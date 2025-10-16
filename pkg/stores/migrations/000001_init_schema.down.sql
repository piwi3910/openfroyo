-- Drop triggers
DROP TRIGGER IF EXISTS update_facts_timestamp;
DROP TRIGGER IF EXISTS update_resource_state_timestamp;
DROP TRIGGER IF EXISTS update_plan_units_timestamp;
DROP TRIGGER IF EXISTS update_runs_timestamp;

-- Drop tables in reverse order (respecting foreign keys)
DROP TABLE IF EXISTS audit;
DROP TABLE IF EXISTS facts;
DROP TABLE IF EXISTS resource_state;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS plan_units;
DROP TABLE IF EXISTS runs;
