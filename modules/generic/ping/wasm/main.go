package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
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
	Status  string                 `json:"status"`  // "ok", "changed", "failed"
	Message string                 `json:"message"` // Human-readable message
	Facts   map[string]interface{} `json:"facts"`   // Discovered facts
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

	// Extract host to ping
	hostInterface, ok := input.Vars["host"]
	if !ok {
		outputError("Missing required variable: host")
		return
	}

	hostStr, ok := hostInterface.(string)
	if !ok {
		outputError("Variable 'host' must be a string")
		return
	}

	if strings.TrimSpace(hostStr) == "" {
		outputError("Host cannot be empty")
		return
	}

	// Extract optional count (default: 4)
	count := 4
	if countInterface, exists := input.Vars["count"]; exists {
		switch v := countInterface.(type) {
		case float64:
			count = int(v)
		case int:
			count = v
		case string:
			if parsedCount, err := strconv.Atoi(v); err == nil {
				count = parsedCount
			}
		}
	}

	if count < 1 {
		outputError("Count must be at least 1")
		return
	}

	// Extract optional timeout (default: 5)
	timeout := 5
	if timeoutInterface, exists := input.Vars["timeout"]; exists {
		switch v := timeoutInterface.(type) {
		case float64:
			timeout = int(v)
		case int:
			timeout = v
		case string:
			if parsedTimeout, err := strconv.Atoi(v); err == nil {
				timeout = parsedTimeout
			}
		}
	}

	if timeout < 1 {
		outputError("Timeout must be at least 1 second")
		return
	}

	// Extract optional interface
	interfaceStr := ""
	if interfaceInterface, exists := input.Vars["interface"]; exists {
		if str, ok := interfaceInterface.(string); ok {
			interfaceStr = strings.TrimSpace(str)
		}
	}

	// Build ping command
	// Basic: ping -c {count} -W {timeout} {host}
	// With interface: ping -c {count} -W {timeout} -I {interface} {host}
	var cmd string
	if interfaceStr != "" {
		cmd = fmt.Sprintf("ping -c %d -W %d -I %s %s 2>&1",
			count, timeout, shellEscape(interfaceStr), shellEscape(hostStr))
	} else {
		cmd = fmt.Sprintf("ping -c %d -W %d %s 2>&1",
			count, timeout, shellEscape(hostStr))
	}

	// Add parsing of output for statistics
	// The ping command output varies by platform, but we'll extract:
	// - Packets transmitted/received/lost
	// - RTT min/avg/max
	parseScript := `
OUTPUT=$(%s)
EXIT_CODE=$?

# Extract statistics
PACKETS_TX=$(echo "$OUTPUT" | grep -oE '[0-9]+ packets transmitted' | grep -oE '[0-9]+' | head -n1)
PACKETS_RX=$(echo "$OUTPUT" | grep -oE '[0-9]+ received' | grep -oE '[0-9]+' | head -n1)
PACKET_LOSS=$(echo "$OUTPUT" | grep -oE '[0-9]+%% packet loss' | grep -oE '[0-9]+' | head -n1)

# Extract RTT stats (format: min/avg/max/mdev or min/avg/max)
RTT_LINE=$(echo "$OUTPUT" | grep -i 'rtt\|round-trip' | tail -n1)
if [ -n "$RTT_LINE" ]; then
    RTT_STATS=$(echo "$RTT_LINE" | grep -oE '[0-9]+\.[0-9]+/[0-9]+\.[0-9]+/[0-9]+\.[0-9]+' | head -n1)
    if [ -n "$RTT_STATS" ]; then
        RTT_MIN=$(echo "$RTT_STATS" | cut -d'/' -f1)
        RTT_AVG=$(echo "$RTT_STATS" | cut -d'/' -f2)
        RTT_MAX=$(echo "$RTT_STATS" | cut -d'/' -f3)
    fi
fi

# Build JSON output
echo "{"
echo "  \"exit_code\": $EXIT_CODE,"
echo "  \"packets_transmitted\": ${PACKETS_TX:-0},"
echo "  \"packets_received\": ${PACKETS_RX:-0},"
echo "  \"packet_loss_percent\": ${PACKET_LOSS:-100},"
if [ -n "$RTT_MIN" ]; then
    echo "  \"rtt_min\": $RTT_MIN,"
    echo "  \"rtt_avg\": $RTT_AVG,"
    echo "  \"rtt_max\": $RTT_MAX,"
fi
echo "  \"output\": \"$(echo "$OUTPUT" | sed 's/"/\\"/g' | tr '\n' ' ')\""
echo "}"
`

	fullCmd := fmt.Sprintf(parseScript, cmd)

	// Return shell_exec fact for the generic handler
	result := TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Pinging %s (%d packets, %ds timeout)", hostStr, count, timeout),
		Facts: map[string]interface{}{
			"shell_exec": []map[string]interface{}{
				{
					"type":    "shell",
					"command": fullCmd,
				},
			},
		},
	}
	printOutput(result)
}

func shellEscape(s string) string {
	// Simple shell escaping: wrap in single quotes and escape any single quotes
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
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
