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

	delimiter, _ := input.Vars["delimiter"].(string)
	regexp, _ := input.Vars["regexp"].(string)
	ignoreHidden, ok := input.Vars["ignore_hidden"].(bool)
	if !ok {
		ignoreHidden = true
	}

	// Build shell commands
	result := buildAssembleCommands(src, dest, delimiter, regexp, ignoreHidden)
	printOutput(result)
}

func buildAssembleCommands(src, dest, delimiter, regexp string, ignoreHidden bool) TaskOutput {
	var shellExec []map[string]interface{}

	// Build the find command based on parameters
	var findCmd string

	if regexp != "" {
		// Use regex to filter files
		if ignoreHidden {
			findCmd = fmt.Sprintf("find '%s' -type f ! -name '.*' -regex '%s' | sort", src, regexp)
		} else {
			findCmd = fmt.Sprintf("find '%s' -type f -regex '%s' | sort", src, regexp)
		}
	} else {
		// No regex filtering
		if ignoreHidden {
			findCmd = fmt.Sprintf("find '%s' -type f ! -name '.*' | sort", src)
		} else {
			findCmd = fmt.Sprintf("find '%s' -type f | sort", src)
		}
	}

	// Build the assembly command
	var assembleCmd string
	if delimiter != "" {
		// With delimiter: need to add delimiter between files
		// This is more complex - we'll use a shell loop
		assembleCmd = fmt.Sprintf(
			"{ first=true; for f in $(%s); do if [ \"$first\" = true ]; then first=false; else echo '%s'; fi; cat \"$f\"; done; } > '%s'",
			findCmd, delimiter, dest,
		)
	} else {
		// Without delimiter: simple concatenation
		assembleCmd = fmt.Sprintf("%s | xargs cat > '%s'", findCmd, dest)
	}

	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": assembleCmd,
	})

	// Verify the destination file was created
	shellExec = append(shellExec, map[string]interface{}{
		"type":    "shell",
		"command": fmt.Sprintf("[ -f '%s' ] && wc -l '%s'", dest, dest),
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
