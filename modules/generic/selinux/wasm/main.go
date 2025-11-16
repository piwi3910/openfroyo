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

	// Extract required state parameter
	stateInterface, ok := input.Vars["state"]
	if !ok {
		outputError("Missing required variable: state")
		return
	}

	stateStr, ok := stateInterface.(string)
	if !ok {
		outputError("Variable 'state' must be a string")
		return
	}

	stateStr = strings.ToLower(strings.TrimSpace(stateStr))

	// Validate state
	validStates := map[string]bool{
		"enforcing":  true,
		"permissive": true,
		"disabled":   true,
	}

	if !validStates[stateStr] {
		outputError(fmt.Sprintf("Invalid state: %s (must be 'enforcing', 'permissive', or 'disabled')", stateStr))
		return
	}

	// Get policy (default: "targeted")
	policy := "targeted"
	if policyInterface, ok := input.Vars["policy"]; ok {
		if policyStr, ok := policyInterface.(string); ok {
			policy = policyStr
		}
	}

	// Get config file (default: "/etc/selinux/config")
	configFile := "/etc/selinux/config"
	if configInterface, ok := input.Vars["configfile"]; ok {
		if configStr, ok := configInterface.(string); ok {
			configFile = configStr
		}
	}

	var shellExec []map[string]interface{}

	// Get current SELinux status
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": "getenforce",
		"capture": "current_mode",
	})

	// Get current policy
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": "sestatus | grep 'Loaded policy' | awk '{print $3}'",
		"capture": "current_policy",
	})

	// Set runtime mode if not disabled
	// Note: setenforce cannot enable/disable, only switch between enforcing/permissive
	if stateStr == "enforcing" {
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": "setenforce 1",
		})
	} else if stateStr == "permissive" {
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": "setenforce 0",
		})
	}

	// Update config file for persistence
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": fmt.Sprintf("sed -i 's/^SELINUX=.*/SELINUX=%s/' %s", stateStr, configFile),
	})

	// Update policy in config file
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": fmt.Sprintf("sed -i 's/^SELINUXTYPE=.*/SELINUXTYPE=%s/' %s", policy, configFile),
	})

	// Add warning message about reboot requirement
	message := ""
	if stateStr == "disabled" {
		message = "SELinux set to disabled. A reboot is required for this change to take effect."
	} else if stateStr == "enforcing" {
		message = "SELinux set to enforcing mode. If switching from disabled, a reboot is required."
	}

	result := TaskOutput{
		Status:  "ok",
		Message: message,
		Facts: map[string]interface{}{
			"shell_exec": shellExec,
			"selinux": map[string]interface{}{
				"state":  stateStr,
				"policy": policy,
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
