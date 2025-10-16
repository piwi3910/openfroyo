package engine

import (
	"context"
	"encoding/json"
	"io"
	"time"
)

// Evaluator parses and validates CUE configurations.
// This is Phase 1: Config evaluation.
type Evaluator interface {
	// Evaluate parses CUE configuration files and returns the parsed configuration.
	Evaluate(ctx context.Context, sources []string) (*Config, error)

	// Validate validates a configuration against schemas and policies.
	Validate(ctx context.Context, config *Config) error

	// EvaluateStarlark executes Starlark scripts for procedural logic.
	EvaluateStarlark(ctx context.Context, script string, input map[string]interface{}) (map[string]interface{}, error)

	// MergeConfigs merges multiple configurations into a single configuration.
	MergeConfigs(ctx context.Context, configs []*Config) (*Config, error)
}

// Discoverer collects facts about target systems.
// This is Phase 2: Facts discovery.
type Discoverer interface {
	// DiscoverFacts collects facts from a target system.
	DiscoverFacts(ctx context.Context, targetID string, namespaces []string) (*Facts, error)

	// RefreshFacts refreshes facts if they are stale (beyond TTL).
	RefreshFacts(ctx context.Context, targetID string, force bool) (*Facts, error)

	// GetCachedFacts retrieves cached facts if available and valid.
	GetCachedFacts(ctx context.Context, targetID string) (*Facts, error)

	// RegisterFactCollector registers a custom fact collector for a namespace.
	RegisterFactCollector(namespace string, collector FactCollector) error
}

// FactCollector is an interface for collecting specific types of facts.
type FactCollector interface {
	// Collect collects facts for a specific namespace.
	Collect(ctx context.Context, target TargetInfo) (json.RawMessage, error)

	// Namespace returns the namespace this collector handles.
	Namespace() string

	// TTL returns the time-to-live for facts from this collector.
	TTL() time.Duration
}

// TargetInfo contains information about a target system.
type TargetInfo struct {
	// ID is the unique identifier of the target.
	ID string `json:"id"`

	// Type is the target type (e.g., "ssh", "local", "winrm").
	Type string `json:"type"`

	// Hostname is the target hostname or IP address.
	Hostname string `json:"hostname,omitempty"`

	// Labels are key-value pairs for organizing targets.
	Labels map[string]string `json:"labels,omitempty"`

	// Metadata contains additional target metadata.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Planner computes differences and builds execution plans.
// This is Phase 3: Plan generation.
type Planner interface {
	// ComputeDiff compares desired state with actual state.
	ComputeDiff(ctx context.Context, desired *Config, actual *Facts) (*DiffResult, error)

	// BuildPlan creates an execution plan from the diff.
	BuildPlan(ctx context.Context, diff *DiffResult) (*Plan, error)

	// BuildDAG creates the dependency graph for plan execution.
	BuildDAG(ctx context.Context, plan *Plan) (*ExecutionGraph, error)

	// ValidatePlan validates a plan for correctness and safety.
	ValidatePlan(ctx context.Context, plan *Plan) error

	// OptimizePlan optimizes the plan for parallel execution.
	OptimizePlan(ctx context.Context, plan *Plan) (*Plan, error)
}

// DiffResult represents the result of comparing desired and actual state.
type DiffResult struct {
	// Resources lists all resources and their differences.
	Resources []ResourceDiff `json:"resources"`

	// Summary provides statistics about the diff.
	Summary DiffSummary `json:"summary"`

	// Timestamp is when the diff was computed.
	Timestamp time.Time `json:"timestamp"`
}

// ResourceDiff represents the difference for a single resource.
type ResourceDiff struct {
	// ResourceID is the ID of the resource.
	ResourceID string `json:"resource_id"`

	// Operation is the required operation.
	Operation OperationType `json:"operation"`

	// DesiredState is the desired state from configuration.
	DesiredState json.RawMessage `json:"desired_state"`

	// ActualState is the actual state from facts.
	ActualState json.RawMessage `json:"actual_state,omitempty"`

	// Changes lists the specific differences.
	Changes []Change `json:"changes"`

	// RequiresRecreate indicates if recreation is needed.
	RequiresRecreate bool `json:"requires_recreate"`
}

// DiffSummary provides statistics about a diff.
type DiffSummary struct {
	// TotalResources is the total number of resources.
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

// Executor executes plans by running the DAG.
// This is Phase 4: Apply/Execute.
type Executor interface {
	// Execute runs a plan to completion.
	Execute(ctx context.Context, plan *Plan) (*Run, error)

	// ExecuteUnit executes a single plan unit.
	ExecuteUnit(ctx context.Context, unit *PlanUnit) (*ExecutionResult, error)

	// Cancel cancels a running execution.
	Cancel(ctx context.Context, runID string) error

	// GetRunStatus retrieves the current status of a run.
	GetRunStatus(ctx context.Context, runID string) (*Run, error)

	// StreamEvents streams execution events as they occur.
	StreamEvents(ctx context.Context, runID string) (<-chan Event, error)
}

// StateManager manages resource state persistence.
type StateManager interface {
	// GetResource retrieves a resource by ID.
	GetResource(ctx context.Context, resourceID string) (*Resource, error)

	// SaveResource persists a resource state.
	SaveResource(ctx context.Context, resource *Resource) error

	// DeleteResource removes a resource from state.
	DeleteResource(ctx context.Context, resourceID string) error

	// ListResources lists all resources matching the selector.
	ListResources(ctx context.Context, selector map[string]string) ([]Resource, error)

	// GetResourceState retrieves only the state portion of a resource.
	GetResourceState(ctx context.Context, resourceID string) (json.RawMessage, error)

	// UpdateResourceState updates only the state portion of a resource.
	UpdateResourceState(ctx context.Context, resourceID string, state json.RawMessage, version int64) error

	// Lock acquires an advisory lock on a resource.
	Lock(ctx context.Context, resourceID string) error

	// Unlock releases an advisory lock on a resource.
	Unlock(ctx context.Context, resourceID string) error

	// GetPlan retrieves a plan by ID.
	GetPlan(ctx context.Context, planID string) (*Plan, error)

	// SavePlan persists a plan.
	SavePlan(ctx context.Context, plan *Plan) error

	// GetRun retrieves a run by ID.
	GetRun(ctx context.Context, runID string) (*Run, error)

	// SaveRun persists a run.
	SaveRun(ctx context.Context, run *Run) error

	// AppendEvent appends an event to the event log.
	AppendEvent(ctx context.Context, event *Event) error

	// GetEvents retrieves events for a run.
	GetEvents(ctx context.Context, runID string) ([]Event, error)
}

// DriftDetector detects configuration drift.
// This is Phase 6: Drift detection.
type DriftDetector interface {
	// DetectDrift compares current state with stored state to find drift.
	DetectDrift(ctx context.Context, resourceID string) (*DriftDetection, error)

	// DetectAllDrift checks drift for all resources matching selector.
	DetectAllDrift(ctx context.Context, selector map[string]string) ([]DriftDetection, error)

	// ReconcileDrift automatically reconciles detected drift.
	ReconcileDrift(ctx context.Context, resourceID string) error

	// ShouldReconcile determines if drift should be auto-reconciled based on policy.
	ShouldReconcile(ctx context.Context, drift *DriftDetection) (bool, error)
}

// PolicyEngine enforces policies on configurations and operations.
type PolicyEngine interface {
	// Evaluate evaluates policies against a configuration.
	Evaluate(ctx context.Context, config *Config) (*PolicyResult, error)

	// EvaluatePlan evaluates policies against a plan.
	EvaluatePlan(ctx context.Context, plan *Plan) (*PolicyResult, error)

	// EvaluateResource evaluates policies against a single resource.
	EvaluateResource(ctx context.Context, resource *Resource) (*PolicyResult, error)

	// LoadPolicies loads policy files.
	LoadPolicies(ctx context.Context, paths []string) error
}

// PolicyResult represents the result of policy evaluation.
type PolicyResult struct {
	// Allowed indicates if the operation is allowed.
	Allowed bool `json:"allowed"`

	// Violations lists policy violations.
	Violations []PolicyViolation `json:"violations,omitempty"`

	// Warnings lists policy warnings.
	Warnings []string `json:"warnings,omitempty"`

	// EvaluatedAt is when the policy was evaluated.
	EvaluatedAt time.Time `json:"evaluated_at"`
}

// PolicyViolation represents a single policy violation.
type PolicyViolation struct {
	// Policy is the policy name that was violated.
	Policy string `json:"policy"`

	// Message is a human-readable violation message.
	Message string `json:"message"`

	// Severity is the violation severity (error, warning).
	Severity string `json:"severity"`

	// ResourceID is the resource that violated the policy, if applicable.
	ResourceID string `json:"resource_id,omitempty"`
}

// ProviderRegistry manages provider plugins.
type ProviderRegistry interface {
	// Register registers a provider plugin.
	Register(ctx context.Context, manifest *ProviderManifest, wasmModule []byte) error

	// Get retrieves a provider by name and version.
	Get(ctx context.Context, name, version string) (Provider, error)

	// List lists all registered providers.
	List(ctx context.Context) ([]ProviderMetadata, error)

	// Unregister removes a provider from the registry.
	Unregister(ctx context.Context, name, version string) error

	// ValidateCapabilities validates that requested capabilities are allowed.
	ValidateCapabilities(ctx context.Context, capabilities []string) error
}

// EventPublisher publishes events to subscribers.
type EventPublisher interface {
	// Publish publishes an event.
	Publish(ctx context.Context, event *Event) error

	// Subscribe subscribes to events matching a filter.
	Subscribe(ctx context.Context, filter EventFilter) (<-chan Event, error)

	// Unsubscribe removes a subscription.
	Unsubscribe(ctx context.Context, subscriptionID string) error
}

// EventFilter represents criteria for filtering events.
type EventFilter struct {
	// RunID filters events by run ID.
	RunID string `json:"run_id,omitempty"`

	// ResourceID filters events by resource ID.
	ResourceID string `json:"resource_id,omitempty"`

	// Types filters events by type.
	Types []EventType `json:"types,omitempty"`

	// MinLevel filters events by minimum log level.
	MinLevel string `json:"min_level,omitempty"`
}

// Scheduler schedules and manages plan executions.
type Scheduler interface {
	// Schedule schedules a plan for execution.
	Schedule(ctx context.Context, plan *Plan, opts ScheduleOptions) (string, error)

	// Cancel cancels a scheduled or running execution.
	Cancel(ctx context.Context, runID string) error

	// GetStatus retrieves the status of a scheduled run.
	GetStatus(ctx context.Context, runID string) (*Run, error)
}

// ScheduleOptions contains options for scheduling a plan execution.
type ScheduleOptions struct {
	// Delay is the delay before execution starts.
	Delay time.Duration `json:"delay,omitempty"`

	// MaxParallel is the maximum number of plan units to execute in parallel.
	MaxParallel int `json:"max_parallel,omitempty"`

	// DryRun executes the plan in dry-run mode (no actual changes).
	DryRun bool `json:"dry_run,omitempty"`

	// FailFast stops execution on first failure.
	FailFast bool `json:"fail_fast,omitempty"`

	// User is the user initiating the execution.
	User string `json:"user,omitempty"`
}

// BackupManager handles backup and restore operations.
type BackupManager interface {
	// Backup creates a backup of all state data.
	Backup(ctx context.Context, dest io.Writer) error

	// Restore restores state data from a backup.
	Restore(ctx context.Context, src io.Reader) error

	// ListBackups lists available backups.
	ListBackups(ctx context.Context) ([]BackupInfo, error)
}

// BackupInfo contains information about a backup.
type BackupInfo struct {
	// ID is the backup identifier.
	ID string `json:"id"`

	// CreatedAt is when the backup was created.
	CreatedAt time.Time `json:"created_at"`

	// Size is the backup size in bytes.
	Size int64 `json:"size"`

	// ResourceCount is the number of resources in the backup.
	ResourceCount int `json:"resource_count"`
}
