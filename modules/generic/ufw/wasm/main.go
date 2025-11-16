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

	// Extract parameters
	rule, _ := input.Vars["rule"].(string)
	port, _ := input.Vars["port"].(string)
	proto, _ := input.Vars["proto"].(string)
	fromIP, _ := input.Vars["from_ip"].(string)
	toIP, _ := input.Vars["to_ip"].(string)
	direction, _ := input.Vars["direction"].(string)
	iface, _ := input.Vars["interface"].(string)
	state, _ := input.Vars["state"].(string)
	deleteRule, _ := input.Vars["delete"].(bool)

	// Set defaults
	if rule == "" {
		rule = "allow"
	}
	if state == "" {
		state = "enabled"
	}

	// Build the UFW command
	result := buildUFWCommand(rule, port, proto, fromIP, toIP, direction, iface, state, deleteRule)
	printOutput(result)
}

func buildUFWCommand(rule, port, proto, fromIP, toIP, direction, iface, state string, deleteRule bool) TaskOutput {
	// Build shell script for UFW management
	script := `
# Check if UFW is installed
if ! command -v ufw >/dev/null 2>&1; then
    echo "UFW is not installed"
    exit 1
fi

echo "Managing UFW configuration..."
`

	// Handle UFW enable/disable state
	if state == "enabled" {
		script += `
# Check if UFW is enabled
if ! sudo ufw status | grep -q "Status: active"; then
    echo "Enabling UFW..."
    echo "y" | sudo ufw enable
    if [ $? -eq 0 ]; then
        echo "UFW enabled successfully"
    else
        echo "Failed to enable UFW"
        exit 1
    fi
else
    echo "UFW is already enabled"
fi
`
	} else if state == "disabled" {
		script += `
# Check if UFW is disabled
if sudo ufw status | grep -q "Status: active"; then
    echo "Disabling UFW..."
    sudo ufw disable
    if [ $? -eq 0 ]; then
        echo "UFW disabled successfully"
    else
        echo "Failed to disable UFW"
        exit 1
    fi
else
    echo "UFW is already disabled"
fi
`
		// If we're just disabling UFW, we don't need to process rules
		shellExec := []map[string]interface{}{
			{
				"type":    "shell",
				"command": script,
			},
		}

		return TaskOutput{
			Status:  "ok",
			Message: "",
			Facts: map[string]interface{}{
				"shell_exec": shellExec,
			},
		}
	}

	// Build rule command if we have port or other rule specifications
	if port != "" || fromIP != "" || toIP != "" || direction != "" || iface != "" {
		var cmdParts []string

		// Add delete prefix if needed
		if deleteRule {
			cmdParts = append(cmdParts, "delete")
		}

		// Add rule type
		cmdParts = append(cmdParts, rule)

		// Add direction
		if direction != "" {
			cmdParts = append(cmdParts, direction)
		}

		// Add interface
		if iface != "" {
			cmdParts = append(cmdParts, "on", iface)
		}

		// Add from IP
		if fromIP != "" {
			cmdParts = append(cmdParts, "from", fromIP)
		} else {
			cmdParts = append(cmdParts, "from", "any")
		}

		// Add to specification
		if toIP != "" {
			cmdParts = append(cmdParts, "to", toIP)
		} else {
			cmdParts = append(cmdParts, "to", "any")
		}

		// Add port and protocol
		if port != "" {
			cmdParts = append(cmdParts, "port", port)
			if proto != "" {
				cmdParts = append(cmdParts, "proto", proto)
			}
		}

		ruleCmd := strings.Join(cmdParts, " ")

		// Check if rule exists (for non-delete operations)
		if !deleteRule {
			script += fmt.Sprintf(`
# Check if rule exists
RULE_EXISTS=0
if sudo ufw status numbered | grep -q "%s"; then
    RULE_EXISTS=1
    echo "Rule appears to exist"
else
    echo "Rule does not exist"
fi

if [ $RULE_EXISTS -eq 0 ]; then
    echo "Adding UFW rule: %s"
    sudo ufw %s
    if [ $? -eq 0 ]; then
        echo "Rule added successfully"
    else
        echo "Failed to add rule"
        exit 1
    fi
else
    echo "Rule already exists, no changes needed"
fi
`, port, ruleCmd, ruleCmd)
		} else {
			// For delete operations
			script += fmt.Sprintf(`
echo "Deleting UFW rule: %s"
sudo ufw %s 2>/dev/null
if [ $? -eq 0 ]; then
    echo "Rule deleted successfully"
else
    echo "Rule may not exist or failed to delete"
fi
`, ruleCmd, ruleCmd)
		}
	}

	shellExec := []map[string]interface{}{
		{
			"type":    "shell",
			"command": script,
		},
	}

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
