package commands

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func newInitCommand() *cobra.Command {
	var (
		solo bool
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize OpenFroyo workspace",
		Long: `Initialize a new OpenFroyo workspace with configuration, keys, and data directories.

The --solo flag initializes a standalone workspace using SQLite and local file storage,
suitable for single-machine or development use.`,
		Example: `  # Initialize a standalone workspace
  froyo init --solo

  # Initialize with custom config path
  froyo init --solo --config /etc/openfroyo/config.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info().
				Bool("solo", solo).
				Str("config", configPath).
				Msg("Initializing workspace")

			// TODO: Implement workspace initialization
			// - Create directory structure (./data/, ./data/blobs/, ./data/keys/)
			// - Generate age keypair
			// - Initialize SQLite database
			// - Create default config file
			// - Setup embedded queue (Badger/Pebble)

			fmt.Println("Not implemented yet: workspace initialization")
			fmt.Printf("Would initialize workspace with solo=%v, config=%s\n", solo, configPath)

			return nil
		},
	}

	cmd.Flags().BoolVar(&solo, "solo", false, "initialize standalone workspace (SQLite + local storage)")
	cmd.MarkFlagRequired("solo")

	return cmd
}
