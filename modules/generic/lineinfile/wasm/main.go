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

	line, _ := input.Vars["line"].(string)
	regexp, _ := input.Vars["regexp"].(string)
	state, _ := input.Vars["state"].(string)
	if state == "" {
		state = "present"
	}

	insertafter, _ := input.Vars["insertafter"].(string)
	insertbefore, _ := input.Vars["insertbefore"].(string)
	backup, _ := input.Vars["backup"].(bool)
	create, _ := input.Vars["create"].(bool)

	// Build shell commands based on parameters
	result := buildLineinfileCommands(path, line, regexp, state, insertafter, insertbefore, backup, create)
	printOutput(result)
}

func buildLineinfileCommands(path, line, regexp, state, insertafter, insertbefore string, backup, create bool) TaskOutput {
	var shellExec []map[string]interface{}

	// Validate parameters
	if state == "present" && line == "" {
		return TaskOutput{
			Status:  "failed",
			Message: "line parameter required when state=present",
			Facts:   make(map[string]interface{}),
		}
	}

	if state == "absent" && regexp == "" {
		return TaskOutput{
			Status:  "failed",
			Message: "regexp parameter required when state=absent",
			Facts:   make(map[string]interface{}),
		}
	}

	// Create file if requested and doesn't exist
	if create {
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("[ -f '%s' ] || touch '%s'", path, path),
		})
	}

	// Check if file exists
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": fmt.Sprintf("[ -f '%s' ] || (echo 'File not found: %s' >&2 && exit 1)", path, path),
	})

	// Create backup if requested
	if backup {
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("cp '%s' '%s.bak'", path, path),
		})
	}

	// Escape single quotes in line for shell safety
	escapedLine := strings.ReplaceAll(line, "'", "'\\''")

	switch state {
	case "present":
		if regexp != "" {
			// Replace line matching regexp, or add if not found
			escapedRegexp := strings.ReplaceAll(regexp, "'", "'\\''")
			shellExec = append(shellExec, buildReplaceCommand(path, escapedLine, escapedRegexp, insertafter, insertbefore))
		} else {
			// Just ensure line exists (add if not found)
			shellExec = append(shellExec, buildEnsureLineCommand(path, escapedLine, insertafter, insertbefore))
		}

	case "absent":
		// Remove lines matching regexp
		escapedRegexp := strings.ReplaceAll(regexp, "'", "'\\''")
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("sed -i.tmp '/%s/d' '%s' && rm -f '%s.tmp'", escapedRegexp, path, path),
		})
	}

	// Verify the result
	if state == "present" {
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("grep -F '%s' '%s' && echo 'Line present'", escapedLine, path),
		})
	} else {
		escapedRegexp := strings.ReplaceAll(regexp, "'", "'\\''")
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("! grep '%s' '%s' && echo 'Lines removed'", escapedRegexp, path),
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

func buildReplaceCommand(path, line, regexp, insertafter, insertbefore string) map[string]interface{} {
	// Try to replace matching line, or insert if not found
	var cmd string
	if insertafter != "" {
		// Replace if regexp matches, otherwise insert after insertafter pattern
		escapedInsertafter := strings.ReplaceAll(insertafter, "'", "'\\''")
		if insertafter == "EOF" {
			cmd = fmt.Sprintf(
				"grep -q '%s' '%s' && sed -i.tmp 's/%s/%s/' '%s' || echo '%s' >> '%s'; rm -f '%s.tmp'",
				regexp, path, regexp, line, path, line, path, path,
			)
		} else {
			cmd = fmt.Sprintf(
				"grep -q '%s' '%s' && sed -i.tmp 's/%s/%s/' '%s' || sed -i.tmp '/%s/a %s' '%s'; rm -f '%s.tmp'",
				regexp, path, regexp, line, path, escapedInsertafter, line, path, path,
			)
		}
	} else if insertbefore != "" {
		// Replace if regexp matches, otherwise insert before insertbefore pattern
		escapedInsertbefore := strings.ReplaceAll(insertbefore, "'", "'\\''")
		if insertbefore == "BOF" {
			cmd = fmt.Sprintf(
				"grep -q '%s' '%s' && sed -i.tmp 's/%s/%s/' '%s' || sed -i.tmp '1i %s' '%s'; rm -f '%s.tmp'",
				regexp, path, regexp, line, path, line, path, path,
			)
		} else {
			cmd = fmt.Sprintf(
				"grep -q '%s' '%s' && sed -i.tmp 's/%s/%s/' '%s' || sed -i.tmp '/%s/i %s' '%s'; rm -f '%s.tmp'",
				regexp, path, regexp, line, path, escapedInsertbefore, line, path, path,
			)
		}
	} else {
		// Just replace if found, append if not
		cmd = fmt.Sprintf(
			"grep -q '%s' '%s' && sed -i.tmp 's/%s/%s/' '%s' || echo '%s' >> '%s'; rm -f '%s.tmp'",
			regexp, path, regexp, line, path, line, path, path,
		)
	}

	return map[string]interface{}{
		"type":    "shell",
		"command": cmd,
	}
}

func buildEnsureLineCommand(path, line, insertafter, insertbefore string) map[string]interface{} {
	// Check if exact line exists, if not insert it
	var cmd string
	if insertafter != "" {
		escapedInsertafter := strings.ReplaceAll(insertafter, "'", "'\\''")
		if insertafter == "EOF" {
			cmd = fmt.Sprintf(
				"grep -qxF '%s' '%s' || echo '%s' >> '%s'",
				line, path, line, path,
			)
		} else {
			cmd = fmt.Sprintf(
				"grep -qxF '%s' '%s' || sed -i.tmp '/%s/a %s' '%s'; rm -f '%s.tmp'",
				line, path, escapedInsertafter, line, path, path,
			)
		}
	} else if insertbefore != "" {
		escapedInsertbefore := strings.ReplaceAll(insertbefore, "'", "'\\''")
		if insertbefore == "BOF" {
			cmd = fmt.Sprintf(
				"grep -qxF '%s' '%s' || sed -i.tmp '1i %s' '%s'; rm -f '%s.tmp'",
				line, path, line, path, path,
			)
		} else {
			cmd = fmt.Sprintf(
				"grep -qxF '%s' '%s' || sed -i.tmp '/%s/i %s' '%s'; rm -f '%s.tmp'",
				line, path, escapedInsertbefore, line, path, path,
			)
		}
	} else {
		// Append to end if not found
		cmd = fmt.Sprintf(
			"grep -qxF '%s' '%s' || echo '%s' >> '%s'",
			line, path, line, path,
		)
	}

	return map[string]interface{}{
		"type":    "shell",
		"command": cmd,
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
