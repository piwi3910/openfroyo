package commands

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func newDevCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dev",
		Short: "Development mode commands",
		Long: `Commands for local development and testing.

These commands help developers run and test OpenFroyo components locally.`,
	}

	cmd.AddCommand(newDevUpCommand())
	cmd.AddCommand(newDevDownCommand())

	return cmd
}

func newDevUpCommand() *cobra.Command {
	var (
		controllerOnly bool
		workerOnly     bool
		workers        int
	)

	cmd := &cobra.Command{
		Use:   "up",
		Short: "Start controller and worker locally",
		Long: `Start OpenFroyo controller and worker processes for local development.

This command runs:
  - Controller: Manages plans, state, and coordination
  - Worker(s): Execute plan units and provider operations

Both components run in-process with shared SQLite database and queue.`,
		Example: `  # Start both controller and worker
  froyo dev up

  # Start controller only
  froyo dev up --controller-only

  # Start worker only (assumes controller running)
  froyo dev up --worker-only

  # Start with multiple workers
  froyo dev up --workers 3`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info().
				Bool("controller_only", controllerOnly).
				Bool("worker_only", workerOnly).
				Int("workers", workers).
				Msg("Starting dev environment")

			// TODO: Implement dev mode
			// - Initialize SQLite database
			// - Start embedded queue
			// - If controller or both:
			//   - Start controller goroutine
			//   - Listen for API requests
			// - If worker or both:
			//   - Start worker goroutine(s)
			//   - Poll queue for tasks
			// - Setup graceful shutdown
			// - Handle signals (Ctrl+C)

			fmt.Println("Not implemented yet: dev environment")
			fmt.Printf("Would start dev mode: controller_only=%v, worker_only=%v, workers=%d\n",
				controllerOnly, workerOnly, workers)

			return nil
		},
	}

	cmd.Flags().BoolVar(&controllerOnly, "controller-only", false, "start controller only")
	cmd.Flags().BoolVar(&workerOnly, "worker-only", false, "start worker only")
	cmd.Flags().IntVar(&workers, "workers", 1, "number of worker processes")

	return cmd
}

func newDevDownCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "down",
		Short: "Stop local dev environment",
		Long: `Stop locally running controller and worker processes.

This sends a graceful shutdown signal to running dev processes.`,
		Example: `  # Stop dev environment
  froyo dev down`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info().Msg("Stopping dev environment")

			// TODO: Implement dev shutdown
			// - Find running dev processes (PID file)
			// - Send SIGTERM signal
			// - Wait for graceful shutdown
			// - Cleanup resources

			fmt.Println("Not implemented yet: dev environment shutdown")

			return nil
		},
	}

	return cmd
}
