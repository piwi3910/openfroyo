package engine

import (
	"fmt"
	"strings"
)

// DAGBuilder builds a directed acyclic graph (DAG) from plan units.
// It performs topological sorting and assigns execution levels for parallel execution.
type DAGBuilder struct {
	// units maps plan unit IDs to their plan units
	units map[string]*PlanUnit

	// adjacencyList maps unit IDs to their dependencies
	adjacencyList map[string][]string

	// reverseAdjacencyList maps unit IDs to their dependents
	reverseAdjacencyList map[string][]string

	// inDegree tracks the number of incoming edges for each node
	inDegree map[string]int

	// levels maps execution level to unit IDs at that level
	levels [][]string
}

// NewDAGBuilder creates a new DAG builder.
func NewDAGBuilder() *DAGBuilder {
	return &DAGBuilder{
		units:                make(map[string]*PlanUnit),
		adjacencyList:        make(map[string][]string),
		reverseAdjacencyList: make(map[string][]string),
		inDegree:             make(map[string]int),
		levels:               make([][]string, 0),
	}
}

// BuildGraph constructs an execution graph from plan units.
// It validates dependencies, detects cycles, and computes execution levels.
func (b *DAGBuilder) BuildGraph(units []PlanUnit) (*ExecutionGraph, error) {
	if len(units) == 0 {
		return &ExecutionGraph{
			Nodes: make(map[string]*GraphNode),
			Edges: make([]GraphEdge, 0),
			Roots: make([]string, 0),
			Depth: 0,
		}, nil
	}

	// Initialize the builder with units
	if err := b.initialize(units); err != nil {
		return nil, err
	}

	// Detect circular dependencies
	if err := b.detectCycles(); err != nil {
		return nil, err
	}

	// Compute topological levels for parallel execution
	if err := b.computeLevels(); err != nil {
		return nil, err
	}

	// Build the execution graph
	graph := b.buildExecutionGraph()

	return graph, nil
}

// initialize sets up the internal data structures from plan units.
func (b *DAGBuilder) initialize(units []PlanUnit) error {
	// First pass: index all units
	for i := range units {
		unit := &units[i]
		if unit.ID == "" {
			return NewPermanentError("plan unit has empty ID", nil).
				WithCode(ErrCodeValidation)
		}

		if _, exists := b.units[unit.ID]; exists {
			return NewPermanentError(fmt.Sprintf("duplicate plan unit ID: %s", unit.ID), nil).
				WithCode(ErrCodeValidation)
		}

		b.units[unit.ID] = unit
		b.adjacencyList[unit.ID] = make([]string, 0)
		b.reverseAdjacencyList[unit.ID] = make([]string, 0)
		b.inDegree[unit.ID] = 0
	}

	// Second pass: build adjacency lists and validate dependencies
	for _, unit := range b.units {
		for _, dep := range unit.Dependencies {
			targetID := dep.TargetID

			// Validate dependency target exists
			if _, exists := b.units[targetID]; !exists {
				return NewPermanentError(
					fmt.Sprintf("plan unit %s depends on non-existent unit %s", unit.ID, targetID),
					nil,
				).WithCode(ErrCodeValidation).WithResource(unit.ID)
			}

			// Add edge from dependency to unit
			// (dependency must complete before unit can start)
			b.adjacencyList[targetID] = append(b.adjacencyList[targetID], unit.ID)
			b.reverseAdjacencyList[unit.ID] = append(b.reverseAdjacencyList[unit.ID], targetID)
			b.inDegree[unit.ID]++
		}
	}

	return nil
}

// detectCycles uses depth-first search to detect circular dependencies.
func (b *DAGBuilder) detectCycles() error {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := make([]string, 0)

	for id := range b.units {
		if !visited[id] {
			if cycle, err := b.detectCyclesUtil(id, visited, recStack, path); err != nil {
				return NewPermanentError(
					fmt.Sprintf("circular dependency detected: %s", formatCycle(cycle)),
					err,
				).WithCode(ErrCodeValidation)
			}
		}
	}

	return nil
}

// detectCyclesUtil performs DFS to detect cycles in the dependency graph.
func (b *DAGBuilder) detectCyclesUtil(
	nodeID string,
	visited map[string]bool,
	recStack map[string]bool,
	path []string,
) ([]string, error) {
	visited[nodeID] = true
	recStack[nodeID] = true
	path = append(path, nodeID)

	// Visit all dependents (units that depend on this node)
	for _, dependent := range b.adjacencyList[nodeID] {
		if !visited[dependent] {
			if cycle, err := b.detectCyclesUtil(dependent, visited, recStack, path); err != nil {
				return cycle, err
			}
		} else if recStack[dependent] {
			// Found a cycle - construct the cycle path
			cycleStart := -1
			for i, id := range path {
				if id == dependent {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				return append(path[cycleStart:], dependent), fmt.Errorf("cycle detected")
			}
		}
	}

	recStack[nodeID] = false
	return nil, nil
}

// computeLevels assigns execution levels to each unit using topological sort.
// Units at the same level can be executed in parallel.
func (b *DAGBuilder) computeLevels() error {
	// Use Kahn's algorithm for topological sorting with level tracking
	inDegreeCopy := make(map[string]int)
	for id, degree := range b.inDegree {
		inDegreeCopy[id] = degree
	}

	// Find all root nodes (nodes with no dependencies)
	currentLevel := make([]string, 0)
	for id, degree := range inDegreeCopy {
		if degree == 0 {
			currentLevel = append(currentLevel, id)
		}
	}

	if len(currentLevel) == 0 && len(b.units) > 0 {
		return NewPermanentError("no root nodes found - all units have dependencies", nil).
			WithCode(ErrCodeValidation)
	}

	// Process nodes level by level
	processedCount := 0
	for len(currentLevel) > 0 {
		// Add current level to levels
		b.levels = append(b.levels, currentLevel)
		processedCount += len(currentLevel)

		// Find next level
		nextLevel := make([]string, 0)
		for _, nodeID := range currentLevel {
			// Process all dependents of this node
			for _, dependent := range b.adjacencyList[nodeID] {
				inDegreeCopy[dependent]--
				if inDegreeCopy[dependent] == 0 {
					nextLevel = append(nextLevel, dependent)
				}
			}
		}

		currentLevel = nextLevel
	}

	// Verify all nodes were processed (should never happen if cycle detection worked)
	if processedCount != len(b.units) {
		return NewPermanentError("failed to process all units - possible cycle", nil).
			WithCode(ErrCodeInternal)
	}

	return nil
}

// buildExecutionGraph creates the final ExecutionGraph structure.
func (b *DAGBuilder) buildExecutionGraph() *ExecutionGraph {
	graph := &ExecutionGraph{
		Nodes: make(map[string]*GraphNode),
		Edges: make([]GraphEdge, 0),
		Roots: make([]string, 0),
		Depth: len(b.levels),
	}

	// Build nodes with their levels
	for level, unitIDs := range b.levels {
		for _, unitID := range unitIDs {
			unit := b.units[unitID]
			node := &GraphNode{
				ID:           unitID,
				Level:        level,
				Dependencies: b.reverseAdjacencyList[unitID],
				Dependents:   b.adjacencyList[unitID],
			}
			graph.Nodes[unitID] = node

			// Set execution order in the original unit
			unit.ExecutionOrder = level

			// Track root nodes
			if level == 0 {
				graph.Roots = append(graph.Roots, unitID)
			}
		}
	}

	// Build edges from dependencies
	for _, unit := range b.units {
		for _, dep := range unit.Dependencies {
			edge := GraphEdge{
				From: dep.TargetID,
				To:   unit.ID,
				Type: dep.Type,
			}
			graph.Edges = append(graph.Edges, edge)
		}
	}

	return graph
}

// GetLevels returns the computed execution levels.
// Each level contains unit IDs that can be executed in parallel.
func (b *DAGBuilder) GetLevels() [][]string {
	return b.levels
}

// ToDOT generates a DOT format representation of the DAG for visualization.
// The output can be rendered with Graphviz tools.
func (b *DAGBuilder) ToDOT() string {
	var sb strings.Builder

	sb.WriteString("digraph ExecutionGraph {\n")
	sb.WriteString("  rankdir=TB;\n")
	sb.WriteString("  node [shape=box, style=rounded];\n\n")

	// Group nodes by level for better visualization
	for level, unitIDs := range b.levels {
		sb.WriteString(fmt.Sprintf("  subgraph cluster_level_%d {\n", level))
		sb.WriteString(fmt.Sprintf("    label=\"Level %d\";\n", level))
		sb.WriteString("    style=dashed;\n")

		for _, unitID := range unitIDs {
			unit := b.units[unitID]
			label := fmt.Sprintf("%s\\n%s", unit.ResourceID, unit.Operation)
			color := getOperationColor(unit.Operation)

			sb.WriteString(fmt.Sprintf("    \"%s\" [label=\"%s\", fillcolor=\"%s\", style=\"filled,rounded\"];\n",
				unitID, label, color))
		}

		sb.WriteString("  }\n\n")
	}

	// Add edges with dependency types
	for _, unit := range b.units {
		for _, dep := range unit.Dependencies {
			style := getDependencyStyle(dep.Type)
			sb.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\" [%s];\n",
				dep.TargetID, unit.ID, style))
		}
	}

	sb.WriteString("}\n")
	return sb.String()
}

// formatCycle formats a cycle path for error messages.
func formatCycle(cycle []string) string {
	if len(cycle) == 0 {
		return ""
	}
	return strings.Join(cycle, " -> ")
}

// getOperationColor returns a color for visualizing operation types.
func getOperationColor(op OperationType) string {
	switch op {
	case OperationCreate:
		return "lightgreen"
	case OperationUpdate:
		return "lightblue"
	case OperationDelete, OperationRecreate:
		return "lightcoral"
	case OperationNoop:
		return "lightgray"
	default:
		return "white"
	}
}

// getDependencyStyle returns a DOT style string for dependency types.
func getDependencyStyle(depType DependencyType) string {
	switch depType {
	case DependencyRequire:
		return "style=solid, color=black"
	case DependencyNotify:
		return "style=dashed, color=blue"
	case DependencyOrder:
		return "style=dotted, color=gray"
	default:
		return "style=solid, color=black"
	}
}

// ValidateGraph performs additional validation on the built graph.
func (b *DAGBuilder) ValidateGraph(graph *ExecutionGraph) error {
	// Verify all units are represented in the graph
	if len(graph.Nodes) != len(b.units) {
		return NewPermanentError("graph node count mismatch", nil).
			WithCode(ErrCodeInternal)
	}

	// Verify all edges are valid
	for _, edge := range graph.Edges {
		if _, exists := graph.Nodes[edge.From]; !exists {
			return NewPermanentError(fmt.Sprintf("edge references non-existent node: %s", edge.From), nil).
				WithCode(ErrCodeInternal)
		}
		if _, exists := graph.Nodes[edge.To]; !exists {
			return NewPermanentError(fmt.Sprintf("edge references non-existent node: %s", edge.To), nil).
				WithCode(ErrCodeInternal)
		}
	}

	// Verify root nodes have no dependencies
	for _, rootID := range graph.Roots {
		node := graph.Nodes[rootID]
		if len(node.Dependencies) > 0 {
			return NewPermanentError(fmt.Sprintf("root node %s has dependencies", rootID), nil).
				WithCode(ErrCodeInternal)
		}
	}

	return nil
}
