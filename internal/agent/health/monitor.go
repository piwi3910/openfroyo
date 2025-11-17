package health

import (
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/piwi3910/openfroyo/internal/agent/config"
	"github.com/piwi3910/openfroyo/internal/protocol"
)

// Monitor tracks agent health metrics
type Monitor struct {
	config     *config.Config
	logger     *log.Logger
	startTime  time.Time
	stats      HealthStats
	statsMutex sync.RWMutex
}

// HealthStats contains health statistics
type HealthStats struct {
	TasksExecuted  int64
	TasksSucceeded int64
	TasksFailed    int64
	LastTaskAt     *time.Time
}

// NewMonitor creates a new health monitor
func NewMonitor(cfg *config.Config, logger *log.Logger) *Monitor {
	return &Monitor{
		config:    cfg,
		logger:    logger,
		startTime: time.Now(),
	}
}

// GetHealthStatus returns current health status
func (m *Monitor) GetHealthStatus() *protocol.HealthStatus {
	m.statsMutex.RLock()
	defer m.statsMutex.RUnlock()

	hostname, _ := os.Hostname()
	uptime := int64(time.Since(m.startTime).Seconds())

	// Get memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	memoryMB := int64(memStats.Alloc / 1024 / 1024)

	// Determine health status
	status := m.determineHealthStatus()

	// Get CPU percent (simplified - would need more sophisticated tracking)
	cpuPercent := m.getCPUPercent()

	return &protocol.HealthStatus{
		AgentID:        m.config.Agent.ID,
		Hostname:       hostname,
		Datacenter:     m.config.Agent.Datacenter,
		Version:        m.config.Agent.Version,
		Status:         status,
		Uptime:         uptime,
		TasksExecuted:  m.stats.TasksExecuted,
		TasksSucceeded: m.stats.TasksSucceeded,
		TasksFailed:    m.stats.TasksFailed,
		LastTaskAt:     m.stats.LastTaskAt,
		CPUPercent:     cpuPercent,
		MemoryMB:       memoryMB,
		Tags:           m.buildTags(),
		Timestamp:      time.Now(),
	}
}

// UpdateTaskStats updates task execution statistics
func (m *Monitor) UpdateTaskStats(status string) {
	m.statsMutex.Lock()
	defer m.statsMutex.Unlock()

	m.stats.TasksExecuted++
	now := time.Now()
	m.stats.LastTaskAt = &now

	switch status {
	case protocol.TaskStatusOK, protocol.TaskStatusChanged:
		m.stats.TasksSucceeded++
	case protocol.TaskStatusFailed:
		m.stats.TasksFailed++
	}
}

// determineHealthStatus determines overall health status
func (m *Monitor) determineHealthStatus() string {
	// Calculate failure rate
	if m.stats.TasksExecuted == 0 {
		return protocol.HealthStatusHealthy
	}

	failureRate := float64(m.stats.TasksFailed) / float64(m.stats.TasksExecuted)

	// Determine status based on failure rate
	if failureRate > 0.5 {
		return protocol.HealthStatusUnhealthy
	} else if failureRate > 0.1 {
		return protocol.HealthStatusDegraded
	}

	return protocol.HealthStatusHealthy
}

// getCPUPercent returns CPU usage percentage (simplified)
func (m *Monitor) getCPUPercent() float64 {
	// This is a simplified version - a real implementation would
	// track CPU usage over time using system calls
	return float64(runtime.NumGoroutine()) * 0.1
}

// buildTags builds agent tags map
func (m *Monitor) buildTags() map[string]string {
	tags := make(map[string]string)
	for _, tag := range m.config.Agent.Tags {
		tags[tag] = "true"
	}
	tags["os"] = runtime.GOOS
	tags["arch"] = runtime.GOARCH
	return tags
}
