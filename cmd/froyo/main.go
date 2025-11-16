package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/piwi3910/openfroyo/internal/executor"
	"github.com/piwi3910/openfroyo/internal/parser"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: froyo apply <stack.ofy.yml>")
		os.Exit(1)
	}

	command := os.Args[1]
	if command != "apply" {
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Available commands: apply")
		os.Exit(1)
	}

	stackPath := os.Args[2]

	if err := runApply(stackPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
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
	runnerPath := "./froyo-runner"
	if _, err := os.Stat(runnerPath); os.IsNotExist(err) {
		// Try in same directory as froyo
		exePath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to locate froyo-runner binary")
		}
		runnerPath = filepath.Join(filepath.Dir(exePath), "froyo-runner")
		if _, err := os.Stat(runnerPath); os.IsNotExist(err) {
			return fmt.Errorf("froyo-runner binary not found (expected at %s)", runnerPath)
		}
	}

	// Execute stack
	exec := executor.NewExecutor(stack, inventory)
	return exec.Execute(runnerPath)
}
