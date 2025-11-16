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
		outputError("Missing required variable: name (group name)")
		return
	}

	state, _ := input.Vars["state"].(string)
	if state == "" {
		state = "present"
	}

	gid, _ := input.Vars["gid"].(string)
	system, _ := input.Vars["system"].(bool)

	// Build the group management command
	result := buildGroupCommand(name, state, gid, system)
	printOutput(result)
}

func buildGroupCommand(name, state, gid string, system bool) TaskOutput {
	// Build shell commands to execute
	var script string

	switch state {
	case "present":
		// Check if group exists
		script = fmt.Sprintf(`
# Check if group exists
if getent group "%s" >/dev/null 2>&1; then
    echo "Group %s already exists"
    EXISTING=1
else
    echo "Creating group %s"
    EXISTING=0
fi
`, name, name, name)

		// Build groupadd command
		groupaddCmd := "sudo groupadd"

		if system {
			groupaddCmd += " --system"
		}

		if gid != "" {
			groupaddCmd += fmt.Sprintf(" --gid %s", gid)
		}

		groupaddCmd += fmt.Sprintf(" %s", name)

		// Build groupmod command for existing groups
		groupmodCmd := "sudo groupmod"

		if gid != "" {
			groupmodCmd += fmt.Sprintf(" --gid %s", gid)
		}

		groupmodCmd += fmt.Sprintf(" %s", name)

		script += fmt.Sprintf(`
if [ $EXISTING -eq 0 ]; then
    %s
    echo "Group %s created"
else
    # Update existing group if needed
    %s
    echo "Group %s updated"
fi
`, groupaddCmd, name, groupmodCmd, name)

	case "absent":
		script = fmt.Sprintf(`
# Check if group exists
if getent group "%s" >/dev/null 2>&1; then
    echo "Removing group %s"
    sudo groupdel %s
    echo "Group %s removed"
else
    echo "Group %s does not exist"
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
