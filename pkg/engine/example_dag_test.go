package engine_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/openfroyo/openfroyo/pkg/engine"
)

// Example demonstrates building and executing a plan with dependencies.
func Example_dagExecution() {
	// Create plan units with dependencies
	// This example creates a simple web application deployment:
	// 1. Create database
	// 2. Run migrations (depends on database)
	// 3. Deploy app server (depends on migrations)
	// 4. Configure load balancer (depends on app server)

	units := []engine.PlanUnit{
		{
			ID:           "db-001",
			ResourceID:   "postgres-primary",
			Operation:    engine.OperationCreate,
			Status:       engine.PlanStatusPending,
			DesiredState: json.RawMessage(`{"version": "14", "size": "large"}`),
			Dependencies: []engine.Dependency{},
			Timeout:      10 * time.Minute,
			MaxRetries:   3,
		},
		{
			ID:           "migrate-001",
			ResourceID:   "db-migrations",
			Operation:    engine.OperationCreate,
			Status:       engine.PlanStatusPending,
			DesiredState: json.RawMessage(`{"version": "v2.5.0"}`),
			Dependencies: []engine.Dependency{
				{TargetID: "db-001", Type: engine.DependencyRequire},
			},
			Timeout:    5 * time.Minute,
			MaxRetries: 3,
		},
		{
			ID:           "app-001",
			ResourceID:   "web-server",
			Operation:    engine.OperationCreate,
			Status:       engine.PlanStatusPending,
			DesiredState: json.RawMessage(`{"replicas": 3, "version": "v1.2.0"}`),
			Dependencies: []engine.Dependency{
				{TargetID: "migrate-001", Type: engine.DependencyRequire},
			},
			Timeout:    10 * time.Minute,
			MaxRetries: 3,
		},
		{
			ID:           "lb-001",
			ResourceID:   "load-balancer",
			Operation:    engine.OperationCreate,
			Status:       engine.PlanStatusPending,
			DesiredState: json.RawMessage(`{"targets": ["app-001"]}`),
			Dependencies: []engine.Dependency{
				{TargetID: "app-001", Type: engine.DependencyRequire},
			},
			Timeout:    5 * time.Minute,
			MaxRetries: 3,
		},
	}

	// Build the DAG
	builder := engine.NewDAGBuilder()
	graph, err := builder.BuildGraph(units)
	if err != nil {
		log.Fatalf("Failed to build DAG: %v", err)
	}

	// Print execution levels
	fmt.Printf("Execution graph depth: %d levels\n", graph.Depth)
	fmt.Printf("Root nodes: %v\n", graph.Roots)

	for level, nodeIDs := range builder.GetLevels() {
		fmt.Printf("Level %d: %v\n", level, nodeIDs)
	}

	// Generate DOT visualization
	dot := builder.ToDOT()
	fmt.Printf("\nGenerated DOT graph (%d bytes)\n", len(dot))

	// Output:
	// Execution graph depth: 4 levels
	// Root nodes: [db-001]
	// Level 0: [db-001]
	// Level 1: [migrate-001]
	// Level 2: [app-001]
	// Level 3: [lb-001]
	//
	// Generated DOT graph (916 bytes)
}

// Example_plannerWorkflow demonstrates the complete planner workflow.
func Example_plannerWorkflow() {
	ctx := context.Background()

	// Mock implementations (in real usage, use actual implementations)
	registry := &mockProviderRegistry{providers: make(map[string]engine.Provider)}
	stateMgr := newMockStateManager()

	// Create planner
	planner := engine.NewPlanner(registry, stateMgr)

	// Define desired configuration
	config := &engine.Config{
		ID:     "config-001",
		Source: "infrastructure.cue",
		Resources: []engine.Resource{
			{
				ID:     "web-server-1",
				Type:   "aws.ec2",
				Name:   "Web Server 1",
				Config: json.RawMessage(`{"instance_type": "t3.medium"}`),
			},
			{
				ID:     "web-server-2",
				Type:   "aws.ec2",
				Name:   "Web Server 2",
				Config: json.RawMessage(`{"instance_type": "t3.medium"}`),
			},
		},
	}

	// Step 1: Compute diff between desired and actual state
	diff, err := planner.ComputeDiff(ctx, config, nil)
	if err != nil {
		log.Fatalf("Failed to compute diff: %v", err)
	}

	fmt.Printf("Diff summary: %d to create, %d to update, %d no change\n",
		diff.Summary.ToCreate, diff.Summary.ToUpdate, diff.Summary.NoChange)

	// Step 2: Build execution plan
	plan, err := planner.BuildPlan(ctx, diff)
	if err != nil {
		log.Fatalf("Failed to build plan: %v", err)
	}

	fmt.Printf("Plan created with %d units\n", len(plan.Units))

	// Step 3: Build DAG for parallel execution
	graph, err := planner.BuildDAG(ctx, plan)
	if err != nil {
		log.Fatalf("Failed to build DAG: %v", err)
	}

	fmt.Printf("Execution graph has %d levels\n", graph.Depth)

	// Step 4: Validate plan
	if err := planner.ValidatePlan(ctx, plan); err != nil {
		log.Fatalf("Plan validation failed: %v", err)
	}

	fmt.Println("Plan validated successfully")

	// Step 5: Optimize plan
	optimized, err := planner.OptimizePlan(ctx, plan)
	if err != nil {
		log.Fatalf("Failed to optimize plan: %v", err)
	}

	fmt.Printf("Plan optimized with %d units\n", len(optimized.Units))

	// Output:
	// Diff summary: 2 to create, 0 to update, 0 no change
	// Plan created with 2 units
	// Execution graph has 1 levels
	// Plan validated successfully
	// Plan optimized with 2 units
}

// Mock implementations for examples

type mockProviderRegistry struct {
	providers map[string]engine.Provider
}

func (m *mockProviderRegistry) Register(ctx context.Context, manifest *engine.ProviderManifest, wasmModule []byte) error {
	return nil
}

func (m *mockProviderRegistry) Get(ctx context.Context, name, version string) (engine.Provider, error) {
	return nil, engine.NewPermanentError("provider not found", nil)
}

func (m *mockProviderRegistry) List(ctx context.Context) ([]engine.ProviderMetadata, error) {
	return []engine.ProviderMetadata{}, nil
}

func (m *mockProviderRegistry) Unregister(ctx context.Context, name, version string) error {
	return nil
}

func (m *mockProviderRegistry) ValidateCapabilities(ctx context.Context, capabilities []string) error {
	return nil
}

type mockStateManager struct {
	resources map[string]*engine.Resource
	states    map[string]json.RawMessage
	plans     map[string]*engine.Plan
	runs      map[string]*engine.Run
}

func newMockStateManager() *mockStateManager {
	return &mockStateManager{
		resources: make(map[string]*engine.Resource),
		states:    make(map[string]json.RawMessage),
		plans:     make(map[string]*engine.Plan),
		runs:      make(map[string]*engine.Run),
	}
}

func (m *mockStateManager) GetResource(ctx context.Context, resourceID string) (*engine.Resource, error) {
	if resource, exists := m.resources[resourceID]; exists {
		return resource, nil
	}
	return nil, engine.NewPermanentError("resource not found", nil)
}

func (m *mockStateManager) SaveResource(ctx context.Context, resource *engine.Resource) error {
	m.resources[resource.ID] = resource
	return nil
}

func (m *mockStateManager) DeleteResource(ctx context.Context, resourceID string) error {
	delete(m.resources, resourceID)
	return nil
}

func (m *mockStateManager) ListResources(ctx context.Context, selector map[string]string) ([]engine.Resource, error) {
	resources := make([]engine.Resource, 0, len(m.resources))
	for _, r := range m.resources {
		resources = append(resources, *r)
	}
	return resources, nil
}

func (m *mockStateManager) GetResourceState(ctx context.Context, resourceID string) (json.RawMessage, error) {
	if state, exists := m.states[resourceID]; exists {
		return state, nil
	}
	return nil, engine.NewPermanentError("resource state not found", nil)
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

func (m *mockStateManager) GetPlan(ctx context.Context, planID string) (*engine.Plan, error) {
	if plan, exists := m.plans[planID]; exists {
		return plan, nil
	}
	return nil, engine.NewPermanentError("plan not found", nil)
}

func (m *mockStateManager) SavePlan(ctx context.Context, plan *engine.Plan) error {
	m.plans[plan.ID] = plan
	return nil
}

func (m *mockStateManager) GetRun(ctx context.Context, runID string) (*engine.Run, error) {
	if run, exists := m.runs[runID]; exists {
		return run, nil
	}
	return nil, engine.NewPermanentError("run not found", nil)
}

func (m *mockStateManager) SaveRun(ctx context.Context, run *engine.Run) error {
	m.runs[run.ID] = run
	return nil
}

func (m *mockStateManager) AppendEvent(ctx context.Context, event *engine.Event) error {
	return nil
}

func (m *mockStateManager) GetEvents(ctx context.Context, runID string) ([]engine.Event, error) {
	return []engine.Event{}, nil
}
