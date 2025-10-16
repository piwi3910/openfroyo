package engine

import (
	"encoding/json"
	"fmt"
)

// RunStatus represents the overall status of a plan execution run.
type RunStatus string

const (
	// RunStatusPending indicates the run is queued but not yet started.
	RunStatusPending RunStatus = "pending"

	// RunStatusRunning indicates the run is currently executing.
	RunStatusRunning RunStatus = "running"

	// RunStatusSucceeded indicates the run completed successfully.
	RunStatusSucceeded RunStatus = "succeeded"

	// RunStatusFailed indicates the run failed with errors.
	RunStatusFailed RunStatus = "failed"

	// RunStatusCancelled indicates the run was cancelled by the user.
	RunStatusCancelled RunStatus = "cancelled"

	// RunStatusPartial indicates the run partially succeeded (some resources failed).
	RunStatusPartial RunStatus = "partial"
)

// IsTerminal returns true if the run status represents a final state.
func (s RunStatus) IsTerminal() bool {
	return s == RunStatusSucceeded || s == RunStatusFailed ||
		s == RunStatusCancelled || s == RunStatusPartial
}

// IsActive returns true if the run is currently active (pending or running).
func (s RunStatus) IsActive() bool {
	return s == RunStatusPending || s == RunStatusRunning
}

// Validate checks if the run status is valid.
func (s RunStatus) Validate() error {
	switch s {
	case RunStatusPending, RunStatusRunning, RunStatusSucceeded,
		RunStatusFailed, RunStatusCancelled, RunStatusPartial:
		return nil
	default:
		return fmt.Errorf("invalid run status: %s", s)
	}
}

// OperationType represents the type of operation to perform on a resource.
type OperationType string

const (
	// OperationCreate indicates a new resource should be created.
	OperationCreate OperationType = "create"

	// OperationUpdate indicates an existing resource should be updated.
	OperationUpdate OperationType = "update"

	// OperationDelete indicates an existing resource should be deleted.
	OperationDelete OperationType = "delete"

	// OperationNoop indicates no operation is needed (resource is in desired state).
	OperationNoop OperationType = "noop"

	// OperationRead indicates a read-only operation to refresh state.
	OperationRead OperationType = "read"

	// OperationRecreate indicates a resource must be destroyed and recreated.
	OperationRecreate OperationType = "recreate"
)

// IsDestructive returns true if the operation modifies or destroys resources.
func (o OperationType) IsDestructive() bool {
	return o == OperationDelete || o == OperationRecreate
}

// IsMutating returns true if the operation changes resource state.
func (o OperationType) IsMutating() bool {
	return o == OperationCreate || o == OperationUpdate ||
		o == OperationDelete || o == OperationRecreate
}

// Validate checks if the operation type is valid.
func (o OperationType) Validate() error {
	switch o {
	case OperationCreate, OperationUpdate, OperationDelete,
		OperationNoop, OperationRead, OperationRecreate:
		return nil
	default:
		return fmt.Errorf("invalid operation type: %s", o)
	}
}

// ResourceStatus represents the current status of a resource.
type ResourceStatus string

const (
	// ResourceStatusUnknown indicates the resource state is not yet known.
	ResourceStatusUnknown ResourceStatus = "unknown"

	// ResourceStatusCreating indicates the resource is being created.
	ResourceStatusCreating ResourceStatus = "creating"

	// ResourceStatusReady indicates the resource is ready and operational.
	ResourceStatusReady ResourceStatus = "ready"

	// ResourceStatusUpdating indicates the resource is being updated.
	ResourceStatusUpdating ResourceStatus = "updating"

	// ResourceStatusDeleting indicates the resource is being deleted.
	ResourceStatusDeleting ResourceStatus = "deleting"

	// ResourceStatusError indicates the resource is in an error state.
	ResourceStatusError ResourceStatus = "error"

	// ResourceStatusDrifted indicates the resource has drifted from desired state.
	ResourceStatusDrifted ResourceStatus = "drifted"

	// ResourceStatusPending indicates the resource operation is pending.
	ResourceStatusPending ResourceStatus = "pending"

	// ResourceStatusDeleted indicates the resource has been deleted.
	ResourceStatusDeleted ResourceStatus = "deleted"
)

// IsTransitional returns true if the status represents a transitional state.
func (s ResourceStatus) IsTransitional() bool {
	return s == ResourceStatusCreating || s == ResourceStatusUpdating ||
		s == ResourceStatusDeleting || s == ResourceStatusPending
}

// IsTerminal returns true if the status represents a final state.
func (s ResourceStatus) IsTerminal() bool {
	return s == ResourceStatusReady || s == ResourceStatusError ||
		s == ResourceStatusDeleted
}

// Validate checks if the resource status is valid.
func (s ResourceStatus) Validate() error {
	switch s {
	case ResourceStatusUnknown, ResourceStatusCreating, ResourceStatusReady,
		ResourceStatusUpdating, ResourceStatusDeleting, ResourceStatusError,
		ResourceStatusDrifted, ResourceStatusPending, ResourceStatusDeleted:
		return nil
	default:
		return fmt.Errorf("invalid resource status: %s", s)
	}
}

// PlanStatus represents the status of a plan unit during execution.
type PlanStatus string

const (
	// PlanStatusPending indicates the plan unit is waiting to execute.
	PlanStatusPending PlanStatus = "pending"

	// PlanStatusBlocked indicates the plan unit is blocked by dependencies.
	PlanStatusBlocked PlanStatus = "blocked"

	// PlanStatusRunning indicates the plan unit is currently executing.
	PlanStatusRunning PlanStatus = "running"

	// PlanStatusSucceeded indicates the plan unit completed successfully.
	PlanStatusSucceeded PlanStatus = "succeeded"

	// PlanStatusFailed indicates the plan unit failed.
	PlanStatusFailed PlanStatus = "failed"

	// PlanStatusSkipped indicates the plan unit was skipped due to failures.
	PlanStatusSkipped PlanStatus = "skipped"

	// PlanStatusCancelled indicates the plan unit was cancelled.
	PlanStatusCancelled PlanStatus = "cancelled"
)

// IsTerminal returns true if the plan status represents a final state.
func (s PlanStatus) IsTerminal() bool {
	return s == PlanStatusSucceeded || s == PlanStatusFailed ||
		s == PlanStatusSkipped || s == PlanStatusCancelled
}

// IsActive returns true if the plan unit is currently active.
func (s PlanStatus) IsActive() bool {
	return s == PlanStatusPending || s == PlanStatusBlocked || s == PlanStatusRunning
}

// Validate checks if the plan status is valid.
func (s PlanStatus) Validate() error {
	switch s {
	case PlanStatusPending, PlanStatusBlocked, PlanStatusRunning,
		PlanStatusSucceeded, PlanStatusFailed, PlanStatusSkipped, PlanStatusCancelled:
		return nil
	default:
		return fmt.Errorf("invalid plan status: %s", s)
	}
}

// DriftStatus represents the drift detection status of a resource.
type DriftStatus string

const (
	// DriftStatusInSync indicates the resource matches desired state.
	DriftStatusInSync DriftStatus = "in_sync"

	// DriftStatusDrifted indicates the resource has drifted from desired state.
	DriftStatusDrifted DriftStatus = "drifted"

	// DriftStatusUnknown indicates drift status could not be determined.
	DriftStatusUnknown DriftStatus = "unknown"

	// DriftStatusNotApplicable indicates drift detection is not applicable.
	DriftStatusNotApplicable DriftStatus = "not_applicable"
)

// Validate checks if the drift status is valid.
func (s DriftStatus) Validate() error {
	switch s {
	case DriftStatusInSync, DriftStatusDrifted, DriftStatusUnknown, DriftStatusNotApplicable:
		return nil
	default:
		return fmt.Errorf("invalid drift status: %s", s)
	}
}

// EventType represents the type of event in the execution timeline.
type EventType string

const (
	// EventTypeRunStarted indicates a run has started.
	EventTypeRunStarted EventType = "run_started"

	// EventTypeRunCompleted indicates a run has completed.
	EventTypeRunCompleted EventType = "run_completed"

	// EventTypeRunFailed indicates a run has failed.
	EventTypeRunFailed EventType = "run_failed"

	// EventTypePlanUnitStarted indicates a plan unit has started execution.
	EventTypePlanUnitStarted EventType = "plan_unit_started"

	// EventTypePlanUnitCompleted indicates a plan unit has completed successfully.
	EventTypePlanUnitCompleted EventType = "plan_unit_completed"

	// EventTypePlanUnitFailed indicates a plan unit has failed.
	EventTypePlanUnitFailed EventType = "plan_unit_failed"

	// EventTypeResourceChanged indicates a resource state has changed.
	EventTypeResourceChanged EventType = "resource_changed"

	// EventTypeProviderInvoked indicates a provider was invoked.
	EventTypeProviderInvoked EventType = "provider_invoked"

	// EventTypeDriftDetected indicates drift was detected.
	EventTypeDriftDetected EventType = "drift_detected"

	// EventTypeError indicates an error occurred.
	EventTypeError EventType = "error"

	// EventTypeWarning indicates a warning was raised.
	EventTypeWarning EventType = "warning"

	// EventTypeInfo indicates informational event.
	EventTypeInfo EventType = "info"
)

// Severity returns the severity level of the event type.
func (e EventType) Severity() string {
	switch e {
	case EventTypeRunFailed, EventTypePlanUnitFailed, EventTypeError:
		return "error"
	case EventTypeWarning:
		return "warning"
	default:
		return "info"
	}
}

// MarshalJSON implements custom JSON marshaling for type-safe enum serialization.
func (s RunStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

// UnmarshalJSON implements custom JSON unmarshaling with validation.
func (s *RunStatus) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*s = RunStatus(str)
	return s.Validate()
}
