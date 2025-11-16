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

	state, _ := input.Vars["state"].(string)
	if state == "" {
		state = "present"
	}

	block, _ := input.Vars["block"].(string)
	if state == "present" && block == "" {
		outputError("Missing required variable: block (required when state=present)")
		return
	}

	marker, _ := input.Vars["marker"].(string)
	markerBegin, _ := input.Vars["marker_begin"].(string)
	markerEnd, _ := input.Vars["marker_end"].(string)
	insertafter, _ := input.Vars["insertafter"].(string)
	insertbefore, _ := input.Vars["insertbefore"].(string)
	create, _ := input.Vars["create"].(bool)
	backup, _ := input.Vars["backup"].(bool)

	// Build shell commands for blockinfile operations
	result := buildBlockInFileCommands(path, block, state, marker, markerBegin, markerEnd,
		insertafter, insertbefore, create, backup)
	printOutput(result)
}

func buildBlockInFileCommands(path, block, state, marker, markerBegin, markerEnd,
	insertafter, insertbefore string, create, backup bool) TaskOutput {

	var shellExec []map[string]interface{}

	// Set default markers if not specified
	if markerBegin == "" && markerEnd == "" {
		if marker != "" {
			// Custom marker template with {mark} placeholder
			markerBegin = strings.Replace(marker, "{mark}", "BEGIN", 1)
			markerEnd = strings.Replace(marker, "{mark}", "END", 1)
		} else {
			// Default markers
			markerBegin = "# BEGIN MANAGED BLOCK"
			markerEnd = "# END MANAGED BLOCK"
		}
	}

	// Create file if it doesn't exist and create=true
	if create {
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("[ -f '%s' ] || touch '%s'", path, path),
		})
	}

	// Check if file exists
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": fmt.Sprintf("[ -f '%s' ] || { echo 'File does not exist: %s' >&2; exit 1; }", path, path),
	})

	// Create backup if requested
	if backup {
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("cp -p '%s' '%s.bak.$(date +%%Y%%m%%d-%%H%%M%%S)'", path, path),
		})
	}

	if state == "present" {
		// Escape single quotes in block content and markers for shell
		escapedBlock := strings.ReplaceAll(block, "'", "'\\''")
		escapedMarkerBegin := strings.ReplaceAll(markerBegin, "'", "'\\''")
		escapedMarkerEnd := strings.ReplaceAll(markerEnd, "'", "'\\''")

		// Build awk script to insert/update block
		awkScript := buildAwkInsertScript(escapedMarkerBegin, escapedMarkerEnd, escapedBlock,
			insertafter, insertbefore)

		// Execute awk script to modify file
		shellExec = append(shellExec, map[string]interface{}{
			"type": "shell",
			"command": fmt.Sprintf(
				"awk '%s' '%s' > '%s.tmp' && mv '%s.tmp' '%s'",
				awkScript, path, path, path, path,
			),
		})

	} else if state == "absent" {
		// Remove block between markers
		escapedMarkerBegin := strings.ReplaceAll(markerBegin, "'", "'\\''")
		escapedMarkerEnd := strings.ReplaceAll(markerEnd, "'", "'\\''")

		awkScript := fmt.Sprintf(
			"BEGIN { in_block=0 } "+
				"/%s/ { in_block=1; next } "+
				"/%s/ { in_block=0; next } "+
				"!in_block { print }",
			escapedMarkerBegin, escapedMarkerEnd,
		)

		shellExec = append(shellExec, map[string]interface{}{
			"type": "shell",
			"command": fmt.Sprintf(
				"awk '%s' '%s' > '%s.tmp' && mv '%s.tmp' '%s'",
				awkScript, path, path, path, path,
			),
		})
	}

	// Verify file exists after operation
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": fmt.Sprintf("[ -f '%s' ] && echo 'Block operation completed' || echo 'Failed'", path),
	})

	return TaskOutput{
		Status:  "ok",
		Message: "",
		Facts: map[string]interface{}{
			"shell_exec": shellExec,
		},
	}
}

func buildAwkInsertScript(markerBegin, markerEnd, block, insertafter, insertbefore string) string {
	// Build the block content with markers
	blockLines := strings.Split(block, "\n")
	var formattedBlock []string
	formattedBlock = append(formattedBlock, markerBegin)
	formattedBlock = append(formattedBlock, blockLines...)
	formattedBlock = append(formattedBlock, markerEnd)

	// Escape block content for awk print statements
	var printStatements []string
	for _, line := range formattedBlock {
		escaped := strings.ReplaceAll(line, "\"", "\\\"")
		printStatements = append(printStatements, fmt.Sprintf("print \"%s\"", escaped))
	}
	blockPrint := strings.Join(printStatements, "; ")

	var awkScript string

	if insertafter != "" {
		// Insert after pattern
		escapedPattern := strings.ReplaceAll(insertafter, "'", "'\\''")
		awkScript = fmt.Sprintf(
			"BEGIN { inserted=0; in_block=0 } "+
				"/%s/ { in_block=1; next } "+
				"/%s/ { in_block=0; next } "+
				"!in_block { print } "+
				"!in_block && !inserted && /%s/ { print; %s; inserted=1; next } "+
				"END { if (!inserted) { %s } }",
			markerBegin, markerEnd, escapedPattern, blockPrint, blockPrint,
		)
	} else if insertbefore != "" {
		// Insert before pattern
		escapedPattern := strings.ReplaceAll(insertbefore, "'", "'\\''")
		awkScript = fmt.Sprintf(
			"BEGIN { inserted=0; in_block=0 } "+
				"/%s/ { in_block=1; next } "+
				"/%s/ { in_block=0; next } "+
				"!in_block && !inserted && /%s/ { %s; inserted=1 } "+
				"!in_block { print }",
			markerBegin, markerEnd, escapedPattern, blockPrint,
		)
	} else {
		// Default: append to end if markers not found, or update existing block
		awkScript = fmt.Sprintf(
			"BEGIN { in_block=0; found=0 } "+
				"/%s/ { in_block=1; found=1; next } "+
				"/%s/ { in_block=0; %s; next } "+
				"!in_block { print } "+
				"END { if (!found) { %s } }",
			markerBegin, markerEnd, blockPrint, blockPrint,
		)
	}

	return awkScript
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
