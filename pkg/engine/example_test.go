package engine_test

import (
	"encoding/json"
	"time"

	"github.com/openfroyo/openfroyo/pkg/engine"
)

// ExampleWorkflow demonstrates how the core types compose together
// in a typical OpenFroyo execution workflow.
func Example_workflow() {
	// 1. Define a resource configuration
	resource := engine.Resource{
		ID:      "apache-pkg-001",
		Type:    "linux.pkg",
		Name:    "apache2",
		Config:  json.RawMessage(`{"name":"apache2","version":"2.4.52","state":"present"}`),
		Status:  engine.ResourceStatusUnknown,
		Labels:  map[string]string{"role": "web", "env": "production"},
		Version: 1,
	}

	// 2. Create a plan unit for the resource
	planUnit := engine.PlanUnit{
		ID:              "pu-001",
		ResourceID:      resource.ID,
		Operation:       engine.OperationCreate,
		Status:          engine.PlanStatusPending,
		DesiredState:    resource.Config,
		ProviderName:    "linux.pkg",
		ProviderVersion: "1.0.0",
		ExecutionOrder:  1,
		MaxRetries:      3,
		Timeout:         5 * time.Minute,
		Changes: []engine.Change{
			{
				Path:   ".config.state",
				Before: nil,
				After:  "present",
				Action: engine.ChangeActionAdd,
			},
		},
	}

	// 3. Add dependencies between plan units
	dependency := engine.Dependency{
		TargetID: "pu-000", // depends on repository update
		Type:     engine.DependencyRequire,
	}
	planUnit.Dependencies = append(planUnit.Dependencies, dependency)

	// 4. Build an execution graph
	graph := engine.ExecutionGraph{
		Nodes: map[string]*engine.GraphNode{
			"pu-001": {
				ID:           "pu-001",
				Level:        1,
				Dependencies: []string{"pu-000"},
				Dependents:   []string{},
			},
		},
		Edges: []engine.GraphEdge{
			{
				From: "pu-000",
				To:   "pu-001",
				Type: engine.DependencyRequire,
			},
		},
		Roots: []string{"pu-000"},
		Depth: 2,
	}

	// 5. Create a complete plan
	plan := engine.Plan{
		ID:        "plan-001",
		RunID:     "run-001",
		CreatedAt: time.Now(),
		Units:     []engine.PlanUnit{planUnit},
		Graph:     &graph,
		Summary: engine.PlanSummary{
			TotalResources: 1,
			ToCreate:       1,
			ToUpdate:       0,
			ToDelete:       0,
			NoChange:       0,
		},
	}

	// 6. Execute and capture results
	result := engine.ExecutionResult{
		PlanUnitID:  "pu-001",
		Status:      engine.PlanStatusSucceeded,
		StartedAt:   time.Now(),
		CompletedAt: time.Now().Add(30 * time.Second),
		Duration:    30 * time.Second,
		NewState:    json.RawMessage(`{"installed":true,"version":"2.4.52"}`),
		Events: []engine.Event{
			{
				ID:         "evt-001",
				Type:       engine.EventTypePlanUnitStarted,
				Timestamp:  time.Now(),
				RunID:      "run-001",
				PlanUnitID: "pu-001",
				ResourceID: "apache-pkg-001",
				Message:    "Installing apache2 package",
				Level:      "info",
			},
		},
	}

	// 7. Track the overall run
	run := engine.Run{
		ID:        "run-001",
		PlanID:    "plan-001",
		Status:    engine.RunStatusSucceeded,
		StartedAt: time.Now(),
		Summary: engine.RunSummary{
			Total:     1,
			Succeeded: 1,
			Failed:    0,
			Skipped:   0,
		},
	}

	// 8. Detect drift after execution
	drift := engine.DriftDetection{
		ResourceID:   resource.ID,
		Status:       engine.DriftStatusInSync,
		DetectedAt:   time.Now(),
		DesiredState: resource.Config,
		ActualState:  result.NewState,
		Reconciled:   true,
	}

	// 9. Handle errors with classification
	if result.Error != nil {
		if engine.IsTransient(result.Error) {
			// Retry the operation with exponential backoff
			_ = result.Error
		} else if engine.IsPermanent(result.Error) {
			// Log and fail the run
			_ = result.Error
		}
	}

	// Types compose cleanly: Resource -> PlanUnit -> Plan -> Run -> Result
	_, _, _, _, _ = resource, plan, run, result, drift
}

// ExampleProvider demonstrates the provider interface contract.
func Example_provider() {
	// Provider request/response cycle
	readReq := engine.ReadRequest{
		ResourceID: "apache-pkg-001",
		Config:     json.RawMessage(`{"name":"apache2"}`),
	}

	planReq := engine.PlanRequest{
		ResourceID:   "apache-pkg-001",
		DesiredState: json.RawMessage(`{"name":"apache2","state":"present"}`),
		ActualState:  json.RawMessage(`{"installed":false}`),
		Operation:    engine.OperationCreate,
	}

	applyReq := engine.ApplyRequest{
		ResourceID:   "apache-pkg-001",
		DesiredState: json.RawMessage(`{"name":"apache2","state":"present"}`),
		Operation:    engine.OperationCreate,
		PlannedChanges: []engine.Change{
			{
				Path:   ".installed",
				Before: false,
				After:  true,
				Action: engine.ChangeActionModify,
			},
		},
	}

	// Provider would process these and return responses
	_, _, _ = readReq, planReq, applyReq
}

// ExampleErrorHandling demonstrates error classification and handling.
func Example_errorHandling() {
	// Create different error types
	transientErr := engine.NewTransientError("network timeout", nil).
		WithResource("apache-pkg-001").
		WithOperation("install")

	permanentErr := engine.NewPermanentError("package not found", nil).
		WithCode(engine.ErrCodeNotFound).
		WithDetail("package", "nonexistent-pkg")

	// Check error classification
	canRetry := engine.IsRetryable(transientErr)     // true - transient errors are retryable
	cannotRetry := !engine.IsRetryable(permanentErr) // true - permanent errors are not retryable

	_, _, _ = transientErr, permanentErr, canRetry && cannotRetry
}

// ExampleStatusValidation demonstrates status enum validation.
func Example_statusValidation() {
	// Validate status enums
	status := engine.RunStatusRunning
	isValid := status.Validate() == nil // Status is valid

	// Check status properties
	isActive := status.IsActive()         // Status is pending or running
	isNotTerminal := !status.IsTerminal() // Status has not reached a final state

	// Check operation properties
	op := engine.OperationDelete
	requiresConfirmation := op.IsDestructive() // Confirm with user before proceeding

	_, _, _, _ = isValid, isActive, isNotTerminal, requiresConfirmation
}
