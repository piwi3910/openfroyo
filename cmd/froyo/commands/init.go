package commands

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"github.com/openfroyo/openfroyo/pkg/stores"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	sshpkg "golang.org/x/crypto/ssh"
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

			ctx := context.Background()

			// Determine data directory
			dataDir := "./data"
			if configPath != "" {
				// If custom config path, use its directory
				dataDir = filepath.Join(filepath.Dir(configPath), "data")
			}

			fmt.Printf("Initializing OpenFroyo workspace in %s\n\n", dataDir)

			// Step 1: Create directory structure
			dirs := []string{
				dataDir,
				filepath.Join(dataDir, "blobs"),
				filepath.Join(dataDir, "keys"),
			}

			for _, dir := range dirs {
				if err := os.MkdirAll(dir, 0700); err != nil {
					return fmt.Errorf("failed to create directory %s: %w", dir, err)
				}
				fmt.Printf("✓ Created directory: %s\n", dir)
			}

			// Step 2: Initialize SQLite database
			dbPath := filepath.Join(dataDir, "openfroyo.db")
			store, err := stores.NewSQLiteStore(stores.Config{
				Path: dbPath,
			})
			if err != nil {
				return fmt.Errorf("failed to create store: %w", err)
			}

			if err := store.Init(ctx); err != nil {
				return fmt.Errorf("failed to initialize store: %w", err)
			}

			if err := store.Migrate(ctx); err != nil {
				return fmt.Errorf("failed to run migrations: %w", err)
			}

			fmt.Printf("✓ Initialized SQLite database: %s\n", dbPath)

			// Step 3: Create default config file
			defaultConfig := `# OpenFroyo Configuration

# Data directory
data_dir: %s

# Database settings
database:
  path: %s

# Telemetry settings
telemetry:
  enabled: true
  log_level: info

# Micro-runner settings
micro_runner:
  binary_path: ./bin/micro-runner
  timeout: 600
`
			configContent := fmt.Sprintf(defaultConfig, dataDir, dbPath)

			if configPath == "" {
				configPath = "./froyo.yaml"
			}

			if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
				return fmt.Errorf("failed to write config file: %w", err)
			}

			fmt.Printf("✓ Created config file: %s\n", configPath)

			// Step 4: Generate default SSH key
			keyPath := filepath.Join(dataDir, "keys", "default-ed25519")
			if _, err := os.Stat(keyPath); os.IsNotExist(err) {
				pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
				if err != nil {
					return fmt.Errorf("failed to generate keypair: %w", err)
				}

				// Marshal private key
				privKeyBytes, err := sshpkg.MarshalPrivateKey(privKey, "")
				if err != nil {
					return fmt.Errorf("failed to marshal private key: %w", err)
				}

				privPEM := pem.EncodeToMemory(privKeyBytes)
				if err := os.WriteFile(keyPath, privPEM, 0600); err != nil {
					return fmt.Errorf("failed to write private key: %w", err)
				}

				// Marshal public key
				sshPubKey, err := sshpkg.NewPublicKey(pubKey)
				if err != nil {
					return fmt.Errorf("failed to create SSH public key: %w", err)
				}

				pubKeyStr := sshpkg.MarshalAuthorizedKey(sshPubKey)
				if err := os.WriteFile(keyPath+".pub", pubKeyStr, 0644); err != nil {
					return fmt.Errorf("failed to write public key: %w", err)
				}

				fmt.Printf("✓ Generated SSH keypair: %s\n", keyPath)
			} else {
				fmt.Printf("✓ SSH keypair already exists: %s\n", keyPath)
			}

			// Done
			fmt.Printf("\n✅ Workspace initialized successfully!\n\n")
			fmt.Printf("Next steps:\n")
			fmt.Printf("  1. Onboard a host:\n")
			fmt.Printf("     froyo onboard ssh --host <ip> --user root --password <pass>\n\n")
			fmt.Printf("  2. Collect facts:\n")
			fmt.Printf("     froyo facts collect\n\n")

			return nil
		},
	}

	cmd.Flags().BoolVar(&solo, "solo", false, "initialize standalone workspace (SQLite + local storage)")
	cmd.MarkFlagRequired("solo")

	return cmd
}
