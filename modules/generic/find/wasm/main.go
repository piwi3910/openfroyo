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

	patterns, _ := input.Vars["patterns"].(string)
	fileType, _ := input.Vars["file_type"].(string)
	if fileType == "" {
		fileType = "any"
	}

	age, _ := input.Vars["age"].(string)
	ageStamp, _ := input.Vars["age_stamp"].(string)
	if ageStamp == "" {
		ageStamp = "mtime"
	}

	size, _ := input.Vars["size"].(string)
	recurse, _ := input.Vars["recurse"].(bool)
	hidden, _ := input.Vars["hidden"].(bool)
	depth, _ := input.Vars["depth"].(string)

	// Build shell commands
	result := buildFindCommand(path, patterns, fileType, age, ageStamp, size, recurse, hidden, depth)
	printOutput(result)
}

func buildFindCommand(path, patterns, fileType, age, ageStamp, size string, recurse, hidden bool, depth string) TaskOutput {
	var shellExec []map[string]interface{}

	// Validate path exists
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": fmt.Sprintf("[ -d '%s' ] || (echo 'Directory not found: %s' >&2 && exit 1)", path, path),
	})

	// Build find command
	cmdParts := []string{"find", fmt.Sprintf("'%s'", path)}

	// Add maxdepth if not recursive or if depth specified
	if !recurse {
		cmdParts = append(cmdParts, "-maxdepth 1")
	} else if depth != "" {
		cmdParts = append(cmdParts, fmt.Sprintf("-maxdepth %s", depth))
	}

	// Add file type filter
	switch fileType {
	case "file":
		cmdParts = append(cmdParts, "-type f")
	case "directory":
		cmdParts = append(cmdParts, "-type d")
	case "link":
		cmdParts = append(cmdParts, "-type l")
	case "any":
		// No filter
	default:
		return TaskOutput{
			Status:  "failed",
			Message: fmt.Sprintf("Invalid file_type: %s (must be file, directory, link, or any)", fileType),
			Facts:   make(map[string]interface{}),
		}
	}

	// Add name pattern
	if patterns != "" {
		// Escape single quotes in pattern
		escapedPattern := strings.ReplaceAll(patterns, "'", "'\\''")
		cmdParts = append(cmdParts, fmt.Sprintf("-name '%s'", escapedPattern))
	}

	// Add hidden file filter
	if !hidden {
		// Exclude hidden files (but not the starting directory itself)
		cmdParts = append(cmdParts, "!", "-path", "'*/.*'")
	}

	// Add age filter
	if age != "" {
		switch ageStamp {
		case "mtime":
			cmdParts = append(cmdParts, fmt.Sprintf("-mtime %s", age))
		case "atime":
			cmdParts = append(cmdParts, fmt.Sprintf("-atime %s", age))
		case "ctime":
			cmdParts = append(cmdParts, fmt.Sprintf("-ctime %s", age))
		default:
			return TaskOutput{
				Status:  "failed",
				Message: fmt.Sprintf("Invalid age_stamp: %s (must be mtime, atime, or ctime)", ageStamp),
				Facts:   make(map[string]interface{}),
			}
		}
	}

	// Add size filter
	if size != "" {
		cmdParts = append(cmdParts, fmt.Sprintf("-size %s", size))
	}

	// Add print to show results
	cmdParts = append(cmdParts, "-print")

	// Join all parts
	findCmd := strings.Join(cmdParts, " ")

	// Execute find command
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": findCmd,
	})

	// Count results for summary
	countCmd := fmt.Sprintf("%s | wc -l", findCmd)
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": countCmd,
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
