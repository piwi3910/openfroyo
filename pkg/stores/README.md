# OpenFroyo SQLite Persistence Layer

This package provides a production-ready SQLite-based persistence layer for the OpenFroyo infrastructure orchestration engine.

## Features

- **WAL Mode**: Write-Ahead Logging enabled for better concurrency
- **Connection Pooling**: Configured limits for optimal performance
- **Migration System**: Versioned schema migrations using golang-migrate
- **Foreign Key Support**: Enforced referential integrity with cascade deletes
- **Transaction Support**: Full ACID compliance with rollback capability
- **TTL Support**: Automatic expiration for facts with cleanup operations
- **Comprehensive Testing**: 77.5% test coverage with race detection

## Database Schema

### Tables

- **runs** - Execution runs and their status
- **plan_units** - Individual units within an execution plan
- **events** - Append-only event log for auditing
- **resource_state** - Current state of managed resources
- **facts** - Discovered system facts with TTL support
- **audit** - Audit trail for all operations

### Indexes

All tables include optimized indexes on:
- Primary keys
- Foreign keys
- Status fields
- Timestamp fields
- Composite queries (e.g., `resource_type + resource_name`)

## Usage

### Initialization

```go
import "github.com/openfroyo/openfroyo/pkg/stores"

// Create store
store, err := stores.NewSQLiteStore(stores.Config{
    Path: "./data/openfroyo.db",
    MaxOpenConns: 25,
    MaxIdleConns: 5,
    ConnMaxLifetime: 5 * time.Minute,
})
if err != nil {
    log.Fatal(err)
}

// Initialize connection
ctx := context.Background()
if err := store.Init(ctx); err != nil {
    log.Fatal(err)
}

// Run migrations
if err := store.Migrate(ctx); err != nil {
    log.Fatal(err)
}

defer store.Close()
```

### Creating a Run

```go
run := &stores.Run{
    ID:        "run-001",
    PlanPath:  "/plans/deploy-apache.json",
    Status:    stores.RunStatusPending,
    StartedAt: time.Now(),
    Metadata:  `{"environment":"production"}`,
    CreatedAt: time.Now(),
    UpdatedAt: time.Now(),
}

err := store.CreateRun(ctx, run)
```

### Creating Plan Units

```go
unit := &stores.PlanUnit{
    ID:           "pu-001",
    RunID:        "run-001",
    ResourceType: "linux.pkg",
    ResourceName: "apache2",
    Action:       "create",
    Status:       stores.PlanUnitStatusPending,
    Dependencies: `[]`,
    DesiredState: `{"state":"present","version":"2.4"}`,
    CreatedAt:    time.Now(),
    UpdatedAt:    time.Now(),
}

err := store.CreatePlanUnit(ctx, unit)
```

### Logging Events

```go
event := &stores.Event{
    RunID:     &run.ID,
    Level:     stores.EventLevelInfo,
    Message:   "Starting package installation",
    Details:   `{"package":"apache2"}`,
    Timestamp: time.Now(),
}

err := store.AppendEvent(ctx, event)
```

### Managing Resource State

```go
state := &stores.ResourceState{
    ID:           "rs-001",
    ResourceType: "linux.pkg",
    ResourceName: "apache2",
    State:        `{"installed":true,"version":"2.4.41"}`,
    Hash:         "abc123...",  // SHA256 of state
    LastRunID:    "run-001",
    LastApplied:  time.Now(),
    CreatedAt:    time.Now(),
    UpdatedAt:    time.Now(),
}

// Upsert creates or updates
err := store.UpsertResourceState(ctx, state)

// Get current state
retrieved, err := store.GetResourceState(ctx, "linux.pkg", "apache2")
```

### Facts with TTL

```go
// Fact without expiry
fact := &stores.Fact{
    ID:        "fact-001",
    TargetID:  "host-001",
    Namespace: "os.basic",
    Key:       "os_family",
    Value:     `"debian"`,
    TTL:       0,  // Never expires
    CreatedAt: time.Now(),
    UpdatedAt: time.Now(),
}

err := store.UpsertFact(ctx, fact)

// Fact with TTL
factWithTTL := &stores.Fact{
    ID:        "fact-002",
    TargetID:  "host-001",
    Namespace: "hw.cpu",
    Key:       "cpu_count",
    Value:     `4`,
    TTL:       3600,  // 1 hour
    ExpiresAt: &expiresAt,
    CreatedAt: time.Now(),
    UpdatedAt: time.Now(),
}

err := store.UpsertFact(ctx, factWithTTL)

// Clean up expired facts
deleted, err := store.DeleteExpiredFacts(ctx)
```

### Audit Logging

```go
entry := &stores.AuditEntry{
    Action:    "run.created",
    Actor:     "system",
    TargetID:  &run.ID,
    Details:   `{"plan":"/plans/deploy.json"}`,
    IPAddress: &ipAddr,
    Timestamp: time.Now(),
}

err := store.CreateAuditEntry(ctx, entry)
```

### Transactions

```go
tx, err := store.BeginTx(ctx)
if err != nil {
    return err
}

// Perform operations...
_, err = tx.ExecContext(ctx, "INSERT INTO runs ...")
if err != nil {
    store.RollbackTx(tx)
    return err
}

err = store.CommitTx(tx)
```

## Migrations

Migrations are embedded in the binary using Go's `embed` package. They are automatically applied when calling `Migrate()`.

To create a new migration:

1. Create `pkg/stores/migrations/NNNNNN_description.up.sql`
2. Create `pkg/stores/migrations/NNNNNN_description.down.sql`

The migration system tracks applied migrations and only runs new ones.

## Testing

```bash
# Run all tests
go test ./pkg/stores/...

# With coverage
go test ./pkg/stores/... -coverprofile=coverage.out
go tool cover -html=coverage.out

# With race detection
go test ./pkg/stores/... -race
```

## Performance Considerations

1. **WAL Mode**: Enables concurrent reads during writes
2. **Connection Pool**: Limits set to 25 max / 5 idle connections
3. **Prepared Statements**: All queries use prepared statements to prevent SQL injection
4. **Indexes**: Strategic indexes on frequently queried columns
5. **Batch Operations**: Use transactions for multiple related operations

## Error Handling

All operations return descriptive errors wrapped with context:

```go
run, err := store.GetRun(ctx, "invalid-id")
if err != nil {
    // Error includes context: "failed to get run: ..."
    log.Printf("Error: %v", err)
}
```

## Security

- **SQL Injection Protection**: All queries use parameterized statements
- **Foreign Key Enforcement**: Prevents orphaned records
- **Context Support**: Operations can be cancelled via context
- **Transaction Isolation**: Serializable isolation level for consistency

## Limitations (Solo Profile)

The current implementation is optimized for the "solo" deployment profile:

- Single SQLite database file
- Local filesystem storage
- No multi-node support
- Advisory locking via SQLite file locks

For cluster deployments, migrate to the PostgreSQL + S3 + NATS backend (future roadmap).

## File Locations

Default data directory: `./data/`
- Database: `./data/openfroyo.db`
- WAL file: `./data/openfroyo.db-wal`
- SHM file: `./data/openfroyo.db-shm`

## Backup & Restore

```bash
# Backup (while running, thanks to WAL mode)
sqlite3 ./data/openfroyo.db ".backup ./backups/openfroyo-$(date +%Y%m%d).db"

# Restore
cp ./backups/openfroyo-20250101.db ./data/openfroyo.db
```

## Dependencies

- `modernc.org/sqlite` - Pure Go SQLite implementation
- `github.com/golang-migrate/migrate/v4` - Database migrations
- Standard library `database/sql`

## Roadmap

Future enhancements planned:

- Metrics instrumentation (Prometheus)
- Query performance logging
- Automatic vacuum scheduling
- Hot backup API
- Read replicas support (cluster profile)
