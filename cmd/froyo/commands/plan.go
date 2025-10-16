package commands

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func newPlanCommand() *cobra.Command {
	var (
		outFile    string
		dotFile    string
		targets    []string
		refresh    bool
		noRefresh  bool
	)

	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Generate execution plan",
		Long: `Generate an execution plan by comparing desired state (CUE configs) with actual state.

The plan:
  - Discovers current state via facts collection
  - Computes diffs between desired and actual
  - Builds a DAG of plan units (PUs) with dependencies
  - Persists the plan for execution with 'apply'`,
		Example: `  # Generate plan and save to file
  froyo plan --out plan.json

  # Generate plan with execution graph visualization
  froyo plan --out plan.json --dot plan.dot

  # Plan for specific targets only
  froyo plan --out plan.json --target host1 --target host2

  # Plan without refreshing facts (use cached)
  froyo plan --out plan.json --no-refresh`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info().
				Str("out", outFile).
				Str("dot", dotFile).
				Strs("targets", targets).
				Bool("refresh", refresh && !noRefresh).
				Msg("Generating plan")

			// TODO: Implement planning
			// - Load and evaluate CUE configs
			// - Collect facts (unless --no-refresh)
			// - Compute diffs (desired vs actual)
			// - Build DAG respecting dependencies
			// - Validate DAG (no cycles)
			// - Persist plan to outFile (JSON)
			// - Generate DOT graph if requested

			fmt.Println("Not implemented yet: plan generation")
			fmt.Printf("Would create plan: out=%s, dot=%s, targets=%v, refresh=%v\n",
				outFile, dotFile, targets, refresh && !noRefresh)

			return nil
		},
	}

	cmd.Flags().StringVarP(&outFile, "out", "o", "plan.json", "output plan file path")
	cmd.Flags().StringVar(&dotFile, "dot", "", "output DOT graph file (optional)")
	cmd.Flags().StringSliceVarP(&targets, "target", "t", nil, "limit plan to specific targets")
	cmd.Flags().BoolVar(&refresh, "refresh", true, "refresh facts before planning")
	cmd.Flags().BoolVar(&noRefresh, "no-refresh", false, "skip facts refresh (use cached)")
	cmd.MarkFlagRequired("out")

	return cmd
}
