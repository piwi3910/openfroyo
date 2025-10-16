// Package engine provides the core orchestration engine for OpenFroyo.
package engine

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/openfroyo/openfroyo/pkg/micro_runner/client"
	"github.com/openfroyo/openfroyo/pkg/micro_runner/protocol"
	"github.com/openfroyo/openfroyo/pkg/stores"
	"github.com/openfroyo/openfroyo/pkg/transports/ssh"
	"github.com/rs/zerolog/log"
	sshpkg "golang.org/x/crypto/ssh"
)

// OnboardingService handles host onboarding operations.
type OnboardingService struct {
	store          stores.Store
	dataDir        string
	runnerBinaryPath string
}

// OnboardingConfig contains configuration for onboarding a host.
type OnboardingConfig struct {
	Host                string
	Port                int
	User                string
	Password            string
	KeyName             string
	CreateUser          string
	SudoRules           string
	DisablePasswordAuth bool
	Labels              map[string]string
}

// OnboardingResult contains the result of an onboarding operation.
type OnboardingResult struct {
	HostID              string            `json:"host_id"`
	Host                string            `json:"host"`
	User                string            `json:"user"`
	PublicKeyInstalled  bool              `json:"public_key_installed"`
	UserCreated         bool              `json:"user_created"`
	SudoersConfigured   bool              `json:"sudoers_configured"`
	SSHHardened         bool              `json:"ssh_hardened"`
	KeyPath             string            `json:"key_path"`
	Labels              map[string]string `json:"labels"`
	OnboardedAt         time.Time         `json:"onboarded_at"`
}

// NewOnboardingService creates a new onboarding service.
func NewOnboardingService(store stores.Store, dataDir string, runnerBinaryPath string) *OnboardingService {
	return &OnboardingService{
		store:          store,
		dataDir:        dataDir,
		runnerBinaryPath: runnerBinaryPath,
	}
}

// OnboardHost performs the complete onboarding workflow for a host.
func (s *OnboardingService) OnboardHost(ctx context.Context, config *OnboardingConfig) (*OnboardingResult, error) {
	log.Info().
		Str("host", config.Host).
		Str("user", config.User).
		Str("create_user", config.CreateUser).
		Msg("Starting host onboarding")

	result := &OnboardingResult{
		HostID:      uuid.New().String(),
		Host:        config.Host,
		User:        config.CreateUser,
		Labels:      config.Labels,
		OnboardedAt: time.Now(),
	}

	// Fallback to initial user if no management user specified
	if result.User == "" {
		result.User = config.User
	}

	// Step 1: Generate or load SSH keypair
	keyPath, pubKey, err := s.ensureSSHKey(config.KeyName)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure SSH key: %w", err)
	}
	result.KeyPath = keyPath

	log.Info().
		Str("key_path", keyPath).
		Msg("SSH keypair ready")

	// Step 2: Connect via SSH with password
	sshConfig := &ssh.Config{
		Host:                  config.Host,
		Port:                  config.Port,
		User:                  config.User,
		AuthMethod:            ssh.AuthMethodPassword,
		Password:              config.Password,
		StrictHostKeyChecking: false,
		ConnectionTimeout:     30 * time.Second,
		CommandTimeout:        5 * time.Minute,
	}

	transport, err := ssh.NewSSHClient(sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH client: %w", err)
	}

	if err := transport.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect via SSH: %w", err)
	}
	defer transport.Disconnect()

	log.Info().Msg("SSH connection established")

	// Step 3: Detect target architecture
	targetOS, targetArch, err := s.detectTargetArchitecture(ctx, transport)
	if err != nil {
		return nil, fmt.Errorf("failed to detect target architecture: %w", err)
	}

	log.Info().
		Str("os", targetOS).
		Str("arch", targetArch).
		Msg("Detected target architecture")

	// Step 3.5: Select appropriate micro-runner binary
	runnerBinaryPath, err := s.selectRunnerBinary(targetOS, targetArch)
	if err != nil {
		return nil, fmt.Errorf("failed to select runner binary: %w", err)
	}

	log.Info().
		Str("binary", runnerBinaryPath).
		Msg("Selected micro-runner binary")

	// Step 4: Create micro-runner transport adapter
	// Get the underlying SSH client for direct session creation
	sshClient, err := transport.GetClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get SSH client: %w", err)
	}

	runnerTransport := &sshTransportAdapter{
		transport: transport,
		client:    sshClient,
	}

	// Step 5: Start micro-runner
	runnerClient, err := client.NewClient(&client.Config{
		Transport:      runnerTransport,
		RunnerPath:     runnerBinaryPath,
		RemotePath:     "/tmp/froyo-micro-runner",
		StartupTimeout: 15 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create runner client: %w", err)
	}

	if err := runnerClient.Start(ctx, &client.Config{
		Transport:      runnerTransport,
		RunnerPath:     runnerBinaryPath,
		RemotePath:     "/tmp/froyo-micro-runner",
		StartupTimeout: 15 * time.Second,
	}); err != nil {
		return nil, fmt.Errorf("failed to start micro-runner: %w", err)
	}
	defer runnerClient.Close(ctx, "/tmp/froyo-micro-runner")

	log.Info().Msg("Micro-runner started")

	// Step 5.5: Detect if we need sudo (check if we're root)
	needsSudo := config.User != "root"
	sudoPass := ""
	if needsSudo {
		sudoPass = config.Password
	}

	log.Info().
		Bool("needs_sudo", needsSudo).
		Str("connected_as", config.User).
		Msg("Determined privilege escalation needs")

	// Step 6: Create management user if specified
	if config.CreateUser != "" && config.CreateUser != config.User {
		if err := s.createUser(ctx, runnerClient, config.CreateUser, pubKey, needsSudo, sudoPass); err != nil {
			return nil, fmt.Errorf("failed to create management user: %w", err)
		}
		result.UserCreated = true
		log.Info().
			Str("user", config.CreateUser).
			Msg("Management user created")
	} else {
		// Install public key for existing user
		if err := s.installSSHKey(ctx, runnerClient, config.User, pubKey, needsSudo, sudoPass); err != nil {
			return nil, fmt.Errorf("failed to install SSH key: %w", err)
		}
		log.Info().
			Str("user", config.User).
			Msg("SSH key installed for user")
	}
	result.PublicKeyInstalled = true

	// Step 7: Configure sudoers if specified
	if config.CreateUser != "" && config.SudoRules != "" {
		if err := s.configureSudoers(ctx, runnerClient, config.CreateUser, config.SudoRules, needsSudo, sudoPass); err != nil {
			return nil, fmt.Errorf("failed to configure sudoers: %w", err)
		}
		result.SudoersConfigured = true
		log.Info().
			Str("user", config.CreateUser).
			Msg("Sudoers configured")
	}

	// Step 8: Harden SSH if requested
	if config.DisablePasswordAuth {
		if err := s.hardenSSH(ctx, runnerClient, result.User); err != nil {
			return nil, fmt.Errorf("failed to harden SSH: %w", err)
		}
		result.SSHHardened = true
		log.Info().Msg("SSH hardened")
	}

	// Step 9: Register host in store
	hostData := &Host{
		ID:          result.HostID,
		Address:     config.Host,
		Port:        config.Port,
		User:        result.User,
		KeyPath:     keyPath,
		Labels:      config.Labels,
		OnboardedAt: result.OnboardedAt,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.registerHost(ctx, hostData); err != nil {
		return nil, fmt.Errorf("failed to register host: %w", err)
	}

	log.Info().
		Str("host_id", result.HostID).
		Str("host", config.Host).
		Msg("Host onboarding completed successfully")

	return result, nil
}

// ensureSSHKey generates or loads an SSH keypair.
func (s *OnboardingService) ensureSSHKey(keyName string) (string, string, error) {
	keysDir := filepath.Join(s.dataDir, "keys")
	if err := os.MkdirAll(keysDir, 0700); err != nil {
		return "", "", fmt.Errorf("failed to create keys directory: %w", err)
	}

	privateKeyPath := filepath.Join(keysDir, keyName)
	publicKeyPath := privateKeyPath + ".pub"

	// Check if key already exists
	if _, err := os.Stat(privateKeyPath); err == nil {
		// Key exists, load it
		pubKeyBytes, err := os.ReadFile(publicKeyPath)
		if err != nil {
			return "", "", fmt.Errorf("failed to read public key: %w", err)
		}
		return privateKeyPath, string(pubKeyBytes), nil
	}

	// Generate new ED25519 keypair
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate keypair: %w", err)
	}

	// Marshal private key to OpenSSH format using PEM encoding
	// ED25519 keys need special handling
	privKeyBytes, err := sshpkg.MarshalPrivateKey(privKey, "")
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal private key: %w", err)
	}

	// Write private key
	privPEM := pem.EncodeToMemory(privKeyBytes)
	if err := os.WriteFile(privateKeyPath, privPEM, 0600); err != nil {
		return "", "", fmt.Errorf("failed to write private key: %w", err)
	}

	// Create SSH public key from ed25519 public key
	sshPubKey, err := sshpkg.NewPublicKey(pubKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to create SSH public key: %w", err)
	}

	// Marshal public key in authorized_keys format
	pubKeyStr := string(sshpkg.MarshalAuthorizedKey(sshPubKey))

	// Write public key
	if err := os.WriteFile(publicKeyPath, []byte(pubKeyStr), 0644); err != nil {
		return "", "", fmt.Errorf("failed to write public key: %w", err)
	}

	log.Info().
		Str("private_key", privateKeyPath).
		Str("public_key", publicKeyPath).
		Msg("Generated new SSH keypair")

	return privateKeyPath, pubKeyStr, nil
}

// createUser creates a management user with SSH key access.
func (s *OnboardingService) createUser(ctx context.Context, runner *client.Client, username string, pubKey string, needsSudo bool, sudoPassword string) error {
	// Create user with home directory
	createUserCmd := &protocol.CommandMessage{
		ID:      uuid.New().String(),
		Type:    protocol.CommandTypeExec,
		Timeout: 60,
	}

	paramsBytes, err := json.Marshal(&protocol.ExecParams{
		Command:      "useradd",
		Args:         []string{"-m", "-s", "/bin/bash", username},
		CaptureOut:   true,
		CaptureErr:   true,
		UseSudo:      needsSudo,
		SudoPassword: sudoPassword,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal exec params: %w", err)
	}
	createUserCmd.Params = paramsBytes

	if _, err := runner.Execute(ctx, createUserCmd); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Install SSH key for the new user
	return s.installSSHKey(ctx, runner, username, pubKey, needsSudo, sudoPassword)
}

// installSSHKey installs an SSH public key for a user.
func (s *OnboardingService) installSSHKey(ctx context.Context, runner *client.Client, username string, pubKey string, needsSudo bool, sudoPassword string) error {
	// Create .ssh directory
	mkdirCmd := &protocol.CommandMessage{
		ID:      uuid.New().String(),
		Type:    protocol.CommandTypeExec,
		Timeout: 30,
	}

	mkdirParams, err := json.Marshal(&protocol.ExecParams{
		Command:      "mkdir",
		Args:         []string{"-p", fmt.Sprintf("/home/%s/.ssh", username)},
		CaptureOut:   true,
		CaptureErr:   true,
		UseSudo:      needsSudo,
		SudoPassword: sudoPassword,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal mkdir params: %w", err)
	}
	mkdirCmd.Params = mkdirParams

	if _, err := runner.Execute(ctx, mkdirCmd); err != nil {
		return fmt.Errorf("failed to create .ssh directory: %w", err)
	}

	// Write authorized_keys file
	writeKeyCmd := &protocol.CommandMessage{
		ID:      uuid.New().String(),
		Type:    protocol.CommandTypeFileWrite,
		Timeout: 30,
	}

	writeParams, err := json.Marshal(&protocol.FileWriteParams{
		Path:         fmt.Sprintf("/home/%s/.ssh/authorized_keys", username),
		Content:      pubKey,
		Mode:         "0600",
		Owner:        username,
		Group:        username,
		Create:       true,
		UseSudo:      needsSudo,
		SudoPassword: sudoPassword,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal file write params: %w", err)
	}
	writeKeyCmd.Params = writeParams

	if _, err := runner.Execute(ctx, writeKeyCmd); err != nil {
		return fmt.Errorf("failed to write authorized_keys: %w", err)
	}

	// Fix ownership of .ssh directory
	chownCmd := &protocol.CommandMessage{
		ID:      uuid.New().String(),
		Type:    protocol.CommandTypeExec,
		Timeout: 30,
	}

	chownParams, err := json.Marshal(&protocol.ExecParams{
		Command:      "chown",
		Args:         []string{"-R", fmt.Sprintf("%s:%s", username, username), fmt.Sprintf("/home/%s/.ssh", username)},
		CaptureOut:   true,
		CaptureErr:   true,
		UseSudo:      needsSudo,
		SudoPassword: sudoPassword,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal chown params: %w", err)
	}
	chownCmd.Params = chownParams

	if _, err := runner.Execute(ctx, chownCmd); err != nil {
		return fmt.Errorf("failed to fix ownership: %w", err)
	}

	return nil
}

// configureSudoers configures sudo access for a user.
func (s *OnboardingService) configureSudoers(ctx context.Context, runner *client.Client, username string, rules string, needsSudo bool, sudoPassword string) error {
	cmd := &protocol.CommandMessage{
		ID:      uuid.New().String(),
		Type:    protocol.CommandTypeSudoersEnsure,
		Timeout: 30,
	}

	// Parse sudo rules (format: "NOPASSWD: /usr/bin/systemctl,/usr/bin/apt")
	commands := []string{"ALL"} // Default to ALL if specific commands not parsed
	noPasswd := true

	params, err := json.Marshal(&protocol.SudoersEnsureParams{
		User:         username,
		Commands:     commands,
		NoPasswd:     noPasswd,
		State:        "present",
		UseSudo:      needsSudo,
		SudoPassword: sudoPassword,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal sudoers params: %w", err)
	}
	cmd.Params = params

	if _, err := runner.Execute(ctx, cmd); err != nil {
		return fmt.Errorf("failed to configure sudoers: %w", err)
	}

	return nil
}

// hardenSSH applies SSH hardening configuration.
func (s *OnboardingService) hardenSSH(ctx context.Context, runner *client.Client, allowUser string) error {
	cmd := &protocol.CommandMessage{
		ID:      uuid.New().String(),
		Type:    protocol.CommandTypeSSHDHarden,
		Timeout: 60,
	}

	params, err := json.Marshal(&protocol.SSHDHardenParams{
		DisablePasswordAuth: true,
		DisableRootLogin:    false,
		AllowUsers:          []string{allowUser},
		TestConnection:      true,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal sshd harden params: %w", err)
	}
	cmd.Params = params

	if _, err := runner.Execute(ctx, cmd); err != nil {
		return fmt.Errorf("failed to harden SSH: %w", err)
	}

	return nil
}

// registerHost stores host information in the database.
func (s *OnboardingService) registerHost(ctx context.Context, host *Host) error {
	// Store host as a fact with special namespace
	labelsJSON, err := json.Marshal(host.Labels)
	if err != nil {
		return fmt.Errorf("failed to marshal labels: %w", err)
	}

	hostDataJSON, err := json.Marshal(host)
	if err != nil {
		return fmt.Errorf("failed to marshal host data: %w", err)
	}

	// Store host metadata
	hostFact := &stores.Fact{
		ID:        uuid.New().String(),
		TargetID:  host.ID,
		Namespace: "host.metadata",
		Key:       "info",
		Value:     string(hostDataJSON),
		TTL:       0, // No expiry for host metadata
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.store.UpsertFact(ctx, hostFact); err != nil {
		return fmt.Errorf("failed to store host metadata: %w", err)
	}

	// Store labels separately for easier querying
	labelsFact := &stores.Fact{
		ID:        uuid.New().String(),
		TargetID:  host.ID,
		Namespace: "host.labels",
		Key:       "all",
		Value:     string(labelsJSON),
		TTL:       0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.store.UpsertFact(ctx, labelsFact); err != nil {
		return fmt.Errorf("failed to store host labels: %w", err)
	}

	// Create audit entry
	auditEntry := &stores.AuditEntry{
		Action:    "host.onboarded",
		Actor:     "system",
		TargetID:  &host.ID,
		Timestamp: time.Now(),
	}

	if err := s.store.CreateAuditEntry(ctx, auditEntry); err != nil {
		log.Warn().Err(err).Msg("Failed to create audit entry")
		// Non-fatal, continue
	}

	return nil
}

// sshTransportAdapter adapts our SSH transport to the micro-runner client.Transport interface.
type sshTransportAdapter struct {
	transport *ssh.SSHClient
	client    *sshpkg.Client // Store the underlying client for direct session creation
}

func (a *sshTransportAdapter) Upload(ctx context.Context, localPath, remotePath string) error {
	return a.transport.UploadFile(ctx, localPath, remotePath, 0755)
}

func (a *sshTransportAdapter) Execute(ctx context.Context, remotePath string) (stdin io.WriteCloser, stdout io.ReadCloser, err error) {
	// First, ensure the binary is executable
	chmodSession, err := a.client.NewSession()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create chmod session: %w", err)
	}

	if err := chmodSession.Run(fmt.Sprintf("chmod +x %s", remotePath)); err != nil {
		chmodSession.Close()
		log.Warn().Err(err).Msg("Failed to chmod runner binary")
		// Continue anyway, might already be executable
	}
	chmodSession.Close()

	// We need to execute the micro-runner directly, not through an interactive shell
	// Create a new SSH session directly
	session, err := a.client.NewSession()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Get stdin pipe
	stdinPipe, err := session.StdinPipe()
	if err != nil {
		session.Close()
		return nil, nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	// Get stdout pipe
	stdoutPipe, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		return nil, nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	// Get stderr pipe for debugging
	stderrPipe, err := session.StderrPipe()
	if err != nil {
		session.Close()
		return nil, nil, fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	// Log stderr output in background
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stderrPipe.Read(buf)
			if n > 0 {
				log.Warn().Str("stderr", string(buf[:n])).Msg("Micro-runner stderr")
			}
			if err != nil {
				break
			}
		}
	}()

	// Start the micro-runner directly (not through a shell)
	if err := session.Start(remotePath); err != nil {
		session.Close()
		return nil, nil, fmt.Errorf("failed to start runner: %w", err)
	}

	// Wrap stdout with a ReadCloser that also closes the session
	wrappedStdout := &readerCloser{
		Reader: stdoutPipe,
		cleanupFunc: func() error {
			return session.Close()
		},
	}

	// Return stdin and stdout for JSON protocol communication
	return stdinPipe, wrappedStdout, nil
}

// readerCloser wraps an io.Reader with a Close method
type readerCloser struct {
	io.Reader
	cleanupFunc func() error
}

func (rc *readerCloser) Close() error {
	if rc.cleanupFunc != nil {
		return rc.cleanupFunc()
	}
	return nil
}

func (a *sshTransportAdapter) Cleanup(ctx context.Context, remotePath string) error {
	// Execute rm command to delete the runner
	_, _, err := a.transport.ExecuteCommand(ctx, fmt.Sprintf("rm -f %s", remotePath))
	return err
}

// detectTargetArchitecture detects the OS and architecture of the remote host.
func (s *OnboardingService) detectTargetArchitecture(ctx context.Context, transport *ssh.SSHClient) (string, string, error) {
	// Run 'uname -s' to get OS
	osOutput, _, err := transport.ExecuteCommand(ctx, "uname -s")
	if err != nil {
		return "", "", fmt.Errorf("failed to detect OS: %w", err)
	}

	// Run 'uname -m' to get architecture
	archOutput, _, err := transport.ExecuteCommand(ctx, "uname -m")
	if err != nil {
		return "", "", fmt.Errorf("failed to detect architecture: %w", err)
	}

	// Normalize OS name
	targetOS := strings.ToLower(strings.TrimSpace(osOutput))

	// Normalize architecture
	rawArch := strings.TrimSpace(archOutput)
	targetArch := normalizeArchitecture(rawArch)

	return targetOS, targetArch, nil
}

// normalizeArchitecture converts various architecture names to Go's GOARCH format.
func normalizeArchitecture(arch string) string {
	switch arch {
	case "x86_64", "amd64":
		return "amd64"
	case "aarch64", "arm64":
		return "arm64"
	case "i386", "i686", "x86":
		return "386"
	case "armv7l", "armv7":
		return "arm"
	default:
		return arch
	}
}

// selectRunnerBinary selects the appropriate micro-runner binary based on target OS and architecture.
func (s *OnboardingService) selectRunnerBinary(targetOS, targetArch string) (string, error) {
	// Build the expected binary name
	binaryName := fmt.Sprintf("micro-runner-%s-%s", targetOS, targetArch)
	binaryPath := filepath.Join(filepath.Dir(s.runnerBinaryPath), binaryName)

	// Check if the specific binary exists
	if _, err := os.Stat(binaryPath); err == nil {
		return binaryPath, nil
	}

	// If not found, provide helpful error message
	return "", fmt.Errorf("micro-runner binary not found for %s/%s at %s\nBuild it with: make build-runner-%s-%s",
		targetOS, targetArch, binaryPath, targetOS, targetArch)
}
