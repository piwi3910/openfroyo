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
		outputError("Missing required variable: name (package name)")
		return
	}

	state, _ := input.Vars["state"].(string)
	if state == "" {
		state = "present" // Default to install
	}

	version, _ := input.Vars["version"].(string)
	global, _ := input.Vars["global"].(bool)
	path, _ := input.Vars["path"].(string)
	production, _ := input.Vars["production"].(bool)

	// Build the npm command
	result := buildNpmCommand(name, state, version, global, path, production)
	printOutput(result)
}

func buildNpmCommand(name, state, version string, global bool, path string, production bool) TaskOutput {
	// Build shell script that checks npm package status and runs appropriate command
	script := buildNpmScript(name, state, version, global, path, production)

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

func buildNpmScript(name, state, version string, global bool, path string, production bool) string {
	script := `#!/bin/bash
# NPM package management script

# Check if npm command exists
if ! command -v npm >/dev/null 2>&1; then
    echo "Error: npm command not found. Please install Node.js and npm first."
    exit 1
fi

# Function to check if a package is installed globally
npm_global_installed() {
    npm list -g "$1" --depth=0 >/dev/null 2>&1
    return $?
}

# Function to check if a package is installed locally
npm_local_installed() {
    local check_path="${2:-.}"
    cd "$check_path" 2>/dev/null || return 1
    npm list "$1" --depth=0 >/dev/null 2>&1
    return $?
}

`

	// Set working directory if path is specified
	if path != "" {
		script += fmt.Sprintf(`
# Change to specified path
if [ ! -d "%s" ]; then
    echo "Error: Directory %s does not exist"
    exit 1
fi
cd "%s"
echo "Working in directory: $(pwd)"
`, path, path, path)
	}

	switch state {
	case "present":
		if global {
			// Global installation
			script += fmt.Sprintf(`
# Check if package is already installed globally
if npm_global_installed "%s"; then
    echo "Package %s is already installed globally"
    exit 0
fi

echo "Installing package %s globally..."

# Build install command
INSTALL_CMD="npm install -g %s`, name, name, name, name)

			if version != "" {
				script += fmt.Sprintf(`@%s`, version)
			}

			script += `"

echo "Running: $INSTALL_CMD"
eval $INSTALL_CMD

if [ $? -eq 0 ]; then
    echo "Package %s installed globally"
else
    echo "Failed to install package %s"
    exit 1
fi
`
		} else {
			// Local installation
			script += fmt.Sprintf(`
# Check if package is already installed locally
if npm_local_installed "%s" "%s"; then
    echo "Package %s is already installed locally"
    exit 0
fi

echo "Installing package %s locally..."

# Build install command
INSTALL_CMD="npm install %s`, name, path, name, name, name)

			if version != "" {
				script += fmt.Sprintf(`@%s`, version)
			}

			if production {
				script += ` --production`
			}

			script += `"

echo "Running: $INSTALL_CMD"
eval $INSTALL_CMD

if [ $? -eq 0 ]; then
    echo "Package %s installed locally"
else
    echo "Failed to install package %s"
    exit 1
fi
`
		}

	case "absent":
		if global {
			// Global uninstall
			script += fmt.Sprintf(`
# Check if package is installed globally
if ! npm_global_installed "%s"; then
    echo "Package %s is not installed globally"
    exit 0
fi

echo "Uninstalling package %s globally..."

npm uninstall -g %s

if [ $? -eq 0 ]; then
    echo "Package %s uninstalled globally"
else
    echo "Failed to uninstall package %s"
    exit 1
fi
`, name, name, name, name, name, name)
		} else {
			// Local uninstall
			script += fmt.Sprintf(`
# Check if package is installed locally
if ! npm_local_installed "%s" "%s"; then
    echo "Package %s is not installed locally"
    exit 0
fi

echo "Uninstalling package %s locally..."

npm uninstall %s

if [ $? -eq 0 ]; then
    echo "Package %s uninstalled locally"
else
    echo "Failed to uninstall package %s"
    exit 1
fi
`, name, path, name, name, name, name, name)
		}

	case "latest":
		if global {
			// Global update
			script += fmt.Sprintf(`
echo "Updating package %s to latest version globally..."

npm update -g %s

if [ $? -eq 0 ]; then
    echo "Package %s updated to latest version globally"
else
    echo "Failed to update package %s"
    exit 1
fi
`, name, name, name, name)
		} else {
			// Local update
			script += fmt.Sprintf(`
echo "Updating package %s to latest version locally..."

npm update %s

if [ $? -eq 0 ]; then
    echo "Package %s updated to latest version locally"
else
    echo "Failed to update package %s"
    exit 1
fi
`, name, name, name, name)
		}

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
