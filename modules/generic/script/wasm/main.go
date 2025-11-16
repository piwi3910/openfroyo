package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// TaskInput represents the input from OpenFroyo
type TaskInput struct {
	Vars    map[string]interface{} `json:"vars"`
	Context TaskContext            `json:"context"`
}

// TaskContext provides execution context
type TaskContext struct {
	Host     string `json:"host"`
	TaskName string `json:"task_name"`
}

// TaskOutput represents the output to OpenFroyo
type TaskOutput struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message"`
	Facts   map[string]interface{} `json:"facts"`
}

func main() {
	var input TaskInput

	// Read input from stdin
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&input); err != nil {
		output := TaskOutput{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to decode input: %v", err),
			Facts:   make(map[string]interface{}),
		}
		printOutput(output)
		return
	}

	// Get script content (base64 encoded) from executor
	scriptContentB64, ok := input.Vars["_script_content"].(string)
	if !ok || scriptContentB64 == "" {
		output := TaskOutput{
			Status:  "failed",
			Message: "No script content provided (_script_content variable missing)",
			Facts:   make(map[string]interface{}),
		}
		printOutput(output)
		return
	}

	// Decode script content
	scriptContent, err := base64.StdEncoding.DecodeString(scriptContentB64)
	if err != nil {
		output := TaskOutput{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to decode script content: %v", err),
			Facts:   make(map[string]interface{}),
		}
		printOutput(output)
		return
	}

	// Get script path for filename
	scriptPath, ok := input.Vars["script"].(string)
	if !ok || scriptPath == "" {
		scriptPath = "script.sh"
	}

	args, _ := input.Vars["args"].(string)
	scriptName := filepath.Base(scriptPath)
	remotePath := fmt.Sprintf("/tmp/froyo/scripts/%s", scriptName)

	// Build the execution command
	execCmd := remotePath
	if args != "" {
		execCmd = fmt.Sprintf("%s %s", remotePath, args)
	}

	// Return shell_exec fact to:
	// 1. Write the script file
	// 2. Execute it
	result := TaskOutput{
		Status:  "ok",
		Message: "", // Will be set by the handler
		Facts: map[string]interface{}{
			"shell_exec": []map[string]interface{}{
				{
					"type":    "file_write",
					"path":    remotePath,
					"content": string(scriptContent),
					"mode":    0755, // executable
				},
				{
					"type":    "shell",
					"command": execCmd,
				},
			},
		},
	}

	printOutput(result)
}

func printOutput(output TaskOutput) {
	outputJSON, _ := json.MarshalIndent(output, "", "  ")
	fmt.Println(string(outputJSON))
}
