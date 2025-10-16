-- Enable foreign keys
PRAGMA foreign_keys = ON;

-- runs table: execution runs and status
CREATE TABLE IF NOT EXISTS runs (
    id TEXT PRIMARY KEY NOT NULL,
    plan_path TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('pending', 'running', 'completed', 'failed', 'cancelled')),
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    error TEXT,
    metadata TEXT NOT NULL DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_runs_status ON runs(status);
CREATE INDEX idx_runs_started_at ON runs(started_at DESC);
CREATE INDEX idx_runs_created_at ON runs(created_at DESC);

-- plan_units table: individual plan unit records
CREATE TABLE IF NOT EXISTS plan_units (
    id TEXT PRIMARY KEY NOT NULL,
    run_id TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_name TEXT NOT NULL,
    action TEXT NOT NULL CHECK(action IN ('create', 'update', 'delete', 'read')),
    status TEXT NOT NULL CHECK(status IN ('pending', 'running', 'completed', 'failed', 'skipped')),
    dependencies TEXT NOT NULL DEFAULT '[]',
    desired_state TEXT NOT NULL,
    actual_state TEXT,
    diff TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    error TEXT,
    retries INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (run_id) REFERENCES runs(id) ON DELETE CASCADE
);

CREATE INDEX idx_plan_units_run_id ON plan_units(run_id);
CREATE INDEX idx_plan_units_status ON plan_units(status);
CREATE INDEX idx_plan_units_resource ON plan_units(resource_type, resource_name);

-- events table: append-only event log
CREATE TABLE IF NOT EXISTS events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id TEXT,
    plan_unit_id TEXT,
    level TEXT NOT NULL CHECK(level IN ('debug', 'info', 'warning', 'error')),
    message TEXT NOT NULL,
    details TEXT,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (run_id) REFERENCES runs(id) ON DELETE CASCADE,
    FOREIGN KEY (plan_unit_id) REFERENCES plan_units(id) ON DELETE CASCADE
);

CREATE INDEX idx_events_run_id ON events(run_id);
CREATE INDEX idx_events_plan_unit_id ON events(plan_unit_id);
CREATE INDEX idx_events_level ON events(level);
CREATE INDEX idx_events_timestamp ON events(timestamp DESC);

-- resource_state table: current state of managed resources
CREATE TABLE IF NOT EXISTS resource_state (
    id TEXT PRIMARY KEY NOT NULL,
    resource_type TEXT NOT NULL,
    resource_name TEXT NOT NULL,
    state TEXT NOT NULL,
    hash TEXT NOT NULL,
    last_run_id TEXT NOT NULL,
    last_applied TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (last_run_id) REFERENCES runs(id) ON DELETE RESTRICT,
    UNIQUE(resource_type, resource_name)
);

CREATE INDEX idx_resource_state_type_name ON resource_state(resource_type, resource_name);
CREATE INDEX idx_resource_state_hash ON resource_state(hash);
CREATE INDEX idx_resource_state_last_applied ON resource_state(last_applied DESC);

-- facts table: discovered facts with TTL
CREATE TABLE IF NOT EXISTS facts (
    id TEXT PRIMARY KEY NOT NULL,
    target_id TEXT NOT NULL,
    namespace TEXT NOT NULL,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    ttl INTEGER NOT NULL DEFAULT 0,
    expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(target_id, namespace, key)
);

CREATE INDEX idx_facts_target_id ON facts(target_id);
CREATE INDEX idx_facts_namespace ON facts(namespace);
CREATE INDEX idx_facts_expires_at ON facts(expires_at);
CREATE INDEX idx_facts_target_namespace ON facts(target_id, namespace);

-- audit table: audit trail
CREATE TABLE IF NOT EXISTS audit (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    action TEXT NOT NULL,
    actor TEXT NOT NULL,
    target_id TEXT,
    details TEXT,
    ip_address TEXT,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_action ON audit(action);
CREATE INDEX idx_audit_actor ON audit(actor);
CREATE INDEX idx_audit_target_id ON audit(target_id);
CREATE INDEX idx_audit_timestamp ON audit(timestamp DESC);

-- Triggers for updated_at timestamps
CREATE TRIGGER update_runs_timestamp
    AFTER UPDATE ON runs
    FOR EACH ROW
BEGIN
    UPDATE runs SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER update_plan_units_timestamp
    AFTER UPDATE ON plan_units
    FOR EACH ROW
BEGIN
    UPDATE plan_units SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER update_resource_state_timestamp
    AFTER UPDATE ON resource_state
    FOR EACH ROW
BEGIN
    UPDATE resource_state SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER update_facts_timestamp
    AFTER UPDATE ON facts
    FOR EACH ROW
BEGIN
    UPDATE facts SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
