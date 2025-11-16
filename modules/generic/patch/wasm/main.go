package main

import (
	"encoding/base64"
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
	src, _ := input.Vars["src"].(string)
	if src == "" {
		outputError("Missing required variable: src")
		return
	}

	// Extract optional parameters
	dest, _ := input.Vars["dest"].(string)
	if dest == "" {
		dest = "."
	}

	// Strip can be a float64 from JSON unmarshaling
	var strip int
	switch v := input.Vars["strip"].(type) {
	case float64:
		strip = int(v)
	case int:
		strip = v
	default:
		strip = 0
	}

	remoteSrc, _ := input.Vars["remote_src"].(bool)

	// Build patch command
	result := buildPatchCommand(src, dest, strip, remoteSrc, input.Vars)
	printOutput(result)
}

func buildPatchCommand(src, dest string, strip int, remoteSrc bool, vars map[string]interface{}) TaskOutput {
	var shellExec []map[string]interface{}

	var patchFile string

	if remoteSrc {
		// Patch file is already on the remote host
		patchFile = src
	} else {
		// Patch file needs to be uploaded from local machine
		// The executor should have set _patch_content with base64-encoded patch
		patchContentB64, ok := vars["_patch_content"].(string)
		if !ok || patchContentB64 == "" {
			return TaskOutput{
				Status:  "failed",
				Message: "Patch content not provided by executor (local patch file should be read and base64-encoded)",
				Facts:   make(map[string]interface{}),
			}
		}

		// Decode the patch content
		patchContent, err := base64.StdEncoding.DecodeString(patchContentB64)
		if err != nil {
			return TaskOutput{
				Status:  "failed",
				Message: fmt.Sprintf("Failed to decode patch content: %v", err),
				Facts:   make(map[string]interface{}),
			}
		}

		// Write patch to temporary file on remote
		patchFile = "/tmp/froyo-patch-$$.patch"
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "file_write",
			"path":    patchFile,
			"content": string(patchContent),
			"mode":    0644,
		})
	}

	// Apply the patch
	patchCmd := fmt.Sprintf("patch -p%d -d '%s' < '%s'", strip, dest, patchFile)
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": patchCmd,
	})

	// Clean up temporary patch file if we created one
	if !remoteSrc {
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("rm -f '%s'", patchFile),
		})
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
