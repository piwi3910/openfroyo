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
	// Detect which package manager is available
	var pkgMgr *PackageManager
	for i := range packageManagers {
		pm := &packageManagers[i]
		// We'll use the check command to detect availability
		pkgMgr = pm
		break // For now, just use the first one; detection happens via shell_exec
	}

	if pkgMgr == nil {
		return TaskOutput{
			Status:  "failed",
			Message: "No package manager configuration found",
			Facts:   make(map[string]interface{}),
		}
	}

	// Build shell commands to execute
	// We create a shell script that detects the package manager and runs the appropriate command
	detectionScript := buildPackageManagerDetectionScript(name, state, updateCache)

	shellExec := []map[string]interface{}{
		{
			"type":    "shell",
			"command": detectionScript,
		},
	}

	return TaskOutput{
		Status:  "ok",
		Message: "", // Will be set by handler
		Facts: map[string]interface{}{
			"shell_exec": shellExec,
		},
	}
}

func buildPackageManagerDetectionScript(name, state string, updateCache bool) string {
	// Build a shell script that detects the package manager and runs the appropriate command
	script := `
# Detect package manager and execute package operation
if command -v apt-get >/dev/null 2>&1; then
    PKG_MGR="apt-get"
    INSTALL_CMD="sudo apt-get install -y"
    REMOVE_CMD="sudo apt-get remove -y"
    UPDATE_CMD="sudo apt-get update"
    UPGRADE_CMD="sudo apt-get install --only-upgrade -y"
elif command -v dnf >/dev/null 2>&1; then
    PKG_MGR="dnf"
    INSTALL_CMD="sudo dnf install -y"
    REMOVE_CMD="sudo dnf remove -y"
    UPDATE_CMD="sudo dnf check-update || true"
    UPGRADE_CMD="sudo dnf upgrade -y"
elif command -v yum >/dev/null 2>&1; then
    PKG_MGR="yum"
    INSTALL_CMD="sudo yum install -y"
    REMOVE_CMD="sudo yum remove -y"
    UPDATE_CMD="sudo yum check-update || true"
    UPGRADE_CMD="sudo yum upgrade -y"
elif command -v pacman >/dev/null 2>&1; then
    PKG_MGR="pacman"
    INSTALL_CMD="sudo pacman -S --noconfirm"
    REMOVE_CMD="sudo pacman -R --noconfirm"
    UPDATE_CMD="sudo pacman -Sy"
    UPGRADE_CMD="sudo pacman -S --noconfirm"
elif command -v brew >/dev/null 2>&1; then
    PKG_MGR="brew"
    INSTALL_CMD="brew install"
    REMOVE_CMD="brew uninstall"
    UPDATE_CMD="brew update"
    UPGRADE_CMD="brew upgrade"
elif command -v choco >/dev/null 2>&1; then
    PKG_MGR="choco"
    INSTALL_CMD="choco install -y"
    REMOVE_CMD="choco uninstall -y"
    UPDATE_CMD=""
    UPGRADE_CMD="choco upgrade -y"
elif command -v winget >/dev/null 2>&1; then
    PKG_MGR="winget"
    INSTALL_CMD="winget install --silent"
    REMOVE_CMD="winget uninstall --silent"
    UPDATE_CMD=""
    UPGRADE_CMD="winget upgrade --silent"
else
    echo "No supported package manager found"
    exit 1
fi

echo "Detected package manager: $PKG_MGR"
`

	// Add cache update if requested
	if updateCache {
		script += `
if [ -n "$UPDATE_CMD" ]; then
    echo "Updating package cache..."
    eval $UPDATE_CMD
fi
`
	}

	// Add the actual package operation
	switch state {
	case "present":
		script += fmt.Sprintf(`
echo "Installing package: %s"
eval $INSTALL_CMD %s
`, name, name)
	case "absent":
		script += fmt.Sprintf(`
echo "Removing package: %s"
eval $REMOVE_CMD %s
`, name, name)
	case "latest":
		script += fmt.Sprintf(`
if [ -n "$UPDATE_CMD" ]; then
    echo "Updating package cache..."
    eval $UPDATE_CMD
fi
echo "Upgrading package: %s"
eval $UPGRADE_CMD %s
`, name, name)
	}

	return script
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
