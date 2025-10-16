package commands

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/openfroyo/openfroyo/pkg/engine"
	"github.com/openfroyo/openfroyo/pkg/stores"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func newOnboardCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "onboard",
		Short: "Onboard new hosts",
		Long: `Onboard new hosts to OpenFroyo management.

Onboarding prepares a host for management by:
  - Creating a management user
  - Installing SSH keys
  - Configuring sudo access
  - Optionally hardening SSH
  - Registering host in inventory`,
	}

	cmd.AddCommand(newOnboardSSHCommand())
	cmd.AddCommand(newOnboardRollbackCommand())

	return cmd
}

func newOnboardSSHCommand() *cobra.Command {
	var (
		host             string
		user             string
		password         string
		keyName          string
		createUser       string
		sudoRules        string
		lockDown         bool
		labels           []string
		port             int
	)

	cmd := &cobra.Command{
		Use:   "ssh",
		Short: "Onboard host via SSH",
		Long: `Onboard a new host via SSH connection.

This command:
  1. Connects via SSH using password authentication
  2. Uploads and runs the micro-runner
  3. Creates management user with SSH key access
  4. Configures sudo permissions
  5. Optionally hardens SSH configuration
  6. Registers host in inventory
  7. Micro-runner self-deletes`,
		Example: `  # Basic onboarding
  froyo onboard ssh --host 10.0.0.42 --user root --password secret

  # Full onboarding with user creation and hardening
  froyo onboard ssh \
    --host 10.0.0.42 \
    --user root \
    --password s3cr3t \
    --key default-ed25519 \
    --create-user froyo \
    --sudo 'NOPASSWD: /usr/bin/systemctl,/usr/bin/apt,/usr/bin/dnf' \
    --lock-down \
    --labels env=dev,role=web

  # Onboard on custom SSH port
  froyo onboard ssh --host server.example.com --port 2222 --user admin`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info().
				Str("host", host).
				Str("user", user).
				Str("key", keyName).
				Str("create_user", createUser).
				Bool("lock_down", lockDown).
				Strs("labels", labels).
				Int("port", port).
				Msg("Onboarding host via SSH")

			ctx := context.Background()

			// Load configuration
			dataDir := "./data"
			if configPath != "" {
				dataDir = filepath.Join(filepath.Dir(configPath), "data")
			}

			// Initialize store
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
			defer store.Close()

			// Parse labels
			labelMap := make(map[string]string)
			for _, label := range labels {
				parts := strings.SplitN(label, "=", 2)
				if len(parts) == 2 {
					labelMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
				}
			}

			// Base runner binary path - the actual architecture-specific binary
			// will be selected automatically based on target host detection
			runnerBinaryPath := filepath.Join("./bin", "micro-runner")

			// Create onboarding service
			onboardingSvc := engine.NewOnboardingService(store, dataDir, runnerBinaryPath)

			// Perform onboarding
			fmt.Printf("🚀 Starting onboarding for %s...\n\n", host)

			result, err := onboardingSvc.OnboardHost(ctx, &engine.OnboardingConfig{
				Host:                host,
				Port:                port,
				User:                user,
				Password:            password,
				KeyName:             keyName,
				CreateUser:          createUser,
				SudoRules:           sudoRules,
				DisablePasswordAuth: lockDown,
				Labels:              labelMap,
			})

			if err != nil {
				return fmt.Errorf("onboarding failed: %w", err)
			}

			// Display results
			fmt.Printf("\n✅ Host onboarded successfully!\n\n")
			fmt.Printf("Host ID:              %s\n", result.HostID)
			fmt.Printf("Address:              %s:%d\n", result.Host, port)
			fmt.Printf("User:                 %s\n", result.User)
			fmt.Printf("Key Path:             %s\n", result.KeyPath)
			fmt.Printf("Public Key Installed: %v\n", result.PublicKeyInstalled)
			if result.UserCreated {
				fmt.Printf("User Created:         %v\n", result.UserCreated)
			}
			if result.SudoersConfigured {
				fmt.Printf("Sudoers Configured:   %v\n", result.SudoersConfigured)
			}
			if result.SSHHardened {
				fmt.Printf("SSH Hardened:         %v\n", result.SSHHardened)
			}
			if len(result.Labels) > 0 {
				fmt.Printf("Labels:\n")
				for k, v := range result.Labels {
					fmt.Printf("  %s: %s\n", k, v)
				}
			}

			fmt.Printf("\n💡 Next steps:\n")
			fmt.Printf("  1. Collect facts from this host:\n")
			fmt.Printf("     froyo facts collect --target %s\n\n", result.HostID)
			fmt.Printf("  2. Test SSH connection with key:\n")
			fmt.Printf("     ssh -i %s %s@%s\n\n", result.KeyPath, result.User, result.Host)

			return nil
		},
	}

	cmd.Flags().StringVar(&host, "host", "", "target host address")
	cmd.Flags().StringVar(&user, "user", "root", "SSH user for initial connection")
	cmd.Flags().StringVar(&password, "password", "", "SSH password (use with caution)")
	cmd.Flags().StringVar(&keyName, "key", "default-ed25519", "SSH key to install")
	cmd.Flags().StringVar(&createUser, "create-user", "", "create management user")
	cmd.Flags().StringVar(&sudoRules, "sudo", "", "sudo rules for management user")
	cmd.Flags().BoolVar(&lockDown, "lock-down", false, "disable password auth after setup")
	cmd.Flags().StringSliceVar(&labels, "labels", nil, "host labels (key=value)")
	cmd.Flags().IntVar(&port, "port", 22, "SSH port")

	cmd.MarkFlagRequired("host")
	cmd.MarkFlagRequired("password")

	return cmd
}

func newOnboardRollbackCommand() *cobra.Command {
	var (
		host string
	)

	cmd := &cobra.Command{
		Use:   "rollback",
		Short: "Rollback onboarding changes",
		Long: `Rollback onboarding changes on a host.

This restores the host to its pre-onboarding state by:
  - Removing the management user
  - Removing SSH keys
  - Restoring original SSH configuration
  - Removing host from inventory`,
		Example: `  # Rollback onboarding for a host
  froyo onboard rollback --host 10.0.0.42`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info().
				Str("host", host).
				Msg("Rolling back onboarding")

			// TODO: Implement rollback
			// - Connect to host with management credentials
			// - Remove management user
			// - Restore SSH config from backup
			// - Remove from inventory
			// - Log rollback event

			fmt.Println("Not implemented yet: onboarding rollback")
			fmt.Printf("Would rollback onboarding for host=%s\n", host)

			return nil
		},
	}

	cmd.Flags().StringVar(&host, "host", "", "target host address")
	cmd.MarkFlagRequired("host")

	return cmd
}
