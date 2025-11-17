package health

import (
	"log"
	"time"

	"github.com/piwi3910/openfroyo/internal/protocol"
)

// Publisher is the interface for publishing health status
type Publisher interface {
	PublishHealthStatus(subject string, health *protocol.HealthStatus) error
}

// Reporter periodically reports health status
type Reporter struct {
	monitor   *Monitor
	publisher Publisher
	subject   string
	interval  time.Duration
	logger    *log.Logger
	stopChan  chan struct{}
}

// NewReporter creates a new health reporter
func NewReporter(monitor *Monitor, publisher Publisher, subject string, interval time.Duration, logger *log.Logger) *Reporter {
	return &Reporter{
		monitor:   monitor,
		publisher: publisher,
		subject:   subject,
		interval:  interval,
		logger:    logger,
		stopChan:  make(chan struct{}),
	}
}

// Start starts the health reporter
func (r *Reporter) Start() {
	r.logger.Printf("[INFO] Starting health reporter (interval: %v)", r.interval)

	// Send initial health status
	r.report()

	// Start periodic reporting
	ticker := time.NewTicker(r.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				r.report()
			case <-r.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

// Stop stops the health reporter
func (r *Reporter) Stop() {
	r.logger.Println("[INFO] Stopping health reporter")
	close(r.stopChan)

	// Send final health status with shutdown status
	health := r.monitor.GetHealthStatus()
	health.Status = protocol.HealthStatusShutdown
	if err := r.publisher.PublishHealthStatus(r.subject, health); err != nil {
		r.logger.Printf("[ERROR] Failed to publish shutdown health status: %v", err)
	}
}

// report publishes current health status
func (r *Reporter) report() {
	health := r.monitor.GetHealthStatus()

	if err := r.publisher.PublishHealthStatus(r.subject, health); err != nil {
		r.logger.Printf("[ERROR] Failed to publish health status: %v", err)
	} else {
		r.logger.Printf("[DEBUG] Published health status: %s (uptime: %ds, tasks: %d/%d/%d)",
			health.Status, health.Uptime,
			health.TasksSucceeded, health.TasksFailed, health.TasksExecuted)
	}
}
