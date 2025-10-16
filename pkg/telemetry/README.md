# OpenFroyo Telemetry Package

Comprehensive observability infrastructure for the OpenFroyo orchestration engine.

## Overview

The telemetry package provides a unified observability system combining:

- **Structured Logging** - Context-aware logging with zerolog
- **Distributed Tracing** - OpenTelemetry traces with OTLP, Jaeger, and stdout exporters
- **Metrics Collection** - Prometheus metrics for operational insights
- **Event Publishing** - Async event system with buffering and filtering

## Quick Start

```go
import "github.com/openfroyo/openfroyo/pkg/telemetry"

// Initialize telemetry
cfg := telemetry.DefaultConfig()
cfg.ServiceName = "openfroyo"
cfg.ServiceVersion = "1.0.0"

tel, err := telemetry.NewTelemetry(cfg)
if err != nil {
    log.Fatal(err)
}
defer tel.Shutdown(context.Background())

// Start metrics server
if err := tel.StartMetricsServer(); err != nil {
    log.Fatal(err)
}

// Add to context
ctx := tel.WithContext(context.Background())
```

## Architecture

### Components

```
┌─────────────────────────────────────────────────────────┐
│                    Telemetry System                     │
├─────────────┬──────────────┬──────────────┬────────────┤
│   Logger    │   Tracer     │   Metrics    │   Events   │
│  (zerolog)  │ (OpenTelemetry)│ (Prometheus) │  (Async)   │
└─────────────┴──────────────┴──────────────┴────────────┘
```

### Data Flow

1. **Context Propagation** - Telemetry attached to context
2. **Instrumentation** - Operations create spans, log entries, metrics
3. **Collection** - Data collected via exporters/endpoints
4. **Export** - Data sent to backends (OTLP, Prometheus, subscribers)

## Files

| File | Purpose |
|------|---------|
| `config.go` | Configuration structures and validation |
| `logger.go` | Structured logging with zerolog |
| `tracer.go` | OpenTelemetry tracing implementation |
| `metrics.go` | Prometheus metrics collection |
| `events.go` | Event publishing system with buffering |
| `context.go` | Context helpers and instrumentation patterns |
| `doc.go` | Package documentation |
| `example_test.go` | Comprehensive usage examples |

## Structured Logging

### Features

- Multiple log levels (trace, debug, info, warn, error, fatal)
- Component-specific loggers
- Context-aware logging with automatic field propagation
- Caller information (file:line)
- Sampling for high-frequency logs
- Multiple output formats (console, JSON)
- Custom hooks support

### Usage

```go
// Create component logger
logger := tel.Logger.NewComponentLogger("engine")

// Add context fields
logger = logger.WithRunID("run-123").WithResourceID("resource-456")

// Log at different levels
logger.Info("Starting operation")
logger.WithError(err).Error("Operation failed")
```

## Distributed Tracing

### Features

- OpenTelemetry standard compliance
- Multiple exporters (OTLP, stdout, Jaeger)
- Configurable sampling (0-100%)
- Automatic context propagation
- Span attributes and events
- Error recording with status codes

### Exporters

| Exporter | Use Case | Configuration |
|----------|----------|--------------|
| `otlp` | Production (OpenTelemetry Collector) | Endpoint, Headers, TLS |
| `stdout` | Development/Debugging | Pretty-print to stdout |
| `jaeger` | Legacy Jaeger deployments | Deprecated |
| `none` | Testing (generate but don't export) | - |

### Usage

```go
// Start a span
ctx, span := tel.Tracer.Start(ctx, "operation.name")
defer span.End()

// Add attributes
span.SetAttributes(
    attribute.String("resource.id", resourceID),
    attribute.String("operation", "create"),
)

// Record error
if err != nil {
    telemetry.RecordError(span, err)
}
```

## Metrics

### Core Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `openfroyo_runs_started_total` | Counter | `user` | Total runs started |
| `openfroyo_runs_completed_total` | Counter | `status` | Total runs completed |
| `openfroyo_run_duration_seconds` | Histogram | `status` | Run execution time |
| `openfroyo_plan_units_executed_total` | Counter | `operation`, `status` | Plan units executed |
| `openfroyo_plan_unit_duration_seconds` | Histogram | `operation`, `resource_type` | Plan unit execution time |
| `openfroyo_provider_calls_total` | Counter | `provider`, `operation` | Provider calls |
| `openfroyo_provider_call_duration_seconds` | Histogram | `provider`, `operation` | Provider call duration |
| `openfroyo_errors_by_class_total` | Counter | `class` | Errors by classification |
| `openfroyo_drift_detections_total` | Counter | `resource_type`, `status` | Drift detections |
| `openfroyo_active_runs` | Gauge | - | Currently active runs |

### Usage

```go
// Record run
tel.Metrics.RecordRunStarted("user@example.com")
tel.Metrics.RecordRunCompleted("succeeded", duration)

// Record plan unit
tel.Metrics.RecordPlanUnitExecution("create", "succeeded", duration, "linux.pkg")

// Record provider call
tel.Metrics.RecordProviderCall("linux.pkg", "apply", duration)

// Record error
tel.Metrics.RecordError("transient", "TIMEOUT")
```

### Metrics Endpoint

Metrics are exposed via HTTP at:
- Default: `http://localhost:9090/metrics`
- Configurable via `MetricsConfig.ListenAddress` and `MetricsConfig.Path`

## Event System

### Event Types

- `run.started`, `run.completed`, `run.failed`
- `plan_unit.started`, `plan_unit.completed`, `plan_unit.failed`
- `resource.state_changed`
- `drift.detected`
- `policy.violation`
- `provider.invoked`

### Features

- Asynchronous publishing with buffering
- Multiple subscribers support
- Event filtering (by level, type, run, resource)
- Batching and periodic flush
- Graceful shutdown with pending event flush

### Usage

```go
// Subscribe to events
tel.Events.Subscribe(func(event telemetry.Event) {
    fmt.Printf("Event: %s - %s\n", event.Type, event.Message)
}, telemetry.FilterByLevel("warning"))

// Publish events
tel.Events.PublishRunStarted(runID, user)
tel.Events.PublishDriftDetected(resourceID, driftCount)
```

## Context Helpers

High-level helpers for common instrumentation patterns:

```go
// Instrument any operation
ic := telemetry.StartOperation(ctx, "plan.execute")
defer ic.End(err)
ic.Logger.Info("Executing plan")

// Run context (automatic metrics, events, tracing)
ctx = telemetry.WithRunContext(ctx, runID, user)
defer telemetry.EndRunContext(ctx, runID, status, err)

// Plan unit context
ctx = telemetry.WithPlanUnitContext(ctx, runID, planUnitID, resourceID, operation)
defer telemetry.EndPlanUnitContext(ctx, runID, planUnitID, resourceID, operation, status, err)

// Provider operation
err := telemetry.RecordProviderOperation(ctx, "linux.pkg", "apply", func() error {
    return provider.Apply(ctx, resource)
})
```

## Configuration

### Pre-configured Environments

```go
// Development (verbose logging, full sampling)
cfg := telemetry.DevelopmentConfig()

// Production (JSON logs, OTLP, 10% sampling)
cfg := telemetry.ProductionConfig()

// Custom
cfg := telemetry.DefaultConfig()
cfg.Tracing.Exporter = "otlp"
cfg.Tracing.Endpoint = "otel-collector:4317"
cfg.Tracing.SamplingRate = 0.1
```

### Configuration Options

| Section | Key | Description | Default |
|---------|-----|-------------|---------|
| **Logging** | Level | Minimum log level | `info` |
| | Format | Log format (console/json) | `console` |
| | EnableCaller | Add file:line info | `true` |
| | EnableSampling | Sample high-frequency logs | `false` |
| **Tracing** | Enabled | Enable tracing | `true` |
| | Exporter | Trace exporter type | `stdout` |
| | SamplingRate | Trace sampling (0-1) | `1.0` |
| | Endpoint | Exporter endpoint | - |
| **Metrics** | Enabled | Enable metrics | `true` |
| | ListenAddress | Metrics HTTP address | `:9090` |
| | Path | Metrics HTTP path | `/metrics` |
| **Events** | Enabled | Enable events | `true` |
| | BufferSize | Event buffer size | `1000` |
| | FlushInterval | Auto-flush interval | `5s` |

## Performance

### Overhead

Typical overhead for the telemetry system:
- **CPU**: <1% for moderate workloads
- **Memory**: <10MB baseline + ~100 bytes per active span
- **Latency**: <100μs per operation instrumentation

### Optimization Tips

1. **Sampling**: Use sampling in production (10-20% is typical)
2. **Buffering**: Increase buffer sizes for high-throughput systems
3. **Filtering**: Filter events by level to reduce processing
4. **Batching**: Events are automatically batched for efficiency
5. **Async**: Enable async mode for events to avoid blocking

## Integration Guide

### 1. Initialize at Startup

```go
func main() {
    cfg := telemetry.ProductionConfig()
    cfg.ServiceName = "openfroyo"
    cfg.ServiceVersion = version.Version

    tel, err := telemetry.NewTelemetry(cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer tel.Shutdown(context.Background())

    if err := tel.StartMetricsServer(); err != nil {
        log.Fatal(err)
    }

    // ... rest of application
}
```

### 2. Add to Context

```go
func handleRequest(ctx context.Context, req Request) error {
    ctx = tel.WithContext(ctx)

    // Telemetry is now available via context
    logger := telemetry.FromContext(ctx)
    logger.Info("Processing request")

    return processRequest(ctx, req)
}
```

### 3. Instrument Operations

```go
func executeRun(ctx context.Context, plan Plan) error {
    // Create run context (auto-instrumentation)
    runID := uuid.New().String()
    ctx = telemetry.WithRunContext(ctx, runID, plan.User)
    defer telemetry.EndRunContext(ctx, runID, "succeeded", nil)

    // Execute plan units
    for _, unit := range plan.Units {
        if err := executePlanUnit(ctx, runID, unit); err != nil {
            return err
        }
    }

    return nil
}

func executePlanUnit(ctx context.Context, runID string, unit PlanUnit) error {
    ctx = telemetry.WithPlanUnitContext(ctx, runID, unit.ID, unit.ResourceID, unit.Operation)
    defer telemetry.EndPlanUnitContext(ctx, runID, unit.ID, unit.ResourceID, unit.Operation, "succeeded", nil)

    logger := telemetry.FromContext(ctx)
    logger.Info("Executing plan unit")

    // ... execute unit

    return nil
}
```

### 4. Instrument Providers

```go
func (p *Provider) Apply(ctx context.Context, resource Resource) error {
    return telemetry.RecordProviderOperation(ctx, p.Name, "apply", func() error {
        logger := telemetry.FromContext(ctx)
        logger.Info("Applying resource configuration")

        // ... actual provider logic

        return nil
    })
}
```

## Graceful Shutdown

Always shut down gracefully to flush pending data:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err := tel.Shutdown(ctx); err != nil {
    log.Printf("Telemetry shutdown error: %v", err)
}
```

## Best Practices

1. **Always use context** - Propagate telemetry via context
2. **Component loggers** - Create component-specific loggers for clarity
3. **Meaningful attributes** - Add relevant attributes to spans
4. **Record both paths** - Track both success and failure
5. **Appropriate levels** - Use correct log levels (debug for verbose, error for failures)
6. **Filter events** - Avoid overwhelming subscribers with too many events
7. **Monitor overhead** - Track telemetry system's own performance
8. **Sample in production** - Use sampling to reduce data volume
9. **Close spans** - Always defer span.End()
10. **Graceful shutdown** - Flush data before exit

## Security Considerations

- Never log sensitive data (credentials, tokens, keys)
- Sanitize resource IDs if they contain PII
- Use TLS for trace exporters in production
- Limit metrics endpoint access via network policies
- Review event data before adding to audit logs
- Set appropriate RBAC for telemetry backends

## Troubleshooting

### Traces not appearing

1. Check exporter configuration (`TracingConfig.Exporter`, `TracingConfig.Endpoint`)
2. Verify sampling rate (`TracingConfig.SamplingRate`)
3. Ensure spans are properly ended (`defer span.End()`)
4. Check backend connectivity and logs

### High memory usage

1. Reduce sampling rate
2. Decrease event buffer size
3. Enable log sampling
4. Check for span leaks (spans not ended)

### Missing metrics

1. Verify metrics are enabled (`MetricsConfig.Enabled`)
2. Check metrics endpoint is accessible
3. Ensure metrics server is started
4. Verify Prometheus scrape configuration

## Examples

See `example_test.go` for comprehensive examples including:
- Basic setup and initialization
- Structured logging patterns
- Distributed tracing usage
- Metrics collection
- Event publishing and filtering
- Run and plan unit instrumentation
- Provider operation recording
- Production configuration

## Dependencies

- `github.com/rs/zerolog` - Structured logging
- `go.opentelemetry.io/otel` - Distributed tracing
- `github.com/prometheus/client_golang` - Metrics
- `github.com/google/uuid` - Event IDs

## License

Part of the OpenFroyo project.
