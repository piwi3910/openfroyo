package telemetry

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics provides Prometheus metrics for OpenFroyo.
type Metrics struct {
	config MetricsConfig

	// Run metrics
	runsStarted   *prometheus.CounterVec
	runsCompleted *prometheus.CounterVec
	runDuration   *prometheus.HistogramVec

	// Plan unit metrics
	planUnitsExecuted *prometheus.CounterVec
	planUnitDuration  *prometheus.HistogramVec

	// Resource metrics
	resourcesManaged *prometheus.GaugeVec
	resourceState    *prometheus.GaugeVec

	// Provider metrics
	providerCalls    *prometheus.CounterVec
	providerDuration *prometheus.HistogramVec
	providerErrors   *prometheus.CounterVec

	// Error metrics
	errorsByClass *prometheus.CounterVec
	errorsByCode  *prometheus.CounterVec

	// Drift detection metrics
	driftDetections *prometheus.CounterVec

	// System metrics
	activeRuns    prometheus.Gauge
	queuedPlanUnits prometheus.Gauge

	registry *prometheus.Registry
}

// NewMetrics creates a new metrics collector with the given configuration.
func NewMetrics(cfg MetricsConfig) (*Metrics, error) {
	if !cfg.Enabled {
		// Return a no-op metrics instance
		return &Metrics{config: cfg}, nil
	}

	namespace := cfg.Namespace
	buckets := cfg.DefaultHistogramBuckets
	if len(buckets) == 0 {
		buckets = prometheus.DefBuckets
	}

	// Create a new registry
	registry := prometheus.NewRegistry()

	m := &Metrics{
		config:   cfg,
		registry: registry,

		// Run metrics
		runsStarted: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "runs_started_total",
				Help:      "Total number of runs started",
			},
			[]string{"user"},
		),
		runsCompleted: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "runs_completed_total",
				Help:      "Total number of runs completed",
			},
			[]string{"status"},
		),
		runDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "run_duration_seconds",
				Help:      "Duration of run execution in seconds",
				Buckets:   buckets,
			},
			[]string{"status"},
		),

		// Plan unit metrics
		planUnitsExecuted: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "plan_units_executed_total",
				Help:      "Total number of plan units executed",
			},
			[]string{"operation", "status"},
		),
		planUnitDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "plan_unit_duration_seconds",
				Help:      "Duration of plan unit execution in seconds",
				Buckets:   buckets,
			},
			[]string{"operation", "resource_type"},
		),

		// Resource metrics
		resourcesManaged: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "resources_managed",
				Help:      "Current number of managed resources",
			},
			[]string{"type", "status"},
		),
		resourceState: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "resource_state",
				Help:      "Current state of resources (1=ready, 0=not ready)",
			},
			[]string{"resource_id", "type"},
		),

		// Provider metrics
		providerCalls: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "provider_calls_total",
				Help:      "Total number of provider calls",
			},
			[]string{"provider", "operation"},
		),
		providerDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "provider_call_duration_seconds",
				Help:      "Duration of provider calls in seconds",
				Buckets:   buckets,
			},
			[]string{"provider", "operation"},
		),
		providerErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "provider_errors_total",
				Help:      "Total number of provider errors",
			},
			[]string{"provider", "operation"},
		),

		// Error metrics
		errorsByClass: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "errors_by_class_total",
				Help:      "Total number of errors by error class",
			},
			[]string{"class"},
		),
		errorsByCode: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "errors_by_code_total",
				Help:      "Total number of errors by error code",
			},
			[]string{"code"},
		),

		// Drift detection metrics
		driftDetections: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "drift_detections_total",
				Help:      "Total number of drift detections",
			},
			[]string{"resource_type", "status"},
		),

		// System metrics
		activeRuns: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "active_runs",
				Help:      "Current number of active runs",
			},
		),
		queuedPlanUnits: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "queued_plan_units",
				Help:      "Current number of queued plan units",
			},
		),
	}

	// Register all metrics
	registry.MustRegister(
		m.runsStarted,
		m.runsCompleted,
		m.runDuration,
		m.planUnitsExecuted,
		m.planUnitDuration,
		m.resourcesManaged,
		m.resourceState,
		m.providerCalls,
		m.providerDuration,
		m.providerErrors,
		m.errorsByClass,
		m.errorsByCode,
		m.driftDetections,
		m.activeRuns,
		m.queuedPlanUnits,
	)

	return m, nil
}

// Run Metrics

// RecordRunStarted increments the counter for started runs.
func (m *Metrics) RecordRunStarted(user string) {
	if m.runsStarted == nil {
		return
	}
	m.runsStarted.WithLabelValues(user).Inc()
	m.activeRuns.Inc()
}

// RecordRunCompleted records a completed run with its status and duration.
func (m *Metrics) RecordRunCompleted(status string, duration time.Duration) {
	if m.runsCompleted == nil {
		return
	}
	m.runsCompleted.WithLabelValues(status).Inc()
	m.runDuration.WithLabelValues(status).Observe(duration.Seconds())
	m.activeRuns.Dec()
}

// Plan Unit Metrics

// RecordPlanUnitExecution records the execution of a plan unit.
func (m *Metrics) RecordPlanUnitExecution(operation, status string, duration time.Duration, resourceType string) {
	if m.planUnitsExecuted == nil {
		return
	}
	m.planUnitsExecuted.WithLabelValues(operation, status).Inc()
	m.planUnitDuration.WithLabelValues(operation, resourceType).Observe(duration.Seconds())
}

// Resource Metrics

// SetResourceCount sets the current count of managed resources.
func (m *Metrics) SetResourceCount(resourceType, status string, count float64) {
	if m.resourcesManaged == nil {
		return
	}
	m.resourcesManaged.WithLabelValues(resourceType, status).Set(count)
}

// SetResourceState sets the state of a specific resource.
func (m *Metrics) SetResourceState(resourceID, resourceType string, ready bool) {
	if m.resourceState == nil {
		return
	}
	value := 0.0
	if ready {
		value = 1.0
	}
	m.resourceState.WithLabelValues(resourceID, resourceType).Set(value)
}

// Provider Metrics

// RecordProviderCall records a provider call with its duration.
func (m *Metrics) RecordProviderCall(provider, operation string, duration time.Duration) {
	if m.providerCalls == nil {
		return
	}
	m.providerCalls.WithLabelValues(provider, operation).Inc()
	m.providerDuration.WithLabelValues(provider, operation).Observe(duration.Seconds())
}

// RecordProviderError records a provider error.
func (m *Metrics) RecordProviderError(provider, operation string) {
	if m.providerErrors == nil {
		return
	}
	m.providerErrors.WithLabelValues(provider, operation).Inc()
}

// Error Metrics

// RecordError records an error by class and optionally by code.
func (m *Metrics) RecordError(errorClass, errorCode string) {
	if m.errorsByClass == nil {
		return
	}
	m.errorsByClass.WithLabelValues(errorClass).Inc()
	if errorCode != "" && m.errorsByCode != nil {
		m.errorsByCode.WithLabelValues(errorCode).Inc()
	}
}

// Drift Metrics

// RecordDriftDetection records a drift detection event.
func (m *Metrics) RecordDriftDetection(resourceType, status string) {
	if m.driftDetections == nil {
		return
	}
	m.driftDetections.WithLabelValues(resourceType, status).Inc()
}

// System Metrics

// SetActiveRuns sets the current number of active runs.
func (m *Metrics) SetActiveRuns(count float64) {
	if m.activeRuns == nil {
		return
	}
	m.activeRuns.Set(count)
}

// SetQueuedPlanUnits sets the current number of queued plan units.
func (m *Metrics) SetQueuedPlanUnits(count float64) {
	if m.queuedPlanUnits == nil {
		return
	}
	m.queuedPlanUnits.Set(count)
}

// Timer provides a convenient way to time operations.
type Timer struct {
	start time.Time
}

// NewTimer creates a new timer.
func NewTimer() *Timer {
	return &Timer{start: time.Now()}
}

// Duration returns the elapsed time since the timer was created.
func (t *Timer) Duration() time.Duration {
	return time.Since(t.start)
}

// ObserveDuration is a helper to time an operation and record it.
func (t *Timer) ObserveDuration(observer prometheus.Observer) {
	observer.Observe(t.Duration().Seconds())
}

// Handler returns an HTTP handler for the metrics endpoint.
func (m *Metrics) Handler() http.Handler {
	if m.registry == nil {
		return http.NotFoundHandler()
	}
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}

// StartMetricsServer starts an HTTP server to expose metrics.
func (m *Metrics) StartMetricsServer() error {
	if !m.config.Enabled {
		return nil
	}

	mux := http.NewServeMux()
	mux.Handle(m.config.Path, m.Handler())

	server := &http.Server{
		Addr:              m.config.ListenAddress,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Log error but don't fail the application
			fmt.Printf("metrics server error: %v\n", err)
		}
	}()

	return nil
}
