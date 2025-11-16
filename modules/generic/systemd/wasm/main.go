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
	name, _ := input.Vars["name"].(string)
	if name == "" {
		outputError("Missing required variable: name (service name)")
		return
	}

	state, _ := input.Vars["state"].(string)
	enabled := extractBoolOrString(input.Vars["enabled"])
	daemonReload, _ := input.Vars["daemon_reload"].(bool)
	masked := extractBoolOrString(input.Vars["masked"])

	// Build the systemd command
	result := buildSystemdCommand(name, state, enabled, daemonReload, masked)
	printOutput(result)
}

func extractBoolOrString(val interface{}) string {
	if val == nil {
		return ""
	}

	switch v := val.(type) {
	case bool:
		if v {
			return "true"
		}
		return "false"
	case string:
		return v
	default:
		return ""
	}
}

func buildSystemdCommand(name, state, enabled string, daemonReload bool, masked string) TaskOutput {
	var commands []string

	// Build shell script that executes systemd commands
	script := `#!/bin/bash
set -e

SERVICE_NAME="` + name + `"
`

	// Check if systemd is available
	script += `
# Check if systemd is available
if ! command -v systemctl >/dev/null 2>&1; then
    echo "systemd is not available on this system"
    exit 1
fi

echo "Managing systemd service: $SERVICE_NAME"
`

	// Daemon reload if requested
	if daemonReload {
		script += `
echo "Reloading systemd daemon..."
sudo systemctl daemon-reload
`
		commands = append(commands, "daemon-reload")
	}

	// Handle masking/unmasking
	if masked != "" {
		if masked == "true" {
			script += `
echo "Masking service..."
sudo systemctl mask "$SERVICE_NAME"
`
			commands = append(commands, "mask")
		} else if masked == "false" {
			script += `
echo "Unmasking service..."
sudo systemctl unmask "$SERVICE_NAME"
`
			commands = append(commands, "unmask")
		}
	}

	// Handle state changes
	if state != "" {
		switch strings.ToLower(state) {
		case "started":
			script += `
# Check if service is active
if systemctl is-active --quiet "$SERVICE_NAME"; then
    echo "Service is already running"
else
    echo "Starting service..."
    sudo systemctl start "$SERVICE_NAME"
fi
`
			commands = append(commands, "start")
		case "stopped":
			script += `
# Check if service is active
if systemctl is-active --quiet "$SERVICE_NAME"; then
    echo "Stopping service..."
    sudo systemctl stop "$SERVICE_NAME"
else
    echo "Service is already stopped"
fi
`
			commands = append(commands, "stop")
		case "restarted":
			script += `
echo "Restarting service..."
sudo systemctl restart "$SERVICE_NAME"
`
			commands = append(commands, "restart")
		case "reloaded":
			script += `
echo "Reloading service..."
sudo systemctl reload "$SERVICE_NAME" || sudo systemctl reload-or-restart "$SERVICE_NAME"
`
			commands = append(commands, "reload")
		}
	}

	// Handle enabled/disabled
	if enabled != "" {
		if enabled == "true" {
			script += `
# Check if service is enabled
if systemctl is-enabled --quiet "$SERVICE_NAME" 2>/dev/null; then
    echo "Service is already enabled"
else
    echo "Enabling service..."
    sudo systemctl enable "$SERVICE_NAME"
fi
`
			commands = append(commands, "enable")
		} else if enabled == "false" {
			script += `
# Check if service is enabled
if systemctl is-enabled --quiet "$SERVICE_NAME" 2>/dev/null; then
    echo "Disabling service..."
    sudo systemctl disable "$SERVICE_NAME"
else
    echo "Service is already disabled"
fi
`
			commands = append(commands, "disable")
		}
	}

	// Get final status
	script += `
# Show service status
echo "Service status:"
systemctl status "$SERVICE_NAME" --no-pager || true
`

	shellExec := []map[string]interface{}{
		{
			"type":    "shell",
			"command": script,
		},
	}

	message := fmt.Sprintf("Executing systemd commands for service '%s': %s", name, strings.Join(commands, ", "))
	if len(commands) == 0 {
		message = fmt.Sprintf("No systemd operations requested for service '%s'", name)
	}

	return TaskOutput{
		Status:  "ok",
		Message: message,
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
