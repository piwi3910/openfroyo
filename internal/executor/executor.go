package executor

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/piwi3910/openfroyo/internal/parser"
	sshclient "github.com/piwi3910/openfroyo/internal/ssh"
)

// Executor executes a stack
type Executor struct {
	stack     *parser.Stack
	inventory *parser.Inventory
}

// NewExecutor creates a new executor
func NewExecutor(stack *parser.Stack, inventory *parser.Inventory) *Executor {
	return &Executor{
		stack:     stack,
		inventory: inventory,
	}
}

// Execute runs the stack
func (e *Executor) Execute(runnerPath string) error {
	fmt.Printf("Executing stack: %s\n", e.stack.Name)
	fmt.Println()

	for i, entry := range e.stack.Run {
		fmt.Printf("[%d/%d] %s\n", i+1, len(e.stack.Run), entry.Name)

		if entry.Module == "" {
			return fmt.Errorf("module not specified for run entry: %s", entry.Name)
		}

		// Resolve hosts
		hosts := entry.Hosts
		if len(hosts) == 0 {
			return fmt.Errorf("no hosts specified for run entry: %s", entry.Name)
		}

		resolvedHosts, err := e.inventory.ResolveHosts(hosts)
		if err != nil {
			return fmt.Errorf("failed to resolve hosts: %w", err)
		}

		// Execute on each host
		for _, hostName := range resolvedHosts {
			if err := e.executeOnHost(hostName, entry, runnerPath); err != nil {
				return fmt.Errorf("failed on host %s: %w", hostName, err)
			}
		}

		fmt.Println()
	}

	fmt.Println("Stack execution completed successfully!")
	return nil
}

func (e *Executor) executeOnHost(hostName string, entry parser.RunEntry, runnerPath string) error {
	host, err := e.inventory.GetHost(hostName)
	if err != nil {
		return err
	}

	fmt.Printf("  → %s (%s)\n", hostName, host.SSHHost)

	// Connect via SSH
	client, err := sshclient.NewClient(host)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()

	// Ensure runner exists
	if err := client.EnsureRunner(runnerPath); err != nil {
		return fmt.Errorf("failed to ensure runner: %w", err)
	}

	// Resolve module path
	// Extract the module name (last part of the path for WASM filename)
	moduleNameParts := strings.Split(entry.Module, "/")
	moduleName := moduleNameParts[len(moduleNameParts)-1]
	modulePath := filepath.Join("modules", entry.Module, "wasm", moduleName+".wasm")

	// Merge vars (defaults + entry-specific)
	vars := make(map[string]interface{})
	for k, v := range e.stack.Defaults {
		vars[k] = v
	}
	for k, v := range entry.Vars {
		vars[k] = v
	}

	// Special handling for script module: read local script file
	if moduleName == "script" {
		if scriptPath, ok := vars["script"].(string); ok {
			if scriptContent, err := os.ReadFile(scriptPath); err == nil {
				// Base64 encode and pass to WASM module
				vars["_script_content"] = base64.StdEncoding.EncodeToString(scriptContent)
			}
		}
	}

	// Special handling for copy module: read local source file
	if moduleName == "copy" {
		if srcPath, ok := vars["src"].(string); ok {
			if fileContent, err := os.ReadFile(srcPath); err == nil {
				// Base64 encode and pass to WASM module
				vars["_file_content"] = base64.StdEncoding.EncodeToString(fileContent)
			} else {
				return fmt.Errorf("failed to read source file %s: %w", srcPath, err)
			}
		}
	}

	// Special handling for patch module: read local patch file
	if moduleName == "patch" {
		// Only read local file if remote_src is false (or not set)
		remoteSrc, _ := vars["remote_src"].(bool)
		if !remoteSrc {
			if srcPath, ok := vars["src"].(string); ok {
				if patchContent, err := os.ReadFile(srcPath); err == nil {
					// Base64 encode and pass to WASM module
					vars["_patch_content"] = base64.StdEncoding.EncodeToString(patchContent)
				} else {
					return fmt.Errorf("failed to read patch file %s: %w", srcPath, err)
				}
			}
		}
	}

	// Special handling for unarchive module: read local archive file
	if moduleName == "unarchive" {
		// Only read local file if remote_src is false (or not set)
		remoteSrc, _ := vars["remote_src"].(bool)
		if !remoteSrc {
			if srcPath, ok := vars["src"].(string); ok {
				if archiveContent, err := os.ReadFile(srcPath); err == nil {
					// Base64 encode and pass to WASM module
					vars["_archive_content"] = base64.StdEncoding.EncodeToString(archiveContent)
				} else {
					return fmt.Errorf("failed to read archive file %s: %w", srcPath, err)
				}
			}
		}
	}

	// Execute module
	output, err := client.ExecuteModule(modulePath, entry.Name, vars)
	if err != nil {
		return fmt.Errorf("failed to execute module: %w", err)
	}

	// Special handling for fetch module: write fetched file to local destination
	if moduleName == "fetch" && output.Status != "failed" {
		if fetchDest, ok := output.Facts["fetch_dest"].(string); ok {
			// The actual file content would come from shell execution results
			// For now, we just mark that fetch handling is needed
			fmt.Printf("      Fetch destination: %s\n", fetchDest)
			// TODO: Implement actual file writing from base64 decoded content
		}
	}

	// Display result
	statusIcon := "✓"
	if output.Status == "failed" {
		statusIcon = "✗"
	} else if output.Status == "changed" {
		statusIcon = "↻"
	}

	fmt.Printf("    %s %s: %s\n", statusIcon, output.Status, output.Message)

	// Display facts
	if len(output.Facts) > 0 {
		for key, value := range output.Facts {
			fmt.Printf("      %s: %v\n", key, value)
		}
	}

	if output.Status == "failed" {
		return fmt.Errorf("task failed: %s", output.Message)
	}

	return nil
}
