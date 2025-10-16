// Package engine provides the core types and interfaces for the OpenFroyo orchestration engine.
//
// # Overview
//
// OpenFroyo is an Infrastructure-as-Code engine that combines declarative state management
// with procedural automation capabilities. The engine operates through a 6-phase workflow:
//
//  1. Config - Parse and validate CUE configurations (Evaluator)
//  2. Facts - Discover current system state (Discoverer)
//  3. Plan - Compute diffs and build execution DAG (Planner)
//  4. Apply - Execute the DAG to achieve desired state (Executor)
//  5. Result - Capture execution outcomes (ExecutionResult)
//  6. Drift - Detect and reconcile configuration drift (DriftDetector)
//
// # Core Domain Types
//
// The package defines several fundamental types that represent the execution model:
//
//   - Resource: A managed infrastructure resource with desired and actual state
//   - PlanUnit: A unit of work in the execution DAG
//   - Dependency: An edge in the execution graph (require/notify/order)
//   - Operation: The type of operation to perform (create/update/delete/noop)
//   - ExecutionResult: The outcome of executing a plan unit
//   - Event: Timeline events during execution
//   - Plan: A complete execution plan with DAG
//   - Run: An execution of a plan with status tracking
//
// # Provider Interface
//
// Providers implement resource management through the Provider interface:
//
//	type Provider interface {
//	    Init(ctx context.Context, config ProviderConfig) error
//	    Read(ctx context.Context, req ReadRequest) (*ReadResponse, error)
//	    Plan(ctx context.Context, req PlanRequest) (*PlanResponse, error)
//	    Apply(ctx context.Context, req ApplyRequest) (*ApplyResponse, error)
//	    Destroy(ctx context.Context, req DestroyRequest) (*DestroyResponse, error)
//	}
//
// Providers are loaded as WASM modules with declared capabilities and schemas.
//
// # Workflow Interfaces
//
// The engine workflow is defined through specialized interfaces:
//
//   - Evaluator: Parses CUE configs and executes Starlark scripts
//   - Discoverer: Collects facts from target systems
//   - Planner: Computes diffs and builds execution plans
//   - Executor: Executes plans by running the DAG
//   - StateManager: Persists resource state and run history
//   - DriftDetector: Detects and reconciles configuration drift
//   - PolicyEngine: Enforces OPA policies on configs and plans
//
// # Error Classification
//
// Errors are classified for intelligent retry logic:
//
//   - Transient: Temporary failures that may succeed on retry
//   - Throttled: Rate limiting that requires backoff
//   - Conflict: Resource conflicts requiring retry
//   - Permanent: Non-recoverable errors
//
// Use the error helper functions to classify and inspect errors:
//
//	if IsTransient(err) {
//	    // Retry the operation
//	}
//
// # Status Tracking
//
// The package provides comprehensive status tracking:
//
//   - RunStatus: Overall execution status (pending/running/succeeded/failed)
//   - OperationType: What operation to perform (create/update/delete/noop)
//   - ResourceStatus: Resource lifecycle state (creating/ready/updating/error)
//   - PlanStatus: Plan unit execution state (pending/running/succeeded/failed)
//   - DriftStatus: Drift detection result (in_sync/drifted/unknown)
//
// # Example Usage
//
// Basic workflow for executing a configuration:
//
//	// 1. Evaluate configuration
//	config, err := evaluator.Evaluate(ctx, sources)
//
//	// 2. Discover facts
//	facts, err := discoverer.DiscoverFacts(ctx, targetID, namespaces)
//
//	// 3. Build plan
//	diff, err := planner.ComputeDiff(ctx, config, facts)
//	plan, err := planner.BuildPlan(ctx, diff)
//	graph, err := planner.BuildDAG(ctx, plan)
//
//	// 4. Execute plan
//	run, err := executor.Execute(ctx, plan)
//
//	// 5. Check results
//	if run.Status == RunStatusSucceeded {
//	    // Success
//	}
//
// # Thread Safety
//
// All interfaces are designed to be safe for concurrent use. Implementations
// must ensure proper synchronization when accessing shared state.
//
// # Immutability
//
// Core types like Resource, PlanUnit, and ExecutionResult are designed to be
// immutable value objects. Modifications create new instances rather than
// mutating existing ones.
package engine
