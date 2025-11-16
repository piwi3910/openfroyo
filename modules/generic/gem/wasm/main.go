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
	name, _ := input.Vars["name"].(string)
	if name == "" {
		outputError("Missing required variable: name (gem name)")
		return
	}

	state, _ := input.Vars["state"].(string)
	if state == "" {
		state = "present" // Default to install
	}

	version, _ := input.Vars["version"].(string)
	userInstall, _ := input.Vars["user_install"].(bool)
	source, _ := input.Vars["source"].(string)
	installDir, _ := input.Vars["install_dir"].(string)

	// Build the gem command
	result := buildGemCommand(name, state, version, userInstall, source, installDir)
	printOutput(result)
}

func buildGemCommand(name, state, version string, userInstall bool, source, installDir string) TaskOutput {
	// Build shell script that checks gem status and runs appropriate command
	script := buildGemScript(name, state, version, userInstall, source, installDir)

	shellExec := []map[string]interface{}{
		{
			"type":    "shell",
			"command": script,
		},
	}

	return TaskOutput{
		Status:  "ok",
		Message: "", // Will be set by handler
		Facts: map[string]interface{}{
			"shell_exec": shellExec,
		},
	}
}

func buildGemScript(name, state, version string, userInstall bool, source, installDir string) string {
	script := `#!/bin/bash
# Gem management script

# Check if gem command exists
if ! command -v gem >/dev/null 2>&1; then
    echo "Error: gem command not found. Please install Ruby and RubyGems first."
    exit 1
fi

# Function to check if a gem is installed
gem_installed() {
    gem list -i "^$1$" >/dev/null 2>&1
    return $?
}

# Function to check if a specific version is installed
gem_version_installed() {
    if [ -n "$2" ]; then
        gem list -i "^$1$" -v "$2" >/dev/null 2>&1
        return $?
    else
        gem_installed "$1"
        return $?
    fi
}

`

	switch state {
	case "present":
		script += fmt.Sprintf(`
# Check if gem is already installed
if gem_version_installed "%s" "%s"; then
    echo "Gem %s`, name, version, name)
		if version != "" {
			script += fmt.Sprintf(` version %s`, version)
		}
		script += ` is already installed"
    exit 0
fi

echo "Installing gem %s`
		if version != "" {
			script += fmt.Sprintf(` version %s`, version)
		}
		script += `..."

# Build install command
INSTALL_CMD="gem install %s`

		// Add version if specified
		if version != "" {
			script += fmt.Sprintf(` -v %s`, version)
		}

		// Add user install flag
		if userInstall {
			script += ` --user-install`
		}

		// Add source if specified
		if source != "" {
			script += fmt.Sprintf(` --source %s`, source)
		}

		// Add install directory if specified
		if installDir != "" {
			script += fmt.Sprintf(` --install-dir %s`, installDir)
		}

		script += `"

echo "Running: $INSTALL_CMD"
eval $INSTALL_CMD
`

	case "absent":
		script += fmt.Sprintf(`
# Check if gem is installed
if ! gem_installed "%s"; then
    echo "Gem %s is not installed"
    exit 0
fi

echo "Uninstalling gem %s..."

# Uninstall gem (remove all versions, executables, ignore dependencies)
gem uninstall %s -x -I

if [ $? -eq 0 ]; then
    echo "Gem %s uninstalled successfully"
else
    echo "Failed to uninstall gem %s"
    exit 1
fi
`, name, name, name, name, name, name)

	case "latest":
		script += fmt.Sprintf(`
# Update to latest version
echo "Updating gem %s to latest version..."

# If source is specified, add it
`, name)

		updateCmd := fmt.Sprintf(`gem update %s`, name)
		if source != "" {
			updateCmd += fmt.Sprintf(` --source %s`, source)
		}

		script += fmt.Sprintf(`
%s

if [ $? -eq 0 ]; then
    echo "Gem %s updated to latest version"
else
    echo "Failed to update gem %s"
    exit 1
fi
`, updateCmd, name, name)

	default:
		script += fmt.Sprintf(`
echo "Error: Invalid state '%s'. Must be 'present', 'absent', or 'latest'"
exit 1
`, state)
	}

	return script
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
