package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
)

// DefaultPlanner implements the Planner interface.
// It computes differences between desired and actual state, builds execution plans,
// and constructs dependency graphs for parallel execution.
type DefaultPlanner struct {
	// providerRegistry is used to invoke provider-specific plan operations
	providerRegistry ProviderRegistry

	// stateManager is used to retrieve resource state
	stateManager StateManager
}

// NewPlanner creates a new default planner implementation.
func NewPlanner(registry ProviderRegistry, stateMgr StateManager) *DefaultPlanner {
	return &DefaultPlanner{
		providerRegistry: registry,
		stateManager:     stateMgr,
	}
}

// ComputeDiff compares desired configuration with actual facts to determine required operations.
func (p *DefaultPlanner) ComputeDiff(ctx context.Context, desired *Config, actual *Facts) (*DiffResult, error) {
	if desired == nil {
		return nil, NewPermanentError("desired configuration is nil", nil).
			WithCode(ErrCodeValidation)
	}

	result := &DiffResult{
		Resources: make([]ResourceDiff, 0, len(desired.Resources)),
		Summary: DiffSummary{
			TotalResources: len(desired.Resources),
		},
		Timestamp: time.Now(),
	}

	// Build a map of actual state by resource ID from facts
	actualStateMap := make(map[string]json.RawMessage)
	if actual != nil {
		// Facts.Data is organized by namespace; we need to extract resource states
		// For now, we'll check the state manager for each resource
	}

	// Process each desired resource
	for _, resource := range desired.Resources {
		diff, err := p.computeResourceDiff(ctx, &resource, actualStateMap)
		if err != nil {
			return nil, fmt.Errorf("failed to compute diff for resource %s: %w", resource.ID, err)
		}

		result.Resources = append(result.Resources, *diff)

		// Update summary statistics
		switch diff.Operation {
		case OperationCreate:
			result.Summary.ToCreate++
		case OperationUpdate:
			result.Summary.ToUpdate++
		case OperationDelete:
			result.Summary.ToDelete++
		case OperationRecreate:
			result.Summary.ToRecreate++
		case OperationNoop:
			result.Summary.NoChange++
		}
	}

	return result, nil
}

// computeResourceDiff computes the diff for a single resource.
func (p *DefaultPlanner) computeResourceDiff(
	ctx context.Context,
	resource *Resource,
	actualStateMap map[string]json.RawMessage,
) (*ResourceDiff, error) {
	diff := &ResourceDiff{
		ResourceID:       resource.ID,
		DesiredState:     resource.Config,
		Changes:          make([]Change, 0),
		RequiresRecreate: false,
	}

	// Try to get actual state from state manager
	actualState, err := p.stateManager.GetResourceState(ctx, resource.ID)
	if err != nil {
		// Resource doesn't exist - needs to be created
		diff.Operation = OperationCreate
		diff.Changes = append(diff.Changes, Change{
			Path:   ".",
			Before: nil,
			After:  resource.Config,
			Action: ChangeActionAdd,
		})
		return diff, nil
	}

	diff.ActualState = actualState

	// Get the provider to compute detailed diff
	provider, err := p.providerRegistry.Get(ctx, resource.Type, "latest")
	if err != nil {
		// If provider not available, do a simple comparison
		if p.statesEqual(resource.Config, actualState) {
			diff.Operation = OperationNoop
			return diff, nil
		}

		// States differ - update needed
		diff.Operation = OperationUpdate
		diff.Changes = p.computeSimpleChanges(resource.Config, actualState)
		return diff, nil
	}

	// Use provider's Plan method to get detailed diff
	planReq := PlanRequest{
		ResourceID:   resource.ID,
		DesiredState: resource.Config,
		ActualState:  actualState,
		Operation:    OperationUpdate,
	}

	planResp, err := provider.Plan(ctx, planReq)
	if err != nil {
		return nil, fmt.Errorf("provider plan failed: %w", err)
	}

	diff.Operation = planResp.Operation
	diff.Changes = planResp.Changes
	diff.RequiresRecreate = planResp.RequiresRecreate

	if planResp.RequiresRecreate {
		diff.Operation = OperationRecreate
	}

	return diff, nil
}

// statesEqual compares two JSON states for equality.
func (p *DefaultPlanner) statesEqual(a, b json.RawMessage) bool {
	var aVal, bVal interface{}

	if err := json.Unmarshal(a, &aVal); err != nil {
		return false
	}
	if err := json.Unmarshal(b, &bVal); err != nil {
		return false
	}

	return reflect.DeepEqual(aVal, bVal)
}

// computeSimpleChanges performs a simple comparison when provider is unavailable.
func (p *DefaultPlanner) computeSimpleChanges(desired, actual json.RawMessage) []Change {
	return []Change{
		{
			Path:   ".",
			Before: actual,
			After:  desired,
			Action: ChangeActionModify,
		},
	}
}

// BuildPlan creates an execution plan from the computed diff.
func (p *DefaultPlanner) BuildPlan(ctx context.Context, diff *DiffResult) (*Plan, error) {
	if diff == nil {
		return nil, NewPermanentError("diff result is nil", nil).
			WithCode(ErrCodeValidation)
	}

	plan := &Plan{
		ID:        uuid.New().String(),
		CreatedAt: time.Now(),
		Units:     make([]PlanUnit, 0, len(diff.Resources)),
		Summary: PlanSummary{
			TotalResources: diff.Summary.TotalResources,
			ToCreate:       diff.Summary.ToCreate,
			ToUpdate:       diff.Summary.ToUpdate,
			ToDelete:       diff.Summary.ToDelete,
			ToRecreate:     diff.Summary.ToRecreate,
			NoChange:       diff.Summary.NoChange,
		},
		Metadata: make(map[string]interface{}),
	}

	// Create plan units from resource diffs
	for _, resourceDiff := range diff.Resources {
		// Skip resources that don't need any changes
		if resourceDiff.Operation == OperationNoop {
			continue
		}

		unit := PlanUnit{
			ID:           uuid.New().String(),
			ResourceID:   resourceDiff.ResourceID,
			Operation:    resourceDiff.Operation,
			Status:       PlanStatusPending,
			DesiredState: resourceDiff.DesiredState,
			ActualState:  resourceDiff.ActualState,
			Changes:      resourceDiff.Changes,
			Timeout:      5 * time.Minute, // Default timeout
			MaxRetries:   3,                // Default max retries
			Metadata:     make(map[string]interface{}),
		}

		// Get resource to extract dependencies
		resource, err := p.stateManager.GetResource(ctx, resourceDiff.ResourceID)
		if err == nil && resource != nil {
			// Map resource dependencies to plan unit dependencies
			unit.Dependencies = p.buildDependencies(ctx, resource.Dependencies, plan.Units)
			unit.ProviderName = resource.Type
		}

		plan.Units = append(plan.Units, unit)
	}

	return plan, nil
}

// buildDependencies converts resource IDs to plan unit dependencies.
func (p *DefaultPlanner) buildDependencies(
	ctx context.Context,
	resourceDeps []string,
	existingUnits []PlanUnit,
) []Dependency {
	deps := make([]Dependency, 0, len(resourceDeps))

	// Build a map of resource ID to plan unit ID
	resourceToUnit := make(map[string]string)
	for _, unit := range existingUnits {
		resourceToUnit[unit.ResourceID] = unit.ID
	}

	// Convert resource dependencies to plan unit dependencies
	for _, resourceID := range resourceDeps {
		if unitID, exists := resourceToUnit[resourceID]; exists {
			deps = append(deps, Dependency{
				TargetID: unitID,
				Type:     DependencyRequire, // Default to required dependency
			})
		}
	}

	return deps
}

// BuildDAG creates the dependency graph for plan execution.
func (p *DefaultPlanner) BuildDAG(ctx context.Context, plan *Plan) (*ExecutionGraph, error) {
	if plan == nil {
		return nil, NewPermanentError("plan is nil", nil).
			WithCode(ErrCodeValidation)
	}

	// Use the DAG builder to construct the graph
	builder := NewDAGBuilder()
	graph, err := builder.BuildGraph(plan.Units)
	if err != nil {
		return nil, fmt.Errorf("failed to build DAG: %w", err)
	}

	// Validate the constructed graph
	if err := builder.ValidateGraph(graph); err != nil {
		return nil, fmt.Errorf("graph validation failed: %w", err)
	}

	// Attach graph to plan
	plan.Graph = graph

	return graph, nil
}

// ValidatePlan validates a plan for correctness and safety.
func (p *DefaultPlanner) ValidatePlan(ctx context.Context, plan *Plan) error {
	if plan == nil {
		return NewPermanentError("plan is nil", nil).
			WithCode(ErrCodeValidation)
	}

	// Validate plan has units
	if len(plan.Units) == 0 {
		return NewPermanentError("plan has no units", nil).
			WithCode(ErrCodeValidation)
	}

	// Validate each plan unit
	for _, unit := range plan.Units {
		if err := p.validatePlanUnit(&unit); err != nil {
			return fmt.Errorf("invalid plan unit %s: %w", unit.ID, err)
		}
	}

	// Validate the execution graph if present
	if plan.Graph != nil {
		builder := NewDAGBuilder()
		if _, err := builder.BuildGraph(plan.Units); err != nil {
			return fmt.Errorf("graph validation failed: %w", err)
		}
	}

	return nil
}

// validatePlanUnit validates a single plan unit.
func (p *DefaultPlanner) validatePlanUnit(unit *PlanUnit) error {
	if unit.ID == "" {
		return NewPermanentError("plan unit has empty ID", nil).
			WithCode(ErrCodeValidation)
	}

	if unit.ResourceID == "" {
		return NewPermanentError("plan unit has empty resource ID", nil).
			WithCode(ErrCodeValidation).
			WithResource(unit.ID)
	}

	if err := unit.Operation.Validate(); err != nil {
		return err
	}

	if err := unit.Status.Validate(); err != nil {
		return err
	}

	if unit.Timeout <= 0 {
		return NewPermanentError("plan unit has invalid timeout", nil).
			WithCode(ErrCodeValidation).
			WithResource(unit.ID)
	}

	if unit.MaxRetries < 0 {
		return NewPermanentError("plan unit has negative max retries", nil).
			WithCode(ErrCodeValidation).
			WithResource(unit.ID)
	}

	return nil
}

// OptimizePlan optimizes the plan for parallel execution.
// This implementation performs basic optimizations like reordering independent units.
func (p *DefaultPlanner) OptimizePlan(ctx context.Context, plan *Plan) (*Plan, error) {
	if plan == nil {
		return nil, NewPermanentError("plan is nil", nil).
			WithCode(ErrCodeValidation)
	}

	// Build or rebuild the DAG to ensure execution levels are optimal
	graph, err := p.BuildDAG(ctx, plan)
	if err != nil {
		return nil, fmt.Errorf("failed to build DAG for optimization: %w", err)
	}

	plan.Graph = graph

	// Optimization: Sort units within each level by operation type
	// Priority: Delete > Create > Update > Noop
	// This ensures deletions happen before creations to avoid conflicts
	p.optimizeExecutionOrder(plan)

	// Optimization: Adjust timeouts based on operation complexity
	p.optimizeTimeouts(plan)

	return plan, nil
}

// optimizeExecutionOrder sorts units within execution levels by priority.
func (p *DefaultPlanner) optimizeExecutionOrder(plan *Plan) {
	// Group units by execution level
	levelUnits := make(map[int][]*PlanUnit)
	for i := range plan.Units {
		unit := &plan.Units[i]
		level := unit.ExecutionOrder
		levelUnits[level] = append(levelUnits[level], unit)
	}

	// Sort units within each level by operation priority
	for _, units := range levelUnits {
		sortPlanUnits(units)
	}
}

// optimizeTimeouts adjusts timeouts based on operation type.
func (p *DefaultPlanner) optimizeTimeouts(plan *Plan) {
	for i := range plan.Units {
		unit := &plan.Units[i]

		// Adjust timeout based on operation type
		switch unit.Operation {
		case OperationCreate, OperationRecreate:
			// Creation operations may take longer
			if unit.Timeout < 10*time.Minute {
				unit.Timeout = 10 * time.Minute
			}
		case OperationDelete:
			// Deletion operations are usually quick
			if unit.Timeout > 3*time.Minute {
				unit.Timeout = 3 * time.Minute
			}
		case OperationUpdate:
			// Update operations have medium timeout
			if unit.Timeout < 5*time.Minute {
				unit.Timeout = 5 * time.Minute
			}
		}
	}
}

// sortPlanUnits sorts plan units by operation priority.
func sortPlanUnits(units []*PlanUnit) {
	// Simple bubble sort for operation priority
	// Priority: Delete (4) > Recreate (3) > Create (2) > Update (1) > Noop (0)
	for i := 0; i < len(units); i++ {
		for j := i + 1; j < len(units); j++ {
			if getOperationPriority(units[i].Operation) < getOperationPriority(units[j].Operation) {
				units[i], units[j] = units[j], units[i]
			}
		}
	}
}

// getOperationPriority returns the priority value for an operation type.
func getOperationPriority(op OperationType) int {
	switch op {
	case OperationDelete:
		return 4
	case OperationRecreate:
		return 3
	case OperationCreate:
		return 2
	case OperationUpdate:
		return 1
	case OperationNoop:
		return 0
	default:
		return 0
	}
}
