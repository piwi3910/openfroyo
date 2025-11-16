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
	mode, _ := input.Vars["mode"].(string)
	owner, _ := input.Vars["owner"].(string)
	group, _ := input.Vars["group"].(string)
	backup, _ := input.Vars["backup"].(bool)
	validate, _ := input.Vars["validate"].(string)

	// Get template content (will be provided by executor after reading local file)
	templateContent, _ := input.Vars["_template_content"].(string)
	if templateContent == "" {
		outputError("Template content not provided by executor (_template_content)")
		return
	}

	// Perform variable substitution
	renderedContent := renderTemplate(templateContent, input.Vars)

	// Build shell commands for template operations
	result := buildTemplateCommands(dest, renderedContent, mode, owner, group, backup, validate)
	printOutput(result)
}

func renderTemplate(template string, vars map[string]interface{}) string {
	result := template

	// Simple variable substitution: {{var_name}} or {{ var_name }}
	// Find all {{...}} patterns and replace with corresponding var values
	for key, value := range vars {
		// Skip internal variables
		if strings.HasPrefix(key, "_") {
			continue
		}

		// Convert value to string
		var strValue string
		switch v := value.(type) {
		case string:
			strValue = v
		case float64:
			strValue = fmt.Sprintf("%v", v)
		case int:
			strValue = fmt.Sprintf("%d", v)
		case bool:
			strValue = fmt.Sprintf("%t", v)
		default:
			strValue = fmt.Sprintf("%v", v)
		}

		// Replace all variants of the variable reference
		result = strings.ReplaceAll(result, "{{"+key+"}}", strValue)
		result = strings.ReplaceAll(result, "{{ "+key+" }}", strValue)
		result = strings.ReplaceAll(result, "{{  "+key+"  }}", strValue)
	}

	return result
}

func buildTemplateCommands(dest, content, mode, owner, group string, backup bool, validate string) TaskOutput {
	var shellExec []map[string]interface{}

	// Create backup if requested and file exists
	if backup {
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("[ -f '%s' ] && cp -p '%s' '%s.bak.$(date +%%Y%%m%%d-%%H%%M%%S)' || true", dest, dest, dest),
		})
	}

	// Write template content to temporary file using heredoc
	tmpFile := fmt.Sprintf("%s.tmp.$$", dest)
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": fmt.Sprintf("cat > '%s' <<'FROYO_TEMPLATE_EOF'\n%s\nFROYO_TEMPLATE_EOF", tmpFile, content),
	})

	// Validate file if validation command provided
	if validate != "" {
		validateCmd := strings.Replace(validate, "%s", tmpFile, -1)
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("%s || { rm -f '%s'; echo 'Validation failed' >&2; exit 1; }", validateCmd, tmpFile),
		})
	}

	// Move temporary file to destination
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": fmt.Sprintf("mv '%s' '%s'", tmpFile, dest),
	})

	// Set permissions if specified
	if mode != "" {
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("chmod %s '%s'", mode, dest),
		})
	}

	// Set ownership if specified
	if owner != "" || group != "" {
		var chownTarget string
		if owner != "" && group != "" {
			chownTarget = fmt.Sprintf("%s:%s", owner, group)
		} else if owner != "" {
			chownTarget = owner
		} else {
			chownTarget = fmt.Sprintf(":%s", group)
		}

		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("chown %s '%s'", chownTarget, dest),
		})
	}

	// Verify file exists and show info
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": fmt.Sprintf("[ -f '%s' ] && ls -l '%s' || { echo 'Template deployment failed' >&2; exit 1; }", dest, dest),
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
