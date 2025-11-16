package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/piwi3910/openfroyo/internal/executor"
	"github.com/piwi3910/openfroyo/internal/onboard"
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
		if len(os.Args) < 3 {
			fmt.Println("Usage: froyo apply <stack.ofy>")
			os.Exit(1)
		}
		stackPath := os.Args[2]
		if err := runApply(stackPath); err != nil {
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
	fmt.Println("  froyo apply <stack.ofy>                    - Execute a stack")
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

func runApply(stackPath string) error {
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
	return exec.Execute(runnerPath)
}
