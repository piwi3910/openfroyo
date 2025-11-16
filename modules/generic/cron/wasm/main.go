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
	name, _ := input.Vars["name"].(string)
	if name == "" {
		outputError("Missing required variable: name (job name)")
		return
	}

	job, _ := input.Vars["job"].(string)
	state, _ := input.Vars["state"].(string)
	if state == "" {
		state = "present"
	}

	user, _ := input.Vars["user"].(string)
	if user == "" {
		user = "root"
	}

	minute, _ := input.Vars["minute"].(string)
	if minute == "" {
		minute = "*"
	}

	hour, _ := input.Vars["hour"].(string)
	if hour == "" {
		hour = "*"
	}

	day, _ := input.Vars["day"].(string)
	if day == "" {
		day = "*"
	}

	month, _ := input.Vars["month"].(string)
	if month == "" {
		month = "*"
	}

	weekday, _ := input.Vars["weekday"].(string)
	if weekday == "" {
		weekday = "*"
	}

	specialTime, _ := input.Vars["special_time"].(string)
	disabled, _ := input.Vars["disabled"].(bool)

	// Validate job command is provided for present state
	if state == "present" && job == "" {
		outputError("Missing required variable: job (command to execute)")
		return
	}

	// Build the cron command
	result := buildCronCommand(name, job, state, user, minute, hour, day, month, weekday, specialTime, disabled)
	printOutput(result)
}

func buildCronCommand(name, job, state, user, minute, hour, day, month, weekday, specialTime string, disabled bool) TaskOutput {
	// Escape double quotes in name and job
	escapedName := strings.ReplaceAll(name, `"`, `\"`)
	escapedJob := strings.ReplaceAll(job, `"`, `\"`)

	script := `#!/bin/bash
set -e

USER="` + user + `"
JOB_NAME="` + escapedName + `"
JOB_CMD="` + escapedJob + `"
STATE="` + state + `"
`

	// Check if cron is available
	script += `
# Check if cron is available
if ! command -v crontab >/dev/null 2>&1; then
    echo "cron/crontab is not available on this system"
    exit 1
fi

echo "Managing cron job '$JOB_NAME' for user '$USER'"
`

	if state == "present" {
		// Build the schedule string
		var schedule string
		if specialTime != "" {
			schedule = "@" + specialTime
		} else {
			schedule = fmt.Sprintf("%s %s %s %s %s", minute, hour, day, month, weekday)
		}

		// Build the cron entry
		commentPrefix := ""
		if disabled {
			commentPrefix = "# DISABLED: "
		}

		script += `
# Get current crontab
CURRENT_CRON=$(crontab -l -u "$USER" 2>/dev/null || echo "")

# Check if job already exists
if echo "$CURRENT_CRON" | grep -F "# $JOB_NAME" >/dev/null 2>&1; then
    echo "Job '$JOB_NAME' already exists, updating..."
    # Remove old job (comment line and cron line)
    TEMP_CRON=$(echo "$CURRENT_CRON" | grep -v "# $JOB_NAME" | grep -v -F "$JOB_CMD" || true)
else
    echo "Adding new job '$JOB_NAME'..."
    TEMP_CRON="$CURRENT_CRON"
fi

# Add new job entry
if [ -n "$TEMP_CRON" ] && [ "$(echo "$TEMP_CRON" | tail -c 1)" != "" ]; then
    TEMP_CRON="$TEMP_CRON
"
fi

TEMP_CRON="${TEMP_CRON}# $JOB_NAME
` + commentPrefix + schedule + ` $JOB_CMD"

# Install new crontab
echo "$TEMP_CRON" | crontab -u "$USER" -

echo "Cron job '$JOB_NAME' has been added/updated"
`
	} else if state == "absent" {
		script += `
# Get current crontab
CURRENT_CRON=$(crontab -l -u "$USER" 2>/dev/null || echo "")

# Check if job exists
if echo "$CURRENT_CRON" | grep -F "# $JOB_NAME" >/dev/null 2>&1; then
    echo "Removing job '$JOB_NAME'..."

    # Remove the job (comment line and the next line which is the actual cron entry)
    TEMP_CRON=$(echo "$CURRENT_CRON" | awk '
        /^# '"$JOB_NAME"'$/ { skip=1; next }
        skip { skip=0; next }
        { print }
    ')

    # Install new crontab
    if [ -z "$TEMP_CRON" ]; then
        # Remove crontab if empty
        crontab -r -u "$USER" 2>/dev/null || true
        echo "Crontab is now empty, removed"
    else
        echo "$TEMP_CRON" | crontab -u "$USER" -
        echo "Cron job '$JOB_NAME' has been removed"
    fi
else
    echo "Job '$JOB_NAME' does not exist, nothing to remove"
fi
`
	}

	// Show current crontab
	script += `
# Show current crontab
echo ""
echo "Current crontab for user '$USER':"
crontab -l -u "$USER" 2>/dev/null || echo "(empty)"
`

	shellExec := []map[string]interface{}{
		{
			"type":    "shell",
			"command": script,
		},
	}

	var message string
	if state == "present" {
		message = fmt.Sprintf("Adding/updating cron job '%s' for user '%s'", name, user)
	} else {
		message = fmt.Sprintf("Removing cron job '%s' for user '%s'", name, user)
	}

	return TaskOutput{
		Status:  "ok",
		Message: message,
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
