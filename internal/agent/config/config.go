package config

import (
	"fmt"
	"time"
)

// Config represents the complete agent configuration
type Config struct {
	Agent     AgentConfig     `yaml:"agent"`
	NATS      NATSConfig      `yaml:"nats"`
	Execution ExecutionConfig `yaml:"execution"`
	Logging   LoggingConfig   `yaml:"logging"`
	Pull      PullConfig      `yaml:"pull"`
	Security  SecurityConfig  `yaml:"security"`
}

// AgentConfig contains agent identification and metadata
type AgentConfig struct {
	ID         string            `yaml:"id"`          // Agent unique identifier
	Datacenter string            `yaml:"datacenter"`  // Datacenter location
	Version    string            `yaml:"version"`     // Agent version
	Tags       []string          `yaml:"tags"`        // Agent tags for classification
}

// NATSConfig contains NATS connection settings
type NATSConfig struct {
	Servers       []string         `yaml:"servers"`         // NATS server URLs
	SubjectPrefix string           `yaml:"subject_prefix"`  // Subject prefix for this datacenter
	Connection    ConnectionConfig `yaml:"connection"`      // Connection options
}

// ConnectionConfig contains NATS connection options
type ConnectionConfig struct {
	MaxReconnects  int           `yaml:"max_reconnects"`  // Max reconnection attempts (-1 = infinite)
	ReconnectWait  time.Duration `yaml:"reconnect_wait"`  // Reconnect wait time
	Timeout        time.Duration `yaml:"timeout"`         // Connection timeout
	PingInterval   time.Duration `yaml:"ping_interval"`   // Ping interval
	MaxPingsOut    int           `yaml:"max_pings_out"`   // Max pings outstanding
}

// ExecutionConfig contains task execution settings
type ExecutionConfig struct {
	WorkDir        string        `yaml:"work_dir"`        // Working directory for execution
	ModuleCache    string        `yaml:"module_cache"`    // Module cache directory
	MaxConcurrent  int           `yaml:"max_concurrent"`  // Max concurrent tasks
	TaskTimeout    time.Duration `yaml:"task_timeout"`    // Default task timeout
	RunnerPath     string        `yaml:"runner_path"`     // Path to froyo-runner binary
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level      string `yaml:"level"`       // Log level (debug, info, warn, error)
	File       string `yaml:"file"`        // Log file path
	MaxSize    int    `yaml:"max_size"`    // Max log file size in MB
	MaxBackups int    `yaml:"max_backups"` // Max log files to keep
	MaxAge     int    `yaml:"max_age"`     // Max age of log files in days
	Compress   bool   `yaml:"compress"`    // Compress old log files
}

// PullConfig contains pull mode settings
type PullConfig struct {
	Enabled  bool            `yaml:"enabled"`  // Enable pull mode
	Interval int             `yaml:"interval"` // Poll interval in seconds
	Stacks   []StackSchedule `yaml:"stacks"`   // Stack schedules
}

// StackSchedule defines a periodic stack execution schedule
type StackSchedule struct {
	Name     string `yaml:"name"`     // Stack name
	Interval int    `yaml:"interval"` // Execution interval in seconds
	Enabled  bool   `yaml:"enabled"`  // Enable/disable this schedule
}

// SecurityConfig contains security settings
type SecurityConfig struct {
	TLS  TLSConfig  `yaml:"tls"`  // TLS configuration
	Auth AuthConfig `yaml:"auth"` // Authentication configuration
}

// TLSConfig contains TLS settings
type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`   // Enable TLS
	CertFile string `yaml:"cert_file"` // Client certificate file
	KeyFile  string `yaml:"key_file"`  // Client key file
	CAFile   string `yaml:"ca_file"`   // CA certificate file
	Verify   bool   `yaml:"verify"`    // Verify server certificate
}

// AuthConfig contains authentication settings
type AuthConfig struct {
	Enabled         bool   `yaml:"enabled"`          // Enable authentication
	Token           string `yaml:"token"`            // Authentication token
	CredentialsFile string `yaml:"credentials_file"` // NATS credentials file
	Username        string `yaml:"username"`         // Username (basic auth)
	Password        string `yaml:"password"`         // Password (basic auth)
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate agent config
	if c.Agent.ID == "" {
		return fmt.Errorf("agent.id is required")
	}
	if c.Agent.Datacenter == "" {
		return fmt.Errorf("agent.datacenter is required")
	}

	// Validate NATS config
	if len(c.NATS.Servers) == 0 {
		return fmt.Errorf("nats.servers must have at least one server")
	}
	if c.NATS.SubjectPrefix == "" {
		return fmt.Errorf("nats.subject_prefix is required")
	}

	// Validate execution config
	if c.Execution.WorkDir == "" {
		return fmt.Errorf("execution.work_dir is required")
	}
	if c.Execution.ModuleCache == "" {
		return fmt.Errorf("execution.module_cache is required")
	}
	if c.Execution.MaxConcurrent <= 0 {
		return fmt.Errorf("execution.max_concurrent must be > 0")
	}
	if c.Execution.RunnerPath == "" {
		c.Execution.RunnerPath = "/usr/local/bin/froyo-runner"
	}

	// Validate logging config
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[c.Logging.Level] {
		return fmt.Errorf("logging.level must be one of: debug, info, warn, error")
	}

	// Validate TLS config
	if c.Security.TLS.Enabled {
		if c.Security.TLS.CertFile == "" {
			return fmt.Errorf("security.tls.cert_file is required when TLS is enabled")
		}
		if c.Security.TLS.KeyFile == "" {
			return fmt.Errorf("security.tls.key_file is required when TLS is enabled")
		}
	}

	// Validate auth config
	if c.Security.Auth.Enabled {
		hasAuth := c.Security.Auth.Token != "" ||
			c.Security.Auth.CredentialsFile != "" ||
			(c.Security.Auth.Username != "" && c.Security.Auth.Password != "")

		if !hasAuth {
			return fmt.Errorf("security.auth requires token, credentials_file, or username/password")
		}
	}

	return nil
}

// SetDefaults sets default values for configuration
func (c *Config) SetDefaults() {
	// NATS connection defaults
	if c.NATS.Connection.MaxReconnects == 0 {
		c.NATS.Connection.MaxReconnects = -1 // Infinite
	}
	if c.NATS.Connection.ReconnectWait == 0 {
		c.NATS.Connection.ReconnectWait = 2 * time.Second
	}
	if c.NATS.Connection.Timeout == 0 {
		c.NATS.Connection.Timeout = 5 * time.Second
	}
	if c.NATS.Connection.PingInterval == 0 {
		c.NATS.Connection.PingInterval = 20 * time.Second
	}
	if c.NATS.Connection.MaxPingsOut == 0 {
		c.NATS.Connection.MaxPingsOut = 2
	}

	// Execution defaults
	if c.Execution.MaxConcurrent == 0 {
		c.Execution.MaxConcurrent = 5
	}
	if c.Execution.TaskTimeout == 0 {
		c.Execution.TaskTimeout = 300 * time.Second // 5 minutes
	}
	if c.Execution.RunnerPath == "" {
		c.Execution.RunnerPath = "/usr/local/bin/froyo-runner"
	}

	// Logging defaults
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.MaxSize == 0 {
		c.Logging.MaxSize = 100
	}
	if c.Logging.MaxBackups == 0 {
		c.Logging.MaxBackups = 5
	}
	if c.Logging.MaxAge == 0 {
		c.Logging.MaxAge = 30
	}

	// Pull mode defaults
	if c.Pull.Interval == 0 {
		c.Pull.Interval = 60
	}
}
