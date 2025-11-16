package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

	remoteSrc := false
	if val, ok := input.Vars["remote_src"].(bool); ok {
		remoteSrc = val
	}

	creates, _ := input.Vars["creates"].(string)

	listFiles := false
	if val, ok := input.Vars["list_files"].(bool); ok {
		listFiles = val
	}

	// Check for uploaded archive content (when remote_src=false)
	archiveContent, _ := input.Vars["_archive_content"].(string)

	// Build shell commands for extraction
	result := buildUnarchiveCommands(src, dest, remoteSrc, creates, listFiles, archiveContent)
	printOutput(result)
}

func buildUnarchiveCommands(src, dest string, remoteSrc bool, creates string, listFiles bool, archiveContent string) TaskOutput {
	var shellExec []map[string]interface{}

	// Check if creates path exists (idempotency)
	if creates != "" {
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("if [ -e '%s' ]; then echo 'File %s already exists, skipping extraction'; exit 0; fi", creates, creates),
		})
	}

	// Determine the actual source path on the remote host
	actualSrc := src
	if !remoteSrc && archiveContent != "" {
		// Archive was uploaded from local machine, create temp file
		actualSrc = fmt.Sprintf("/tmp/openfroyo-archive-%s", filepath.Base(src))
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("echo '%s' | base64 -d > '%s'", archiveContent, actualSrc),
		})
	}

	// Verify source exists
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": fmt.Sprintf("[ -f '%s' ] || { echo 'Archive not found: %s' >&2; exit 1; }", actualSrc, actualSrc),
	})

	// Create destination directory
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": fmt.Sprintf("mkdir -p '%s'", dest),
	})

	// Detect archive format from extension
	format := detectFormat(src)

	// List files if requested
	if listFiles {
		listCmd := buildListCommand(actualSrc, format)
		if listCmd != "" {
			shellExec = append(shellExec, map[string]interface{}{
				"type":    "shell",
				"command": listCmd,
			})
		}
	}

	// Extract archive
	extractCmd := buildExtractCommand(actualSrc, dest, format)
	if extractCmd == "" {
		return TaskOutput{
			Status:  "failed",
			Message: fmt.Sprintf("Unsupported archive format: %s", src),
			Facts:   make(map[string]interface{}),
		}
	}

	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": extractCmd,
	})

	// Verify extraction
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": fmt.Sprintf("ls -la '%s' | head -20", dest),
	})

	// Cleanup temp file if we created one
	if !remoteSrc && archiveContent != "" {
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("rm -f '%s'", actualSrc),
		})
	}

	return TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Extracting %s to %s", src, dest),
		Facts: map[string]interface{}{
			"shell_exec":      shellExec,
			"archive_format":  format,
			"extraction_dest": dest,
		},
	}
}

func detectFormat(filename string) string {
	lower := strings.ToLower(filename)
	if strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz") {
		return "gz"
	} else if strings.HasSuffix(lower, ".tar.bz2") || strings.HasSuffix(lower, ".tbz2") {
		return "bz2"
	} else if strings.HasSuffix(lower, ".tar.xz") || strings.HasSuffix(lower, ".txz") {
		return "xz"
	} else if strings.HasSuffix(lower, ".zip") {
		return "zip"
	} else if strings.HasSuffix(lower, ".tar") {
		return "tar"
	}
	return "unknown"
}

func buildListCommand(src, format string) string {
	switch format {
	case "gz", "bz2", "xz", "tar":
		return fmt.Sprintf("tar -tf '%s' | head -20", src)
	case "zip":
		return fmt.Sprintf("unzip -l '%s' | head -20", src)
	default:
		return ""
	}
}

func buildExtractCommand(src, dest, format string) string {
	switch format {
	case "gz":
		return fmt.Sprintf("tar -xzf '%s' -C '%s'", src, dest)
	case "bz2":
		return fmt.Sprintf("tar -xjf '%s' -C '%s'", src, dest)
	case "xz":
		return fmt.Sprintf("tar -xJf '%s' -C '%s'", src, dest)
	case "tar":
		return fmt.Sprintf("tar -xf '%s' -C '%s'", src, dest)
	case "zip":
		return fmt.Sprintf("unzip -q '%s' -d '%s'", src, dest)
	default:
		return ""
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
