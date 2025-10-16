package commands

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func newBackupCommand() *cobra.Command {
	var (
		outFile     string
		compress    bool
		includeBlobs bool
	)

	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup OpenFroyo data",
		Long: `Create a backup of OpenFroyo workspace data.

The backup includes:
  - SQLite database (hot-copy)
  - Configuration files
  - Keys and secrets
  - Optionally: blob storage

The backup can be used to restore state on the same or different machine.`,
		Example: `  # Create compressed backup
  froyo backup --out backup.tar.gz --compress

  # Backup including blob storage
  froyo backup --out backup.tar.gz --include-blobs

  # Simple backup
  froyo backup --out backup.tar`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info().
				Str("out", outFile).
				Bool("compress", compress).
				Bool("include_blobs", includeBlobs).
				Msg("Creating backup")

			// TODO: Implement backup
			// - Perform SQLite hot-copy (VACUUM INTO)
			// - Archive configuration files
			// - Archive keys directory (encrypted)
			// - Optionally include blobs directory
			// - Create tar/tar.gz archive
			// - Verify backup integrity
			// - Log backup metadata

			fmt.Println("Not implemented yet: backup creation")
			fmt.Printf("Would create backup: out=%s, compress=%v, include_blobs=%v\n",
				outFile, compress, includeBlobs)

			return nil
		},
	}

	cmd.Flags().StringVarP(&outFile, "out", "o", "froyo-backup.tar.gz", "backup output file")
	cmd.Flags().BoolVar(&compress, "compress", true, "compress backup with gzip")
	cmd.Flags().BoolVar(&includeBlobs, "include-blobs", true, "include blob storage in backup")

	return cmd
}
