package telemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Telemetry provides a unified telemetry interface combining logging, tracing, metrics, and events.
type Telemetry struct {
	Logger    *Logger
	Tracer    *Tracer
	Metrics   *Metrics
	Events    *EventPublisher
	Config    *Config
}

// telemetryContextKey is the context key for telemetry instances.
type telemetryContextKey struct{}

// NewTelemetry creates a new telemetry instance from configuration.
func NewTelemetry(cfg *Config) (*Telemetry, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Initialize logger
	logger, err := NewLogger(cfg.Logging)
	if err != nil {
		return nil, err
	}

	// Initialize tracer
	tracer, err := NewTracer(cfg.Tracing, cfg.ServiceName, cfg.ServiceVersion, cfg.Environment)
	if err != nil {
		return nil, err
	}

	// Initialize metrics
	metrics, err := NewMetrics(cfg.Metrics)
	if err != nil {
		return nil, err
	}

	// Initialize event publisher
	events, err := NewEventPublisher(cfg.Events)
	if err != nil {
		return nil, err
	}

	return &Telemetry{
		Logger:  logger,
		Tracer:  tracer,
		Metrics: metrics,
		Events:  events,
		Config:  cfg,
	}, nil
}

// WithContext adds the telemetry instance to the context.
func (t *Telemetry) WithContext(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, telemetryContextKey{}, t)
	ctx = t.Logger.WithContext(ctx)
	return ctx
}

// FromContext retrieves the telemetry instance from the context.
// If no telemetry is found, it returns nil.
func FromTelemetryContext(ctx context.Context) *Telemetry {
	if t, ok := ctx.Value(telemetryContextKey{}).(*Telemetry); ok {
		return t
	}
	return nil
}

// Shutdown gracefully shuts down all telemetry components.
func (t *Telemetry) Shutdown(ctx context.Context) error {
	// Shutdown in reverse order of initialization
	if err := t.Events.Shutdown(ctx); err != nil {
		return err
	}

	if err := t.Tracer.Shutdown(ctx); err != nil {
		return err
	}

	// Metrics server is not explicitly shut down here as it may need to continue
	// serving metrics until the very end of the application lifecycle

	return nil
}

// Flush forces all pending telemetry data to be exported.
func (t *Telemetry) Flush(ctx context.Context) error {
	return t.Tracer.ForceFlush(ctx)
}

// StartMetricsServer starts the metrics HTTP server if metrics are enabled.
func (t *Telemetry) StartMetricsServer() error {
	return t.Metrics.StartMetricsServer()
}

// Context Helpers for common instrumentation patterns

// InstrumentedContext creates a context with telemetry, logger fields, and a trace span.
type InstrumentedContext struct {
	Ctx    context.Context
	Span   trace.Span
	Logger *Logger
	Timer  *Timer
}

// StartOperation begins an instrumented operation with logging, tracing, and timing.
func StartOperation(ctx context.Context, operation string, attrs ...attribute.KeyValue) *InstrumentedContext {
	tel := FromTelemetryContext(ctx)
	if tel == nil {
		return &InstrumentedContext{
			Ctx:    ctx,
			Logger: FromContext(ctx),
			Timer:  NewTimer(),
		}
	}

	// Start trace span
	spanCtx, span := tel.Tracer.StartSpan(ctx, operation, attrs...)

	// Create logger with operation field
	logger := tel.Logger.WithField("operation", operation)

	// Add trace context to logger if available
	if span.SpanContext().IsValid() {
		logger = logger.WithFields(map[string]interface{}{
			"trace_id": span.SpanContext().TraceID().String(),
			"span_id":  span.SpanContext().SpanID().String(),
		})
	}

	return &InstrumentedContext{
		Ctx:    spanCtx,
		Span:   span,
		Logger: logger,
		Timer:  NewTimer(),
	}
}

// End finishes the instrumented operation, recording success or failure.
func (ic *InstrumentedContext) End(err error) {
	if ic.Span != nil {
		if err != nil {
			RecordError(ic.Span, err)
		} else {
			RecordSuccess(ic.Span)
		}
		ic.Span.End()
	}
}

// WithRunContext creates a context enriched with run-specific telemetry.
func WithRunContext(ctx context.Context, runID, user string) context.Context {
	tel := FromTelemetryContext(ctx)
	if tel == nil {
		return ctx
	}

	// Start run span
	spanCtx, span := tel.Tracer.StartRunSpan(ctx, runID)

	// Create run-specific logger
	logger := tel.Logger.WithRunID(runID).WithField("user", user)
	spanCtx = logger.WithContext(spanCtx)

	// Record run started metric
	tel.Metrics.RecordRunStarted(user)

	// Publish run started event
	_ = tel.Events.PublishRunStarted(runID, user)

	// Store the span in context for later retrieval
	spanCtx = context.WithValue(spanCtx, runSpanKey{}, span)

	return spanCtx
}

// runSpanKey is the context key for run spans.
type runSpanKey struct{}

// EndRunContext completes the run context, recording metrics and events.
func EndRunContext(ctx context.Context, runID, status string, err error) {
	tel := FromTelemetryContext(ctx)
	if tel == nil {
		return
	}

	// Get the run span from context
	if span, ok := ctx.Value(runSpanKey{}).(trace.Span); ok {
		if err != nil {
			RecordError(span, err)
		} else {
			RecordSuccess(span)
		}
		span.End()
	}

	// Calculate duration (this is approximate, real duration should come from run metadata)
	timer := NewTimer()
	duration := timer.Duration()

	// Record metrics
	tel.Metrics.RecordRunCompleted(status, duration)

	// Publish events
	if err != nil {
		_ = tel.Events.PublishRunFailed(runID, err.Error())
	} else {
		_ = tel.Events.PublishRunCompleted(runID, status, duration)
	}
}

// WithPlanUnitContext creates a context enriched with plan unit-specific telemetry.
func WithPlanUnitContext(ctx context.Context, runID, planUnitID, resourceID, operation string) context.Context {
	tel := FromTelemetryContext(ctx)
	if tel == nil {
		return ctx
	}

	// Start plan unit span
	spanCtx, span := tel.Tracer.StartPlanUnitSpan(ctx, planUnitID, resourceID, operation)

	// Create plan unit-specific logger
	logger := tel.Logger.
		WithRunID(runID).
		WithPlanUnitID(planUnitID).
		WithResourceID(resourceID).
		WithField("operation", operation)
	spanCtx = logger.WithContext(spanCtx)

	// Publish plan unit started event
	_ = tel.Events.PublishPlanUnitStarted(runID, planUnitID, resourceID, operation)

	// Store the span and timer in context
	spanCtx = context.WithValue(spanCtx, planUnitSpanKey{}, span)
	spanCtx = context.WithValue(spanCtx, planUnitTimerKey{}, NewTimer())

	return spanCtx
}

// planUnitSpanKey is the context key for plan unit spans.
type planUnitSpanKey struct{}

// planUnitTimerKey is the context key for plan unit timers.
type planUnitTimerKey struct{}

// EndPlanUnitContext completes the plan unit context, recording metrics and events.
func EndPlanUnitContext(ctx context.Context, runID, planUnitID, resourceID, operation, status string, err error) {
	tel := FromTelemetryContext(ctx)
	if tel == nil {
		return
	}

	// Get the span from context
	if span, ok := ctx.Value(planUnitSpanKey{}).(trace.Span); ok {
		if err != nil {
			RecordError(span, err)
		} else {
			RecordSuccess(span)
		}
		span.End()
	}

	// Get the timer from context
	var duration time.Duration
	if timer, ok := ctx.Value(planUnitTimerKey{}).(*Timer); ok {
		duration = timer.Duration()
	}

	// Record metrics (assuming resource type from resource ID for now)
	resourceType := "unknown" // This should be passed or derived properly
	tel.Metrics.RecordPlanUnitExecution(operation, status, duration, resourceType)

	// Publish events
	if err != nil {
		_ = tel.Events.PublishPlanUnitFailed(runID, planUnitID, resourceID, err.Error())
	} else {
		_ = tel.Events.PublishPlanUnitCompleted(runID, planUnitID, resourceID, duration)
	}
}

// WithProviderContext creates a context enriched with provider-specific telemetry.
func WithProviderContext(ctx context.Context, providerName, providerVersion string) context.Context {
	tel := FromTelemetryContext(ctx)
	if tel == nil {
		return ctx
	}

	// Create provider-specific logger
	logger := tel.Logger.WithProvider(providerName, providerVersion)
	return logger.WithContext(ctx)
}

// RecordProviderOperation records a provider operation with metrics and tracing.
func RecordProviderOperation(ctx context.Context, providerName, operation string, fn func() error) error {
	tel := FromTelemetryContext(ctx)

	// Start span
	var span trace.Span
	if tel != nil {
		ctx, span = tel.Tracer.StartProviderSpan(ctx, providerName, operation)
		defer span.End()
	}

	// Start timer
	timer := NewTimer()

	// Execute operation
	err := fn()

	// Record metrics
	if tel != nil {
		duration := timer.Duration()
		tel.Metrics.RecordProviderCall(providerName, operation, duration)
		if err != nil {
			tel.Metrics.RecordProviderError(providerName, operation)
			RecordError(span, err)
		} else {
			RecordSuccess(span)
		}
	}

	return err
}
