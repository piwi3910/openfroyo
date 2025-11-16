package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// TaskInput represents the input to this module
type TaskInput struct {
	Vars    map[string]interface{} `json:"vars"`
	Context TaskContext            `json:"context"`
}

// TaskContext provides execution context
type TaskContext struct {
	Host     string `json:"host"`
	TaskName string `json:"task_name"`
}

// TaskOutput represents the output from this module
type TaskOutput struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message"`
	Facts   map[string]interface{} `json:"facts"`
}

func main() {
	// Read input from stdin
	inputBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		outputError(fmt.Sprintf("Failed to read stdin: %v", err))
		return
	}

	var input TaskInput
	if err := json.Unmarshal(inputBytes, &input); err != nil {
		outputError(fmt.Sprintf("Failed to parse input JSON: %v", err))
		return
	}

	// Extract required parameters
	path, _ := input.Vars["path"].(string)
	if path == "" {
		outputError("Missing required variable: path")
		return
	}

	delimiter, _ := input.Vars["delimiter"].(string)
	if delimiter == "" {
		delimiter = ","
	}

	key, _ := input.Vars["key"].(string)
	skipinitialspace, _ := input.Vars["skipinitialspace"].(bool)

	// Build shell commands to read and parse CSV
	result := buildReadCSVCommands(path, delimiter, key, skipinitialspace)
	printOutput(result)
}

func buildReadCSVCommands(path, delimiter, key string, skipinitialspace bool) TaskOutput {
	var shellExec []map[string]interface{}

	// First check if file exists
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": fmt.Sprintf("[ -f '%s' ] || (echo 'CSV_ERROR: File not found' && exit 1)", path),
	})

	// Read the CSV file and output it
	// Use cat to just read the file - let the executor parse it
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": fmt.Sprintf("cat '%s'", path),
	})

	// Store metadata about the CSV parsing request
	metadata := map[string]interface{}{
		"csv_path":           path,
		"csv_delimiter":      delimiter,
		"csv_key":            key,
		"csv_skipinitspace":  skipinitialspace,
	}

	return TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Reading CSV file: %s", path),
		Facts: map[string]interface{}{
			"shell_exec":   shellExec,
			"csv_metadata": metadata,
		},
	}
}

func buildAwkScript(delimiter, key string, skipinitialspace bool) string {
	// Use a simpler approach - just output CSV as is, let the runner parse it
	// This avoids complex shell escaping issues

	// We'll use cat to read the file and return it as-is
	// The executor can parse CSV in Go if needed
	return ""
}

func escapeShellArg(arg string) string {
	// Escape single quotes for shell
	return strings.ReplaceAll(arg, "'", "'\\''")
}

func outputError(msg string) {
	result := TaskOutput{
		Status:  "failed",
		Message: msg,
		Facts:   make(map[string]interface{}),
	}
	printOutput(result)
}

func printOutput(output TaskOutput) {
	outputJSON, _ := json.Marshal(output)
	fmt.Println(string(outputJSON))
}
