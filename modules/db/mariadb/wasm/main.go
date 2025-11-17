package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Input represents the module input structure
type Input struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	User         string `json:"user"`
	Password     string `json:"password"`
	Socket       string `json:"socket"`
	Command      string `json:"command"`
	DatabaseName string `json:"database_name"`
	UserName     string `json:"user_name"`
	UserPassword string `json:"user_password"`
	UserHost     string `json:"user_host"`
	Privileges   string `json:"privileges"`
	TableName    string `json:"table_name"`
	Query        string `json:"query"`
	ScriptFile   string `json:"script_file"`
	BackupFile   string `json:"backup_file"`
	Charset      string `json:"charset"`
	Collation    string `json:"collation"`
}

// Output represents the module output structure
type Output struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message"`
	Facts   map[string]interface{} `json:"facts"`
}

func main() {
	var input Input
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		outputError(fmt.Sprintf("Failed to decode input: %v", err))
		return
	}

	// Validate required fields
	if input.User == "" {
		outputError("mariadb_user is required")
		return
	}

	if input.Command == "" {
		outputError("command is required")
		return
	}

	// Execute the appropriate command
	var result Output
	var err error

	switch input.Command {
	// Database Management
	case "create_database":
		result, err = createDatabase(input)
	case "drop_database":
		result, err = dropDatabase(input)
	case "list_databases":
		result, err = listDatabases(input)
	case "show_database_size":
		result, err = showDatabaseSize(input)
	case "check_database_exists":
		result, err = checkDatabaseExists(input)

	// User Management
	case "create_user":
		result, err = createUser(input)
	case "drop_user":
		result, err = dropUser(input)
	case "list_users":
		result, err = listUsers(input)
	case "grant_privileges":
		result, err = grantPrivileges(input)
	case "revoke_privileges":
		result, err = revokePrivileges(input)
	case "show_user_grants":
		result, err = showUserGrants(input)
	case "change_password":
		result, err = changePassword(input)

	// Table Management
	case "list_tables":
		result, err = listTables(input)
	case "describe_table":
		result, err = describeTable(input)
	case "optimize_table":
		result, err = optimizeTable(input)
	case "repair_table":
		result, err = repairTable(input)

	// Query Execution
	case "execute_query":
		result, err = executeQuery(input)
	case "execute_script":
		result, err = executeScript(input)
	case "show_processlist":
		result, err = showProcesslist(input)

	// Replication
	case "show_master_status":
		result, err = showMasterStatus(input)
	case "show_slave_status":
		result, err = showSlaveStatus(input)
	case "start_slave":
		result, err = startSlave(input)
	case "stop_slave":
		result, err = stopSlave(input)

	// Backup and Restore
	case "backup_database":
		result, err = backupDatabase(input)
	case "restore_database":
		result, err = restoreDatabase(input)

	default:
		outputError(fmt.Sprintf("Unknown command: %s", input.Command))
		return
	}

	if err != nil {
		outputError(err.Error())
		return
	}

	outputResult(result)
}

// Database Management Functions

func createDatabase(input Input) (Output, error) {
	if input.DatabaseName == "" {
		return Output{}, fmt.Errorf("database_name is required")
	}

	// Check if database already exists
	exists, err := databaseExists(input, input.DatabaseName)
	if err != nil {
		return Output{}, err
	}

	if exists {
		return Output{
			Status:  "ok",
			Message: fmt.Sprintf("Database '%s' already exists", input.DatabaseName),
			Facts:   map[string]interface{}{"database": input.DatabaseName, "created": false},
		}, nil
	}

	query := fmt.Sprintf("CREATE DATABASE `%s` CHARACTER SET %s COLLATE %s",
		escapeName(input.DatabaseName), input.Charset, input.Collation)

	if err := executeMariaDBCommand(input, query); err != nil {
		return Output{}, err
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Database '%s' created successfully", input.DatabaseName),
		Facts:   map[string]interface{}{"database": input.DatabaseName, "created": true},
	}, nil
}

func dropDatabase(input Input) (Output, error) {
	if input.DatabaseName == "" {
		return Output{}, fmt.Errorf("database_name is required")
	}

	// Check if database exists
	exists, err := databaseExists(input, input.DatabaseName)
	if err != nil {
		return Output{}, err
	}

	if !exists {
		return Output{
			Status:  "ok",
			Message: fmt.Sprintf("Database '%s' does not exist", input.DatabaseName),
			Facts:   map[string]interface{}{"database": input.DatabaseName, "dropped": false},
		}, nil
	}

	query := fmt.Sprintf("DROP DATABASE `%s`", escapeName(input.DatabaseName))

	if err := executeMariaDBCommand(input, query); err != nil {
		return Output{}, err
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Database '%s' dropped successfully", input.DatabaseName),
		Facts:   map[string]interface{}{"database": input.DatabaseName, "dropped": true},
	}, nil
}

func listDatabases(input Input) (Output, error) {
	output, err := executeMariaDBQuery(input, "SHOW DATABASES")
	if err != nil {
		return Output{}, err
	}

	databases := strings.Split(strings.TrimSpace(output), "\n")
	if len(databases) > 0 {
		// Remove header
		databases = databases[1:]
	}

	return Output{
		Status:  "ok",
		Message: fmt.Sprintf("Found %d databases", len(databases)),
		Facts:   map[string]interface{}{"databases": databases, "count": len(databases)},
	}, nil
}

func showDatabaseSize(input Input) (Output, error) {
	if input.DatabaseName == "" {
		return Output{}, fmt.Errorf("database_name is required")
	}

	query := fmt.Sprintf(`
		SELECT
			table_schema AS 'database',
			ROUND(SUM(data_length + index_length) / 1024 / 1024, 2) AS 'size_mb'
		FROM information_schema.tables
		WHERE table_schema = '%s'
		GROUP BY table_schema
	`, escapeString(input.DatabaseName))

	output, err := executeMariaDBQuery(input, query)
	if err != nil {
		return Output{}, err
	}

	return Output{
		Status:  "ok",
		Message: fmt.Sprintf("Database size for '%s'", input.DatabaseName),
		Facts:   map[string]interface{}{"database": input.DatabaseName, "size_info": output},
	}, nil
}

func checkDatabaseExists(input Input) (Output, error) {
	if input.DatabaseName == "" {
		return Output{}, fmt.Errorf("database_name is required")
	}

	exists, err := databaseExists(input, input.DatabaseName)
	if err != nil {
		return Output{}, err
	}

	return Output{
		Status:  "ok",
		Message: fmt.Sprintf("Database '%s' exists: %v", input.DatabaseName, exists),
		Facts:   map[string]interface{}{"database": input.DatabaseName, "exists": exists},
	}, nil
}

// User Management Functions

func createUser(input Input) (Output, error) {
	if input.UserName == "" {
		return Output{}, fmt.Errorf("user_name is required")
	}
	if input.UserPassword == "" {
		return Output{}, fmt.Errorf("user_password is required")
	}

	// Check if user already exists
	exists, err := userExists(input, input.UserName, input.UserHost)
	if err != nil {
		return Output{}, err
	}

	if exists {
		return Output{
			Status:  "ok",
			Message: fmt.Sprintf("User '%s'@'%s' already exists", input.UserName, input.UserHost),
			Facts:   map[string]interface{}{"user": input.UserName, "host": input.UserHost, "created": false},
		}, nil
	}

	query := fmt.Sprintf("CREATE USER '%s'@'%s' IDENTIFIED BY '%s'",
		escapeString(input.UserName), escapeString(input.UserHost), escapeString(input.UserPassword))

	if err := executeMariaDBCommand(input, query); err != nil {
		return Output{}, err
	}

	// Flush privileges
	if err := executeMariaDBCommand(input, "FLUSH PRIVILEGES"); err != nil {
		return Output{}, err
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("User '%s'@'%s' created successfully", input.UserName, input.UserHost),
		Facts:   map[string]interface{}{"user": input.UserName, "host": input.UserHost, "created": true},
	}, nil
}

func dropUser(input Input) (Output, error) {
	if input.UserName == "" {
		return Output{}, fmt.Errorf("user_name is required")
	}

	// Check if user exists
	exists, err := userExists(input, input.UserName, input.UserHost)
	if err != nil {
		return Output{}, err
	}

	if !exists {
		return Output{
			Status:  "ok",
			Message: fmt.Sprintf("User '%s'@'%s' does not exist", input.UserName, input.UserHost),
			Facts:   map[string]interface{}{"user": input.UserName, "host": input.UserHost, "dropped": false},
		}, nil
	}

	query := fmt.Sprintf("DROP USER '%s'@'%s'",
		escapeString(input.UserName), escapeString(input.UserHost))

	if err := executeMariaDBCommand(input, query); err != nil {
		return Output{}, err
	}

	// Flush privileges
	if err := executeMariaDBCommand(input, "FLUSH PRIVILEGES"); err != nil {
		return Output{}, err
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("User '%s'@'%s' dropped successfully", input.UserName, input.UserHost),
		Facts:   map[string]interface{}{"user": input.UserName, "host": input.UserHost, "dropped": true},
	}, nil
}

func listUsers(input Input) (Output, error) {
	query := "SELECT User, Host FROM mysql.user ORDER BY User, Host"
	output, err := executeMariaDBQuery(input, query)
	if err != nil {
		return Output{}, err
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	userCount := 0
	if len(lines) > 1 {
		userCount = len(lines) - 1 // Exclude header
	}

	return Output{
		Status:  "ok",
		Message: fmt.Sprintf("Found %d users", userCount),
		Facts:   map[string]interface{}{"users": output, "count": userCount},
	}, nil
}

func grantPrivileges(input Input) (Output, error) {
	if input.UserName == "" {
		return Output{}, fmt.Errorf("user_name is required")
	}
	if input.DatabaseName == "" {
		return Output{}, fmt.Errorf("database_name is required")
	}

	query := fmt.Sprintf("GRANT %s ON `%s`.* TO '%s'@'%s'",
		input.Privileges,
		escapeName(input.DatabaseName),
		escapeString(input.UserName),
		escapeString(input.UserHost))

	if err := executeMariaDBCommand(input, query); err != nil {
		return Output{}, err
	}

	// Flush privileges
	if err := executeMariaDBCommand(input, "FLUSH PRIVILEGES"); err != nil {
		return Output{}, err
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Granted %s on '%s' to '%s'@'%s'", input.Privileges, input.DatabaseName, input.UserName, input.UserHost),
		Facts: map[string]interface{}{
			"user":       input.UserName,
			"host":       input.UserHost,
			"database":   input.DatabaseName,
			"privileges": input.Privileges,
		},
	}, nil
}

func revokePrivileges(input Input) (Output, error) {
	if input.UserName == "" {
		return Output{}, fmt.Errorf("user_name is required")
	}
	if input.DatabaseName == "" {
		return Output{}, fmt.Errorf("database_name is required")
	}

	query := fmt.Sprintf("REVOKE %s ON `%s`.* FROM '%s'@'%s'",
		input.Privileges,
		escapeName(input.DatabaseName),
		escapeString(input.UserName),
		escapeString(input.UserHost))

	if err := executeMariaDBCommand(input, query); err != nil {
		return Output{}, err
	}

	// Flush privileges
	if err := executeMariaDBCommand(input, "FLUSH PRIVILEGES"); err != nil {
		return Output{}, err
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Revoked %s on '%s' from '%s'@'%s'", input.Privileges, input.DatabaseName, input.UserName, input.UserHost),
		Facts: map[string]interface{}{
			"user":       input.UserName,
			"host":       input.UserHost,
			"database":   input.DatabaseName,
			"privileges": input.Privileges,
		},
	}, nil
}

func showUserGrants(input Input) (Output, error) {
	if input.UserName == "" {
		return Output{}, fmt.Errorf("user_name is required")
	}

	query := fmt.Sprintf("SHOW GRANTS FOR '%s'@'%s'",
		escapeString(input.UserName), escapeString(input.UserHost))

	output, err := executeMariaDBQuery(input, query)
	if err != nil {
		return Output{}, err
	}

	return Output{
		Status:  "ok",
		Message: fmt.Sprintf("Grants for '%s'@'%s'", input.UserName, input.UserHost),
		Facts:   map[string]interface{}{"user": input.UserName, "host": input.UserHost, "grants": output},
	}, nil
}

func changePassword(input Input) (Output, error) {
	if input.UserName == "" {
		return Output{}, fmt.Errorf("user_name is required")
	}
	if input.UserPassword == "" {
		return Output{}, fmt.Errorf("user_password is required")
	}

	query := fmt.Sprintf("ALTER USER '%s'@'%s' IDENTIFIED BY '%s'",
		escapeString(input.UserName), escapeString(input.UserHost), escapeString(input.UserPassword))

	if err := executeMariaDBCommand(input, query); err != nil {
		return Output{}, err
	}

	// Flush privileges
	if err := executeMariaDBCommand(input, "FLUSH PRIVILEGES"); err != nil {
		return Output{}, err
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Password changed for '%s'@'%s'", input.UserName, input.UserHost),
		Facts:   map[string]interface{}{"user": input.UserName, "host": input.UserHost},
	}, nil
}

// Table Management Functions

func listTables(input Input) (Output, error) {
	if input.DatabaseName == "" {
		return Output{}, fmt.Errorf("database_name is required")
	}

	query := fmt.Sprintf("SHOW TABLES FROM `%s`", escapeName(input.DatabaseName))
	output, err := executeMariaDBQuery(input, query)
	if err != nil {
		return Output{}, err
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	tableCount := 0
	if len(lines) > 1 {
		tableCount = len(lines) - 1 // Exclude header
	}

	return Output{
		Status:  "ok",
		Message: fmt.Sprintf("Found %d tables in database '%s'", tableCount, input.DatabaseName),
		Facts:   map[string]interface{}{"database": input.DatabaseName, "tables": output, "count": tableCount},
	}, nil
}

func describeTable(input Input) (Output, error) {
	if input.DatabaseName == "" {
		return Output{}, fmt.Errorf("database_name is required")
	}
	if input.TableName == "" {
		return Output{}, fmt.Errorf("table_name is required")
	}

	query := fmt.Sprintf("DESCRIBE `%s`.`%s`", escapeName(input.DatabaseName), escapeName(input.TableName))
	output, err := executeMariaDBQuery(input, query)
	if err != nil {
		return Output{}, err
	}

	return Output{
		Status:  "ok",
		Message: fmt.Sprintf("Table structure for '%s.%s'", input.DatabaseName, input.TableName),
		Facts:   map[string]interface{}{"database": input.DatabaseName, "table": input.TableName, "structure": output},
	}, nil
}

func optimizeTable(input Input) (Output, error) {
	if input.DatabaseName == "" {
		return Output{}, fmt.Errorf("database_name is required")
	}
	if input.TableName == "" {
		return Output{}, fmt.Errorf("table_name is required")
	}

	query := fmt.Sprintf("OPTIMIZE TABLE `%s`.`%s`", escapeName(input.DatabaseName), escapeName(input.TableName))
	output, err := executeMariaDBQuery(input, query)
	if err != nil {
		return Output{}, err
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Optimized table '%s.%s'", input.DatabaseName, input.TableName),
		Facts:   map[string]interface{}{"database": input.DatabaseName, "table": input.TableName, "result": output},
	}, nil
}

func repairTable(input Input) (Output, error) {
	if input.DatabaseName == "" {
		return Output{}, fmt.Errorf("database_name is required")
	}
	if input.TableName == "" {
		return Output{}, fmt.Errorf("table_name is required")
	}

	query := fmt.Sprintf("REPAIR TABLE `%s`.`%s`", escapeName(input.DatabaseName), escapeName(input.TableName))
	output, err := executeMariaDBQuery(input, query)
	if err != nil {
		return Output{}, err
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Repaired table '%s.%s'", input.DatabaseName, input.TableName),
		Facts:   map[string]interface{}{"database": input.DatabaseName, "table": input.TableName, "result": output},
	}, nil
}

// Query Execution Functions

func executeQuery(input Input) (Output, error) {
	if input.Query == "" {
		return Output{}, fmt.Errorf("query is required")
	}

	output, err := executeMariaDBQuery(input, input.Query)
	if err != nil {
		return Output{}, err
	}

	status := "ok"
	// Check if query modifies data
	queryUpper := strings.ToUpper(strings.TrimSpace(input.Query))
	if strings.HasPrefix(queryUpper, "INSERT") ||
		strings.HasPrefix(queryUpper, "UPDATE") ||
		strings.HasPrefix(queryUpper, "DELETE") ||
		strings.HasPrefix(queryUpper, "ALTER") ||
		strings.HasPrefix(queryUpper, "CREATE") ||
		strings.HasPrefix(queryUpper, "DROP") {
		status = "changed"
	}

	return Output{
		Status:  status,
		Message: "Query executed successfully",
		Facts:   map[string]interface{}{"query": input.Query, "result": output},
	}, nil
}

func executeScript(input Input) (Output, error) {
	if input.ScriptFile == "" {
		return Output{}, fmt.Errorf("script_file is required")
	}

	// Check if file exists
	if _, err := os.Stat(input.ScriptFile); os.IsNotExist(err) {
		return Output{}, fmt.Errorf("script file not found: %s", input.ScriptFile)
	}

	args := buildMariaDBArgs(input)
	if input.DatabaseName != "" {
		args = append(args, input.DatabaseName)
	}

	cmd := exec.Command("mysql", args...)

	// Read script file
	scriptContent, err := os.ReadFile(input.ScriptFile)
	if err != nil {
		return Output{}, fmt.Errorf("failed to read script file: %v", err)
	}

	cmd.Stdin = strings.NewReader(string(scriptContent))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return Output{}, fmt.Errorf("failed to execute script: %v - %s", err, string(output))
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Script '%s' executed successfully", input.ScriptFile),
		Facts:   map[string]interface{}{"script": input.ScriptFile, "result": string(output)},
	}, nil
}

func showProcesslist(input Input) (Output, error) {
	query := "SHOW FULL PROCESSLIST"
	output, err := executeMariaDBQuery(input, query)
	if err != nil {
		return Output{}, err
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	processCount := 0
	if len(lines) > 1 {
		processCount = len(lines) - 1 // Exclude header
	}

	return Output{
		Status:  "ok",
		Message: fmt.Sprintf("Found %d running processes", processCount),
		Facts:   map[string]interface{}{"processlist": output, "count": processCount},
	}, nil
}

// Replication Functions

func showMasterStatus(input Input) (Output, error) {
	query := "SHOW MASTER STATUS"
	output, err := executeMariaDBQuery(input, query)
	if err != nil {
		return Output{}, err
	}

	return Output{
		Status:  "ok",
		Message: "Master status retrieved",
		Facts:   map[string]interface{}{"master_status": output},
	}, nil
}

func showSlaveStatus(input Input) (Output, error) {
	query := "SHOW SLAVE STATUS\\G"
	output, err := executeMariaDBQuery(input, query)
	if err != nil {
		return Output{}, err
	}

	return Output{
		Status:  "ok",
		Message: "Slave status retrieved",
		Facts:   map[string]interface{}{"slave_status": output},
	}, nil
}

func startSlave(input Input) (Output, error) {
	if err := executeMariaDBCommand(input, "START SLAVE"); err != nil {
		return Output{}, err
	}

	return Output{
		Status:  "changed",
		Message: "Slave started successfully",
		Facts:   map[string]interface{}{},
	}, nil
}

func stopSlave(input Input) (Output, error) {
	if err := executeMariaDBCommand(input, "STOP SLAVE"); err != nil {
		return Output{}, err
	}

	return Output{
		Status:  "changed",
		Message: "Slave stopped successfully",
		Facts:   map[string]interface{}{},
	}, nil
}

// Backup and Restore Functions

func backupDatabase(input Input) (Output, error) {
	if input.DatabaseName == "" {
		return Output{}, fmt.Errorf("database_name is required")
	}
	if input.BackupFile == "" {
		return Output{}, fmt.Errorf("backup_file is required")
	}

	args := []string{
		fmt.Sprintf("--host=%s", input.Host),
		fmt.Sprintf("--port=%d", input.Port),
		fmt.Sprintf("--user=%s", input.User),
	}

	if input.Password != "" {
		args = append(args, fmt.Sprintf("--password=%s", input.Password))
	}

	if input.Socket != "" {
		args = append(args, fmt.Sprintf("--socket=%s", input.Socket))
	}

	args = append(args, "--single-transaction", "--routines", "--triggers", input.DatabaseName)

	cmd := exec.Command("mysqldump", args...)

	output, err := cmd.Output()
	if err != nil {
		return Output{}, fmt.Errorf("backup failed: %v", err)
	}

	// Write to file
	if err := os.WriteFile(input.BackupFile, output, 0644); err != nil {
		return Output{}, fmt.Errorf("failed to write backup file: %v", err)
	}

	fileInfo, _ := os.Stat(input.BackupFile)

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Database '%s' backed up to '%s'", input.DatabaseName, input.BackupFile),
		Facts: map[string]interface{}{
			"database":    input.DatabaseName,
			"backup_file": input.BackupFile,
			"size_bytes":  fileInfo.Size(),
		},
	}, nil
}

func restoreDatabase(input Input) (Output, error) {
	if input.DatabaseName == "" {
		return Output{}, fmt.Errorf("database_name is required")
	}
	if input.BackupFile == "" {
		return Output{}, fmt.Errorf("backup_file is required")
	}

	// Check if backup file exists
	if _, err := os.Stat(input.BackupFile); os.IsNotExist(err) {
		return Output{}, fmt.Errorf("backup file not found: %s", input.BackupFile)
	}

	args := buildMariaDBArgs(input)
	args = append(args, input.DatabaseName)

	cmd := exec.Command("mysql", args...)

	// Read backup file
	backupContent, err := os.ReadFile(input.BackupFile)
	if err != nil {
		return Output{}, fmt.Errorf("failed to read backup file: %v", err)
	}

	cmd.Stdin = strings.NewReader(string(backupContent))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return Output{}, fmt.Errorf("restore failed: %v - %s", err, string(output))
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Database '%s' restored from '%s'", input.DatabaseName, input.BackupFile),
		Facts: map[string]interface{}{
			"database":    input.DatabaseName,
			"backup_file": input.BackupFile,
		},
	}, nil
}

// Helper Functions

func buildMariaDBArgs(input Input) []string {
	args := []string{
		fmt.Sprintf("--host=%s", input.Host),
		fmt.Sprintf("--port=%d", input.Port),
		fmt.Sprintf("--user=%s", input.User),
	}

	if input.Password != "" {
		args = append(args, fmt.Sprintf("--password=%s", input.Password))
	}

	if input.Socket != "" {
		args = append(args, fmt.Sprintf("--socket=%s", input.Socket))
	}

	return args
}

func executeMariaDBCommand(input Input, query string) error {
	args := buildMariaDBArgs(input)
	args = append(args, "-e", query)

	cmd := exec.Command("mysql", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %v - %s", err, string(output))
	}

	return nil
}

func executeMariaDBQuery(input Input, query string) (string, error) {
	args := buildMariaDBArgs(input)
	if input.DatabaseName != "" {
		args = append(args, input.DatabaseName)
	}
	args = append(args, "-e", query)

	cmd := exec.Command("mysql", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("query failed: %v", err)
	}

	return string(output), nil
}

func databaseExists(input Input, dbName string) (bool, error) {
	query := fmt.Sprintf("SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = '%s'",
		escapeString(dbName))

	output, err := executeMariaDBQuery(input, query)
	if err != nil {
		return false, err
	}

	return strings.Contains(output, dbName), nil
}

func userExists(input Input, userName, userHost string) (bool, error) {
	query := fmt.Sprintf("SELECT User FROM mysql.user WHERE User = '%s' AND Host = '%s'",
		escapeString(userName), escapeString(userHost))

	output, err := executeMariaDBQuery(input, query)
	if err != nil {
		return false, err
	}

	return strings.Contains(output, userName), nil
}

func escapeString(s string) string {
	// Basic SQL string escaping
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}

func escapeName(s string) string {
	// Escape backticks in identifiers
	return strings.ReplaceAll(s, "`", "``")
}

func outputResult(result Output) {
	json.NewEncoder(os.Stdout).Encode(result)
}

func outputError(message string) {
	result := Output{
		Status:  "failed",
		Message: message,
		Facts:   map[string]interface{}{},
	}
	json.NewEncoder(os.Stdout).Encode(result)
}
