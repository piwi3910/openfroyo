# OpenFroyo Database Modules

This directory contains modules for managing relational databases and database clusters.

## Available Modules (3 Total)

### 1. mariadb (MariaDB Database)
**Purpose:** Complete MariaDB database server management

**Key Features:**
- 25+ operations
- Database and table management
- User and privilege management
- Query execution
- Replication (master/slave)
- Backup and restore (mysqldump)
- Table optimization and repair

**Compatibility:** MariaDB 10.x, MySQL 5.7+

**Size:** 3.6MB

**Transport:** mysql CLI client

---

### 2. postgresql (PostgreSQL Database)
**Purpose:** Complete PostgreSQL database server management

**Key Features:**
- 30+ operations
- Database and schema management
- User and role management
- Extension management (uuid-ossp, PostGIS, pg_trgm)
- Query execution
- Table maintenance (VACUUM, ANALYZE)
- Backup and restore (pg_dump, pg_restore)
- Multiple backup formats (plain, custom, directory, tar)

**Compatibility:** PostgreSQL 10+, 11, 12, 13, 14, 15, 16

**Size:** 3.3MB

**Transport:** psql CLI client

---

### 3. galeradb (Galera Cluster)
**Purpose:** Galera multi-master cluster management

**Key Features:**
- 20 operations
- Cluster bootstrap and recovery
- Node management (join, leave, status)
- Replication monitoring (flow control, lag, conflicts)
- Split-brain detection and resolution
- grastate.dat management
- wsrep status monitoring

**Compatibility:** MariaDB Galera Cluster, Percona XtraDB Cluster

**Size:** 3.5MB

**Transport:** mysql CLI + systemctl

---

## Comparison Matrix

| Feature | MariaDB | PostgreSQL | GaleraDB |
|---------|---------|------------|----------|
| **Type** | Single-node | Single-node | Multi-master cluster |
| Database Management | ✅ (5 ops) | ✅ (6 ops) | ❌ (uses MariaDB) |
| User Management | ✅ (7 ops) | ✅ (8 ops) | ❌ (uses MariaDB) |
| Schema Management | ❌ | ✅ (4 ops) | ❌ |
| Extension Management | ❌ | ✅ (3 ops) | ❌ |
| Table Management | ✅ (4 ops) | ✅ (4 ops) | ❌ |
| Query Execution | ✅ (3 ops) | ✅ (3 ops) | ❌ |
| Replication | ✅ Master/Slave (4 ops) | ✅ Streaming (docs) | ✅ Multi-master (9 ops) |
| Backup/Restore | ✅ (2 ops) | ✅ (2 ops, 4 formats) | ❌ (uses MariaDB) |
| Cluster Management | ❌ | ❌ | ✅ (20 ops) |

## When to Use Each Module

### Use **mariadb** when:
- You need a drop-in MySQL replacement
- Running web applications (WordPress, Drupal)
- Need master/slave replication
- Require MySQL compatibility
- Want simple database management
- Running small to medium applications

### Use **postgresql** when:
- You need advanced SQL features
- Require ACID compliance and reliability
- Need complex queries and analytics
- Want extension support (PostGIS, full-text search)
- Need multiple schema support
- Require advanced data types (JSON, arrays, hstore)
- Running enterprise applications

### Use **galeradb** when:
- You need high availability (multi-master)
- Require automatic failover
- Need synchronous replication
- Want zero data loss
- Require read/write scaling
- Need continuous availability
- Running mission-critical applications

## Common Usage Patterns

### MariaDB Application Setup

```yaml
# Create database and user
- module: db/mariadb
  vars:
    mariadb_host: localhost
    mariadb_user: root
    mariadb_password: "{{ vault_root_password }}"
    command: create_database
    database_name: wordpress_db
    charset: utf8mb4
    collation: utf8mb4_unicode_ci

- module: db/mariadb
  vars:
    command: create_user
    user_name: wp_user
    user_password: "{{ vault_wp_password }}"
    user_host: localhost

- module: db/mariadb
  vars:
    command: grant_privileges
    database_name: wordpress_db
    user_name: wp_user
    user_host: localhost
    privileges: ALL
```

### PostgreSQL with Extensions

```yaml
# Create database with extensions
- module: db/postgresql
  vars:
    postgresql_host: localhost
    postgresql_user: postgres
    postgresql_password: "{{ vault_pg_password }}"
    command: create_database
    database_name: myapp_db
    owner: postgres
    encoding: UTF8

- module: db/postgresql
  vars:
    command: create_extension
    database_name: myapp_db
    extension_name: uuid-ossp

- module: db/postgresql
  vars:
    command: create_extension
    database_name: myapp_db
    extension_name: pg_trgm

- module: db/postgresql
  vars:
    command: create_user
    user_name: app_user
    user_password: "{{ vault_app_password }}"

- module: db/postgresql
  vars:
    command: grant_privileges
    database_name: myapp_db
    user_name: app_user
    privileges: ALL
```

### Galera Cluster Setup

```yaml
# Bootstrap first node
- module: db/galeradb
  vars:
    galera_host: node1.example.com
    galera_user: root
    galera_password: "{{ vault_galera_password }}"
    command: bootstrap_cluster

# Join additional nodes
- module: db/galeradb
  vars:
    galera_host: node2.example.com
    command: join_cluster
    cluster_address: "gcomm://node1,node2,node3"

- module: db/galeradb
  vars:
    galera_host: node3.example.com
    command: join_cluster
    cluster_address: "gcomm://node1,node2,node3"

# Monitor cluster health
- module: db/galeradb
  vars:
    galera_host: node1.example.com
    command: show_cluster_status
```

## High Availability Architectures

### MariaDB Master-Slave Replication

```yaml
# Configure master
- module: db/mariadb
  vars:
    mariadb_host: master.example.com
    command: show_master_status

# Configure slave
- module: db/mariadb
  vars:
    mariadb_host: slave.example.com
    command: start_slave
```

### PostgreSQL Streaming Replication

```yaml
# Primary server setup (via configuration files)
# Standby server setup (via pg_basebackup + recovery.conf)
# Then monitor with PostgreSQL module
- module: db/postgresql
  vars:
    command: show_activity
```

### Galera Multi-Master Cluster

```yaml
# 3-node cluster with automatic synchronous replication
# All nodes are read/write
# Automatic failover
# No single point of failure
```

## Backup and Recovery

### MariaDB Backup

```yaml
# Full database backup
- module: db/mariadb
  vars:
    command: backup_database
    database_name: wordpress_db
    backup_file: /backup/wordpress_db_{{ ansible_date_time.date }}.sql

# Restore database
- module: db/mariadb
  vars:
    command: restore_database
    database_name: wordpress_db
    backup_file: /backup/wordpress_db_2024-11-17.sql
```

### PostgreSQL Backup (Multiple Formats)

```yaml
# Plain SQL backup
- module: db/postgresql
  vars:
    command: backup_database
    database_name: myapp_db
    backup_file: /backup/myapp_db.sql
    format: plain

# Custom format (compressed, supports parallel restore)
- module: db/postgresql
  vars:
    command: backup_database
    database_name: myapp_db
    backup_file: /backup/myapp_db.dump
    format: custom

# Directory format (supports parallel backup/restore)
- module: db/postgresql
  vars:
    command: backup_database
    database_name: myapp_db
    backup_file: /backup/myapp_db_dir
    format: directory
```

## Security Best Practices

### Common Security Measures
- Use strong passwords (minimum 16 characters)
- Create application-specific users (not root/postgres)
- Grant minimum required privileges
- Use localhost connections when possible
- Enable SSL/TLS for remote connections
- Regular security updates
- Audit user access regularly

### MariaDB Security
- Remove anonymous users
- Disable remote root login
- Use mysql_secure_installation
- Enable slow query log for monitoring
- Set max_connections appropriately

### PostgreSQL Security
- Use .pgpass file for credentials
- Configure pg_hba.conf restrictively
- Use SSL certificates
- Enable connection logging
- Set statement_timeout to prevent DoS

### Galera Security
- Secure wsrep SST with SSL
- Use strong cluster passwords
- Isolate cluster traffic on private network
- Monitor for split-brain conditions

## Performance Tuning

### MariaDB Performance
- Optimize InnoDB buffer pool size
- Enable query cache (if beneficial)
- Regular OPTIMIZE TABLE
- Use indexes effectively
- Monitor slow query log

### PostgreSQL Performance
- Tune shared_buffers (25% of RAM)
- Configure work_mem appropriately
- Regular VACUUM and ANALYZE
- Use connection pooling (PgBouncer)
- Monitor with pg_stat_statements

### Galera Performance
- Monitor flow control (wsrep_flow_control_paused)
- Check replication lag
- Optimize wsrep_slave_threads
- Use appropriate write-set size
- Monitor certification conflicts

## Monitoring and Metrics

### Health Checks

**MariaDB:**
```yaml
- module: db/mariadb
  vars:
    command: show_processlist

- module: db/mariadb
  vars:
    command: show_database_size
    database_name: myapp_db
```

**PostgreSQL:**
```yaml
- module: db/postgresql
  vars:
    command: show_activity

- module: db/postgresql
  vars:
    command: show_database_size
    database_name: myapp_db
```

**Galera:**
```yaml
- module: db/galeradb
  vars:
    command: show_cluster_status

- module: db/galeradb
  vars:
    command: show_flow_control

- module: db/galeradb
  vars:
    command: show_replication_status
```

## Troubleshooting

### MariaDB Issues

**Connection Failures:**
- Check MariaDB service status
- Verify user credentials
- Check host/port configuration
- Review firewall rules
- Check max_connections limit

**Replication Issues:**
- Check master status
- Verify slave IO and SQL threads
- Review binary log position
- Check for replication errors

### PostgreSQL Issues

**Connection Failures:**
- Verify PostgreSQL service running
- Check pg_hba.conf configuration
- Verify user credentials
- Test with psql command
- Review PostgreSQL logs

**Performance Issues:**
- Run VACUUM ANALYZE
- Check for bloated tables
- Review slow queries
- Monitor connection pool
- Check disk I/O

### Galera Issues

**Node Sync Problems:**
- Check wsrep_local_state_comment
- Verify cluster address configuration
- Review Galera logs
- Check network connectivity between nodes
- Verify firewall allows Galera ports (4567, 4568, 4444)

**Split-Brain:**
- Check wsrep_cluster_status (should be "Primary")
- Review cluster size
- Follow recovery procedures in documentation

## Documentation

Each module includes comprehensive documentation:
- **README.md** - Complete usage guide
- **COMMANDS.md** - Detailed command reference
- **WORKFLOWS.md** - Real-world scenarios
- **test.ofy** - Example test stack

## Module Statistics

- **3 database modules**
- **75+ total operations**
- **~10.4MB total WASM binaries** (source only, binaries in .gitignore)
- **Complete database coverage** (single-node + high-availability)
- **Production-ready** with comprehensive documentation

## Future Enhancements

Potential additions:
- MongoDB module
- Redis module
- MySQL module (distinct from MariaDB)
- Oracle Database module
- SQL Server module
- Cassandra module
- Elasticsearch module
