package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/piwi3910/openfroyo/internal/executor"
	"github.com/piwi3910/openfroyo/internal/onboard"
	orchestratorExecutor "github.com/piwi3910/openfroyo/internal/orchestrator/executor"
	orchestratorNats "github.com/piwi3910/openfroyo/internal/orchestrator/nats"
	"github.com/piwi3910/openfroyo/internal/parser"
	"github.com/piwi3910/openfroyo/internal/validator"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "apply":
		if err := runApply(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "plan":
		if err := runPlan(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "validate":
		if err := runValidate(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "onboard":
		if err := runOnboard(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  froyo apply <stack.ofy> [options]          - Execute a stack")
	fmt.Println("  froyo plan <stack.ofy> [options]           - Show execution plan (dry run)")
	fmt.Println("  froyo validate <stack.ofy>                 - Validate a stack file")
	fmt.Println("  froyo onboard [options]                    - Set up openfroyo user on remote host")
	fmt.Println()
	fmt.Println("Apply/Plan options:")
	fmt.Println("  --nats-server <url>    NATS server URL for agent mode")
	fmt.Println("  --nats-token <token>   NATS authentication token")
	fmt.Println("  --from <name>          Start execution from named entry")
	fmt.Println("  --until <name>         Stop execution after named entry")
	fmt.Println("  --verbose, -v          Show detailed task execution info")
	fmt.Println("  --debug, -d            Show full execution trace")
	fmt.Println()
	fmt.Println("Onboard options:")
	fmt.Println("  --host <ip>            Remote host IP or hostname (required)")
	fmt.Println("  --user <user>          SSH username (required)")
	fmt.Println("  --password <pass>      SSH password (required)")
	fmt.Println("  --port <port>          SSH port (default: 22)")
	fmt.Println("  --inventory <file>     Inventory file to update")
}

func runValidate(args []string) error {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Println("Usage: froyo validate <stack.ofy>")
		os.Exit(1)
	}

	stackPath := fs.Arg(0)

	// Create validator with default module directory
	v := validator.NewValidator("modules")

	// Run validation
	result, err := v.Validate(stackPath)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Print results
	validator.PrintResult(result, stackPath)

	if !result.Valid {
		os.Exit(1)
	}

	return nil
}

func runOnboard(args []string) error {
	fs := flag.NewFlagSet("onboard", flag.ExitOnError)
	host := fs.String("host", "", "Remote host IP or hostname")
	user := fs.String("user", "", "SSH username")
	password := fs.String("password", "", "SSH password")
	port := fs.Int("port", 22, "SSH port (default: 22)")
	inventory := fs.String("inventory", "", "Inventory file to update (optional)")

	fs.Parse(args)

	if *host == "" || *user == "" || *password == "" {
		fmt.Println("Error: --host, --user, and --password are required")
		fmt.Println("\nUsage: froyo onboard --host <ip> --user <user> --password <pass> [--port <port>] [--inventory <file>]")
		os.Exit(1)
	}

	return onboard.OnboardHost(*host, *port, *user, *password, *inventory)
}

func runPlan(args []string) error {
	// Parse flags
	fs := flag.NewFlagSet("plan", flag.ExitOnError)
	fromEntry := fs.String("from", "", "Start from named entry")
	untilEntry := fs.String("until", "", "Stop after named entry")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Println("Usage: froyo plan <stack.ofy> [--from <name>] [--until <name>]")
		os.Exit(1)
	}

	stackPath := fs.Arg(0)

	// Load stack
	fmt.Printf("Planning stack: %s\n", stackPath)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	stack, err := parser.LoadStack(stackPath)
	if err != nil {
		return fmt.Errorf("failed to load stack: %w", err)
	}

	// Load inventory
	inventory, err := parser.LoadInventory(stack.Inventory)
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	// Filter entries based on --from and --until
	entries, err := filterEntries(stack.Run, *fromEntry, *untilEntry)
	if err != nil {
		return err
	}

	// Print plan summary
	fmt.Printf("Stack: %s\n", stack.Name)
	fmt.Printf("Inventory: %v\n", stack.Inventory)
	fmt.Printf("Total hosts: %d\n", len(inventory.Hosts))
	fmt.Printf("Total groups: %d\n", len(inventory.Groups))
	fmt.Printf("Run entries: %d (of %d total)\n", len(entries), len(stack.Run))
	fmt.Println()

	if *fromEntry != "" || *untilEntry != "" {
		fmt.Println("Partial execution:")
		if *fromEntry != "" {
			fmt.Printf("  Starting from: %s\n", *fromEntry)
		}
		if *untilEntry != "" {
			fmt.Printf("  Stopping after: %s\n", *untilEntry)
		}
		fmt.Println()
	}

	// Print defaults if any
	if len(stack.Defaults) > 0 {
		fmt.Println("Default variables:")
		for k, v := range stack.Defaults {
			fmt.Printf("  %s: %v\n", k, v)
		}
		fmt.Println()
	}

	// Print execution plan
	fmt.Println("Execution plan:")
	fmt.Println(strings.Repeat("-", 60))

	for i, entry := range entries {
		fmt.Printf("\n[%d/%d] %s\n", i+1, len(entries), entry.Name)
		fmt.Printf("      Module: %s\n", entry.Module)

		// Handle task blocks
		if len(entry.Tasks) > 0 {
			fmt.Printf("      Tasks: %d inline tasks\n", len(entry.Tasks))
			for j, task := range entry.Tasks {
				fmt.Printf("        [%d] %s (module: %s)\n", j+1, task.Name, task.Module)
			}
		}

		// Resolve hosts
		if len(entry.Hosts) > 0 {
			resolvedHosts, err := inventory.ResolveHosts(entry.Hosts)
			if err != nil {
				fmt.Printf("      Hosts: ERROR - %v\n", err)
			} else {
				fmt.Printf("      Hosts: %d (%s)\n", len(resolvedHosts), strings.Join(entry.Hosts, ", "))
				for _, hostName := range resolvedHosts {
					host, _ := inventory.GetHost(hostName)
					mode := host.Mode
					if mode == "" {
						mode = "ssh"
					}
					if mode == "ssh" {
						fmt.Printf("        - %s (%s:%d) [SSH]\n", hostName, host.SSHHost, host.SSHPort)
					} else {
						fmt.Printf("        - %s (agent: %s, dc: %s) [AGENT]\n", hostName, host.AgentID, host.Datacenter)
					}
				}
			}
		}

		// Show variables
		if len(entry.Vars) > 0 {
			fmt.Println("      Variables:")
			for k, v := range entry.Vars {
				fmt.Printf("        %s: %v\n", k, v)
			}
		}

		// Show strategy
		strategy := entry.Strategy
		if strategy == "" {
			strategy = "serial"
		}
		fmt.Printf("      Strategy: %s\n", strategy)
	}

	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("Plan complete. No changes made.")
	fmt.Printf("Run 'froyo apply %s' to execute this stack.\n", stackPath)

	return nil
}

// filterEntries filters run entries based on --from and --until flags
func filterEntries(entries []parser.RunEntry, fromEntry, untilEntry string) ([]parser.RunEntry, error) {
	if fromEntry == "" && untilEntry == "" {
		return entries, nil
	}

	startIdx := 0
	endIdx := len(entries) - 1

	// Find start index
	if fromEntry != "" {
		found := false
		for i, entry := range entries {
			if entry.Name == fromEntry {
				startIdx = i
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("entry not found: %s", fromEntry)
		}
	}

	// Find end index
	if untilEntry != "" {
		found := false
		for i, entry := range entries {
			if entry.Name == untilEntry {
				endIdx = i
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("entry not found: %s", untilEntry)
		}
	}

	if startIdx > endIdx {
		return nil, fmt.Errorf("--from entry '%s' comes after --until entry '%s'", fromEntry, untilEntry)
	}

	return entries[startIdx : endIdx+1], nil
}

func runApply(args []string) error {
	// Parse flags
	fs := flag.NewFlagSet("apply", flag.ExitOnError)
	natsServer := fs.String("nats-server", "", "NATS server URL (e.g., nats://localhost:4222)")
	natsToken := fs.String("nats-token", "", "NATS authentication token")
	fromEntry := fs.String("from", "", "Start from named entry")
	untilEntry := fs.String("until", "", "Stop after named entry")
	verbose := fs.Bool("verbose", false, "Show detailed task execution info")
	verboseShort := fs.Bool("v", false, "Show detailed task execution info (short)")
	debug := fs.Bool("debug", false, "Show full execution trace")
	debugShort := fs.Bool("d", false, "Show full execution trace (short)")
	fs.Parse(args)

	// Combine short and long flags
	isVerbose := *verbose || *verboseShort
	isDebug := *debug || *debugShort
	// Debug implies verbose
	if isDebug {
		isVerbose = true
	}

	if fs.NArg() < 1 {
		fmt.Println("Usage: froyo apply <stack.ofy> [--nats-server <url>] [--nats-token <token>] [--from <name>] [--until <name>]")
		os.Exit(1)
	}

	stackPath := fs.Arg(0)

	// Load stack
	fmt.Printf("Loading stack: %s\n", stackPath)
	stack, err := parser.LoadStack(stackPath)
	if err != nil {
		return fmt.Errorf("failed to load stack: %w", err)
	}

	// Load inventory
	fmt.Println("Loading inventory...")
	inventory, err := parser.LoadInventory(stack.Inventory)
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	fmt.Printf("Loaded %d hosts, %d groups\n\n", len(inventory.Hosts), len(inventory.Groups))

	// Check if any hosts use agent mode
	hasAgentMode := false
	for _, host := range inventory.Hosts {
		if host.Mode == "agent" {
			hasAgentMode = true
			break
		}
	}

	// Initialize NATS if agent mode is used
	var agentExec *orchestratorExecutor.AgentExecutor
	var natsClient *orchestratorNats.Client
	if hasAgentMode {
		if *natsServer == "" {
			return fmt.Errorf("NATS server required for agent mode hosts (use --nats-server)")
		}

		fmt.Printf("Connecting to NATS: %s\n", *natsServer)
		logger := log.New(os.Stdout, "[nats] ", 0)

		natsConfig := &orchestratorNats.Config{
			Servers: strings.Split(*natsServer, ","),
			Timeout: 5 * time.Second,
			Token:   *natsToken,
		}

		natsClient, err = orchestratorNats.NewClient(natsConfig, logger)
		if err != nil {
			return fmt.Errorf("failed to create NATS client: %w", err)
		}

		if err := natsClient.Connect(); err != nil {
			return fmt.Errorf("failed to connect to NATS: %w", err)
		}
		defer natsClient.Close()

		// Create agent executor
		agentExec = orchestratorExecutor.NewAgentExecutor(natsClient, logger)

		// Initialize for each datacenter
		datacenters := make(map[string]bool)
		for _, host := range inventory.Hosts {
			if host.Mode == "agent" && host.Datacenter != "" {
				datacenters[host.Datacenter] = true
			}
		}

		for dc := range datacenters {
			if err := agentExec.Initialize(dc); err != nil {
				return fmt.Errorf("failed to initialize agent executor for %s: %w", dc, err)
			}
		}

		fmt.Println()
	}

	// Find froyo-runner binary
	// Try bin/froyo-runner first (when running from project root)
	runnerPath := "bin/froyo-runner"
	if _, err := os.Stat(runnerPath); os.IsNotExist(err) {
		// Try in same directory as froyo executable
		exePath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to locate froyo-runner binary")
		}
		runnerPath = filepath.Join(filepath.Dir(exePath), "froyo-runner")
		if _, err := os.Stat(runnerPath); os.IsNotExist(err) {
			return fmt.Errorf("froyo-runner binary not found (expected at bin/froyo-runner or %s)", runnerPath)
		}
	}

	// Filter entries based on --from and --until
	if *fromEntry != "" || *untilEntry != "" {
		filteredEntries, err := filterEntries(stack.Run, *fromEntry, *untilEntry)
		if err != nil {
			return err
		}
		stack.Run = filteredEntries
		fmt.Printf("Partial execution: %d entries selected\n", len(filteredEntries))
		if *fromEntry != "" {
			fmt.Printf("  Starting from: %s\n", *fromEntry)
		}
		if *untilEntry != "" {
			fmt.Printf("  Stopping after: %s\n", *untilEntry)
		}
		fmt.Println()
	}

	// Execute stack
	exec := executor.NewExecutor(stack, inventory)

	// Set verbose and debug modes
	exec.SetVerbose(isVerbose)
	exec.SetDebug(isDebug)

	// Set agent executor if agent mode is used
	if agentExec != nil {
		exec.SetAgentExecutor(agentExec)
	}

	return exec.Execute(runnerPath)
}
