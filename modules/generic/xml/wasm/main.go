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

	xpath, _ := input.Vars["xpath"].(string)
	if xpath == "" {
		outputError("Missing required variable: xpath")
		return
	}

	state, _ := input.Vars["state"].(string)
	if state == "" {
		state = "present"
	}

	value, _ := input.Vars["value"].(string)
	attribute, _ := input.Vars["attribute"].(string)
	prettyPrint, _ := input.Vars["pretty_print"].(bool)
	count, _ := input.Vars["count"].(bool)

	// Build shell commands based on operation
	result := buildXMLCommands(path, xpath, value, attribute, state, prettyPrint, count)
	printOutput(result)
}

func buildXMLCommands(path, xpath, value, attribute, state string, prettyPrint, count bool) TaskOutput {
	var shellExec []map[string]interface{}

	// Check if xmlstarlet is available
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": "command -v xmlstarlet >/dev/null 2>&1 || echo 'xmlstarlet not found'",
	})

	// Handle count operation
	if count {
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("xmlstarlet sel -t -v \"count(%s)\" '%s'", escapeXPath(xpath), path),
		})
		return TaskOutput{
			Status:  "ok",
			Message: "",
			Facts: map[string]interface{}{
				"shell_exec": shellExec,
			},
		}
	}

	switch state {
	case "present":
		// Modify or insert XML content
		if attribute != "" {
			// Set attribute value
			if value == "" {
				return TaskOutput{
					Status:  "failed",
					Message: "value parameter required when setting attribute",
					Facts:   make(map[string]interface{}),
				}
			}

			cmd := fmt.Sprintf(
				"xmlstarlet ed -L -u \"%s/@%s\" -v \"%s\" '%s'",
				escapeXPath(xpath), attribute, escapeValue(value), path,
			)

			if prettyPrint {
				cmd = cmd + fmt.Sprintf(" && xmlstarlet fo -s 2 '%s' > '%s.tmp' && mv '%s.tmp' '%s'",
					path, path, path, path)
			}

			shellExec = append(shellExec, map[string]interface{}{
				"type":    "shell",
				"command": cmd,
			})

		} else if value != "" {
			// Set element value
			cmd := fmt.Sprintf(
				"xmlstarlet ed -L -u \"%s\" -v \"%s\" '%s'",
				escapeXPath(xpath), escapeValue(value), path,
			)

			if prettyPrint {
				cmd = cmd + fmt.Sprintf(" && xmlstarlet fo -s 2 '%s' > '%s.tmp' && mv '%s.tmp' '%s'",
					path, path, path, path)
			}

			shellExec = append(shellExec, map[string]interface{}{
				"type":    "shell",
				"command": cmd,
			})

		} else {
			return TaskOutput{
				Status:  "failed",
				Message: "value parameter required for state=present",
				Facts:   make(map[string]interface{}),
			}
		}

	case "absent":
		// Delete XML nodes
		cmd := fmt.Sprintf(
			"xmlstarlet ed -L -d \"%s\" '%s'",
			escapeXPath(xpath), path,
		)

		if prettyPrint {
			cmd = cmd + fmt.Sprintf(" && xmlstarlet fo -s 2 '%s' > '%s.tmp' && mv '%s.tmp' '%s'",
				path, path, path, path)
		}

		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": cmd,
		})

	default:
		return TaskOutput{
			Status:  "failed",
			Message: fmt.Sprintf("Invalid state: %s (must be present or absent)", state),
			Facts:   make(map[string]interface{}),
		}
	}

	// Add verification command to show the modified section
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": fmt.Sprintf("xmlstarlet sel -t -c \"%s\" '%s' 2>/dev/null || echo 'no matches'", escapeXPath(xpath), path),
	})

	return TaskOutput{
		Status:  "ok",
		Message: "",
		Facts: map[string]interface{}{
			"shell_exec": shellExec,
		},
	}
}

// escapeXPath escapes special characters in XPath expressions
func escapeXPath(xpath string) string {
	// Basic escaping - in production, more sophisticated escaping may be needed
	return strings.ReplaceAll(xpath, "\"", "\\\"")
}

// escapeValue escapes special characters in values
func escapeValue(value string) string {
	// Escape double quotes and backslashes
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\"", "\\\"")
	return value
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
