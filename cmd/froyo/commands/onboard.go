package commands

import (
	"fmt"

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

			// TODO: Implement SSH onboarding
			// - Establish SSH connection with password auth
			// - Upload micro-runner binary (verify signature)
			// - Execute micro-runner
			// - Send commands via stdio protocol:
			//   - Create user if specified
			//   - Install SSH authorized_keys
			//   - Configure sudoers
			//   - Optionally disable password auth
			//   - Restart sshd safely
			// - Register host in inventory with labels
			// - Store host fingerprint and key handle
			// - Verify micro-runner self-deletion

			fmt.Println("Not implemented yet: SSH onboarding")
			fmt.Printf("Would onboard host=%s, user=%s, create_user=%s, lock_down=%v, labels=%v\n",
				host, user, createUser, lockDown, labels)

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
