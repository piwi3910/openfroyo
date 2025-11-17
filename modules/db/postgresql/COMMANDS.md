# PostgreSQL Module - Command Reference

Complete reference for all 30+ PostgreSQL module commands.

## Table of Contents

- [Database Management](#database-management)
- [User/Role Management](#userrole-management)
- [Schema Management](#schema-management)
- [Table Management](#table-management)
- [Extension Management](#extension-management)
- [Query Execution](#query-execution)
- [Backup and Restore](#backup-and-restore)

---

## Database Management

### create_database

Create a new PostgreSQL database.

**Variables:**
- `database_name` (required): Name of the database to create
- `owner` (optional): Database owner username
- `encoding` (optional, default: UTF8): Database encoding

**Example:**
```yaml
- module: db/postgresql
  vars:
    command: create_database
    database_name: myapp_production
    owner: appuser
    encoding: UTF8
```

**Output:**
```json
{
  "status": "changed",
  "message": "Database 'myapp_production' created successfully",
  "facts": {
    "database_name": "myapp_production",
    "owner": "appuser",
    "encoding": "UTF8"
  }
}
```

**Idempotency:** Returns `ok` if database already exists.

---

### drop_database

Drop an existing PostgreSQL database.

**Variables:**
- `database_name` (required): Name of the database to drop

**Example:**
```yaml
- module: db/postgresql
  vars:
    command: drop_database
    database_name: old_database
```

**Output:**
```json
{
  "status": "changed",
  "message": "Database 'old_database' dropped successfully"
}
```

**Idempotency:** Returns `ok` if database doesn't exist.

**Warning:** This operation is destructive and cannot be undone.

---

### list_databases

List all non-template databases on the PostgreSQL server.

**Variables:** None (uses connection settings)

**Example:**
```yaml
- module: db/postgresql
  vars:
    command: list_databases
```

**Output:**
```json
{
  "status": "ok",
  "message": "Databases retrieved successfully",
  "facts": {
    "databases": "datname|owner|encoding|size\nmyapp_db|postgres|UTF8|42 MB\ntest_db|appuser|UTF8|15 MB"
  }
}
```

---

### rename_database

Rename an existing database.

**Variables:**
- `old_name` (required): Current database name
- `new_name` (required): New database name

**Example:**
```yaml
- module: db/postgresql
  vars:
    command: rename_database
    old_name: myapp_dev
    new_name: myapp_staging
```

**Output:**
```json
{
  "status": "changed",
  "message": "Database renamed from 'myapp_dev' to 'myapp_staging'",
  "facts": {
    "old_name": "myapp_dev",
    "new_name": "myapp_staging"
  }
}
```

**Note:** No active connections should exist to the database being renamed.

---

### show_database_size

Show the size of a database in human-readable format.

**Variables:**
- `database_name` (required): Name of the database

**Example:**
```yaml
- module: db/postgresql
  vars:
    command: show_database_size
    database_name: myapp_db
```

**Output:**
```json
{
  "status": "ok",
  "message": "Database 'myapp_db' size retrieved",
  "facts": {
    "database_name": "myapp_db",
    "size": "42 MB"
  }
}
```

---

### check_database_exists

Check if a database exists on the server.

**Variables:**
- `database_name` (required): Name of the database to check

**Example:**
```yaml
- module: db/postgresql
  vars:
    command: check_database_exists
    database_name: myapp_db
```

**Output:**
```json
{
  "status": "ok",
  "message": "Database 'myapp_db' existence checked",
  "facts": {
    "database_name": "myapp_db",
    "exists": true
  }
}
```

---

## User/Role Management

### create_user

Create a new PostgreSQL user with specified privileges.

**Variables:**
- `user_name` (required): Username to create
- `user_password` (optional): User password
- `superuser` (optional, default: false): Grant superuser privilege
- `createdb` (optional, default: false): Grant create database privilege
- `createrole` (optional, default: false): Grant create role privilege

**Example:**
```yaml
- module: db/postgresql
  vars:
    command: create_user
    user_name: appuser
    user_password: secure_password_123
    createdb: true
    createrole: false
    superuser: false
```

**Output:**
```json
{
  "status": "changed",
  "message": "User 'appuser' created successfully",
  "facts": {
    "user_name": "appuser"
  }
}
```

**Idempotency:** Returns `ok` if user already exists.

**Security Note:** Avoid storing passwords in stack files. Use .pgpass or environment variables.

---

### drop_user

Drop an existing PostgreSQL user.

**Variables:**
- `user_name` (required): Username to drop

**Example:**
```yaml
- module: db/postgresql
  vars:
    command: drop_user
    user_name: old_user
```

**Output:**
```json
{
  "status": "changed",
  "message": "User 'old_user' dropped successfully"
}
```

**Idempotency:** Returns `ok` if user doesn't exist.

**Note:** User must not own any databases or have active connections.

---

### list_users

List all users/roles on the PostgreSQL server.

**Variables:** None (uses connection settings)

**Example:**
```yaml
- module: db/postgresql
  vars:
    command: list_users
```

**Output:**
```json
{
  "status": "ok",
  "message": "Users retrieved successfully",
  "facts": {
    "users": "username|superuser|createdb|createrole\npostgres|t|t|t\nappuser|f|t|f"
  }
}
```

---

### alter_user

Alter an existing user's password.

**Variables:**
- `user_name` (required): Username to alter
- `user_password` (required): New password

**Example:**
```yaml
- module: db/postgresql
  vars:
    command: alter_user
    user_name: appuser
    user_password: new_secure_password
```

**Output:**
```json
{
  "status": "changed",
  "message": "User 'appuser' altered successfully"
}
```

**Security Note:** Always use secure password storage methods.

---

### create_role

Create a new PostgreSQL role (group).

**Variables:**
- `role_name` (required): Role name to create

**Example:**
```yaml
- module: db/postgresql
  vars:
    command: create_role
    role_name: app_readers
```

**Output:**
```json
{
  "status": "changed",
  "message": "Role 'app_readers' created successfully"
}
```

**Idempotency:** Returns `ok` if role already exists.

---

### drop_role

Drop an existing PostgreSQL role.

**Variables:**
- `role_name` (required): Role name to drop

**Example:**
```yaml
- module: db/postgresql
  vars:
    command: drop_role
    role_name: old_role
```

**Output:**
```json
{
  "status": "changed",
  "message": "Role 'old_role' dropped successfully"
}
```

**Idempotency:** Returns `ok` if role doesn't exist.

---

### grant_privileges

Grant database privileges to a user.

**Variables:**
- `database_name` (required): Database name
- `user_name` (required): Username to grant privileges to
- `privileges` (optional, default: ALL): Privileges to grant (ALL, SELECT, INSERT, UPDATE, DELETE, etc.)

**Example:**
```yaml
- module: db/postgresql
  vars:
    command: grant_privileges
    database_name: myapp_db
    user_name: appuser
    privileges: ALL
```

**Output:**
```json
{
  "status": "changed",
  "message": "Granted ALL privileges on 'myapp_db' to 'appuser'"
}
```

**Note:** For table-level privileges, use execute_query instead.

---

### revoke_privileges

Revoke database privileges from a user.

**Variables:**
- `database_name` (required): Database name
- `user_name` (required): Username to revoke privileges from
- `privileges` (optional, default: ALL): Privileges to revoke

**Example:**
```yaml
- module: db/postgresql
  vars:
    command: revoke_privileges
    database_name: myapp_db
    user_name: appuser
    privileges: INSERT,UPDATE,DELETE
```

**Output:**
```json
{
  "status": "changed",
  "message": "Revoked INSERT,UPDATE,DELETE privileges on 'myapp_db' from 'appuser'"
}
```

---

## Schema Management

### create_schema

Create a new schema in a database.

**Variables:**
- `postgresql_database` (required): Target database
- `schema_name` (required): Schema name to create
- `owner` (optional): Schema owner

**Example:**
```yaml
- module: db/postgresql
  vars:
    command: create_schema
    postgresql_database: myapp_db
    schema_name: app
    owner: appuser
```

**Output:**
```json
{
  "status": "changed",
  "message": "Schema 'app' created successfully"
}
```

**Idempotency:** Uses `CREATE SCHEMA IF NOT EXISTS`.

---

### drop_schema

Drop a schema from a database.

**Variables:**
- `postgresql_database` (required): Target database
- `schema_name` (required): Schema name to drop
- `cascade` (optional, default: false): Drop objects in schema

**Example:**
```yaml
- module: db/postgresql
  vars:
    command: drop_schema
    postgresql_database: myapp_db
    schema_name: old_schema
    cascade: true
```

**Output:**
```json
{
  "status": "changed",
  "message": "Schema 'old_schema' dropped successfully"
}
```

**Warning:** Using `cascade: true` will drop all objects in the schema.

---

### list_schemas

List all user-created schemas in a database.

**Variables:**
- `postgresql_database` (required): Target database

**Example:**
```yaml
- module: db/postgresql
  vars:
    command: list_schemas
    postgresql_database: myapp_db
```

**Output:**
```json
{
  "status": "ok",
  "message": "Schemas retrieved successfully",
  "facts": {
    "schemas": "schema_name\npublic\napp\narchive"
  }
}
```

---

### set_search_path

Set the default schema search path for a database.

**Variables:**
- `postgresql_database` (required): Target database
- `search_path` (required): Comma-separated schema list

**Example:**
```yaml
- module: db/postgresql
  vars:
    command: set_search_path
    postgresql_database: myapp_db
    search_path: app,public
```

**Output:**
```json
{
  "status": "changed",
  "message": "Search path set to 'app,public' for database 'myapp_db'"
}
```

**Note:** Affects all future connections to the database.

---

## Table Management

### list_tables

List all tables in a schema.

**Variables:**
- `postgresql_database` (required): Target database
- `schema_name` (optional, default: public): Schema to list tables from

**Example:**
```yaml
- module: db/postgresql
  vars:
    command: list_tables
    postgresql_database: myapp_db
    schema_name: app
```

**Output:**
```json
{
  "status": "ok",
  "message": "Tables in schema 'app' retrieved successfully",
  "facts": {
    "tables": "tablename\nusers\nproducts\norders"
  }
}
```

---

### describe_table

Show detailed information about a table structure.

**Variables:**
- `postgresql_database` (required): Target database
- `table_name` (required): Table to describe

**Example:**
```yaml
- module: db/postgresql
  vars:
    command: describe_table
    postgresql_database: myapp_db
    table_name: users
```

**Output:**
```json
{
  "status": "ok",
  "message": "Table 'users' description retrieved",
  "facts": {
    "table_description": "Column | Type | Modifiers\nid | integer | not null\nname | varchar(255) | \n..."
  }
}
```

---

### vacuum_table

Vacuum a table or entire database to reclaim storage.

**Variables:**
- `postgresql_database` (required): Target database
- `table_name` (optional): Specific table (omit for full database)
- `full` (optional, default: false): Perform full vacuum
- `analyze` (optional, default: false): Update statistics after vacuum

**Example:**
```yaml
- module: db/postgresql
  vars:
    command: vacuum_table
    postgresql_database: myapp_db
    table_name: users
    full: false
    analyze: true
```

**Output:**
```json
{
  "status": "changed",
  "message": "Table 'users' vacuumed successfully"
}
```

**Note:** Full vacuum requires exclusive lock and may take significant time.

---

### analyze_table

Update optimizer statistics for a table or database.

**Variables:**
- `postgresql_database` (required): Target database
- `table_name` (optional): Specific table (omit for full database)

**Example:**
```yaml
- module: db/postgresql
  vars:
    command: analyze_table
    postgresql_database: myapp_db
    table_name: users
```

**Output:**
```json
{
  "status": "changed",
  "message": "Table 'users' analyzed successfully"
}
```

**Tip:** Run regularly for optimal query performance.

---

## Extension Management

### create_extension

Create/enable a PostgreSQL extension.

**Variables:**
- `postgresql_database` (required): Target database
- `extension_name` (required): Extension name to enable

**Popular Extensions:**
- `uuid-ossp` - UUID generation functions
- `pg_trgm` - Trigram matching for similarity searches
- `postgis` - Geographic information system support
- `hstore` - Key-value store
- `pgcrypto` - Cryptographic functions
- `citext` - Case-insensitive text type

**Example:**
```yaml
- module: db/postgresql
  vars:
    command: create_extension
    postgresql_database: myapp_db
    extension_name: uuid-ossp
```

**Output:**
```json
{
  "status": "changed",
  "message": "Extension 'uuid-ossp' created successfully"
}
```

**Idempotency:** Uses `CREATE EXTENSION IF NOT EXISTS`.

---

### drop_extension

Drop/disable a PostgreSQL extension.

**Variables:**
- `postgresql_database` (required): Target database
- `extension_name` (required): Extension name to disable
- `cascade` (optional, default: false): Drop dependent objects

**Example:**
```yaml
- module: db/postgresql
  vars:
    command: drop_extension
    postgresql_database: myapp_db
    extension_name: old_extension
    cascade: true
```

**Output:**
```json
{
  "status": "changed",
  "message": "Extension 'old_extension' dropped successfully"
}
```

---

### list_extensions

List all installed extensions in a database.

**Variables:**
- `postgresql_database` (required): Target database

**Example:**
```yaml
- module: db/postgresql
  vars:
    command: list_extensions
    postgresql_database: myapp_db
```

**Output:**
```json
{
  "status": "ok",
  "message": "Extensions retrieved successfully",
  "facts": {
    "extensions": "extname|extversion\nuuid-ossp|1.1\npg_trgm|1.6"
  }
}
```

---

## Query Execution

### execute_query

Execute a single SQL query.

**Variables:**
- `postgresql_database` (required): Target database
- `query` (required): SQL query to execute

**Example:**
```yaml
- module: db/postgresql
  vars:
    command: execute_query
    postgresql_database: myapp_db
    query: "SELECT COUNT(*) FROM users WHERE active = true"
```

**Output:**
```json
{
  "status": "ok",
  "message": "Query executed successfully",
  "facts": {
    "result": "42"
  }
}
```

**Warning:** Be cautious with DML statements (INSERT, UPDATE, DELETE).

---

### execute_script

Execute a SQL script file.

**Variables:**
- `postgresql_database` (required): Target database
- `script_file` (required): Path to SQL script file

**Example:**
```yaml
- module: db/postgresql
  vars:
    command: execute_script
    postgresql_database: myapp_db
    script_file: /tmp/schema_migration.sql
```

**Output:**
```json
{
  "status": "changed",
  "message": "Script '/tmp/schema_migration.sql' executed successfully",
  "facts": {
    "result": "CREATE TABLE\nINSERT 0 100"
  }
}
```

**Note:** Script file must be accessible on the target host.

---

### show_activity

Show active database connections and queries.

**Variables:**
- `postgresql_database` (required): Target database

**Example:**
```yaml
- module: db/postgresql
  vars:
    command: show_activity
    postgresql_database: myapp_db
```

**Output:**
```json
{
  "status": "ok",
  "message": "Database activity retrieved successfully",
  "facts": {
    "activity": "pid|usename|application_name|client_addr|state|query_start|query\n1234|appuser|myapp|192.168.1.100|active|2024-01-15 10:30:00|SELECT * FROM users"
  }
}
```

**Use Case:** Monitoring, troubleshooting, finding long-running queries.

---

## Backup and Restore

### backup_database

Backup a PostgreSQL database to a file.

**Variables:**
- `database_name` (required): Database to backup
- `backup_file` (required): Output file path
- `format` (optional, default: plain): Backup format

**Backup Formats:**
- `plain` - Plain SQL text (human-readable, large)
- `custom` - Custom compressed format (recommended)
- `directory` - Directory with per-table files
- `tar` - TAR archive format

**Example (Plain SQL):**
```yaml
- module: db/postgresql
  vars:
    command: backup_database
    database_name: myapp_db
    backup_file: /backups/myapp_db_20240115.sql
    format: plain
```

**Example (Custom Compressed):**
```yaml
- module: db/postgresql
  vars:
    command: backup_database
    database_name: myapp_db
    backup_file: /backups/myapp_db_20240115.dump
    format: custom
```

**Output:**
```json
{
  "status": "changed",
  "message": "Database 'myapp_db' backed up to '/backups/myapp_db_20240115.dump' (format: custom)",
  "facts": {
    "backup_file": "/backups/myapp_db_20240115.dump",
    "format": "custom"
  }
}
```

**Best Practices:**
- Use `custom` format for compression and flexibility
- Include timestamp in filename
- Store backups on separate storage
- Verify backup integrity after creation

---

### restore_database

Restore a PostgreSQL database from a backup file.

**Variables:**
- `database_name` (required): Target database (must exist)
- `backup_file` (required): Backup file path

**Example:**
```yaml
- module: db/postgresql
  vars:
    command: restore_database
    database_name: myapp_db_restored
    backup_file: /backups/myapp_db_20240115.dump
```

**Output:**
```json
{
  "status": "changed",
  "message": "Database 'myapp_db_restored' restored from '/backups/myapp_db_20240115.dump'",
  "facts": {
    "backup_file": "/backups/myapp_db_20240115.dump"
  }
}
```

**Notes:**
- Target database must be created first (use `create_database`)
- Custom format backups use `pg_restore`
- Plain SQL backups use `psql -f`
- Format is automatically detected from file extension

**Typical Restore Workflow:**
```yaml
run:
  # 1. Create empty database
  - module: db/postgresql
    vars:
      command: create_database
      database_name: myapp_db_restored
      owner: appuser

  # 2. Restore from backup
  - module: db/postgresql
    vars:
      command: restore_database
      database_name: myapp_db_restored
      backup_file: /backups/myapp_db.dump
```

---

## Exit Codes and Status

All commands return one of three status values:

| Status | Meaning | When Used |
|--------|---------|-----------|
| `ok` | Success, no changes | Resource already in desired state |
| `changed` | Success, changes made | Resource was created/modified/deleted |
| `failed` | Operation failed | Error occurred, see message field |

## Common Patterns

### Check-Create Pattern
```yaml
# Check if database exists before creating
- module: db/postgresql
  vars:
    command: check_database_exists
    database_name: myapp_db
  register: db_check

- module: db/postgresql
  vars:
    command: create_database
    database_name: myapp_db
  when: not db_check.facts.exists
```

### Backup-Before-Modify Pattern
```yaml
# Backup before making changes
- module: db/postgresql
  vars:
    command: backup_database
    database_name: myapp_db
    backup_file: /backups/myapp_db_before_migration.dump
    format: custom

# Make changes...
- module: db/postgresql
  vars:
    command: execute_script
    postgresql_database: myapp_db
    script_file: /tmp/migration.sql
```

### Full Setup Pattern
```yaml
# Complete database setup
run:
  - module: db/postgresql
    vars:
      command: create_database
      database_name: myapp_db
      encoding: UTF8

  - module: db/postgresql
    vars:
      command: create_user
      user_name: appuser
      user_password: "{{ vault_db_password }}"

  - module: db/postgresql
    vars:
      command: grant_privileges
      database_name: myapp_db
      user_name: appuser
      privileges: ALL

  - module: db/postgresql
    vars:
      command: create_extension
      postgresql_database: myapp_db
      extension_name: uuid-ossp
```

## See Also

- [README.md](README.md) - Module overview and quick start
- [WORKFLOWS.md](WORKFLOWS.md) - Common usage workflows
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
