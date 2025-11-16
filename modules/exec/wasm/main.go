package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
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
	Status  string                 `json:"status"`  // "ok", "changed", "failed"
	Message string                 `json:"message"` // Human-readable message
	Facts   map[string]interface{} `json:"facts"`   // Discovered facts
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

	// Extract command from vars
	cmdInterface, ok := input.Vars["cmd"]
	if !ok {
		outputError("Missing required variable: cmd")
		return
	}

	cmdStr, ok := cmdInterface.(string)
	if !ok {
		outputError("Variable 'cmd' must be a string")
		return
	}

	// Parse command (simple shell-style parsing)
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		outputError("Command cannot be empty")
		return
	}

	// Execute command
	command := exec.Command(parts[0], parts[1:]...)

	output, err := command.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		// Command failed
		result := TaskOutput{
			Status:  "failed",
			Message: fmt.Sprintf("Command failed: %v\nOutput: %s", err, outputStr),
			Facts: map[string]interface{}{
				"stdout": outputStr,
				"error":  err.Error(),
			},
		}
		printOutput(result)
		return
	}

	// Command succeeded
	result := TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Command executed successfully: %s", cmdStr),
		Facts: map[string]interface{}{
			"stdout": strings.TrimSpace(outputStr),
		},
	}
	printOutput(result)
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
