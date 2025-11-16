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

	// Extract optional parameters
	deleteParm, _ := input.Vars["delete"].(bool)
	recursive, _ := input.Vars["recursive"].(bool)
	archive, _ := input.Vars["archive"].(bool)
	checksum, _ := input.Vars["checksum"].(bool)

	// Build rsync command
	result := buildRsyncCommand(src, dest, deleteParm, recursive, archive, checksum)
	printOutput(result)
}

func buildRsyncCommand(src, dest string, deleteParm, recursive, archive, checksum bool) TaskOutput {
	var shellExec []map[string]interface{}

	// Build rsync command with flags
	var flags []string

	// Archive mode includes -rlptgoD (recursive, links, perms, times, group, owner, devices)
	if archive {
		flags = append(flags, "-a")
	} else if recursive {
		flags = append(flags, "-r")
	}

	// Verbose and compress
	flags = append(flags, "-v", "-z")

	// Delete extraneous files
	if deleteParm {
		flags = append(flags, "--delete")
	}

	// Use checksums
	if checksum {
		flags = append(flags, "-c")
	}

	// Build the command
	flagStr := strings.Join(flags, " ")
	rsyncCmd := fmt.Sprintf("rsync %s '%s' '%s'", flagStr, src, dest)

	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": rsyncCmd,
	})

	// Add verification command
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": fmt.Sprintf("ls -la '%s'", dest),
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
