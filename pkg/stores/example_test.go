package stores_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/openfroyo/openfroyo/pkg/stores"
)

// ExampleNewSQLiteStore demonstrates creating and initializing a new SQLite store.
func ExampleNewSQLiteStore() {
	// Create store configuration
	store, err := stores.NewSQLiteStore(stores.Config{
		Path:            ":memory:", // Use in-memory database for example
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Initialize the database connection
	ctx := context.Background()
	if err := store.Init(ctx); err != nil {
		log.Fatal(err)
	}

	// Run migrations
	if err := store.Migrate(ctx); err != nil {
		log.Fatal(err)
	}

	defer store.Close()

	// Store is now ready to use
	fmt.Println("Store initialized successfully")
	// Output: Store initialized successfully
}

// ExampleSQLiteStore_CreateRun demonstrates creating a new run record.
func ExampleSQLiteStore_CreateRun() {
	store, _ := stores.NewSQLiteStore(stores.Config{Path: ":memory:"})
	ctx := context.Background()
	_ = store.Init(ctx)
	_ = store.Migrate(ctx)
	defer store.Close()

	// Create a new run
	run := &stores.Run{
		ID:        "run-001",
		PlanPath:  "/plans/deploy-apache.json",
		Status:    stores.RunStatusPending,
		StartedAt: time.Now(),
		Metadata:  `{"environment":"production"}`,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := store.CreateRun(ctx, run); err != nil {
		log.Fatal(err)
	}

	// Retrieve the run
	retrieved, err := store.GetRun(ctx, "run-001")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Run ID: %s, Status: %s\n", retrieved.ID, retrieved.Status)
	// Output: Run ID: run-001, Status: pending
}

// ExampleSQLiteStore_UpsertResourceState demonstrates managing resource state.
func ExampleSQLiteStore_UpsertResourceState() {
	store, _ := stores.NewSQLiteStore(stores.Config{Path: ":memory:"})
	ctx := context.Background()
	_ = store.Init(ctx)
	_ = store.Migrate(ctx)
	defer store.Close()

	// Create a run (required for foreign key)
	run := &stores.Run{
		ID:        "run-002",
		PlanPath:  "/plans/test.json",
		Status:    stores.RunStatusCompleted,
		StartedAt: time.Now(),
		Metadata:  `{}`,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = store.CreateRun(ctx, run)

	// Upsert resource state (insert)
	state := &stores.ResourceState{
		ID:           "rs-001",
		ResourceType: "linux.pkg",
		ResourceName: "nginx",
		State:        `{"installed":true,"version":"1.18.0"}`,
		Hash:         "abc123def456",
		LastRunID:    "run-002",
		LastApplied:  time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := store.UpsertResourceState(ctx, state); err != nil {
		log.Fatal(err)
	}

	// Get the state
	retrieved, err := store.GetResourceState(ctx, "linux.pkg", "nginx")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Resource: %s/%s, Hash: %s\n",
		retrieved.ResourceType, retrieved.ResourceName, retrieved.Hash)
	// Output: Resource: linux.pkg/nginx, Hash: abc123def456
}

// ExampleSQLiteStore_AppendEvent demonstrates logging events.
func ExampleSQLiteStore_AppendEvent() {
	store, _ := stores.NewSQLiteStore(stores.Config{Path: ":memory:"})
	ctx := context.Background()
	_ = store.Init(ctx)
	_ = store.Migrate(ctx)
	defer store.Close()

	// Create a run
	run := &stores.Run{
		ID:        "run-003",
		PlanPath:  "/plans/test.json",
		Status:    stores.RunStatusRunning,
		StartedAt: time.Now(),
		Metadata:  `{}`,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = store.CreateRun(ctx, run)

	// Log an event
	details := `{"target":"production"}`
	event := &stores.Event{
		RunID:     &run.ID,
		Level:     stores.EventLevelInfo,
		Message:   "Starting deployment",
		Details:   &details,
		Timestamp: time.Now(),
	}

	if err := store.AppendEvent(ctx, event); err != nil {
		log.Fatal(err)
	}

	// Retrieve events
	events, err := store.GetEvents(ctx, &run.ID, nil, nil, 10, 0)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Event count: %d, Message: %s\n", len(events), events[0].Message)
	// Output: Event count: 1, Message: Starting deployment
}

// ExampleSQLiteStore_UpsertFact demonstrates storing facts with TTL.
func ExampleSQLiteStore_UpsertFact() {
	store, _ := stores.NewSQLiteStore(stores.Config{Path: ":memory:"})
	ctx := context.Background()
	_ = store.Init(ctx)
	_ = store.Migrate(ctx)
	defer store.Close()

	// Store a fact without expiry
	fact := &stores.Fact{
		ID:        "fact-001",
		TargetID:  "host-001",
		Namespace: "os.basic",
		Key:       "os_family",
		Value:     `"debian"`,
		TTL:       0, // Never expires
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := store.UpsertFact(ctx, fact); err != nil {
		log.Fatal(err)
	}

	// Retrieve the fact
	retrieved, err := store.GetFact(ctx, "host-001", "os.basic", "os_family")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Fact: %s/%s/%s = %s\n",
		retrieved.TargetID, retrieved.Namespace, retrieved.Key, retrieved.Value)
	// Output: Fact: host-001/os.basic/os_family = "debian"
}

// ExampleSQLiteStore_BeginTx demonstrates using transactions.
func ExampleSQLiteStore_BeginTx() {
	store, _ := stores.NewSQLiteStore(stores.Config{Path: ":memory:"})
	ctx := context.Background()
	_ = store.Init(ctx)
	_ = store.Migrate(ctx)
	defer store.Close()

	// Begin transaction
	tx, err := store.BeginTx(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Perform operations within transaction
	query := `
		INSERT INTO runs (id, plan_path, status, started_at, metadata, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	_, err = tx.ExecContext(ctx, query, "run-tx-001", "/plans/test.json",
		"pending", now, "{}", now, now)

	if err != nil {
		_ = store.RollbackTx(tx)
		log.Fatal(err)
	}

	// Commit transaction
	if err := store.CommitTx(tx); err != nil {
		log.Fatal(err)
	}

	// Verify run was created
	run, err := store.GetRun(ctx, "run-tx-001")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Transaction committed: Run %s created\n", run.ID)
	// Output: Transaction committed: Run run-tx-001 created
}
