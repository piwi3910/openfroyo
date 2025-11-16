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
	image, _ := input.Vars["image"].(string)
	if image == "" {
		outputError("Missing required variable: image")
		return
	}

	dest, _ := input.Vars["dest"].(string)
	if dest == "" {
		outputError("Missing required variable: dest")
		return
	}

	creates, _ := input.Vars["creates"].(string)

	// Extract files list if provided
	var files []string
	if filesRaw, ok := input.Vars["files"].([]interface{}); ok {
		for _, f := range filesRaw {
			if fileStr, ok := f.(string); ok {
				files = append(files, fileStr)
			}
		}
	}

	// Build shell commands to extract ISO
	result := buildISOExtractCommands(image, dest, files, creates)
	printOutput(result)
}

func buildISOExtractCommands(image, dest string, files []string, creates string) TaskOutput {
	var shellExec []map[string]interface{}

	// If creates parameter is set, check if file exists (idempotency)
	if creates != "" {
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("[ -e '%s' ] && echo 'SKIP: Already extracted' && exit 0 || true", creates),
		})
	}

	// Check if ISO image exists
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": fmt.Sprintf("[ -f '%s' ] || (echo 'ERROR: ISO image not found: %s' && exit 1)", image, image),
	})

	// Create destination directory
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": fmt.Sprintf("mkdir -p '%s'", dest),
	})

	// Try multiple extraction methods in order of preference
	// Method 1: Try 7z (most reliable and widely available)
	// Method 2: Try bsdtar (available on macOS and BSD systems)
	// Method 3: Try mount loop (Linux only, requires root)

	extractCmd := buildExtractionCommand(image, dest, files)

	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": extractCmd,
	})

	// Verify extraction succeeded
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": fmt.Sprintf("[ -d '%s' ] && [ \"$(ls -A '%s')\" ] && echo 'Extraction complete' || (echo 'ERROR: Extraction failed' && exit 1)", dest, dest),
	})

	return TaskOutput{
		Status:  "ok",
		Message: "",
		Facts: map[string]interface{}{
			"shell_exec": shellExec,
		},
	}
}

func buildExtractionCommand(image, dest string, files []string) string {
	// Build a robust extraction command that tries multiple methods

	filePattern := ""
	if len(files) > 0 {
		// If specific files requested, build pattern for extraction
		filePattern = strings.Join(files, " ")
	}

	// Create a command that tries multiple extraction tools
	var cmdParts []string

	// Try 7z first (works on Linux, macOS, Windows with 7zip installed)
	if filePattern != "" {
		cmdParts = append(cmdParts, fmt.Sprintf("7z x '%s' -o'%s' %s 2>/dev/null", image, dest, filePattern))
	} else {
		cmdParts = append(cmdParts, fmt.Sprintf("7z x '%s' -o'%s' 2>/dev/null", image, dest))
	}

	// Try bsdtar (available on macOS and BSD)
	if filePattern != "" {
		// bsdtar with specific files
		cmdParts = append(cmdParts, fmt.Sprintf("bsdtar -xf '%s' -C '%s' %s 2>/dev/null", image, dest, filePattern))
	} else {
		cmdParts = append(cmdParts, fmt.Sprintf("bsdtar -xf '%s' -C '%s' 2>/dev/null", image, dest))
	}

	// Try mount method (Linux, requires privileges)
	mountCmd := fmt.Sprintf("TMPDIR=$(mktemp -d) && mount -o loop '%s' \"$TMPDIR\" 2>/dev/null && cp -r \"$TMPDIR\"/* '%s'/ && umount \"$TMPDIR\" && rmdir \"$TMPDIR\" 2>/dev/null", image, dest)
	cmdParts = append(cmdParts, mountCmd)

	// Join with || to try each method in sequence
	return strings.Join(cmdParts, " || ")
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
