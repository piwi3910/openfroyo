package protocol

import "time"

// TaskMessage represents a task sent from orchestrator to agent
type TaskMessage struct {
	TaskID    string                 `json:"task_id"`    // Unique task identifier
	Type      string                 `json:"type"`       // "module" or "command"
	Module    string                 `json:"module"`     // Module path (e.g., "generic/file")
	Vars      map[string]interface{} `json:"vars"`       // Input variables
	Context   TaskContext            `json:"context"`    // Execution context
	Timeout   int                    `json:"timeout"`    // Timeout in seconds
	Priority  int                    `json:"priority"`   // Task priority (1-10, higher = more priority)
	CreatedAt time.Time              `json:"created_at"` // Task creation timestamp
}

// TaskContext provides execution context for a task
type TaskContext struct {
	StackName  string            `json:"stack_name"`  // Source stack name
	TaskName   string            `json:"task_name"`   // Human-readable task name
	Host       string            `json:"host"`        // Target hostname
	Datacenter string            `json:"datacenter"`  // Datacenter identifier
	Tags       map[string]string `json:"tags"`        // Additional metadata tags
}

// TaskResult represents the result of a task execution
type TaskResult struct {
	TaskID      string                 `json:"task_id"`      // Matching task ID
	AgentID     string                 `json:"agent_id"`     // Agent identifier
	Status      string                 `json:"status"`       // "ok", "changed", "failed"
	Message     string                 `json:"message"`      // Result message
	Facts       map[string]interface{} `json:"facts"`        // Discovered facts
	StartedAt   time.Time              `json:"started_at"`   // Execution start time
	CompletedAt time.Time              `json:"completed_at"` // Execution end time
	Duration    float64                `json:"duration"`     // Duration in seconds
	Error       string                 `json:"error"`        // Error message if failed
	Output      string                 `json:"output"`       // Command output (for debugging)
}

// HealthStatus represents agent health information
type HealthStatus struct {
	AgentID        string            `json:"agent_id"`         // Agent identifier
	Hostname       string            `json:"hostname"`         // System hostname
	Datacenter     string            `json:"datacenter"`       // Datacenter location
	Version        string            `json:"version"`          // Agent version
	Status         string            `json:"status"`           // "healthy", "degraded", "unhealthy", "shutdown"
	Uptime         int64             `json:"uptime"`           // Uptime in seconds
	TasksExecuted  int64             `json:"tasks_executed"`   // Total tasks executed
	TasksSucceeded int64             `json:"tasks_succeeded"`  // Successful tasks
	TasksFailed    int64             `json:"tasks_failed"`     // Failed tasks
	LastTaskAt     *time.Time        `json:"last_task_at"`     // Last task execution (nullable)
	CPUPercent     float64           `json:"cpu_percent"`      // CPU usage percentage
	MemoryMB       int64             `json:"memory_mb"`        // Memory usage in MB
	Tags           map[string]string `json:"tags"`             // Agent tags
	Timestamp      time.Time         `json:"timestamp"`        // Status timestamp
}

// PullSchedule defines a periodic stack execution schedule
type PullSchedule struct {
	StackName string                 `json:"stack_name"` // Stack to execute
	Interval  int                    `json:"interval"`   // Interval in seconds
	Enabled   bool                   `json:"enabled"`    // Enable/disable schedule
	Vars      map[string]interface{} `json:"vars"`       // Stack variables
	Tags      []string               `json:"tags"`       // Required agent tags (filter)
	LastRun   *time.Time             `json:"last_run"`   // Last execution time (runtime)
	NextRun   *time.Time             `json:"next_run"`   // Next scheduled run (runtime)
}

// ControlMessage represents control messages sent to agents
type ControlMessage struct {
	Type      string                 `json:"type"`       // "shutdown", "reload", "update_schedule"
	Payload   map[string]interface{} `json:"payload"`    // Control message payload
	Timestamp time.Time              `json:"timestamp"`  // Message timestamp
}

// AgentRegistration represents agent registration/announcement
type AgentRegistration struct {
	AgentID    string            `json:"agent_id"`    // Agent identifier
	Hostname   string            `json:"hostname"`    // System hostname
	Datacenter string            `json:"datacenter"`  // Datacenter location
	Version    string            `json:"version"`     // Agent version
	Tags       map[string]string `json:"tags"`        // Agent tags
	Timestamp  time.Time         `json:"timestamp"`   // Registration timestamp
}

// Task status constants
const (
	TaskStatusOK      = "ok"
	TaskStatusChanged = "changed"
	TaskStatusFailed  = "failed"
)

// Agent health status constants
const (
	HealthStatusHealthy   = "healthy"
	HealthStatusDegraded  = "degraded"
	HealthStatusUnhealthy = "unhealthy"
	HealthStatusShutdown  = "shutdown"
)

// Task type constants
const (
	TaskTypeModule  = "module"
	TaskTypeCommand = "command"
)

// Control message type constants
const (
	ControlTypeShutdown       = "shutdown"
	ControlTypeReload         = "reload"
	ControlTypeUpdateSchedule = "update_schedule"
)
