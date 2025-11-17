# MariaDB Database Management Module

Comprehensive MariaDB/MySQL database management module for OpenFroyo that provides 25+ commands for database administration, user management, table operations, query execution, replication, and backup/restore operations.

## Features

- **Database Management**: Create, drop, list, and check databases
- **User Management**: Create users, manage privileges, change passwords
- **Table Management**: List, describe, optimize, and repair tables
- **Query Execution**: Execute queries and SQL scripts
- **Replication**: Monitor and control master-slave replication
- **Backup & Restore**: Database backup and restore operations
- **Security**: SQL injection prevention, secure password handling
- **Idempotency**: Safe to run multiple times without side effects

## Requirements

### Target Host Requirements

- MariaDB or MySQL server installed and running
- `mysql` client binary available in PATH
- `mysqldump` binary available for backup operations
- SSH access with appropriate permissions
- User with sufficient database privileges

### Connection Methods

The module supports two connection methods:

1. **TCP/IP Connection** (default):
   - Uses host and port
   - Suitable for remote connections
   - Example: `localhost:3306`

2. **Unix Socket Connection** (optional):
   - Uses local socket file
   - Faster for local connections
   - Example: `/var/run/mysqld/mysqld.sock`

## Installation

The module is located at `modules/db/mariadb/` and includes:

```
modules/db/mariadb/
├── module.ofy.yml      # Module definition
├── defaults.ofy.yml    # Default variables
├── wasm/
│   ├── main.go         # Go source code
│   ├── Makefile        # Build configuration
│   └── mariadb.wasm    # Compiled WASM module
├── test.ofy            # Test stack file
├── README.md           # This file
├── COMMANDS.md         # Command reference
└── WORKFLOWS.md        # Common workflows
```

## Quick Start

### 1. Create a Database

```yaml
- run: db/mariadb
  on: @database-servers
  vars:
    mariadb_user: root
    mariadb_password: secret
    command: create_database
    database_name: myapp_db
```

### 2. Create a User and Grant Privileges

```yaml
- run: db/mariadb
  on: @database-servers
  vars:
    mariadb_user: root
    mariadb_password: secret
    command: create_user
    user_name: appuser
    user_password: apppass123
    user_host: localhost

- run: db/mariadb
  on: @database-servers
  vars:
    mariadb_user: root
    mariadb_password: secret
    command: grant_privileges
    database_name: myapp_db
    user_name: appuser
    user_host: localhost
    privileges: ALL
```

### 3. Backup a Database

```yaml
- run: db/mariadb
  on: @database-servers
  vars:
    mariadb_user: root
    mariadb_password: secret
    command: backup_database
    database_name: myapp_db
    backup_file: /tmp/myapp_db_backup.sql
```

## Configuration Variables

### Connection Variables

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `mariadb_host` | string | `localhost` | Database host |
| `mariadb_port` | int | `3306` | Database port |
| `mariadb_user` | string | `root` | Database user |
| `mariadb_password` | string | (required) | Database password |
| `mariadb_socket` | string | `""` | Unix socket path (optional) |

### Database Variables

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `database_name` | string | `""` | Target database name |
| `charset` | string | `utf8mb4` | Database character set |
| `collation` | string | `utf8mb4_unicode_ci` | Database collation |

### User Variables

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `user_name` | string | `""` | Database user name |
| `user_password` | string | `""` | User password |
| `user_host` | string | `localhost` | User host (use `%` for any) |
| `privileges` | string | `ALL` | Privileges to grant/revoke |

### Query Variables

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `query` | string | `""` | SQL query to execute |
| `script_file` | string | `""` | Path to SQL script file |

### Table Variables

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `table_name` | string | `""` | Target table name |

### Backup Variables

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `backup_file` | string | `""` | Path to backup file |

## Available Commands

The module supports 25+ commands across six categories:

### Database Management (5 commands)
- `create_database` - Create a new database
- `drop_database` - Drop an existing database
- `list_databases` - List all databases
- `show_database_size` - Show database size in MB
- `check_database_exists` - Check if database exists

### User Management (7 commands)
- `create_user` - Create a new user
- `drop_user` - Drop an existing user
- `list_users` - List all users
- `grant_privileges` - Grant privileges to user
- `revoke_privileges` - Revoke privileges from user
- `show_user_grants` - Show user's current grants
- `change_password` - Change user password

### Table Management (4 commands)
- `list_tables` - List all tables in database
- `describe_table` - Show table structure
- `optimize_table` - Optimize table
- `repair_table` - Repair table

### Query Execution (3 commands)
- `execute_query` - Execute a SQL query
- `execute_script` - Execute SQL script from file
- `show_processlist` - Show running processes

### Replication (4 commands)
- `show_master_status` - Show master replication status
- `show_slave_status` - Show slave replication status
- `start_slave` - Start slave replication
- `stop_slave` - Stop slave replication

### Backup and Restore (2 commands)
- `backup_database` - Backup database to file
- `restore_database` - Restore database from file

See [COMMANDS.md](COMMANDS.md) for detailed command documentation.

## Security Best Practices

### 1. Password Management

**DO NOT** hardcode passwords in stack files:

```yaml
# Bad - password in plain text
vars:
  mariadb_password: mysecretpassword
```

**DO** use environment variables or secure vaults:

```yaml
# Good - use environment variable
vars:
  mariadb_password: "{{ env.MARIADB_ROOT_PASSWORD }}"
```

### 2. Principle of Least Privilege

Grant only the minimum required privileges:

```yaml
# Instead of ALL privileges
privileges: ALL

# Grant specific privileges
privileges: SELECT,INSERT,UPDATE,DELETE
```

### 3. Host Restrictions

Restrict user access by host:

```yaml
# Local access only
user_host: localhost

# Specific host
user_host: 192.168.1.100

# Any host (use with caution)
user_host: "%"
```

### 4. Secure Connections

For production environments:

- Use SSL/TLS connections
- Implement firewall rules
- Use SSH tunneling for remote access
- Regular password rotation
- Monitor access logs

### 5. Backup Security

Protect backup files:

```yaml
# Store backups in secure location
backup_file: /secure/backups/db_backup_{{ timestamp }}.sql

# Set restrictive permissions
# (handle in post-backup step)
```

## Performance Tuning Tips

### 1. Connection Optimization

Use Unix sockets for local connections:

```yaml
mariadb_socket: /var/run/mysqld/mysqld.sock
```

### 2. Backup Performance

For large databases, use additional options:

```bash
# In backup operations, the module uses:
mysqldump --single-transaction  # InnoDB consistency
          --routines            # Include stored procedures
          --triggers            # Include triggers
```

### 3. Table Maintenance

Regular maintenance improves performance:

```yaml
# Optimize tables weekly
- run: db/mariadb
  vars:
    command: optimize_table
    database_name: myapp_db
    table_name: large_table
```

### 4. Query Optimization

Monitor slow queries:

```yaml
# Check running processes
- run: db/mariadb
  vars:
    command: show_processlist
```

### 5. Replication Monitoring

Regular replication health checks:

```yaml
# Check slave status
- run: db/mariadb
  on: @slave-servers
  vars:
    command: show_slave_status
```

## Troubleshooting

### Connection Issues

**Problem**: Cannot connect to database

```
Error: command failed: ERROR 2002 (HY000): Can't connect to local MySQL server
```

**Solutions**:
1. Verify MariaDB is running: `systemctl status mariadb`
2. Check connection credentials
3. Verify firewall allows port 3306
4. Try socket connection instead of TCP

### Authentication Issues

**Problem**: Access denied

```
Error: ERROR 1045 (28000): Access denied for user 'root'@'localhost'
```

**Solutions**:
1. Verify password is correct
2. Check user exists and has proper host
3. Ensure privileges are granted
4. Try resetting root password

### Permission Issues

**Problem**: Insufficient privileges

```
Error: ERROR 1142 (42000): CREATE command denied
```

**Solutions**:
1. Grant required privileges
2. Use a user with sufficient permissions
3. Check privilege scope (database, table, global)

### Backup/Restore Issues

**Problem**: mysqldump not found

```
Error: exec: "mysqldump": executable file not found in $PATH
```

**Solutions**:
1. Install MariaDB client tools
2. Verify mysqldump is in PATH
3. Use full path to mysqldump

### Replication Issues

**Problem**: Slave not replicating

```
Slave_IO_Running: No
Slave_SQL_Running: No
```

**Solutions**:
1. Check master status
2. Verify replication user credentials
3. Check network connectivity
4. Review error log
5. Use `start_slave` command

## Examples

See [WORKFLOWS.md](WORKFLOWS.md) for complete workflow examples including:

- Setting up a new application database
- Configuring master-slave replication
- Automated backup strategies
- Database migration workflows
- User privilege management

## Module Output

All commands return structured JSON output:

```json
{
  "status": "ok|changed|failed",
  "message": "Descriptive message",
  "facts": {
    "key": "value"
  }
}
```

### Status Values

- **ok**: Operation completed, no changes made
- **changed**: Operation completed, changes made
- **failed**: Operation failed with error

### Facts

Commands return relevant facts that can be used in subsequent tasks:

```yaml
- run: db/mariadb
  vars:
    command: list_databases
  # Returns: facts.databases, facts.count
```

## Building from Source

To rebuild the WASM module:

```bash
cd modules/db/mariadb/wasm
make build
```

Requirements:
- Go 1.21 or later
- WASM support (GOOS=wasip1 GOARCH=wasm)

## Testing

Run the test stack:

```bash
froyo apply modules/db/mariadb/test.ofy
```

## Support

For issues, questions, or contributions:

- GitHub Issues: https://github.com/piwi3910/openfroyo/issues
- Documentation: See COMMANDS.md and WORKFLOWS.md
- Examples: See test.ofy

## License

Part of the OpenFroyo project.
