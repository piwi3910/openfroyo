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
	path, _ := input.Vars["path"].(string)
	if path == "" {
		outputError("Missing required variable: path")
		return
	}

	regexp, _ := input.Vars["regexp"].(string)
	if regexp == "" {
		outputError("Missing required variable: regexp")
		return
	}

	replace, _ := input.Vars["replace"].(string)
	if replace == "" {
		outputError("Missing required variable: replace")
		return
	}

	backup, _ := input.Vars["backup"].(bool)

	// Build shell commands
	result := buildReplaceCommands(path, regexp, replace, backup)
	printOutput(result)
}

func buildReplaceCommands(path, regexp, replace string, backup bool) TaskOutput {
	var shellExec []map[string]interface{}

	// Escape special characters in the replacement string for sed
	// We need to escape forward slashes, backslashes, and ampersands
	escapedReplace := strings.ReplaceAll(replace, "\\", "\\\\")
	escapedReplace = strings.ReplaceAll(escapedReplace, "/", "\\/")
	escapedReplace = strings.ReplaceAll(escapedReplace, "&", "\\&")

	// Escape special characters in the regexp for sed
	escapedRegexp := strings.ReplaceAll(regexp, "/", "\\/")

	var sedCommand string
	if backup {
		// Create backup with .bak extension
		sedCommand = fmt.Sprintf("sed -i.bak 's/%s/%s/g' '%s'", escapedRegexp, escapedReplace, path)
	} else {
		// No backup
		sedCommand = fmt.Sprintf("sed -i '' 's/%s/%s/g' '%s'", escapedRegexp, escapedReplace, path)
	}

	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": sedCommand,
	})

	// Verify the file exists after replacement
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": fmt.Sprintf("[ -f '%s' ] && echo 'File modified successfully'", path),
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
