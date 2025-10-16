package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Tracer wraps the OpenTelemetry tracer with OpenFroyo-specific functionality.
type Tracer struct {
	provider *sdktrace.TracerProvider
	tracer   trace.Tracer
	config   TracingConfig
}

// NewTracer creates a new tracer with the given configuration.
func NewTracer(cfg TracingConfig, serviceName, serviceVersion, environment string) (*Tracer, error) {
	if !cfg.Enabled {
		// Return a tracer with no-op provider
		return &Tracer{
			provider: sdktrace.NewTracerProvider(),
			tracer:   otel.Tracer(serviceName),
			config:   cfg,
		}, nil
	}

	// Create resource with service information
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(serviceVersion),
			attribute.String("environment", environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace resource: %w", err)
	}

	// Create exporter based on configuration
	var exporter sdktrace.SpanExporter
	switch cfg.Exporter {
	case "otlp":
		exporter, err = createOTLPExporter(cfg)
	case "stdout":
		exporter, err = createStdoutExporter(cfg)
	case "none":
		// No exporter - traces are generated but not exported
		exporter = nil
	default:
		return nil, fmt.Errorf("unsupported trace exporter: %s", cfg.Exporter)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Configure sampler
	sampler := sdktrace.ParentBased(
		sdktrace.TraceIDRatioBased(cfg.SamplingRate),
	)

	// Create trace provider
	opts := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	}

	if exporter != nil {
		opts = append(opts, sdktrace.WithBatcher(
			exporter,
			sdktrace.WithMaxExportBatchSize(cfg.MaxExportBatchSize),
			sdktrace.WithExportTimeout(cfg.ExportTimeout),
		))
	}

	provider := sdktrace.NewTracerProvider(opts...)

	// Set global trace provider
	otel.SetTracerProvider(provider)

	// Set global propagator for context propagation
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	return &Tracer{
		provider: provider,
		tracer:   provider.Tracer(serviceName),
		config:   cfg,
	}, nil
}

// createOTLPExporter creates an OTLP gRPC exporter.
func createOTLPExporter(cfg TracingConfig) (sdktrace.SpanExporter, error) {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.Endpoint),
	}

	if cfg.Insecure {
		opts = append(opts, otlptracegrpc.WithTLSCredentials(insecure.NewCredentials()))
	}

	// Add custom headers if provided
	if len(cfg.Headers) > 0 {
		opts = append(opts, otlptracegrpc.WithHeaders(cfg.Headers))
	}

	// Add dial options for connection timeout
	opts = append(opts, otlptracegrpc.WithDialOption(
		grpc.WithBlock(),
	))

	return otlptracegrpc.New(context.Background(), opts...)
}

// createStdoutExporter creates a stdout exporter for debugging.
func createStdoutExporter(cfg TracingConfig) (sdktrace.SpanExporter, error) {
	return stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
	)
}

// Start begins a new span with the given name.
func (t *Tracer) Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, spanName, opts...)
}

// StartSpan is a convenience method that starts a span with common attributes.
func (t *Tracer) StartSpan(ctx context.Context, operation string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, operation, trace.WithAttributes(attrs...))
}

// StartRunSpan starts a span for a run execution.
func (t *Tracer) StartRunSpan(ctx context.Context, runID string) (context.Context, trace.Span) {
	return t.StartSpan(ctx, "run.execute",
		attribute.String("run.id", runID),
		attribute.String("span.kind", "run"),
	)
}

// StartPlanUnitSpan starts a span for a plan unit execution.
func (t *Tracer) StartPlanUnitSpan(ctx context.Context, planUnitID, resourceID, operation string) (context.Context, trace.Span) {
	return t.StartSpan(ctx, "plan_unit.execute",
		attribute.String("plan_unit.id", planUnitID),
		attribute.String("resource.id", resourceID),
		attribute.String("operation", operation),
		attribute.String("span.kind", "plan_unit"),
	)
}

// StartProviderSpan starts a span for a provider operation.
func (t *Tracer) StartProviderSpan(ctx context.Context, providerName, operation string) (context.Context, trace.Span) {
	return t.StartSpan(ctx, fmt.Sprintf("provider.%s", operation),
		attribute.String("provider.name", providerName),
		attribute.String("provider.operation", operation),
		attribute.String("span.kind", "provider"),
	)
}

// RecordError records an error on the current span.
func RecordError(span trace.Span, err error) {
	if err == nil {
		return
	}
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// RecordSuccess marks the span as successful.
func RecordSuccess(span trace.Span) {
	span.SetStatus(codes.Ok, "")
}

// SetAttributes sets multiple attributes on a span.
func SetAttributes(span trace.Span, attrs ...attribute.KeyValue) {
	span.SetAttributes(attrs...)
}

// AddEvent adds an event to the span.
func AddEvent(span trace.Span, name string, attrs ...attribute.KeyValue) {
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// AddRunEvent adds a run-related event to the span.
func AddRunEvent(span trace.Span, eventType, message string) {
	span.AddEvent(eventType, trace.WithAttributes(
		attribute.String("event.message", message),
		attribute.String("event.category", "run"),
	))
}

// AddResourceEvent adds a resource-related event to the span.
func AddResourceEvent(span trace.Span, resourceID, eventType, message string) {
	span.AddEvent(eventType, trace.WithAttributes(
		attribute.String("resource.id", resourceID),
		attribute.String("event.message", message),
		attribute.String("event.category", "resource"),
	))
}

// Shutdown gracefully shuts down the tracer, flushing any pending spans.
func (t *Tracer) Shutdown(ctx context.Context) error {
	if t.provider == nil {
		return nil
	}
	return t.provider.Shutdown(ctx)
}

// ForceFlush forces all pending spans to be exported immediately.
func (t *Tracer) ForceFlush(ctx context.Context) error {
	if t.provider == nil {
		return nil
	}
	return t.provider.ForceFlush(ctx)
}

// SpanFromContext returns the current span from the context.
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// TraceID returns the trace ID of the current span in the context.
func TraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return ""
	}
	return span.SpanContext().TraceID().String()
}

// SpanID returns the span ID of the current span in the context.
func SpanID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return ""
	}
	return span.SpanContext().SpanID().String()
}

// Common attribute keys for OpenFroyo tracing.
var (
	// Run attributes
	AttrRunID        = attribute.Key("run.id")
	AttrRunStatus    = attribute.Key("run.status")
	AttrPlanID       = attribute.Key("plan.id")

	// Plan unit attributes
	AttrPlanUnitID   = attribute.Key("plan_unit.id")
	AttrResourceID   = attribute.Key("resource.id")
	AttrResourceType = attribute.Key("resource.type")
	AttrOperation    = attribute.Key("operation")
	AttrOperationType = attribute.Key("operation.type")

	// Provider attributes
	AttrProviderName    = attribute.Key("provider.name")
	AttrProviderVersion = attribute.Key("provider.version")
	AttrProviderOp      = attribute.Key("provider.operation")

	// Error attributes
	AttrErrorClass   = attribute.Key("error.class")
	AttrErrorCode    = attribute.Key("error.code")
	AttrErrorMessage = attribute.Key("error.message")

	// Metadata attributes
	AttrTargetHost = attribute.Key("target.host")
	AttrTargetID   = attribute.Key("target.id")
)
