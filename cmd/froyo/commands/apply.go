package commands

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func newApplyCommand() *cobra.Command {
	var (
		planFile    string
		autoApprove bool
		parallelism int
	)

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Execute a plan",
		Long: `Execute a previously generated plan.

This command:
  - Loads the plan file
  - Optionally prompts for approval (unless --auto-approve)
  - Executes the DAG in parallel (respecting dependencies)
  - Runs provider operations in WASM sandbox
  - Delegates complex operations to micro-runner
  - Updates state and logs events`,
		Example: `  # Apply plan with approval prompt
  froyo apply --plan plan.json

  # Auto-approve and apply
  froyo apply --plan plan.json --auto-approve

  # Apply with limited parallelism
  froyo apply --plan plan.json --parallelism 5`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info().
				Str("plan", planFile).
				Bool("auto_approve", autoApprove).
				Int("parallelism", parallelism).
				Msg("Applying plan")

			// TODO: Implement apply
			// - Load plan from planFile
			// - Validate plan is current (no config changes since plan)
			// - Show summary and prompt for approval (unless auto-approve)
			// - Execute DAG with parallelism limit
			// - For each PU:
			//   - Load provider WASM
			//   - Execute Apply() in sandbox
			//   - Or delegate to micro-runner for local ops
			//   - Update state
			//   - Log events
			// - Trigger post-apply handlers
			// - Run smoke tests if configured

			fmt.Println("Not implemented yet: plan execution")
			fmt.Printf("Would apply plan=%s, auto_approve=%v, parallelism=%d\n",
				planFile, autoApprove, parallelism)

			return nil
		},
	}

	cmd.Flags().StringVarP(&planFile, "plan", "p", "plan.json", "plan file to execute")
	cmd.Flags().BoolVar(&autoApprove, "auto-approve", false, "skip approval prompt")
	cmd.Flags().IntVar(&parallelism, "parallelism", 10, "max parallel operations")
	cmd.MarkFlagRequired("plan")

	return cmd
}
