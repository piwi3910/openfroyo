package main

import (
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
	path, _ := input.Vars["path"].(string)
	if path == "" {
		outputError("Missing required variable: path")
		return
	}

	// Build shell commands to stat the file
	result := buildStatCommands(path)
	printOutput(result)
}

func buildStatCommands(path string) TaskOutput {
	var shellExec []map[string]interface{}

	// Use stat command with format strings to get detailed info
	// We use stat with fallback to ls -la for portability
	// The command will check if path exists and get its details
	statCmd := fmt.Sprintf(`
if [ -e '%s' ]; then
  if command -v stat >/dev/null 2>&1; then
    # Try GNU stat format
    if stat -c '%%s|%%f|%%U|%%G|%%Y|%%X' '%s' 2>/dev/null; then
      :
    # Try BSD stat format
    elif stat -f '%%z|%%p|%%Su|%%Sg|%%m|%%a' '%s' 2>/dev/null; then
      :
    else
      # Fallback to ls
      ls -la '%s'
    fi
  else
    # No stat command, use ls
    ls -la '%s'
  fi
  echo "EXISTS=true"
else
  echo "EXISTS=false"
fi
`, path, path, path, path, path)

	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": statCmd,
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
