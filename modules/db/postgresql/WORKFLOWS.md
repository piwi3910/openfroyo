# PostgreSQL Module - Common Workflows

Real-world usage scenarios and workflows for the PostgreSQL module.

## Table of Contents

- [Application Setup](#application-setup)
- [Database Migration](#database-migration)
- [Backup and Recovery](#backup-and-recovery)
- [User and Access Management](#user-and-access-management)
- [Performance Optimization](#performance-optimization)
- [Extension Management](#extension-management)
- [Multi-Environment Setup](#multi-environment-setup)
- [Disaster Recovery](#disaster-recovery)

---

## Application Setup

### New Application Database Setup

Complete setup for a new application database with user, schema, and extensions.

```yaml
name: Setup New Application Database
inventory: @group:db-servers

vars:
  app_name: myapp
  db_user: myapp_user
  db_password: "{{ vault_myapp_db_password }}"

run:
  # 1. Create database
  - name: Create application database
    module: db/postgresql
    vars:
      command: create_database
      database_name: "{{ app_name }}_production"
      encoding: UTF8

  # 2. Create application user
  - name: Create application user
    module: db/postgresql
    vars:
      command: create_user
      user_name: "{{ db_user }}"
      user_password: "{{ db_password }}"
      createdb: false
      createrole: false
      superuser: false

  # 3. Grant privileges
  - name: Grant database privileges
    module: db/postgresql
    vars:
      command: grant_privileges
      database_name: "{{ app_name }}_production"
      user_name: "{{ db_user }}"
      privileges: ALL

  # 4. Create application schema
  - name: Create app schema
    module: db/postgresql
    vars:
      command: create_schema
      postgresql_database: "{{ app_name }}_production"
      schema_name: app
      owner: "{{ db_user }}"

  # 5. Set search path
  - name: Set schema search path
    module: db/postgresql
    vars:
      command: set_search_path
      postgresql_database: "{{ app_name }}_production"
      search_path: app,public

  # 6. Enable required extensions
  - name: Enable UUID extension
    module: db/postgresql
    vars:
      command: create_extension
      postgresql_database: "{{ app_name }}_production"
      extension_name: uuid-ossp

  - name: Enable trigram extension
    module: db/postgresql
    vars:
      command: create_extension
      postgresql_database: "{{ app_name }}_production"
      extension_name: pg_trgm

  # 7. Run initial schema
  - name: Execute schema creation script
    module: db/postgresql
    vars:
      command: execute_script
      postgresql_database: "{{ app_name }}_production"
      script_file: /tmp/schema.sql
```

---

## Database Migration

### Schema Migration Workflow

Safe database migration with backup and rollback capability.

```yaml
name: Database Schema Migration
inventory: @group:db-servers

vars:
  db_name: myapp_production
  migration_file: /tmp/migration_v2.sql
  backup_dir: /backups
  timestamp: "{{ now | date('YmdHis') }}"

run:
  # 1. Backup before migration
  - name: Backup database before migration
    module: db/postgresql
    vars:
      command: backup_database
      database_name: "{{ db_name }}"
      backup_file: "{{ backup_dir }}/{{ db_name }}_pre_migration_{{ timestamp }}.dump"
      format: custom

  # 2. Verify backup
  - name: Verify backup file exists
    shell: ls -lh {{ backup_dir }}/{{ db_name }}_pre_migration_{{ timestamp }}.dump

  # 3. Show current activity (ensure low load)
  - name: Check database activity
    module: db/postgresql
    vars:
      command: show_activity
      postgresql_database: "{{ db_name }}"

  # 4. Execute migration
  - name: Run migration script
    module: db/postgresql
    vars:
      command: execute_script
      postgresql_database: "{{ db_name }}"
      script_file: "{{ migration_file }}"
    register: migration_result

  # 5. Analyze database after migration
  - name: Analyze database
    module: db/postgresql
    vars:
      command: analyze_table
      postgresql_database: "{{ db_name }}"

  # 6. Verify migration (optional query check)
  - name: Verify migration
    module: db/postgresql
    vars:
      command: execute_query
      postgresql_database: "{{ db_name }}"
      query: "SELECT COUNT(*) FROM pg_tables WHERE schemaname = 'app'"
```

### Rollback Migration

```yaml
name: Rollback Database Migration
inventory: @group:db-servers

vars:
  db_name: myapp_production
  backup_file: /backups/myapp_production_pre_migration_20240115120000.dump

run:
  # 1. Drop existing database
  - name: Drop current database
    module: db/postgresql
    vars:
      command: drop_database
      database_name: "{{ db_name }}"

  # 2. Recreate database
  - name: Recreate database
    module: db/postgresql
    vars:
      command: create_database
      database_name: "{{ db_name }}"
      owner: myapp_user

  # 3. Restore from backup
  - name: Restore from pre-migration backup
    module: db/postgresql
    vars:
      command: restore_database
      database_name: "{{ db_name }}"
      backup_file: "{{ backup_file }}"

  # 4. Verify restoration
  - name: Check database size
    module: db/postgresql
    vars:
      command: show_database_size
      database_name: "{{ db_name }}"
```

---

## Backup and Recovery

### Automated Daily Backup

```yaml
name: Daily Database Backup
inventory: @group:db-servers

vars:
  backup_root: /backups/postgresql
  retention_days: 30
  timestamp: "{{ now | date('Ymd') }}"

run:
  # Backup all production databases
  - name: Backup myapp production
    module: db/postgresql
    vars:
      command: backup_database
      database_name: myapp_production
      backup_file: "{{ backup_root }}/myapp_production_{{ timestamp }}.dump"
      format: custom

  - name: Backup analytics database
    module: db/postgresql
    vars:
      command: backup_database
      database_name: analytics_db
      backup_file: "{{ backup_root }}/analytics_db_{{ timestamp }}.dump"
      format: custom

  # Cleanup old backups
  - name: Remove backups older than retention period
    shell: find {{ backup_root }} -name "*.dump" -mtime +{{ retention_days }} -delete

  # Verify today's backups
  - name: List today's backups
    shell: ls -lh {{ backup_root }}/*{{ timestamp }}.dump
```

### Point-in-Time Recovery

```yaml
name: Point-in-Time Database Recovery
inventory: @group:db-servers

vars:
  source_db: myapp_production
  target_db: myapp_production_recovery
  backup_file: /backups/myapp_production_20240115.dump

run:
  # 1. Create recovery database
  - name: Create recovery database
    module: db/postgresql
    vars:
      command: create_database
      database_name: "{{ target_db }}"
      owner: myapp_user

  # 2. Restore backup
  - name: Restore from backup
    module: db/postgresql
    vars:
      command: restore_database
      database_name: "{{ target_db }}"
      backup_file: "{{ backup_file }}"

  # 3. Verify data
  - name: Check restored data
    module: db/postgresql
    vars:
      command: execute_query
      postgresql_database: "{{ target_db }}"
      query: "SELECT COUNT(*) FROM app.users"

  # 4. Compare with production
  - name: Compare record counts
    module: db/postgresql
    vars:
      command: execute_query
      postgresql_database: "{{ source_db }}"
      query: "SELECT COUNT(*) FROM app.users"
```

---

## User and Access Management

### Onboard New Team Member

```yaml
name: Onboard Database User
inventory: @group:db-servers

vars:
  new_user: john_doe
  user_password: "{{ vault_john_password }}"
  databases:
    - myapp_dev
    - myapp_staging

run:
  # 1. Create user account
  - name: Create user account
    module: db/postgresql
    vars:
      command: create_user
      user_name: "{{ new_user }}"
      user_password: "{{ user_password }}"
      createdb: false

  # 2. Grant read-only access to development
  - name: Grant dev database access
    module: db/postgresql
    vars:
      command: grant_privileges
      database_name: myapp_dev
      user_name: "{{ new_user }}"
      privileges: SELECT

  # 3. Grant full access to staging
  - name: Grant staging database access
    module: db/postgresql
    vars:
      command: grant_privileges
      database_name: myapp_staging
      user_name: "{{ new_user }}"
      privileges: ALL

  # 4. Verify user creation
  - name: List all users
    module: db/postgresql
    vars:
      command: list_users
```

### Offboard Team Member

```yaml
name: Offboard Database User
inventory: @group:db-servers

vars:
  user_to_remove: john_doe

run:
  # 1. Revoke all privileges
  - name: Revoke dev database privileges
    module: db/postgresql
    vars:
      command: revoke_privileges
      database_name: myapp_dev
      user_name: "{{ user_to_remove }}"
      privileges: ALL

  - name: Revoke staging database privileges
    module: db/postgresql
    vars:
      command: revoke_privileges
      database_name: myapp_staging
      user_name: "{{ user_to_remove }}"
      privileges: ALL

  # 2. Terminate active connections
  - name: Check for active connections
    module: db/postgresql
    vars:
      command: show_activity
      postgresql_database: postgres

  # 3. Drop user
  - name: Drop user account
    module: db/postgresql
    vars:
      command: drop_user
      user_name: "{{ user_to_remove }}"
```

### Rotate Application Password

```yaml
name: Rotate Application Database Password
inventory: @group:db-servers

vars:
  app_user: myapp_user
  new_password: "{{ vault_new_password }}"

run:
  # 1. Update password
  - name: Update user password
    module: db/postgresql
    vars:
      command: alter_user
      user_name: "{{ app_user }}"
      user_password: "{{ new_password }}"

  # 2. Restart application (external)
  - name: Notify application restart required
    debug:
      msg: "Password updated. Restart application with new credentials."
```

---

## Performance Optimization

### Weekly Maintenance

```yaml
name: Weekly Database Maintenance
inventory: @group:db-servers

vars:
  databases:
    - myapp_production
    - analytics_db

run:
  # Perform maintenance on each database
  - name: Vacuum and analyze databases
    loop: "{{ databases }}"
    module: db/postgresql
    vars:
      command: vacuum_table
      postgresql_database: "{{ item }}"
      full: false
      analyze: true

  # Check database sizes
  - name: Check database sizes
    loop: "{{ databases }}"
    module: db/postgresql
    vars:
      command: show_database_size
      database_name: "{{ item }}"

  # Show active connections
  - name: Show database activity
    loop: "{{ databases }}"
    module: db/postgresql
    vars:
      command: show_activity
      postgresql_database: "{{ item }}"
```

### Performance Analysis

```yaml
name: Database Performance Analysis
inventory: @group:db-servers

vars:
  db_name: myapp_production

run:
  # 1. List all tables
  - name: Get table list
    module: db/postgresql
    vars:
      command: list_tables
      postgresql_database: "{{ db_name }}"
      schema_name: app
    register: tables

  # 2. Analyze each table
  - name: Analyze all tables
    module: db/postgresql
    vars:
      command: analyze_table
      postgresql_database: "{{ db_name }}"

  # 3. Check for bloat
  - name: Check table sizes
    module: db/postgresql
    vars:
      command: execute_query
      postgresql_database: "{{ db_name }}"
      query: |
        SELECT schemaname, tablename,
               pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
        FROM pg_tables
        WHERE schemaname = 'app'
        ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC
        LIMIT 10

  # 4. Show active queries
  - name: Show long-running queries
    module: db/postgresql
    vars:
      command: show_activity
      postgresql_database: "{{ db_name }}"
```

---

## Extension Management

### Setup PostGIS for Geographic Data

```yaml
name: Setup PostGIS Extension
inventory: @group:db-servers

vars:
  db_name: mapping_app

run:
  # 1. Create database
  - name: Create mapping database
    module: db/postgresql
    vars:
      command: create_database
      database_name: "{{ db_name }}"

  # 2. Enable PostGIS
  - name: Enable PostGIS extension
    module: db/postgresql
    vars:
      command: create_extension
      postgresql_database: "{{ db_name }}"
      extension_name: postgis

  # 3. Enable PostGIS Topology
  - name: Enable PostGIS Topology
    module: db/postgresql
    vars:
      command: create_extension
      postgresql_database: "{{ db_name }}"
      extension_name: postgis_topology

  # 4. Verify extensions
  - name: List installed extensions
    module: db/postgresql
    vars:
      command: list_extensions
      postgresql_database: "{{ db_name }}"
```

### Setup Full-Text Search

```yaml
name: Setup Full-Text Search Extensions
inventory: @group:db-servers

vars:
  db_name: search_app

run:
  # Enable trigram matching
  - name: Enable pg_trgm
    module: db/postgresql
    vars:
      command: create_extension
      postgresql_database: "{{ db_name }}"
      extension_name: pg_trgm

  # Enable unaccent for accent-insensitive search
  - name: Enable unaccent
    module: db/postgresql
    vars:
      command: create_extension
      postgresql_database: "{{ db_name }}"
      extension_name: unaccent

  # Verify
  - name: List extensions
    module: db/postgresql
    vars:
      command: list_extensions
      postgresql_database: "{{ db_name }}"
```

---

## Multi-Environment Setup

### Create Development, Staging, Production

```yaml
name: Multi-Environment Database Setup
inventory: @group:db-servers

vars:
  app_name: myapp
  environments:
    - name: dev
      owner: myapp_dev_user
    - name: staging
      owner: myapp_staging_user
    - name: production
      owner: myapp_prod_user

run:
  # Create databases for each environment
  - name: Create environment databases
    loop: "{{ environments }}"
    module: db/postgresql
    vars:
      command: create_database
      database_name: "{{ app_name }}_{{ item.name }}"
      owner: "{{ item.owner }}"
      encoding: UTF8

  # Create users for each environment
  - name: Create environment users
    loop: "{{ environments }}"
    module: db/postgresql
    vars:
      command: create_user
      user_name: "{{ item.owner }}"
      user_password: "{{ lookup('vault', item.owner + '_password') }}"
      createdb: false

  # Grant privileges
  - name: Grant environment privileges
    loop: "{{ environments }}"
    module: db/postgresql
    vars:
      command: grant_privileges
      database_name: "{{ app_name }}_{{ item.name }}"
      user_name: "{{ item.owner }}"
      privileges: ALL

  # Create schemas
  - name: Create app schemas
    loop: "{{ environments }}"
    module: db/postgresql
    vars:
      command: create_schema
      postgresql_database: "{{ app_name }}_{{ item.name }}"
      schema_name: app
      owner: "{{ item.owner }}"
```

### Promote Staging to Production

```yaml
name: Promote Staging to Production
inventory: @group:db-servers

vars:
  app_name: myapp
  backup_dir: /backups

run:
  # 1. Backup current production
  - name: Backup production database
    module: db/postgresql
    vars:
      command: backup_database
      database_name: "{{ app_name }}_production"
      backup_file: "{{ backup_dir }}/{{ app_name }}_prod_pre_promotion_{{ timestamp }}.dump"
      format: custom

  # 2. Backup staging
  - name: Backup staging database
    module: db/postgresql
    vars:
      command: backup_database
      database_name: "{{ app_name }}_staging"
      backup_file: "{{ backup_dir }}/{{ app_name }}_staging_{{ timestamp }}.dump"
      format: custom

  # 3. Drop and recreate production
  - name: Drop production database
    module: db/postgresql
    vars:
      command: drop_database
      database_name: "{{ app_name }}_production"

  - name: Recreate production database
    module: db/postgresql
    vars:
      command: create_database
      database_name: "{{ app_name }}_production"
      owner: myapp_prod_user

  # 4. Restore staging backup to production
  - name: Restore staging to production
    module: db/postgresql
    vars:
      command: restore_database
      database_name: "{{ app_name }}_production"
      backup_file: "{{ backup_dir }}/{{ app_name }}_staging_{{ timestamp }}.dump"

  # 5. Verify
  - name: Verify production database
    module: db/postgresql
    vars:
      command: show_database_size
      database_name: "{{ app_name }}_production"
```

---

## Disaster Recovery

### Full Disaster Recovery Plan

```yaml
name: Disaster Recovery - Full Restore
inventory: @group:db-servers-dr

vars:
  backup_location: /backups/disaster_recovery
  databases:
    - name: myapp_production
      backup: myapp_production_latest.dump
      owner: myapp_user
    - name: analytics_db
      backup: analytics_db_latest.dump
      owner: analytics_user

run:
  # 1. Verify backup files exist
  - name: Verify backup files
    loop: "{{ databases }}"
    shell: ls -lh {{ backup_location }}/{{ item.backup }}

  # 2. Create users
  - name: Create database users
    loop: "{{ databases }}"
    module: db/postgresql
    vars:
      command: create_user
      user_name: "{{ item.owner }}"
      user_password: "{{ lookup('vault', item.owner + '_password') }}"

  # 3. Create databases
  - name: Create databases
    loop: "{{ databases }}"
    module: db/postgresql
    vars:
      command: create_database
      database_name: "{{ item.name }}"
      owner: "{{ item.owner }}"

  # 4. Restore from backups
  - name: Restore databases
    loop: "{{ databases }}"
    module: db/postgresql
    vars:
      command: restore_database
      database_name: "{{ item.name }}"
      backup_file: "{{ backup_location }}/{{ item.backup }}"

  # 5. Verify restoration
  - name: Verify database sizes
    loop: "{{ databases }}"
    module: db/postgresql
    vars:
      command: show_database_size
      database_name: "{{ item.name }}"

  # 6. Analyze databases
  - name: Analyze restored databases
    loop: "{{ databases }}"
    module: db/postgresql
    vars:
      command: analyze_table
      postgresql_database: "{{ item.name }}"
```

### Database Replication Setup

```yaml
name: Setup Database Replication User
inventory: @group:db-primary

vars:
  replication_user: replicator
  replication_password: "{{ vault_replication_password }}"

run:
  # Create replication user
  - name: Create replication user
    module: db/postgresql
    vars:
      command: execute_query
      postgresql_database: postgres
      query: |
        CREATE USER {{ replication_user }} REPLICATION LOGIN
        ENCRYPTED PASSWORD '{{ replication_password }}'

  # Verify user
  - name: Verify replication user
    module: db/postgresql
    vars:
      command: list_users
```

---

## Best Practices

### Pre-Deployment Checklist

```yaml
name: Pre-Deployment Database Check
inventory: @group:db-servers

run:
  # 1. Backup current state
  - name: Create pre-deployment backup
    module: db/postgresql
    vars:
      command: backup_database
      database_name: myapp_production
      backup_file: /backups/pre_deploy_{{ timestamp }}.dump
      format: custom

  # 2. Check database size
  - name: Check current database size
    module: db/postgresql
    vars:
      command: show_database_size
      database_name: myapp_production

  # 3. Check active connections
  - name: Monitor active connections
    module: db/postgresql
    vars:
      command: show_activity
      postgresql_database: myapp_production

  # 4. Verify all extensions
  - name: List current extensions
    module: db/postgresql
    vars:
      command: list_extensions
      postgresql_database: myapp_production

  # 5. List all users
  - name: Audit database users
    module: db/postgresql
    vars:
      command: list_users
```

### Post-Deployment Verification

```yaml
name: Post-Deployment Database Verification
inventory: @group:db-servers

run:
  # 1. Verify migrations applied
  - name: Check schema version
    module: db/postgresql
    vars:
      command: execute_query
      postgresql_database: myapp_production
      query: "SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1"

  # 2. Verify data integrity
  - name: Run integrity checks
    module: db/postgresql
    vars:
      command: execute_query
      postgresql_database: myapp_production
      query: "SELECT COUNT(*) FROM app.users WHERE created_at > NOW() - INTERVAL '1 day'"

  # 3. Analyze for performance
  - name: Analyze database
    module: db/postgresql
    vars:
      command: analyze_table
      postgresql_database: myapp_production

  # 4. Check for errors
  - name: Show recent activity
    module: db/postgresql
    vars:
      command: show_activity
      postgresql_database: myapp_production
```

---

## See Also

- [README.md](README.md) - Module overview and quick start
- [COMMANDS.md](COMMANDS.md) - Detailed command reference
- [PostgreSQL Best Practices](https://www.postgresql.org/docs/current/tutorial.html)
