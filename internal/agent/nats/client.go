package nats

import (
	"crypto/tls"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/piwi3910/openfroyo/internal/agent/config"
)

// Client wraps a NATS connection with OpenFroyo-specific functionality
type Client struct {
	conn      *nats.Conn
	config    *config.Config
	connected bool
	logger    *log.Logger
}

// NewClient creates a new NATS client
func NewClient(cfg *config.Config, logger *log.Logger) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	return &Client{
		config: cfg,
		logger: logger,
	}, nil
}

// Connect establishes a connection to NATS server(s)
func (c *Client) Connect() error {
	opts := []nats.Option{
		nats.Name(fmt.Sprintf("froyo-agent-%s", c.config.Agent.ID)),
		nats.MaxReconnects(c.config.NATS.Connection.MaxReconnects),
		nats.ReconnectWait(c.config.NATS.Connection.ReconnectWait),
		nats.Timeout(c.config.NATS.Connection.Timeout),
		nats.PingInterval(c.config.NATS.Connection.PingInterval),
		nats.MaxPingsOutstanding(c.config.NATS.Connection.MaxPingsOut),
		nats.ReconnectHandler(c.onReconnect),
		nats.DisconnectErrHandler(c.onDisconnect),
		nats.ClosedHandler(c.onClosed),
		nats.ErrorHandler(c.onError),
	}

	// Add TLS if enabled
	if c.config.Security.TLS.Enabled {
		tlsConfig, err := c.buildTLSConfig()
		if err != nil {
			return fmt.Errorf("failed to build TLS config: %w", err)
		}
		opts = append(opts, nats.Secure(tlsConfig))
	}

	// Add authentication if enabled
	if c.config.Security.Auth.Enabled {
		authOpts, err := c.buildAuthOptions()
		if err != nil {
			return fmt.Errorf("failed to build auth options: %w", err)
		}
		opts = append(opts, authOpts...)
	}

	// Connect to NATS
	// Join all server URLs with comma for multi-server support
	serverURL := strings.Join(c.config.NATS.Servers, ",")

	conn, err := nats.Connect(serverURL, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}

	c.conn = conn
	c.connected = true
	c.logger.Printf("[INFO] Connected to NATS: %s", conn.ConnectedUrl())

	return nil
}

// Close closes the NATS connection
func (c *Client) Close() error {
	if c.conn != nil {
		c.connected = false
		c.conn.Close()
		c.logger.Println("[INFO] NATS connection closed")
	}
	return nil
}

// Publish publishes a message to a subject
func (c *Client) Publish(subject string, data []byte) error {
	if c.conn == nil {
		return fmt.Errorf("not connected to NATS")
	}
	return c.conn.Publish(subject, data)
}

// Subscribe subscribes to a subject with a message handler
func (c *Client) Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("not connected to NATS")
	}

	sub, err := c.conn.Subscribe(subject, handler)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to %s: %w", subject, err)
	}

	c.logger.Printf("[INFO] Subscribed to: %s", subject)
	return sub, nil
}

// QueueSubscribe subscribes to a subject as part of a queue group
func (c *Client) QueueSubscribe(subject, queue string, handler nats.MsgHandler) (*nats.Subscription, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("not connected to NATS")
	}

	sub, err := c.conn.QueueSubscribe(subject, queue, handler)
	if err != nil {
		return nil, fmt.Errorf("failed to queue subscribe to %s: %w", subject, err)
	}

	c.logger.Printf("[INFO] Queue subscribed to: %s (queue: %s)", subject, queue)
	return sub, nil
}

// Request makes a request and waits for a response
func (c *Client) Request(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("not connected to NATS")
	}
	return c.conn.Request(subject, data, timeout)
}

// IsConnected returns whether the client is connected
func (c *Client) IsConnected() bool {
	return c.connected && c.conn != nil && !c.conn.IsClosed()
}

// buildTLSConfig builds TLS configuration from config
func (c *Client) buildTLSConfig() (*tls.Config, error) {
	tlsCfg := &tls.Config{
		InsecureSkipVerify: !c.config.Security.TLS.Verify,
	}

	// Load client certificate if provided
	if c.config.Security.TLS.CertFile != "" && c.config.Security.TLS.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(
			c.config.Security.TLS.CertFile,
			c.config.Security.TLS.KeyFile,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		tlsCfg.Certificates = []tls.Certificate{cert}
	}

	// Load CA certificate if provided
	if c.config.Security.TLS.CAFile != "" {
		// Note: You'd need to implement CA loading here
		// This is a placeholder
		c.logger.Println("[WARN] CA file loading not yet implemented")
	}

	return tlsCfg, nil
}

// buildAuthOptions builds authentication options from config
func (c *Client) buildAuthOptions() ([]nats.Option, error) {
	var opts []nats.Option

	// Token authentication
	if c.config.Security.Auth.Token != "" {
		opts = append(opts, nats.Token(c.config.Security.Auth.Token))
		return opts, nil
	}

	// Credentials file authentication
	if c.config.Security.Auth.CredentialsFile != "" {
		opts = append(opts, nats.UserCredentials(c.config.Security.Auth.CredentialsFile))
		return opts, nil
	}

	// Username/password authentication
	if c.config.Security.Auth.Username != "" {
		opts = append(opts, nats.UserInfo(
			c.config.Security.Auth.Username,
			c.config.Security.Auth.Password,
		))
		return opts, nil
	}

	return nil, fmt.Errorf("no valid authentication method configured")
}

// Event handlers

func (c *Client) onReconnect(nc *nats.Conn) {
	c.logger.Printf("[INFO] Reconnected to NATS: %s", nc.ConnectedUrl())
	c.connected = true
}

func (c *Client) onDisconnect(nc *nats.Conn, err error) {
	if err != nil {
		c.logger.Printf("[WARN] Disconnected from NATS: %v", err)
	} else {
		c.logger.Println("[WARN] Disconnected from NATS")
	}
	c.connected = false
}

func (c *Client) onClosed(nc *nats.Conn) {
	c.logger.Println("[INFO] NATS connection closed")
	c.connected = false
}

func (c *Client) onError(nc *nats.Conn, sub *nats.Subscription, err error) {
	if sub != nil {
		c.logger.Printf("[ERROR] NATS error on subject %s: %v", sub.Subject, err)
	} else {
		c.logger.Printf("[ERROR] NATS error: %v", err)
	}
}
