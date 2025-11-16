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
	chain, _ := input.Vars["chain"].(string)
	protocol, _ := input.Vars["protocol"].(string)
	destinationPort, _ := input.Vars["destination_port"].(string)
	source, _ := input.Vars["source"].(string)
	destination, _ := input.Vars["destination"].(string)
	jump, _ := input.Vars["jump"].(string)
	state, _ := input.Vars["state"].(string)
	table, _ := input.Vars["table"].(string)

	// Set defaults
	if chain == "" {
		chain = "INPUT"
	}
	if jump == "" {
		jump = "ACCEPT"
	}
	if state == "" {
		state = "present"
	}
	if table == "" {
		table = "filter"
	}

	// Build the iptables command
	result := buildIptablesCommand(chain, protocol, destinationPort, source, destination, jump, state, table)
	printOutput(result)
}

func buildIptablesCommand(chain, protocol, destinationPort, source, destination, jump, state, table string) TaskOutput {
	// Build rule specification parts
	var ruleParts []string

	if protocol != "" {
		ruleParts = append(ruleParts, fmt.Sprintf("-p %s", protocol))
	}

	if source != "" {
		ruleParts = append(ruleParts, fmt.Sprintf("-s %s", source))
	}

	if destination != "" {
		ruleParts = append(ruleParts, fmt.Sprintf("-d %s", destination))
	}

	if destinationPort != "" {
		ruleParts = append(ruleParts, fmt.Sprintf("--dport %s", destinationPort))
	}

	ruleParts = append(ruleParts, fmt.Sprintf("-j %s", jump))

	ruleSpec := strings.Join(ruleParts, " ")

	// Build shell script for iptables management
	script := `
# Check if iptables is installed
if ! command -v iptables >/dev/null 2>&1; then
    echo "iptables is not installed"
    exit 1
fi

echo "Managing iptables rule..."
`

	script += fmt.Sprintf(`
# Define rule components
TABLE="%s"
CHAIN="%s"
RULE_SPEC="%s"

# Check if rule exists
echo "Checking if rule exists..."
if sudo iptables -t $TABLE -C $CHAIN $RULE_SPEC 2>/dev/null; then
    RULE_EXISTS=1
    echo "Rule exists"
else
    RULE_EXISTS=0
    echo "Rule does not exist"
fi
`, table, chain, ruleSpec)

	if state == "present" {
		script += `
# Add rule if it doesn't exist
if [ $RULE_EXISTS -eq 0 ]; then
    echo "Adding iptables rule..."
    sudo iptables -t $TABLE -A $CHAIN $RULE_SPEC
    if [ $? -eq 0 ]; then
        echo "Rule added successfully"

        # Save rules based on distribution
        if command -v netfilter-persistent >/dev/null 2>&1; then
            echo "Saving rules with netfilter-persistent..."
            sudo netfilter-persistent save
        elif [ -f /etc/debian_version ]; then
            echo "Saving rules to /etc/iptables/rules.v4..."
            sudo mkdir -p /etc/iptables
            sudo iptables-save | sudo tee /etc/iptables/rules.v4 >/dev/null
        elif [ -f /etc/redhat-release ]; then
            echo "Saving rules to /etc/sysconfig/iptables..."
            sudo iptables-save | sudo tee /etc/sysconfig/iptables >/dev/null
        else
            echo "Warning: Could not determine how to persist iptables rules"
        fi
    else
        echo "Failed to add rule"
        exit 1
    fi
else
    echo "Rule already exists, no changes needed"
fi
`
	} else if state == "absent" {
		script += `
# Remove rule if it exists
if [ $RULE_EXISTS -eq 1 ]; then
    echo "Removing iptables rule..."
    sudo iptables -t $TABLE -D $CHAIN $RULE_SPEC
    if [ $? -eq 0 ]; then
        echo "Rule removed successfully"

        # Save rules based on distribution
        if command -v netfilter-persistent >/dev/null 2>&1; then
            echo "Saving rules with netfilter-persistent..."
            sudo netfilter-persistent save
        elif [ -f /etc/debian_version ]; then
            echo "Saving rules to /etc/iptables/rules.v4..."
            sudo mkdir -p /etc/iptables
            sudo iptables-save | sudo tee /etc/iptables/rules.v4 >/dev/null
        elif [ -f /etc/redhat-release ]; then
            echo "Saving rules to /etc/sysconfig/iptables..."
            sudo iptables-save | sudo tee /etc/sysconfig/iptables >/dev/null
        else
            echo "Warning: Could not determine how to persist iptables rules"
        fi
    else
        echo "Failed to remove rule"
        exit 1
    fi
else
    echo "Rule does not exist, no changes needed"
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
