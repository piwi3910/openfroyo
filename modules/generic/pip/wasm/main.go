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

	// Extract parameters
	name, _ := input.Vars["name"].(string)
	state, _ := input.Vars["state"].(string)
	if state == "" {
		state = "present"
	}
	version, _ := input.Vars["version"].(string)
	requirements, _ := input.Vars["requirements"].(string)
	virtualenv, _ := input.Vars["virtualenv"].(string)
	extraArgs, _ := input.Vars["extra_args"].(string)
	executable, _ := input.Vars["executable"].(string)
	if executable == "" {
		executable = "pip3"
	}
	user, _ := input.Vars["user"].(bool)

	// Validate input
	if name == "" && requirements == "" {
		outputError("Missing required variable: name or requirements must be specified")
		return
	}

	// Build the pip command
	result := buildPipCommand(name, state, version, requirements, virtualenv, extraArgs, executable, user)
	printOutput(result)
}

func buildPipCommand(name, state, version, requirements, virtualenv, extraArgs, executable string, user bool) TaskOutput {
	var script strings.Builder

	// Determine the pip executable to use
	pipExec := executable
	if virtualenv != "" {
		pipExec = fmt.Sprintf("%s/bin/pip", virtualenv)
	}

	script.WriteString("# Manage pip package\n")

	// If requirements file is specified, use that instead
	if requirements != "" {
		script.WriteString(fmt.Sprintf("echo 'Installing from requirements file: %s'\n", requirements))

		cmd := fmt.Sprintf("%s install -r %s", pipExec, requirements)
		if extraArgs != "" {
			cmd += " " + extraArgs
		}
		script.WriteString(cmd + "\n")

		return buildShellExecOutput(script.String())
	}

	// Handle individual package operations
	switch state {
	case "present":
		// Check if package is installed
		script.WriteString(fmt.Sprintf("if %s show %s >/dev/null 2>&1; then\n", pipExec, name))
		script.WriteString(fmt.Sprintf("  echo 'Package %s is already installed'\n", name))
		script.WriteString("else\n")
		script.WriteString(fmt.Sprintf("  echo 'Installing package: %s'\n", name))

		// Build install command
		packageSpec := name
		if version != "" {
			packageSpec = fmt.Sprintf("%s==%s", name, version)
		}

		cmd := fmt.Sprintf("  %s install", pipExec)
		if user {
			cmd += " --user"
		}
		cmd += " " + packageSpec
		if extraArgs != "" {
			cmd += " " + extraArgs
		}
		script.WriteString(cmd + "\n")
		script.WriteString("fi\n")

	case "absent":
		// Check if package is installed before removing
		script.WriteString(fmt.Sprintf("if %s show %s >/dev/null 2>&1; then\n", pipExec, name))
		script.WriteString(fmt.Sprintf("  echo 'Removing package: %s'\n", name))
		script.WriteString(fmt.Sprintf("  %s uninstall -y %s\n", pipExec, name))
		script.WriteString("else\n")
		script.WriteString(fmt.Sprintf("  echo 'Package %s is not installed'\n", name))
		script.WriteString("fi\n")

	case "latest":
		// Check if package is installed
		script.WriteString(fmt.Sprintf("if %s show %s >/dev/null 2>&1; then\n", pipExec, name))
		script.WriteString(fmt.Sprintf("  echo 'Upgrading package to latest: %s'\n", name))

		cmd := fmt.Sprintf("  %s install --upgrade %s", pipExec, name)
		if user {
			cmd += " --user"
		}
		if extraArgs != "" {
			cmd += " " + extraArgs
		}
		script.WriteString(cmd + "\n")
		script.WriteString("else\n")
		script.WriteString(fmt.Sprintf("  echo 'Installing package: %s'\n", name))

		cmd = fmt.Sprintf("  %s install %s", pipExec, name)
		if user {
			cmd += " --user"
		}
		if extraArgs != "" {
			cmd += " " + extraArgs
		}
		script.WriteString(cmd + "\n")
		script.WriteString("fi\n")

	default:
		return TaskOutput{
			Status:  "failed",
			Message: fmt.Sprintf("Invalid state: %s (must be present, absent, or latest)", state),
			Facts:   make(map[string]interface{}),
		}
	}

	return buildShellExecOutput(script.String())
}

func buildShellExecOutput(script string) TaskOutput {
	shellExec := []map[string]interface{}{
		{
			"type":    "shell",
			"command": script,
		},
	}

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
