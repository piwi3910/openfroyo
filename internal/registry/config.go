package registry

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the registry configuration
type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Storage StorageConfig `yaml:"storage"`
	Auth    AuthConfig    `yaml:"auth"`
	Logging LoggingConfig `yaml:"logging"`
}

// ServerConfig contains HTTP server settings
type ServerConfig struct {
	Listen           string        `yaml:"listen"`            // Listen address
	ReadTimeout      time.Duration `yaml:"read_timeout"`      // Read timeout
	WriteTimeout     time.Duration `yaml:"write_timeout"`     // Write timeout
	MaxHeaderBytes   int           `yaml:"max_header_bytes"`  // Max header bytes
	EnableCORS       bool          `yaml:"enable_cors"`       // Enable CORS
	EnableMetrics    bool          `yaml:"enable_metrics"`    // Enable /metrics endpoint
	EnableHealthz    bool          `yaml:"enable_healthz"`    // Enable /healthz endpoint
	TrustedProxies   []string      `yaml:"trusted_proxies"`   // Trusted proxy IPs
}

// StorageConfig contains storage settings
type StorageConfig struct {
	ModulesDir      string `yaml:"modules_dir"`       // Directory containing WASM modules
	ChecksumsFile   string `yaml:"checksums_file"`    // Path to checksums file
	MetadataFile    string `yaml:"metadata_file"`     // Path to metadata file
	EnableCache     bool   `yaml:"enable_cache"`      // Enable in-memory caching
	CacheTTL        int    `yaml:"cache_ttl"`         // Cache TTL in seconds
}

// AuthConfig contains authentication settings
type AuthConfig struct {
	Enabled    bool              `yaml:"enabled"`     // Enable authentication
	Type       string            `yaml:"type"`        // Auth type: "none", "api_key", "jwt"
	APIKeys    []string          `yaml:"api_keys"`    // Valid API keys
	JWTSecret  string            `yaml:"jwt_secret"`  // JWT signing secret
	JWTIssuer  string            `yaml:"jwt_issuer"`  // JWT issuer
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level       string `yaml:"level"`        // Log level
	Format      string `yaml:"format"`       // Log format: "text", "json"
	AccessLog   bool   `yaml:"access_log"`   // Enable access logging
	File        string `yaml:"file"`         // Log file path
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	cfg.SetDefaults()

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	cfg := &Config{
		Server: ServerConfig{
			Listen:         ":8080",
			ReadTimeout:    30 * time.Second,
			WriteTimeout:   30 * time.Second,
			MaxHeaderBytes: 1 << 20, // 1MB
			EnableCORS:     true,
			EnableMetrics:  true,
			EnableHealthz:  true,
		},
		Storage: StorageConfig{
			ModulesDir:    "/var/lib/froyo-registry/modules",
			ChecksumsFile: "/var/lib/froyo-registry/checksums.yml",
			MetadataFile:  "/var/lib/froyo-registry/metadata.yml",
			EnableCache:   true,
			CacheTTL:      3600,
		},
		Auth: AuthConfig{
			Enabled: false,
			Type:    "none",
		},
		Logging: LoggingConfig{
			Level:     "info",
			Format:    "text",
			AccessLog: true,
		},
	}
	return cfg
}

// SetDefaults sets default values for missing fields
func (c *Config) SetDefaults() {
	if c.Server.Listen == "" {
		c.Server.Listen = ":8080"
	}
	if c.Server.ReadTimeout == 0 {
		c.Server.ReadTimeout = 30 * time.Second
	}
	if c.Server.WriteTimeout == 0 {
		c.Server.WriteTimeout = 30 * time.Second
	}
	if c.Server.MaxHeaderBytes == 0 {
		c.Server.MaxHeaderBytes = 1 << 20
	}

	if c.Storage.ModulesDir == "" {
		c.Storage.ModulesDir = "/var/lib/froyo-registry/modules"
	}
	if c.Storage.ChecksumsFile == "" {
		c.Storage.ChecksumsFile = "/var/lib/froyo-registry/checksums.yml"
	}
	if c.Storage.CacheTTL == 0 {
		c.Storage.CacheTTL = 3600
	}

	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.Format == "" {
		c.Logging.Format = "text"
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Storage.ModulesDir == "" {
		return fmt.Errorf("storage.modules_dir is required")
	}

	if c.Auth.Enabled {
		if c.Auth.Type == "" {
			return fmt.Errorf("auth.type is required when auth is enabled")
		}
		if c.Auth.Type == "api_key" && len(c.Auth.APIKeys) == 0 {
			return fmt.Errorf("auth.api_keys is required when type is api_key")
		}
		if c.Auth.Type == "jwt" && c.Auth.JWTSecret == "" {
			return fmt.Errorf("auth.jwt_secret is required when type is jwt")
		}
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

	return nil
}
