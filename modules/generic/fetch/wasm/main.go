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
	src, _ := input.Vars["src"].(string)
	if src == "" {
		outputError("Missing required variable: src")
		return
	}

	dest, _ := input.Vars["dest"].(string)
	if dest == "" {
		outputError("Missing required variable: dest")
		return
	}

	failOnMissing := true
	if val, ok := input.Vars["fail_on_missing"].(bool); ok {
		failOnMissing = val
	}

	// Build shell commands for fetching
	result := buildFetchCommands(src, dest, failOnMissing)
	printOutput(result)
}

func buildFetchCommands(src, dest string, failOnMissing bool) TaskOutput {
	var shellExec []map[string]interface{}

	// Check if source file exists
	if failOnMissing {
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("[ -f '%s' ] || { echo 'Source file not found: %s' >&2; exit 1; }", src, src),
		})
	} else {
		// Just check and warn if missing
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("[ -f '%s' ] || echo 'Warning: Source file not found: %s'", src, src),
		})
	}

	// Read file and base64 encode it
	// The output will be captured and used to write the local file
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": fmt.Sprintf("cat '%s' | base64", src),
	})

	// Get file stats for verification
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": fmt.Sprintf("stat -c '%%s %%a %%U %%G' '%s' 2>/dev/null || stat -f '%%z %%Lp %%Su %%Sg' '%s'", src, src),
	})

	return TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Fetching %s to %s", src, dest),
		Facts: map[string]interface{}{
			"shell_exec": shellExec,
			"fetch_src":  src,
			"fetch_dest": dest,
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
