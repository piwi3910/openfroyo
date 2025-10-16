package engine

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ParallelScheduler implements parallel execution of plan units with dependency management.
// It executes plan units level-by-level, running independent units in parallel within each level.
type ParallelScheduler struct {
	// maxParallel is the maximum number of concurrent workers
	maxParallel int

	// executor is used to execute individual plan units
	executor Executor

	// eventPublisher publishes execution events
	eventPublisher EventPublisher

	// stateManager manages resource state
	stateManager StateManager

	// mu protects shared state during execution
	mu sync.RWMutex

	// unitResults maps plan unit IDs to their execution results
	unitResults map[string]*ExecutionResult

	// unitStatus tracks the current status of each unit
	unitStatus map[string]PlanStatus
}

// NewParallelScheduler creates a new parallel scheduler.
func NewParallelScheduler(
	maxParallel int,
	executor Executor,
	eventPublisher EventPublisher,
	stateManager StateManager,
) *ParallelScheduler {
	if maxParallel <= 0 {
		maxParallel = 10 // Default to 10 concurrent workers
	}

	return &ParallelScheduler{
		maxParallel:    maxParallel,
		executor:       executor,
		eventPublisher: eventPublisher,
		stateManager:   stateManager,
		unitResults:    make(map[string]*ExecutionResult),
		unitStatus:     make(map[string]PlanStatus),
	}
}

// Schedule schedules a plan for execution with the given options.
func (s *ParallelScheduler) Schedule(
	ctx context.Context,
	plan *Plan,
	opts ScheduleOptions,
) (string, error) {
	if plan == nil {
		return "", NewPermanentError("plan is nil", nil).WithCode(ErrCodeValidation)
	}

	// Ensure the plan has a valid execution graph
	if plan.Graph == nil {
		return "", NewPermanentError("plan has no execution graph", nil).
			WithCode(ErrCodeValidation)
	}

	// Create a new run
	run := &Run{
		ID:        uuid.New().String(),
		PlanID:    plan.ID,
		Status:    RunStatusPending,
		StartedAt: time.Now(),
		User:      opts.User,
		Summary: RunSummary{
			Total:   len(plan.Units),
			Pending: len(plan.Units),
		},
		Metadata: make(map[string]interface{}),
	}

	// Store run ID in plan metadata for tracking
	if plan.Metadata == nil {
		plan.Metadata = make(map[string]interface{})
	}
	plan.Metadata["run_id"] = run.ID

	// Apply delay if specified
	if opts.Delay > 0 {
		select {
		case <-time.After(opts.Delay):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

	// Save the initial run state
	if err := s.stateManager.SaveRun(ctx, run); err != nil {
		return "", fmt.Errorf("failed to save run: %w", err)
	}

	// Publish run started event
	s.publishEvent(ctx, run.ID, "", EventTypeRunStarted, "Run started", "info")

	// Start execution in a goroutine
	go func() {
		execCtx := context.Background()
		if err := s.executeRun(execCtx, run, plan, opts); err != nil {
			s.publishEvent(execCtx, run.ID, "", EventTypeRunFailed,
				fmt.Sprintf("Run failed: %v", err), "error")
		}
	}()

	return run.ID, nil
}

// executeRun executes the plan and updates the run status.
func (s *ParallelScheduler) executeRun(
	ctx context.Context,
	run *Run,
	plan *Plan,
	opts ScheduleOptions,
) error {
	// Update run status to running
	run.Status = RunStatusRunning
	if err := s.stateManager.SaveRun(ctx, run); err != nil {
		return fmt.Errorf("failed to update run status: %w", err)
	}

	// Initialize unit status map
	s.mu.Lock()
	for _, unit := range plan.Units {
		s.unitStatus[unit.ID] = PlanStatusPending
	}
	s.mu.Unlock()

	// Execute the plan level by level
	err := s.executePlanLevels(ctx, run, plan, opts)

	// Calculate final run statistics
	s.mu.RLock()
	summary := s.calculateRunSummary(plan.Units)
	s.mu.RUnlock()

	// Update final run status
	run.Summary = summary
	completedAt := time.Now()
	run.CompletedAt = &completedAt
	run.Duration = completedAt.Sub(run.StartedAt)

	// Determine final run status
	if err != nil {
		run.Status = RunStatusFailed
	} else if summary.Failed > 0 {
		if summary.Succeeded > 0 {
			run.Status = RunStatusPartial
		} else {
			run.Status = RunStatusFailed
		}
	} else if summary.Skipped > 0 {
		run.Status = RunStatusPartial
	} else {
		run.Status = RunStatusSucceeded
	}

	// Save final run state
	if saveErr := s.stateManager.SaveRun(ctx, run); saveErr != nil {
		return fmt.Errorf("failed to save final run state: %w", saveErr)
	}

	// Publish completion event
	if run.Status == RunStatusSucceeded {
		s.publishEvent(ctx, run.ID, "", EventTypeRunCompleted, "Run completed successfully", "info")
	} else {
		s.publishEvent(ctx, run.ID, "", EventTypeRunFailed,
			fmt.Sprintf("Run completed with status: %s", run.Status), "error")
	}

	return err
}

// executePlanLevels executes the plan level by level, with parallelism within each level.
func (s *ParallelScheduler) executePlanLevels(
	ctx context.Context,
	run *Run,
	plan *Plan,
	opts ScheduleOptions,
) error {
	// Build unit map for quick lookups
	unitMap := make(map[string]*PlanUnit)
	for i := range plan.Units {
		unit := &plan.Units[i]
		unitMap[unit.ID] = unit
	}

	// Process each level in sequence
	for level := 0; level < plan.Graph.Depth; level++ {
		// Get units at this level
		levelUnits := s.getUnitsAtLevel(plan.Graph, level, unitMap)
		if len(levelUnits) == 0 {
			continue
		}

		// Execute units at this level in parallel
		if err := s.executeLevelParallel(ctx, run, levelUnits, opts); err != nil {
			if opts.FailFast {
				return fmt.Errorf("level %d failed: %w", level, err)
			}
			// Continue to next level even if this level had failures
		}

		// Check for cancellation
		select {
		case <-ctx.Done():
			return s.handleCancellation(ctx, run, plan)
		default:
		}
	}

	return nil
}

// getUnitsAtLevel returns all plan units at the specified execution level.
func (s *ParallelScheduler) getUnitsAtLevel(
	graph *ExecutionGraph,
	level int,
	unitMap map[string]*PlanUnit,
) []*PlanUnit {
	units := make([]*PlanUnit, 0)

	for _, node := range graph.Nodes {
		if node.Level == level {
			if unit, exists := unitMap[node.ID]; exists {
				units = append(units, unit)
			}
		}
	}

	return units
}

// executeLevelParallel executes all units at a level in parallel using a worker pool.
func (s *ParallelScheduler) executeLevelParallel(
	ctx context.Context,
	run *Run,
	units []*PlanUnit,
	opts ScheduleOptions,
) error {
	// Determine worker count (min of maxParallel and number of units)
	workerCount := s.maxParallel
	if opts.MaxParallel > 0 && opts.MaxParallel < workerCount {
		workerCount = opts.MaxParallel
	}
	if len(units) < workerCount {
		workerCount = len(units)
	}

	// Create work queue
	workQueue := make(chan *PlanUnit, len(units))
	for _, unit := range units {
		workQueue <- unit
	}
	close(workQueue)

	// Create worker pool
	var wg sync.WaitGroup
	errChan := make(chan error, len(units))

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for unit := range workQueue {
				// Check if dependencies succeeded
				if !s.checkDependencies(unit) {
					s.markUnitSkipped(unit, "Dependencies failed")
					continue
				}

				// Execute the unit
				if err := s.executeUnit(ctx, run, unit, opts); err != nil {
					errChan <- fmt.Errorf("unit %s failed: %w", unit.ID, err)
				}

				// Check for cancellation
				select {
				case <-ctx.Done():
					return
				default:
				}
			}
		}()
	}

	// Wait for all workers to complete
	wg.Wait()
	close(errChan)

	// Collect errors
	var firstErr error
	for err := range errChan {
		if firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

// executeUnit executes a single plan unit with retry logic.
func (s *ParallelScheduler) executeUnit(
	ctx context.Context,
	run *Run,
	unit *PlanUnit,
	opts ScheduleOptions,
) error {
	// Mark unit as running
	s.updateUnitStatus(unit.ID, PlanStatusRunning)

	// Publish unit started event
	s.publishEvent(ctx, run.ID, unit.ID, EventTypePlanUnitStarted,
		fmt.Sprintf("Started execution of %s", unit.ResourceID), "info")

	startTime := time.Now()

	// Execute with retry logic
	var result *ExecutionResult
	var err error

	for attempt := 0; attempt <= unit.MaxRetries; attempt++ {
		// Create timeout context
		execCtx, cancel := context.WithTimeout(ctx, unit.Timeout)

		// Execute the unit
		if opts.DryRun {
			result = s.simulateDryRun(unit)
			err = nil
		} else {
			result, err = s.executor.ExecuteUnit(execCtx, unit)
		}

		cancel()

		// Check if execution succeeded
		if err == nil && result != nil && result.Status == PlanStatusSucceeded {
			break
		}

		// Check if error is retryable
		if err != nil && !IsRetryable(err) {
			break
		}

		// Don't retry on last attempt
		if attempt >= unit.MaxRetries {
			break
		}

		// Calculate backoff delay
		backoff := s.calculateBackoff(attempt, err)

		// Publish retry event
		s.publishEvent(ctx, run.ID, unit.ID, EventTypeWarning,
			fmt.Sprintf("Retrying after failure (attempt %d/%d)", attempt+1, unit.MaxRetries+1),
			"warning")

		// Wait with exponential backoff
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Store result
	if result == nil {
		result = &ExecutionResult{
			PlanUnitID:  unit.ID,
			Status:      PlanStatusFailed,
			StartedAt:   startTime,
			CompletedAt: time.Now(),
			Duration:    time.Since(startTime),
		}
	}

	if err != nil {
		result.Error = s.classifyError(err)
		result.Status = PlanStatusFailed
	}

	s.storeUnitResult(unit.ID, result)
	unit.Result = result

	// Update unit status
	if result.Status == PlanStatusSucceeded {
		s.updateUnitStatus(unit.ID, PlanStatusSucceeded)
		s.publishEvent(ctx, run.ID, unit.ID, EventTypePlanUnitCompleted,
			fmt.Sprintf("Completed execution of %s", unit.ResourceID), "info")
	} else {
		s.updateUnitStatus(unit.ID, PlanStatusFailed)
		s.publishEvent(ctx, run.ID, unit.ID, EventTypePlanUnitFailed,
			fmt.Sprintf("Failed execution of %s: %v", unit.ResourceID, err), "error")
		return err
	}

	return nil
}

// checkDependencies verifies that all required dependencies succeeded.
func (s *ParallelScheduler) checkDependencies(unit *PlanUnit) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, dep := range unit.Dependencies {
		status, exists := s.unitStatus[dep.TargetID]
		if !exists {
			return false
		}

		// Check dependency type
		switch dep.Type {
		case DependencyRequire:
			// Required dependencies must succeed
			if status != PlanStatusSucceeded {
				return false
			}
		case DependencyOrder:
			// Order dependencies must complete (success or failure)
			if !status.IsTerminal() {
				return false
			}
		case DependencyNotify:
			// Notify dependencies don't block execution
			continue
		}
	}

	return true
}

// calculateBackoff calculates exponential backoff with jitter.
func (s *ParallelScheduler) calculateBackoff(attempt int, err error) time.Duration {
	baseDelay := 1 * time.Second

	// Use different base delays for different error types
	if IsThrottled(err) {
		baseDelay = 5 * time.Second
	} else if IsConflict(err) {
		baseDelay = 2 * time.Second
	}

	// Exponential backoff: delay = baseDelay * 2^attempt
	delay := baseDelay * time.Duration(math.Pow(2, float64(attempt)))

	// Cap at 1 minute
	if delay > time.Minute {
		delay = time.Minute
	}

	// Add jitter (±25%)
	jitter := time.Duration(float64(delay) * 0.25)
	delay = delay + jitter/2

	return delay
}

// classifyError converts a regular error to an EngineError.
func (s *ParallelScheduler) classifyError(err error) *EngineError {
	if err == nil {
		return nil
	}

	// Check if already an EngineError
	var engineErr *EngineError
	if IsRetryable(err) {
		return engineErr
	}

	// Classify as permanent by default
	return NewPermanentError("execution failed", err).
		WithCode(ErrCodeProviderFailed)
}

// simulateDryRun simulates a dry-run execution.
func (s *ParallelScheduler) simulateDryRun(unit *PlanUnit) *ExecutionResult {
	return &ExecutionResult{
		PlanUnitID:  unit.ID,
		Status:      PlanStatusSucceeded,
		StartedAt:   time.Now(),
		CompletedAt: time.Now(),
		Duration:    0,
		NewState:    unit.DesiredState,
	}
}

// handleCancellation handles graceful cancellation of execution.
func (s *ParallelScheduler) handleCancellation(
	ctx context.Context,
	run *Run,
	plan *Plan,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Mark all pending and blocked units as cancelled
	for _, unit := range plan.Units {
		status := s.unitStatus[unit.ID]
		if status == PlanStatusPending || status == PlanStatusBlocked {
			s.unitStatus[unit.ID] = PlanStatusCancelled
		}
	}

	run.Status = RunStatusCancelled
	return NewPermanentError("execution cancelled", ctx.Err()).
		WithCode(ErrCodeInternal)
}

// updateUnitStatus updates the status of a plan unit.
func (s *ParallelScheduler) updateUnitStatus(unitID string, status PlanStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.unitStatus[unitID] = status
}

// storeUnitResult stores the execution result for a plan unit.
func (s *ParallelScheduler) storeUnitResult(unitID string, result *ExecutionResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.unitResults[unitID] = result
}

// markUnitSkipped marks a unit as skipped.
func (s *ParallelScheduler) markUnitSkipped(unit *PlanUnit, reason string) {
	s.updateUnitStatus(unit.ID, PlanStatusSkipped)

	result := &ExecutionResult{
		PlanUnitID:  unit.ID,
		Status:      PlanStatusSkipped,
		StartedAt:   time.Now(),
		CompletedAt: time.Now(),
		Duration:    0,
		Error: NewPermanentError(reason, nil).
			WithCode(ErrCodeDependencyFailed).
			WithResource(unit.ResourceID),
	}

	s.storeUnitResult(unit.ID, result)
	unit.Result = result
}

// calculateRunSummary calculates the final run summary statistics.
func (s *ParallelScheduler) calculateRunSummary(units []PlanUnit) RunSummary {
	summary := RunSummary{
		Total: len(units),
	}

	for _, unit := range units {
		status := s.unitStatus[unit.ID]
		switch status {
		case PlanStatusSucceeded:
			summary.Succeeded++
		case PlanStatusFailed:
			summary.Failed++
		case PlanStatusSkipped:
			summary.Skipped++
		case PlanStatusPending, PlanStatusBlocked:
			summary.Pending++
		case PlanStatusRunning:
			summary.Running++
		}
	}

	return summary
}

// publishEvent publishes an execution event.
func (s *ParallelScheduler) publishEvent(
	ctx context.Context,
	runID, planUnitID string,
	eventType EventType,
	message, level string,
) {
	if s.eventPublisher == nil {
		return
	}

	event := &Event{
		ID:         uuid.New().String(),
		Type:       eventType,
		Timestamp:  time.Now(),
		RunID:      runID,
		PlanUnitID: planUnitID,
		Message:    message,
		Level:      level,
	}

	// Publish event asynchronously to avoid blocking
	go func() {
		if err := s.eventPublisher.Publish(ctx, event); err != nil {
			// Log error but don't fail execution
		}
	}()
}

// Cancel cancels a running execution.
func (s *ParallelScheduler) Cancel(ctx context.Context, runID string) error {
	// Retrieve the run
	run, err := s.stateManager.GetRun(ctx, runID)
	if err != nil {
		return fmt.Errorf("failed to get run: %w", err)
	}

	// Check if run is active
	if !run.Status.IsActive() {
		return NewPermanentError("run is not active", nil).
			WithCode(ErrCodeValidation)
	}

	// Update run status to cancelled
	run.Status = RunStatusCancelled
	completedAt := time.Now()
	run.CompletedAt = &completedAt
	run.Duration = completedAt.Sub(run.StartedAt)

	// Save updated run
	if err := s.stateManager.SaveRun(ctx, run); err != nil {
		return fmt.Errorf("failed to save cancelled run: %w", err)
	}

	return nil
}

// GetStatus retrieves the status of a scheduled run.
func (s *ParallelScheduler) GetStatus(ctx context.Context, runID string) (*Run, error) {
	run, err := s.stateManager.GetRun(ctx, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to get run: %w", err)
	}

	return run, nil
}
