package engine

import (
	"encoding/json"
	"time"
)

// Resource represents a managed infrastructure resource.
type Resource struct {
	// ID is the unique identifier for this resource.
	ID string `json:"id"`

	// Type is the resource type (e.g., "linux.pkg", "linux.service").
	Type string `json:"type"`

	// Name is the human-readable name of the resource.
	Name string `json:"name"`

	// Config is the desired configuration for this resource.
	Config json.RawMessage `json:"config"`

	// State is the current state of the resource.
	State json.RawMessage `json:"state,omitempty"`

	// Status is the current status of the resource.
	Status ResourceStatus `json:"status"`

	// Labels are key-value pairs for organizing and selecting resources.
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations are additional metadata for the resource.
	Annotations map[string]string `json:"annotations,omitempty"`

	// Dependencies lists resource IDs that this resource depends on.
	Dependencies []string `json:"dependencies,omitempty"`

	// CreatedAt is when the resource was first created.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the resource was last updated.
	UpdatedAt time.Time `json:"updated_at"`

	// Version is the resource version for optimistic locking.
	Version int64 `json:"version"`
}

// PlanUnit represents a unit of work in the execution DAG.
type PlanUnit struct {
	// ID is the unique identifier for this plan unit.
	ID string `json:"id"`

	// ResourceID is the ID of the resource this plan unit operates on.
	ResourceID string `json:"resource_id"`

	// Operation is the type of operation to perform.
	Operation OperationType `json:"operation"`

	// Status is the current execution status of this plan unit.
	Status PlanStatus `json:"status"`

	// Dependencies lists plan unit IDs that must complete before this unit.
	Dependencies []Dependency `json:"dependencies,omitempty"`

	// DesiredState is the desired configuration after this operation.
	DesiredState json.RawMessage `json:"desired_state,omitempty"`

	// ActualState is the current state before this operation.
	ActualState json.RawMessage `json:"actual_state,omitempty"`

	// Changes describes what will change if this operation is applied.
	Changes []Change `json:"changes,omitempty"`

	// ProviderName is the name of the provider that will execute this operation.
	ProviderName string `json:"provider_name"`

	// ProviderVersion is the version of the provider.
	ProviderVersion string `json:"provider_version"`

	// ExecutionOrder is the topological order for execution.
	ExecutionOrder int `json:"execution_order"`

	// Retries is the number of retry attempts for this plan unit.
	Retries int `json:"retries"`

	// MaxRetries is the maximum number of retry attempts allowed.
	MaxRetries int `json:"max_retries"`

	// Timeout is the maximum duration for executing this plan unit.
	Timeout time.Duration `json:"timeout"`

	// Metadata contains additional plan unit metadata.
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Result is the execution result once the plan unit completes.
	Result *ExecutionResult `json:"result,omitempty"`
}

// Dependency represents an edge in the execution DAG.
type Dependency struct {
	// TargetID is the ID of the plan unit this depends on.
	TargetID string `json:"target_id"`

	// Type is the type of dependency relationship.
	Type DependencyType `json:"type"`
}

// DependencyType represents the type of dependency between plan units.
type DependencyType string

const (
	// DependencyRequire indicates a hard dependency that must succeed.
	DependencyRequire DependencyType = "require"

	// DependencyNotify indicates a soft dependency for triggering handlers.
	DependencyNotify DependencyType = "notify"

	// DependencyOrder indicates ordering without success requirement.
	DependencyOrder DependencyType = "order"
)

// Change represents a single change to be applied to a resource.
type Change struct {
	// Path is the JSON path to the field being changed (e.g., ".config.version").
	Path string `json:"path"`

	// Before is the value before the change.
	Before interface{} `json:"before,omitempty"`

	// After is the value after the change.
	After interface{} `json:"after,omitempty"`

	// Action describes the change action (add, remove, modify).
	Action ChangeAction `json:"action"`
}

// ChangeAction represents the type of change being made.
type ChangeAction string

const (
	// ChangeActionAdd indicates a new field is being added.
	ChangeActionAdd ChangeAction = "add"

	// ChangeActionRemove indicates a field is being removed.
	ChangeActionRemove ChangeAction = "remove"

	// ChangeActionModify indicates a field value is being changed.
	ChangeActionModify ChangeAction = "modify"
)

// ExecutionResult represents the outcome of executing a plan unit.
type ExecutionResult struct {
	// PlanUnitID is the ID of the plan unit this result belongs to.
	PlanUnitID string `json:"plan_unit_id"`

	// Status indicates whether the execution succeeded or failed.
	Status PlanStatus `json:"status"`

	// StartedAt is when the execution started.
	StartedAt time.Time `json:"started_at"`

	// CompletedAt is when the execution completed.
	CompletedAt time.Time `json:"completed_at"`

	// Duration is the total execution time.
	Duration time.Duration `json:"duration"`

	// NewState is the resulting state after the operation.
	NewState json.RawMessage `json:"new_state,omitempty"`

	// Output contains any output data from the provider.
	Output json.RawMessage `json:"output,omitempty"`

	// Error is the error that occurred, if any.
	Error *EngineError `json:"error,omitempty"`

	// Events are timeline events that occurred during execution.
	Events []Event `json:"events,omitempty"`

	// Metrics contains execution metrics (e.g., resource usage).
	Metrics map[string]interface{} `json:"metrics,omitempty"`
}

// Event represents a timeline event during execution.
type Event struct {
	// ID is the unique identifier for this event.
	ID string `json:"id"`

	// Type is the type of event.
	Type EventType `json:"type"`

	// Timestamp is when the event occurred.
	Timestamp time.Time `json:"timestamp"`

	// RunID is the ID of the run this event belongs to.
	RunID string `json:"run_id"`

	// PlanUnitID is the ID of the plan unit, if applicable.
	PlanUnitID string `json:"plan_unit_id,omitempty"`

	// ResourceID is the ID of the resource, if applicable.
	ResourceID string `json:"resource_id,omitempty"`

	// Message is a human-readable event message.
	Message string `json:"message"`

	// Details contains additional event-specific data.
	Details map[string]interface{} `json:"details,omitempty"`

	// Level is the log level (info, warning, error).
	Level string `json:"level"`
}

// Plan represents a complete execution plan.
type Plan struct {
	// ID is the unique identifier for this plan.
	ID string `json:"id"`

	// RunID is the ID of the run that created this plan.
	RunID string `json:"run_id"`

	// CreatedAt is when the plan was created.
	CreatedAt time.Time `json:"created_at"`

	// Units are all the plan units to be executed.
	Units []PlanUnit `json:"units"`

	// Graph is the DAG representation of the plan.
	Graph *ExecutionGraph `json:"graph,omitempty"`

	// Summary provides high-level statistics about the plan.
	Summary PlanSummary `json:"summary"`

	// Metadata contains additional plan metadata.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// PlanSummary provides statistics about a plan.
type PlanSummary struct {
	// TotalResources is the total number of resources in the plan.
	TotalResources int `json:"total_resources"`

	// ToCreate is the number of resources to create.
	ToCreate int `json:"to_create"`

	// ToUpdate is the number of resources to update.
	ToUpdate int `json:"to_update"`

	// ToDelete is the number of resources to delete.
	ToDelete int `json:"to_delete"`

	// ToRecreate is the number of resources to recreate.
	ToRecreate int `json:"to_recreate"`

	// NoChange is the number of resources with no changes.
	NoChange int `json:"no_change"`
}

// ExecutionGraph represents the DAG of plan units.
type ExecutionGraph struct {
	// Nodes maps plan unit IDs to their graph nodes.
	Nodes map[string]*GraphNode `json:"nodes"`

	// Edges lists all dependency edges in the graph.
	Edges []GraphEdge `json:"edges"`

	// Roots are the plan unit IDs with no dependencies.
	Roots []string `json:"roots"`

	// Depth is the maximum depth of the graph.
	Depth int `json:"depth"`
}

// GraphNode represents a node in the execution graph.
type GraphNode struct {
	// ID is the plan unit ID.
	ID string `json:"id"`

	// Level is the topological level (depth from roots).
	Level int `json:"level"`

	// Dependencies are the incoming edges (units this depends on).
	Dependencies []string `json:"dependencies"`

	// Dependents are the outgoing edges (units that depend on this).
	Dependents []string `json:"dependents"`
}

// GraphEdge represents an edge in the execution graph.
type GraphEdge struct {
	// From is the source plan unit ID.
	From string `json:"from"`

	// To is the target plan unit ID.
	To string `json:"to"`

	// Type is the dependency type.
	Type DependencyType `json:"type"`
}

// Run represents an execution run of a plan.
type Run struct {
	// ID is the unique identifier for this run.
	ID string `json:"id"`

	// PlanID is the ID of the plan being executed.
	PlanID string `json:"plan_id"`

	// Status is the current status of the run.
	Status RunStatus `json:"status"`

	// StartedAt is when the run started.
	StartedAt time.Time `json:"started_at"`

	// CompletedAt is when the run completed.
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// Duration is the total run duration.
	Duration time.Duration `json:"duration"`

	// User is the user who initiated the run.
	User string `json:"user,omitempty"`

	// Summary provides statistics about the run.
	Summary RunSummary `json:"summary"`

	// Metadata contains additional run metadata.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// RunSummary provides statistics about a run.
type RunSummary struct {
	// Total is the total number of plan units.
	Total int `json:"total"`

	// Succeeded is the number of plan units that succeeded.
	Succeeded int `json:"succeeded"`

	// Failed is the number of plan units that failed.
	Failed int `json:"failed"`

	// Skipped is the number of plan units that were skipped.
	Skipped int `json:"skipped"`

	// Pending is the number of plan units still pending.
	Pending int `json:"pending"`

	// Running is the number of plan units currently running.
	Running int `json:"running"`
}

// DriftDetection represents drift detection results for a resource.
type DriftDetection struct {
	// ResourceID is the ID of the resource.
	ResourceID string `json:"resource_id"`

	// Status is the drift status.
	Status DriftStatus `json:"status"`

	// DetectedAt is when drift was detected.
	DetectedAt time.Time `json:"detected_at"`

	// DesiredState is the desired state from configuration.
	DesiredState json.RawMessage `json:"desired_state"`

	// ActualState is the actual state discovered.
	ActualState json.RawMessage `json:"actual_state"`

	// Drifts lists the specific drifts detected.
	Drifts []Change `json:"drifts,omitempty"`

	// Reconciled indicates if the drift has been reconciled.
	Reconciled bool `json:"reconciled"`

	// ReconciledAt is when the drift was reconciled.
	ReconciledAt *time.Time `json:"reconciled_at,omitempty"`
}

// Facts represents discovered system state.
type Facts struct {
	// TargetID is the ID of the target system.
	TargetID string `json:"target_id"`

	// CollectedAt is when the facts were collected.
	CollectedAt time.Time `json:"collected_at"`

	// TTL is how long the facts are considered valid.
	TTL time.Duration `json:"ttl"`

	// Data contains the actual fact data organized by namespace.
	Data map[string]json.RawMessage `json:"data"`

	// Version is the facts schema version.
	Version string `json:"version"`
}

// Config represents parsed CUE configuration.
type Config struct {
	// ID is the unique identifier for this configuration.
	ID string `json:"id"`

	// Source is the source file or content.
	Source string `json:"source"`

	// ParsedAt is when the configuration was parsed.
	ParsedAt time.Time `json:"parsed_at"`

	// Resources are the resources defined in the configuration.
	Resources []Resource `json:"resources"`

	// Variables are the configuration variables.
	Variables map[string]interface{} `json:"variables,omitempty"`

	// Metadata contains additional configuration metadata.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}
