package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	src, _ := input.Vars["src"].(string)
	if src == "" {
		outputError("Missing required variable: src")
		return
	}

	dest, _ := input.Vars["dest"].(string)
	if dest == "" {
		outputError("Missing required variable: dest")
		return
	}

	// Get file content (base64 encoded) from executor
	fileContentB64, ok := input.Vars["_file_content"].(string)
	if !ok || fileContentB64 == "" {
		outputError("No file content provided (_file_content variable missing)")
		return
	}

	// Decode file content
	fileContent, err := base64.StdEncoding.DecodeString(fileContentB64)
	if err != nil {
		outputError(fmt.Sprintf("Failed to decode file content: %v", err))
		return
	}

	// Extract optional parameters
	mode, _ := input.Vars["mode"].(string)
	owner, _ := input.Vars["owner"].(string)
	group, _ := input.Vars["group"].(string)
	backup, _ := input.Vars["backup"].(bool)

	// Build shell commands
	result := buildCopyCommands(dest, string(fileContent), mode, owner, group, backup)
	printOutput(result)
}

func buildCopyCommands(dest, content, mode, owner, group string, backup bool) TaskOutput {
	var shellExec []map[string]interface{}

	// Create destination directory if needed
	destDir := filepath.Dir(dest)
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": fmt.Sprintf("mkdir -p '%s'", destDir),
	})

	// Backup existing file if requested
	if backup {
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("[ -f '%s' ] && cp '%s' '%s.backup' || true", dest, dest, dest),
		})
	}

	// Write the file content
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "file_write",
		"path":    dest,
		"content": content,
		"mode":    0644, // Default mode, will be changed if mode is specified
	})

	// Set permissions if specified
	if mode != "" {
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("chmod %s '%s'", mode, dest),
		})
	}

	// Set ownership if specified
	if owner != "" || group != "" {
		var chownTarget string
		if owner != "" && group != "" {
			chownTarget = fmt.Sprintf("%s:%s", owner, group)
		} else if owner != "" {
			chownTarget = owner
		} else {
			chownTarget = fmt.Sprintf(":%s", group)
		}

		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("chown %s '%s'", chownTarget, dest),
		})
	}

	// Verify the file was created
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": fmt.Sprintf("[ -f '%s' ] && ls -l '%s'", dest, dest),
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
