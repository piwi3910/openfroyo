// Package client provides a client library for communicating with the micro-runner.
package client

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/openfroyo/openfroyo/pkg/micro_runner/protocol"
)

// Transport defines the interface for uploading and executing the runner.
type Transport interface {
	// Upload uploads the runner binary to the remote host
	Upload(ctx context.Context, localPath, remotePath string) error
	// Execute starts the runner process and returns stdin/stdout
	Execute(ctx context.Context, remotePath string) (stdin io.WriteCloser, stdout io.ReadCloser, err error)
	// Cleanup removes the runner binary from the remote host
	Cleanup(ctx context.Context, remotePath string) error
}

// Client manages communication with a micro-runner instance.
type Client struct {
	transport Transport
	encoder   *protocol.Encoder
	decoder   *protocol.Decoder
	stdin     io.WriteCloser
	stdout    io.ReadCloser
	ready     *protocol.ReadyMessage
	mu        sync.Mutex
	closed    bool
}

// Config contains client configuration options.
type Config struct {
	Transport     Transport
	RunnerPath    string // Path to local runner binary
	RemotePath    string // Path on remote host
	StartupTimeout time.Duration
}

// NewClient creates a new micro-runner client.
func NewClient(cfg *Config) (*Client, error) {
	if cfg.Transport == nil {
		return nil, fmt.Errorf("transport is required")
	}
	if cfg.RunnerPath == "" {
		return nil, fmt.Errorf("runner path is required")
	}
	if cfg.RemotePath == "" {
		cfg.RemotePath = "/tmp/micro-runner"
	}
	if cfg.StartupTimeout == 0 {
		cfg.StartupTimeout = 10 * time.Second
	}

	return &Client{
		transport: cfg.Transport,
	}, nil
}

// Start uploads the runner binary and starts the runner process.
func (c *Client) Start(ctx context.Context, cfg *Config) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return fmt.Errorf("client is closed")
	}

	// Upload runner binary
	if err := c.transport.Upload(ctx, cfg.RunnerPath, cfg.RemotePath); err != nil {
		return fmt.Errorf("failed to upload runner: %w", err)
	}

	// Start runner process
	stdin, stdout, err := c.transport.Execute(ctx, cfg.RemotePath)
	if err != nil {
		return fmt.Errorf("failed to start runner: %w", err)
	}

	c.stdin = stdin
	c.stdout = stdout
	c.encoder = protocol.NewEncoder(stdin)
	c.decoder = protocol.NewDecoder(stdout)

	// Wait for READY message
	readyCtx, cancel := context.WithTimeout(ctx, cfg.StartupTimeout)
	defer cancel()

	readyCh := make(chan *protocol.ReadyMessage, 1)
	errCh := make(chan error, 1)

	go func() {
		msg, err := c.decoder.Decode()
		if err != nil {
			errCh <- err
			return
		}
		if msg.Type != protocol.MessageTypeReady {
			errCh <- fmt.Errorf("expected READY, got %s", msg.Type)
			return
		}
		var ready protocol.ReadyMessage
		if err := protocol.ParseParams(msg.Data, &ready); err != nil {
			errCh <- err
			return
		}
		readyCh <- &ready
	}()

	select {
	case <-readyCtx.Done():
		return fmt.Errorf("timeout waiting for READY message")
	case err := <-errCh:
		return fmt.Errorf("failed to receive READY: %w", err)
	case ready := <-readyCh:
		c.ready = ready
		return nil
	}
}

// Execute sends a command to the runner and waits for completion.
func (c *Client) Execute(ctx context.Context, cmd *protocol.CommandMessage) (*protocol.DoneMessage, error) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil, fmt.Errorf("client is closed")
	}
	c.mu.Unlock()

	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("invalid command: %w", err)
	}

	// Send command
	if err := c.encoder.Encode(protocol.MessageTypeCommand, cmd); err != nil {
		return nil, fmt.Errorf("failed to send command: %w", err)
	}

	// Wait for response
	for {
		msg, err := c.decoder.Decode()
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		switch msg.Type {
		case protocol.MessageTypeEvent:
			// Handle progress events (could emit to a channel)
			var event protocol.EventMessage
			if err := protocol.ParseParams(msg.Data, &event); err != nil {
				return nil, fmt.Errorf("failed to parse event: %w", err)
			}
			// For now, just ignore events
			// In a real implementation, you'd emit these to a channel

		case protocol.MessageTypeDone:
			var done protocol.DoneMessage
			if err := protocol.ParseParams(msg.Data, &done); err != nil {
				return nil, fmt.Errorf("failed to parse done: %w", err)
			}
			if done.CommandID != cmd.ID {
				return nil, fmt.Errorf("command ID mismatch: expected %s, got %s", cmd.ID, done.CommandID)
			}
			return &done, nil

		case protocol.MessageTypeError:
			var errMsg protocol.ErrorMessage
			if err := protocol.ParseParams(msg.Data, &errMsg); err != nil {
				return nil, fmt.Errorf("failed to parse error: %w", err)
			}
			if errMsg.CommandID != "" && errMsg.CommandID != cmd.ID {
				return nil, fmt.Errorf("command ID mismatch: expected %s, got %s", cmd.ID, errMsg.CommandID)
			}
			return nil, fmt.Errorf("command failed: %s - %s", errMsg.Code, errMsg.Message)

		case protocol.MessageTypeExit:
			return nil, fmt.Errorf("runner exited unexpectedly")

		default:
			return nil, fmt.Errorf("unexpected message type: %s", msg.Type)
		}
	}
}

// ExecuteWithEvents sends a command and streams events to a channel.
func (c *Client) ExecuteWithEvents(ctx context.Context, cmd *protocol.CommandMessage, eventCh chan<- *protocol.EventMessage) (*protocol.DoneMessage, error) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil, fmt.Errorf("client is closed")
	}
	c.mu.Unlock()

	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("invalid command: %w", err)
	}

	// Send command
	if err := c.encoder.Encode(protocol.MessageTypeCommand, cmd); err != nil {
		return nil, fmt.Errorf("failed to send command: %w", err)
	}

	// Wait for response
	for {
		msg, err := c.decoder.Decode()
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		switch msg.Type {
		case protocol.MessageTypeEvent:
			var event protocol.EventMessage
			if err := protocol.ParseParams(msg.Data, &event); err != nil {
				return nil, fmt.Errorf("failed to parse event: %w", err)
			}
			if eventCh != nil {
				eventCh <- &event
			}

		case protocol.MessageTypeDone:
			var done protocol.DoneMessage
			if err := protocol.ParseParams(msg.Data, &done); err != nil {
				return nil, fmt.Errorf("failed to parse done: %w", err)
			}
			if done.CommandID != cmd.ID {
				return nil, fmt.Errorf("command ID mismatch: expected %s, got %s", cmd.ID, done.CommandID)
			}
			return &done, nil

		case protocol.MessageTypeError:
			var errMsg protocol.ErrorMessage
			if err := protocol.ParseParams(msg.Data, &errMsg); err != nil {
				return nil, fmt.Errorf("failed to parse error: %w", err)
			}
			if errMsg.CommandID != "" && errMsg.CommandID != cmd.ID {
				return nil, fmt.Errorf("command ID mismatch: expected %s, got %s", cmd.ID, errMsg.CommandID)
			}
			return nil, fmt.Errorf("command failed: %s - %s", errMsg.Code, errMsg.Message)

		case protocol.MessageTypeExit:
			return nil, fmt.Errorf("runner exited unexpectedly")

		default:
			return nil, fmt.Errorf("unexpected message type: %s", msg.Type)
		}
	}
}

// Ready returns the READY message received during startup.
func (c *Client) Ready() *protocol.ReadyMessage {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ready
}

// Close closes the client and cleans up resources.
func (c *Client) Close(ctx context.Context, remotePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true

	var errs []error

	// Close stdin to signal runner to exit
	if c.stdin != nil {
		if err := c.stdin.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close stdin: %w", err))
		}
	}

	// Close stdout
	if c.stdout != nil {
		if err := c.stdout.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close stdout: %w", err))
		}
	}

	// Cleanup remote binary (if it wasn't self-deleted)
	if remotePath != "" {
		if err := c.transport.Cleanup(ctx, remotePath); err != nil {
			// Ignore errors as runner may have self-deleted
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}

	return nil
}

// BuildRunnerBinary builds the micro-runner binary for the target platform.
func BuildRunnerBinary(ctx context.Context, outputPath, goos, goarch string) error {
	cmd := fmt.Sprintf("CGO_ENABLED=0 GOOS=%s GOARCH=%s go build -ldflags='-s -w' -o %s ./cmd/micro-runner",
		goos, goarch, outputPath)

	// This is a placeholder - in production you'd execute this command
	// For now, just return an error if the file doesn't exist
	if _, err := os.Stat(outputPath); err != nil {
		return fmt.Errorf("runner binary not found at %s, build with: %s", outputPath, cmd)
	}

	return nil
}
