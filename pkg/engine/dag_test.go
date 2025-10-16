package engine

import (
	"testing"
	"time"
)

func TestDAGBuilder_BuildGraph_EmptyUnits(t *testing.T) {
	builder := NewDAGBuilder()
	graph, err := builder.BuildGraph([]PlanUnit{})

	if err != nil {
		t.Fatalf("Expected no error for empty units, got: %v", err)
	}

	if len(graph.Nodes) != 0 {
		t.Errorf("Expected 0 nodes, got %d", len(graph.Nodes))
	}

	if len(graph.Edges) != 0 {
		t.Errorf("Expected 0 edges, got %d", len(graph.Edges))
	}

	if graph.Depth != 0 {
		t.Errorf("Expected depth 0, got %d", graph.Depth)
	}
}

func TestDAGBuilder_BuildGraph_SingleUnit(t *testing.T) {
	units := []PlanUnit{
		{
			ID:           "unit1",
			ResourceID:   "resource1",
			Operation:    OperationCreate,
			Status:       PlanStatusPending,
			Dependencies: []Dependency{},
			Timeout:      time.Minute,
			MaxRetries:   3,
		},
	}

	builder := NewDAGBuilder()
	graph, err := builder.BuildGraph(units)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(graph.Nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(graph.Nodes))
	}

	if len(graph.Roots) != 1 {
		t.Errorf("Expected 1 root, got %d", len(graph.Roots))
	}

	if graph.Depth != 1 {
		t.Errorf("Expected depth 1, got %d", graph.Depth)
	}

	node := graph.Nodes["unit1"]
	if node.Level != 0 {
		t.Errorf("Expected level 0, got %d", node.Level)
	}
}

func TestDAGBuilder_BuildGraph_LinearDependencies(t *testing.T) {
	units := []PlanUnit{
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
	}

	builder := NewDAGBuilder()
	graph, err := builder.BuildGraph(units)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(graph.Nodes) != 3 {
		t.Errorf("Expected 3 nodes, got %d", len(graph.Nodes))
	}

	if graph.Depth != 3 {
		t.Errorf("Expected depth 3, got %d", graph.Depth)
	}

	// Verify levels
	if graph.Nodes["unit1"].Level != 0 {
		t.Errorf("unit1 should be at level 0, got %d", graph.Nodes["unit1"].Level)
	}
	if graph.Nodes["unit2"].Level != 1 {
		t.Errorf("unit2 should be at level 1, got %d", graph.Nodes["unit2"].Level)
	}
	if graph.Nodes["unit3"].Level != 2 {
		t.Errorf("unit3 should be at level 2, got %d", graph.Nodes["unit3"].Level)
	}

	// Verify edges
	if len(graph.Edges) != 2 {
		t.Errorf("Expected 2 edges, got %d", len(graph.Edges))
	}
}

func TestDAGBuilder_BuildGraph_ParallelUnits(t *testing.T) {
	units := []PlanUnit{
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
	}

	builder := NewDAGBuilder()
	graph, err := builder.BuildGraph(units)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(graph.Nodes) != 3 {
		t.Errorf("Expected 3 nodes, got %d", len(graph.Nodes))
	}

	if graph.Depth != 1 {
		t.Errorf("Expected depth 1, got %d", graph.Depth)
	}

	// All units should be at level 0 (parallel)
	for _, unit := range units {
		if graph.Nodes[unit.ID].Level != 0 {
			t.Errorf("%s should be at level 0, got %d", unit.ID, graph.Nodes[unit.ID].Level)
		}
	}

	if len(graph.Roots) != 3 {
		t.Errorf("Expected 3 roots, got %d", len(graph.Roots))
	}
}

func TestDAGBuilder_BuildGraph_DiamondDependencies(t *testing.T) {
	// Diamond pattern: unit1 -> unit2,unit3 -> unit4
	units := []PlanUnit{
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
				{TargetID: "unit1", Type: DependencyRequire},
			},
			Timeout:    time.Minute,
			MaxRetries: 3,
		},
		{
			ID:         "unit4",
			ResourceID: "resource4",
			Operation:  OperationCreate,
			Status:     PlanStatusPending,
			Dependencies: []Dependency{
				{TargetID: "unit2", Type: DependencyRequire},
				{TargetID: "unit3", Type: DependencyRequire},
			},
			Timeout:    time.Minute,
			MaxRetries: 3,
		},
	}

	builder := NewDAGBuilder()
	graph, err := builder.BuildGraph(units)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(graph.Nodes) != 4 {
		t.Errorf("Expected 4 nodes, got %d", len(graph.Nodes))
	}

	if graph.Depth != 3 {
		t.Errorf("Expected depth 3, got %d", graph.Depth)
	}

	// Verify levels
	if graph.Nodes["unit1"].Level != 0 {
		t.Errorf("unit1 should be at level 0, got %d", graph.Nodes["unit1"].Level)
	}
	if graph.Nodes["unit2"].Level != 1 {
		t.Errorf("unit2 should be at level 1, got %d", graph.Nodes["unit2"].Level)
	}
	if graph.Nodes["unit3"].Level != 1 {
		t.Errorf("unit3 should be at level 1, got %d", graph.Nodes["unit3"].Level)
	}
	if graph.Nodes["unit4"].Level != 2 {
		t.Errorf("unit4 should be at level 2, got %d", graph.Nodes["unit4"].Level)
	}

	if len(graph.Edges) != 4 {
		t.Errorf("Expected 4 edges, got %d", len(graph.Edges))
	}
}

func TestDAGBuilder_DetectCycles_SimpleCycle(t *testing.T) {
	// Simple cycle: unit1 -> unit2 -> unit1
	units := []PlanUnit{
		{
			ID:         "unit1",
			ResourceID: "resource1",
			Operation:  OperationCreate,
			Status:     PlanStatusPending,
			Dependencies: []Dependency{
				{TargetID: "unit2", Type: DependencyRequire},
			},
			Timeout:    time.Minute,
			MaxRetries: 3,
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
	}

	builder := NewDAGBuilder()
	_, err := builder.BuildGraph(units)

	if err == nil {
		t.Fatal("Expected error for circular dependency, got nil")
	}

	if !IsPermanent(err) {
		t.Error("Expected permanent error for circular dependency")
	}
}

func TestDAGBuilder_DetectCycles_ComplexCycle(t *testing.T) {
	// Complex cycle: unit1 -> unit2 -> unit3 -> unit1
	units := []PlanUnit{
		{
			ID:         "unit1",
			ResourceID: "resource1",
			Operation:  OperationCreate,
			Status:     PlanStatusPending,
			Dependencies: []Dependency{
				{TargetID: "unit3", Type: DependencyRequire},
			},
			Timeout:    time.Minute,
			MaxRetries: 3,
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
	}

	builder := NewDAGBuilder()
	_, err := builder.BuildGraph(units)

	if err == nil {
		t.Fatal("Expected error for circular dependency, got nil")
	}
}

func TestDAGBuilder_InvalidDependency(t *testing.T) {
	units := []PlanUnit{
		{
			ID:         "unit1",
			ResourceID: "resource1",
			Operation:  OperationCreate,
			Status:     PlanStatusPending,
			Dependencies: []Dependency{
				{TargetID: "nonexistent", Type: DependencyRequire},
			},
			Timeout:    time.Minute,
			MaxRetries: 3,
		},
	}

	builder := NewDAGBuilder()
	_, err := builder.BuildGraph(units)

	if err == nil {
		t.Fatal("Expected error for invalid dependency, got nil")
	}
}

func TestDAGBuilder_DuplicateIDs(t *testing.T) {
	units := []PlanUnit{
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
			ID:           "unit1", // Duplicate ID
			ResourceID:   "resource2",
			Operation:    OperationCreate,
			Status:       PlanStatusPending,
			Dependencies: []Dependency{},
			Timeout:      time.Minute,
			MaxRetries:   3,
		},
	}

	builder := NewDAGBuilder()
	_, err := builder.BuildGraph(units)

	if err == nil {
		t.Fatal("Expected error for duplicate IDs, got nil")
	}
}

func TestDAGBuilder_ToDOT(t *testing.T) {
	units := []PlanUnit{
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
			Operation:  OperationUpdate,
			Status:     PlanStatusPending,
			Dependencies: []Dependency{
				{TargetID: "unit1", Type: DependencyRequire},
			},
			Timeout:    time.Minute,
			MaxRetries: 3,
		},
	}

	builder := NewDAGBuilder()
	_, err := builder.BuildGraph(units)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	dot := builder.ToDOT()

	// Check that DOT output contains expected elements
	if len(dot) == 0 {
		t.Error("Expected non-empty DOT output")
	}

	// Should contain digraph declaration
	if !contains(dot, "digraph ExecutionGraph") {
		t.Error("DOT output missing digraph declaration")
	}

	// Should contain nodes
	if !contains(dot, "unit1") || !contains(dot, "unit2") {
		t.Error("DOT output missing expected nodes")
	}

	// Should contain edge
	if !contains(dot, "->") {
		t.Error("DOT output missing edge")
	}
}

func TestDAGBuilder_DifferentDependencyTypes(t *testing.T) {
	units := []PlanUnit{
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
				{TargetID: "unit1", Type: DependencyNotify},
			},
			Timeout:    time.Minute,
			MaxRetries: 3,
		},
		{
			ID:         "unit4",
			ResourceID: "resource4",
			Operation:  OperationCreate,
			Status:     PlanStatusPending,
			Dependencies: []Dependency{
				{TargetID: "unit1", Type: DependencyOrder},
			},
			Timeout:    time.Minute,
			MaxRetries: 3,
		},
	}

	builder := NewDAGBuilder()
	graph, err := builder.BuildGraph(units)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify all dependency types are preserved in edges
	dependencyTypes := make(map[DependencyType]int)
	for _, edge := range graph.Edges {
		dependencyTypes[edge.Type]++
	}

	if dependencyTypes[DependencyRequire] != 1 {
		t.Errorf("Expected 1 require dependency, got %d", dependencyTypes[DependencyRequire])
	}
	if dependencyTypes[DependencyNotify] != 1 {
		t.Errorf("Expected 1 notify dependency, got %d", dependencyTypes[DependencyNotify])
	}
	if dependencyTypes[DependencyOrder] != 1 {
		t.Errorf("Expected 1 order dependency, got %d", dependencyTypes[DependencyOrder])
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
