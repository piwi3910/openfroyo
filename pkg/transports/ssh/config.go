package ssh

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// AuthMethod represents the type of SSH authentication.
type AuthMethod string

const (
	// AuthMethodPassword uses password authentication
	AuthMethodPassword AuthMethod = "password"

	// AuthMethodKey uses private key authentication
	AuthMethodKey AuthMethod = "key"

	// AuthMethodAgent uses SSH agent authentication
	AuthMethodAgent AuthMethod = "agent"
)

// Config holds SSH connection configuration.
type Config struct {
	// Host is the remote hostname or IP address
	Host string

	// Port is the SSH port (default: 22)
	Port int

	// User is the SSH username
	User string

	// AuthMethod specifies which authentication method to use
	AuthMethod AuthMethod

	// Password for password-based authentication
	Password string

	// PrivateKeyPath is the path to the private key file
	PrivateKeyPath string

	// PrivateKeyPassphrase is the passphrase for encrypted private keys
	PrivateKeyPassphrase string

	// KnownHostsPath is the path to the known_hosts file
	// If empty, host key verification is disabled (not recommended for production)
	KnownHostsPath string

	// StrictHostKeyChecking enables strict host key verification
	// When true, unknown hosts will be rejected
	StrictHostKeyChecking bool

	// ConnectionTimeout is the timeout for establishing a connection
	ConnectionTimeout time.Duration

	// CommandTimeout is the default timeout for command execution
	CommandTimeout time.Duration

	// KeepAliveInterval is the interval for sending keep-alive messages
	// Set to 0 to disable keep-alive
	KeepAliveInterval time.Duration

	// MaxKeepAliveRetries is the maximum number of keep-alive retries before giving up
	MaxKeepAliveRetries int

	// EnableConnectionPooling enables connection pooling for efficiency
	EnableConnectionPooling bool

	// MaxPoolSize is the maximum number of connections in the pool
	MaxPoolSize int

	// PoolIdleTimeout is how long idle connections are kept in the pool
	PoolIdleTimeout time.Duration

	// EnableCompression enables SSH compression
	EnableCompression bool

	// ProxyHost is the hostname of a jump/proxy host (optional)
	ProxyHost string

	// ProxyPort is the port of the proxy host
	ProxyPort int

	// ProxyUser is the username for the proxy host
	ProxyUser string

	// ProxyAuthMethod is the authentication method for the proxy
	ProxyAuthMethod AuthMethod

	// ProxyPassword is the password for proxy authentication
	ProxyPassword string

	// ProxyPrivateKeyPath is the path to the proxy's private key
	ProxyPrivateKeyPath string
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig(host string, user string) *Config {
	return &Config{
		Host:                    host,
		Port:                    22,
		User:                    user,
		AuthMethod:              AuthMethodKey,
		KnownHostsPath:          filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts"),
		StrictHostKeyChecking:   true,
		ConnectionTimeout:       30 * time.Second,
		CommandTimeout:          5 * time.Minute,
		KeepAliveInterval:       0, // Disabled by default
		MaxKeepAliveRetries:     3,
		EnableConnectionPooling: true,
		MaxPoolSize:             5,
		PoolIdleTimeout:         10 * time.Minute,
		EnableCompression:       false,
		ProxyPort:               22,
	}
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}

	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Port)
	}

	if c.User == "" {
		return fmt.Errorf("user is required")
	}

	switch c.AuthMethod {
	case AuthMethodPassword:
		if c.Password == "" {
			return fmt.Errorf("password is required for password authentication")
		}
	case AuthMethodKey:
		if c.PrivateKeyPath == "" {
			// Try default key locations
			homeDir := os.Getenv("HOME")
			defaultKeys := []string{
				filepath.Join(homeDir, ".ssh", "id_ed25519"),
				filepath.Join(homeDir, ".ssh", "id_rsa"),
				filepath.Join(homeDir, ".ssh", "id_ecdsa"),
			}
			for _, keyPath := range defaultKeys {
				if _, err := os.Stat(keyPath); err == nil {
					c.PrivateKeyPath = keyPath
					break
				}
			}
			if c.PrivateKeyPath == "" {
				return fmt.Errorf("private key path is required for key authentication and no default key found")
			}
		}
		if _, err := os.Stat(c.PrivateKeyPath); os.IsNotExist(err) {
			return fmt.Errorf("private key file not found: %s", c.PrivateKeyPath)
		}
	case AuthMethodAgent:
		// Agent authentication doesn't require additional config
	default:
		return fmt.Errorf("unsupported auth method: %s", c.AuthMethod)
	}

	if c.ConnectionTimeout <= 0 {
		return fmt.Errorf("connection timeout must be positive")
	}

	if c.CommandTimeout <= 0 {
		return fmt.Errorf("command timeout must be positive")
	}

	if c.EnableConnectionPooling && c.MaxPoolSize <= 0 {
		return fmt.Errorf("max pool size must be positive when pooling is enabled")
	}

	// Validate proxy configuration if proxy is specified
	if c.ProxyHost != "" {
		if c.ProxyPort <= 0 || c.ProxyPort > 65535 {
			return fmt.Errorf("invalid proxy port: %d", c.ProxyPort)
		}
		if c.ProxyUser == "" {
			return fmt.Errorf("proxy user is required when proxy host is specified")
		}
	}

	return nil
}

// BuildSSHClientConfig creates an ssh.ClientConfig from the Config.
func (c *Config) BuildSSHClientConfig() (*ssh.ClientConfig, error) {
	var authMethods []ssh.AuthMethod

	switch c.AuthMethod {
	case AuthMethodPassword:
		// Add password authentication
		authMethods = append(authMethods, ssh.Password(c.Password))

		// Add keyboard-interactive authentication (required by many SSH servers)
		// This handles the common "Password:" prompt
		authMethods = append(authMethods, ssh.KeyboardInteractive(
			func(user, instruction string, questions []string, echos []bool) ([]string, error) {
				answers := make([]string, len(questions))
				for i := range answers {
					answers[i] = c.Password
				}
				return answers, nil
			},
		))

	case AuthMethodKey:
		keyBytes, err := os.ReadFile(c.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key: %w", err)
		}

		var signer ssh.Signer
		if c.PrivateKeyPassphrase != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(keyBytes, []byte(c.PrivateKeyPassphrase))
		} else {
			signer, err = ssh.ParsePrivateKey(keyBytes)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}

		authMethods = append(authMethods, ssh.PublicKeys(signer))

	case AuthMethodAgent:
		// SSH agent support would be implemented here
		// This requires connecting to the SSH agent socket
		return nil, fmt.Errorf("SSH agent authentication not yet implemented")
	}

	// Configure host key callback
	var hostKeyCallback ssh.HostKeyCallback
	if c.KnownHostsPath != "" && c.StrictHostKeyChecking {
		var err error
		hostKeyCallback, err = knownhosts.New(c.KnownHostsPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load known_hosts: %w", err)
		}
	} else {
		// Insecure: accept any host key (only for testing/development)
		hostKeyCallback = ssh.InsecureIgnoreHostKey()
	}

	clientConfig := &ssh.ClientConfig{
		User:            c.User,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         c.ConnectionTimeout,
	}

	return clientConfig, nil
}

// Address returns the formatted SSH address (host:port).
func (c *Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// ProxyAddress returns the formatted proxy address (host:port).
func (c *Config) ProxyAddress() string {
	if c.ProxyHost == "" {
		return ""
	}
	return fmt.Sprintf("%s:%d", c.ProxyHost, c.ProxyPort)
}

// IsProxyEnabled returns true if a proxy/jump host is configured.
func (c *Config) IsProxyEnabled() bool {
	return c.ProxyHost != ""
}
