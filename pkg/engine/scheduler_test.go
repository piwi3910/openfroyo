package engine

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"
)

// Mock executor for testing
type mockExecutor struct {
	mu             sync.Mutex
	executionDelay time.Duration
	failUnits      map[string]bool
	executedUnits  []string
}

func newMockExecutor() *mockExecutor {
	return &mockExecutor{
		executionDelay: 10 * time.Millisecond,
		failUnits:      make(map[string]bool),
		executedUnits:  make([]string, 0),
	}
}

func (m *mockExecutor) Execute(ctx context.Context, plan *Plan) (*Run, error) {
	return nil, nil
}

func (m *mockExecutor) ExecuteUnit(ctx context.Context, unit *PlanUnit) (*ExecutionResult, error) {
	m.mu.Lock()
	m.executedUnits = append(m.executedUnits, unit.ID)
	shouldFail := m.failUnits[unit.ID]
	m.mu.Unlock()

	// Simulate execution time
	select {
	case <-time.After(m.executionDelay):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	result := &ExecutionResult{
		PlanUnitID:  unit.ID,
		StartedAt:   time.Now(),
		CompletedAt: time.Now(),
		Duration:    m.executionDelay,
		NewState:    unit.DesiredState,
	}

	if shouldFail {
		result.Status = PlanStatusFailed
		result.Error = NewTransientError("mock failure", nil)
		return result, NewTransientError("mock failure", nil)
	}

	result.Status = PlanStatusSucceeded
	return result, nil
}

func (m *mockExecutor) Cancel(ctx context.Context, runID string) error {
	return nil
}

func (m *mockExecutor) GetRunStatus(ctx context.Context, runID string) (*Run, error) {
	return nil, nil
}

func (m *mockExecutor) StreamEvents(ctx context.Context, runID string) (<-chan Event, error) {
	ch := make(chan Event)
	close(ch)
	return ch, nil
}

// Mock event publisher for testing
type mockEventPublisher struct {
	mu     sync.Mutex
	events []Event
}

func newMockEventPublisher() *mockEventPublisher {
	return &mockEventPublisher{
		events: make([]Event, 0),
	}
}

func (m *mockEventPublisher) Publish(ctx context.Context, event *Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, *event)
	return nil
}

func (m *mockEventPublisher) Subscribe(ctx context.Context, filter EventFilter) (<-chan Event, error) {
	ch := make(chan Event)
	close(ch)
	return ch, nil
}

func (m *mockEventPublisher) Unsubscribe(ctx context.Context, subscriptionID string) error {
	return nil
}

func (m *mockEventPublisher) getEvents() []Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]Event{}, m.events...)
}

func TestNewParallelScheduler(t *testing.T) {
	executor := newMockExecutor()
	publisher := newMockEventPublisher()
	stateMgr := newMockStateManager()

	scheduler := NewParallelScheduler(5, executor, publisher, stateMgr)

	if scheduler == nil {
		t.Fatal("Expected non-nil scheduler")
	}

	if scheduler.maxParallel != 5 {
		t.Errorf("Expected maxParallel=5, got %d", scheduler.maxParallel)
	}
}

func TestNewParallelScheduler_DefaultMaxParallel(t *testing.T) {
	executor := newMockExecutor()
	publisher := newMockEventPublisher()
	stateMgr := newMockStateManager()

	scheduler := NewParallelScheduler(0, executor, publisher, stateMgr)

	if scheduler.maxParallel != 10 {
		t.Errorf("Expected default maxParallel=10, got %d", scheduler.maxParallel)
	}
}

func TestScheduler_Schedule_NilPlan(t *testing.T) {
	executor := newMockExecutor()
	publisher := newMockEventPublisher()
	stateMgr := newMockStateManager()
	scheduler := NewParallelScheduler(5, executor, publisher, stateMgr)

	ctx := context.Background()
	_, err := scheduler.Schedule(ctx, nil, ScheduleOptions{})

	if err == nil {
		t.Fatal("Expected error for nil plan, got nil")
	}
}

func TestScheduler_Schedule_PlanWithoutGraph(t *testing.T) {
	executor := newMockExecutor()
	publisher := newMockEventPublisher()
	stateMgr := newMockStateManager()
	scheduler := NewParallelScheduler(5, executor, publisher, stateMgr)

	ctx := context.Background()

	plan := &Plan{
		ID:        "plan1",
		CreatedAt: time.Now(),
		Units:     []PlanUnit{},
		Graph:     nil, // No graph
	}

	_, err := scheduler.Schedule(ctx, plan, ScheduleOptions{})

	if err == nil {
		t.Fatal("Expected error for plan without graph, got nil")
	}
}

func TestScheduler_Schedule_SingleUnit(t *testing.T) {
	executor := newMockExecutor()
	publisher := newMockEventPublisher()
	stateMgr := newMockStateManager()
	scheduler := NewParallelScheduler(5, executor, publisher, stateMgr)

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
				DesiredState: json.RawMessage(`{"key": "value"}`),
				Timeout:      time.Minute,
				MaxRetries:   3,
			},
		},
		Graph: &ExecutionGraph{
			Nodes: map[string]*GraphNode{
				"unit1": {
					ID:           "unit1",
					Level:        0,
					Dependencies: []string{},
					Dependents:   []string{},
				},
			},
			Edges: []GraphEdge{},
			Roots: []string{"unit1"},
			Depth: 1,
		},
	}

	runID, err := scheduler.Schedule(ctx, plan, ScheduleOptions{})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if runID == "" {
		t.Error("Expected non-empty run ID")
	}

	// Wait for execution to complete
	time.Sleep(100 * time.Millisecond)

	// Verify run was saved
	run, err := stateMgr.GetRun(ctx, runID)
	if err != nil {
		t.Fatalf("Failed to get run: %v", err)
	}

	if run.PlanID != plan.ID {
		t.Errorf("Expected plan ID %s, got %s", plan.ID, run.PlanID)
	}
}

func TestScheduler_Schedule_LinearDependencies(t *testing.T) {
	executor := newMockExecutor()
	publisher := newMockEventPublisher()
	stateMgr := newMockStateManager()
	scheduler := NewParallelScheduler(5, executor, publisher, stateMgr)

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
			{
				ID:         "unit3",
				ResourceID: "resource3",
				Operation:  OperationCreate,
				Status:     PlanStatusPending,
				Dependencies: []Dependency{
					{TargetID: "unit2", Type: DependencyRequire},
				},
				Timeout:    time.Minute,
				MaxRetries: 3,
			},
		},
		Graph: &ExecutionGraph{
			Nodes: map[string]*GraphNode{
				"unit1": {ID: "unit1", Level: 0, Dependencies: []string{}, Dependents: []string{"unit2"}},
				"unit2": {ID: "unit2", Level: 1, Dependencies: []string{"unit1"}, Dependents: []string{"unit3"}},
				"unit3": {ID: "unit3", Level: 2, Dependencies: []string{"unit2"}, Dependents: []string{}},
			},
			Edges: []GraphEdge{
				{From: "unit1", To: "unit2", Type: DependencyRequire},
				{From: "unit2", To: "unit3", Type: DependencyRequire},
			},
			Roots: []string{"unit1"},
			Depth: 3,
		},
	}

	runID, err := scheduler.Schedule(ctx, plan, ScheduleOptions{})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Wait for execution to complete
	time.Sleep(200 * time.Millisecond)

	// Verify execution order (should be sequential)
	executor.mu.Lock()
	executedUnits := executor.executedUnits
	executor.mu.Unlock()

	if len(executedUnits) != 3 {
		t.Fatalf("Expected 3 executed units, got %d", len(executedUnits))
	}

	// Verify order: unit1 -> unit2 -> unit3
	if executedUnits[0] != "unit1" {
		t.Errorf("Expected unit1 first, got %s", executedUnits[0])
	}
	if executedUnits[1] != "unit2" {
		t.Errorf("Expected unit2 second, got %s", executedUnits[1])
	}
	if executedUnits[2] != "unit3" {
		t.Errorf("Expected unit3 third, got %s", executedUnits[2])
	}

	// Verify run completed successfully
	run, err := stateMgr.GetRun(ctx, runID)
	if err != nil {
		t.Fatalf("Failed to get run: %v", err)
	}

	if run.Status != RunStatusSucceeded {
		t.Errorf("Expected run status SUCCEEDED, got %s", run.Status)
	}

	if run.Summary.Succeeded != 3 {
		t.Errorf("Expected 3 succeeded units, got %d", run.Summary.Succeeded)
	}
}

func TestScheduler_Schedule_ParallelExecution(t *testing.T) {
	executor := newMockExecutor()
	executor.executionDelay = 50 * time.Millisecond
	publisher := newMockEventPublisher()
	stateMgr := newMockStateManager()
	scheduler := NewParallelScheduler(5, executor, publisher, stateMgr)

	ctx := context.Background()

	// Three parallel units at level 0
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
				ID:           "unit2",
				ResourceID:   "resource2",
				Operation:    OperationCreate,
				Status:       PlanStatusPending,
				Dependencies: []Dependency{},
				Timeout:      time.Minute,
				MaxRetries:   3,
			},
			{
				ID:           "unit3",
				ResourceID:   "resource3",
				Operation:    OperationCreate,
				Status:       PlanStatusPending,
				Dependencies: []Dependency{},
				Timeout:      time.Minute,
				MaxRetries:   3,
			},
		},
		Graph: &ExecutionGraph{
			Nodes: map[string]*GraphNode{
				"unit1": {ID: "unit1", Level: 0, Dependencies: []string{}, Dependents: []string{}},
				"unit2": {ID: "unit2", Level: 0, Dependencies: []string{}, Dependents: []string{}},
				"unit3": {ID: "unit3", Level: 0, Dependencies: []string{}, Dependents: []string{}},
			},
			Edges: []GraphEdge{},
			Roots: []string{"unit1", "unit2", "unit3"},
			Depth: 1,
		},
	}

	startTime := time.Now()
	runID, err := scheduler.Schedule(ctx, plan, ScheduleOptions{})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Wait for execution to complete
	time.Sleep(200 * time.Millisecond)
	duration := time.Since(startTime)

	// With parallel execution, should take about 50ms (one execution delay)
	// rather than 150ms (three sequential executions)
	// Allow some overhead for goroutine scheduling and synchronization
	if duration > 250*time.Millisecond {
		t.Errorf("Execution took too long (%v), expected parallel execution", duration)
	}

	// Verify all units executed
	executor.mu.Lock()
	executedCount := len(executor.executedUnits)
	executor.mu.Unlock()

	if executedCount != 3 {
		t.Errorf("Expected 3 executed units, got %d", executedCount)
	}

	// Verify run completed successfully
	run, err := stateMgr.GetRun(ctx, runID)
	if err != nil {
		t.Fatalf("Failed to get run: %v", err)
	}

	if run.Status != RunStatusSucceeded {
		t.Errorf("Expected run status SUCCEEDED, got %s", run.Status)
	}
}

func TestScheduler_Schedule_FailedUnit(t *testing.T) {
	executor := newMockExecutor()
	executor.failUnits["unit2"] = true // Make unit2 fail
	publisher := newMockEventPublisher()
	stateMgr := newMockStateManager()
	scheduler := NewParallelScheduler(5, executor, publisher, stateMgr)

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
				MaxRetries:   0, // No retries
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
				MaxRetries: 0, // No retries
			},
			{
				ID:         "unit3",
				ResourceID: "resource3",
				Operation:  OperationCreate,
				Status:     PlanStatusPending,
				Dependencies: []Dependency{
					{TargetID: "unit2", Type: DependencyRequire},
				},
				Timeout:    time.Minute,
				MaxRetries: 0,
			},
		},
		Graph: &ExecutionGraph{
			Nodes: map[string]*GraphNode{
				"unit1": {ID: "unit1", Level: 0, Dependencies: []string{}, Dependents: []string{"unit2"}},
				"unit2": {ID: "unit2", Level: 1, Dependencies: []string{"unit1"}, Dependents: []string{"unit3"}},
				"unit3": {ID: "unit3", Level: 2, Dependencies: []string{"unit2"}, Dependents: []string{}},
			},
			Edges: []GraphEdge{
				{From: "unit1", To: "unit2", Type: DependencyRequire},
				{From: "unit2", To: "unit3", Type: DependencyRequire},
			},
			Roots: []string{"unit1"},
			Depth: 3,
		},
	}

	runID, err := scheduler.Schedule(ctx, plan, ScheduleOptions{})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Wait for execution to complete
	time.Sleep(200 * time.Millisecond)

	// Verify run completed with failures
	run, err := stateMgr.GetRun(ctx, runID)
	if err != nil {
		t.Fatalf("Failed to get run: %v", err)
	}

	if run.Status == RunStatusSucceeded {
		t.Error("Expected run to fail, got SUCCEEDED")
	}

	// unit1 should succeed, unit2 should fail, unit3 should be skipped
	if run.Summary.Succeeded != 1 {
		t.Errorf("Expected 1 succeeded unit, got %d", run.Summary.Succeeded)
	}

	if run.Summary.Failed != 1 {
		t.Errorf("Expected 1 failed unit, got %d", run.Summary.Failed)
	}

	if run.Summary.Skipped != 1 {
		t.Errorf("Expected 1 skipped unit, got %d", run.Summary.Skipped)
	}
}

func TestScheduler_Schedule_DryRun(t *testing.T) {
	executor := newMockExecutor()
	publisher := newMockEventPublisher()
	stateMgr := newMockStateManager()
	scheduler := NewParallelScheduler(5, executor, publisher, stateMgr)

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
				DesiredState: json.RawMessage(`{"key": "value"}`),
				Timeout:      time.Minute,
				MaxRetries:   3,
			},
		},
		Graph: &ExecutionGraph{
			Nodes: map[string]*GraphNode{
				"unit1": {ID: "unit1", Level: 0, Dependencies: []string{}, Dependents: []string{}},
			},
			Edges: []GraphEdge{},
			Roots: []string{"unit1"},
			Depth: 1,
		},
	}

	runID, err := scheduler.Schedule(ctx, plan, ScheduleOptions{DryRun: true})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Wait for execution to complete
	time.Sleep(100 * time.Millisecond)

	// Verify executor was not called (dry run)
	executor.mu.Lock()
	executedCount := len(executor.executedUnits)
	executor.mu.Unlock()

	if executedCount != 0 {
		t.Errorf("Expected 0 executed units in dry run, got %d", executedCount)
	}

	// Verify run completed successfully
	run, err := stateMgr.GetRun(ctx, runID)
	if err != nil {
		t.Fatalf("Failed to get run: %v", err)
	}

	if run.Status != RunStatusSucceeded {
		t.Errorf("Expected run status SUCCEEDED, got %s", run.Status)
	}
}

func TestScheduler_Schedule_WithDelay(t *testing.T) {
	executor := newMockExecutor()
	publisher := newMockEventPublisher()
	stateMgr := newMockStateManager()
	scheduler := NewParallelScheduler(5, executor, publisher, stateMgr)

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
				DesiredState: json.RawMessage(`{"key": "value"}`),
				Timeout:      time.Minute,
				MaxRetries:   3,
			},
		},
		Graph: &ExecutionGraph{
			Nodes: map[string]*GraphNode{
				"unit1": {ID: "unit1", Level: 0, Dependencies: []string{}, Dependents: []string{}},
			},
			Edges: []GraphEdge{},
			Roots: []string{"unit1"},
			Depth: 1,
		},
	}

	delay := 50 * time.Millisecond
	startTime := time.Now()

	runID, err := scheduler.Schedule(ctx, plan, ScheduleOptions{Delay: delay})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Wait for delay + execution to complete
	time.Sleep(150 * time.Millisecond)
	duration := time.Since(startTime)

	// Verify delay was applied
	if duration < delay {
		t.Errorf("Expected delay of at least %v, got %v", delay, duration)
	}

	// Verify run was saved
	_, err = stateMgr.GetRun(ctx, runID)
	if err != nil {
		t.Fatalf("Failed to get run: %v", err)
	}
}

func TestScheduler_GetStatus(t *testing.T) {
	executor := newMockExecutor()
	publisher := newMockEventPublisher()
	stateMgr := newMockStateManager()
	scheduler := NewParallelScheduler(5, executor, publisher, stateMgr)

	ctx := context.Background()

	// Save a test run
	testRun := &Run{
		ID:        "run1",
		PlanID:    "plan1",
		Status:    RunStatusRunning,
		StartedAt: time.Now(),
	}
	stateMgr.SaveRun(ctx, testRun)

	// Get status
	run, err := scheduler.GetStatus(ctx, "run1")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if run.ID != "run1" {
		t.Errorf("Expected run ID 'run1', got %s", run.ID)
	}

	if run.Status != RunStatusRunning {
		t.Errorf("Expected status RUNNING, got %s", run.Status)
	}
}

func TestScheduler_GetStatus_NotFound(t *testing.T) {
	executor := newMockExecutor()
	publisher := newMockEventPublisher()
	stateMgr := newMockStateManager()
	scheduler := NewParallelScheduler(5, executor, publisher, stateMgr)

	ctx := context.Background()

	_, err := scheduler.GetStatus(ctx, "nonexistent")

	if err == nil {
		t.Fatal("Expected error for non-existent run, got nil")
	}
}
