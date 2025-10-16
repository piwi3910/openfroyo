package engine

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

// Mock implementations for testing

type mockProviderRegistry struct {
	providers map[string]Provider
}

func (m *mockProviderRegistry) Register(ctx context.Context, manifest *ProviderManifest, wasmModule []byte) error {
	return nil
}

func (m *mockProviderRegistry) Get(ctx context.Context, name, version string) (Provider, error) {
	if provider, exists := m.providers[name]; exists {
		return provider, nil
	}
	return nil, NewPermanentError("provider not found", nil).WithCode(ErrCodeNotFound)
}

func (m *mockProviderRegistry) List(ctx context.Context) ([]ProviderMetadata, error) {
	return []ProviderMetadata{}, nil
}

func (m *mockProviderRegistry) Unregister(ctx context.Context, name, version string) error {
	return nil
}

func (m *mockProviderRegistry) ValidateCapabilities(ctx context.Context, capabilities []string) error {
	return nil
}

type mockStateManager struct {
	resources map[string]*Resource
	states    map[string]json.RawMessage
	plans     map[string]*Plan
	runs      map[string]*Run
}

func newMockStateManager() *mockStateManager {
	return &mockStateManager{
		resources: make(map[string]*Resource),
		states:    make(map[string]json.RawMessage),
		plans:     make(map[string]*Plan),
		runs:      make(map[string]*Run),
	}
}

func (m *mockStateManager) GetResource(ctx context.Context, resourceID string) (*Resource, error) {
	if resource, exists := m.resources[resourceID]; exists {
		return resource, nil
	}
	return nil, NewPermanentError("resource not found", nil).WithCode(ErrCodeNotFound)
}

func (m *mockStateManager) SaveResource(ctx context.Context, resource *Resource) error {
	m.resources[resource.ID] = resource
	return nil
}

func (m *mockStateManager) DeleteResource(ctx context.Context, resourceID string) error {
	delete(m.resources, resourceID)
	return nil
}

func (m *mockStateManager) ListResources(ctx context.Context, selector map[string]string) ([]Resource, error) {
	resources := make([]Resource, 0, len(m.resources))
	for _, r := range m.resources {
		resources = append(resources, *r)
	}
	return resources, nil
}

func (m *mockStateManager) GetResourceState(ctx context.Context, resourceID string) (json.RawMessage, error) {
	if state, exists := m.states[resourceID]; exists {
		return state, nil
	}
	return nil, NewPermanentError("resource state not found", nil).WithCode(ErrCodeNotFound)
}

func (m *mockStateManager) UpdateResourceState(ctx context.Context, resourceID string, state json.RawMessage, version int64) error {
	m.states[resourceID] = state
	return nil
}

func (m *mockStateManager) Lock(ctx context.Context, resourceID string) error {
	return nil
}

func (m *mockStateManager) Unlock(ctx context.Context, resourceID string) error {
	return nil
}

func (m *mockStateManager) GetPlan(ctx context.Context, planID string) (*Plan, error) {
	if plan, exists := m.plans[planID]; exists {
		return plan, nil
	}
	return nil, NewPermanentError("plan not found", nil).WithCode(ErrCodeNotFound)
}

func (m *mockStateManager) SavePlan(ctx context.Context, plan *Plan) error {
	m.plans[plan.ID] = plan
	return nil
}

func (m *mockStateManager) GetRun(ctx context.Context, runID string) (*Run, error) {
	if run, exists := m.runs[runID]; exists {
		return run, nil
	}
	return nil, NewPermanentError("run not found", nil).WithCode(ErrCodeNotFound)
}

func (m *mockStateManager) SaveRun(ctx context.Context, run *Run) error {
	m.runs[run.ID] = run
	return nil
}

func (m *mockStateManager) AppendEvent(ctx context.Context, event *Event) error {
	return nil
}

func (m *mockStateManager) GetEvents(ctx context.Context, runID string) ([]Event, error) {
	return []Event{}, nil
}

func TestNewPlanner(t *testing.T) {
	registry := &mockProviderRegistry{providers: make(map[string]Provider)}
	stateMgr := newMockStateManager()

	planner := NewPlanner(registry, stateMgr)

	if planner == nil {
		t.Fatal("Expected non-nil planner")
	}

	if planner.providerRegistry != registry {
		t.Error("Provider registry not set correctly")
	}

	if planner.stateManager != stateMgr {
		t.Error("State manager not set correctly")
	}
}

func TestPlanner_ComputeDiff_NilConfig(t *testing.T) {
	registry := &mockProviderRegistry{providers: make(map[string]Provider)}
	stateMgr := newMockStateManager()
	planner := NewPlanner(registry, stateMgr)

	ctx := context.Background()
	_, err := planner.ComputeDiff(ctx, nil, nil)

	if err == nil {
		t.Fatal("Expected error for nil config, got nil")
	}

	if !IsPermanent(err) {
		t.Error("Expected permanent error for nil config")
	}
}

func TestPlanner_ComputeDiff_NewResource(t *testing.T) {
	registry := &mockProviderRegistry{providers: make(map[string]Provider)}
	stateMgr := newMockStateManager()
	planner := NewPlanner(registry, stateMgr)

	ctx := context.Background()

	config := &Config{
		ID:     "config1",
		Source: "test.cue",
		Resources: []Resource{
			{
				ID:     "resource1",
				Type:   "test.resource",
				Name:   "Test Resource",
				Config: json.RawMessage(`{"key": "value"}`),
				Status: ResourceStatusUnknown,
			},
		},
	}

	diff, err := planner.ComputeDiff(ctx, config, nil)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(diff.Resources) != 1 {
		t.Fatalf("Expected 1 resource diff, got %d", len(diff.Resources))
	}

	resourceDiff := diff.Resources[0]
	if resourceDiff.Operation != OperationCreate {
		t.Errorf("Expected CREATE operation, got %s", resourceDiff.Operation)
	}

	if diff.Summary.ToCreate != 1 {
		t.Errorf("Expected 1 to create, got %d", diff.Summary.ToCreate)
	}

	if diff.Summary.TotalResources != 1 {
		t.Errorf("Expected 1 total resource, got %d", diff.Summary.TotalResources)
	}
}

func TestPlanner_ComputeDiff_ExistingResourceNoChange(t *testing.T) {
	registry := &mockProviderRegistry{providers: make(map[string]Provider)}
	stateMgr := newMockStateManager()

	// Add existing state
	state := json.RawMessage(`{"key": "value"}`)
	stateMgr.states["resource1"] = state

	planner := NewPlanner(registry, stateMgr)
	ctx := context.Background()

	config := &Config{
		ID:     "config1",
		Source: "test.cue",
		Resources: []Resource{
			{
				ID:     "resource1",
				Type:   "test.resource",
				Name:   "Test Resource",
				Config: json.RawMessage(`{"key": "value"}`),
				Status: ResourceStatusReady,
			},
		},
	}

	diff, err := planner.ComputeDiff(ctx, config, nil)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(diff.Resources) != 1 {
		t.Fatalf("Expected 1 resource diff, got %d", len(diff.Resources))
	}

	resourceDiff := diff.Resources[0]
	if resourceDiff.Operation != OperationNoop {
		t.Errorf("Expected NOOP operation, got %s", resourceDiff.Operation)
	}

	if diff.Summary.NoChange != 1 {
		t.Errorf("Expected 1 no change, got %d", diff.Summary.NoChange)
	}
}

func TestPlanner_ComputeDiff_ExistingResourceChanged(t *testing.T) {
	registry := &mockProviderRegistry{providers: make(map[string]Provider)}
	stateMgr := newMockStateManager()

	// Add existing state with different value
	state := json.RawMessage(`{"key": "old_value"}`)
	stateMgr.states["resource1"] = state

	planner := NewPlanner(registry, stateMgr)
	ctx := context.Background()

	config := &Config{
		ID:     "config1",
		Source: "test.cue",
		Resources: []Resource{
			{
				ID:     "resource1",
				Type:   "test.resource",
				Name:   "Test Resource",
				Config: json.RawMessage(`{"key": "new_value"}`),
				Status: ResourceStatusReady,
			},
		},
	}

	diff, err := planner.ComputeDiff(ctx, config, nil)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(diff.Resources) != 1 {
		t.Fatalf("Expected 1 resource diff, got %d", len(diff.Resources))
	}

	resourceDiff := diff.Resources[0]
	if resourceDiff.Operation != OperationUpdate {
		t.Errorf("Expected UPDATE operation, got %s", resourceDiff.Operation)
	}

	if diff.Summary.ToUpdate != 1 {
		t.Errorf("Expected 1 to update, got %d", diff.Summary.ToUpdate)
	}
}

func TestPlanner_BuildPlan_NilDiff(t *testing.T) {
	registry := &mockProviderRegistry{providers: make(map[string]Provider)}
	stateMgr := newMockStateManager()
	planner := NewPlanner(registry, stateMgr)

	ctx := context.Background()
	_, err := planner.BuildPlan(ctx, nil)

	if err == nil {
		t.Fatal("Expected error for nil diff, got nil")
	}
}

func TestPlanner_BuildPlan_Success(t *testing.T) {
	registry := &mockProviderRegistry{providers: make(map[string]Provider)}
	stateMgr := newMockStateManager()

	// Add resource to state manager
	resource := &Resource{
		ID:           "resource1",
		Type:         "test.resource",
		Name:         "Test Resource",
		Dependencies: []string{},
	}
	stateMgr.resources["resource1"] = resource

	planner := NewPlanner(registry, stateMgr)
	ctx := context.Background()

	diff := &DiffResult{
		Resources: []ResourceDiff{
			{
				ResourceID:   "resource1",
				Operation:    OperationCreate,
				DesiredState: json.RawMessage(`{"key": "value"}`),
				Changes: []Change{
					{
						Path:   ".",
						Before: nil,
						After:  json.RawMessage(`{"key": "value"}`),
						Action: ChangeActionAdd,
					},
				},
			},
		},
		Summary: DiffSummary{
			TotalResources: 1,
			ToCreate:       1,
		},
		Timestamp: time.Now(),
	}

	plan, err := planner.BuildPlan(ctx, diff)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(plan.Units) != 1 {
		t.Fatalf("Expected 1 plan unit, got %d", len(plan.Units))
	}

	unit := plan.Units[0]
	if unit.ResourceID != "resource1" {
		t.Errorf("Expected resource1, got %s", unit.ResourceID)
	}

	if unit.Operation != OperationCreate {
		t.Errorf("Expected CREATE operation, got %s", unit.Operation)
	}

	if unit.Status != PlanStatusPending {
		t.Errorf("Expected PENDING status, got %s", unit.Status)
	}

	if plan.Summary.ToCreate != 1 {
		t.Errorf("Expected 1 to create in summary, got %d", plan.Summary.ToCreate)
	}
}

func TestPlanner_BuildPlan_SkipNoopOperations(t *testing.T) {
	registry := &mockProviderRegistry{providers: make(map[string]Provider)}
	stateMgr := newMockStateManager()
	planner := NewPlanner(registry, stateMgr)

	ctx := context.Background()

	diff := &DiffResult{
		Resources: []ResourceDiff{
			{
				ResourceID:   "resource1",
				Operation:    OperationCreate,
				DesiredState: json.RawMessage(`{"key": "value1"}`),
			},
			{
				ResourceID:   "resource2",
				Operation:    OperationNoop,
				DesiredState: json.RawMessage(`{"key": "value2"}`),
			},
			{
				ResourceID:   "resource3",
				Operation:    OperationUpdate,
				DesiredState: json.RawMessage(`{"key": "value3"}`),
			},
		},
		Summary: DiffSummary{
			TotalResources: 3,
			ToCreate:       1,
			ToUpdate:       1,
			NoChange:       1,
		},
		Timestamp: time.Now(),
	}

	plan, err := planner.BuildPlan(ctx, diff)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should only have 2 units (skipping noop)
	if len(plan.Units) != 2 {
		t.Fatalf("Expected 2 plan units, got %d", len(plan.Units))
	}

	// Verify noop was skipped
	for _, unit := range plan.Units {
		if unit.ResourceID == "resource2" {
			t.Error("Noop operation should have been skipped")
		}
	}
}

func TestPlanner_BuildDAG_NilPlan(t *testing.T) {
	registry := &mockProviderRegistry{providers: make(map[string]Provider)}
	stateMgr := newMockStateManager()
	planner := NewPlanner(registry, stateMgr)

	ctx := context.Background()
	_, err := planner.BuildDAG(ctx, nil)

	if err == nil {
		t.Fatal("Expected error for nil plan, got nil")
	}
}

func TestPlanner_BuildDAG_Success(t *testing.T) {
	registry := &mockProviderRegistry{providers: make(map[string]Provider)}
	stateMgr := newMockStateManager()
	planner := NewPlanner(registry, stateMgr)

	ctx := context.Background()

	plan := &Plan{
		ID:        "plan1",
		CreatedAt: time.Now(),
		Units: []PlanUnit{
			{
				ID:           "unit1",
				ResourceID:   "resource1",
				Operation:    OperationCreate,
				Status:       PlanStatusPending,
				Dependencies: []Dependency{},
				Timeout:      time.Minute,
				MaxRetries:   3,
			},
			{
				ID:         "unit2",
				ResourceID: "resource2",
				Operation:  OperationCreate,
				Status:     PlanStatusPending,
				Dependencies: []Dependency{
					{TargetID: "unit1", Type: DependencyRequire},
				},
				Timeout:    time.Minute,
				MaxRetries: 3,
			},
		},
	}

	graph, err := planner.BuildDAG(ctx, plan)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(graph.Nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(graph.Nodes))
	}

	if graph.Depth != 2 {
		t.Errorf("Expected depth 2, got %d", graph.Depth)
	}

	// Verify graph is attached to plan
	if plan.Graph == nil {
		t.Error("Expected graph to be attached to plan")
	}
}

func TestPlanner_ValidatePlan_NilPlan(t *testing.T) {
	registry := &mockProviderRegistry{providers: make(map[string]Provider)}
	stateMgr := newMockStateManager()
	planner := NewPlanner(registry, stateMgr)

	ctx := context.Background()
	err := planner.ValidatePlan(ctx, nil)

	if err == nil {
		t.Fatal("Expected error for nil plan, got nil")
	}
}

func TestPlanner_ValidatePlan_EmptyPlan(t *testing.T) {
	registry := &mockProviderRegistry{providers: make(map[string]Provider)}
	stateMgr := newMockStateManager()
	planner := NewPlanner(registry, stateMgr)

	ctx := context.Background()

	plan := &Plan{
		ID:        "plan1",
		CreatedAt: time.Now(),
		Units:     []PlanUnit{},
	}

	err := planner.ValidatePlan(ctx, plan)

	if err == nil {
		t.Fatal("Expected error for empty plan, got nil")
	}
}

func TestPlanner_ValidatePlan_ValidPlan(t *testing.T) {
	registry := &mockProviderRegistry{providers: make(map[string]Provider)}
	stateMgr := newMockStateManager()
	planner := NewPlanner(registry, stateMgr)

	ctx := context.Background()

	plan := &Plan{
		ID:        "plan1",
		CreatedAt: time.Now(),
		Units: []PlanUnit{
			{
				ID:           "unit1",
				ResourceID:   "resource1",
				Operation:    OperationCreate,
				Status:       PlanStatusPending,
				Dependencies: []Dependency{},
				Timeout:      time.Minute,
				MaxRetries:   3,
			},
		},
	}

	err := planner.ValidatePlan(ctx, plan)

	if err != nil {
		t.Fatalf("Expected no error for valid plan, got: %v", err)
	}
}

func TestPlanner_ValidatePlan_InvalidUnit(t *testing.T) {
	registry := &mockProviderRegistry{providers: make(map[string]Provider)}
	stateMgr := newMockStateManager()
	planner := NewPlanner(registry, stateMgr)

	ctx := context.Background()

	plan := &Plan{
		ID:        "plan1",
		CreatedAt: time.Now(),
		Units: []PlanUnit{
			{
				ID:           "", // Invalid empty ID
				ResourceID:   "resource1",
				Operation:    OperationCreate,
				Status:       PlanStatusPending,
				Dependencies: []Dependency{},
				Timeout:      time.Minute,
				MaxRetries:   3,
			},
		},
	}

	err := planner.ValidatePlan(ctx, plan)

	if err == nil {
		t.Fatal("Expected error for invalid unit, got nil")
	}
}

func TestPlanner_OptimizePlan_Success(t *testing.T) {
	registry := &mockProviderRegistry{providers: make(map[string]Provider)}
	stateMgr := newMockStateManager()
	planner := NewPlanner(registry, stateMgr)

	ctx := context.Background()

	plan := &Plan{
		ID:        "plan1",
		CreatedAt: time.Now(),
		Units: []PlanUnit{
			{
				ID:           "unit1",
				ResourceID:   "resource1",
				Operation:    OperationCreate,
				Status:       PlanStatusPending,
				Dependencies: []Dependency{},
				Timeout:      time.Second, // Short timeout
				MaxRetries:   3,
			},
		},
	}

	optimized, err := planner.OptimizePlan(ctx, plan)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if optimized.Graph == nil {
		t.Error("Expected graph to be built during optimization")
	}

	// Check that timeout was optimized (create operations get longer timeouts)
	if optimized.Units[0].Timeout < 10*time.Minute {
		t.Errorf("Expected timeout >= 10 minutes for create operation, got %v", optimized.Units[0].Timeout)
	}
}
