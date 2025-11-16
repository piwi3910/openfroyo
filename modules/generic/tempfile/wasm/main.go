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

	// Extract parameters with defaults
	state, _ := input.Vars["state"].(string)
	if state == "" {
		state = "file"
	}

	path, _ := input.Vars["path"].(string)
	if path == "" {
		path = "/tmp"
	}

	prefix, _ := input.Vars["prefix"].(string)
	if prefix == "" {
		prefix = "ansible."
	}

	suffix, _ := input.Vars["suffix"].(string)

	// Build shell commands
	result := buildTempfileCommands(state, path, prefix, suffix)
	printOutput(result)
}

func buildTempfileCommands(state, path, prefix, suffix string) TaskOutput {
	var shellExec []map[string]interface{}

	// Construct the mktemp template
	template := prefix + "XXXXXX" + suffix

	var command string
	switch state {
	case "file":
		// Create temporary file
		command = fmt.Sprintf("mktemp -p '%s' '%s'", path, template)

	case "directory":
		// Create temporary directory
		command = fmt.Sprintf("mktemp -d -p '%s' '%s'", path, template)

	default:
		return TaskOutput{
			Status:  "failed",
			Message: fmt.Sprintf("Invalid state: %s (must be file or directory)", state),
			Facts:   make(map[string]interface{}),
		}
	}

	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": command,
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
