// Package telemetry provides comprehensive observability instrumentation for OpenFroyo.
//
// The telemetry package integrates structured logging (zerolog), distributed tracing
// (OpenTelemetry), metrics (Prometheus), and event publishing into a unified system
// for monitoring and debugging OpenFroyo operations.
//
// # Architecture
//
// The telemetry system is built on four pillars:
//
//  1. Structured Logging - Context-aware logging with zerolog
//  2. Distributed Tracing - OpenTelemetry traces with multiple exporters
//  3. Metrics Collection - Prometheus metrics for operational insights
//  4. Event Publishing - Async event system for audit and notifications
//
// # Usage
//
// Initialize telemetry at application startup:
//
//	cfg := telemetry.DefaultConfig()
//	cfg.ServiceName = "openfroyo"
//	cfg.ServiceVersion = "1.0.0"
//
//	tel, err := telemetry.NewTelemetry(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer tel.Shutdown(context.Background())
//
//	// Start metrics server
//	if err := tel.StartMetricsServer(); err != nil {
//	    log.Fatal(err)
//	}
//
// Add telemetry to context:
//
//	ctx = tel.WithContext(ctx)
//
// # Structured Logging
//
// The logger provides component-specific logging with automatic context propagation:
//
//	logger := tel.Logger.NewComponentLogger("engine")
//	logger = logger.WithRunID("run-123").WithResourceID("resource-456")
//	logger.Info("Starting resource provisioning")
//	logger.WithError(err).Error("Provisioning failed")
//
// Log levels: trace, debug, info, warn, error, fatal
//
// # Distributed Tracing
//
// Tracing provides visibility into request flow and performance:
//
//	ctx, span := tel.Tracer.Start(ctx, "operation.name")
//	defer span.End()
//
//	// Add attributes
//	span.SetAttributes(
//	    attribute.String("resource.id", resourceID),
//	    attribute.String("operation", "create"),
//	)
//
//	// Record events
//	span.AddEvent("validation.complete")
//
//	// Record errors
//	if err != nil {
//	    telemetry.RecordError(span, err)
//	}
//
// Supported exporters: OTLP (production), Stdout (development), Jaeger (legacy)
//
// # Metrics
//
// Prometheus metrics track system behavior and performance:
//
//	// Record run execution
//	tel.Metrics.RecordRunStarted("user@example.com")
//	tel.Metrics.RecordRunCompleted("succeeded", duration)
//
//	// Record plan unit execution
//	tel.Metrics.RecordPlanUnitExecution("create", "succeeded", duration, "linux.pkg")
//
//	// Record provider calls
//	tel.Metrics.RecordProviderCall("linux.pkg", "apply", duration)
//
//	// Record errors
//	tel.Metrics.RecordError("transient", "TIMEOUT")
//
// Metrics are exposed via HTTP at /metrics (default: :9090/metrics)
//
// # Event Publishing
//
// The event system provides async publishing with buffering and filtering:
//
//	// Publish events
//	tel.Events.PublishRunStarted(runID, user)
//	tel.Events.PublishPlanUnitCompleted(runID, planUnitID, resourceID, duration)
//	tel.Events.PublishDriftDetected(resourceID, driftCount)
//
//	// Subscribe to events
//	tel.Events.Subscribe(func(event telemetry.Event) {
//	    fmt.Printf("Event: %s - %s\n", event.Type, event.Message)
//	}, telemetry.FilterByLevel("warning"))
//
// Event filters: FilterByLevel, FilterByType, FilterByRunID, FilterByResourceID
//
// # Context Helpers
//
// High-level helpers simplify common instrumentation patterns:
//
//	// Instrument an operation
//	ic := telemetry.StartOperation(ctx, "plan.execute",
//	    attribute.String("plan.id", planID))
//	defer ic.End(err)
//
//	ic.Logger.Info("Executing plan")
//
//	// Run context
//	ctx = telemetry.WithRunContext(ctx, runID, user)
//	defer telemetry.EndRunContext(ctx, runID, status, err)
//
//	// Plan unit context
//	ctx = telemetry.WithPlanUnitContext(ctx, runID, planUnitID, resourceID, operation)
//	defer telemetry.EndPlanUnitContext(ctx, runID, planUnitID, resourceID, operation, status, err)
//
//	// Provider operation
//	err := telemetry.RecordProviderOperation(ctx, "linux.pkg", "apply", func() error {
//	    return provider.Apply(ctx, resource)
//	})
//
// # Configuration
//
// The package provides pre-configured setups for different environments:
//
//	// Development (verbose logging, stdout traces, full sampling)
//	cfg := telemetry.DevelopmentConfig()
//
//	// Production (JSON logs, OTLP traces, 10% sampling)
//	cfg := telemetry.ProductionConfig()
//
//	// Custom configuration
//	cfg := &telemetry.Config{
//	    ServiceName: "openfroyo",
//	    ServiceVersion: "1.0.0",
//	    Environment: "staging",
//	    Logging: telemetry.LoggingConfig{
//	        Level: "info",
//	        Format: "json",
//	    },
//	    Tracing: telemetry.TracingConfig{
//	        Enabled: true,
//	        Exporter: "otlp",
//	        Endpoint: "otel-collector:4317",
//	        SamplingRate: 0.1,
//	    },
//	    Metrics: telemetry.MetricsConfig{
//	        Enabled: true,
//	        ListenAddress: ":9090",
//	    },
//	}
//
// # Performance Considerations
//
// The telemetry system is designed for minimal overhead:
//
//  - Structured logging uses zerolog's zero-allocation approach
//  - Tracing uses sampling to reduce data volume in production
//  - Metrics use Prometheus's efficient storage format
//  - Events are buffered and batched to reduce I/O
//  - All operations are non-blocking when possible
//
// Typical overhead: <1% CPU, <10MB memory for moderate workloads
//
// # Graceful Shutdown
//
// Always shut down telemetry gracefully to flush pending data:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	if err := tel.Shutdown(ctx); err != nil {
//	    log.Printf("Telemetry shutdown error: %v", err)
//	}
//
// This ensures:
//  - All buffered events are published
//  - All pending traces are exported
//  - Metrics are finalized
//
// # Integration with OpenFroyo Engine
//
// The engine components automatically integrate with telemetry when available:
//
//  1. Run execution: Automatic run-level tracing and metrics
//  2. Plan units: Per-unit tracing with resource context
//  3. Providers: Provider call tracking and error classification
//  4. Drift detection: Drift events and metrics
//  5. Policy engine: Policy violation events
//
// # Exporters
//
// Tracing supports multiple exporters:
//
//  - "stdout": Print traces to stdout (development)
//  - "otlp": Export via OTLP/gRPC (production, works with collectors)
//  - "jaeger": Direct export to Jaeger (legacy, deprecated)
//  - "none": Generate traces but don't export (testing)
//
// Configure via TracingConfig.Exporter and TracingConfig.Endpoint
//
// # Common Metrics
//
// Key metrics exposed:
//
//  - openfroyo_runs_started_total{user}
//  - openfroyo_runs_completed_total{status}
//  - openfroyo_run_duration_seconds{status}
//  - openfroyo_plan_units_executed_total{operation,status}
//  - openfroyo_plan_unit_duration_seconds{operation,resource_type}
//  - openfroyo_provider_calls_total{provider,operation}
//  - openfroyo_provider_call_duration_seconds{provider,operation}
//  - openfroyo_errors_by_class_total{class}
//  - openfroyo_drift_detections_total{resource_type,status}
//  - openfroyo_active_runs
//
// # Best Practices
//
//  1. Always use context to propagate telemetry
//  2. Use component-specific loggers for clarity
//  3. Add meaningful attributes to spans
//  4. Record both success and failure metrics
//  5. Use appropriate log levels
//  6. Filter events to avoid overwhelming subscribers
//  7. Monitor telemetry overhead in production
//  8. Configure sampling for high-volume systems
//  9. Always call defer span.End() after starting a span
//  10. Shut down gracefully to avoid data loss
//
// # Security Considerations
//
//  - Never log sensitive data (credentials, keys, tokens)
//  - Sanitize resource IDs if they contain PII
//  - Use secure connections (TLS) for trace exporters in production
//  - Limit metrics endpoint access via network policies
//  - Consider event data before adding to audit logs
//
package telemetry
