package ssh

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
)

// SSHClient implements the Transport interface with connection pooling.
type SSHClient struct {
	config *Config

	// Connection management
	client      *ssh.Client
	connMu      sync.RWMutex
	isConnected bool
	connectedAt time.Time

	// Connection pool for reusable sessions
	pool       *connectionPool
	lastUsedAt time.Time

	// Executor for command execution
	executor *executor

	// File transfer handler
	fileTransfer *fileTransfer
}

// connectionPool manages a pool of SSH connections for efficiency.
type connectionPool struct {
	config   *Config
	mu       sync.Mutex
	conns    []*pooledConnection
	maxSize  int
	idleTime time.Duration
}

// pooledConnection wraps an SSH client connection with metadata.
type pooledConnection struct {
	client      *ssh.Client
	createdAt   time.Time
	lastUsedAt  time.Time
	inUse       bool
	usageCount  int
	mu          sync.Mutex
}

// NewSSHClient creates a new SSH transport client.
func NewSSHClient(config *Config) (*SSHClient, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	client := &SSHClient{
		config: config,
	}

	if config.EnableConnectionPooling {
		client.pool = &connectionPool{
			config:   config,
			conns:    make([]*pooledConnection, 0, config.MaxPoolSize),
			maxSize:  config.MaxPoolSize,
			idleTime: config.PoolIdleTimeout,
		}
	}

	return client, nil
}

// Connect establishes an SSH connection to the remote host.
func (c *SSHClient) Connect(ctx context.Context) error {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	if c.isConnected && c.client != nil {
		// Already connected, verify connection is still alive
		if err := c.healthCheckInternal(); err == nil {
			return nil
		}
		// Connection is dead, close it and reconnect
		log.Warn().Msg("existing connection is dead, reconnecting")
		_ = c.client.Close()
	}

	// Build SSH client config
	clientConfig, err := c.config.BuildSSHClientConfig()
	if err != nil {
		return &TransportError{
			Op:          "connect",
			Err:         err,
			IsTemporary: false,
			IsAuthError: true,
		}
	}

	// Handle proxy/jump host if configured
	if c.config.IsProxyEnabled() {
		return c.connectViaProxy(ctx, clientConfig)
	}

	// Direct connection
	return c.connectDirect(ctx, clientConfig)
}

// connectDirect establishes a direct SSH connection.
func (c *SSHClient) connectDirect(ctx context.Context, clientConfig *ssh.ClientConfig) error {
	address := c.config.Address()
	log.Debug().Str("address", address).Msg("establishing SSH connection")

	// Create a channel to handle connection with timeout
	connChan := make(chan *ssh.Client, 1)
	errChan := make(chan error, 1)

	go func() {
		client, err := ssh.Dial("tcp", address, clientConfig)
		if err != nil {
			errChan <- err
			return
		}
		connChan <- client
	}()

	select {
	case <-ctx.Done():
		return &TransportError{
			Op:          "connect",
			Err:         ctx.Err(),
			IsTemporary: true,
			IsAuthError: false,
		}
	case err := <-errChan:
		return &TransportError{
			Op:          "connect",
			Err:         err,
			IsTemporary: true,
			IsAuthError: false,
		}
	case client := <-connChan:
		c.client = client
		c.isConnected = true
		c.connectedAt = time.Now()
		c.lastUsedAt = time.Now()

		// Initialize executor and file transfer
		c.executor = &executor{
			client: c,
			config: c.config,
		}
		c.fileTransfer = &fileTransfer{
			client: c,
			config: c.config,
		}

		// Start keep-alive if configured
		if c.config.KeepAliveInterval > 0 {
			go c.keepAlive()
		}

		log.Info().Str("address", address).Msg("SSH connection established")
		return nil
	}
}

// connectViaProxy establishes an SSH connection through a proxy/jump host.
func (c *SSHClient) connectViaProxy(ctx context.Context, targetConfig *ssh.ClientConfig) error {
	// First, connect to the proxy
	proxyConfig := &Config{
		Host:                  c.config.ProxyHost,
		Port:                  c.config.ProxyPort,
		User:                  c.config.ProxyUser,
		AuthMethod:            c.config.ProxyAuthMethod,
		Password:              c.config.ProxyPassword,
		PrivateKeyPath:        c.config.ProxyPrivateKeyPath,
		ConnectionTimeout:     c.config.ConnectionTimeout,
		StrictHostKeyChecking: c.config.StrictHostKeyChecking,
		KnownHostsPath:        c.config.KnownHostsPath,
	}

	proxyClientConfig, err := proxyConfig.BuildSSHClientConfig()
	if err != nil {
		return fmt.Errorf("failed to build proxy config: %w", err)
	}

	log.Debug().Str("proxy", proxyConfig.Address()).Msg("connecting to proxy host")

	proxyClient, err := ssh.Dial("tcp", proxyConfig.Address(), proxyClientConfig)
	if err != nil {
		return &TransportError{
			Op:          "connect-proxy",
			Err:         err,
			IsTemporary: true,
			IsAuthError: false,
		}
	}

	// Now connect to the target through the proxy
	targetAddress := c.config.Address()
	log.Debug().Str("target", targetAddress).Msg("connecting to target through proxy")

	proxyConn, err := proxyClient.Dial("tcp", targetAddress)
	if err != nil {
		_ = proxyClient.Close()
		return &TransportError{
			Op:          "connect-via-proxy",
			Err:         err,
			IsTemporary: true,
			IsAuthError: false,
		}
	}

	ncc, chans, reqs, err := ssh.NewClientConn(proxyConn, targetAddress, targetConfig)
	if err != nil {
		_ = proxyConn.Close()
		_ = proxyClient.Close()
		return &TransportError{
			Op:          "connect-via-proxy",
			Err:         err,
			IsTemporary: true,
			IsAuthError: true,
		}
	}

	c.client = ssh.NewClient(ncc, chans, reqs)
	c.isConnected = true
	c.connectedAt = time.Now()
	c.lastUsedAt = time.Now()

	// Initialize executor and file transfer
	c.executor = &executor{
		client: c,
		config: c.config,
	}
	c.fileTransfer = &fileTransfer{
		client: c,
		config: c.config,
	}

	if c.config.KeepAliveInterval > 0 {
		go c.keepAlive()
	}

	log.Info().Str("target", targetAddress).Str("proxy", proxyConfig.Address()).Msg("SSH connection established via proxy")
	return nil
}

// Disconnect closes the SSH connection and releases all resources.
func (c *SSHClient) Disconnect() error {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	if !c.isConnected || c.client == nil {
		return nil
	}

	log.Debug().Str("host", c.config.Host).Msg("closing SSH connection")

	err := c.client.Close()
	c.client = nil
	c.isConnected = false

	if err != nil {
		return &TransportError{
			Op:          "disconnect",
			Err:         err,
			IsTemporary: false,
			IsAuthError: false,
		}
	}

	return nil
}

// IsConnected returns true if the transport has an active connection.
func (c *SSHClient) IsConnected() bool {
	c.connMu.RLock()
	defer c.connMu.RUnlock()
	return c.isConnected
}

// HealthCheck verifies the connection is still alive and responsive.
func (c *SSHClient) HealthCheck(ctx context.Context) error {
	c.connMu.RLock()
	defer c.connMu.RUnlock()

	if !c.isConnected || c.client == nil {
		return &TransportError{
			Op:          "healthcheck",
			Err:         fmt.Errorf("not connected"),
			IsTemporary: false,
			IsAuthError: false,
		}
	}

	return c.healthCheckInternal()
}

// healthCheckInternal performs the actual health check (must be called with lock held).
func (c *SSHClient) healthCheckInternal() error {
	// Create a new session to test the connection
	session, err := c.client.NewSession()
	if err != nil {
		return &TransportError{
			Op:          "healthcheck",
			Err:         err,
			IsTemporary: true,
			IsAuthError: false,
		}
	}
	defer session.Close()

	// Run a simple command
	if err := session.Run("true"); err != nil {
		return &TransportError{
			Op:          "healthcheck",
			Err:         err,
			IsTemporary: true,
			IsAuthError: false,
		}
	}

	return nil
}

// keepAlive sends periodic keep-alive messages to keep the connection alive.
func (c *SSHClient) keepAlive() {
	ticker := time.NewTicker(c.config.KeepAliveInterval)
	defer ticker.Stop()

	retries := 0
	maxRetries := c.config.MaxKeepAliveRetries

	for range ticker.C {
		c.connMu.RLock()
		if !c.isConnected || c.client == nil {
			c.connMu.RUnlock()
			return
		}
		c.connMu.RUnlock()

		// Send a keep-alive request
		_, _, err := c.client.SendRequest("keepalive@openssh.com", true, nil)
		if err != nil {
			retries++
			log.Warn().Err(err).Int("retries", retries).Msg("keep-alive failed")
			if retries >= maxRetries {
				log.Error().Msg("keep-alive failed too many times, connection may be dead")
				return
			}
		} else {
			retries = 0
			c.lastUsedAt = time.Now()
		}
	}
}

// GetConnectionInfo returns information about the current connection.
func (c *SSHClient) GetConnectionInfo() ConnectionInfo {
	c.connMu.RLock()
	defer c.connMu.RUnlock()

	return ConnectionInfo{
		Host:         c.config.Host,
		Port:         c.config.Port,
		User:         c.config.User,
		ConnectedAt:  c.connectedAt,
		LastActivity: c.lastUsedAt,
		IsPooled:     c.config.EnableConnectionPooling,
	}
}

// getClient returns the underlying SSH client (used internally by executor and file transfer).
func (c *SSHClient) getClient() (*ssh.Client, error) {
	c.connMu.RLock()
	defer c.connMu.RUnlock()

	if !c.isConnected || c.client == nil {
		return nil, &TransportError{
			Op:          "get-client",
			Err:         fmt.Errorf("not connected"),
			IsTemporary: false,
			IsAuthError: false,
		}
	}

	c.lastUsedAt = time.Now()
	return c.client, nil
}
