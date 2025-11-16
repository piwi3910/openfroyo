package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// TaskInput represents the input from OpenFroyo
type TaskInput struct {
	Vars    map[string]interface{} `json:"vars"`
	Context TaskContext            `json:"context"`
}

// TaskContext provides execution context
type TaskContext struct {
	Host     string `json:"host"`
	TaskName string `json:"task_name"`
}

// TaskOutput represents the output to OpenFroyo
type TaskOutput struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message"`
	Facts   map[string]interface{} `json:"facts"`
}

func main() {
	var input TaskInput

	// Read input from stdin
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&input); err != nil {
		output := TaskOutput{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to decode input: %v", err),
			Facts:   make(map[string]interface{}),
		}
		printOutput(output)
		return
	}

	// Extract required variables
	device, ok := input.Vars["device"].(string)
	if !ok || device == "" {
		outputError("Missing required variable: device")
		return
	}

	path, ok := input.Vars["path"].(string)
	if !ok || path == "" {
		outputError("Missing required variable: path")
		return
	}

	fstype, ok := input.Vars["fstype"].(string)
	if !ok || fstype == "" {
		outputError("Missing required variable: fstype")
		return
	}

	// Extract optional variables with defaults
	opts := getString(input.Vars, "opts", "defaults")
	state := getString(input.Vars, "state", "mounted")
	fstab := getBool(input.Vars, "fstab", true)
	dump := getInt(input.Vars, "dump", 0)
	passno := getInt(input.Vars, "passno", 0)

	// Validate state
	validStates := map[string]bool{"mounted": true, "unmounted": true, "present": true, "absent": true}
	if !validStates[state] {
		outputError(fmt.Sprintf("Invalid state: %s (must be: mounted, unmounted, present, absent)", state))
		return
	}

	// Build shell_exec commands
	var commands []map[string]interface{}

	switch state {
	case "mounted":
		// 1. Check if already mounted
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("mount | grep -q ' %s ' && echo 'MOUNTED' || echo 'NOT_MOUNTED'", escapePath(path)),
		})

		// 2. Create mount point if needed
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("mkdir -p %s", escapePath(path)),
		})

		// 3. Mount the filesystem
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("mount -t %s -o %s %s %s", fstype, opts, device, escapePath(path)),
		})

		// 4. Add to fstab if requested
		if fstab {
			fstabEntry := fmt.Sprintf("%s %s %s %s %d %d", device, path, fstype, opts, dump, passno)
			commands = append(commands, map[string]interface{}{
				"type":    "shell",
				"command": fmt.Sprintf("grep -q '^%s ' /etc/fstab || echo '%s' >> /etc/fstab", escapePath(path), fstabEntry),
			})
		}

	case "unmounted":
		// 1. Check if mounted
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("mount | grep -q ' %s ' && echo 'MOUNTED' || echo 'NOT_MOUNTED'", escapePath(path)),
		})

		// 2. Unmount if mounted
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("umount %s || true", escapePath(path)),
		})

	case "present":
		// 1. Create mount point
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("mkdir -p %s", escapePath(path)),
		})

		// 2. Add to fstab if requested
		if fstab {
			fstabEntry := fmt.Sprintf("%s %s %s %s %d %d", device, path, fstype, opts, dump, passno)
			commands = append(commands, map[string]interface{}{
				"type":    "shell",
				"command": fmt.Sprintf("grep -q '^%s ' /etc/fstab || echo '%s' >> /etc/fstab", escapePath(path), fstabEntry),
			})
		}

	case "absent":
		// 1. Unmount if mounted
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("umount %s || true", escapePath(path)),
		})

		// 2. Remove from fstab
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("sed -i.bak '\\|^%s |d' /etc/fstab || true", escapePath(path)),
		})

		// 3. Remove mount point
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("rmdir %s || true", escapePath(path)),
		})
	}

	// Return shell_exec fact
	result := TaskOutput{
		Status:  "ok",
		Message: "",
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}

	printOutput(result)
}

func getString(vars map[string]interface{}, key string, defaultVal string) string {
	if val, ok := vars[key].(string); ok {
		return val
	}
	return defaultVal
}

func getBool(vars map[string]interface{}, key string, defaultVal bool) bool {
	if val, ok := vars[key].(bool); ok {
		return val
	}
	return defaultVal
}

func getInt(vars map[string]interface{}, key string, defaultVal int) int {
	if val, ok := vars[key].(float64); ok {
		return int(val)
	}
	if val, ok := vars[key].(string); ok {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func escapePath(path string) string {
	// Escape spaces and special characters for shell
	path = strings.ReplaceAll(path, " ", "\\ ")
	return path
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
	outputJSON, _ := json.MarshalIndent(output, "", "  ")
	fmt.Println(string(outputJSON))
}
