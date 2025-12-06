package ssh

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"

	"github.com/piwi3910/openfroyo/internal/parser"
)

// HostKeyMode defines the SSH host key checking behavior
type HostKeyMode string

const (
	// HostKeyModeStrict fails if the host key is not in known_hosts
	HostKeyModeStrict HostKeyMode = "strict"
	// HostKeyModeWarn warns but allows connection if host key is not in known_hosts
	HostKeyModeWarn HostKeyMode = "warn"
	// HostKeyModeIgnore ignores host key validation (insecure, for development only)
	HostKeyModeIgnore HostKeyMode = "ignore"
	// HostKeyModeAutoAdd automatically adds unknown keys to known_hosts
	HostKeyModeAutoAdd HostKeyMode = "auto_add"
)

// Client represents an SSH client to a remote host
type Client struct {
	host      parser.Host
	client    *ssh.Client
	runnerDir string
}

// TaskInput represents the input to a WASM module
type TaskInput struct {
	Vars    map[string]interface{} `json:"vars"`
	Context TaskContext            `json:"context"`
}

// TaskContext provides execution context
type TaskContext struct {
	Host     string `json:"host"`
	TaskName string `json:"task_name"`
}

// TaskOutput represents the output from a WASM module
type TaskOutput struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message"`
	Facts   map[string]interface{} `json:"facts"`
}

// getKnownHostsPath returns the path to the known_hosts file
func getKnownHostsPath(customPath string) string {
	if customPath != "" {
		return customPath
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".ssh", "known_hosts")
}

// createHostKeyCallback creates the appropriate host key callback based on mode
func createHostKeyCallback(host parser.Host) (ssh.HostKeyCallback, error) {
	mode := HostKeyMode(host.SSHHostKeyMode)
	if mode == "" {
		mode = HostKeyModeWarn // Default to warn mode
	}

	knownHostsPath := getKnownHostsPath(host.SSHKnownHostsFile)

	switch mode {
	case HostKeyModeIgnore:
		return ssh.InsecureIgnoreHostKey(), nil

	case HostKeyModeStrict:
		if knownHostsPath == "" {
			return nil, fmt.Errorf("known_hosts file path required for strict mode")
		}
		callback, err := knownhosts.New(knownHostsPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read known_hosts: %w", err)
		}
		return callback, nil

	case HostKeyModeAutoAdd:
		return createAutoAddCallback(knownHostsPath), nil

	case HostKeyModeWarn:
		fallthrough
	default:
		return createWarnCallback(knownHostsPath), nil
	}
}

// createWarnCallback creates a callback that warns about unknown keys but allows connection
func createWarnCallback(knownHostsPath string) ssh.HostKeyCallback {
	var knownHostsCallback ssh.HostKeyCallback

	if knownHostsPath != "" {
		if _, err := os.Stat(knownHostsPath); err == nil {
			callback, err := knownhosts.New(knownHostsPath)
			if err == nil {
				knownHostsCallback = callback
			}
		}
	}

	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		if knownHostsCallback != nil {
			err := knownHostsCallback(hostname, remote, key)
			if err == nil {
				return nil // Key is known
			}
			// Key is unknown or changed - warn but allow
			fmt.Printf("WARNING: Host key for %s is not in known_hosts or has changed\n", hostname)
			fmt.Printf("  Fingerprint: %s\n", ssh.FingerprintSHA256(key))
		} else {
			fmt.Printf("WARNING: No known_hosts file found, accepting host key for %s\n", hostname)
			fmt.Printf("  Fingerprint: %s\n", ssh.FingerprintSHA256(key))
		}
		return nil
	}
}

// createAutoAddCallback creates a callback that automatically adds unknown keys
func createAutoAddCallback(knownHostsPath string) ssh.HostKeyCallback {
	var knownHostsCallback ssh.HostKeyCallback

	if knownHostsPath != "" {
		if _, err := os.Stat(knownHostsPath); err == nil {
			callback, err := knownhosts.New(knownHostsPath)
			if err == nil {
				knownHostsCallback = callback
			}
		}
	}

	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		if knownHostsCallback != nil {
			err := knownHostsCallback(hostname, remote, key)
			if err == nil {
				return nil // Key is known
			}
		}

		// Add the key to known_hosts
		if knownHostsPath != "" {
			if err := addHostKey(knownHostsPath, hostname, key); err != nil {
				fmt.Printf("WARNING: Failed to add host key to known_hosts: %v\n", err)
			} else {
				fmt.Printf("Added host key for %s to known_hosts\n", hostname)
			}
		}
		return nil
	}
}

// addHostKey adds a host key to the known_hosts file
func addHostKey(knownHostsPath, hostname string, key ssh.PublicKey) error {
	// Ensure .ssh directory exists
	sshDir := filepath.Dir(knownHostsPath)
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return fmt.Errorf("failed to create .ssh directory: %w", err)
	}

	// Open file for appending
	f, err := os.OpenFile(knownHostsPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open known_hosts: %w", err)
	}
	defer f.Close()

	// Format the entry
	line := knownhosts.Line([]string{hostname}, key)

	// Write the entry
	if _, err := f.WriteString(line + "\n"); err != nil {
		return fmt.Errorf("failed to write to known_hosts: %w", err)
	}

	return nil
}

// NewClient creates a new SSH client
func NewClient(host parser.Host) (*Client, error) {
	hostKeyCallback, err := createHostKeyCallback(host)
	if err != nil {
		return nil, fmt.Errorf("failed to create host key callback: %w", err)
	}

	config := &ssh.ClientConfig{
		User:            host.SSHUser,
		HostKeyCallback: hostKeyCallback,
	}

	// Add authentication method
	if host.SSHKeyFile != "" {
		key, err := os.ReadFile(host.SSHKeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read SSH key: %w", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("failed to parse SSH key: %w", err)
		}

		config.Auth = append(config.Auth, ssh.PublicKeys(signer))
	} else if host.SSHPassword != "" {
		config.Auth = append(config.Auth, ssh.Password(host.SSHPassword))
	} else {
		return nil, fmt.Errorf("no authentication method specified (password or key file required)")
	}

	// Connect
	addr := fmt.Sprintf("%s:%d", host.SSHHost, host.SSHPort)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	return &Client{
		host:      host,
		client:    client,
		runnerDir: "/tmp/froyo",
	}, nil
}

// NewSimpleClient creates a basic SSH client with host, port, user, and password
func NewSimpleClient(host string, port int, user, password string) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	config.Auth = append(config.Auth, ssh.Password(password))
	config.Auth = append(config.Auth, ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
		answers := make([]string, len(questions))
		for i := range answers {
			answers[i] = password
		}
		return answers, nil
	}))

	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	return client, nil
}

// Close closes the SSH connection
func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// EnsureRunner ensures the froyo-runner binary exists on the remote host
func (c *Client) EnsureRunner(localRunnerPath string) error {
	// Detect remote host OS and architecture
	goos, goarch, err := c.detectRemoteSystem()
	if err != nil {
		return fmt.Errorf("failed to detect remote system: %w", err)
	}

	// Determine the correct binary name based on OS/arch
	var binaryName string
	if goos == "windows" {
		binaryName = fmt.Sprintf("froyo-runner-%s-%s.exe", goos, goarch)
	} else {
		binaryName = fmt.Sprintf("froyo-runner-%s-%s", goos, goarch)
	}

	// Construct path to platform-specific binary
	localBinDir := filepath.Dir(localRunnerPath)
	platformBinary := filepath.Join(localBinDir, binaryName)

	// Check if we have this binary
	if _, err := os.Stat(platformBinary); os.IsNotExist(err) {
		return fmt.Errorf("froyo-runner binary not found for %s/%s: %s\nPlease run 'make build-runner' to build all platform binaries", goos, goarch, platformBinary)
	}

	remoteRunnerPath := filepath.Join(c.runnerDir, "froyo-runner")

	// Create directory
	if err := c.runCommand(fmt.Sprintf("mkdir -p %s", c.runnerDir)); err != nil {
		return fmt.Errorf("failed to create runner directory: %w", err)
	}

	// For development: always upload the latest runner
	// TODO: Add version checking or hash comparison for production

	// Upload the correct platform-specific runner
	if err := c.uploadFile(platformBinary, remoteRunnerPath); err != nil {
		return fmt.Errorf("failed to upload runner: %w", err)
	}

	// Make executable (Unix-like systems)
	if goos != "windows" {
		if err := c.runCommand(fmt.Sprintf("chmod +x %s", remoteRunnerPath)); err != nil {
			return fmt.Errorf("failed to make runner executable: %w", err)
		}
	}

	return nil
}

// detectRemoteSystem detects the OS and architecture of the remote host
func (c *Client) detectRemoteSystem() (goos, goarch string, err error) {
	// Detect OS
	osOutput, err := c.runCommandOutput("uname -s")
	if err != nil {
		// Try Windows detection
		winOutput, winErr := c.runCommandOutput("echo %OS%")
		if winErr == nil && strings.Contains(strings.ToLower(winOutput), "windows") {
			goos = "windows"
		} else {
			return "", "", fmt.Errorf("failed to detect OS: %w", err)
		}
	} else {
		osOutput = strings.TrimSpace(strings.ToLower(osOutput))
		switch {
		case strings.Contains(osOutput, "linux"):
			goos = "linux"
		case strings.Contains(osOutput, "darwin"):
			goos = "darwin"
		case strings.Contains(osOutput, "freebsd"):
			goos = "freebsd"
		default:
			return "", "", fmt.Errorf("unsupported OS: %s", osOutput)
		}
	}

	// Detect architecture
	archOutput, err := c.runCommandOutput("uname -m")
	if err != nil {
		return "", "", fmt.Errorf("failed to detect architecture: %w", err)
	}

	archOutput = strings.TrimSpace(strings.ToLower(archOutput))
	switch archOutput {
	case "x86_64", "amd64":
		goarch = "amd64"
	case "aarch64", "arm64":
		goarch = "arm64"
	case "armv7l", "armv6l":
		goarch = "arm"
	case "i386", "i686":
		goarch = "386"
	default:
		return "", "", fmt.Errorf("unsupported architecture: %s", archOutput)
	}

	return goos, goarch, nil
}

// ExecuteModule executes a WASM module on the remote host
func (c *Client) ExecuteModule(modulePath, taskName string, vars map[string]interface{}) (*TaskOutput, error) {
	// Upload module
	moduleFilename := filepath.Base(modulePath)
	remoteModulePath := filepath.Join(c.runnerDir, moduleFilename)

	if err := c.uploadFile(modulePath, remoteModulePath); err != nil {
		return nil, fmt.Errorf("failed to upload module: %w", err)
	}

	// Prepare input
	input := TaskInput{
		Vars: vars,
		Context: TaskContext{
			Host:     c.host.SSHHost,
			TaskName: taskName,
		},
	}

	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	inputBase64 := base64.StdEncoding.EncodeToString(inputJSON)

	// Execute
	runnerPath := filepath.Join(c.runnerDir, "froyo-runner")
	cmd := fmt.Sprintf("%s --module %s --input-base64 '%s'", runnerPath, remoteModulePath, inputBase64)

	output, err := c.runCommandOutput(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to execute module: %w\nOutput: %s", err, output)
	}

	// Parse output
	var taskOutput TaskOutput
	if err := json.Unmarshal([]byte(output), &taskOutput); err != nil {
		return nil, fmt.Errorf("failed to parse module output: %w\nOutput: %s", err, output)
	}

	return &taskOutput, nil
}

// runCommand runs a command and returns an error if it fails
func (c *Client) runCommand(cmd string) error {
	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	return session.Run(cmd)
}

// runCommandOutput runs a command and returns its output
func (c *Client) runCommandOutput(cmd string) (string, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	return string(output), err
}

// uploadFile uploads a local file to the remote host using native scp command
func (c *Client) uploadFile(localPath, remotePath string) error {
	// Create remote directory if needed
	remoteDir := filepath.Dir(remotePath)
	if err := c.runCommand(fmt.Sprintf("mkdir -p %s", remoteDir)); err != nil {
		return fmt.Errorf("failed to create remote directory: %w", err)
	}

	// Build scp command using the existing SSH connection details
	var scpCmd []string
	addr := fmt.Sprintf("%s@%s", c.host.SSHUser, c.host.SSHHost)

	// Add port if not default
	if c.host.SSHPort != 22 {
		scpCmd = append(scpCmd, "scp", "-P", fmt.Sprintf("%d", c.host.SSHPort))
	} else {
		scpCmd = append(scpCmd, "scp")
	}

	// Add SSH options
	scpCmd = append(scpCmd, "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null")

	// Add key file if specified
	if c.host.SSHKeyFile != "" {
		scpCmd = append(scpCmd, "-i", c.host.SSHKeyFile)
	}

	// Add source and destination
	scpCmd = append(scpCmd, localPath, fmt.Sprintf("%s:%s", addr, remotePath))

	// Execute scp command
	cmd := exec.Command(scpCmd[0], scpCmd[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("scp failed: %w, output: %s", err, string(output))
	}

	// Make the file executable
	if err := c.runCommand(fmt.Sprintf("chmod +x %s", remotePath)); err != nil {
		return fmt.Errorf("failed to chmod: %w", err)
	}

	return nil
}
