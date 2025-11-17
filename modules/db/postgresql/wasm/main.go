package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unsafe"
)

// Input represents the JSON input to the WASM module
type Input struct {
	Vars    map[string]interface{} `json:"vars"`
	Context Context                `json:"context"`
}

// Context provides execution context information
type Context struct {
	Host     string `json:"host"`
	TaskName string `json:"task_name"`
}

// Output represents the JSON output from the WASM module
type Output struct {
	Status  string                 `json:"status"` // ok, changed, failed
	Message string                 `json:"message"`
	Facts   map[string]interface{} `json:"facts,omitempty"`
}

func main() {
	var input Input
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		outputError(fmt.Sprintf("Failed to decode input: %v", err))
		return
	}

	command, ok := input.Vars["command"].(string)
	if !ok {
		outputError("Missing required variable: command")
		return
	}

	var output Output
	var err error

	switch command {
	// Database Management
	case "create_database":
		output, err = createDatabase(input.Vars)
	case "drop_database":
		output, err = dropDatabase(input.Vars)
	case "list_databases":
		output, err = listDatabases(input.Vars)
	case "rename_database":
		output, err = renameDatabase(input.Vars)
	case "show_database_size":
		output, err = showDatabaseSize(input.Vars)
	case "check_database_exists":
		output, err = checkDatabaseExists(input.Vars)

	// User/Role Management
	case "create_user":
		output, err = createUser(input.Vars)
	case "drop_user":
		output, err = dropUser(input.Vars)
	case "list_users":
		output, err = listUsers(input.Vars)
	case "alter_user":
		output, err = alterUser(input.Vars)
	case "create_role":
		output, err = createRole(input.Vars)
	case "drop_role":
		output, err = dropRole(input.Vars)
	case "grant_privileges":
		output, err = grantPrivileges(input.Vars)
	case "revoke_privileges":
		output, err = revokePrivileges(input.Vars)

	// Schema Management
	case "create_schema":
		output, err = createSchema(input.Vars)
	case "drop_schema":
		output, err = dropSchema(input.Vars)
	case "list_schemas":
		output, err = listSchemas(input.Vars)
	case "set_search_path":
		output, err = setSearchPath(input.Vars)

	// Table Management
	case "list_tables":
		output, err = listTables(input.Vars)
	case "describe_table":
		output, err = describeTable(input.Vars)
	case "vacuum_table":
		output, err = vacuumTable(input.Vars)
	case "analyze_table":
		output, err = analyzeTable(input.Vars)

	// Extension Management
	case "create_extension":
		output, err = createExtension(input.Vars)
	case "drop_extension":
		output, err = dropExtension(input.Vars)
	case "list_extensions":
		output, err = listExtensions(input.Vars)

	// Query Execution
	case "execute_query":
		output, err = executeQuery(input.Vars)
	case "execute_script":
		output, err = executeScript(input.Vars)
	case "show_activity":
		output, err = showActivity(input.Vars)

	// Backup and Restore
	case "backup_database":
		output, err = backupDatabase(input.Vars)
	case "restore_database":
		output, err = restoreDatabase(input.Vars)

	default:
		outputError(fmt.Sprintf("Unknown command: %s", command))
		return
	}

	if err != nil {
		outputError(err.Error())
		return
	}

	outputResult(output)
}

// Database Management Functions

func createDatabase(vars map[string]interface{}) (Output, error) {
	dbName := getStringVar(vars, "database_name", "")
	if dbName == "" {
		return Output{}, fmt.Errorf("database_name is required")
	}

	owner := getStringVar(vars, "owner", "")
	encoding := getStringVar(vars, "encoding", "UTF8")

	var query string
	if owner != "" {
		query = fmt.Sprintf("CREATE DATABASE %s OWNER %s ENCODING '%s'", dbName, owner, encoding)
	} else {
		query = fmt.Sprintf("CREATE DATABASE %s ENCODING '%s'", dbName, encoding)
	}

	_, stderr, exitCode := execPSQL(vars, "postgres", query)

	if exitCode != 0 {
		if strings.Contains(stderr, "already exists") {
			return Output{
				Status:  "ok",
				Message: fmt.Sprintf("Database '%s' already exists", dbName),
			}, nil
		}
		return Output{}, fmt.Errorf("failed to create database: %s", stderr)
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Database '%s' created successfully", dbName),
		Facts: map[string]interface{}{
			"database_name": dbName,
			"owner":         owner,
			"encoding":      encoding,
		},
	}, nil
}

func dropDatabase(vars map[string]interface{}) (Output, error) {
	dbName := getStringVar(vars, "database_name", "")
	if dbName == "" {
		return Output{}, fmt.Errorf("database_name is required")
	}

	query := fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName)
	stdout, stderr, exitCode := execPSQL(vars, "postgres", query)

	if exitCode != 0 {
		return Output{}, fmt.Errorf("failed to drop database: %s", stderr)
	}

	changed := !strings.Contains(stdout, "does not exist")

	return Output{
		Status:  getStatus(changed),
		Message: fmt.Sprintf("Database '%s' dropped successfully", dbName),
	}, nil
}

func listDatabases(vars map[string]interface{}) (Output, error) {
	query := "SELECT datname, pg_catalog.pg_get_userbyid(datdba) as owner, pg_encoding_to_char(encoding) as encoding, pg_size_pretty(pg_database_size(datname)) as size FROM pg_catalog.pg_database WHERE datistemplate = false ORDER BY datname"

	stdout, stderr, exitCode := execPSQL(vars, "postgres", query)

	if exitCode != 0 {
		return Output{}, fmt.Errorf("failed to list databases: %s", stderr)
	}

	return Output{
		Status:  "ok",
		Message: "Databases retrieved successfully",
		Facts: map[string]interface{}{
			"databases": stdout,
		},
	}, nil
}

func renameDatabase(vars map[string]interface{}) (Output, error) {
	oldName := getStringVar(vars, "old_name", "")
	newName := getStringVar(vars, "new_name", "")

	if oldName == "" || newName == "" {
		return Output{}, fmt.Errorf("old_name and new_name are required")
	}

	query := fmt.Sprintf("ALTER DATABASE %s RENAME TO %s", oldName, newName)
	_, stderr, exitCode := execPSQL(vars, "postgres", query)

	if exitCode != 0 {
		return Output{}, fmt.Errorf("failed to rename database: %s", stderr)
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Database renamed from '%s' to '%s'", oldName, newName),
		Facts: map[string]interface{}{
			"old_name": oldName,
			"new_name": newName,
		},
	}, nil
}

func showDatabaseSize(vars map[string]interface{}) (Output, error) {
	dbName := getStringVar(vars, "database_name", "")
	if dbName == "" {
		return Output{}, fmt.Errorf("database_name is required")
	}

	query := fmt.Sprintf("SELECT pg_size_pretty(pg_database_size('%s')) as size", dbName)
	stdout, stderr, exitCode := execPSQL(vars, "postgres", query)

	if exitCode != 0 {
		return Output{}, fmt.Errorf("failed to get database size: %s", stderr)
	}

	return Output{
		Status:  "ok",
		Message: fmt.Sprintf("Database '%s' size retrieved", dbName),
		Facts: map[string]interface{}{
			"database_name": dbName,
			"size":          strings.TrimSpace(stdout),
		},
	}, nil
}

func checkDatabaseExists(vars map[string]interface{}) (Output, error) {
	dbName := getStringVar(vars, "database_name", "")
	if dbName == "" {
		return Output{}, fmt.Errorf("database_name is required")
	}

	query := fmt.Sprintf("SELECT 1 FROM pg_catalog.pg_database WHERE datname = '%s'", dbName)
	stdout, stderr, exitCode := execPSQL(vars, "postgres", query)

	if exitCode != 0 {
		return Output{}, fmt.Errorf("failed to check database existence: %s", stderr)
	}

	exists := strings.Contains(stdout, "1")

	return Output{
		Status:  "ok",
		Message: fmt.Sprintf("Database '%s' existence checked", dbName),
		Facts: map[string]interface{}{
			"database_name": dbName,
			"exists":        exists,
		},
	}, nil
}

// User/Role Management Functions

func createUser(vars map[string]interface{}) (Output, error) {
	userName := getStringVar(vars, "user_name", "")
	if userName == "" {
		return Output{}, fmt.Errorf("user_name is required")
	}

	password := getStringVar(vars, "user_password", "")
	superuser := getBoolVar(vars, "superuser", false)
	createdb := getBoolVar(vars, "createdb", false)
	createrole := getBoolVar(vars, "createrole", false)

	var query strings.Builder
	query.WriteString(fmt.Sprintf("CREATE USER %s", userName))

	if password != "" {
		query.WriteString(fmt.Sprintf(" WITH PASSWORD '%s'", password))
	}

	if superuser {
		query.WriteString(" SUPERUSER")
	}
	if createdb {
		query.WriteString(" CREATEDB")
	}
	if createrole {
		query.WriteString(" CREATEROLE")
	}

	_, stderr, exitCode := execPSQL(vars, "postgres", query.String())

	if exitCode != 0 {
		if strings.Contains(stderr, "already exists") {
			return Output{
				Status:  "ok",
				Message: fmt.Sprintf("User '%s' already exists", userName),
			}, nil
		}
		return Output{}, fmt.Errorf("failed to create user: %s", stderr)
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("User '%s' created successfully", userName),
		Facts: map[string]interface{}{
			"user_name": userName,
		},
	}, nil
}

func dropUser(vars map[string]interface{}) (Output, error) {
	userName := getStringVar(vars, "user_name", "")
	if userName == "" {
		return Output{}, fmt.Errorf("user_name is required")
	}

	query := fmt.Sprintf("DROP USER IF EXISTS %s", userName)
	stdout, stderr, exitCode := execPSQL(vars, "postgres", query)

	if exitCode != 0 {
		return Output{}, fmt.Errorf("failed to drop user: %s", stderr)
	}

	changed := !strings.Contains(stdout, "does not exist")

	return Output{
		Status:  getStatus(changed),
		Message: fmt.Sprintf("User '%s' dropped successfully", userName),
	}, nil
}

func listUsers(vars map[string]interface{}) (Output, error) {
	query := "SELECT usename as username, usesuper as superuser, usecreatedb as createdb, usecreaterole as createrole FROM pg_catalog.pg_user ORDER BY usename"

	stdout, stderr, exitCode := execPSQL(vars, "postgres", query)

	if exitCode != 0 {
		return Output{}, fmt.Errorf("failed to list users: %s", stderr)
	}

	return Output{
		Status:  "ok",
		Message: "Users retrieved successfully",
		Facts: map[string]interface{}{
			"users": stdout,
		},
	}, nil
}

func alterUser(vars map[string]interface{}) (Output, error) {
	userName := getStringVar(vars, "user_name", "")
	if userName == "" {
		return Output{}, fmt.Errorf("user_name is required")
	}

	password := getStringVar(vars, "user_password", "")
	if password == "" {
		return Output{}, fmt.Errorf("user_password is required for alter_user")
	}

	query := fmt.Sprintf("ALTER USER %s WITH PASSWORD '%s'", userName, password)
	_, stderr, exitCode := execPSQL(vars, "postgres", query)

	if exitCode != 0 {
		return Output{}, fmt.Errorf("failed to alter user: %s", stderr)
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("User '%s' altered successfully", userName),
	}, nil
}

func createRole(vars map[string]interface{}) (Output, error) {
	roleName := getStringVar(vars, "role_name", "")
	if roleName == "" {
		return Output{}, fmt.Errorf("role_name is required")
	}

	query := fmt.Sprintf("CREATE ROLE %s", roleName)
	_, stderr, exitCode := execPSQL(vars, "postgres", query)

	if exitCode != 0 {
		if strings.Contains(stderr, "already exists") {
			return Output{
				Status:  "ok",
				Message: fmt.Sprintf("Role '%s' already exists", roleName),
			}, nil
		}
		return Output{}, fmt.Errorf("failed to create role: %s", stderr)
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Role '%s' created successfully", roleName),
	}, nil
}

func dropRole(vars map[string]interface{}) (Output, error) {
	roleName := getStringVar(vars, "role_name", "")
	if roleName == "" {
		return Output{}, fmt.Errorf("role_name is required")
	}

	query := fmt.Sprintf("DROP ROLE IF EXISTS %s", roleName)
	stdout, stderr, exitCode := execPSQL(vars, "postgres", query)

	if exitCode != 0 {
		return Output{}, fmt.Errorf("failed to drop role: %s", stderr)
	}

	changed := !strings.Contains(stdout, "does not exist")

	return Output{
		Status:  getStatus(changed),
		Message: fmt.Sprintf("Role '%s' dropped successfully", roleName),
	}, nil
}

func grantPrivileges(vars map[string]interface{}) (Output, error) {
	dbName := getStringVar(vars, "database_name", "")
	userName := getStringVar(vars, "user_name", "")
	privileges := getStringVar(vars, "privileges", "ALL")

	if dbName == "" || userName == "" {
		return Output{}, fmt.Errorf("database_name and user_name are required")
	}

	query := fmt.Sprintf("GRANT %s PRIVILEGES ON DATABASE %s TO %s", privileges, dbName, userName)
	_, stderr, exitCode := execPSQL(vars, "postgres", query)

	if exitCode != 0 {
		return Output{}, fmt.Errorf("failed to grant privileges: %s", stderr)
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Granted %s privileges on '%s' to '%s'", privileges, dbName, userName),
	}, nil
}

func revokePrivileges(vars map[string]interface{}) (Output, error) {
	dbName := getStringVar(vars, "database_name", "")
	userName := getStringVar(vars, "user_name", "")
	privileges := getStringVar(vars, "privileges", "ALL")

	if dbName == "" || userName == "" {
		return Output{}, fmt.Errorf("database_name and user_name are required")
	}

	query := fmt.Sprintf("REVOKE %s PRIVILEGES ON DATABASE %s FROM %s", privileges, dbName, userName)
	_, stderr, exitCode := execPSQL(vars, "postgres", query)

	if exitCode != 0 {
		return Output{}, fmt.Errorf("failed to revoke privileges: %s", stderr)
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Revoked %s privileges on '%s' from '%s'", privileges, dbName, userName),
	}, nil
}

// Schema Management Functions

func createSchema(vars map[string]interface{}) (Output, error) {
	dbName := getStringVar(vars, "postgresql_database", "")
	schemaName := getStringVar(vars, "schema_name", "")
	if schemaName == "" {
		return Output{}, fmt.Errorf("schema_name is required")
	}

	owner := getStringVar(vars, "owner", "")
	var query string
	if owner != "" {
		query = fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s AUTHORIZATION %s", schemaName, owner)
	} else {
		query = fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName)
	}

	stdout, stderr, exitCode := execPSQL(vars, dbName, query)

	if exitCode != 0 {
		return Output{}, fmt.Errorf("failed to create schema: %s", stderr)
	}

	changed := !strings.Contains(stdout, "already exists")

	return Output{
		Status:  getStatus(changed),
		Message: fmt.Sprintf("Schema '%s' created successfully", schemaName),
	}, nil
}

func dropSchema(vars map[string]interface{}) (Output, error) {
	dbName := getStringVar(vars, "postgresql_database", "")
	schemaName := getStringVar(vars, "schema_name", "")
	if schemaName == "" {
		return Output{}, fmt.Errorf("schema_name is required")
	}

	cascade := getBoolVar(vars, "cascade", false)
	var query string
	if cascade {
		query = fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName)
	} else {
		query = fmt.Sprintf("DROP SCHEMA IF EXISTS %s", schemaName)
	}

	stdout, stderr, exitCode := execPSQL(vars, dbName, query)

	if exitCode != 0 {
		return Output{}, fmt.Errorf("failed to drop schema: %s", stderr)
	}

	changed := !strings.Contains(stdout, "does not exist")

	return Output{
		Status:  getStatus(changed),
		Message: fmt.Sprintf("Schema '%s' dropped successfully", schemaName),
	}, nil
}

func listSchemas(vars map[string]interface{}) (Output, error) {
	dbName := getStringVar(vars, "postgresql_database", "")
	query := "SELECT schema_name FROM information_schema.schemata WHERE schema_name NOT LIKE 'pg_%' AND schema_name != 'information_schema' ORDER BY schema_name"

	stdout, stderr, exitCode := execPSQL(vars, dbName, query)

	if exitCode != 0 {
		return Output{}, fmt.Errorf("failed to list schemas: %s", stderr)
	}

	return Output{
		Status:  "ok",
		Message: "Schemas retrieved successfully",
		Facts: map[string]interface{}{
			"schemas": stdout,
		},
	}, nil
}

func setSearchPath(vars map[string]interface{}) (Output, error) {
	dbName := getStringVar(vars, "postgresql_database", "")
	searchPath := getStringVar(vars, "search_path", "")
	if searchPath == "" {
		return Output{}, fmt.Errorf("search_path is required")
	}

	query := fmt.Sprintf("ALTER DATABASE %s SET search_path TO %s", dbName, searchPath)
	_, stderr, exitCode := execPSQL(vars, "postgres", query)

	if exitCode != 0 {
		return Output{}, fmt.Errorf("failed to set search path: %s", stderr)
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Search path set to '%s' for database '%s'", searchPath, dbName),
	}, nil
}

// Table Management Functions

func listTables(vars map[string]interface{}) (Output, error) {
	dbName := getStringVar(vars, "postgresql_database", "")
	schemaName := getStringVar(vars, "schema_name", "public")

	query := fmt.Sprintf("SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname = '%s' ORDER BY tablename", schemaName)
	stdout, stderr, exitCode := execPSQL(vars, dbName, query)

	if exitCode != 0 {
		return Output{}, fmt.Errorf("failed to list tables: %s", stderr)
	}

	return Output{
		Status:  "ok",
		Message: fmt.Sprintf("Tables in schema '%s' retrieved successfully", schemaName),
		Facts: map[string]interface{}{
			"tables": stdout,
		},
	}, nil
}

func describeTable(vars map[string]interface{}) (Output, error) {
	dbName := getStringVar(vars, "postgresql_database", "")
	tableName := getStringVar(vars, "table_name", "")
	if tableName == "" {
		return Output{}, fmt.Errorf("table_name is required")
	}

	// Use \d+ command for detailed table description
	stdout, stderr, exitCode := execPSQLCommand(vars, dbName, fmt.Sprintf("\\d+ %s", tableName))

	if exitCode != 0 {
		return Output{}, fmt.Errorf("failed to describe table: %s", stderr)
	}

	return Output{
		Status:  "ok",
		Message: fmt.Sprintf("Table '%s' description retrieved", tableName),
		Facts: map[string]interface{}{
			"table_description": stdout,
		},
	}, nil
}

func vacuumTable(vars map[string]interface{}) (Output, error) {
	dbName := getStringVar(vars, "postgresql_database", "")
	tableName := getStringVar(vars, "table_name", "")
	full := getBoolVar(vars, "full", false)
	analyze := getBoolVar(vars, "analyze", false)

	var query strings.Builder
	query.WriteString("VACUUM")

	if full {
		query.WriteString(" FULL")
	}
	if analyze {
		query.WriteString(" ANALYZE")
	}

	if tableName != "" {
		query.WriteString(fmt.Sprintf(" %s", tableName))
	}

	_, stderr, exitCode := execPSQL(vars, dbName, query.String())

	if exitCode != 0 {
		return Output{}, fmt.Errorf("failed to vacuum table: %s", stderr)
	}

	var msg string
	if tableName != "" {
		msg = fmt.Sprintf("Table '%s' vacuumed successfully", tableName)
	} else {
		msg = "Database vacuumed successfully"
	}

	return Output{
		Status:  "changed",
		Message: msg,
	}, nil
}

func analyzeTable(vars map[string]interface{}) (Output, error) {
	dbName := getStringVar(vars, "postgresql_database", "")
	tableName := getStringVar(vars, "table_name", "")

	var query string
	if tableName != "" {
		query = fmt.Sprintf("ANALYZE %s", tableName)
	} else {
		query = "ANALYZE"
	}

	_, stderr, exitCode := execPSQL(vars, dbName, query)

	if exitCode != 0 {
		return Output{}, fmt.Errorf("failed to analyze table: %s", stderr)
	}

	var msg string
	if tableName != "" {
		msg = fmt.Sprintf("Table '%s' analyzed successfully", tableName)
	} else {
		msg = "Database analyzed successfully"
	}

	return Output{
		Status:  "changed",
		Message: msg,
	}, nil
}

// Extension Management Functions

func createExtension(vars map[string]interface{}) (Output, error) {
	dbName := getStringVar(vars, "postgresql_database", "")
	extName := getStringVar(vars, "extension_name", "")
	if extName == "" {
		return Output{}, fmt.Errorf("extension_name is required")
	}

	query := fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS \"%s\"", extName)
	stdout, stderr, exitCode := execPSQL(vars, dbName, query)

	if exitCode != 0 {
		return Output{}, fmt.Errorf("failed to create extension: %s", stderr)
	}

	changed := !strings.Contains(stdout, "already exists")

	return Output{
		Status:  getStatus(changed),
		Message: fmt.Sprintf("Extension '%s' created successfully", extName),
	}, nil
}

func dropExtension(vars map[string]interface{}) (Output, error) {
	dbName := getStringVar(vars, "postgresql_database", "")
	extName := getStringVar(vars, "extension_name", "")
	if extName == "" {
		return Output{}, fmt.Errorf("extension_name is required")
	}

	cascade := getBoolVar(vars, "cascade", false)
	var query string
	if cascade {
		query = fmt.Sprintf("DROP EXTENSION IF EXISTS \"%s\" CASCADE", extName)
	} else {
		query = fmt.Sprintf("DROP EXTENSION IF EXISTS \"%s\"", extName)
	}

	stdout, stderr, exitCode := execPSQL(vars, dbName, query)

	if exitCode != 0 {
		return Output{}, fmt.Errorf("failed to drop extension: %s", stderr)
	}

	changed := !strings.Contains(stdout, "does not exist")

	return Output{
		Status:  getStatus(changed),
		Message: fmt.Sprintf("Extension '%s' dropped successfully", extName),
	}, nil
}

func listExtensions(vars map[string]interface{}) (Output, error) {
	dbName := getStringVar(vars, "postgresql_database", "")
	query := "SELECT extname, extversion FROM pg_extension ORDER BY extname"

	stdout, stderr, exitCode := execPSQL(vars, dbName, query)

	if exitCode != 0 {
		return Output{}, fmt.Errorf("failed to list extensions: %s", stderr)
	}

	return Output{
		Status:  "ok",
		Message: "Extensions retrieved successfully",
		Facts: map[string]interface{}{
			"extensions": stdout,
		},
	}, nil
}

// Query Execution Functions

func executeQuery(vars map[string]interface{}) (Output, error) {
	dbName := getStringVar(vars, "postgresql_database", "")
	query := getStringVar(vars, "query", "")
	if query == "" {
		return Output{}, fmt.Errorf("query is required")
	}

	stdout, stderr, exitCode := execPSQL(vars, dbName, query)

	if exitCode != 0 {
		return Output{}, fmt.Errorf("failed to execute query: %s", stderr)
	}

	return Output{
		Status:  "ok",
		Message: "Query executed successfully",
		Facts: map[string]interface{}{
			"result": stdout,
		},
	}, nil
}

func executeScript(vars map[string]interface{}) (Output, error) {
	dbName := getStringVar(vars, "postgresql_database", "")
	scriptFile := getStringVar(vars, "script_file", "")
	if scriptFile == "" {
		return Output{}, fmt.Errorf("script_file is required")
	}

	stdout, stderr, exitCode := execPSQLFile(vars, dbName, scriptFile)

	if exitCode != 0 {
		return Output{}, fmt.Errorf("failed to execute script: %s", stderr)
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Script '%s' executed successfully", scriptFile),
		Facts: map[string]interface{}{
			"result": stdout,
		},
	}, nil
}

func showActivity(vars map[string]interface{}) (Output, error) {
	dbName := getStringVar(vars, "postgresql_database", "")
	query := "SELECT pid, usename, application_name, client_addr, state, query_start, query FROM pg_stat_activity WHERE state != 'idle' ORDER BY query_start"

	stdout, stderr, exitCode := execPSQL(vars, dbName, query)

	if exitCode != 0 {
		return Output{}, fmt.Errorf("failed to show activity: %s", stderr)
	}

	return Output{
		Status:  "ok",
		Message: "Database activity retrieved successfully",
		Facts: map[string]interface{}{
			"activity": stdout,
		},
	}, nil
}

// Backup and Restore Functions

func backupDatabase(vars map[string]interface{}) (Output, error) {
	host := getStringVar(vars, "postgresql_host", "localhost")
	port := getStringVar(vars, "postgresql_port", "5432")
	user := getStringVar(vars, "postgresql_user", "postgres")
	dbName := getStringVar(vars, "database_name", "")
	backupFile := getStringVar(vars, "backup_file", "")
	format := getStringVar(vars, "format", "plain")

	if dbName == "" || backupFile == "" {
		return Output{}, fmt.Errorf("database_name and backup_file are required")
	}

	// Build pg_dump command
	args := []string{
		"-h", host,
		"-p", port,
		"-U", user,
		"-d", dbName,
		"-f", backupFile,
	}

	// Add format option
	switch format {
	case "custom":
		args = append(args, "-Fc")
	case "directory":
		args = append(args, "-Fd")
	case "tar":
		args = append(args, "-Ft")
	default:
		// plain format (default)
		args = append(args, "-Fp")
	}

	_, stderr, exitCode := hostExec("pg_dump", args, 300)

	if exitCode != 0 {
		return Output{}, fmt.Errorf("failed to backup database: %s", stderr)
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Database '%s' backed up to '%s' (format: %s)", dbName, backupFile, format),
		Facts: map[string]interface{}{
			"backup_file": backupFile,
			"format":      format,
		},
	}, nil
}

func restoreDatabase(vars map[string]interface{}) (Output, error) {
	host := getStringVar(vars, "postgresql_host", "localhost")
	port := getStringVar(vars, "postgresql_port", "5432")
	user := getStringVar(vars, "postgresql_user", "postgres")
	dbName := getStringVar(vars, "database_name", "")
	backupFile := getStringVar(vars, "backup_file", "")

	if dbName == "" || backupFile == "" {
		return Output{}, fmt.Errorf("database_name and backup_file are required")
	}

	// Check if backup file exists
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		return Output{}, fmt.Errorf("backup file does not exist: %s", backupFile)
	}

	// Determine if this is a custom format or plain SQL
	format := "plain"
	if ext := filepath.Ext(backupFile); ext == ".dump" || ext == ".backup" {
		format = "custom"
	}

	var stderr string
	var exitCode int

	if format == "custom" {
		// Use pg_restore for custom format
		args := []string{
			"-h", host,
			"-p", port,
			"-U", user,
			"-d", dbName,
			backupFile,
		}
		_, stderr, exitCode = hostExec("pg_restore", args, 300)
	} else {
		// Use psql for plain SQL
		_, stderr, exitCode = execPSQLFile(vars, dbName, backupFile)
	}

	if exitCode != 0 {
		return Output{}, fmt.Errorf("failed to restore database: %s", stderr)
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Database '%s' restored from '%s'", dbName, backupFile),
		Facts: map[string]interface{}{
			"backup_file": backupFile,
		},
	}, nil
}

// Helper Functions

func execPSQL(vars map[string]interface{}, dbName, query string) (string, string, int) {
	host := getStringVar(vars, "postgresql_host", "localhost")
	port := getStringVar(vars, "postgresql_port", "5432")
	user := getStringVar(vars, "postgresql_user", "postgres")

	if dbName == "" {
		dbName = getStringVar(vars, "postgresql_database", "postgres")
	}

	args := []string{
		"-h", host,
		"-p", port,
		"-U", user,
		"-d", dbName,
		"-t", // tuples only (no headers)
		"-A", // unaligned output
		"-c", query,
	}

	return hostExec("psql", args, 60)
}

func execPSQLCommand(vars map[string]interface{}, dbName, command string) (string, string, int) {
	host := getStringVar(vars, "postgresql_host", "localhost")
	port := getStringVar(vars, "postgresql_port", "5432")
	user := getStringVar(vars, "postgresql_user", "postgres")

	if dbName == "" {
		dbName = getStringVar(vars, "postgresql_database", "postgres")
	}

	args := []string{
		"-h", host,
		"-p", port,
		"-U", user,
		"-d", dbName,
		"-c", command,
	}

	return hostExec("psql", args, 60)
}

func execPSQLFile(vars map[string]interface{}, dbName, filePath string) (string, string, int) {
	host := getStringVar(vars, "postgresql_host", "localhost")
	port := getStringVar(vars, "postgresql_port", "5432")
	user := getStringVar(vars, "postgresql_user", "postgres")

	if dbName == "" {
		dbName = getStringVar(vars, "postgresql_database", "postgres")
	}

	args := []string{
		"-h", host,
		"-p", port,
		"-U", user,
		"-d", dbName,
		"-f", filePath,
	}

	return hostExec("psql", args, 300)
}

func getStringVar(vars map[string]interface{}, key, defaultValue string) string {
	if val, ok := vars[key].(string); ok {
		return val
	}
	return defaultValue
}

func getBoolVar(vars map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := vars[key].(bool); ok {
		return val
	}
	return defaultValue
}

func getStatus(changed bool) string {
	if changed {
		return "changed"
	}
	return "ok"
}

func outputResult(output Output) {
	json.NewEncoder(os.Stdout).Encode(output)
}

func outputError(message string) {
	output := Output{
		Status:  "failed",
		Message: message,
	}
	json.NewEncoder(os.Stdout).Encode(output)
}

// Host API functions (imported from host)
// These are provided by the froyo-runner host environment

//go:wasmimport env host_exec
func host_exec(cmd_ptr, cmd_len, args_ptr, args_len, timeout uint32) uint64

//go:wasmimport env host_log
func host_log(level_ptr, level_len, msg_ptr, msg_len uint32)

func hostExec(cmd string, args []string, timeout int) (stdout, stderr string, exitCode int) {
	// Marshal args to JSON
	argsJSON, _ := json.Marshal(args)

	// Convert strings to pointers
	cmdBytes := []byte(cmd)
	argsBytes := []byte(argsJSON)

	// Call host_exec
	result := host_exec(
		uint32(uintptr(unsafe.Pointer(&cmdBytes[0]))), uint32(len(cmd)),
		uint32(uintptr(unsafe.Pointer(&argsBytes[0]))), uint32(len(argsJSON)),
		uint32(timeout),
	)

	// Parse result (simplified - actual implementation would decode from memory)
	exitCode = int(result & 0xFFFFFFFF)
	// stdout and stderr would be read from shared memory in real implementation
	stdout = ""
	stderr = ""

	return stdout, stderr, exitCode
}

func logMessage(level, message string) {
	levelBytes := []byte(level)
	msgBytes := []byte(message)

	host_log(
		uint32(uintptr(unsafe.Pointer(&levelBytes[0]))), uint32(len(level)),
		uint32(uintptr(unsafe.Pointer(&msgBytes[0]))), uint32(len(message)),
	)
}
