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

	// Extract required parameters
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

	if strings.TrimSpace(nameStr) == "" {
		outputError("Parameter name cannot be empty")
		return
	}

	// Get state (default: "present")
	state := "present"
	if stateInterface, ok := input.Vars["state"]; ok {
		if stateStr, ok := stateInterface.(string); ok {
			state = stateStr
		}
	}

	// Get sysctl_file (default: "/etc/sysctl.conf")
	sysctlFile := "/etc/sysctl.conf"
	if fileInterface, ok := input.Vars["sysctl_file"]; ok {
		if fileStr, ok := fileInterface.(string); ok {
			sysctlFile = fileStr
		}
	}

	// Get reload (default: true)
	reload := true
	if reloadInterface, ok := input.Vars["reload"]; ok {
		if reloadBool, ok := reloadInterface.(bool); ok {
			reload = reloadBool
		}
	}

	var shellExec []map[string]interface{}

	// Get current value
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": fmt.Sprintf("sysctl -n %s", nameStr),
		"capture": "current_value",
	})

	if state == "present" {
		// Value is required for present state
		valueInterface, ok := input.Vars["value"]
		if !ok {
			outputError("Missing required variable: value (required when state=present)")
			return
		}

		valueStr, ok := valueInterface.(string)
		if !ok {
			outputError("Variable 'value' must be a string")
			return
		}

		// Set value immediately
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("sysctl -w %s=%s", nameStr, valueStr),
		})

		// Persist to config file
		// First remove any existing entries for this parameter
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("sed -i '/%s/d' %s", strings.ReplaceAll(nameStr, ".", "\\."), sysctlFile),
		})

		// Then append the new setting
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("echo '%s=%s' >> %s", nameStr, valueStr, sysctlFile),
		})

	} else if state == "absent" {
		// Remove from config file
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("sed -i '/%s/d' %s", strings.ReplaceAll(nameStr, ".", "\\."), sysctlFile),
		})
	} else {
		outputError(fmt.Sprintf("Invalid state: %s (must be 'present' or 'absent')", state))
		return
	}

	// Reload sysctl if requested
	if reload {
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("sysctl -p %s", sysctlFile),
		})
	}

	result := TaskOutput{
		Status:  "ok",
		Message: "",
		Facts: map[string]interface{}{
			"shell_exec": shellExec,
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
