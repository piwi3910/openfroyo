package executor

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	orchestratorExecutor "github.com/piwi3910/openfroyo/internal/orchestrator/executor"
	"github.com/piwi3910/openfroyo/internal/parser"
	sshclient "github.com/piwi3910/openfroyo/internal/ssh"
	"github.com/piwi3910/openfroyo/internal/template"
)

// Executor executes a stack
type Executor struct {
	stack         *parser.Stack
	inventory     *parser.Inventory
	agentExecutor *orchestratorExecutor.AgentExecutor
	logger        *log.Logger
}

// NewExecutor creates a new executor
func NewExecutor(stack *parser.Stack, inventory *parser.Inventory) *Executor {
	return &Executor{
		stack:     stack,
		inventory: inventory,
		logger:    log.New(os.Stdout, "", 0),
	}
}

// SetAgentExecutor sets the agent executor for agent mode hosts
func (e *Executor) SetAgentExecutor(agentExec *orchestratorExecutor.AgentExecutor) {
	e.agentExecutor = agentExec
}

// Execute runs the stack
func (e *Executor) Execute(runnerPath string) error {
	fmt.Printf("Executing stack: %s\n", e.stack.Name)
	fmt.Println()

	for i, entry := range e.stack.Run {
		fmt.Printf("[%d/%d] %s\n", i+1, len(e.stack.Run), entry.Name)

		// Check if this is a task block or module invocation
		if len(entry.Tasks) > 0 {
			// Task block: execute inline tasks
			if err := e.executeTaskBlock(entry, runnerPath); err != nil {
				return err
			}
		} else if entry.Module != "" {
			// Module invocation
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
		} else {
			return fmt.Errorf("run entry '%s' must have either 'module' or 'tasks' specified", entry.Name)
		}

		fmt.Println()
	}

	fmt.Println("Stack execution completed successfully!")
	return nil
}

// executeTaskBlock executes an inline task block
func (e *Executor) executeTaskBlock(entry parser.RunEntry, runnerPath string) error {
	// Resolve hosts
	hosts := entry.Hosts
	if len(hosts) == 0 {
		return fmt.Errorf("no hosts specified for task block: %s", entry.Name)
	}

	resolvedHosts, err := e.inventory.ResolveHosts(hosts)
	if err != nil {
		return fmt.Errorf("failed to resolve hosts: %w", err)
	}

	// Execute each task in the block on each host
	for _, hostName := range resolvedHosts {
		host, err := e.inventory.GetHost(hostName)
		if err != nil {
			return err
		}

		mode := host.Mode
		if mode == "" {
			mode = "ssh"
		}

		for j, task := range entry.Tasks {
			fmt.Printf("  Task [%d/%d] %s\n", j+1, len(entry.Tasks), task.Name)

			// Create a temporary RunEntry for this task
			taskEntry := parser.RunEntry{
				Name:   task.Name,
				Module: task.Module,
				Hosts:  []string{hostName},
				Vars:   make(map[string]interface{}),
			}

			// Merge vars: defaults -> entry vars -> task vars
			for k, v := range e.stack.Defaults {
				taskEntry.Vars[k] = v
			}
			for k, v := range entry.Vars {
				taskEntry.Vars[k] = v
			}
			for k, v := range task.Vars {
				taskEntry.Vars[k] = v
			}

			// Execute based on mode
			switch mode {
			case "agent":
				if err := e.executeOnAgent(hostName, host, taskEntry); err != nil {
					return fmt.Errorf("task '%s' failed on host %s: %w", task.Name, hostName, err)
				}
			case "ssh":
				if err := e.executeOnSSH(hostName, host, taskEntry, runnerPath); err != nil {
					return fmt.Errorf("task '%s' failed on host %s: %w", task.Name, hostName, err)
				}
			default:
				return fmt.Errorf("unknown execution mode: %s", mode)
			}
		}
	}

	return nil
}

func (e *Executor) executeOnHost(hostName string, entry parser.RunEntry, runnerPath string) error {
	host, err := e.inventory.GetHost(hostName)
	if err != nil {
		return err
	}

	// Determine execution mode (default to SSH if not specified)
	mode := host.Mode
	if mode == "" {
		mode = "ssh"
	}

	// Route to appropriate executor based on mode
	switch mode {
	case "agent":
		return e.executeOnAgent(hostName, host, entry)
	case "ssh":
		return e.executeOnSSH(hostName, host, entry, runnerPath)
	default:
		return fmt.Errorf("unknown execution mode: %s", mode)
	}
}

func (e *Executor) executeOnSSH(hostName string, host parser.Host, entry parser.RunEntry, runnerPath string) error {
	fmt.Printf("  → %s (%s) [SSH]\n", hostName, host.SSHHost)

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

	// Process template expressions in variables
	mode := host.Mode
	if mode == "" {
		mode = "ssh"
	}
	templateCtx := &template.Context{
		Vars: vars,
		Host: template.HostContext{
			Name:       hostName,
			SSHHost:    host.SSHHost,
			SSHPort:    host.SSHPort,
			SSHUser:    host.SSHUser,
			Datacenter: host.Datacenter,
			AgentID:    host.AgentID,
			Mode:       mode,
		},
	}
	vars, err = template.ProcessVars(vars, templateCtx)
	if err != nil {
		return fmt.Errorf("failed to process template variables: %w", err)
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

func (e *Executor) executeOnAgent(hostName string, host parser.Host, entry parser.RunEntry) error {
	// Check if agent executor is configured
	if e.agentExecutor == nil {
		return fmt.Errorf("agent mode requested but NATS not configured (use --nats-server flag)")
	}

	// Validate agent configuration
	if host.AgentID == "" {
		return fmt.Errorf("agent_id is required for agent mode")
	}
	if host.Datacenter == "" {
		return fmt.Errorf("datacenter is required for agent mode")
	}

	fmt.Printf("  → %s (agent: %s) [AGENT]\n", hostName, host.AgentID)

	// Merge vars (defaults + entry-specific)
	vars := make(map[string]interface{})
	for k, v := range e.stack.Defaults {
		vars[k] = v
	}
	for k, v := range entry.Vars {
		vars[k] = v
	}

	// Process template expressions in variables
	templateCtx := &template.Context{
		Vars: vars,
		Host: template.HostContext{
			Name:       hostName,
			SSHHost:    host.SSHHost,
			SSHPort:    host.SSHPort,
			SSHUser:    host.SSHUser,
			Datacenter: host.Datacenter,
			AgentID:    host.AgentID,
			Mode:       "agent",
		},
	}
	var err error
	vars, err = template.ProcessVars(vars, templateCtx)
	if err != nil {
		return fmt.Errorf("failed to process template variables: %w", err)
	}

	// Execute task on agent via NATS
	result, err := e.agentExecutor.ExecuteTask(
		host.Datacenter,
		host.AgentID,
		entry.Name,
		entry.Module,
		vars,
		300, // 5 minute timeout
	)

	if err != nil {
		return fmt.Errorf("failed to execute on agent: %w", err)
	}

	// Display result
	statusIcon := "✓"
	if result.Status == "failed" {
		statusIcon = "✗"
	} else if result.Status == "changed" {
		statusIcon = "↻"
	}

	fmt.Printf("    %s %s: %s\n", statusIcon, result.Status, result.Message)

	// Display facts
	if len(result.Facts) > 0 {
		for key, value := range result.Facts {
			fmt.Printf("      %s: %v\n", key, value)
		}
	}

	if result.Status == "failed" {
		return fmt.Errorf("task failed: %s", result.Message)
	}

	return nil
}
