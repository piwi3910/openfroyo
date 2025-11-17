# MariaDB Module - Command Reference

Complete reference for all 25+ commands available in the MariaDB module.

## Table of Contents

- [Database Management](#database-management)
- [User Management](#user-management)
- [Table Management](#table-management)
- [Query Execution](#query-execution)
- [Replication](#replication)
- [Backup and Restore](#backup-and-restore)

---

## Database Management

### create_database

Create a new database with specified character set and collation.

**Required Variables:**
- `database_name` - Name of the database to create

**Optional Variables:**
- `charset` - Character set (default: `utf8mb4`)
- `collation` - Collation (default: `utf8mb4_unicode_ci`)

**Returns:**
- `status`: `ok` (already exists) or `changed` (created)
- `facts.database` - Database name
- `facts.created` - Boolean indicating if created

**Example:**
```yaml
- run: db/mariadb
  vars:
    command: create_database
    database_name: myapp_production
    charset: utf8mb4
    collation: utf8mb4_unicode_ci
```

**Idempotency:** Yes - will not fail if database already exists

---

### drop_database

Drop an existing database.

**Required Variables:**
- `database_name` - Name of the database to drop

**Returns:**
- `status`: `ok` (doesn't exist) or `changed` (dropped)
- `facts.database` - Database name
- `facts.dropped` - Boolean indicating if dropped

**Example:**
```yaml
- run: db/mariadb
  vars:
    command: drop_database
    database_name: old_database
```

**Warning:** This operation is destructive and cannot be undone!

**Idempotency:** Yes - will not fail if database doesn't exist

---

### list_databases

List all databases on the server.

**Required Variables:** None

**Returns:**
- `status`: `ok`
- `facts.databases` - Array of database names
- `facts.count` - Number of databases

**Example:**
```yaml
- run: db/mariadb
  vars:
    command: list_databases
```

**Output:**
```
Database
information_schema
mysql
performance_schema
myapp_db
test_db
```

---

### show_database_size

Show the size of a database in megabytes.

**Required Variables:**
- `database_name` - Name of the database

**Returns:**
- `status`: `ok`
- `facts.database` - Database name
- `facts.size_info` - Size information in MB

**Example:**
```yaml
- run: db/mariadb
  vars:
    command: show_database_size
    database_name: myapp_production
```

**Output:**
```
database          size_mb
myapp_production  1523.45
```

---

### check_database_exists

Check if a database exists.

**Required Variables:**
- `database_name` - Name of the database to check

**Returns:**
- `status`: `ok`
- `facts.database` - Database name
- `facts.exists` - Boolean indicating existence

**Example:**
```yaml
- run: db/mariadb
  vars:
    command: check_database_exists
    database_name: myapp_db
```

---

## User Management

### create_user

Create a new database user.

**Required Variables:**
- `user_name` - Username to create
- `user_password` - Password for the user

**Optional Variables:**
- `user_host` - Host from which user can connect (default: `localhost`)

**Returns:**
- `status`: `ok` (already exists) or `changed` (created)
- `facts.user` - Username
- `facts.host` - User host
- `facts.created` - Boolean indicating if created

**Example:**
```yaml
- run: db/mariadb
  vars:
    command: create_user
    user_name: appuser
    user_password: "{{ vault.db_password }}"
    user_host: "%"  # Allow from any host
```

**Security Note:** Always use strong passwords and restrict hosts when possible.

**Idempotency:** Yes - will not fail if user already exists

---

### drop_user

Drop an existing database user.

**Required Variables:**
- `user_name` - Username to drop

**Optional Variables:**
- `user_host` - User host (default: `localhost`)

**Returns:**
- `status`: `ok` (doesn't exist) or `changed` (dropped)
- `facts.user` - Username
- `facts.host` - User host
- `facts.dropped` - Boolean indicating if dropped

**Example:**
```yaml
- run: db/mariadb
  vars:
    command: drop_user
    user_name: old_user
    user_host: localhost
```

**Idempotency:** Yes - will not fail if user doesn't exist

---

### list_users

List all database users.

**Required Variables:** None

**Returns:**
- `status`: `ok`
- `facts.users` - User list with hosts
- `facts.count` - Number of users

**Example:**
```yaml
- run: db/mariadb
  vars:
    command: list_users
```

**Output:**
```
User      Host
root      localhost
appuser   %
readonly  192.168.1.%
```

---

### grant_privileges

Grant privileges to a user on a database.

**Required Variables:**
- `user_name` - Username
- `database_name` - Database name
- `user_host` - User host

**Optional Variables:**
- `privileges` - Privileges to grant (default: `ALL`)

**Returns:**
- `status`: `changed`
- `facts.user` - Username
- `facts.host` - User host
- `facts.database` - Database name
- `facts.privileges` - Granted privileges

**Example:**
```yaml
# Grant all privileges
- run: db/mariadb
  vars:
    command: grant_privileges
    user_name: appuser
    user_host: localhost
    database_name: myapp_db
    privileges: ALL

# Grant specific privileges
- run: db/mariadb
  vars:
    command: grant_privileges
    user_name: readonly_user
    user_host: "%"
    database_name: myapp_db
    privileges: SELECT
```

**Common Privilege Combinations:**
- `ALL` - All privileges
- `SELECT,INSERT,UPDATE,DELETE` - Data manipulation
- `SELECT` - Read-only access
- `CREATE,DROP,ALTER` - Schema modification
- `GRANT OPTION` - Allow granting privileges to others

---

### revoke_privileges

Revoke privileges from a user on a database.

**Required Variables:**
- `user_name` - Username
- `database_name` - Database name
- `user_host` - User host

**Optional Variables:**
- `privileges` - Privileges to revoke (default: `ALL`)

**Returns:**
- `status`: `changed`
- `facts.user` - Username
- `facts.host` - User host
- `facts.database` - Database name
- `facts.privileges` - Revoked privileges

**Example:**
```yaml
- run: db/mariadb
  vars:
    command: revoke_privileges
    user_name: appuser
    user_host: localhost
    database_name: myapp_db
    privileges: DELETE,DROP
```

---

### show_user_grants

Show all privileges granted to a user.

**Required Variables:**
- `user_name` - Username

**Optional Variables:**
- `user_host` - User host (default: `localhost`)

**Returns:**
- `status`: `ok`
- `facts.user` - Username
- `facts.host` - User host
- `facts.grants` - Grant statements

**Example:**
```yaml
- run: db/mariadb
  vars:
    command: show_user_grants
    user_name: appuser
    user_host: localhost
```

**Output:**
```
GRANT USAGE ON *.* TO 'appuser'@'localhost'
GRANT ALL PRIVILEGES ON `myapp_db`.* TO 'appuser'@'localhost'
```

---

### change_password

Change a user's password.

**Required Variables:**
- `user_name` - Username
- `user_password` - New password

**Optional Variables:**
- `user_host` - User host (default: `localhost`)

**Returns:**
- `status`: `changed`
- `facts.user` - Username
- `facts.host` - User host

**Example:**
```yaml
- run: db/mariadb
  vars:
    command: change_password
    user_name: appuser
    user_host: localhost
    user_password: "{{ vault.new_password }}"
```

**Security Note:** Use secure password storage and rotation policies.

---

## Table Management

### list_tables

List all tables in a database.

**Required Variables:**
- `database_name` - Database name

**Returns:**
- `status`: `ok`
- `facts.database` - Database name
- `facts.tables` - Table list
- `facts.count` - Number of tables

**Example:**
```yaml
- run: db/mariadb
  vars:
    command: list_tables
    database_name: myapp_db
```

---

### describe_table

Show the structure of a table.

**Required Variables:**
- `database_name` - Database name
- `table_name` - Table name

**Returns:**
- `status`: `ok`
- `facts.database` - Database name
- `facts.table` - Table name
- `facts.structure` - Table structure

**Example:**
```yaml
- run: db/mariadb
  vars:
    command: describe_table
    database_name: myapp_db
    table_name: users
```

**Output:**
```
Field      Type          Null  Key  Default  Extra
id         int(11)       NO    PRI  NULL     auto_increment
username   varchar(50)   NO    UNI  NULL
email      varchar(100)  NO         NULL
created_at timestamp     YES        NULL
```

---

### optimize_table

Optimize a table to reclaim unused space and defragment data.

**Required Variables:**
- `database_name` - Database name
- `table_name` - Table name

**Returns:**
- `status`: `changed`
- `facts.database` - Database name
- `facts.table` - Table name
- `facts.result` - Optimization result

**Example:**
```yaml
- run: db/mariadb
  vars:
    command: optimize_table
    database_name: myapp_db
    table_name: large_table
```

**Use Cases:**
- After large DELETE operations
- Regular maintenance (weekly/monthly)
- Before backups of large tables

**Note:** Locks the table during optimization. Use during maintenance windows.

---

### repair_table

Repair a corrupted table.

**Required Variables:**
- `database_name` - Database name
- `table_name` - Table name

**Returns:**
- `status`: `changed`
- `facts.database` - Database name
- `facts.table` - Table name
- `facts.result` - Repair result

**Example:**
```yaml
- run: db/mariadb
  vars:
    command: repair_table
    database_name: myapp_db
    table_name: corrupted_table
```

**When to Use:**
- Table corruption detected
- After server crash
- Data inconsistencies

**Warning:** Always backup before attempting repair!

---

## Query Execution

### execute_query

Execute a SQL query.

**Required Variables:**
- `query` - SQL query to execute

**Optional Variables:**
- `database_name` - Database to use

**Returns:**
- `status`: `ok` (SELECT) or `changed` (INSERT/UPDATE/DELETE)
- `facts.query` - Executed query
- `facts.result` - Query result

**Example:**
```yaml
# SELECT query
- run: db/mariadb
  vars:
    command: execute_query
    database_name: myapp_db
    query: "SELECT * FROM users WHERE active = 1"

# UPDATE query
- run: db/mariadb
  vars:
    command: execute_query
    database_name: myapp_db
    query: "UPDATE settings SET value = 'production' WHERE key = 'environment'"
```

**Security Warning:** Ensure queries are properly escaped. Avoid user input in queries.

---

### execute_script

Execute SQL statements from a file.

**Required Variables:**
- `script_file` - Path to SQL script file

**Optional Variables:**
- `database_name` - Database to use

**Returns:**
- `status`: `changed`
- `facts.script` - Script file path
- `facts.result` - Execution result

**Example:**
```yaml
- run: db/mariadb
  vars:
    command: execute_script
    database_name: myapp_db
    script_file: /tmp/schema_updates.sql
```

**Script File Format:**
```sql
-- schema_updates.sql
CREATE TABLE IF NOT EXISTS sessions (
    id VARCHAR(128) PRIMARY KEY,
    data TEXT,
    expires TIMESTAMP
);

CREATE INDEX idx_expires ON sessions(expires);
```

---

### show_processlist

Show currently running database processes.

**Required Variables:** None

**Returns:**
- `status`: `ok`
- `facts.processlist` - Process list
- `facts.count` - Number of processes

**Example:**
```yaml
- run: db/mariadb
  vars:
    command: show_processlist
```

**Output:**
```
Id  User     Host           db        Command  Time  State    Info
1   root     localhost      myapp_db  Query    0     NULL     SHOW FULL PROCESSLIST
45  appuser  app-server:123 myapp_db  Sleep    15    NULL     NULL
```

**Use Cases:**
- Monitoring active connections
- Identifying long-running queries
- Troubleshooting performance issues

---

## Replication

### show_master_status

Show master replication status.

**Required Variables:** None

**Returns:**
- `status`: `ok`
- `facts.master_status` - Master status information

**Example:**
```yaml
- run: db/mariadb
  on: @master-db
  vars:
    command: show_master_status
```

**Output:**
```
File              Position  Binlog_Do_DB  Binlog_Ignore_DB
mysql-bin.000003  73        myapp_db
```

**Use Cases:**
- Setting up new slave servers
- Monitoring replication health
- Point-in-time recovery

---

### show_slave_status

Show slave replication status (detailed).

**Required Variables:** None

**Returns:**
- `status`: `ok`
- `facts.slave_status` - Detailed slave status

**Example:**
```yaml
- run: db/mariadb
  on: @slave-db
  vars:
    command: show_slave_status
```

**Output:**
```
Slave_IO_Running: Yes
Slave_SQL_Running: Yes
Master_Host: 192.168.1.10
Master_User: replication
Master_Port: 3306
Seconds_Behind_Master: 0
```

**Key Fields to Monitor:**
- `Slave_IO_Running` - Should be "Yes"
- `Slave_SQL_Running` - Should be "Yes"
- `Seconds_Behind_Master` - Replication lag
- `Last_Error` - Any replication errors

---

### start_slave

Start slave replication.

**Required Variables:** None

**Returns:**
- `status`: `changed`

**Example:**
```yaml
- run: db/mariadb
  on: @slave-db
  vars:
    command: start_slave
```

**Use Cases:**
- After slave maintenance
- After replication configuration
- Resuming after stopped replication

---

### stop_slave

Stop slave replication.

**Required Variables:** None

**Returns:**
- `status`: `changed`

**Example:**
```yaml
- run: db/mariadb
  on: @slave-db
  vars:
    command: stop_slave
```

**Use Cases:**
- Before slave maintenance
- Before configuration changes
- Pausing replication temporarily

---

## Backup and Restore

### backup_database

Backup a database to a SQL file using mysqldump.

**Required Variables:**
- `database_name` - Database to backup
- `backup_file` - Path for backup file

**Returns:**
- `status`: `changed`
- `facts.database` - Database name
- `facts.backup_file` - Backup file path
- `facts.size_bytes` - Backup file size

**Example:**
```yaml
- run: db/mariadb
  vars:
    command: backup_database
    database_name: myapp_production
    backup_file: /backups/myapp_{{ timestamp }}.sql
```

**Backup Options Used:**
- `--single-transaction` - Consistent snapshot for InnoDB
- `--routines` - Include stored procedures
- `--triggers` - Include triggers

**Best Practices:**
- Include timestamp in filename
- Store in separate backup location
- Verify backup file after creation
- Test restoration periodically
- Implement backup retention policy

---

### restore_database

Restore a database from a SQL backup file.

**Required Variables:**
- `database_name` - Database to restore to
- `backup_file` - Path to backup file

**Returns:**
- `status`: `changed`
- `facts.database` - Database name
- `facts.backup_file` - Backup file path

**Example:**
```yaml
- run: db/mariadb
  vars:
    command: restore_database
    database_name: myapp_production
    backup_file: /backups/myapp_2024-01-15.sql
```

**Important Notes:**
- Database must exist before restore
- Existing data will be replaced
- Large restores may take significant time
- Ensure sufficient disk space

**Recommended Workflow:**
```yaml
# 1. Create database if needed
- run: db/mariadb
  vars:
    command: create_database
    database_name: myapp_production

# 2. Restore from backup
- run: db/mariadb
  vars:
    command: restore_database
    database_name: myapp_production
    backup_file: /backups/myapp_latest.sql
```

---

## Status and Return Values

All commands return consistent status values:

- **ok**: Command executed successfully, no changes made
- **changed**: Command executed successfully, changes made to database
- **failed**: Command failed with error

Each command also returns relevant facts that can be used in subsequent tasks or for reporting.

## Error Handling

Common error scenarios and their meanings:

| Error | Meaning | Solution |
|-------|---------|----------|
| `connection refused` | Cannot connect to MariaDB | Check server is running, verify host/port |
| `access denied` | Authentication failed | Verify username/password |
| `database not found` | Database doesn't exist | Create database first |
| `table not found` | Table doesn't exist | Check table name, create if needed |
| `permission denied` | Insufficient privileges | Grant required privileges |
| `syntax error` | Invalid SQL | Check query syntax |

## Additional Resources

- [README.md](README.md) - Module overview and setup
- [WORKFLOWS.md](WORKFLOWS.md) - Common workflow examples
- MariaDB Documentation: https://mariadb.com/kb/en/documentation/
- MySQL Documentation: https://dev.mysql.com/doc/
