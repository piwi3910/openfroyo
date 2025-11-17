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
	fmt.Println("  froyo apply <stack.ofy> [--nats-server <url>] [--nats-token <token>]")
	fmt.Println("                                             - Execute a stack")
	fmt.Println("  froyo onboard --host <ip> --user <user> --password <pass> [--port <port>] [--inventory <file>]")
	fmt.Println("                                             - Set up openfroyo user on remote host")
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

func runApply(args []string) error {
	// Parse flags
	fs := flag.NewFlagSet("apply", flag.ExitOnError)
	natsServer := fs.String("nats-server", "", "NATS server URL (e.g., nats://localhost:4222)")
	natsToken := fs.String("nats-token", "", "NATS authentication token")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Println("Usage: froyo apply <stack.ofy> [--nats-server <url>] [--nats-token <token>]")
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

	// Execute stack
	exec := executor.NewExecutor(stack, inventory)

	// Set agent executor if agent mode is used
	if agentExec != nil {
		exec.SetAgentExecutor(agentExec)
	}

	return exec.Execute(runnerPath)
}
