package telemetry

import (
	"fmt"
	"time"
)

// Config contains the telemetry configuration for the OpenFroyo engine.
type Config struct {
	// ServiceName is the name of the service for telemetry identification.
	ServiceName string

	// ServiceVersion is the version of the service.
	ServiceVersion string

	// Environment specifies the deployment environment (dev, staging, prod).
	Environment string

	// Logging contains logging configuration.
	Logging LoggingConfig

	// Tracing contains distributed tracing configuration.
	Tracing TracingConfig

	// Metrics contains metrics collection configuration.
	Metrics MetricsConfig

	// Events contains event publishing configuration.
	Events EventsConfig

	// ResourceAttributes are additional resource attributes for telemetry.
	ResourceAttributes map[string]string
}

// LoggingConfig configures structured logging.
type LoggingConfig struct {
	// Level sets the minimum log level (trace, debug, info, warn, error, fatal).
	Level string

	// Format specifies the log format (console, json).
	Format string

	// Output specifies where logs are written (stdout, stderr, file path).
	Output string

	// EnableCaller adds file:line caller information to logs.
	EnableCaller bool

	// EnableSampling enables log sampling for high-frequency logs.
	EnableSampling bool

	// SamplingInitial is the number of messages logged per second initially.
	SamplingInitial int

	// SamplingThereafter logs every Nth message after the initial sample.
	SamplingThereafter int

	// TimeFormat specifies the timestamp format (unix, rfc3339, etc.).
	TimeFormat string
}

// TracingConfig configures distributed tracing.
type TracingConfig struct {
	// Enabled controls whether tracing is active.
	Enabled bool

	// Exporter specifies the trace exporter (jaeger, otlp, stdout, none).
	Exporter string

	// Endpoint is the exporter endpoint (e.g., "localhost:14268" for Jaeger).
	Endpoint string

	// SamplingRate is the trace sampling rate (0.0 to 1.0).
	SamplingRate float64

	// MaxExportBatchSize is the maximum batch size for export.
	MaxExportBatchSize int

	// ExportTimeout is the timeout for trace export.
	ExportTimeout time.Duration

	// Headers are additional headers for OTLP exporter.
	Headers map[string]string

	// Insecure disables TLS for the exporter connection.
	Insecure bool
}

// MetricsConfig configures metrics collection.
type MetricsConfig struct {
	// Enabled controls whether metrics collection is active.
	Enabled bool

	// ListenAddress is the address for the metrics HTTP endpoint.
	ListenAddress string

	// Path is the HTTP path for metrics (default: /metrics).
	Path string

	// Namespace is the metrics namespace prefix.
	Namespace string

	// DefaultHistogramBuckets are the default latency buckets in seconds.
	DefaultHistogramBuckets []float64
}

// EventsConfig configures the event publishing system.
type EventsConfig struct {
	// Enabled controls whether event publishing is active.
	Enabled bool

	// BufferSize is the size of the event buffer.
	BufferSize int

	// FlushInterval is how often to flush buffered events.
	FlushInterval time.Duration

	// MaxBatchSize is the maximum number of events to publish in one batch.
	MaxBatchSize int

	// EnableAsync enables asynchronous event publishing.
	EnableAsync bool
}

// DefaultConfig returns a default telemetry configuration.
func DefaultConfig() *Config {
	return &Config{
		ServiceName:    "openfroyo",
		ServiceVersion: "dev",
		Environment:    "development",
		Logging: LoggingConfig{
			Level:              "info",
			Format:             "console",
			Output:             "stdout",
			EnableCaller:       true,
			EnableSampling:     false,
			SamplingInitial:    100,
			SamplingThereafter: 100,
			TimeFormat:         "rfc3339",
		},
		Tracing: TracingConfig{
			Enabled:            true,
			Exporter:           "stdout",
			Endpoint:           "",
			SamplingRate:       1.0,
			MaxExportBatchSize: 512,
			ExportTimeout:      30 * time.Second,
			Headers:            make(map[string]string),
			Insecure:           true,
		},
		Metrics: MetricsConfig{
			Enabled:       true,
			ListenAddress: ":9090",
			Path:          "/metrics",
			Namespace:     "openfroyo",
			DefaultHistogramBuckets: []float64{
				0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0,
			},
		},
		Events: EventsConfig{
			Enabled:       true,
			BufferSize:    1000,
			FlushInterval: 5 * time.Second,
			MaxBatchSize:  100,
			EnableAsync:   true,
		},
		ResourceAttributes: make(map[string]string),
	}
}

// ProductionConfig returns a production-optimized telemetry configuration.
func ProductionConfig() *Config {
	cfg := DefaultConfig()
	cfg.Environment = "production"
	cfg.Logging.Level = "info"
	cfg.Logging.Format = "json"
	cfg.Logging.EnableSampling = true
	cfg.Logging.TimeFormat = "unix"
	cfg.Tracing.Exporter = "otlp"
	cfg.Tracing.SamplingRate = 0.1 // Sample 10% in production
	cfg.Tracing.Insecure = false
	return cfg
}

// DevelopmentConfig returns a development-optimized telemetry configuration.
func DevelopmentConfig() *Config {
	cfg := DefaultConfig()
	cfg.Environment = "development"
	cfg.Logging.Level = "debug"
	cfg.Logging.Format = "console"
	cfg.Logging.EnableCaller = true
	cfg.Tracing.Exporter = "stdout"
	cfg.Tracing.SamplingRate = 1.0 // Sample all traces in development
	return cfg
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.ServiceName == "" {
		return fmt.Errorf("service name is required")
	}

	if c.ServiceVersion == "" {
		return fmt.Errorf("service version is required")
	}

	// Validate logging level
	validLevels := map[string]bool{
		"trace": true, "debug": true, "info": true,
		"warn": true, "error": true, "fatal": true,
	}
	if !validLevels[c.Logging.Level] {
		return fmt.Errorf("invalid log level: %s", c.Logging.Level)
	}

	// Validate logging format
	if c.Logging.Format != "console" && c.Logging.Format != "json" {
		return fmt.Errorf("invalid log format: %s (must be 'console' or 'json')", c.Logging.Format)
	}

	// Validate tracing exporter
	validExporters := map[string]bool{
		"jaeger": true, "otlp": true, "stdout": true, "none": true,
	}
	if c.Tracing.Enabled && !validExporters[c.Tracing.Exporter] {
		return fmt.Errorf("invalid trace exporter: %s", c.Tracing.Exporter)
	}

	// Validate sampling rate
	if c.Tracing.SamplingRate < 0 || c.Tracing.SamplingRate > 1 {
		return fmt.Errorf("trace sampling rate must be between 0 and 1, got: %f", c.Tracing.SamplingRate)
	}

	// Validate metrics listen address
	if c.Metrics.Enabled && c.Metrics.ListenAddress == "" {
		return fmt.Errorf("metrics listen address is required when metrics are enabled")
	}

	// Validate event buffer size
	if c.Events.Enabled && c.Events.BufferSize <= 0 {
		return fmt.Errorf("event buffer size must be positive, got: %d", c.Events.BufferSize)
	}

	return nil
}
