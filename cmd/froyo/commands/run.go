package commands

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func newRunCommand() *cobra.Command {
	var (
		params    []string
		extraVars map[string]string
		targets   []string
	)

	cmd := &cobra.Command{
		Use:   "run <action>",
		Short: "Run an action or runbook",
		Long: `Execute a predefined action or runbook.

Actions are standalone operational tasks defined in CUE configs, such as:
  - Service restarts
  - Health checks
  - Backup operations
  - Custom automation scripts

Runbooks are sequences of actions with conditional logic.`,
		Example: `  # Run a simple action
  froyo run restart-nginx

  # Run action on specific targets
  froyo run health-check --target web1 --target web2

  # Run action with parameters
  froyo run deploy --param version=1.2.3 --param env=production

  # Run with extra variables
  froyo run backup --extra-vars backup_dir=/mnt/backups`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			action := args[0]

			log.Info().
				Str("action", action).
				Strs("params", params).
				Interface("extra_vars", extraVars).
				Strs("targets", targets).
				Msg("Running action")

			// TODO: Implement action/runbook execution
			// - Load action definition from CUE configs
			// - Parse and validate parameters
			// - Resolve target hosts
			// - Execute action steps
			// - Handle conditional logic for runbooks
			// - Log results and update state

			fmt.Println("Not implemented yet: action execution")
			fmt.Printf("Would run action=%s, params=%v, extra_vars=%v, targets=%v\n",
				action, params, extraVars, targets)

			return nil
		},
	}

	cmd.Flags().StringSliceVarP(&params, "param", "p", nil, "action parameters (key=value)")
	cmd.Flags().StringToStringVarP(&extraVars, "extra-vars", "e", nil, "extra variables")
	cmd.Flags().StringSliceVarP(&targets, "target", "t", nil, "target hosts/groups")

	return cmd
}
