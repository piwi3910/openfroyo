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

	// Extract parameters
	service, _ := input.Vars["service"].(string)
	port, _ := input.Vars["port"].(string)
	state, _ := input.Vars["state"].(string)
	zone, _ := input.Vars["zone"].(string)
	permanent, _ := input.Vars["permanent"].(bool)
	immediate, _ := input.Vars["immediate"].(bool)

	// Validate input
	if service == "" && port == "" {
		outputError("Either 'service' or 'port' must be specified")
		return
	}

	if state == "" {
		state = "enabled"
	}

	if zone == "" {
		zone = "public"
	}

	// Build the firewalld command
	result := buildFirewalldCommand(service, port, state, zone, permanent, immediate)
	printOutput(result)
}

func buildFirewalldCommand(service, port, state, zone string, permanent, immediate bool) TaskOutput {
	// Build shell script for firewalld management
	script := `
# Check if firewalld is installed and running
if ! command -v firewall-cmd >/dev/null 2>&1; then
    echo "firewalld is not installed"
    exit 1
fi

if ! systemctl is-active --quiet firewalld; then
    echo "firewalld is not running"
    exit 1
fi

echo "Managing firewalld configuration..."
`

	permanentFlag := ""
	if permanent {
		permanentFlag = "--permanent"
	}

	if service != "" {
		// Managing a service
		script += fmt.Sprintf(`
# Check current state for service: %s
CURRENT_STATE=""
if firewall-cmd --zone=%s --query-service=%s >/dev/null 2>&1; then
    CURRENT_STATE="enabled"
else
    CURRENT_STATE="disabled"
fi

echo "Current state for service %s: $CURRENT_STATE"
`, service, zone, service, service)

		if state == "enabled" {
			script += fmt.Sprintf(`
if [ "$CURRENT_STATE" = "disabled" ]; then
    echo "Enabling service: %s in zone: %s"
    firewall-cmd --zone=%s --add-service=%s %s
    if [ $? -eq 0 ]; then
        echo "Service %s enabled successfully"
    else
        echo "Failed to enable service %s"
        exit 1
    fi
else
    echo "Service %s is already enabled"
fi
`, service, zone, zone, service, permanentFlag, service, service, service)
		} else if state == "disabled" {
			script += fmt.Sprintf(`
if [ "$CURRENT_STATE" = "enabled" ]; then
    echo "Disabling service: %s in zone: %s"
    firewall-cmd --zone=%s --remove-service=%s %s
    if [ $? -eq 0 ]; then
        echo "Service %s disabled successfully"
    else
        echo "Failed to disable service %s"
        exit 1
    fi
else
    echo "Service %s is already disabled"
fi
`, service, zone, zone, service, permanentFlag, service, service, service)
		}
	} else if port != "" {
		// Managing a port
		script += fmt.Sprintf(`
# Check current state for port: %s
CURRENT_STATE=""
if firewall-cmd --zone=%s --query-port=%s >/dev/null 2>&1; then
    CURRENT_STATE="enabled"
else
    CURRENT_STATE="disabled"
fi

echo "Current state for port %s: $CURRENT_STATE"
`, port, zone, port, port)

		if state == "enabled" {
			script += fmt.Sprintf(`
if [ "$CURRENT_STATE" = "disabled" ]; then
    echo "Enabling port: %s in zone: %s"
    firewall-cmd --zone=%s --add-port=%s %s
    if [ $? -eq 0 ]; then
        echo "Port %s enabled successfully"
    else
        echo "Failed to enable port %s"
        exit 1
    fi
else
    echo "Port %s is already enabled"
fi
`, port, zone, zone, port, permanentFlag, port, port, port)
		} else if state == "disabled" {
			script += fmt.Sprintf(`
if [ "$CURRENT_STATE" = "enabled" ]; then
    echo "Disabling port: %s in zone: %s"
    firewall-cmd --zone=%s --remove-port=%s %s
    if [ $? -eq 0 ]; then
        echo "Port %s disabled successfully"
    else
        echo "Failed to disable port %s"
        exit 1
    fi
else
    echo "Port %s is already disabled"
fi
`, port, zone, zone, port, permanentFlag, port, port, port)
		}
	}

	// Reload firewalld if immediate and permanent flags are both set
	if immediate && permanent {
		script += `
# Reload firewalld to apply permanent changes
echo "Reloading firewalld..."
firewall-cmd --reload
if [ $? -eq 0 ]; then
    echo "Firewalld reloaded successfully"
else
    echo "Failed to reload firewalld"
    exit 1
fi
`
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
