package executor

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/piwi3910/openfroyo/internal/condition"
	orchestratorExecutor "github.com/piwi3910/openfroyo/internal/orchestrator/executor"
	"github.com/piwi3910/openfroyo/internal/parser"
	sshclient "github.com/piwi3910/openfroyo/internal/ssh"
	"github.com/piwi3910/openfroyo/internal/template"
)

// Executor executes a stack
type Executor struct {
	stack            *parser.Stack
	inventory        *parser.Inventory
	agentExecutor    *orchestratorExecutor.AgentExecutor
	logger           *log.Logger
	registeredVars   map[string]interface{} // Stores registered task results
	notifiedHandlers map[string]bool        // Tracks which handlers have been notified
	verbose          bool                   // Verbose output mode
	debug            bool                   // Debug output mode
}

// NewExecutor creates a new executor
func NewExecutor(stack *parser.Stack, inventory *parser.Inventory) *Executor {
	return &Executor{
		stack:            stack,
		inventory:        inventory,
		logger:           log.New(os.Stdout, "", 0),
		registeredVars:   make(map[string]interface{}),
		notifiedHandlers: make(map[string]bool),
	}
}

// SetVerbose enables verbose output
func (e *Executor) SetVerbose(verbose bool) {
	e.verbose = verbose
}

// SetDebug enables debug output
func (e *Executor) SetDebug(debug bool) {
	e.debug = debug
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
			if err := e.executeRunEntry(entry, runnerPath); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("run entry '%s' must have either 'module' or 'tasks' specified", entry.Name)
		}

		fmt.Println()
	}

	// Execute notified handlers
	if err := e.executeHandlers(runnerPath); err != nil {
		return fmt.Errorf("handler execution failed: %w", err)
	}

	fmt.Println("Stack execution completed successfully!")
	return nil
}

// executeRunEntry executes a single run entry with conditional and loop support
func (e *Executor) executeRunEntry(entry parser.RunEntry, runnerPath string) error {
	// Resolve hosts first
	hosts := entry.Hosts
	if len(hosts) == 0 {
		return fmt.Errorf("no hosts specified for run entry: %s", entry.Name)
	}

	resolvedHosts, err := e.inventory.ResolveHosts(hosts)
	if err != nil {
		return fmt.Errorf("failed to resolve hosts: %w", err)
	}

	// Check for loop
	loopItems, hasLoop := e.resolveLoop(entry.Loop)

	// Execute on each host
	for _, hostName := range resolvedHosts {
		host, err := e.inventory.GetHost(hostName)
		if err != nil {
			return err
		}

		// Build template context for condition evaluation
		vars := e.buildVars(entry.Vars)
		templateCtx := e.buildTemplateContext(hostName, host, vars)

		// Evaluate condition
		if entry.When != "" {
			shouldRun, err := e.evaluateCondition(entry.When, templateCtx)
			if err != nil {
				if e.debug {
					fmt.Printf("    [DEBUG] Condition evaluation error: %v\n", err)
				}
			}
			if !shouldRun {
				fmt.Printf("  → %s: skipped (when: %s)\n", hostName, entry.When)
				continue
			}
		}

		if hasLoop {
			// Execute with loop
			for idx, item := range loopItems {
				// Add item to vars
				loopVars := e.buildVars(entry.Vars)
				loopVars["item"] = item
				loopVars["item_index"] = idx

				if e.verbose {
					fmt.Printf("    Loop iteration %d: item=%v\n", idx+1, item)
				}

				result, err := e.executeOnHostWithVars(hostName, host, entry, loopVars, runnerPath)
				if err != nil {
					return fmt.Errorf("failed on host %s (loop item %d): %w", hostName, idx, err)
				}

				// Handle registration and notification for loop item
				e.handleTaskResult(entry, result, hostName, idx)
			}
		} else {
			// Execute without loop
			result, err := e.executeOnHostWithVars(hostName, host, entry, vars, runnerPath)
			if err != nil {
				return fmt.Errorf("failed on host %s: %w", hostName, err)
			}

			// Handle registration and notification
			e.handleTaskResult(entry, result, hostName, -1)
		}
	}

	return nil
}

// buildVars builds the vars map from defaults and entry vars
func (e *Executor) buildVars(entryVars map[string]interface{}) map[string]interface{} {
	vars := make(map[string]interface{})
	for k, v := range e.stack.Defaults {
		vars[k] = v
	}
	for k, v := range entryVars {
		vars[k] = v
	}
	// Add registered variables
	for k, v := range e.registeredVars {
		vars[k] = v
	}
	return vars
}

// buildTemplateContext builds the template context for a host
func (e *Executor) buildTemplateContext(hostName string, host parser.Host, vars map[string]interface{}) *template.Context {
	mode := host.Mode
	if mode == "" {
		mode = "ssh"
	}
	return &template.Context{
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
}

// evaluateCondition evaluates a when condition
func (e *Executor) evaluateCondition(when string, ctx *template.Context) (bool, error) {
	evaluator := condition.NewEvaluator(ctx)
	return evaluator.Evaluate(when)
}

// resolveLoop resolves the loop value to a list of items
func (e *Executor) resolveLoop(loop interface{}) ([]interface{}, bool) {
	if loop == nil {
		return nil, false
	}

	switch l := loop.(type) {
	case []interface{}:
		return l, true
	case []string:
		result := make([]interface{}, len(l))
		for i, s := range l {
			result[i] = s
		}
		return result, true
	case string:
		// Check if it's a template expression
		if strings.HasPrefix(l, "{{") && strings.HasSuffix(l, "}}") {
			// Extract variable name and look up in registered vars
			expr := strings.TrimSpace(l[2 : len(l)-2])
			if strings.HasPrefix(expr, "var.") || strings.HasPrefix(expr, "vars.") {
				varName := strings.TrimPrefix(strings.TrimPrefix(expr, "vars."), "var.")
				if val, ok := e.stack.Defaults[varName]; ok {
					if items, ok := val.([]interface{}); ok {
						return items, true
					}
				}
				if val, ok := e.registeredVars[varName]; ok {
					if items, ok := val.([]interface{}); ok {
						return items, true
					}
				}
			}
		}
		// Single item loop
		return []interface{}{l}, true
	default:
		return nil, false
	}
}

// handleTaskResult handles registration and notification for a task result
func (e *Executor) handleTaskResult(entry parser.RunEntry, result *TaskResult, hostName string, loopIdx int) {
	if result == nil {
		return
	}

	// Register result if requested
	if entry.Register != "" {
		regName := entry.Register
		if loopIdx >= 0 {
			// For loops, store results in a list
			listName := regName + "_results"
			if _, ok := e.registeredVars[listName]; !ok {
				e.registeredVars[listName] = []interface{}{}
			}
			list := e.registeredVars[listName].([]interface{})
			e.registeredVars[listName] = append(list, map[string]interface{}{
				"status":  result.Status,
				"message": result.Message,
				"facts":   result.Facts,
				"changed": result.Changed,
				"host":    hostName,
				"index":   loopIdx,
			})
			// Also store the last result
			e.registeredVars[regName] = map[string]interface{}{
				"status":  result.Status,
				"message": result.Message,
				"facts":   result.Facts,
				"changed": result.Changed,
			}
		} else {
			e.registeredVars[regName] = map[string]interface{}{
				"status":  result.Status,
				"message": result.Message,
				"facts":   result.Facts,
				"changed": result.Changed,
			}
		}
		if e.debug {
			fmt.Printf("    [DEBUG] Registered: %s = {status: %s, changed: %v}\n", regName, result.Status, result.Changed)
		}
	}

	// Notify handlers if task changed
	if result.Changed && len(entry.Notify) > 0 {
		for _, handlerName := range entry.Notify {
			e.notifiedHandlers[handlerName] = true
			if e.verbose {
				fmt.Printf("    Notified handler: %s\n", handlerName)
			}
		}
	}
}

// TaskResult represents the result of a task execution
type TaskResult struct {
	Status  string
	Message string
	Facts   map[string]interface{}
	Changed bool
}

// executeOnHostWithVars executes on a host with pre-built vars
func (e *Executor) executeOnHostWithVars(hostName string, host parser.Host, entry parser.RunEntry, vars map[string]interface{}, runnerPath string) (*TaskResult, error) {
	mode := host.Mode
	if mode == "" {
		mode = "ssh"
	}

	switch mode {
	case "agent":
		return e.executeOnAgentWithResult(hostName, host, entry, vars)
	case "ssh":
		return e.executeOnSSHWithResult(hostName, host, entry, vars, runnerPath)
	default:
		return nil, fmt.Errorf("unknown execution mode: %s", mode)
	}
}

// executeHandlers runs all notified handlers
func (e *Executor) executeHandlers(runnerPath string) error {
	if len(e.notifiedHandlers) == 0 {
		return nil
	}

	fmt.Println()
	fmt.Println("Running handlers:")
	fmt.Println(strings.Repeat("-", 40))

	for _, handler := range e.stack.Handlers {
		if !e.notifiedHandlers[handler.Name] {
			continue
		}

		fmt.Printf("  Handler: %s\n", handler.Name)

		// Resolve hosts (use handler's hosts or all hosts)
		hosts := handler.Hosts
		if len(hosts) == 0 {
			// Get all hosts from the stack's first run entry as default
			for _, h := range e.inventory.Hosts {
				_ = h // Use all inventory hosts
			}
			// For simplicity, require hosts to be specified in handler
			fmt.Printf("    WARNING: No hosts specified for handler %s, skipping\n", handler.Name)
			continue
		}

		resolvedHosts, err := e.inventory.ResolveHosts(hosts)
		if err != nil {
			return fmt.Errorf("failed to resolve handler hosts: %w", err)
		}

		// Create a RunEntry from the handler
		handlerEntry := parser.RunEntry{
			Name:   handler.Name,
			Module: handler.Module,
			Hosts:  handler.Hosts,
			Vars:   handler.Vars,
		}

		// Execute on each host
		for _, hostName := range resolvedHosts {
			host, err := e.inventory.GetHost(hostName)
			if err != nil {
				return err
			}

			vars := e.buildVars(handler.Vars)
			_, err = e.executeOnHostWithVars(hostName, host, handlerEntry, vars, runnerPath)
			if err != nil {
				return fmt.Errorf("handler %s failed on %s: %w", handler.Name, hostName, err)
			}
		}
	}

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

// executeOnSSHWithResult executes on SSH and returns a TaskResult
func (e *Executor) executeOnSSHWithResult(hostName string, host parser.Host, entry parser.RunEntry, vars map[string]interface{}, runnerPath string) (*TaskResult, error) {
	fmt.Printf("  → %s (%s) [SSH]\n", hostName, host.SSHHost)

	if e.debug {
		fmt.Printf("    [DEBUG] Connecting to %s:%d as %s\n", host.SSHHost, host.SSHPort, host.SSHUser)
	}

	// Connect via SSH
	client, err := sshclient.NewClient(host)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()

	// Ensure runner exists
	if err := client.EnsureRunner(runnerPath); err != nil {
		return nil, fmt.Errorf("failed to ensure runner: %w", err)
	}

	// Resolve module path
	moduleNameParts := strings.Split(entry.Module, "/")
	moduleName := moduleNameParts[len(moduleNameParts)-1]
	modulePath := filepath.Join("modules", entry.Module, "wasm", moduleName+".wasm")

	if e.debug {
		fmt.Printf("    [DEBUG] Module: %s -> %s\n", entry.Module, modulePath)
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
		return nil, fmt.Errorf("failed to process template variables: %w", err)
	}

	// Handle special modules that need local file reading
	vars = e.handleSpecialModules(moduleName, vars)

	if e.verbose {
		fmt.Printf("    Variables: %v\n", vars)
	}

	// Execute module
	output, err := client.ExecuteModule(modulePath, entry.Name, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to execute module: %w", err)
	}

	// Display result
	statusIcon := "✓"
	changed := false
	if output.Status == "failed" {
		statusIcon = "✗"
	} else if output.Status == "changed" {
		statusIcon = "↻"
		changed = true
	}

	fmt.Printf("    %s %s: %s\n", statusIcon, output.Status, output.Message)

	// Display facts in verbose mode
	if e.verbose && len(output.Facts) > 0 {
		for key, value := range output.Facts {
			fmt.Printf("      %s: %v\n", key, value)
		}
	}

	result := &TaskResult{
		Status:  output.Status,
		Message: output.Message,
		Facts:   output.Facts,
		Changed: changed,
	}

	if output.Status == "failed" {
		return result, fmt.Errorf("task failed: %s", output.Message)
	}

	return result, nil
}

// executeOnAgentWithResult executes on agent and returns a TaskResult
func (e *Executor) executeOnAgentWithResult(hostName string, host parser.Host, entry parser.RunEntry, vars map[string]interface{}) (*TaskResult, error) {
	// Check if agent executor is configured
	if e.agentExecutor == nil {
		return nil, fmt.Errorf("agent mode requested but NATS not configured (use --nats-server flag)")
	}

	// Validate agent configuration
	if host.AgentID == "" {
		return nil, fmt.Errorf("agent_id is required for agent mode")
	}
	if host.Datacenter == "" {
		return nil, fmt.Errorf("datacenter is required for agent mode")
	}

	fmt.Printf("  → %s (agent: %s) [AGENT]\n", hostName, host.AgentID)

	if e.debug {
		fmt.Printf("    [DEBUG] Datacenter: %s, AgentID: %s\n", host.Datacenter, host.AgentID)
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
		return nil, fmt.Errorf("failed to process template variables: %w", err)
	}

	if e.verbose {
		fmt.Printf("    Variables: %v\n", vars)
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
		return nil, fmt.Errorf("failed to execute on agent: %w", err)
	}

	// Display result
	statusIcon := "✓"
	changed := false
	if result.Status == "failed" {
		statusIcon = "✗"
	} else if result.Status == "changed" {
		statusIcon = "↻"
		changed = true
	}

	fmt.Printf("    %s %s: %s\n", statusIcon, result.Status, result.Message)

	// Display facts in verbose mode
	if e.verbose && len(result.Facts) > 0 {
		for key, value := range result.Facts {
			fmt.Printf("      %s: %v\n", key, value)
		}
	}

	taskResult := &TaskResult{
		Status:  result.Status,
		Message: result.Message,
		Facts:   result.Facts,
		Changed: changed,
	}

	if result.Status == "failed" {
		return taskResult, fmt.Errorf("task failed: %s", result.Message)
	}

	return taskResult, nil
}

// handleSpecialModules handles special module processing (file reading, etc.)
func (e *Executor) handleSpecialModules(moduleName string, vars map[string]interface{}) map[string]interface{} {
	// Special handling for script module: read local script file
	if moduleName == "script" {
		if scriptPath, ok := vars["script"].(string); ok {
			if scriptContent, err := os.ReadFile(scriptPath); err == nil {
				vars["_script_content"] = base64.StdEncoding.EncodeToString(scriptContent)
			}
		}
	}

	// Special handling for copy module: read local source file
	if moduleName == "copy" {
		if srcPath, ok := vars["src"].(string); ok {
			if fileContent, err := os.ReadFile(srcPath); err == nil {
				vars["_file_content"] = base64.StdEncoding.EncodeToString(fileContent)
			}
		}
	}

	// Special handling for patch module: read local patch file
	if moduleName == "patch" {
		remoteSrc, _ := vars["remote_src"].(bool)
		if !remoteSrc {
			if srcPath, ok := vars["src"].(string); ok {
				if patchContent, err := os.ReadFile(srcPath); err == nil {
					vars["_patch_content"] = base64.StdEncoding.EncodeToString(patchContent)
				}
			}
		}
	}

	// Special handling for unarchive module: read local archive file
	if moduleName == "unarchive" {
		remoteSrc, _ := vars["remote_src"].(bool)
		if !remoteSrc {
			if srcPath, ok := vars["src"].(string); ok {
				if archiveContent, err := os.ReadFile(srcPath); err == nil {
					vars["_archive_content"] = base64.StdEncoding.EncodeToString(archiveContent)
				}
			}
		}
	}

	return vars
}
