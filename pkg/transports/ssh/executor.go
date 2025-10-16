package ssh

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
)

// executor handles command execution over SSH.
type executor struct {
	client *SSHClient
	config *Config
}

// ExecuteCommand runs a command on the remote host.
func (c *SSHClient) ExecuteCommand(ctx context.Context, cmd string) (stdout string, stderr string, err error) {
	if c.executor == nil {
		return "", "", &TransportError{
			Op:          "execute",
			Err:         fmt.Errorf("executor not initialized"),
			IsTemporary: false,
			IsAuthError: false,
		}
	}
	return c.executor.execute(ctx, cmd, false, "")
}

// ExecuteCommandWithSudo runs a command with sudo privileges.
func (c *SSHClient) ExecuteCommandWithSudo(ctx context.Context, cmd string, sudoPassword string) (stdout string, stderr string, err error) {
	if c.executor == nil {
		return "", "", &TransportError{
			Op:          "execute-sudo",
			Err:         fmt.Errorf("executor not initialized"),
			IsTemporary: false,
			IsAuthError: false,
		}
	}
	return c.executor.execute(ctx, cmd, true, sudoPassword)
}

// execute is the internal implementation of command execution.
func (e *executor) execute(ctx context.Context, cmd string, useSudo bool, sudoPassword string) (stdout string, stderr string, err error) {
	startTime := time.Now()

	log.Debug().
		Str("command", cmd).
		Bool("sudo", useSudo).
		Msg("executing command")

	// Get the SSH client
	sshClient, err := e.client.GetClient()
	if err != nil {
		return "", "", err
	}

	// Create a new session
	session, err := sshClient.NewSession()
	if err != nil {
		return "", "", &TransportError{
			Op:          "execute",
			Err:         fmt.Errorf("failed to create session: %w", err),
			IsTemporary: true,
			IsAuthError: false,
		}
	}
	defer session.Close()

	// Set up buffers for stdout and stderr
	var stdoutBuf, stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	// Prepare the command
	finalCmd := cmd
	if useSudo {
		if sudoPassword != "" {
			// If sudo password is provided, use it via stdin
			finalCmd = fmt.Sprintf("echo '%s' | sudo -S %s", sudoPassword, cmd)
		} else {
			// NOPASSWD sudo
			finalCmd = fmt.Sprintf("sudo %s", cmd)
		}
	}

	// Create a channel for command completion
	doneChan := make(chan error, 1)

	go func() {
		doneChan <- session.Run(finalCmd)
	}()

	// Wait for command to complete or timeout
	var execErr error
	select {
	case <-ctx.Done():
		// Context cancelled, try to signal the session
		_ = session.Signal(ssh.SIGTERM)
		time.Sleep(100 * time.Millisecond)
		_ = session.Signal(ssh.SIGKILL)
		execErr = ctx.Err()
	case execErr = <-doneChan:
		// Command completed
	}

	duration := time.Since(startTime)

	stdout = strings.TrimSpace(stdoutBuf.String())
	stderr = strings.TrimSpace(stderrBuf.String())

	log.Debug().
		Str("command", cmd).
		Int("stdout_len", len(stdout)).
		Int("stderr_len", len(stderr)).
		Dur("duration", duration).
		Err(execErr).
		Msg("command completed")

	if execErr != nil {
		// Check if it's an exit error
		if exitErr, ok := execErr.(*ssh.ExitError); ok {
			// Command ran but returned non-zero exit code
			return stdout, stderr, &TransportError{
				Op:          "execute",
				Err:         fmt.Errorf("command exited with code %d: %s", exitErr.ExitStatus(), stderr),
				IsTemporary: false,
				IsAuthError: false,
			}
		}
		// Other error (connection issue, etc.)
		return stdout, stderr, &TransportError{
			Op:          "execute",
			Err:         execErr,
			IsTemporary: true,
			IsAuthError: false,
		}
	}

	return stdout, stderr, nil
}

// StartInteractiveSession starts an interactive SSH session.
func (c *SSHClient) StartInteractiveSession(ctx context.Context) (stdin io.WriteCloser, stdout io.Reader, stderr io.Reader, cleanup func() error, err error) {
	log.Debug().Msg("starting interactive session")

	// Get the SSH client
	sshClient, err := c.GetClient()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Create a new session
	session, err := sshClient.NewSession()
	if err != nil {
		return nil, nil, nil, nil, &TransportError{
			Op:          "interactive-session",
			Err:         fmt.Errorf("failed to create session: %w", err),
			IsTemporary: true,
			IsAuthError: false,
		}
	}

	// Set up pipes for stdin, stdout, and stderr
	stdinPipe, err := session.StdinPipe()
	if err != nil {
		session.Close()
		return nil, nil, nil, nil, &TransportError{
			Op:          "interactive-session",
			Err:         fmt.Errorf("failed to create stdin pipe: %w", err),
			IsTemporary: true,
			IsAuthError: false,
		}
	}

	stdoutPipe, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		return nil, nil, nil, nil, &TransportError{
			Op:          "interactive-session",
			Err:         fmt.Errorf("failed to create stdout pipe: %w", err),
			IsTemporary: true,
			IsAuthError: false,
		}
	}

	stderrPipe, err := session.StderrPipe()
	if err != nil {
		session.Close()
		return nil, nil, nil, nil, &TransportError{
			Op:          "interactive-session",
			Err:         fmt.Errorf("failed to create stderr pipe: %w", err),
			IsTemporary: true,
			IsAuthError: false,
		}
	}

	// Request a pseudo-terminal
	if err := session.RequestPty("xterm", 80, 40, ssh.TerminalModes{
		ssh.ECHO:          1,     // enable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}); err != nil {
		session.Close()
		return nil, nil, nil, nil, &TransportError{
			Op:          "interactive-session",
			Err:         fmt.Errorf("failed to request pseudo-terminal: %w", err),
			IsTemporary: true,
			IsAuthError: false,
		}
	}

	// Start the shell
	if err := session.Shell(); err != nil {
		session.Close()
		return nil, nil, nil, nil, &TransportError{
			Op:          "interactive-session",
			Err:         fmt.Errorf("failed to start shell: %w", err),
			IsTemporary: true,
			IsAuthError: false,
		}
	}

	// Cleanup function to close the session
	cleanupFunc := func() error {
		log.Debug().Msg("closing interactive session")
		return session.Close()
	}

	log.Info().Msg("interactive session started")
	return stdinPipe, stdoutPipe, stderrPipe, cleanupFunc, nil
}

// ExecuteScript uploads and executes a script on the remote host.
// This is useful for complex operations that are better expressed as scripts.
func (e *executor) ExecuteScript(ctx context.Context, script string, interpreter string, useSudo bool, sudoPassword string) (stdout string, stderr string, err error) {
	// Create a temporary file for the script
	tmpFile := fmt.Sprintf("/tmp/openfroyo-script-%d.sh", time.Now().UnixNano())

	log.Debug().
		Str("tmpfile", tmpFile).
		Str("interpreter", interpreter).
		Bool("sudo", useSudo).
		Msg("executing script")

	// Write the script to the remote host
	writeCmd := fmt.Sprintf("cat > %s << 'OPENFROYO_SCRIPT_EOF'\n%s\nOPENFROYO_SCRIPT_EOF", tmpFile, script)
	_, _, err = e.execute(ctx, writeCmd, false, "")
	if err != nil {
		return "", "", fmt.Errorf("failed to write script: %w", err)
	}

	// Make the script executable
	chmodCmd := fmt.Sprintf("chmod +x %s", tmpFile)
	_, _, err = e.execute(ctx, chmodCmd, useSudo, sudoPassword)
	if err != nil {
		return "", "", fmt.Errorf("failed to make script executable: %w", err)
	}

	// Execute the script
	var execCmd string
	if interpreter != "" {
		execCmd = fmt.Sprintf("%s %s", interpreter, tmpFile)
	} else {
		execCmd = tmpFile
	}

	stdout, stderr, err = e.execute(ctx, execCmd, useSudo, sudoPassword)

	// Clean up the temporary file
	rmCmd := fmt.Sprintf("rm -f %s", tmpFile)
	_, _, cleanupErr := e.execute(ctx, rmCmd, useSudo, sudoPassword)
	if cleanupErr != nil {
		log.Warn().Err(cleanupErr).Msg("failed to clean up script file")
	}

	return stdout, stderr, err
}

// ExecuteWithTimeout executes a command with a specific timeout.
func (e *executor) ExecuteWithTimeout(cmd string, timeout time.Duration, useSudo bool, sudoPassword string) (stdout string, stderr string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return e.execute(ctx, cmd, useSudo, sudoPassword)
}

// ExecuteBatch executes multiple commands in sequence.
// Stops on first error if stopOnError is true.
func (e *executor) ExecuteBatch(ctx context.Context, commands []string, stopOnError bool, useSudo bool, sudoPassword string) ([]ExecResult, error) {
	results := make([]ExecResult, 0, len(commands))

	for i, cmd := range commands {
		startTime := time.Now()

		log.Debug().
			Int("index", i).
			Str("command", cmd).
			Msg("executing batch command")

		stdout, stderr, err := e.execute(ctx, cmd, useSudo, sudoPassword)

		result := ExecResult{
			Stdout:     stdout,
			Stderr:     stderr,
			ExitCode:   0,
			StartedAt:  startTime,
			FinishedAt: time.Now(),
			Duration:   time.Since(startTime),
		}

		if err != nil {
			if transportErr, ok := err.(*TransportError); ok {
				if exitErr, ok := transportErr.Err.(*ssh.ExitError); ok {
					result.ExitCode = exitErr.ExitStatus()
				} else {
					result.ExitCode = -1
				}
			} else {
				result.ExitCode = -1
			}

			results = append(results, result)

			if stopOnError {
				return results, fmt.Errorf("command %d failed: %w", i, err)
			}
		} else {
			results = append(results, result)
		}
	}

	return results, nil
}
