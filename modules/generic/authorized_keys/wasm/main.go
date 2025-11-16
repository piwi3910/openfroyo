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
	user, _ := input.Vars["user"].(string)
	if user == "" {
		outputError("Missing required variable: user (username)")
		return
	}

	key, _ := input.Vars["key"].(string)
	if key == "" {
		outputError("Missing required variable: key (SSH public key)")
		return
	}

	state, _ := input.Vars["state"].(string)
	if state == "" {
		state = "present"
	}

	keyOptions, _ := input.Vars["key_options"].(string)
	exclusive, _ := input.Vars["exclusive"].(bool)
	customPath, _ := input.Vars["path"].(string)

	// Build the authorized_keys management command
	result := buildAuthorizedKeysCommand(user, key, state, keyOptions, exclusive, customPath)
	printOutput(result)
}

func buildAuthorizedKeysCommand(user, key, state, keyOptions string, exclusive bool, customPath string) TaskOutput {
	// Escape single quotes in the key
	escapedKey := strings.ReplaceAll(key, "'", "'\\''")

	// Build shell commands to execute
	var script string

	// Determine the authorized_keys path
	if customPath == "" {
		script = fmt.Sprintf(`
# Get user home directory
USER_HOME=$(getent passwd "%s" | cut -d: -f6)
if [ -z "$USER_HOME" ]; then
    echo "User %s does not exist"
    exit 1
fi
AUTH_KEYS="$USER_HOME/.ssh/authorized_keys"
`, user, user)
	} else {
		script = fmt.Sprintf(`
AUTH_KEYS="%s"
`, customPath)
	}

	switch state {
	case "present":
		// Prepare the key with options if provided
		fullKey := key
		if keyOptions != "" {
			fullKey = fmt.Sprintf("%s %s", keyOptions, key)
		}
		escapedFullKey := strings.ReplaceAll(fullKey, "'", "'\\''")

		script += fmt.Sprintf(`
# Ensure .ssh directory exists
SSH_DIR=$(dirname "$AUTH_KEYS")
if [ ! -d "$SSH_DIR" ]; then
    echo "Creating .ssh directory: $SSH_DIR"
    sudo mkdir -p "$SSH_DIR"
    sudo chown %s "$SSH_DIR"
    sudo chmod 700 "$SSH_DIR"
fi

# Ensure authorized_keys file exists
if [ ! -f "$AUTH_KEYS" ]; then
    echo "Creating authorized_keys file: $AUTH_KEYS"
    sudo touch "$AUTH_KEYS"
    sudo chown %s "$AUTH_KEYS"
    sudo chmod 600 "$AUTH_KEYS"
fi
`, user, user)

		if exclusive {
			// Replace all keys with this one
			script += fmt.Sprintf(`
# Exclusive mode: replace all keys
echo "Replacing all keys with the specified key"
echo '%s' | sudo tee "$AUTH_KEYS" > /dev/null
sudo chown %s "$AUTH_KEYS"
sudo chmod 600 "$AUTH_KEYS"
echo "Key set (exclusive mode)"
`, escapedFullKey, user)
		} else {
			// Add key if not present
			script += fmt.Sprintf(`
# Extract the key part (without options) for checking
KEY_ONLY='%s'

# Check if key already exists (check for the key itself, ignoring options)
if sudo grep -qF "$KEY_ONLY" "$AUTH_KEYS" 2>/dev/null; then
    echo "SSH key already present in authorized_keys"
else
    echo "Adding SSH key to authorized_keys"
    echo '%s' | sudo tee -a "$AUTH_KEYS" > /dev/null
    sudo chown %s "$AUTH_KEYS"
    sudo chmod 600 "$AUTH_KEYS"
    echo "Key added"
fi

# Remove duplicates
sudo awk '!seen[$0]++' "$AUTH_KEYS" | sudo tee "$AUTH_KEYS.tmp" > /dev/null
sudo mv "$AUTH_KEYS.tmp" "$AUTH_KEYS"
sudo chown %s "$AUTH_KEYS"
sudo chmod 600 "$AUTH_KEYS"
`, escapedKey, escapedFullKey, user, user)
		}

	case "absent":
		script += fmt.Sprintf(`
# Check if authorized_keys file exists
if [ ! -f "$AUTH_KEYS" ]; then
    echo "authorized_keys file does not exist"
    exit 0
fi

# Extract the key part for removal
KEY_ONLY='%s'

# Remove the key
if sudo grep -qF "$KEY_ONLY" "$AUTH_KEYS" 2>/dev/null; then
    echo "Removing SSH key from authorized_keys"
    sudo grep -vF "$KEY_ONLY" "$AUTH_KEYS" | sudo tee "$AUTH_KEYS.tmp" > /dev/null
    sudo mv "$AUTH_KEYS.tmp" "$AUTH_KEYS"
    sudo chown %s "$AUTH_KEYS"
    sudo chmod 600 "$AUTH_KEYS"
    echo "Key removed"
else
    echo "SSH key not found in authorized_keys"
fi
`, escapedKey, user)
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
