package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
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

	dest, _ := input.Vars["dest"].(string)
	if dest == "" {
		outputError("Missing required variable: dest")
		return
	}

	format, _ := input.Vars["format"].(string)
	if format == "" {
		format = "gz"
	}

	remove := false
	if val, ok := input.Vars["remove"].(bool); ok {
		remove = val
	}

	// Build shell commands for archiving
	result := buildArchiveCommands(path, dest, format, remove)
	printOutput(result)
}

func buildArchiveCommands(path, dest, format string, remove bool) TaskOutput {
	var shellExec []map[string]interface{}

	// Validate format
	validFormats := map[string]bool{"gz": true, "bz2": true, "xz": true, "zip": true}
	if !validFormats[format] {
		return TaskOutput{
			Status:  "failed",
			Message: fmt.Sprintf("Invalid format: %s (must be gz, bz2, xz, or zip)", format),
			Facts:   make(map[string]interface{}),
		}
	}

	// Check if source path exists
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": fmt.Sprintf("[ -e '%s' ] || { echo 'Source path not found: %s' >&2; exit 1; }", path, path),
	})

	// Create destination directory if needed
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": fmt.Sprintf("mkdir -p \"$(dirname '%s')\"", dest),
	})

	// Build archive command based on format
	var archiveCmd string
	switch format {
	case "gz":
		archiveCmd = fmt.Sprintf("tar -czf '%s' '%s'", dest, path)
	case "bz2":
		archiveCmd = fmt.Sprintf("tar -cjf '%s' '%s'", dest, path)
	case "xz":
		archiveCmd = fmt.Sprintf("tar -cJf '%s' '%s'", dest, path)
	case "zip":
		// For zip, we need to handle directories differently
		archiveCmd = fmt.Sprintf("if [ -d '%s' ]; then cd \"$(dirname '%s')\" && zip -r '%s' \"$(basename '%s')\"; else zip '%s' '%s'; fi",
			path, path, dest, path, dest, path)
	}

	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": archiveCmd,
	})

	// Verify archive was created
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": fmt.Sprintf("[ -f '%s' ] && ls -lh '%s'", dest, dest),
	})

	// Remove original files if requested
	if remove {
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("rm -rf '%s'", path),
		})
	}

	return TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Creating %s archive: %s", format, dest),
		Facts: map[string]interface{}{
			"shell_exec":     shellExec,
			"archive_format": format,
			"archive_dest":   dest,
		},
	}
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
