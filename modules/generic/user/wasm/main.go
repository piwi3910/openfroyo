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
		outputError("Missing required variable: name (username)")
		return
	}

	state, _ := input.Vars["state"].(string)
	if state == "" {
		state = "present"
	}

	uid, _ := input.Vars["uid"].(string)
	group, _ := input.Vars["group"].(string)
	shell, _ := input.Vars["shell"].(string)
	if shell == "" {
		shell = "/bin/bash"
	}
	home, _ := input.Vars["home"].(string)
	createHome, _ := input.Vars["create_home"].(bool)
	password, _ := input.Vars["password"].(string)
	comment, _ := input.Vars["comment"].(string)
	system, _ := input.Vars["system"].(bool)

	// Handle groups array
	var groups []string
	if groupsVar, ok := input.Vars["groups"]; ok {
		if groupsArray, ok := groupsVar.([]interface{}); ok {
			for _, g := range groupsArray {
				if groupStr, ok := g.(string); ok {
					groups = append(groups, groupStr)
				}
			}
		}
	}

	// Build the user management command
	result := buildUserCommand(name, state, uid, group, groups, shell, home, createHome, password, comment, system)
	printOutput(result)
}

func buildUserCommand(name, state, uid, primaryGroup string, groups []string, shell, home string, createHome bool, password, comment string, system bool) TaskOutput {
	// Build shell commands to execute
	var script string

	switch state {
	case "present":
		// Check if user exists
		script = fmt.Sprintf(`
# Check if user exists
if id "%s" >/dev/null 2>&1; then
    echo "User %s already exists"
    EXISTING=1
else
    echo "Creating user %s"
    EXISTING=0
fi
`, name, name, name)

		// Build useradd command
		useraddCmd := "sudo useradd"

		if system {
			useraddCmd += " --system"
		}

		if uid != "" {
			useraddCmd += fmt.Sprintf(" --uid %s", uid)
		}

		if primaryGroup != "" {
			useraddCmd += fmt.Sprintf(" --gid %s", primaryGroup)
		}

		if len(groups) > 0 {
			useraddCmd += fmt.Sprintf(" --groups %s", strings.Join(groups, ","))
		}

		if shell != "" {
			useraddCmd += fmt.Sprintf(" --shell %s", shell)
		}

		if home != "" {
			useraddCmd += fmt.Sprintf(" --home %s", home)
		}

		if createHome {
			useraddCmd += " --create-home"
		} else {
			useraddCmd += " --no-create-home"
		}

		if comment != "" {
			useraddCmd += fmt.Sprintf(" --comment '%s'", comment)
		}

		useraddCmd += fmt.Sprintf(" %s", name)

		// Build usermod command for existing users
		usermodCmd := "sudo usermod"

		if uid != "" {
			usermodCmd += fmt.Sprintf(" --uid %s", uid)
		}

		if primaryGroup != "" {
			usermodCmd += fmt.Sprintf(" --gid %s", primaryGroup)
		}

		if len(groups) > 0 {
			usermodCmd += fmt.Sprintf(" --groups %s", strings.Join(groups, ","))
		}

		if shell != "" {
			usermodCmd += fmt.Sprintf(" --shell %s", shell)
		}

		if home != "" {
			usermodCmd += fmt.Sprintf(" --home %s", home)
		}

		if comment != "" {
			usermodCmd += fmt.Sprintf(" --comment '%s'", comment)
		}

		usermodCmd += fmt.Sprintf(" %s", name)

		script += fmt.Sprintf(`
if [ $EXISTING -eq 0 ]; then
    %s
    echo "User %s created"
else
    # Update existing user if needed
    %s
    echo "User %s updated"
fi
`, useraddCmd, name, usermodCmd, name)

		// Set password if provided
		if password != "" {
			script += fmt.Sprintf(`
# Set password (encrypted)
echo '%s:%s' | sudo chpasswd -e
echo "Password set for user %s"
`, name, password, name)
		}

	case "absent":
		script = fmt.Sprintf(`
# Check if user exists
if id "%s" >/dev/null 2>&1; then
    echo "Removing user %s"
    sudo userdel -r %s
    echo "User %s removed"
else
    echo "User %s does not exist"
fi
`, name, name, name, name, name)
	}

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
