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

	// Extract timezone name from vars
	nameInterface, ok := input.Vars["name"]
	if !ok {
		outputError("Missing required variable: name")
		return
	}

	nameStr, ok := nameInterface.(string)
	if !ok {
		outputError("Variable 'name' must be a string")
		return
	}

	// Validate timezone is not empty
	if strings.TrimSpace(nameStr) == "" {
		outputError("Timezone cannot be empty")
		return
	}

	// Return shell_exec facts to:
	// 1. Get current timezone
	// 2. Set timezone with timedatectl
	// 3. Create symlink as alternative method
	result := TaskOutput{
		Status:  "ok",
		Message: "",
		Facts: map[string]interface{}{
			"shell_exec": []map[string]interface{}{
				{
					"type":    "shell",
					"command": "timedatectl show -p Timezone --value",
					"capture": "current_timezone",
				},
				{
					"type":    "shell",
					"command": fmt.Sprintf("timedatectl set-timezone %s", nameStr),
				},
				{
					"type":    "shell",
					"command": fmt.Sprintf("ln -sf /usr/share/zoneinfo/%s /etc/localtime", nameStr),
				},
			},
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
