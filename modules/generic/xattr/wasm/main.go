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

	key, _ := input.Vars["key"].(string)
	if key == "" {
		outputError("Missing required variable: key")
		return
	}

	state, _ := input.Vars["state"].(string)
	if state == "" {
		state = "present"
	}

	value, _ := input.Vars["value"].(string)

	// Build shell commands based on state
	result := buildXattrCommands(path, key, value, state)
	printOutput(result)
}

func buildXattrCommands(path, key, value, state string) TaskOutput {
	var shellExec []map[string]interface{}

	// Detect OS type (we'll use uname to determine Linux vs macOS)
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": "uname -s",
	})

	switch state {
	case "present":
		// Set extended attribute
		if value == "" {
			return TaskOutput{
				Status:  "failed",
				Message: "value parameter required for state=present",
				Facts:   make(map[string]interface{}),
			}
		}

		// Try Linux-style setfattr first, fallback to macOS xattr
		// Linux: setfattr -n key -v value path
		// macOS: xattr -w key value path
		shellExec = append(shellExec, map[string]interface{}{
			"type": "shell",
			"command": fmt.Sprintf(
				"if command -v setfattr >/dev/null 2>&1; then setfattr -n '%s' -v '%s' '%s'; else xattr -w '%s' '%s' '%s'; fi",
				key, value, path, key, value, path,
			),
		})

	case "absent":
		// Remove extended attribute
		// Linux: setfattr -x key path
		// macOS: xattr -d key path
		shellExec = append(shellExec, map[string]interface{}{
			"type": "shell",
			"command": fmt.Sprintf(
				"if command -v setfattr >/dev/null 2>&1; then setfattr -x '%s' '%s'; else xattr -d '%s' '%s'; fi",
				key, path, key, path,
			),
		})

	default:
		return TaskOutput{
			Status:  "failed",
			Message: fmt.Sprintf("Invalid state: %s (must be present or absent)", state),
			Facts:   make(map[string]interface{}),
		}
	}

	// Add verification command to show the specific attribute
	// Linux: getfattr -n key path
	// macOS: xattr -p key path
	shellExec = append(shellExec, map[string]interface{}{
		"type": "shell",
		"command": fmt.Sprintf(
			"if command -v getfattr >/dev/null 2>&1; then getfattr -n '%s' '%s' 2>/dev/null || echo 'attribute not set'; else xattr -p '%s' '%s' 2>/dev/null || echo 'attribute not set'; fi",
			key, path, key, path,
		),
	})

	return TaskOutput{
		Status:  "ok",
		Message: "",
		Facts: map[string]interface{}{
			"shell_exec": shellExec,
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
