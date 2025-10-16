package stores

import (
	"context"
	"os"
	"testing"
	"time"
)

// setupTestStore creates an in-memory SQLite store for testing
func setupTestStore(t *testing.T) *SQLiteStore {
	t.Helper()

	store, err := NewSQLiteStore(Config{
		Path: ":memory:",
	})
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	ctx := context.Background()
	if err := store.Init(ctx); err != nil {
		t.Fatalf("failed to initialize store: %v", err)
	}

	if err := store.Migrate(ctx); err != nil {
		t.Fatalf("failed to migrate store: %v", err)
	}

	return store
}

// TestStoreLifecycle tests database initialization and closure
func TestStoreLifecycle(t *testing.T) {
	store, err := NewSQLiteStore(Config{
		Path: ":memory:",
	})
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	ctx := context.Background()
	if err := store.Init(ctx); err != nil {
		t.Fatalf("failed to initialize store: %v", err)
	}

	if err := store.HealthCheck(ctx); err != nil {
		t.Fatalf("health check failed: %v", err)
	}

	if err := store.Close(); err != nil {
		t.Fatalf("failed to close store: %v", err)
	}
}

// TestStoreMigrations tests database migrations
func TestStoreMigrations(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	ctx := context.Background()

	// Check that tables exist by querying them
	tables := []string{"runs", "plan_units", "events", "resource_state", "facts", "audit"}
	for _, table := range tables {
		query := "SELECT COUNT(*) FROM " + table
		var count int
		err := store.db.QueryRowContext(ctx, query).Scan(&count)
		if err != nil {
			t.Errorf("table %s does not exist or is not accessible: %v", table, err)
		}
	}
}

// TestRunCRUD tests Run CRUD operations
func TestRunCRUD(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	ctx := context.Background()
	now := time.Now()

	// Create
	run := &Run{
		ID:        "run-001",
		PlanPath:  "/plans/test.json",
		Status:    RunStatusPending,
		StartedAt: now,
		Metadata:  `{"env":"test"}`,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := store.CreateRun(ctx, run); err != nil {
		t.Fatalf("failed to create run: %v", err)
	}

	// Read
	retrieved, err := store.GetRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("failed to get run: %v", err)
	}

	if retrieved.ID != run.ID {
		t.Errorf("expected ID %s, got %s", run.ID, retrieved.ID)
	}
	if retrieved.PlanPath != run.PlanPath {
		t.Errorf("expected PlanPath %s, got %s", run.PlanPath, retrieved.PlanPath)
	}
	if retrieved.Status != run.Status {
		t.Errorf("expected Status %s, got %s", run.Status, retrieved.Status)
	}

	// Update
	errMsg := "test error"
	if err := store.UpdateRunStatus(ctx, run.ID, RunStatusFailed, &errMsg); err != nil {
		t.Fatalf("failed to update run status: %v", err)
	}

	updated, err := store.GetRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("failed to get updated run: %v", err)
	}

	if updated.Status != RunStatusFailed {
		t.Errorf("expected Status %s, got %s", RunStatusFailed, updated.Status)
	}
	if updated.Error == nil || *updated.Error != errMsg {
		t.Errorf("expected Error %s, got %v", errMsg, updated.Error)
	}
	if updated.CompletedAt == nil {
		t.Error("expected CompletedAt to be set")
	}

	// List
	runs, err := store.ListRuns(ctx, 10, 0)
	if err != nil {
		t.Fatalf("failed to list runs: %v", err)
	}

	if len(runs) != 1 {
		t.Errorf("expected 1 run, got %d", len(runs))
	}

	// Delete
	if err := store.DeleteRun(ctx, run.ID); err != nil {
		t.Fatalf("failed to delete run: %v", err)
	}

	_, err = store.GetRun(ctx, run.ID)
	if err == nil {
		t.Error("expected error when getting deleted run")
	}
}

// TestPlanUnitCRUD tests PlanUnit CRUD operations
func TestPlanUnitCRUD(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	ctx := context.Background()
	now := time.Now()

	// Create a run first (required for foreign key)
	run := &Run{
		ID:        "run-002",
		PlanPath:  "/plans/test.json",
		Status:    RunStatusRunning,
		StartedAt: now,
		Metadata:  `{}`,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := store.CreateRun(ctx, run); err != nil {
		t.Fatalf("failed to create run: %v", err)
	}

	// Create
	unit := &PlanUnit{
		ID:           "pu-001",
		RunID:        run.ID,
		ResourceType: "linux.pkg",
		ResourceName: "apache2",
		Action:       "create",
		Status:       PlanUnitStatusPending,
		Dependencies: `[]`,
		DesiredState: `{"state":"present"}`,
		Retries:      0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := store.CreatePlanUnit(ctx, unit); err != nil {
		t.Fatalf("failed to create plan unit: %v", err)
	}

	// Read
	retrieved, err := store.GetPlanUnit(ctx, unit.ID)
	if err != nil {
		t.Fatalf("failed to get plan unit: %v", err)
	}

	if retrieved.ID != unit.ID {
		t.Errorf("expected ID %s, got %s", unit.ID, retrieved.ID)
	}
	if retrieved.ResourceType != unit.ResourceType {
		t.Errorf("expected ResourceType %s, got %s", unit.ResourceType, retrieved.ResourceType)
	}

	// Update Status
	actualState := `{"state":"present","version":"2.4.41"}`
	if err := store.UpdatePlanUnitStatus(ctx, unit.ID, PlanUnitStatusCompleted, &actualState, nil); err != nil {
		t.Fatalf("failed to update plan unit status: %v", err)
	}

	updated, err := store.GetPlanUnit(ctx, unit.ID)
	if err != nil {
		t.Fatalf("failed to get updated plan unit: %v", err)
	}

	if updated.Status != PlanUnitStatusCompleted {
		t.Errorf("expected Status %s, got %s", PlanUnitStatusCompleted, updated.Status)
	}
	if updated.ActualState == nil || *updated.ActualState != actualState {
		t.Errorf("expected ActualState %s, got %v", actualState, updated.ActualState)
	}

	// Increment Retries
	if err := store.IncrementPlanUnitRetries(ctx, unit.ID); err != nil {
		t.Fatalf("failed to increment retries: %v", err)
	}

	retried, err := store.GetPlanUnit(ctx, unit.ID)
	if err != nil {
		t.Fatalf("failed to get plan unit after retry increment: %v", err)
	}

	if retried.Retries != 1 {
		t.Errorf("expected Retries 1, got %d", retried.Retries)
	}

	// List by Run
	units, err := store.ListPlanUnitsByRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("failed to list plan units: %v", err)
	}

	if len(units) != 1 {
		t.Errorf("expected 1 plan unit, got %d", len(units))
	}

	// Delete
	if err := store.DeletePlanUnit(ctx, unit.ID); err != nil {
		t.Fatalf("failed to delete plan unit: %v", err)
	}

	_, err = store.GetPlanUnit(ctx, unit.ID)
	if err == nil {
		t.Error("expected error when getting deleted plan unit")
	}
}

// TestEventOperations tests Event operations
func TestEventOperations(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	ctx := context.Background()
	now := time.Now()

	// Create a run first
	run := &Run{
		ID:        "run-003",
		PlanPath:  "/plans/test.json",
		Status:    RunStatusRunning,
		StartedAt: now,
		Metadata:  `{}`,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := store.CreateRun(ctx, run); err != nil {
		t.Fatalf("failed to create run: %v", err)
	}

	// Append events
	events := []*Event{
		{
			RunID:     &run.ID,
			Level:     EventLevelInfo,
			Message:   "Starting execution",
			Timestamp: now,
		},
		{
			RunID:     &run.ID,
			Level:     EventLevelWarning,
			Message:   "Resource already exists",
			Timestamp: now.Add(1 * time.Second),
		},
		{
			RunID:     &run.ID,
			Level:     EventLevelError,
			Message:   "Failed to apply",
			Timestamp: now.Add(2 * time.Second),
		},
	}

	for _, event := range events {
		if err := store.AppendEvent(ctx, event); err != nil {
			t.Fatalf("failed to append event: %v", err)
		}
		if event.ID == 0 {
			t.Error("expected event ID to be set after insert")
		}
	}

	// Get all events for run
	retrieved, err := store.GetEvents(ctx, &run.ID, nil, nil, 10, 0)
	if err != nil {
		t.Fatalf("failed to get events: %v", err)
	}

	if len(retrieved) != 3 {
		t.Errorf("expected 3 events, got %d", len(retrieved))
	}

	// Filter by level
	errorLevel := EventLevelError
	filtered, err := store.GetEvents(ctx, nil, nil, &errorLevel, 10, 0)
	if err != nil {
		t.Fatalf("failed to get filtered events: %v", err)
	}

	if len(filtered) != 1 {
		t.Errorf("expected 1 error event, got %d", len(filtered))
	}
	if filtered[0].Level != EventLevelError {
		t.Errorf("expected level %s, got %s", EventLevelError, filtered[0].Level)
	}
}

// TestResourceStateOperations tests ResourceState operations
func TestResourceStateOperations(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	ctx := context.Background()
	now := time.Now()

	// Create a run first
	run := &Run{
		ID:        "run-004",
		PlanPath:  "/plans/test.json",
		Status:    RunStatusCompleted,
		StartedAt: now,
		Metadata:  `{}`,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := store.CreateRun(ctx, run); err != nil {
		t.Fatalf("failed to create run: %v", err)
	}

	// Upsert (insert)
	state := &ResourceState{
		ID:           "rs-001",
		ResourceType: "linux.pkg",
		ResourceName: "nginx",
		State:        `{"state":"present","version":"1.18.0"}`,
		Hash:         "abc123def456",
		LastRunID:    run.ID,
		LastApplied:  now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := store.UpsertResourceState(ctx, state); err != nil {
		t.Fatalf("failed to upsert resource state: %v", err)
	}

	// Get
	retrieved, err := store.GetResourceState(ctx, state.ResourceType, state.ResourceName)
	if err != nil {
		t.Fatalf("failed to get resource state: %v", err)
	}

	if retrieved.Hash != state.Hash {
		t.Errorf("expected Hash %s, got %s", state.Hash, retrieved.Hash)
	}

	// Upsert (update)
	state.State = `{"state":"present","version":"1.20.0"}`
	state.Hash = "xyz789ghi012"

	if err := store.UpsertResourceState(ctx, state); err != nil {
		t.Fatalf("failed to upsert resource state (update): %v", err)
	}

	updated, err := store.GetResourceState(ctx, state.ResourceType, state.ResourceName)
	if err != nil {
		t.Fatalf("failed to get updated resource state: %v", err)
	}

	if updated.Hash != "xyz789ghi012" {
		t.Errorf("expected updated Hash xyz789ghi012, got %s", updated.Hash)
	}

	// List
	states, err := store.ListResourceStates(ctx, 10, 0)
	if err != nil {
		t.Fatalf("failed to list resource states: %v", err)
	}

	if len(states) != 1 {
		t.Errorf("expected 1 resource state, got %d", len(states))
	}

	// Delete
	if err := store.DeleteResourceState(ctx, state.ID); err != nil {
		t.Fatalf("failed to delete resource state: %v", err)
	}

	_, err = store.GetResourceState(ctx, state.ResourceType, state.ResourceName)
	if err == nil {
		t.Error("expected error when getting deleted resource state")
	}
}

// TestFactOperations tests Fact operations including TTL
func TestFactOperations(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	ctx := context.Background()
	now := time.Now().UTC()

	// Upsert fact without expiry
	fact1 := &Fact{
		ID:        "fact-001",
		TargetID:  "host-001",
		Namespace: "os.basic",
		Key:       "os_family",
		Value:     `"linux"`,
		TTL:       0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := store.UpsertFact(ctx, fact1); err != nil {
		t.Fatalf("failed to upsert fact: %v", err)
	}

	// Upsert fact with TTL (future expiry)
	expiresAt := now.Add(1 * time.Hour)
	fact2 := &Fact{
		ID:        "fact-002",
		TargetID:  "host-001",
		Namespace: "hw.cpu",
		Key:       "cpu_count",
		Value:     `4`,
		TTL:       3600,
		ExpiresAt: &expiresAt,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := store.UpsertFact(ctx, fact2); err != nil {
		t.Fatalf("failed to upsert fact with TTL: %v", err)
	}

	// Upsert expired fact (past expiry)
	expiredAt := now.Add(-1 * time.Hour)
	fact3 := &Fact{
		ID:        "fact-003",
		TargetID:  "host-001",
		Namespace: "net.ifaces",
		Key:       "interface_count",
		Value:     `2`,
		TTL:       3600,
		ExpiresAt: &expiredAt,
		CreatedAt: now.Add(-2 * time.Hour),
		UpdatedAt: now.Add(-2 * time.Hour),
	}

	if err := store.UpsertFact(ctx, fact3); err != nil {
		t.Fatalf("failed to upsert expired fact: %v", err)
	}

	// Get non-expired fact
	retrieved, err := store.GetFact(ctx, fact1.TargetID, fact1.Namespace, fact1.Key)
	if err != nil {
		t.Fatalf("failed to get fact: %v", err)
	}

	if retrieved.Value != fact1.Value {
		t.Errorf("expected Value %s, got %s", fact1.Value, retrieved.Value)
	}

	// Try to get expired fact (should fail because GetFact filters expired facts)
	_, err = store.GetFact(ctx, fact3.TargetID, fact3.Namespace, fact3.Key)
	if err == nil {
		t.Error("expected error when getting expired fact")
	}

	// List facts (should not include expired ones)
	targetID := "host-001"
	facts, err := store.ListFacts(ctx, &targetID, nil, 10, 0)
	if err != nil {
		t.Fatalf("failed to list facts: %v", err)
	}

	// Should get fact1 (no expiry) and fact2 (future expiry), but not fact3 (expired)
	if len(facts) != 2 {
		t.Errorf("expected 2 non-expired facts, got %d", len(facts))
		for i, f := range facts {
			t.Logf("fact[%d]: id=%s, expires_at=%v", i, f.ID, f.ExpiresAt)
		}
	}

	// Delete expired facts
	deleted, err := store.DeleteExpiredFacts(ctx)
	if err != nil {
		t.Fatalf("failed to delete expired facts: %v", err)
	}

	if deleted != 1 {
		t.Errorf("expected 1 expired fact deleted, got %d", deleted)
	}

	// Verify fact3 is really gone
	var count int
	err = store.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM facts WHERE id = ?", fact3.ID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to count fact3: %v", err)
	}
	if count != 0 {
		t.Errorf("expected fact3 to be deleted, but it still exists")
	}

	// Delete fact by ID
	if err := store.DeleteFact(ctx, fact1.ID); err != nil {
		t.Fatalf("failed to delete fact: %v", err)
	}

	_, err = store.GetFact(ctx, fact1.TargetID, fact1.Namespace, fact1.Key)
	if err == nil {
		t.Error("expected error when getting deleted fact")
	}
}

// TestAuditOperations tests Audit operations
func TestAuditOperations(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	ctx := context.Background()
	now := time.Now()

	// Create audit entries
	entries := []*AuditEntry{
		{
			Action:    "run.created",
			Actor:     "admin",
			Timestamp: now,
		},
		{
			Action:    "state.updated",
			Actor:     "system",
			Timestamp: now.Add(1 * time.Second),
		},
		{
			Action:    "run.created",
			Actor:     "user1",
			Timestamp: now.Add(2 * time.Second),
		},
	}

	for _, entry := range entries {
		if err := store.CreateAuditEntry(ctx, entry); err != nil {
			t.Fatalf("failed to create audit entry: %v", err)
		}
		if entry.ID == 0 {
			t.Error("expected audit entry ID to be set after insert")
		}
	}

	// List all
	retrieved, err := store.ListAuditEntries(ctx, nil, nil, 10, 0)
	if err != nil {
		t.Fatalf("failed to list audit entries: %v", err)
	}

	if len(retrieved) != 3 {
		t.Errorf("expected 3 audit entries, got %d", len(retrieved))
	}

	// Filter by action
	action := "run.created"
	filtered, err := store.ListAuditEntries(ctx, &action, nil, 10, 0)
	if err != nil {
		t.Fatalf("failed to list filtered audit entries: %v", err)
	}

	if len(filtered) != 2 {
		t.Errorf("expected 2 run.created entries, got %d", len(filtered))
	}

	// Filter by actor
	actor := "admin"
	actorFiltered, err := store.ListAuditEntries(ctx, nil, &actor, 10, 0)
	if err != nil {
		t.Fatalf("failed to list actor filtered audit entries: %v", err)
	}

	if len(actorFiltered) != 1 {
		t.Errorf("expected 1 admin entry, got %d", len(actorFiltered))
	}
}

// TestTransactions tests transaction support
func TestTransactions(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	ctx := context.Background()
	now := time.Now()

	// Begin transaction
	tx, err := store.BeginTx(ctx)
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	// Create run within transaction
	run := &Run{
		ID:        "run-tx-001",
		PlanPath:  "/plans/test.json",
		Status:    RunStatusPending,
		StartedAt: now,
		Metadata:  `{}`,
		CreatedAt: now,
		UpdatedAt: now,
	}

	query := `
		INSERT INTO runs (id, plan_path, status, started_at, metadata, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err = tx.ExecContext(ctx, query, run.ID, run.PlanPath, run.Status, run.StartedAt, run.Metadata, run.CreatedAt, run.UpdatedAt)
	if err != nil {
		store.RollbackTx(tx)
		t.Fatalf("failed to insert run in transaction: %v", err)
	}

	// Rollback
	if err := store.RollbackTx(tx); err != nil {
		t.Fatalf("failed to rollback transaction: %v", err)
	}

	// Verify run was not created
	_, err = store.GetRun(ctx, run.ID)
	if err == nil {
		t.Error("expected error when getting rolled back run")
	}

	// Begin new transaction and commit
	tx, err = store.BeginTx(ctx)
	if err != nil {
		t.Fatalf("failed to begin second transaction: %v", err)
	}

	_, err = tx.ExecContext(ctx, query, run.ID, run.PlanPath, run.Status, run.StartedAt, run.Metadata, run.CreatedAt, run.UpdatedAt)
	if err != nil {
		store.RollbackTx(tx)
		t.Fatalf("failed to insert run in second transaction: %v", err)
	}

	if err := store.CommitTx(tx); err != nil {
		t.Fatalf("failed to commit transaction: %v", err)
	}

	// Verify run was created
	retrieved, err := store.GetRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("failed to get committed run: %v", err)
	}

	if retrieved.ID != run.ID {
		t.Errorf("expected ID %s, got %s", run.ID, retrieved.ID)
	}
}

// TestCascadeDelete tests foreign key cascading
func TestCascadeDelete(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	ctx := context.Background()
	now := time.Now()

	// Create run
	run := &Run{
		ID:        "run-cascade-001",
		PlanPath:  "/plans/test.json",
		Status:    RunStatusRunning,
		StartedAt: now,
		Metadata:  `{}`,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := store.CreateRun(ctx, run); err != nil {
		t.Fatalf("failed to create run: %v", err)
	}

	// Create plan unit
	unit := &PlanUnit{
		ID:           "pu-cascade-001",
		RunID:        run.ID,
		ResourceType: "linux.pkg",
		ResourceName: "test",
		Action:       "create",
		Status:       PlanUnitStatusPending,
		Dependencies: `[]`,
		DesiredState: `{}`,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := store.CreatePlanUnit(ctx, unit); err != nil {
		t.Fatalf("failed to create plan unit: %v", err)
	}

	// Create event
	event := &Event{
		RunID:     &run.ID,
		Level:     EventLevelInfo,
		Message:   "test event",
		Timestamp: now,
	}
	if err := store.AppendEvent(ctx, event); err != nil {
		t.Fatalf("failed to append event: %v", err)
	}

	// Delete run (should cascade to plan_units and events)
	if err := store.DeleteRun(ctx, run.ID); err != nil {
		t.Fatalf("failed to delete run: %v", err)
	}

	// Verify plan unit was deleted
	_, err := store.GetPlanUnit(ctx, unit.ID)
	if err == nil {
		t.Error("expected error when getting cascaded deleted plan unit")
	}

	// Verify events were deleted
	events, err := store.GetEvents(ctx, &run.ID, nil, nil, 10, 0)
	if err != nil {
		t.Fatalf("failed to get events: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events after cascade delete, got %d", len(events))
	}
}

// TestMain sets up and tears down test environment
func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()

	// Exit
	os.Exit(code)
}
