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
		state = "file"
	}

	mode, _ := input.Vars["mode"].(string)
	owner, _ := input.Vars["owner"].(string)
	group, _ := input.Vars["group"].(string)
	src, _ := input.Vars["src"].(string)
	recurse, _ := input.Vars["recurse"].(bool)

	// Build shell commands based on state
	result := buildFileCommands(path, state, mode, owner, group, src, recurse)
	printOutput(result)
}

func buildFileCommands(path, state, mode, owner, group, src string, recurse bool) TaskOutput {
	var shellExec []map[string]interface{}

	switch state {
	case "file":
		// Ensure file exists
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("[ -f '%s' ] || touch '%s'", path, path),
		})

	case "directory":
		// Create directory
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("mkdir -p '%s'", path),
		})

	case "link":
		// Create symlink
		if src == "" {
			return TaskOutput{
				Status:  "failed",
				Message: "src parameter required for state=link",
				Facts:   make(map[string]interface{}),
			}
		}
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("ln -sf '%s' '%s'", src, path),
		})

	case "absent":
		// Remove file/directory
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("rm -rf '%s'", path),
		})

	case "touch":
		// Touch file (update timestamp)
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("touch '%s'", path),
		})

	default:
		return TaskOutput{
			Status:  "failed",
			Message: fmt.Sprintf("Invalid state: %s (must be file, directory, link, absent, or touch)", state),
			Facts:   make(map[string]interface{}),
		}
	}

	// Set permissions if specified
	if mode != "" && state != "absent" {
		chmodCmd := fmt.Sprintf("chmod %s '%s'", mode, path)
		if recurse && state == "directory" {
			chmodCmd = fmt.Sprintf("chmod -R %s '%s'", mode, path)
		}
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": chmodCmd,
		})
	}

	// Set ownership if specified
	if (owner != "" || group != "") && state != "absent" {
		var chownTarget string
		if owner != "" && group != "" {
			chownTarget = fmt.Sprintf("%s:%s", owner, group)
		} else if owner != "" {
			chownTarget = owner
		} else {
			chownTarget = fmt.Sprintf(":%s", group)
		}

		chownCmd := fmt.Sprintf("chown %s '%s'", chownTarget, path)
		if recurse && state == "directory" {
			chownCmd = fmt.Sprintf("chown -R %s '%s'", chownTarget, path)
		}
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": chownCmd,
		})
	}

	// Add verification command
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": buildVerifyCommand(path, state),
	})

	return TaskOutput{
		Status:  "ok",
		Message: "",
		Facts: map[string]interface{}{
			"shell_exec": shellExec,
		},
	}
}

func buildVerifyCommand(path, state string) string {
	switch state {
	case "absent":
		return fmt.Sprintf("[ ! -e '%s' ] && echo 'removed' || echo 'exists'", path)
	case "directory":
		return fmt.Sprintf("[ -d '%s' ] && ls -ld '%s'", path, path)
	case "link":
		return fmt.Sprintf("[ -L '%s' ] && ls -l '%s'", path, path)
	default:
		return fmt.Sprintf("[ -e '%s' ] && ls -l '%s'", path, path)
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
