package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
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

	// Extract state (default: "reboot")
	state := "reboot"
	if stateInterface, exists := input.Vars["state"]; exists {
		if str, ok := stateInterface.(string); ok {
			state = strings.ToLower(strings.TrimSpace(str))
		}
	}

	// Validate state
	validStates := map[string]bool{
		"reboot":   true,
		"shutdown": true,
		"halt":     true,
		"cancel":   true,
	}
	if !validStates[state] {
		outputError(fmt.Sprintf("Invalid state '%s'. Must be one of: reboot, shutdown, halt, cancel", state))
		return
	}

	// Extract delay (default: 0)
	delay := 0
	if delayInterface, exists := input.Vars["delay"]; exists {
		switch v := delayInterface.(type) {
		case float64:
			delay = int(v)
		case int:
			delay = v
		case string:
			if parsedDelay, err := strconv.Atoi(v); err == nil {
				delay = parsedDelay
			}
		}
	}

	if delay < 0 {
		outputError("Delay cannot be negative")
		return
	}

	// Extract message
	message := ""
	if msgInterface, exists := input.Vars["message"]; exists {
		if str, ok := msgInterface.(string); ok {
			message = str
		}
	}

	// Extract force flag (default: false)
	force := false
	if forceInterface, exists := input.Vars["force"]; exists {
		switch v := forceInterface.(type) {
		case bool:
			force = v
		case string:
			force = strings.ToLower(v) == "true" || v == "1"
		case float64:
			force = v != 0
		case int:
			force = v != 0
		}
	}

	var cmd string
	var actionMsg string

	switch state {
	case "cancel":
		// Cancel scheduled shutdown/reboot
		cmd = "shutdown -c 2>&1 || echo 'No shutdown scheduled'"
		actionMsg = "Cancelling scheduled shutdown/reboot"

	case "reboot":
		if force {
			// Force immediate reboot (no graceful shutdown)
			cmd = "reboot -f"
			actionMsg = "Forcing immediate reboot"
		} else if delay > 0 {
			// Scheduled reboot
			if message != "" {
				cmd = fmt.Sprintf("shutdown -r +%d %s", delay, shellEscape(message))
			} else {
				cmd = fmt.Sprintf("shutdown -r +%d", delay)
			}
			actionMsg = fmt.Sprintf("Scheduling reboot in %d minutes", delay)
		} else {
			// Immediate reboot (but graceful)
			if message != "" {
				cmd = fmt.Sprintf("shutdown -r now %s", shellEscape(message))
			} else {
				cmd = "shutdown -r now"
			}
			actionMsg = "Rebooting system now"
		}

	case "shutdown":
		if force {
			// Force immediate shutdown
			cmd = "poweroff -f"
			actionMsg = "Forcing immediate shutdown"
		} else if delay > 0 {
			// Scheduled shutdown
			if message != "" {
				cmd = fmt.Sprintf("shutdown -h +%d %s", delay, shellEscape(message))
			} else {
				cmd = fmt.Sprintf("shutdown -h +%d", delay)
			}
			actionMsg = fmt.Sprintf("Scheduling shutdown in %d minutes", delay)
		} else {
			// Immediate shutdown (but graceful)
			if message != "" {
				cmd = fmt.Sprintf("shutdown -h now %s", shellEscape(message))
			} else {
				cmd = "shutdown -h now"
			}
			actionMsg = "Shutting down system now"
		}

	case "halt":
		// System halt (stop CPU without power off)
		cmd = "halt"
		actionMsg = "Halting system"
	}

	// Wrap command to capture output and exit code
	fullCmd := fmt.Sprintf(`
OUTPUT=$(%s 2>&1)
EXIT_CODE=$?
echo "{\"exit_code\": $EXIT_CODE, \"output\": \"$(echo "$OUTPUT" | sed 's/"/\\"/g' | tr '\n' ' ')\"}"
`, cmd)

	// Return shell_exec fact for the generic handler
	result := TaskOutput{
		Status:  "changed", // System state will change
		Message: actionMsg,
		Facts: map[string]interface{}{
			"shell_exec": []map[string]interface{}{
				{
					"type":    "shell",
					"command": fullCmd,
				},
			},
			"reboot_state": state,
			"reboot_delay": delay,
			"force":        force,
		},
	}
	printOutput(result)
}

func shellEscape(s string) string {
	// Simple shell escaping: wrap in single quotes and escape any single quotes
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
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
