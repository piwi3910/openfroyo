package stores

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	// SQLite driver
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// SQLiteStore implements the Store interface using SQLite
type SQLiteStore struct {
	db   *sql.DB
	path string
}

// Config holds SQLite store configuration
type Config struct {
	Path            string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// NewSQLiteStore creates a new SQLite store instance
func NewSQLiteStore(cfg Config) (*SQLiteStore, error) {
	if cfg.Path == "" {
		return nil, fmt.Errorf("database path is required")
	}

	// Set defaults
	if cfg.MaxOpenConns == 0 {
		cfg.MaxOpenConns = 25
	}
	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = 5
	}
	if cfg.ConnMaxLifetime == 0 {
		cfg.ConnMaxLifetime = 5 * time.Minute
	}

	return &SQLiteStore{
		path: cfg.Path,
	}, nil
}

// Init initializes the database connection and enables WAL mode.
func (s *SQLiteStore) Init(ctx context.Context) error {
	// Open database with SQLite-specific connection parameters
	dsn := fmt.Sprintf("%s?_foreign_keys=on&_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL&_txlock=immediate", s.path)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection and set PRAGMAs
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Ensure foreign keys are enabled (connection-level setting)
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		_ = db.Close()
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	s.db = db
	return nil
}

// Close closes the database connection
func (s *SQLiteStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Migrate runs database migrations.
func (s *SQLiteStore) Migrate(_ context.Context) error {
	if s.db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Create migration source from embedded FS
	sourceDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	// Create database driver
	driver, err := sqlite3.WithInstance(s.db, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("failed to create database driver: %w", err)
	}

	// Create migration instance
	m, err := migrate.NewWithInstance("iofs", sourceDriver, "sqlite3", driver)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	// Run migrations
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// BeginTx starts a new transaction
func (s *SQLiteStore) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return s.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
}

// CommitTx commits a transaction
func (s *SQLiteStore) CommitTx(tx *sql.Tx) error {
	return tx.Commit()
}

// RollbackTx rolls back a transaction
func (s *SQLiteStore) RollbackTx(tx *sql.Tx) error {
	return tx.Rollback()
}

// CreateRun creates a new run record
func (s *SQLiteStore) CreateRun(ctx context.Context, run *Run) error {
	query := `
		INSERT INTO runs (id, plan_path, status, started_at, completed_at, error, metadata, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
		run.ID,
		run.PlanPath,
		run.Status,
		run.StartedAt,
		run.CompletedAt,
		run.Error,
		run.Metadata,
		run.CreatedAt,
		run.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create run: %w", err)
	}

	return nil
}

// GetRun retrieves a run by ID
func (s *SQLiteStore) GetRun(ctx context.Context, id string) (*Run, error) {
	query := `
		SELECT id, plan_path, status, started_at, completed_at, error, metadata, created_at, updated_at
		FROM runs
		WHERE id = ?
	`

	run := &Run{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&run.ID,
		&run.PlanPath,
		&run.Status,
		&run.StartedAt,
		&run.CompletedAt,
		&run.Error,
		&run.Metadata,
		&run.CreatedAt,
		&run.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("run not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get run: %w", err)
	}

	return run, nil
}

// UpdateRunStatus updates the status of a run
func (s *SQLiteStore) UpdateRunStatus(ctx context.Context, id string, status RunStatus, errMsg *string) error {
	query := `
		UPDATE runs
		SET status = ?, error = ?, completed_at = ?
		WHERE id = ?
	`

	var completedAt *time.Time
	if status == RunStatusCompleted || status == RunStatusFailed || status == RunStatusCancelled {
		now := time.Now()
		completedAt = &now
	}

	result, err := s.db.ExecContext(ctx, query, status, errMsg, completedAt, id)
	if err != nil {
		return fmt.Errorf("failed to update run status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("run not found: %s", id)
	}

	return nil
}

// ListRuns lists runs with pagination
func (s *SQLiteStore) ListRuns(ctx context.Context, limit, offset int) ([]*Run, error) {
	query := `
		SELECT id, plan_path, status, started_at, completed_at, error, metadata, created_at, updated_at
		FROM runs
		ORDER BY started_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list runs: %w", err)
	}
	defer rows.Close()

	runs := []*Run{}
	for rows.Next() {
		run := &Run{}
		err := rows.Scan(
			&run.ID,
			&run.PlanPath,
			&run.Status,
			&run.StartedAt,
			&run.CompletedAt,
			&run.Error,
			&run.Metadata,
			&run.CreatedAt,
			&run.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan run: %w", err)
		}
		runs = append(runs, run)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating runs: %w", err)
	}

	return runs, nil
}

// DeleteRun deletes a run by ID
func (s *SQLiteStore) DeleteRun(ctx context.Context, id string) error {
	query := `DELETE FROM runs WHERE id = ?`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete run: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("run not found: %s", id)
	}

	return nil
}

// CreatePlanUnit creates a new plan unit record
func (s *SQLiteStore) CreatePlanUnit(ctx context.Context, unit *PlanUnit) error {
	query := `
		INSERT INTO plan_units (
			id, run_id, resource_type, resource_name, action, status,
			dependencies, desired_state, actual_state, diff,
			started_at, completed_at, error, retries, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
		unit.ID,
		unit.RunID,
		unit.ResourceType,
		unit.ResourceName,
		unit.Action,
		unit.Status,
		unit.Dependencies,
		unit.DesiredState,
		unit.ActualState,
		unit.Diff,
		unit.StartedAt,
		unit.CompletedAt,
		unit.Error,
		unit.Retries,
		unit.CreatedAt,
		unit.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create plan unit: %w", err)
	}

	return nil
}

// GetPlanUnit retrieves a plan unit by ID
func (s *SQLiteStore) GetPlanUnit(ctx context.Context, id string) (*PlanUnit, error) {
	query := `
		SELECT id, run_id, resource_type, resource_name, action, status,
			   dependencies, desired_state, actual_state, diff,
			   started_at, completed_at, error, retries, created_at, updated_at
		FROM plan_units
		WHERE id = ?
	`

	unit := &PlanUnit{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&unit.ID,
		&unit.RunID,
		&unit.ResourceType,
		&unit.ResourceName,
		&unit.Action,
		&unit.Status,
		&unit.Dependencies,
		&unit.DesiredState,
		&unit.ActualState,
		&unit.Diff,
		&unit.StartedAt,
		&unit.CompletedAt,
		&unit.Error,
		&unit.Retries,
		&unit.CreatedAt,
		&unit.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("plan unit not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get plan unit: %w", err)
	}

	return unit, nil
}

// UpdatePlanUnitStatus updates the status of a plan unit
func (s *SQLiteStore) UpdatePlanUnitStatus(ctx context.Context, id string, status PlanUnitStatus, actualState *string, errMsg *string) error {
	query := `
		UPDATE plan_units
		SET status = ?, actual_state = ?, error = ?,
			started_at = CASE WHEN started_at IS NULL AND ? = 'running' THEN CURRENT_TIMESTAMP ELSE started_at END,
			completed_at = CASE WHEN ? IN ('completed', 'failed', 'skipped') THEN CURRENT_TIMESTAMP ELSE completed_at END
		WHERE id = ?
	`

	result, err := s.db.ExecContext(ctx, query, status, actualState, errMsg, status, status, id)
	if err != nil {
		return fmt.Errorf("failed to update plan unit status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("plan unit not found: %s", id)
	}

	return nil
}

// ListPlanUnitsByRun lists all plan units for a run
func (s *SQLiteStore) ListPlanUnitsByRun(ctx context.Context, runID string) ([]*PlanUnit, error) {
	query := `
		SELECT id, run_id, resource_type, resource_name, action, status,
			   dependencies, desired_state, actual_state, diff,
			   started_at, completed_at, error, retries, created_at, updated_at
		FROM plan_units
		WHERE run_id = ?
		ORDER BY created_at ASC
	`

	rows, err := s.db.QueryContext(ctx, query, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to list plan units: %w", err)
	}
	defer rows.Close()

	units := []*PlanUnit{}
	for rows.Next() {
		unit := &PlanUnit{}
		err := rows.Scan(
			&unit.ID,
			&unit.RunID,
			&unit.ResourceType,
			&unit.ResourceName,
			&unit.Action,
			&unit.Status,
			&unit.Dependencies,
			&unit.DesiredState,
			&unit.ActualState,
			&unit.Diff,
			&unit.StartedAt,
			&unit.CompletedAt,
			&unit.Error,
			&unit.Retries,
			&unit.CreatedAt,
			&unit.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan plan unit: %w", err)
		}
		units = append(units, unit)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating plan units: %w", err)
	}

	return units, nil
}

// DeletePlanUnit deletes a plan unit by ID
func (s *SQLiteStore) DeletePlanUnit(ctx context.Context, id string) error {
	query := `DELETE FROM plan_units WHERE id = ?`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete plan unit: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("plan unit not found: %s", id)
	}

	return nil
}

// IncrementPlanUnitRetries increments the retry counter for a plan unit
func (s *SQLiteStore) IncrementPlanUnitRetries(ctx context.Context, id string) error {
	query := `UPDATE plan_units SET retries = retries + 1 WHERE id = ?`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment retries: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("plan unit not found: %s", id)
	}

	return nil
}

// AppendEvent appends a new event to the log
func (s *SQLiteStore) AppendEvent(ctx context.Context, event *Event) error {
	query := `
		INSERT INTO events (run_id, plan_unit_id, level, message, details, timestamp)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := s.db.ExecContext(ctx, query,
		event.RunID,
		event.PlanUnitID,
		event.Level,
		event.Message,
		event.Details,
		event.Timestamp,
	)

	if err != nil {
		return fmt.Errorf("failed to append event: %w", err)
	}

	// Get the auto-generated ID
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get event ID: %w", err)
	}

	event.ID = id
	return nil
}

// GetEvents retrieves events with optional filters and pagination
func (s *SQLiteStore) GetEvents(ctx context.Context, runID *string, planUnitID *string, level *EventLevel, limit, offset int) ([]*Event, error) {
	query := `
		SELECT id, run_id, plan_unit_id, level, message, details, timestamp
		FROM events
		WHERE (? IS NULL OR run_id = ?)
		  AND (? IS NULL OR plan_unit_id = ?)
		  AND (? IS NULL OR level = ?)
		ORDER BY timestamp DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.QueryContext(ctx, query, runID, runID, planUnitID, planUnitID, level, level, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}
	defer rows.Close()

	events := []*Event{}
	for rows.Next() {
		event := &Event{}
		err := rows.Scan(
			&event.ID,
			&event.RunID,
			&event.PlanUnitID,
			&event.Level,
			&event.Message,
			&event.Details,
			&event.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	return events, nil
}

// UpsertResourceState inserts or updates resource state
func (s *SQLiteStore) UpsertResourceState(ctx context.Context, state *ResourceState) error {
	query := `
		INSERT INTO resource_state (
			id, resource_type, resource_name, state, hash, last_run_id, last_applied, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(resource_type, resource_name) DO UPDATE SET
			state = excluded.state,
			hash = excluded.hash,
			last_run_id = excluded.last_run_id,
			last_applied = excluded.last_applied
	`

	_, err := s.db.ExecContext(ctx, query,
		state.ID,
		state.ResourceType,
		state.ResourceName,
		state.State,
		state.Hash,
		state.LastRunID,
		state.LastApplied,
		state.CreatedAt,
		state.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to upsert resource state: %w", err)
	}

	return nil
}

// GetResourceState retrieves resource state by type and name
func (s *SQLiteStore) GetResourceState(ctx context.Context, resourceType, resourceName string) (*ResourceState, error) {
	query := `
		SELECT id, resource_type, resource_name, state, hash, last_run_id, last_applied, created_at, updated_at
		FROM resource_state
		WHERE resource_type = ? AND resource_name = ?
	`

	state := &ResourceState{}
	err := s.db.QueryRowContext(ctx, query, resourceType, resourceName).Scan(
		&state.ID,
		&state.ResourceType,
		&state.ResourceName,
		&state.State,
		&state.Hash,
		&state.LastRunID,
		&state.LastApplied,
		&state.CreatedAt,
		&state.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("resource state not found: %s/%s", resourceType, resourceName)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get resource state: %w", err)
	}

	return state, nil
}

// ListResourceStates lists all resource states with pagination
func (s *SQLiteStore) ListResourceStates(ctx context.Context, limit, offset int) ([]*ResourceState, error) {
	query := `
		SELECT id, resource_type, resource_name, state, hash, last_run_id, last_applied, created_at, updated_at
		FROM resource_state
		ORDER BY last_applied DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list resource states: %w", err)
	}
	defer rows.Close()

	states := []*ResourceState{}
	for rows.Next() {
		state := &ResourceState{}
		err := rows.Scan(
			&state.ID,
			&state.ResourceType,
			&state.ResourceName,
			&state.State,
			&state.Hash,
			&state.LastRunID,
			&state.LastApplied,
			&state.CreatedAt,
			&state.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan resource state: %w", err)
		}
		states = append(states, state)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating resource states: %w", err)
	}

	return states, nil
}

// DeleteResourceState deletes a resource state by ID
func (s *SQLiteStore) DeleteResourceState(ctx context.Context, id string) error {
	query := `DELETE FROM resource_state WHERE id = ?`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete resource state: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("resource state not found: %s", id)
	}

	return nil
}

// UpsertFact inserts or updates a fact
func (s *SQLiteStore) UpsertFact(ctx context.Context, fact *Fact) error {
	query := `
		INSERT INTO facts (
			id, target_id, namespace, key, value, ttl, expires_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(target_id, namespace, key) DO UPDATE SET
			value = excluded.value,
			ttl = excluded.ttl,
			expires_at = excluded.expires_at
	`

	// Format expires_at to SQLite-compatible datetime string
	var expiresAtStr *string
	if fact.ExpiresAt != nil {
		formatted := fact.ExpiresAt.UTC().Format("2006-01-02 15:04:05")
		expiresAtStr = &formatted
	}

	_, err := s.db.ExecContext(ctx, query,
		fact.ID,
		fact.TargetID,
		fact.Namespace,
		fact.Key,
		fact.Value,
		fact.TTL,
		expiresAtStr,
		fact.CreatedAt.UTC().Format("2006-01-02 15:04:05"),
		fact.UpdatedAt.UTC().Format("2006-01-02 15:04:05"),
	)

	if err != nil {
		return fmt.Errorf("failed to upsert fact: %w", err)
	}

	return nil
}

// GetFact retrieves a fact by target, namespace, and key
func (s *SQLiteStore) GetFact(ctx context.Context, targetID, namespace, key string) (*Fact, error) {
	query := `
		SELECT id, target_id, namespace, key, value, ttl, expires_at, created_at, updated_at
		FROM facts
		WHERE target_id = ? AND namespace = ? AND key = ?
		  AND (expires_at IS NULL OR datetime(expires_at) > datetime('now'))
	`

	fact := &Fact{}
	err := s.db.QueryRowContext(ctx, query, targetID, namespace, key).Scan(
		&fact.ID,
		&fact.TargetID,
		&fact.Namespace,
		&fact.Key,
		&fact.Value,
		&fact.TTL,
		&fact.ExpiresAt,
		&fact.CreatedAt,
		&fact.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("fact not found or expired: %s/%s/%s", targetID, namespace, key)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get fact: %w", err)
	}

	return fact, nil
}

// ListFacts lists facts with optional filters and pagination
func (s *SQLiteStore) ListFacts(ctx context.Context, targetID *string, namespace *string, limit, offset int) ([]*Fact, error) {
	query := `
		SELECT id, target_id, namespace, key, value, ttl, expires_at, created_at, updated_at
		FROM facts
		WHERE (? IS NULL OR target_id = ?)
		  AND (? IS NULL OR namespace = ?)
		  AND (expires_at IS NULL OR datetime(expires_at) > datetime('now'))
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.QueryContext(ctx, query, targetID, targetID, namespace, namespace, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list facts: %w", err)
	}
	defer rows.Close()

	facts := []*Fact{}
	for rows.Next() {
		fact := &Fact{}
		err := rows.Scan(
			&fact.ID,
			&fact.TargetID,
			&fact.Namespace,
			&fact.Key,
			&fact.Value,
			&fact.TTL,
			&fact.ExpiresAt,
			&fact.CreatedAt,
			&fact.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan fact: %w", err)
		}
		facts = append(facts, fact)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating facts: %w", err)
	}

	return facts, nil
}

// DeleteExpiredFacts deletes all expired facts
func (s *SQLiteStore) DeleteExpiredFacts(ctx context.Context) (int64, error) {
	query := `DELETE FROM facts WHERE expires_at IS NOT NULL AND datetime(expires_at) <= datetime('now')`

	result, err := s.db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired facts: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rows, nil
}

// DeleteFact deletes a fact by ID
func (s *SQLiteStore) DeleteFact(ctx context.Context, id string) error {
	query := `DELETE FROM facts WHERE id = ?`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete fact: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("fact not found: %s", id)
	}

	return nil
}

// CreateAuditEntry creates a new audit log entry
func (s *SQLiteStore) CreateAuditEntry(ctx context.Context, entry *AuditEntry) error {
	query := `
		INSERT INTO audit (action, actor, target_id, details, ip_address, timestamp)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := s.db.ExecContext(ctx, query,
		entry.Action,
		entry.Actor,
		entry.TargetID,
		entry.Details,
		entry.IPAddress,
		entry.Timestamp,
	)

	if err != nil {
		return fmt.Errorf("failed to create audit entry: %w", err)
	}

	// Get the auto-generated ID
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get audit entry ID: %w", err)
	}

	entry.ID = id
	return nil
}

// ListAuditEntries lists audit entries with optional filters and pagination
func (s *SQLiteStore) ListAuditEntries(ctx context.Context, action *string, actor *string, limit, offset int) ([]*AuditEntry, error) {
	query := `
		SELECT id, action, actor, target_id, details, ip_address, timestamp
		FROM audit
		WHERE (? IS NULL OR action = ?)
		  AND (? IS NULL OR actor = ?)
		ORDER BY timestamp DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.QueryContext(ctx, query, action, action, actor, actor, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit entries: %w", err)
	}
	defer rows.Close()

	entries := []*AuditEntry{}
	for rows.Next() {
		entry := &AuditEntry{}
		err := rows.Scan(
			&entry.ID,
			&entry.Action,
			&entry.Actor,
			&entry.TargetID,
			&entry.Details,
			&entry.IPAddress,
			&entry.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit entry: %w", err)
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating audit entries: %w", err)
	}

	return entries, nil
}

// HealthCheck verifies the database connection is healthy
func (s *SQLiteStore) HealthCheck(ctx context.Context) error {
	if s.db == nil {
		return fmt.Errorf("database not initialized")
	}

	return s.db.PingContext(ctx)
}
