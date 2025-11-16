package ssh

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"

	"github.com/piwi3910/openfroyo/internal/parser"
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

// NewClient creates a new SSH client
func NewClient(host parser.Host) (*Client, error) {
	config := &ssh.ClientConfig{
		User:            host.SSHUser,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // For MVP - should validate in production
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

// Close closes the SSH connection
func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// EnsureRunner ensures the froyo-runner binary exists on the remote host
func (c *Client) EnsureRunner(localRunnerPath string) error {
	remoteRunnerPath := filepath.Join(c.runnerDir, "froyo-runner")

	// Create directory
	if err := c.runCommand(fmt.Sprintf("mkdir -p %s", c.runnerDir)); err != nil {
		return fmt.Errorf("failed to create runner directory: %w", err)
	}

	// For development: always upload the latest runner
	// TODO: Add version checking or hash comparison for production
	// checkCmd := fmt.Sprintf("test -f %s && echo exists || echo missing", remoteRunnerPath)
	// output, err := c.runCommandOutput(checkCmd)
	// if err != nil {
	// 	return fmt.Errorf("failed to check for runner: %w", err)
	// }
	//
	// if strings.TrimSpace(output) == "exists" {
	// 	return nil // Runner already exists
	// }

	// Upload runner
	if err := c.uploadFile(localRunnerPath, remoteRunnerPath); err != nil {
		return fmt.Errorf("failed to upload runner: %w", err)
	}

	// Make executable
	if err := c.runCommand(fmt.Sprintf("chmod +x %s", remoteRunnerPath)); err != nil {
		return fmt.Errorf("failed to make runner executable: %w", err)
	}

	return nil
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

// uploadFile uploads a local file to the remote host using SCP
func (c *Client) uploadFile(localPath, remotePath string) error {
	// Read local file
	data, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("failed to read local file: %w", err)
	}

	// Create remote directory if needed
	remoteDir := filepath.Dir(remotePath)
	if err := c.runCommand(fmt.Sprintf("mkdir -p %s", remoteDir)); err != nil {
		return fmt.Errorf("failed to create remote directory: %w", err)
	}

	// Open SCP session
	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	// Create pipe for stdin
	stdin, err := session.StdinPipe()
	if err != nil {
		return err
	}

	// Start SCP command
	if err := session.Start(fmt.Sprintf("scp -t %s", remotePath)); err != nil {
		return err
	}

	// Send file via SCP protocol
	fmt.Fprintf(stdin, "C0755 %d %s\n", len(data), filepath.Base(remotePath))
	io.Copy(stdin, strings.NewReader(string(data)))
	fmt.Fprint(stdin, "\x00")
	stdin.Close()

	return session.Wait()
}
