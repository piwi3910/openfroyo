# PostgreSQL Module

Comprehensive PostgreSQL database management module for OpenFroyo. This module provides 30+ commands for managing PostgreSQL databases, users, roles, schemas, tables, extensions, and more.

## Overview

The PostgreSQL module enables agentless database management operations over SSH using WebAssembly (WASM). It uses the `psql` CLI client and `pg_dump`/`pg_restore` utilities for all operations.

## Features

- **30+ PostgreSQL Commands** covering all major operations
- **Agentless Execution** via SSH and WASM
- **Idempotent Operations** for safe repeated execution
- **Comprehensive Error Handling** with detailed output
- **Connection Security** via `.pgpass` file support
- **Flexible Authentication** supporting multiple connection methods

## Requirements

### Target Host Requirements

- PostgreSQL client tools (`psql`, `pg_dump`, `pg_restore`)
- SSH access
- Network connectivity to PostgreSQL server

### PostgreSQL Versions

- PostgreSQL 10+
- PostgreSQL 11+
- PostgreSQL 12+
- PostgreSQL 13+
- PostgreSQL 14+
- PostgreSQL 15+
- PostgreSQL 16+

## Installation

The PostgreSQL module is located at `modules/db/postgresql/` in the OpenFroyo repository.

```bash
# Verify module structure
ls -la modules/db/postgresql/
# module.ofy.yml     # Module definition
# defaults.ofy.yml   # Default variables
# wasm/              # WASM module directory
# README.md          # This file
# COMMANDS.md        # Command reference
# WORKFLOWS.md       # Usage workflows
# test.ofy           # Test stack
```

## Quick Start

### Basic Database Creation

```yaml
# stacks/create_database.ofy
name: Create PostgreSQL Database
inventory: @group:db-servers

vars:
  postgresql_host: localhost
  postgresql_user: postgres
  postgresql_password: secretpassword

run:
  - module: db/postgresql
    vars:
      command: create_database
      database_name: myapp_db
      owner: postgres
      encoding: UTF8
```

### Create User and Grant Privileges

```yaml
# stacks/setup_user.ofy
name: Setup PostgreSQL User
inventory: @group:db-servers

vars:
  postgresql_host: localhost
  postgresql_user: postgres

run:
  - module: db/postgresql
    vars:
      command: create_user
      user_name: appuser
      user_password: apppassword
      createdb: true

  - module: db/postgresql
    vars:
      command: grant_privileges
      database_name: myapp_db
      user_name: appuser
      privileges: ALL
```

## Connection Configuration

### Environment Variables

```bash
# PostgreSQL connection settings
export PGHOST=localhost
export PGPORT=5432
export PGUSER=postgres
export PGPASSWORD=secretpassword
export PGDATABASE=postgres
```

### Using .pgpass File (Recommended)

Create a `.pgpass` file in the user's home directory:

```
# Format: hostname:port:database:username:password
localhost:5432:*:postgres:secretpassword
db.example.com:5432:*:appuser:apppassword
```

Set proper permissions:

```bash
chmod 600 ~/.pgpass
```

### Connection String

```yaml
vars:
  postgresql_host: db.example.com
  postgresql_port: 5432
  postgresql_user: postgres
  postgresql_password: secretpassword
  postgresql_database: postgres  # Connection database
```

## Command Categories

### Database Management (6 commands)

- `create_database` - Create a new database
- `drop_database` - Drop an existing database
- `list_databases` - List all databases
- `rename_database` - Rename a database
- `show_database_size` - Show database size
- `check_database_exists` - Check if database exists

### User/Role Management (8 commands)

- `create_user` - Create a new user
- `drop_user` - Drop an existing user
- `list_users` - List all users
- `alter_user` - Alter user password
- `create_role` - Create a new role
- `drop_role` - Drop an existing role
- `grant_privileges` - Grant privileges to user
- `revoke_privileges` - Revoke privileges from user

### Schema Management (4 commands)

- `create_schema` - Create a new schema
- `drop_schema` - Drop an existing schema
- `list_schemas` - List all schemas
- `set_search_path` - Set database search path

### Table Management (4 commands)

- `list_tables` - List tables in schema
- `describe_table` - Describe table structure
- `vacuum_table` - Vacuum table or database
- `analyze_table` - Analyze table or database

### Extension Management (3 commands)

- `create_extension` - Create/enable extension
- `drop_extension` - Drop/disable extension
- `list_extensions` - List installed extensions

### Query Execution (3 commands)

- `execute_query` - Execute a single query
- `execute_script` - Execute SQL script file
- `show_activity` - Show active connections

### Backup and Restore (2 commands)

- `backup_database` - Backup database to file
- `restore_database` - Restore database from file

## Usage Examples

### Database Operations

```yaml
# Create database
- module: db/postgresql
  vars:
    command: create_database
    database_name: myapp_db
    owner: postgres
    encoding: UTF8

# Check if database exists
- module: db/postgresql
  vars:
    command: check_database_exists
    database_name: myapp_db

# Show database size
- module: db/postgresql
  vars:
    command: show_database_size
    database_name: myapp_db

# Drop database
- module: db/postgresql
  vars:
    command: drop_database
    database_name: old_db
```

### User Management

```yaml
# Create user with permissions
- module: db/postgresql
  vars:
    command: create_user
    user_name: appuser
    user_password: secure_password
    createdb: true
    createrole: false
    superuser: false

# Alter user password
- module: db/postgresql
  vars:
    command: alter_user
    user_name: appuser
    user_password: new_password

# Grant privileges
- module: db/postgresql
  vars:
    command: grant_privileges
    database_name: myapp_db
    user_name: appuser
    privileges: ALL

# Revoke privileges
- module: db/postgresql
  vars:
    command: revoke_privileges
    database_name: myapp_db
    user_name: appuser
    privileges: INSERT,UPDATE,DELETE
```

### Schema Management

```yaml
# Create schema
- module: db/postgresql
  vars:
    command: create_schema
    postgresql_database: myapp_db
    schema_name: app
    owner: appuser

# Set search path
- module: db/postgresql
  vars:
    command: set_search_path
    postgresql_database: myapp_db
    search_path: app,public

# Drop schema (with cascade)
- module: db/postgresql
  vars:
    command: drop_schema
    postgresql_database: myapp_db
    schema_name: old_schema
    cascade: true
```

### Extension Management

```yaml
# Create popular extensions
- module: db/postgresql
  vars:
    command: create_extension
    postgresql_database: myapp_db
    extension_name: uuid-ossp

- module: db/postgresql
  vars:
    command: create_extension
    postgresql_database: myapp_db
    extension_name: pg_trgm

- module: db/postgresql
  vars:
    command: create_extension
    postgresql_database: myapp_db
    extension_name: postgis
```

### Maintenance Operations

```yaml
# Vacuum and analyze specific table
- module: db/postgresql
  vars:
    command: vacuum_table
    postgresql_database: myapp_db
    table_name: users
    full: false
    analyze: true

# Analyze entire database
- module: db/postgresql
  vars:
    command: analyze_table
    postgresql_database: myapp_db
```

### Backup and Restore

```yaml
# Backup database (plain SQL format)
- module: db/postgresql
  vars:
    command: backup_database
    database_name: myapp_db
    backup_file: /tmp/myapp_db.sql
    format: plain

# Backup database (custom format - compressed)
- module: db/postgresql
  vars:
    command: backup_database
    database_name: myapp_db
    backup_file: /tmp/myapp_db.dump
    format: custom

# Restore database
- module: db/postgresql
  vars:
    command: restore_database
    database_name: myapp_db_restored
    backup_file: /tmp/myapp_db.sql
```

## Variable Reference

### Connection Variables

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `postgresql_host` | string | localhost | PostgreSQL host |
| `postgresql_port` | string | 5432 | PostgreSQL port |
| `postgresql_user` | string | postgres | PostgreSQL user |
| `postgresql_password` | string | "" | PostgreSQL password |
| `postgresql_database` | string | postgres | Connection database |

### Command Variable

| Variable | Type | Required | Description |
|----------|------|----------|-------------|
| `command` | string | Yes | Operation to perform |

### Database Variables

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `database_name` | string | "" | Target database name |
| `owner` | string | "" | Database owner |
| `encoding` | string | UTF8 | Database encoding |
| `old_name` | string | "" | Old database name (rename) |
| `new_name` | string | "" | New database name (rename) |

### User/Role Variables

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `user_name` | string | "" | Username |
| `user_password` | string | "" | User password |
| `superuser` | boolean | false | Superuser privilege |
| `createdb` | boolean | false | Create database privilege |
| `createrole` | boolean | false | Create role privilege |
| `role_name` | string | "" | Role name |
| `privileges` | string | ALL | Privileges to grant/revoke |

### Schema Variables

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `schema_name` | string | public | Schema name |
| `search_path` | string | "" | Database search path |
| `cascade` | boolean | false | Cascade drop operation |

### Table Variables

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `table_name` | string | "" | Table name |
| `full` | boolean | false | Full vacuum |
| `analyze` | boolean | false | Analyze after vacuum |

### Extension Variables

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `extension_name` | string | "" | Extension name |

### Query Variables

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `query` | string | "" | SQL query to execute |
| `script_file` | string | "" | SQL script file path |

### Backup Variables

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `backup_file` | string | "" | Backup file path |
| `format` | string | plain | Backup format (plain, custom, directory, tar) |

## Output and Facts

Each command returns structured output with status and facts:

```json
{
  "status": "changed",
  "message": "Database 'myapp_db' created successfully",
  "facts": {
    "database_name": "myapp_db",
    "owner": "postgres",
    "encoding": "UTF8"
  }
}
```

### Status Values

- `ok` - Operation completed, no changes made
- `changed` - Operation completed with changes
- `failed` - Operation failed (see message)

## Security Best Practices

### Password Management

1. **Use .pgpass file** instead of passing passwords in stack files
2. **Set proper file permissions** (`chmod 600 ~/.pgpass`)
3. **Use environment variables** for sensitive data
4. **Rotate passwords regularly** using `alter_user`
5. **Use strong passwords** with mixed case, numbers, and symbols

### Access Control

1. **Principle of least privilege** - Grant minimal required permissions
2. **Use roles** for grouping permissions
3. **Separate admin and app users** - Don't use superuser for applications
4. **Audit user access** regularly with `list_users`
5. **Revoke unused privileges** promptly

### Network Security

1. **Use SSL/TLS connections** when possible
2. **Restrict host access** via pg_hba.conf
3. **Use SSH tunneling** for remote connections
4. **Firewall rules** to limit database access
5. **Monitor connections** with `show_activity`

### Backup Security

1. **Encrypt backups** at rest
2. **Secure backup storage** with proper permissions
3. **Test restore procedures** regularly
4. **Rotate backup files** and cleanup old backups
5. **Verify backup integrity** after creation

## Performance Tuning

### Vacuum and Analyze

Regular maintenance improves query performance:

```yaml
# Weekly full vacuum
- module: db/postgresql
  vars:
    command: vacuum_table
    postgresql_database: myapp_db
    full: true
    analyze: true

# Daily analyze
- module: db/postgresql
  vars:
    command: analyze_table
    postgresql_database: myapp_db
```

### Monitoring Activity

Track active connections and queries:

```yaml
- module: db/postgresql
  vars:
    command: show_activity
    postgresql_database: myapp_db
```

### Database Size Monitoring

Monitor database growth:

```yaml
- module: db/postgresql
  vars:
    command: show_database_size
    database_name: myapp_db
```

## Troubleshooting

### Connection Issues

**Problem:** Cannot connect to PostgreSQL

**Solutions:**
- Verify PostgreSQL is running: `systemctl status postgresql`
- Check firewall rules: `sudo ufw status`
- Verify pg_hba.conf allows connections
- Test connection manually: `psql -h localhost -U postgres`

### Authentication Failures

**Problem:** Authentication failed for user

**Solutions:**
- Verify password in .pgpass file
- Check user exists: `command: list_users`
- Verify pg_hba.conf authentication method
- Check user privileges: `command: list_users`

### Permission Denied

**Problem:** Permission denied for database/table

**Solutions:**
- Verify user privileges: `command: list_users`
- Grant required privileges: `command: grant_privileges`
- Check database owner: `command: list_databases`
- Verify schema permissions

### Backup/Restore Failures

**Problem:** Backup or restore operation fails

**Solutions:**
- Verify disk space: `df -h`
- Check file permissions: `ls -l /path/to/backup`
- Verify PostgreSQL version compatibility
- Check backup format matches restore method

## Module Files

```
modules/db/postgresql/
├── module.ofy.yml          # Module definition
├── defaults.ofy.yml        # Default variables
├── wasm/
│   ├── main.go            # WASM module source
│   ├── Makefile           # Build configuration
│   └── postgresql.wasm    # Compiled WASM module (3.3MB)
├── README.md              # This file
├── COMMANDS.md            # Detailed command reference
├── WORKFLOWS.md           # Common usage workflows
└── test.ofy               # Test stack file
```

## Building the Module

To rebuild the WASM module:

```bash
cd modules/db/postgresql/wasm
make build
```

To clean build artifacts:

```bash
make clean
```

## Testing

Test the module with the provided test stack:

```bash
froyo apply modules/db/postgresql/test.ofy
```

## See Also

- [COMMANDS.md](COMMANDS.md) - Detailed command reference
- [WORKFLOWS.md](WORKFLOWS.md) - Common usage workflows
- [OpenFroyo Documentation](../../docs/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)

## Support

For issues, questions, or contributions:
- GitHub Issues: https://github.com/piwi3910/openfroyo/issues
- OpenFroyo Documentation: ../../docs/

## License

This module is part of the OpenFroyo project.
