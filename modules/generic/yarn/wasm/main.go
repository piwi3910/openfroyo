package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type TaskInput struct {
	Vars    map[string]interface{} `json:"vars"`
	Context map[string]interface{} `json:"context"`
}

type TaskOutput struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message"`
	Facts   map[string]interface{} `json:"facts"`
}

func main() {
	if len(os.Args) < 2 {
		outputError("no input provided")
		return
	}

	var input TaskInput
	if err := json.Unmarshal([]byte(os.Args[1]), &input); err != nil {
		outputError(fmt.Sprintf("failed to parse input: %v", err))
		return
	}

	result := manageYarnPackage(input)
	output, _ := json.Marshal(result)
	fmt.Println(string(output))
}

func manageYarnPackage(input TaskInput) TaskOutput {
	name, _ := input.Vars["name"].(string)
	if name == "" {
		return TaskOutput{
			Status:  "failed",
			Message: "package name is required",
			Facts:   make(map[string]interface{}),
		}
	}

	state, _ := input.Vars["state"].(string)
	if state == "" {
		state = "present"
	}

	version, _ := input.Vars["version"].(string)
	global, _ := input.Vars["global"].(bool)
	path, _ := input.Vars["path"].(string)
	production, _ := input.Vars["production"].(bool)

	script := buildYarnScript(name, state, version, global, path, production)

	return TaskOutput{
		Status:  "ok",
		Message: "",
		Facts: map[string]interface{}{
			"shell_exec": []map[string]interface{}{
				{
					"type":    "shell",
					"command": script,
				},
			},
		},
	}
}

func buildYarnScript(name, state, version string, global bool, path string, production bool) string {
	script := `#!/bin/bash
set -e

# Check if yarn is installed
if ! command -v yarn >/dev/null 2>&1; then
  echo "ERROR: yarn is not installed"
  exit 1
fi

PACKAGE_NAME="%s"
STATE="%s"
VERSION="%s"
GLOBAL=%t
PATH_DIR="%s"
PRODUCTION=%t

`
	script = fmt.Sprintf(script, name, state, version, global, path, production)

	// Build package spec with version if provided
	script += `PACKAGE_SPEC="$PACKAGE_NAME"
if [ -n "$VERSION" ]; then
  PACKAGE_SPEC="${PACKAGE_NAME}@${VERSION}"
fi

`

	// Check installation status
	if global {
		script += `# Check if package is installed globally
if yarn global list --pattern "^$PACKAGE_NAME$" 2>/dev/null | grep -q "$PACKAGE_NAME"; then
  INSTALLED=true
else
  INSTALLED=false
fi

`
	} else {
		script += `# For local installation, change to specified directory
if [ -n "$PATH_DIR" ]; then
  if [ ! -d "$PATH_DIR" ]; then
    echo "ERROR: Directory $PATH_DIR does not exist"
    exit 1
  fi
  cd "$PATH_DIR"
fi

# Check if package is installed locally
if [ -f "package.json" ] && yarn list --pattern "^$PACKAGE_NAME$" 2>/dev/null | grep -q "$PACKAGE_NAME"; then
  INSTALLED=true
else
  INSTALLED=false
fi

`
	}

	// Handle different states
	script += `case "$STATE" in
  present)
    if [ "$INSTALLED" = "false" ]; then
`

	if global {
		script += `      echo "Installing package globally: $PACKAGE_SPEC"
      yarn global add "$PACKAGE_SPEC"
`
	} else {
		if production {
			script += `      echo "Installing production dependencies"
      yarn install --production
`
		} else {
			script += `      echo "Installing package locally: $PACKAGE_SPEC"
      yarn add "$PACKAGE_SPEC"
`
		}
	}

	script += `      echo "CHANGED=true"
    else
      echo "Package $PACKAGE_NAME is already installed"
      echo "CHANGED=false"
    fi
    ;;

  absent)
    if [ "$INSTALLED" = "true" ]; then
`

	if global {
		script += `      echo "Removing package globally: $PACKAGE_NAME"
      yarn global remove "$PACKAGE_NAME"
`
	} else {
		script += `      echo "Removing package locally: $PACKAGE_NAME"
      yarn remove "$PACKAGE_NAME"
`
	}

	script += `      echo "CHANGED=true"
    else
      echo "Package $PACKAGE_NAME is not installed"
      echo "CHANGED=false"
    fi
    ;;

  latest)
`

	if global {
		script += `    echo "Upgrading package globally: $PACKAGE_NAME"
    yarn global upgrade "$PACKAGE_NAME"
`
	} else {
		script += `    echo "Upgrading package locally: $PACKAGE_NAME"
    yarn upgrade "$PACKAGE_NAME"
`
	}

	script += `    echo "CHANGED=true"
    ;;

  *)
    echo "ERROR: Invalid state: $STATE (must be present, absent, or latest)"
    exit 1
    ;;
esac

echo "SUCCESS=true"
`

	return script
}

func outputError(message string) {
	result := TaskOutput{
		Status:  "failed",
		Message: message,
		Facts:   make(map[string]interface{}),
	}
	output, _ := json.Marshal(result)
	fmt.Println(string(output))
}
