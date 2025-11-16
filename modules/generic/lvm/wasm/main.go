package main

import (
	"encoding/json"
	"fmt"
	"os"
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

	// Extract optional variables
	vg := getString(input.Vars, "vg", "")
	lv := getString(input.Vars, "lv", "")
	size := getString(input.Vars, "size", "")
	state := getString(input.Vars, "state", "present")
	force := getBool(input.Vars, "force", false)
	opts := getString(input.Vars, "opts", "")

	// Extract pvs array
	var pvs []string
	if pvsInterface, ok := input.Vars["pvs"]; ok {
		if pvsArray, ok := pvsInterface.([]interface{}); ok {
			for _, pv := range pvsArray {
				if pvStr, ok := pv.(string); ok {
					pvs = append(pvs, pvStr)
				}
			}
		}
	}

	// Validate state
	validStates := map[string]bool{"present": true, "absent": true}
	if !validStates[state] {
		outputError(fmt.Sprintf("Invalid state: %s (must be: present, absent)", state))
		return
	}

	// Build shell_exec commands based on operation type
	var commands []map[string]interface{}

	// Determine operation type
	if vg != "" && lv != "" {
		// Logical Volume operation
		commands = buildLVCommands(vg, lv, size, state, force, opts)
	} else if vg != "" && len(pvs) > 0 {
		// Volume Group operation
		commands = buildVGCommands(vg, pvs, state, force)
	} else if len(pvs) > 0 {
		// Physical Volume operation
		commands = buildPVCommands(pvs, state, force)
	} else {
		outputError("Invalid parameters: must specify either (vg + lv), (vg + pvs), or (pvs)")
		return
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

func buildLVCommands(vg, lv, size, state string, force bool, opts string) []map[string]interface{} {
	var commands []map[string]interface{}

	lvPath := fmt.Sprintf("%s/%s", vg, lv)

	if state == "present" {
		// 1. Check if LV exists
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("lvdisplay %s >/dev/null 2>&1 && echo 'EXISTS' || echo 'NOT_EXISTS'", lvPath),
		})

		// 2. Create LV if it doesn't exist
		createCmd := fmt.Sprintf("lvcreate -n %s", lv)
		if size != "" {
			// Determine size option based on format
			if strings.HasSuffix(size, "%") || strings.Contains(size, "FREE") {
				createCmd += fmt.Sprintf(" -l %s", size)
			} else {
				createCmd += fmt.Sprintf(" -L %s", size)
			}
		}
		if opts != "" {
			createCmd += fmt.Sprintf(" %s", opts)
		}
		createCmd += fmt.Sprintf(" %s", vg)

		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("lvdisplay %s >/dev/null 2>&1 || %s", lvPath, createCmd),
		})

		// 3. Extend LV if size is specified and LV exists
		if size != "" {
			extendCmd := fmt.Sprintf("lvextend -L %s %s", size, lvPath)
			commands = append(commands, map[string]interface{}{
				"type":    "shell",
				"command": fmt.Sprintf("lvdisplay %s >/dev/null 2>&1 && %s || true", lvPath, extendCmd),
			})
		}

	} else if state == "absent" {
		// 1. Check if LV exists
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("lvdisplay %s >/dev/null 2>&1 && echo 'EXISTS' || echo 'NOT_EXISTS'", lvPath),
		})

		// 2. Remove LV
		removeCmd := fmt.Sprintf("lvremove %s", lvPath)
		if force {
			removeCmd = fmt.Sprintf("lvremove -f %s", lvPath)
		}

		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("lvdisplay %s >/dev/null 2>&1 && %s || true", lvPath, removeCmd),
		})
	}

	return commands
}

func buildVGCommands(vg string, pvs []string, state string, force bool) []map[string]interface{} {
	var commands []map[string]interface{}

	pvsStr := strings.Join(pvs, " ")

	if state == "present" {
		// 1. Check if VG exists
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("vgdisplay %s >/dev/null 2>&1 && echo 'EXISTS' || echo 'NOT_EXISTS'", vg),
		})

		// 2. Create PVs if they don't exist
		for _, pv := range pvs {
			commands = append(commands, map[string]interface{}{
				"type":    "shell",
				"command": fmt.Sprintf("pvdisplay %s >/dev/null 2>&1 || pvcreate %s", pv, pv),
			})
		}

		// 3. Create VG if it doesn't exist
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("vgdisplay %s >/dev/null 2>&1 || vgcreate %s %s", vg, vg, pvsStr),
		})

		// 4. Extend VG with additional PVs
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("vgdisplay %s >/dev/null 2>&1 && vgextend %s %s || true", vg, vg, pvsStr),
		})

	} else if state == "absent" {
		// 1. Check if VG exists
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("vgdisplay %s >/dev/null 2>&1 && echo 'EXISTS' || echo 'NOT_EXISTS'", vg),
		})

		// 2. Remove VG
		removeCmd := fmt.Sprintf("vgremove %s", vg)
		if force {
			removeCmd = fmt.Sprintf("vgremove -f %s", vg)
		}

		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": fmt.Sprintf("vgdisplay %s >/dev/null 2>&1 && %s || true", vg, removeCmd),
		})
	}

	return commands
}

func buildPVCommands(pvs []string, state string, force bool) []map[string]interface{} {
	var commands []map[string]interface{}

	if state == "present" {
		// Create each PV
		for _, pv := range pvs {
			// 1. Check if PV exists
			commands = append(commands, map[string]interface{}{
				"type":    "shell",
				"command": fmt.Sprintf("pvdisplay %s >/dev/null 2>&1 && echo 'EXISTS' || echo 'NOT_EXISTS'", pv),
			})

			// 2. Create PV if it doesn't exist
			createCmd := fmt.Sprintf("pvcreate %s", pv)
			if force {
				createCmd = fmt.Sprintf("pvcreate -f %s", pv)
			}

			commands = append(commands, map[string]interface{}{
				"type":    "shell",
				"command": fmt.Sprintf("pvdisplay %s >/dev/null 2>&1 || %s", pv, createCmd),
			})
		}

	} else if state == "absent" {
		// Remove each PV
		for _, pv := range pvs {
			// 1. Check if PV exists
			commands = append(commands, map[string]interface{}{
				"type":    "shell",
				"command": fmt.Sprintf("pvdisplay %s >/dev/null 2>&1 && echo 'EXISTS' || echo 'NOT_EXISTS'", pv),
			})

			// 2. Remove PV
			removeCmd := fmt.Sprintf("pvremove %s", pv)
			if force {
				removeCmd = fmt.Sprintf("pvremove -f %s", pv)
			}

			commands = append(commands, map[string]interface{}{
				"type":    "shell",
				"command": fmt.Sprintf("pvdisplay %s >/dev/null 2>&1 && %s || true", pv, removeCmd),
			})
		}
	}

	return commands
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
