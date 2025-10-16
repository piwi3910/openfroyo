package telemetry

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Logger wraps zerolog.Logger with OpenFroyo-specific functionality.
type Logger struct {
	zlog   zerolog.Logger
	config LoggingConfig
}

// loggerContextKey is the context key for logger instances.
type loggerContextKey struct{}

// NewLogger creates a new logger with the given configuration.
func NewLogger(cfg LoggingConfig) (*Logger, error) {
	// Configure output writer
	var writer io.Writer
	switch cfg.Output {
	case "stdout":
		writer = os.Stdout
	case "stderr":
		writer = os.Stderr
	default:
		// If it's not stdout/stderr, assume it's a file path
		file, err := os.OpenFile(cfg.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, err
		}
		writer = file
	}

	// Configure format
	if cfg.Format == "console" {
		writer = zerolog.ConsoleWriter{
			Out:        writer,
			TimeFormat: getTimeFormat(cfg.TimeFormat),
			NoColor:    false,
		}
	}

	// Configure time format
	switch cfg.TimeFormat {
	case "unix":
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	case "unixms":
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	case "unixmicro":
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMicro
	default: // rfc3339
		zerolog.TimeFieldFormat = time.RFC3339
	}

	// Create base logger
	zlog := zerolog.New(writer).With().Timestamp().Logger()

	// Set log level
	level := parseLogLevel(cfg.Level)
	zlog = zlog.Level(level)

	// Enable caller information if requested
	if cfg.EnableCaller {
		zlog = zlog.With().Caller().Logger()
	}

	// Configure sampling if enabled
	if cfg.EnableSampling {
		sampler := &zerolog.BurstSampler{
			Burst:       uint32(cfg.SamplingInitial),
			Period:      1 * time.Second,
			NextSampler: &zerolog.BasicSampler{N: uint32(cfg.SamplingThereafter)},
		}
		zlog = zlog.Sample(sampler)
	}

	return &Logger{
		zlog:   zlog,
		config: cfg,
	}, nil
}

// NewComponentLogger creates a child logger for a specific component.
func (l *Logger) NewComponentLogger(component string) *Logger {
	return &Logger{
		zlog:   l.zlog.With().Str("component", component).Logger(),
		config: l.config,
	}
}

// WithContext adds the logger to the context.
func (l *Logger) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, loggerContextKey{}, l)
}

// FromContext retrieves the logger from the context.
// If no logger is found, it returns a default logger.
func FromContext(ctx context.Context) *Logger {
	if l, ok := ctx.Value(loggerContextKey{}).(*Logger); ok {
		return l
	}
	// Return a minimal default logger
	return &Logger{
		zlog: zerolog.New(os.Stdout).With().Timestamp().Logger(),
	}
}

// WithFields returns a logger with additional fields.
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	ctx := l.zlog.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	return &Logger{
		zlog:   ctx.Logger(),
		config: l.config,
	}
}

// WithField returns a logger with a single additional field.
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{
		zlog:   l.zlog.With().Interface(key, value).Logger(),
		config: l.config,
	}
}

// WithRunID adds a run_id field to the logger.
func (l *Logger) WithRunID(runID string) *Logger {
	return l.WithField("run_id", runID)
}

// WithResourceID adds a resource_id field to the logger.
func (l *Logger) WithResourceID(resourceID string) *Logger {
	return l.WithField("resource_id", resourceID)
}

// WithPlanUnitID adds a plan_unit_id field to the logger.
func (l *Logger) WithPlanUnitID(planUnitID string) *Logger {
	return l.WithField("plan_unit_id", planUnitID)
}

// WithProvider adds provider information to the logger.
func (l *Logger) WithProvider(name, version string) *Logger {
	return &Logger{
		zlog: l.zlog.With().
			Str("provider_name", name).
			Str("provider_version", version).
			Logger(),
		config: l.config,
	}
}

// WithError adds error information to the logger.
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		zlog:   l.zlog.With().Err(err).Logger(),
		config: l.config,
	}
}

// Trace logs a trace-level message.
func (l *Logger) Trace(msg string) {
	l.zlog.Trace().Msg(msg)
}

// Tracef logs a formatted trace-level message.
func (l *Logger) Tracef(format string, args ...interface{}) {
	l.zlog.Trace().Msgf(format, args...)
}

// Debug logs a debug-level message.
func (l *Logger) Debug(msg string) {
	l.zlog.Debug().Msg(msg)
}

// Debugf logs a formatted debug-level message.
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.zlog.Debug().Msgf(format, args...)
}

// Info logs an info-level message.
func (l *Logger) Info(msg string) {
	l.zlog.Info().Msg(msg)
}

// Infof logs a formatted info-level message.
func (l *Logger) Infof(format string, args ...interface{}) {
	l.zlog.Info().Msgf(format, args...)
}

// Warn logs a warning-level message.
func (l *Logger) Warn(msg string) {
	l.zlog.Warn().Msg(msg)
}

// Warnf logs a formatted warning-level message.
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.zlog.Warn().Msgf(format, args...)
}

// Error logs an error-level message.
func (l *Logger) Error(msg string) {
	l.zlog.Error().Msg(msg)
}

// Errorf logs a formatted error-level message.
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.zlog.Error().Msgf(format, args...)
}

// Fatal logs a fatal-level message and exits.
func (l *Logger) Fatal(msg string) {
	l.zlog.Fatal().Msg(msg)
}

// Fatalf logs a formatted fatal-level message and exits.
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.zlog.Fatal().Msgf(format, args...)
}

// parseLogLevel converts a string log level to zerolog.Level.
func parseLogLevel(level string) zerolog.Level {
	switch level {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

// getTimeFormat returns the appropriate time format for console output.
func getTimeFormat(format string) string {
	switch format {
	case "unix":
		return "unix"
	case "rfc3339":
		return time.RFC3339
	default:
		return time.RFC3339
	}
}

// Hook provides a way to add custom hooks to the logger.
type Hook interface {
	Run(e *zerolog.Event, level zerolog.Level, msg string)
}

// AddHook adds a hook to the logger.
func (l *Logger) AddHook(hook zerolog.Hook) *Logger {
	return &Logger{
		zlog:   l.zlog.Hook(hook),
		config: l.config,
	}
}
