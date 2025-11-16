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
	path, _ := input.Vars["path"].(string)
	if path == "" {
		outputError("Missing required variable: path")
		return
	}

	state, _ := input.Vars["state"].(string)
	if state == "" {
		state = "present"
	}

	entity, _ := input.Vars["entity"].(string)
	etype, _ := input.Vars["etype"].(string)
	if etype == "" {
		etype = "user"
	}
	permissions, _ := input.Vars["permissions"].(string)
	recursive, _ := input.Vars["recursive"].(bool)

	// Build shell commands based on state
	result := buildACLCommands(path, state, entity, etype, permissions, recursive)
	printOutput(result)
}

func buildACLCommands(path, state, entity, etype, permissions string, recursive bool) TaskOutput {
	var shellExec []map[string]interface{}

	// Validate required parameters for present state
	if state == "present" {
		if entity == "" {
			return TaskOutput{
				Status:  "failed",
				Message: "entity parameter required for state=present",
				Facts:   make(map[string]interface{}),
			}
		}
		if permissions == "" {
			return TaskOutput{
				Status:  "failed",
				Message: "permissions parameter required for state=present",
				Facts:   make(map[string]interface{}),
			}
		}
	}

	// Validate etype
	if etype != "user" && etype != "group" && etype != "mask" && etype != "other" {
		return TaskOutput{
			Status:  "failed",
			Message: fmt.Sprintf("Invalid etype: %s (must be user, group, mask, or other)", etype),
			Facts:   make(map[string]interface{}),
		}
	}

	switch state {
	case "present":
		// Add or modify ACL entry
		aclSpec := fmt.Sprintf("%s:%s:%s", etype, entity, permissions)

		if recursive {
			shellExec = append(shellExec, map[string]interface{}{
				"type":    "shell",
				"command": fmt.Sprintf("setfacl -R -m '%s' '%s'", aclSpec, path),
			})
		} else {
			shellExec = append(shellExec, map[string]interface{}{
				"type":    "shell",
				"command": fmt.Sprintf("setfacl -m '%s' '%s'", aclSpec, path),
			})
		}

	case "absent":
		// Remove ACL entry
		if entity == "" {
			return TaskOutput{
				Status:  "failed",
				Message: "entity parameter required for state=absent",
				Facts:   make(map[string]interface{}),
			}
		}

		aclSpec := fmt.Sprintf("%s:%s", etype, entity)

		if recursive {
			shellExec = append(shellExec, map[string]interface{}{
				"type":    "shell",
				"command": fmt.Sprintf("setfacl -R -x '%s' '%s'", aclSpec, path),
			})
		} else {
			shellExec = append(shellExec, map[string]interface{}{
				"type":    "shell",
				"command": fmt.Sprintf("setfacl -x '%s' '%s'", aclSpec, path),
			})
		}

	default:
		return TaskOutput{
			Status:  "failed",
			Message: fmt.Sprintf("Invalid state: %s (must be present or absent)", state),
			Facts:   make(map[string]interface{}),
		}
	}

	// Add verification command to show current ACLs
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": fmt.Sprintf("getfacl '%s'", path),
	})

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
