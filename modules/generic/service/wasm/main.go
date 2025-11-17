package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Input represents the module input parameters
type Input struct {
	Name         string `json:"name"`
	State        string `json:"state"`
	Enabled      *bool  `json:"enabled"`
	DaemonReload bool   `json:"daemon_reload"`
	Sleep        int    `json:"sleep"`
}

// Output represents the module output
type Output struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message"`
	Facts   map[string]interface{} `json:"facts"`
}

// ShellExec represents a shell command to execute
type ShellExec struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

func main() {
	// Read input from stdin
	var input Input
	decoder := json.NewDecoder(strings.NewReader(getInput()))
	if err := decoder.Decode(&input); err != nil {
		output := Output{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to parse input: %v", err),
			Facts:   map[string]interface{}{},
		}
		printOutput(output)
		return
	}

	// Validate required parameters
	if input.Name == "" {
		output := Output{
			Status:  "failed",
			Message: "Service name is required",
			Facts:   map[string]interface{}{},
		}
		printOutput(output)
		return
	}

	// Validate state
	validStates := map[string]bool{
		"started":   true,
		"stopped":   true,
		"restarted": true,
		"reloaded":  true,
	}
	if input.State != "" && !validStates[input.State] {
		output := Output{
			Status:  "failed",
			Message: fmt.Sprintf("Invalid state: %s (must be started, stopped, restarted, or reloaded)", input.State),
			Facts:   map[string]interface{}{},
		}
		printOutput(output)
		return
	}

	// Set default state
	if input.State == "" {
		input.State = "started"
	}

	// Build commands to manage service
	commands := buildServiceCommands(input)

	output := Output{
		Status:  "changed",
		Message: fmt.Sprintf("Managing service %s (state=%s, enabled=%v)", input.Name, input.State, input.Enabled),
		Facts: map[string]interface{}{
			"shell_exec":    commands,
			"service_name":  input.Name,
			"service_state": input.State,
		},
	}

	printOutput(output)
}

func buildServiceCommands(input Input) []ShellExec {
	var commands []ShellExec

	// Detect OS first
	commands = append(commands, ShellExec{
		Type:    "shell",
		Command: "uname -s || echo Windows_NT",
	})

	// Store OS detection result in variable for conditional execution
	// Note: This is a simplified version - actual implementation would need
	// the runner to support conditional command execution based on previous results

	// Generate Linux systemd commands
	linuxCommands := buildLinuxCommands(input)

	// Generate Windows commands
	windowsCommands := buildWindowsCommands(input)

	// For now, we'll generate both and the runner will need to execute based on OS
	// In a more advanced implementation, we'd have conditional execution support

	// Add platform detection and conditional execution
	commands = append(commands, ShellExec{
		Type:    "shell",
		Command: fmt.Sprintf("if command -v systemctl >/dev/null 2>&1; then OS_TYPE=linux; else OS_TYPE=windows; fi && echo \"OS: $OS_TYPE\""),
	})

	// Add Linux commands (will be executed if systemctl exists)
	for _, cmd := range linuxCommands {
		commands = append(commands, cmd)
	}

	// Add Windows commands (will be executed if systemctl doesn't exist)
	// Note: In production, these should be conditional
	for _, cmd := range windowsCommands {
		commands = append(commands, cmd)
	}

	return commands
}

func buildLinuxCommands(input Input) []ShellExec {
	var commands []ShellExec

	// Check if systemctl exists (guard clause)
	commands = append(commands, ShellExec{
		Type:    "shell",
		Command: "command -v systemctl >/dev/null 2>&1 || exit 0",
	})

	// Daemon reload if requested
	if input.DaemonReload {
		commands = append(commands, ShellExec{
			Type:    "shell",
			Command: "systemctl daemon-reload",
		})
	}

	// Manage enabled state first (if specified)
	if input.Enabled != nil {
		if *input.Enabled {
			commands = append(commands, ShellExec{
				Type:    "shell",
				Command: fmt.Sprintf("systemctl enable '%s'", escapeShell(input.Name)),
			})
		} else {
			commands = append(commands, ShellExec{
				Type:    "shell",
				Command: fmt.Sprintf("systemctl disable '%s'", escapeShell(input.Name)),
			})
		}
	}

	// Manage service state
	switch input.State {
	case "started":
		// Check if service is already running (idempotency)
		commands = append(commands, ShellExec{
			Type:    "shell",
			Command: fmt.Sprintf("systemctl is-active '%s' >/dev/null 2>&1 && echo 'Service already running' || systemctl start '%s'",
				escapeShell(input.Name), escapeShell(input.Name)),
		})
	case "stopped":
		// Check if service is already stopped (idempotency)
		commands = append(commands, ShellExec{
			Type:    "shell",
			Command: fmt.Sprintf("systemctl is-active '%s' >/dev/null 2>&1 && systemctl stop '%s' || echo 'Service already stopped'",
				escapeShell(input.Name), escapeShell(input.Name)),
		})
	case "restarted":
		commands = append(commands, ShellExec{
			Type:    "shell",
			Command: fmt.Sprintf("systemctl restart '%s'", escapeShell(input.Name)),
		})
	case "reloaded":
		commands = append(commands, ShellExec{
			Type:    "shell",
			Command: fmt.Sprintf("systemctl reload '%s'", escapeShell(input.Name)),
		})
	}

	// Sleep if requested
	if input.Sleep > 0 {
		commands = append(commands, ShellExec{
			Type:    "shell",
			Command: fmt.Sprintf("sleep %d", input.Sleep),
		})
	}

	// Verify service status
	commands = append(commands, ShellExec{
		Type:    "shell",
		Command: fmt.Sprintf("systemctl status '%s' --no-pager || true", escapeShell(input.Name)),
	})

	return commands
}

func buildWindowsCommands(input Input) []ShellExec {
	var commands []ShellExec

	// Check if we're on Windows (guard clause)
	commands = append(commands, ShellExec{
		Type:    "shell",
		Command: "command -v systemctl >/dev/null 2>&1 && exit 0 || true",
	})

	// Manage enabled state (startup type) if specified
	if input.Enabled != nil {
		if *input.Enabled {
			// Set service to automatic startup
			commands = append(commands, ShellExec{
				Type:    "shell",
				Command: fmt.Sprintf("sc.exe config \"%s\" start= auto", escapeShell(input.Name)),
			})
		} else {
			// Set service to manual startup (disabled)
			commands = append(commands, ShellExec{
				Type:    "shell",
				Command: fmt.Sprintf("sc.exe config \"%s\" start= demand", escapeShell(input.Name)),
			})
		}
	}

	// Manage service state
	switch input.State {
	case "started":
		// Check if service is running, start if not
		commands = append(commands, ShellExec{
			Type:    "shell",
			Command: fmt.Sprintf("sc.exe query \"%s\" | findstr \"RUNNING\" >nul && echo Service already running || sc.exe start \"%s\"",
				escapeShell(input.Name), escapeShell(input.Name)),
		})
	case "stopped":
		// Check if service is stopped, stop if not
		commands = append(commands, ShellExec{
			Type:    "shell",
			Command: fmt.Sprintf("sc.exe query \"%s\" | findstr \"STOPPED\" >nul && echo Service already stopped || sc.exe stop \"%s\"",
				escapeShell(input.Name), escapeShell(input.Name)),
		})
	case "restarted":
		// Stop then start
		commands = append(commands, ShellExec{
			Type:    "shell",
			Command: fmt.Sprintf("sc.exe stop \"%s\" && timeout /t 2 /nobreak >nul && sc.exe start \"%s\"",
				escapeShell(input.Name), escapeShell(input.Name)),
		})
	case "reloaded":
		// Windows doesn't have reload, so we'll restart
		commands = append(commands, ShellExec{
			Type:    "shell",
			Command: fmt.Sprintf("sc.exe stop \"%s\" && timeout /t 2 /nobreak >nul && sc.exe start \"%s\"",
				escapeShell(input.Name), escapeShell(input.Name)),
		})
	}

	// Sleep if requested
	if input.Sleep > 0 {
		commands = append(commands, ShellExec{
			Type:    "shell",
			Command: fmt.Sprintf("timeout /t %d /nobreak >nul", input.Sleep),
		})
	}

	// Query service status
	commands = append(commands, ShellExec{
		Type:    "shell",
		Command: fmt.Sprintf("sc.exe query \"%s\"", escapeShell(input.Name)),
	})

	return commands
}

func escapeShell(s string) string {
	// Escape single quotes for shell command
	return strings.ReplaceAll(s, "'", "'\\''")
}

func getInput() string {
	// This would normally read from stdin, but in WASM we get it differently
	// For now, return empty string and let the host provide it
	return ""
}

func printOutput(output Output) {
	data, _ := json.Marshal(output)
	fmt.Println(string(data))
}
