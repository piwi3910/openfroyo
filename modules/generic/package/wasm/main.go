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

// PackageManager represents a package manager configuration
type PackageManager struct {
	Name          string
	CheckCmd      string
	InstallCmd    string
	RemoveCmd     string
	UpdateCmd     string
	UpgradeCmd    string
	NeedsSudo     bool
	UpdateCache   string
}

var packageManagers = []PackageManager{
	// Debian/Ubuntu - apt
	{
		Name:         "apt",
		CheckCmd:     "which apt-get",
		InstallCmd:   "apt-get install -y",
		RemoveCmd:    "apt-get remove -y",
		UpdateCmd:    "apt-get update",
		UpgradeCmd:   "apt-get upgrade -y",
		NeedsSudo:    true,
		UpdateCache:  "apt-get update",
	},
	// RHEL/CentOS/Fedora - dnf
	{
		Name:         "dnf",
		CheckCmd:     "which dnf",
		InstallCmd:   "dnf install -y",
		RemoveCmd:    "dnf remove -y",
		UpdateCmd:    "dnf check-update",
		UpgradeCmd:   "dnf upgrade -y",
		NeedsSudo:    true,
		UpdateCache:  "dnf check-update",
	},
	// RHEL/CentOS (older) - yum
	{
		Name:         "yum",
		CheckCmd:     "which yum",
		InstallCmd:   "yum install -y",
		RemoveCmd:    "yum remove -y",
		UpdateCmd:    "yum check-update",
		UpgradeCmd:   "yum upgrade -y",
		NeedsSudo:    true,
		UpdateCache:  "yum check-update",
	},
	// Arch Linux - pacman
	{
		Name:         "pacman",
		CheckCmd:     "which pacman",
		InstallCmd:   "pacman -S --noconfirm",
		RemoveCmd:    "pacman -R --noconfirm",
		UpdateCmd:    "pacman -Sy",
		UpgradeCmd:   "pacman -Syu --noconfirm",
		NeedsSudo:    true,
		UpdateCache:  "pacman -Sy",
	},
	// macOS - brew
	{
		Name:         "brew",
		CheckCmd:     "which brew",
		InstallCmd:   "brew install",
		RemoveCmd:    "brew uninstall",
		UpdateCmd:    "brew update",
		UpgradeCmd:   "brew upgrade",
		NeedsSudo:    false,
		UpdateCache:  "brew update",
	},
	// Windows - chocolatey
	{
		Name:         "choco",
		CheckCmd:     "where choco",
		InstallCmd:   "choco install -y",
		RemoveCmd:    "choco uninstall -y",
		UpdateCmd:    "choco outdated",
		UpgradeCmd:   "choco upgrade all -y",
		NeedsSudo:    false,
		UpdateCache:  "",
	},
	// Windows - winget
	{
		Name:         "winget",
		CheckCmd:     "where winget",
		InstallCmd:   "winget install --silent",
		RemoveCmd:    "winget uninstall --silent",
		UpdateCmd:    "winget upgrade --all",
		UpgradeCmd:   "winget upgrade --all",
		NeedsSudo:    false,
		UpdateCache:  "",
	},
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
		outputError("Missing required variable: name (package name)")
		return
	}

	state, _ := input.Vars["state"].(string)
	if state == "" {
		state = "present" // Default to install
	}

	updateCache, _ := input.Vars["update_cache"].(bool)

	// Build the package manager detection and command
	result := buildPackageCommand(name, state, updateCache)
	printOutput(result)
}

func buildPackageCommand(name, state string, updateCache bool) TaskOutput {
	// Generate commands to detect package manager and execute action
	var commands []string

	// Add package manager detection
	commands = append(commands, "detect_pkg_mgr")

	// Add cache update if requested
	if updateCache {
		commands = append(commands, "update_cache")
	}

	// Add the actual package operation
	switch state {
	case "present":
		commands = append(commands, fmt.Sprintf("install:%s", name))
	case "absent":
		commands = append(commands, fmt.Sprintf("remove:%s", name))
	case "latest":
		commands = append(commands, "update_cache", fmt.Sprintf("upgrade:%s", name))
	default:
		return TaskOutput{
			Status:  "failed",
			Message: fmt.Sprintf("Invalid state: %s (must be present, absent, or latest)", state),
			Facts:   make(map[string]interface{}),
		}
	}

	return TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Package command prepared: %s -> %s", name, state),
		Facts: map[string]interface{}{
			"package_command": strings.Join(commands, ";"),
			"package_name":    name,
			"package_state":   state,
			"update_cache":    updateCache,
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
