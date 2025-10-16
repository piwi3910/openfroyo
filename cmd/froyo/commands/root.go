package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	configPath string
	verbose    bool
	jsonOutput bool
)

// Execute runs the root command
func Execute(ctx context.Context, version, commit, buildDate string) error {
	rootCmd := newRootCommand(version, commit, buildDate)
	return rootCmd.ExecuteContext(ctx)
}

func newRootCommand(version, commit, buildDate string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "froyo",
		Short: "OpenFroyo - Infrastructure Orchestration Engine",
		Long: `OpenFroyo is a next-generation Infrastructure-as-Code engine that combines
declarative state and planning with procedural configuration capabilities.

Features:
  - Typed configs via CUE
  - Light procedural scripting via Starlark
  - WASM-based provider system
  - Ephemeral micro-runner for complex local ops
  - Drift detection and state management
  - Policy enforcement`,
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, buildDate),
	}

	// Persistent flags available to all commands
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "config file path")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output in JSON format")

	// Add subcommands
	rootCmd.AddCommand(newInitCommand())
	rootCmd.AddCommand(newValidateCommand())
	rootCmd.AddCommand(newPlanCommand())
	rootCmd.AddCommand(newApplyCommand())
	rootCmd.AddCommand(newRunCommand())
	rootCmd.AddCommand(newDriftCommand())
	rootCmd.AddCommand(newOnboardCommand())
	rootCmd.AddCommand(newBackupCommand())
	rootCmd.AddCommand(newRestoreCommand())
	rootCmd.AddCommand(newDevCommand())
	rootCmd.AddCommand(newFactsCommand())

	return rootCmd
}
