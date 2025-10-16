package stores

import (
	"context"
	"database/sql"
	"time"
)

// RunStatus represents the status of an execution run
type RunStatus string

const (
	RunStatusPending   RunStatus = "pending"
	RunStatusRunning   RunStatus = "running"
	RunStatusCompleted RunStatus = "completed"
	RunStatusFailed    RunStatus = "failed"
	RunStatusCancelled RunStatus = "cancelled"
)

// PlanUnitStatus represents the status of a plan unit
type PlanUnitStatus string

const (
	PlanUnitStatusPending   PlanUnitStatus = "pending"
	PlanUnitStatusRunning   PlanUnitStatus = "running"
	PlanUnitStatusCompleted PlanUnitStatus = "completed"
	PlanUnitStatusFailed    PlanUnitStatus = "failed"
	PlanUnitStatusSkipped   PlanUnitStatus = "skipped"
)

// EventLevel represents the severity level of an event
type EventLevel string

const (
	EventLevelDebug   EventLevel = "debug"
	EventLevelInfo    EventLevel = "info"
	EventLevelWarning EventLevel = "warning"
	EventLevelError   EventLevel = "error"
)

// Run represents an execution run
type Run struct {
	ID          string     `json:"id"`
	PlanPath    string     `json:"plan_path"`
	Status      RunStatus  `json:"status"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Error       *string    `json:"error,omitempty"`
	Metadata    string     `json:"metadata"` // JSON blob
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// PlanUnit represents a single unit in an execution plan
type PlanUnit struct {
	ID           string         `json:"id"`
	RunID        string         `json:"run_id"`
	ResourceType string         `json:"resource_type"`
	ResourceName string         `json:"resource_name"`
	Action       string         `json:"action"` // create, update, delete, read
	Status       PlanUnitStatus `json:"status"`
	Dependencies string         `json:"dependencies"`           // JSON array of PlanUnit IDs
	DesiredState string         `json:"desired_state"`          // JSON blob
	ActualState  *string        `json:"actual_state,omitempty"` // JSON blob
	Diff         *string        `json:"diff,omitempty"`         // JSON blob
	StartedAt    *time.Time     `json:"started_at,omitempty"`
	CompletedAt  *time.Time     `json:"completed_at,omitempty"`
	Error        *string        `json:"error,omitempty"`
	Retries      int            `json:"retries"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// Event represents an append-only log event
type Event struct {
	ID         int64      `json:"id"`
	RunID      *string    `json:"run_id,omitempty"`
	PlanUnitID *string    `json:"plan_unit_id,omitempty"`
	Level      EventLevel `json:"level"`
	Message    string     `json:"message"`
	Details    *string    `json:"details,omitempty"` // JSON blob
	Timestamp  time.Time  `json:"timestamp"`
}

// ResourceState represents the current state of a managed resource
type ResourceState struct {
	ID           string    `json:"id"`
	ResourceType string    `json:"resource_type"`
	ResourceName string    `json:"resource_name"`
	State        string    `json:"state"` // JSON blob
	Hash         string    `json:"hash"`  // SHA256 of state for drift detection
	LastRunID    string    `json:"last_run_id"`
	LastApplied  time.Time `json:"last_applied"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Fact represents discovered facts about managed systems
type Fact struct {
	ID        string     `json:"id"`
	TargetID  string     `json:"target_id"` // host/system identifier
	Namespace string     `json:"namespace"` // e.g., "os.basic", "hw.cpu", "net.ifaces"
	Key       string     `json:"key"`
	Value     string     `json:"value"` // JSON blob
	TTL       int        `json:"ttl"`   // seconds, 0 = no expiry
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// AuditEntry represents an audit trail entry
type AuditEntry struct {
	ID        int64     `json:"id"`
	Action    string    `json:"action"`              // e.g., "run.created", "state.updated", "plan.applied"
	Actor     string    `json:"actor"`               // user or system identifier
	TargetID  *string   `json:"target_id,omitempty"` // resource/run/etc ID
	Details   *string   `json:"details,omitempty"`   // JSON blob
	IPAddress *string   `json:"ip_address,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// Store defines the interface for the persistence layer
type Store interface {
	// Lifecycle
	Init(ctx context.Context) error
	Close() error
	Migrate(ctx context.Context) error

	// Transaction support
	BeginTx(ctx context.Context) (*sql.Tx, error)
	CommitTx(tx *sql.Tx) error
	RollbackTx(tx *sql.Tx) error

	// Run operations
	CreateRun(ctx context.Context, run *Run) error
	GetRun(ctx context.Context, id string) (*Run, error)
	UpdateRunStatus(ctx context.Context, id string, status RunStatus, err *string) error
	ListRuns(ctx context.Context, limit, offset int) ([]*Run, error)
	DeleteRun(ctx context.Context, id string) error

	// PlanUnit operations
	CreatePlanUnit(ctx context.Context, unit *PlanUnit) error
	GetPlanUnit(ctx context.Context, id string) (*PlanUnit, error)
	UpdatePlanUnitStatus(ctx context.Context, id string, status PlanUnitStatus, actualState *string, err *string) error
	ListPlanUnitsByRun(ctx context.Context, runID string) ([]*PlanUnit, error)
	DeletePlanUnit(ctx context.Context, id string) error
	IncrementPlanUnitRetries(ctx context.Context, id string) error

	// Event operations
	AppendEvent(ctx context.Context, event *Event) error
	GetEvents(ctx context.Context, runID *string, planUnitID *string, level *EventLevel, limit, offset int) ([]*Event, error)

	// ResourceState operations
	UpsertResourceState(ctx context.Context, state *ResourceState) error
	GetResourceState(ctx context.Context, resourceType, resourceName string) (*ResourceState, error)
	ListResourceStates(ctx context.Context, limit, offset int) ([]*ResourceState, error)
	DeleteResourceState(ctx context.Context, id string) error

	// Facts operations
	UpsertFact(ctx context.Context, fact *Fact) error
	GetFact(ctx context.Context, targetID, namespace, key string) (*Fact, error)
	ListFacts(ctx context.Context, targetID *string, namespace *string, limit, offset int) ([]*Fact, error)
	DeleteExpiredFacts(ctx context.Context) (int64, error)
	DeleteFact(ctx context.Context, id string) error

	// Audit operations
	CreateAuditEntry(ctx context.Context, entry *AuditEntry) error
	ListAuditEntries(ctx context.Context, action *string, actor *string, limit, offset int) ([]*AuditEntry, error)

	// Utility
	HealthCheck(ctx context.Context) error
}
