package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// TaskInput represents the input to a WASM module
type TaskInput struct {
	Vars    map[string]interface{} `json:"vars"`
	Context TaskContext            `json:"context"`
}

// TaskContext provides execution context to the module
type TaskContext struct {
	Host     string `json:"host"`
	TaskName string `json:"task_name"`
}

// TaskOutput represents the output from a WASM module
type TaskOutput struct {
	Status  string                 `json:"status"`  // "ok", "changed", "failed"
	Message string                 `json:"message"` // Human-readable message
	Facts   map[string]interface{} `json:"facts"`   // Discovered facts
}

func main() {
	var modulePath string
	var inputBase64 string

	flag.StringVar(&modulePath, "module", "", "Path to WASM module")
	flag.StringVar(&inputBase64, "input-base64", "", "Base64-encoded JSON input")
	flag.Parse()

	if modulePath == "" || inputBase64 == "" {
		fmt.Fprintln(os.Stderr, "Usage: froyo-runner --module <path.wasm> --input-base64 <base64>")
		os.Exit(1)
	}

	// Decode input
	inputJSON, err := base64.StdEncoding.DecodeString(inputBase64)
	if err != nil {
		output := TaskOutput{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to decode input: %v", err),
			Facts:   make(map[string]interface{}),
		}
		printOutput(output)
		os.Exit(1)
	}

	var input TaskInput
	if err := json.Unmarshal(inputJSON, &input); err != nil {
		output := TaskOutput{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to parse input JSON: %v", err),
			Facts:   make(map[string]interface{}),
		}
		printOutput(output)
		os.Exit(1)
	}

	// Read WASM module
	wasmBytes, err := os.ReadFile(modulePath)
	if err != nil {
		output := TaskOutput{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to read WASM module: %v", err),
			Facts:   make(map[string]interface{}),
		}
		printOutput(output)
		os.Exit(1)
	}

	// Execute WASM module
	output := executeWASM(wasmBytes, input)
	printOutput(output)

	if output.Status == "failed" {
		os.Exit(1)
	}
}

func executeWASM(wasmBytes []byte, input TaskInput) TaskOutput {
	ctx := context.Background()

	// Create a new WebAssembly runtime
	r := wazero.NewRuntime(ctx)
	defer r.Close(ctx)

	// Instantiate WASI, which provides host functions for I/O, env vars, etc.
	wasi_snapshot_preview1.Instantiate(ctx, r)

	// Create stdin with input JSON
	inputJSON, _ := json.Marshal(input)
	stdin := strings.NewReader(string(inputJSON))

	// Capture stdout
	var stdout, stderr strings.Builder

	// Configure module with I/O, filesystem access, and environment
	config := wazero.NewModuleConfig().
		WithStdin(stdin).
		WithStdout(&stdout).
		WithStderr(&stderr).
		WithSysNanosleep().
		WithSysWalltime().
		WithFS(os.DirFS("/")).  // Grant access to root filesystem for command execution
		WithEnv("PATH", "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin")

	// Instantiate the module
	mod, err := r.InstantiateWithConfig(ctx, wasmBytes, config)
	if err != nil {
		return TaskOutput{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to instantiate WASM module: %v", err),
			Facts:   make(map[string]interface{}),
		}
	}
	defer mod.Close(ctx)

	// The WASM module should print JSON to stdout
	var output TaskOutput
	if err := json.Unmarshal([]byte(stdout.String()), &output); err != nil {
		return TaskOutput{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to parse module output: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String()),
			Facts:   make(map[string]interface{}),
		}
	}

	// Check if the module wants us to execute a command
	if execCmdInterface, exists := output.Facts["exec_command"]; exists {
		if execCmd, ok := execCmdInterface.(string); ok && execCmd != "" {
			output = executeCommand(output, execCmd)
		}
	}

	// Check if the module wants us to manage packages
	if pkgCmdInterface, exists := output.Facts["package_command"]; exists {
		if pkgCmd, ok := pkgCmdInterface.(string); ok && pkgCmd != "" {
			output = executePackageCommand(output, pkgCmd)
		}
	}

	// Check if the module wants us to execute a script
	if scriptCmdInterface, exists := output.Facts["script_command"]; exists {
		if scriptCmd, ok := scriptCmdInterface.(string); ok && scriptCmd != "" {
			// Get script details from facts
			scriptPath, _ := output.Facts["script_path"].(string)
			scriptContent, _ := output.Facts["script_content"].(string)
			output = executeScript(output, scriptPath, scriptContent, scriptCmd)
		}
	}

	return output
}

func executeCommand(output TaskOutput, execCmd string) TaskOutput {
	var cmd *exec.Cmd

	// Platform-specific command execution
	if os.Getenv("OS") != "" && strings.Contains(strings.ToLower(os.Getenv("OS")), "windows") {
		// Windows: use cmd.exe /c for command execution
		cmd = exec.Command("cmd.exe", "/c", execCmd)
	} else {
		// Unix-like (Linux, macOS, FreeBSD): parse and execute directly
		parts := strings.Fields(execCmd)
		if len(parts) > 0 {
			cmd = exec.Command(parts[0], parts[1:]...)
		}
	}

	if cmd != nil {
		cmdOutput, cmdErr := cmd.CombinedOutput()

		if cmdErr != nil {
			output.Status = "failed"
			output.Message = fmt.Sprintf("Command failed: %v", cmdErr)
			output.Facts["stdout"] = string(cmdOutput)
			output.Facts["error"] = cmdErr.Error()
		} else {
			output.Status = "ok"
			output.Message = fmt.Sprintf("Command executed successfully: %s", execCmd)
			output.Facts["stdout"] = strings.TrimSpace(string(cmdOutput))
		}
		delete(output.Facts, "exec_command")
	}

	return output
}

func executePackageCommand(output TaskOutput, pkgCmd string) TaskOutput {
	commands := strings.Split(pkgCmd, ";")

	var packageManager string
	var packageName string
	var results []string

	for _, cmdDirective := range commands {
		cmdDirective = strings.TrimSpace(cmdDirective)

		if cmdDirective == "detect_pkg_mgr" {
			// Detect available package manager
			pkgMgr := detectPackageManager()
			if pkgMgr == "" {
				output.Status = "failed"
				output.Message = "No supported package manager found"
				delete(output.Facts, "package_command")
				return output
			}
			packageManager = pkgMgr
			results = append(results, fmt.Sprintf("Detected package manager: %s", pkgMgr))
			continue
		}

		if cmdDirective == "update_cache" {
			// Update package cache
			updateCmd := getUpdateCacheCommand(packageManager)
			if updateCmd != "" {
				cmd := exec.Command("sh", "-c", updateCmd)
				cmdOutput, _ := cmd.CombinedOutput()
				results = append(results, fmt.Sprintf("Updated package cache: %s", strings.TrimSpace(string(cmdOutput))))
			}
			continue
		}

		// Handle install/remove/upgrade commands
		if strings.HasPrefix(cmdDirective, "install:") {
			packageName = strings.TrimPrefix(cmdDirective, "install:")
			installCmd := getInstallCommand(packageManager, packageName)
			cmd := exec.Command("sh", "-c", installCmd)
			cmdOutput, cmdErr := cmd.CombinedOutput()

			if cmdErr != nil {
				output.Status = "failed"
				output.Message = fmt.Sprintf("Failed to install %s: %v", packageName, cmdErr)
				output.Facts["error"] = cmdErr.Error()
			} else {
				output.Status = "ok"
				output.Message = fmt.Sprintf("Package %s installed successfully", packageName)
				results = append(results, string(cmdOutput))
			}
		} else if strings.HasPrefix(cmdDirective, "remove:") {
			packageName = strings.TrimPrefix(cmdDirective, "remove:")
			removeCmd := getRemoveCommand(packageManager, packageName)
			cmd := exec.Command("sh", "-c", removeCmd)
			cmdOutput, cmdErr := cmd.CombinedOutput()

			if cmdErr != nil {
				output.Status = "failed"
				output.Message = fmt.Sprintf("Failed to remove %s: %v", packageName, cmdErr)
				output.Facts["error"] = cmdErr.Error()
			} else {
				output.Status = "ok"
				output.Message = fmt.Sprintf("Package %s removed successfully", packageName)
				results = append(results, string(cmdOutput))
			}
		} else if strings.HasPrefix(cmdDirective, "upgrade:") {
			packageName = strings.TrimPrefix(cmdDirective, "upgrade:")
			upgradeCmd := getUpgradeCommand(packageManager, packageName)
			cmd := exec.Command("sh", "-c", upgradeCmd)
			cmdOutput, cmdErr := cmd.CombinedOutput()

			if cmdErr != nil {
				output.Status = "failed"
				output.Message = fmt.Sprintf("Failed to upgrade %s: %v", packageName, cmdErr)
				output.Facts["error"] = cmdErr.Error()
			} else {
				output.Status = "ok"
				output.Message = fmt.Sprintf("Package %s upgraded successfully", packageName)
				results = append(results, string(cmdOutput))
			}
		}
	}

	output.Facts["package_manager"] = packageManager
	output.Facts["package_name"] = packageName
	output.Facts["stdout"] = strings.Join(results, "\n")
	delete(output.Facts, "package_command")

	return output
}

func detectPackageManager() string {
	managers := []struct {
		name  string
		check string
	}{
		{"apt", "which apt-get"},
		{"dnf", "which dnf"},
		{"yum", "which yum"},
		{"pacman", "which pacman"},
		{"brew", "which brew"},
		{"choco", "where choco"},
		{"winget", "where winget"},
	}

	for _, mgr := range managers {
		cmd := exec.Command("sh", "-c", mgr.check)
		if err := cmd.Run(); err == nil {
			return mgr.name
		}
	}

	return ""
}

func getUpdateCacheCommand(pkgMgr string) string {
	switch pkgMgr {
	case "apt":
		return "sudo apt-get update"
	case "dnf":
		return "sudo dnf check-update || true"
	case "yum":
		return "sudo yum check-update || true"
	case "pacman":
		return "sudo pacman -Sy"
	case "brew":
		return "brew update"
	default:
		return ""
	}
}

func getInstallCommand(pkgMgr, packageName string) string {
	switch pkgMgr {
	case "apt":
		return fmt.Sprintf("sudo apt-get install -y %s", packageName)
	case "dnf":
		return fmt.Sprintf("sudo dnf install -y %s", packageName)
	case "yum":
		return fmt.Sprintf("sudo yum install -y %s", packageName)
	case "pacman":
		return fmt.Sprintf("sudo pacman -S --noconfirm %s", packageName)
	case "brew":
		return fmt.Sprintf("brew install %s", packageName)
	case "choco":
		return fmt.Sprintf("choco install -y %s", packageName)
	case "winget":
		return fmt.Sprintf("winget install --silent %s", packageName)
	default:
		return ""
	}
}

func getRemoveCommand(pkgMgr, packageName string) string {
	switch pkgMgr {
	case "apt":
		return fmt.Sprintf("sudo apt-get remove -y %s", packageName)
	case "dnf":
		return fmt.Sprintf("sudo dnf remove -y %s", packageName)
	case "yum":
		return fmt.Sprintf("sudo yum remove -y %s", packageName)
	case "pacman":
		return fmt.Sprintf("sudo pacman -R --noconfirm %s", packageName)
	case "brew":
		return fmt.Sprintf("brew uninstall %s", packageName)
	case "choco":
		return fmt.Sprintf("choco uninstall -y %s", packageName)
	case "winget":
		return fmt.Sprintf("winget uninstall --silent %s", packageName)
	default:
		return ""
	}
}

func getUpgradeCommand(pkgMgr, packageName string) string {
	switch pkgMgr {
	case "apt":
		return fmt.Sprintf("sudo apt-get install --only-upgrade -y %s", packageName)
	case "dnf":
		return fmt.Sprintf("sudo dnf upgrade -y %s", packageName)
	case "yum":
		return fmt.Sprintf("sudo yum upgrade -y %s", packageName)
	case "pacman":
		return fmt.Sprintf("sudo pacman -S --noconfirm %s", packageName)
	case "brew":
		return fmt.Sprintf("brew upgrade %s", packageName)
	case "choco":
		return fmt.Sprintf("choco upgrade -y %s", packageName)
	case "winget":
		return fmt.Sprintf("winget upgrade --silent %s", packageName)
	default:
		return ""
	}
}

func printOutput(output TaskOutput) {
	outputJSON, _ := json.MarshalIndent(output, "", "  ")
	fmt.Println(string(outputJSON))
}

func executeScript(output TaskOutput, scriptPath, scriptContent, scriptCmd string) TaskOutput {
	// Create script directory
	scriptDir := filepath.Dir(scriptPath)
	if err := os.MkdirAll(scriptDir, 0755); err != nil {
		output.Status = "failed"
		output.Message = fmt.Sprintf("Failed to create script directory: %v", err)
		output.Facts["error"] = err.Error()
		return output
	}

	// Write script content to file
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		output.Status = "failed"
		output.Message = fmt.Sprintf("Failed to write script file: %v", err)
		output.Facts["error"] = err.Error()
		return output
	}

	// Execute the script
	parts := strings.Fields(scriptCmd)
	var cmd *exec.Cmd
	if len(parts) > 1 {
		cmd = exec.Command(parts[0], parts[1:]...)
	} else {
		cmd = exec.Command(parts[0])
	}

	cmdOutput, cmdErr := cmd.CombinedOutput()

	if cmdErr != nil {
		output.Status = "failed"
		output.Message = fmt.Sprintf("Script execution failed: %v", cmdErr)
		output.Facts["stdout"] = string(cmdOutput)
		output.Facts["error"] = cmdErr.Error()
	} else {
		output.Status = "ok"
		output.Message = fmt.Sprintf("Script executed successfully: %s", scriptPath)
		output.Facts["stdout"] = strings.TrimSpace(string(cmdOutput))
	}

	// Clean up script facts that were just for execution
	delete(output.Facts, "script_command")
	delete(output.Facts, "script_content")

	return output
}

// hostExec is called by WASM modules to execute shell commands
func hostExec(cmd string, args []string, timeoutSecs int) (string, string, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSecs)*time.Second)
	defer cancel()

	command := exec.CommandContext(ctx, cmd, args...)

	var stdout, stderr strings.Builder
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
	}

	return stdout.String(), stderr.String(), exitCode, nil
}
