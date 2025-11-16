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

	section, _ := input.Vars["section"].(string)
	if section == "" {
		outputError("Missing required variable: section")
		return
	}

	// Extract optional parameters
	option, _ := input.Vars["option"].(string)
	value, _ := input.Vars["value"].(string)
	state, _ := input.Vars["state"].(string)
	if state == "" {
		state = "present"
	}

	noExtraSpaces, _ := input.Vars["no_extra_spaces"].(bool)
	create, _ := input.Vars["create"].(bool)

	// Build INI file manipulation commands
	result := buildIniFileCommands(path, section, option, value, state, noExtraSpaces, create)
	printOutput(result)
}

func buildIniFileCommands(path, section, option, value, state string, noExtraSpaces, create bool) TaskOutput {
	var shellExec []map[string]interface{}

	// Create file if needed
	if create {
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("[ -f '%s' ] || touch '%s'", path, path),
		})
	}

	// Try to use crudini if available, otherwise fall back to sed/awk
	if state == "present" {
		if option != "" && value != "" {
			// Set an option value
			delimiter := " = "
			if noExtraSpaces {
				delimiter = "="
			}

			// First try crudini
			crudiniCmd := fmt.Sprintf("crudini --set '%s' '%s' '%s' '%s' 2>/dev/null", path, section, option, value)

			// Fallback to manual manipulation using awk
			fallbackScript := buildSetOptionScript(path, section, option, value, delimiter)

			shellExec = append(shellExec, map[string]interface{}{
				"type":    "shell",
				"command": fmt.Sprintf("%s || { %s; }", crudiniCmd, fallbackScript),
			})
		} else if option == "" {
			// Just ensure section exists
			crudiniCmd := fmt.Sprintf("crudini --set '%s' '%s' 2>/dev/null", path, section)
			fallbackScript := buildEnsureSectionScript(path, section)

			shellExec = append(shellExec, map[string]interface{}{
				"type":    "shell",
				"command": fmt.Sprintf("%s || { %s; }", crudiniCmd, fallbackScript),
			})
		}
	} else if state == "absent" {
		if option != "" {
			// Remove an option
			crudiniCmd := fmt.Sprintf("crudini --del '%s' '%s' '%s' 2>/dev/null", path, section, option)
			fallbackScript := fmt.Sprintf("sed -i '/^\\[%s\\]/,/^\\[/{/^%s[[:space:]]*=/d}' '%s'",
				escapeForSed(section), escapeForSed(option), path)

			shellExec = append(shellExec, map[string]interface{}{
				"type":    "shell",
				"command": fmt.Sprintf("%s || { %s; }", crudiniCmd, fallbackScript),
			})
		} else {
			// Remove entire section
			crudiniCmd := fmt.Sprintf("crudini --del '%s' '%s' 2>/dev/null", path, section)
			fallbackScript := fmt.Sprintf("sed -i '/^\\[%s\\]/,/^\\[/d' '%s'", escapeForSed(section), path)

			shellExec = append(shellExec, map[string]interface{}{
				"type":    "shell",
				"command": fmt.Sprintf("%s || { %s; }", crudiniCmd, fallbackScript),
			})
		}
	}

	// Verify the result
	if option != "" {
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("grep -A 10 '\\[%s\\]' '%s' | head -n 15", section, path),
		})
	} else {
		shellExec = append(shellExec, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("grep '\\[%s\\]' '%s'", section, path),
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

func buildSetOptionScript(path, section, option, value, delimiter string) string {
	// This awk script ensures section exists and sets the option
	script := fmt.Sprintf(`
awk -v section="[%s]" -v option="%s" -v value="%s" -v delim="%s" '
BEGIN { found_section=0; found_option=0; }
/^\[.*\]/ {
  if (found_section && !found_option) {
    print option delim value
    found_option=1
  }
  if ($0 == section) found_section=1
  print; next
}
found_section && $0 ~ "^" option "[[:space:]]*=" {
  print option delim value
  found_option=1
  next
}
{ print }
END {
  if (!found_section) {
    print section
    print option delim value
  } else if (!found_option) {
    print option delim value
  }
}' '%s' > '%s.tmp' && mv '%s.tmp' '%s'
`, section, option, value, delimiter, path, path, path, path)

	return strings.TrimSpace(script)
}

func buildEnsureSectionScript(path, section string) string {
	script := fmt.Sprintf(`grep -q '^\[%s\]' '%s' || echo -e '\n[%s]' >> '%s'`,
		section, path, section, path)
	return script
}

func escapeForSed(s string) string {
	// Escape special sed characters
	s = strings.ReplaceAll(s, "/", "\\/")
	s = strings.ReplaceAll(s, ".", "\\.")
	s = strings.ReplaceAll(s, "[", "\\[")
	s = strings.ReplaceAll(s, "]", "\\]")
	return s
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
