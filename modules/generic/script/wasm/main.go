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

	// Get script content (base64 encoded) and args
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

	args, _ := input.Vars["args"].(string)

	// Get original script name for better debugging
	scriptName, ok := input.Vars["script"].(string)
	if !ok || scriptName == "" {
		scriptName = "script.sh"
	} else {
		scriptName = filepath.Base(scriptName)
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

	// Create remote script path
	remotePath := fmt.Sprintf("/tmp/froyo/scripts/%s", scriptName)

	// Return facts to tell froyo-runner to:
	// 1. Create the script file
	// 2. Make it executable
	// 3. Execute it
	result := TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Script prepared: %s", scriptName),
		Facts: map[string]interface{}{
			"script_path":    remotePath,
			"script_content": string(scriptContent),
			"script_args":    args,
			"script_command": buildScriptCommand(remotePath, args),
		},
	}

	printOutput(result)
}

func buildScriptCommand(scriptPath, args string) string {
	if args != "" {
		return fmt.Sprintf("%s %s", scriptPath, args)
	}
	return scriptPath
}

func printOutput(output TaskOutput) {
	outputJSON, _ := json.MarshalIndent(output, "", "  ")
	fmt.Println(string(outputJSON))
}
