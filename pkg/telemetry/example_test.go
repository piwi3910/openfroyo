package telemetry_test

import (
	"context"
	"fmt"
	"time"

	"github.com/openfroyo/openfroyo/pkg/telemetry"
	"go.opentelemetry.io/otel/attribute"
)

// Example_basicSetup demonstrates basic telemetry setup.
func Example_basicSetup() {
	// Create configuration
	cfg := telemetry.DefaultConfig()
	cfg.ServiceName = "openfroyo"
	cfg.ServiceVersion = "1.0.0"

	// Initialize telemetry
	tel, err := telemetry.NewTelemetry(cfg)
	if err != nil {
		panic(err)
	}
	defer tel.Shutdown(context.Background())

	// Start metrics server (non-blocking)
	if err := tel.StartMetricsServer(); err != nil {
		panic(err)
	}

	// Add telemetry to context
	ctx := tel.WithContext(context.Background())

	// Use telemetry
	logger := telemetry.FromContext(ctx)
	logger.Info("Application started")

	// Output can vary, so we don't specify output for this example
}

// Example_structuredLogging demonstrates structured logging features.
func Example_structuredLogging() {
	cfg := telemetry.DevelopmentConfig()
	cfg.Logging.Output = "stdout"

	tel, _ := telemetry.NewTelemetry(cfg)
	defer tel.Shutdown(context.Background())

	// Component-specific logger
	logger := tel.Logger.NewComponentLogger("engine")

	// Add context fields
	logger = logger.WithFields(map[string]interface{}{
		"run_id":      "run-123",
		"resource_id": "resource-456",
	})

	// Log at different levels
	logger.Debug("Starting resource provisioning")
	logger.Info("Resource created successfully")
	logger.Warn("Resource configuration drift detected")

	// Log with error
	err := fmt.Errorf("network timeout")
	logger.WithError(err).Error("Failed to connect to remote host")

	// Output varies, no output specified
}

// Example_distributedTracing demonstrates distributed tracing usage.
func Example_distributedTracing() {
	cfg := telemetry.DevelopmentConfig()
	cfg.Tracing.Exporter = "stdout"

	tel, _ := telemetry.NewTelemetry(cfg)
	defer tel.Shutdown(context.Background())

	ctx := tel.WithContext(context.Background())

	// Start a span
	ctx, span := tel.Tracer.Start(ctx, "execute_plan")
	defer span.End()

	// Add attributes
	span.SetAttributes(
		attribute.String("plan.id", "plan-789"),
		attribute.Int("plan.units", 5),
	)

	// Add event
	span.AddEvent("validation.complete")

	// Nested span
	ctx, childSpan := tel.Tracer.Start(ctx, "apply_resource")
	defer childSpan.End()

	childSpan.SetAttributes(
		attribute.String("resource.id", "resource-456"),
		attribute.String("operation", "create"),
	)

	// Simulate work
	time.Sleep(10 * time.Millisecond)

	// Record success
	telemetry.RecordSuccess(childSpan)

	// Output varies, no output specified
}

// Example_metricsCollection demonstrates metrics collection.
func Example_metricsCollection() {
	cfg := telemetry.DefaultConfig()
	cfg.Metrics.Enabled = true

	tel, _ := telemetry.NewTelemetry(cfg)
	defer tel.Shutdown(context.Background())

	// Record run metrics
	tel.Metrics.RecordRunStarted("user@example.com")

	// Simulate run execution
	start := time.Now()
	time.Sleep(50 * time.Millisecond)
	duration := time.Since(start)

	tel.Metrics.RecordRunCompleted("succeeded", duration)

	// Record plan unit metrics
	tel.Metrics.RecordPlanUnitExecution(
		"create",          // operation
		"succeeded",       // status
		25*time.Millisecond, // duration
		"linux.pkg",       // resource type
	)

	// Record provider metrics
	tel.Metrics.RecordProviderCall("linux.pkg", "apply", 15*time.Millisecond)

	// Record error metrics
	tel.Metrics.RecordError("transient", "TIMEOUT")

	// Set resource counts
	tel.Metrics.SetResourceCount("linux.pkg", "ready", 10)
	tel.Metrics.SetResourceCount("linux.service", "ready", 5)

	fmt.Println("Metrics recorded successfully")
	// Output: Metrics recorded successfully
}

// Example_eventPublishing demonstrates event publishing and subscription.
func Example_eventPublishing() {
	cfg := telemetry.DefaultConfig()
	cfg.Events.Enabled = true
	cfg.Events.EnableAsync = false // Synchronous for example

	tel, _ := telemetry.NewTelemetry(cfg)
	defer tel.Shutdown(context.Background())

	// Subscribe to events
	tel.Events.Subscribe(func(event telemetry.Event) {
		fmt.Printf("Event: %s - %s\n", event.Type, event.Message)
	}, nil) // No filter, receive all events

	// Publish events
	tel.Events.PublishRunStarted("run-123", "user@example.com")
	tel.Events.PublishPlanUnitStarted("run-123", "pu-1", "resource-456", "create")
	tel.Events.PublishPlanUnitCompleted("run-123", "pu-1", "resource-456", 25*time.Millisecond)

	// Output varies due to async nature, no output specified
}

// Example_runInstrumentation demonstrates instrumenting a complete run.
func Example_runInstrumentation() {
	cfg := telemetry.DevelopmentConfig()
	tel, _ := telemetry.NewTelemetry(cfg)
	defer tel.Shutdown(context.Background())

	ctx := tel.WithContext(context.Background())

	// Start run context
	runID := "run-123"
	user := "admin@example.com"
	ctx = telemetry.WithRunContext(ctx, runID, user)

	// Execute run (simulated)
	executeRun(ctx, runID)

	// End run context
	telemetry.EndRunContext(ctx, runID, "succeeded", nil)

	fmt.Println("Run instrumentation complete")
	// Output: Run instrumentation complete
}

func executeRun(ctx context.Context, runID string) {
	// Simulate plan unit execution
	planUnitID := "pu-1"
	resourceID := "resource-456"
	operation := "create"

	ctx = telemetry.WithPlanUnitContext(ctx, runID, planUnitID, resourceID, operation)

	// Get logger from context
	logger := telemetry.FromContext(ctx)
	logger.Info("Executing plan unit")

	// Simulate work
	time.Sleep(10 * time.Millisecond)

	// End plan unit context
	telemetry.EndPlanUnitContext(ctx, runID, planUnitID, resourceID, operation, "succeeded", nil)
}

// Example_providerInstrumentation demonstrates instrumenting provider calls.
func Example_providerInstrumentation() {
	cfg := telemetry.DevelopmentConfig()
	tel, _ := telemetry.NewTelemetry(cfg)
	defer tel.Shutdown(context.Background())

	ctx := tel.WithContext(context.Background())

	// Add provider context
	ctx = telemetry.WithProviderContext(ctx, "linux.pkg", "1.0.0")

	// Record provider operation
	err := telemetry.RecordProviderOperation(ctx, "linux.pkg", "apply", func() error {
		// Simulate provider work
		time.Sleep(15 * time.Millisecond)
		return nil
	})

	if err == nil {
		fmt.Println("Provider operation completed successfully")
	}

	// Output: Provider operation completed successfully
}

// Example_instrumentedOperation demonstrates using the InstrumentedContext helper.
func Example_instrumentedOperation() {
	cfg := telemetry.DevelopmentConfig()
	tel, _ := telemetry.NewTelemetry(cfg)
	defer tel.Shutdown(context.Background())

	ctx := tel.WithContext(context.Background())

	// Start instrumented operation
	ic := telemetry.StartOperation(ctx, "validate_config",
		attribute.String("config.path", "/etc/openfroyo/config.cue"),
	)
	defer ic.End(nil)

	// Use the instrumented context
	ic.Logger.Info("Validating configuration")

	// Simulate validation
	time.Sleep(5 * time.Millisecond)

	ic.Logger.Debug("Configuration validation complete")

	fmt.Println("Operation instrumentation complete")
	// Output: Operation instrumentation complete
}

// Example_eventFiltering demonstrates event filtering.
func Example_eventFiltering() {
	cfg := telemetry.DefaultConfig()
	cfg.Events.Enabled = true
	cfg.Events.EnableAsync = false

	tel, _ := telemetry.NewTelemetry(cfg)
	defer tel.Shutdown(context.Background())

	// Subscribe with level filter (only warnings and errors)
	tel.Events.Subscribe(func(event telemetry.Event) {
		fmt.Printf("Important event: %s\n", event.Type)
	}, telemetry.FilterByLevel(telemetry.EventLevelWarning))

	// Subscribe with type filter (only drift events)
	tel.Events.Subscribe(func(event telemetry.Event) {
		fmt.Printf("Drift event: %s\n", event.Message)
	}, telemetry.FilterByType("drift.detected"))

	// Publish various events
	tel.Events.PublishRunStarted("run-123", "user") // Info - filtered by level filter
	tel.Events.PublishDriftDetected("resource-1", 3) // Warning - passes level filter
	tel.Events.PublishRunFailed("run-123", "error") // Error - passes level filter

	// Output varies, no output specified
}

// Example_productionConfiguration demonstrates production-ready configuration.
func Example_productionConfiguration() {
	cfg := telemetry.ProductionConfig()

	// Customize for your environment
	cfg.ServiceName = "openfroyo"
	cfg.ServiceVersion = "1.2.3"
	cfg.Environment = "production"

	// Configure OTLP exporter
	cfg.Tracing.Exporter = "otlp"
	cfg.Tracing.Endpoint = "otel-collector.monitoring.svc.cluster.local:4317"
	cfg.Tracing.SamplingRate = 0.1 // 10% sampling
	cfg.Tracing.Insecure = false   // Use TLS in production

	// Configure metrics
	cfg.Metrics.ListenAddress = ":9090"
	cfg.Metrics.Namespace = "openfroyo"

	// Configure events
	cfg.Events.BufferSize = 10000
	cfg.Events.FlushInterval = 5 * time.Second

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		panic(err)
	}

	fmt.Println("Production configuration validated")
	// Output: Production configuration validated
}

// Example_errorRecording demonstrates error recording with proper classification.
func Example_errorRecording() {
	cfg := telemetry.DevelopmentConfig()
	tel, _ := telemetry.NewTelemetry(cfg)
	defer tel.Shutdown(context.Background())

	ctx := tel.WithContext(context.Background())

	// Start a span
	ctx, span := tel.Tracer.Start(ctx, "risky_operation")
	defer span.End()

	// Simulate an error
	err := fmt.Errorf("connection timeout")

	if err != nil {
		// Record error on span
		telemetry.RecordError(span, err)

		// Record error metric with classification
		tel.Metrics.RecordError("transient", "TIMEOUT")

		// Log error
		logger := telemetry.FromContext(ctx)
		logger.WithError(err).Error("Operation failed")
	}

	fmt.Println("Error recording complete")
	// Output: Error recording complete
}

// Example_multipleComponents demonstrates telemetry in a multi-component system.
func Example_multipleComponents() {
	cfg := telemetry.DevelopmentConfig()
	tel, _ := telemetry.NewTelemetry(cfg)
	defer tel.Shutdown(context.Background())

	// Component-specific loggers
	engineLogger := tel.Logger.NewComponentLogger("engine")
	plannerLogger := tel.Logger.NewComponentLogger("planner")
	providerLogger := tel.Logger.NewComponentLogger("provider")

	engineLogger.Info("Engine initialized")
	plannerLogger.Info("Building execution plan")
	providerLogger.Info("Loading provider plugins")

	fmt.Println("Multi-component logging complete")
	// Output: Multi-component logging complete
}
