# MariaDB Module - Common Workflows

This document provides complete, production-ready workflows for common MariaDB management scenarios.

## Table of Contents

- [New Application Setup](#new-application-setup)
- [User and Privilege Management](#user-and-privilege-management)
- [Backup Strategies](#backup-strategies)
- [Database Migration](#database-migration)
- [Master-Slave Replication](#master-slave-replication)
- [Maintenance and Optimization](#maintenance-and-optimization)
- [Monitoring and Health Checks](#monitoring-and-health-checks)
- [Disaster Recovery](#disaster-recovery)

---

## New Application Setup

Complete workflow for setting up a new application database with proper users and privileges.

### Scenario: WordPress Installation

```yaml
name: WordPress Database Setup
inventory: inventory/
defaults:
  mariadb_host: localhost
  mariadb_user: root
  mariadb_password: "{{ env.MARIADB_ROOT_PASSWORD }}"

run:
  # 1. Create the WordPress database
  - run: db/mariadb
    on: web-db-01
    vars:
      command: create_database
      database_name: wordpress_production
      charset: utf8mb4
      collation: utf8mb4_unicode_ci

  # 2. Create WordPress user
  - run: db/mariadb
    on: web-db-01
    vars:
      command: create_user
      user_name: wp_user
      user_password: "{{ env.WP_DB_PASSWORD }}"
      user_host: "192.168.1.%"  # Allow from web server subnet

  # 3. Grant privileges to WordPress user
  - run: db/mariadb
    on: web-db-01
    vars:
      command: grant_privileges
      database_name: wordpress_production
      user_name: wp_user
      user_host: "192.168.1.%"
      privileges: ALL

  # 4. Verify setup
  - run: db/mariadb
    on: web-db-01
    vars:
      command: show_user_grants
      user_name: wp_user
      user_host: "192.168.1.%"
```

### Scenario: Microservices Setup

```yaml
name: Microservices Database Setup
inventory: inventory/
defaults:
  mariadb_host: localhost
  mariadb_user: root
  mariadb_password: "{{ env.MARIADB_ROOT_PASSWORD }}"

run:
  # Create databases for each service
  - run: db/mariadb
    on: @db-cluster
    vars:
      command: create_database
      database_name: "{{ item }}"
    loop:
      - auth_service
      - user_service
      - order_service
      - inventory_service

  # Create service-specific users with limited privileges
  - run: db/mariadb
    on: @db-cluster
    vars:
      command: create_user
      user_name: auth_user
      user_password: "{{ vault.auth_db_password }}"
      user_host: "%"

  - run: db/mariadb
    on: @db-cluster
    vars:
      command: grant_privileges
      database_name: auth_service
      user_name: auth_user
      user_host: "%"
      privileges: SELECT,INSERT,UPDATE,DELETE

  # Repeat for other services...
```

---

## User and Privilege Management

### Scenario: Create Read-Only User

```yaml
name: Create Read-Only User
inventory: inventory/
defaults:
  mariadb_user: root
  mariadb_password: "{{ env.MARIADB_ROOT_PASSWORD }}"

run:
  # Create read-only user for reporting
  - run: db/mariadb
    on: db-primary
    vars:
      command: create_user
      user_name: readonly_reports
      user_password: "{{ vault.readonly_password }}"
      user_host: "10.0.2.%"  # Reports subnet

  # Grant SELECT only
  - run: db/mariadb
    on: db-primary
    vars:
      command: grant_privileges
      database_name: analytics_db
      user_name: readonly_reports
      user_host: "10.0.2.%"
      privileges: SELECT

  # Verify grants
  - run: db/mariadb
    on: db-primary
    vars:
      command: show_user_grants
      user_name: readonly_reports
      user_host: "10.0.2.%"
```

### Scenario: Rotate Application Password

```yaml
name: Rotate Application Database Password
inventory: inventory/
defaults:
  mariadb_user: root
  mariadb_password: "{{ env.MARIADB_ROOT_PASSWORD }}"

run:
  # Change password for application user
  - run: db/mariadb
    on: @db-servers
    vars:
      command: change_password
      user_name: app_user
      user_host: "%"
      user_password: "{{ vault.new_app_password }}"

  # Note: Update application configuration separately
```

### Scenario: Revoke Privileges for Departing Team Member

```yaml
name: Remove Developer Database Access
inventory: inventory/
defaults:
  mariadb_user: root
  mariadb_password: "{{ env.MARIADB_ROOT_PASSWORD }}"

run:
  # Revoke all privileges
  - run: db/mariadb
    on: @db-servers
    vars:
      command: revoke_privileges
      database_name: "*"
      user_name: john_doe
      user_host: "%"
      privileges: ALL

  # Drop the user
  - run: db/mariadb
    on: @db-servers
    vars:
      command: drop_user
      user_name: john_doe
      user_host: "%"
```

---

## Backup Strategies

### Scenario: Daily Automated Backups

```yaml
name: Daily Database Backup
inventory: inventory/
defaults:
  mariadb_user: backup_user
  mariadb_password: "{{ env.BACKUP_PASSWORD }}"
  backup_timestamp: "{{ timestamp('%Y%m%d_%H%M%S') }}"

run:
  # List all databases to backup
  - run: db/mariadb
    on: db-primary
    vars:
      command: list_databases

  # Backup each production database
  - run: db/mariadb
    on: db-primary
    vars:
      command: backup_database
      database_name: production_app
      backup_file: "/backups/production_app_{{ backup_timestamp }}.sql"

  - run: db/mariadb
    on: db-primary
    vars:
      command: backup_database
      database_name: user_data
      backup_file: "/backups/user_data_{{ backup_timestamp }}.sql"

  # Compress backups
  - run: system/file
    on: db-primary
    vars:
      command: compress
      path: /backups/*.sql
      method: gzip

  # Copy to remote backup location
  - run: system/copy
    on: db-primary
    vars:
      src: /backups/*.sql.gz
      dest: backup-server:/mnt/db-backups/
      remote: true

  # Clean up old backups (keep 30 days)
  - run: system/file
    on: db-primary
    vars:
      command: delete_older_than
      path: /backups/
      days: 30
```

### Scenario: Pre-Migration Backup

```yaml
name: Pre-Migration Backup
inventory: inventory/
defaults:
  mariadb_user: root
  mariadb_password: "{{ env.MARIADB_ROOT_PASSWORD }}"

run:
  # Create timestamped backup before major changes
  - run: db/mariadb
    on: db-primary
    vars:
      command: backup_database
      database_name: production_app
      backup_file: "/backups/pre_migration_{{ timestamp }}.sql"

  # Verify backup file exists and has content
  - run: system/file
    on: db-primary
    vars:
      command: stat
      path: "/backups/pre_migration_{{ timestamp }}.sql"

  # Store backup hash for integrity verification
  - run: system/checksum
    on: db-primary
    vars:
      path: "/backups/pre_migration_{{ timestamp }}.sql"
      algorithm: sha256
```

---

## Database Migration

### Scenario: Development to Production Migration

```yaml
name: Database Migration - Dev to Prod
inventory: inventory/
defaults:
  mariadb_user: root
  mariadb_password: "{{ env.MARIADB_ROOT_PASSWORD }}"

run:
  # 1. Backup production database first
  - run: db/mariadb
    on: prod-db
    vars:
      command: backup_database
      database_name: production_app
      backup_file: "/backups/prod_pre_migration.sql"

  # 2. Export development schema
  - run: db/mariadb
    on: dev-db
    vars:
      command: backup_database
      database_name: development_app
      backup_file: "/tmp/dev_schema.sql"

  # 3. Copy schema to production
  - run: system/copy
    vars:
      src: dev-db:/tmp/dev_schema.sql
      dest: prod-db:/tmp/schema_update.sql

  # 4. Apply schema updates to production
  - run: db/mariadb
    on: prod-db
    vars:
      command: execute_script
      database_name: production_app
      script_file: /tmp/schema_update.sql

  # 5. Verify table structure
  - run: db/mariadb
    on: prod-db
    vars:
      command: list_tables
      database_name: production_app

  # 6. Cleanup temporary files
  - run: system/file
    on: prod-db
    vars:
      command: delete
      path: /tmp/schema_update.sql
```

### Scenario: Schema Update Workflow

```yaml
name: Apply Schema Updates
inventory: inventory/
defaults:
  mariadb_user: root
  mariadb_password: "{{ env.MARIADB_ROOT_PASSWORD }}"

run:
  # 1. Backup before schema changes
  - run: db/mariadb
    on: @production-db
    vars:
      command: backup_database
      database_name: myapp_db
      backup_file: "/backups/pre_schema_{{ timestamp }}.sql"

  # 2. Copy migration script to server
  - run: system/copy
    vars:
      src: ./migrations/v2.0_schema.sql
      dest: "@production-db:/tmp/migration.sql"

  # 3. Execute migration
  - run: db/mariadb
    on: @production-db
    vars:
      command: execute_script
      database_name: myapp_db
      script_file: /tmp/migration.sql

  # 4. Verify new tables exist
  - run: db/mariadb
    on: @production-db
    vars:
      command: describe_table
      database_name: myapp_db
      table_name: new_features_table

  # 5. Optimize tables after changes
  - run: db/mariadb
    on: @production-db
    vars:
      command: optimize_table
      database_name: myapp_db
      table_name: "{{ item }}"
    loop:
      - users
      - new_features_table
```

---

## Master-Slave Replication

### Scenario: Setup Master-Slave Replication

```yaml
name: Configure MariaDB Replication
inventory: inventory/
defaults:
  mariadb_user: root
  mariadb_password: "{{ env.MARIADB_ROOT_PASSWORD }}"

run:
  # 1. Create replication user on master
  - run: db/mariadb
    on: db-master
    vars:
      command: create_user
      user_name: replication
      user_password: "{{ vault.replication_password }}"
      user_host: "10.0.1.%"

  # 2. Grant replication privileges
  - run: db/mariadb
    on: db-master
    vars:
      command: execute_query
      query: "GRANT REPLICATION SLAVE ON *.* TO 'replication'@'10.0.1.%'"

  # 3. Get master status
  - run: db/mariadb
    on: db-master
    vars:
      command: show_master_status
    # Save output: master_log_file, master_log_pos

  # 4. Backup master database
  - run: db/mariadb
    on: db-master
    vars:
      command: backup_database
      database_name: production_db
      backup_file: /tmp/master_backup.sql

  # 5. Copy backup to slave
  - run: system/copy
    vars:
      src: db-master:/tmp/master_backup.sql
      dest: db-slave:/tmp/master_backup.sql

  # 6. Restore on slave
  - run: db/mariadb
    on: db-slave
    vars:
      command: restore_database
      database_name: production_db
      backup_file: /tmp/master_backup.sql

  # 7. Configure slave (manual step - use execute_query)
  - run: db/mariadb
    on: db-slave
    vars:
      command: execute_query
      query: >
        CHANGE MASTER TO
        MASTER_HOST='10.0.1.10',
        MASTER_USER='replication',
        MASTER_PASSWORD='{{ vault.replication_password }}',
        MASTER_LOG_FILE='mysql-bin.000001',
        MASTER_LOG_POS=154

  # 8. Start slave
  - run: db/mariadb
    on: db-slave
    vars:
      command: start_slave

  # 9. Verify replication
  - run: db/mariadb
    on: db-slave
    vars:
      command: show_slave_status
```

### Scenario: Replication Health Check

```yaml
name: Check Replication Health
inventory: inventory/
defaults:
  mariadb_user: root
  mariadb_password: "{{ env.MARIADB_ROOT_PASSWORD }}"

run:
  # Check master status
  - run: db/mariadb
    on: db-master
    vars:
      command: show_master_status

  # Check slave status on all slaves
  - run: db/mariadb
    on: @db-slaves
    vars:
      command: show_slave_status

  # Look for:
  # - Slave_IO_Running: Yes
  # - Slave_SQL_Running: Yes
  # - Seconds_Behind_Master: 0 (or small number)
  # - Last_Error: (empty)
```

### Scenario: Promote Slave to Master

```yaml
name: Promote Slave to Master (Failover)
inventory: inventory/
defaults:
  mariadb_user: root
  mariadb_password: "{{ env.MARIADB_ROOT_PASSWORD }}"

run:
  # 1. Stop slave replication
  - run: db/mariadb
    on: db-slave-01
    vars:
      command: stop_slave

  # 2. Verify all changes are applied
  - run: db/mariadb
    on: db-slave-01
    vars:
      command: show_slave_status
    # Check Seconds_Behind_Master is 0

  # 3. Reset slave configuration
  - run: db/mariadb
    on: db-slave-01
    vars:
      command: execute_query
      query: "RESET SLAVE ALL"

  # 4. Enable binary logging on new master (if not already)
  # (Requires my.cnf configuration - manual step)

  # 5. Point other slaves to new master
  # (Repeat configuration steps from master-slave setup)

  # 6. Update application to use new master
  # (Manual configuration update)
```

---

## Maintenance and Optimization

### Scenario: Weekly Table Optimization

```yaml
name: Weekly Database Maintenance
inventory: inventory/
defaults:
  mariadb_user: root
  mariadb_password: "{{ env.MARIADB_ROOT_PASSWORD }}"

run:
  # List all tables
  - run: db/mariadb
    on: db-primary
    vars:
      command: list_tables
      database_name: production_app

  # Optimize frequently updated tables
  - run: db/mariadb
    on: db-primary
    vars:
      command: optimize_table
      database_name: production_app
      table_name: "{{ item }}"
    loop:
      - user_sessions
      - activity_log
      - cached_data
      - search_index

  # Check table sizes after optimization
  - run: db/mariadb
    on: db-primary
    vars:
      command: show_database_size
      database_name: production_app
```

### Scenario: Database Health Check

```yaml
name: Database Health Check
inventory: inventory/
defaults:
  mariadb_user: root
  mariadb_password: "{{ env.MARIADB_ROOT_PASSWORD }}"

run:
  # Check database size
  - run: db/mariadb
    on: @db-servers
    vars:
      command: show_database_size
      database_name: production_app

  # Check active connections
  - run: db/mariadb
    on: @db-servers
    vars:
      command: show_processlist

  # Verify critical users exist
  - run: db/mariadb
    on: @db-servers
    vars:
      command: list_users

  # Check replication status (if applicable)
  - run: db/mariadb
    on: @db-slaves
    vars:
      command: show_slave_status
```

---

## Monitoring and Health Checks

### Scenario: Automated Monitoring

```yaml
name: Database Monitoring
inventory: inventory/
defaults:
  mariadb_user: monitor
  mariadb_password: "{{ env.MONITOR_PASSWORD }}"

run:
  # Monitor connection count
  - run: db/mariadb
    on: @db-servers
    vars:
      command: show_processlist

  # Monitor database sizes
  - run: db/mariadb
    on: @db-servers
    vars:
      command: show_database_size
      database_name: "{{ item }}"
    loop:
      - production_app
      - analytics_db
      - logs_db

  # Check for long-running queries
  - run: db/mariadb
    on: @db-servers
    vars:
      command: execute_query
      query: >
        SELECT * FROM information_schema.processlist
        WHERE command != 'Sleep' AND time > 300

  # Monitor replication lag
  - run: db/mariadb
    on: @db-slaves
    vars:
      command: show_slave_status
```

---

## Disaster Recovery

### Scenario: Point-in-Time Recovery

```yaml
name: Point-in-Time Recovery
inventory: inventory/
defaults:
  mariadb_user: root
  mariadb_password: "{{ env.MARIADB_ROOT_PASSWORD }}"

run:
  # 1. Stop application access
  # (Manual step or use application module)

  # 2. Create current state backup
  - run: db/mariadb
    on: db-primary
    vars:
      command: backup_database
      database_name: production_app
      backup_file: "/backups/before_recovery_{{ timestamp }}.sql"

  # 3. Restore from last good backup
  - run: db/mariadb
    on: db-primary
    vars:
      command: restore_database
      database_name: production_app
      backup_file: /backups/production_app_20240115_020000.sql

  # 4. Apply binary logs up to specific point
  # (Manual step using mysqlbinlog)

  # 5. Verify data integrity
  - run: db/mariadb
    on: db-primary
    vars:
      command: execute_query
      database_name: production_app
      query: "SELECT COUNT(*) FROM users"

  # 6. Resume application access
```

### Scenario: Database Corruption Recovery

```yaml
name: Recover from Table Corruption
inventory: inventory/
defaults:
  mariadb_user: root
  mariadb_password: "{{ env.MARIADB_ROOT_PASSWORD }}"

run:
  # 1. Backup current state (even if corrupted)
  - run: db/mariadb
    on: db-primary
    vars:
      command: backup_database
      database_name: production_app
      backup_file: "/backups/corrupted_state_{{ timestamp }}.sql"
    continue_on_error: true

  # 2. Attempt table repair
  - run: db/mariadb
    on: db-primary
    vars:
      command: repair_table
      database_name: production_app
      table_name: corrupted_table

  # 3. If repair fails, restore from backup
  - run: db/mariadb
    on: db-primary
    vars:
      command: restore_database
      database_name: production_app
      backup_file: /backups/last_known_good.sql
    when: repair_failed

  # 4. Verify table integrity
  - run: db/mariadb
    on: db-primary
    vars:
      command: describe_table
      database_name: production_app
      table_name: corrupted_table

  # 5. Optimize recovered table
  - run: db/mariadb
    on: db-primary
    vars:
      command: optimize_table
      database_name: production_app
      table_name: corrupted_table
```

---

## Best Practices Summary

### Always Do:
- Backup before major changes
- Test restores regularly
- Use strong passwords
- Restrict user hosts
- Monitor replication lag
- Optimize tables regularly
- Document all changes
- Use version control for scripts

### Never Do:
- Hardcode passwords in scripts
- Grant ALL to application users
- Skip backups before migrations
- Ignore replication errors
- Run untested queries on production
- Allow root access from anywhere
- Forget to flush privileges after changes

### Regular Tasks:
- Daily: Automated backups
- Weekly: Table optimization
- Weekly: Replication health checks
- Monthly: Backup restoration tests
- Monthly: User access review
- Quarterly: Performance analysis
- Yearly: Disaster recovery drills

---

## Additional Resources

- [README.md](README.md) - Module overview
- [COMMANDS.md](COMMANDS.md) - Complete command reference
- MariaDB Documentation: https://mariadb.com/kb/
- Replication Guide: https://mariadb.com/kb/en/replication/
- Backup Best Practices: https://mariadb.com/kb/en/backup-and-restore/
