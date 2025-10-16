// Package ssh provides SSH-based transport for remote operations.
package ssh

import (
	"context"
	"io"
	"time"
)

// Transport defines the interface for SSH-based remote operations.
// It provides methods for command execution, file transfer, and connection management.
type Transport interface {
	// Connect establishes an SSH connection to the remote host.
	// Returns an error if connection fails or authentication is rejected.
	Connect(ctx context.Context) error

	// Disconnect closes the SSH connection and releases all resources.
	// Should be called when the transport is no longer needed.
	Disconnect() error

	// IsConnected returns true if the transport has an active connection.
	IsConnected() bool

	// HealthCheck verifies the connection is still alive and responsive.
	// Returns an error if the connection is dead or unresponsive.
	HealthCheck(ctx context.Context) error

	// ExecuteCommand runs a command on the remote host.
	// Returns stdout, stderr, and any error that occurred.
	ExecuteCommand(ctx context.Context, cmd string) (stdout string, stderr string, err error)

	// ExecuteCommandWithSudo runs a command with sudo privileges.
	// The sudoPassword parameter can be empty if NOPASSWD is configured.
	ExecuteCommandWithSudo(ctx context.Context, cmd string, sudoPassword string) (stdout string, stderr string, err error)

	// StartInteractiveSession starts an interactive SSH session.
	// Returns stdin writer, stdout reader, stderr reader, and a cleanup function.
	StartInteractiveSession(ctx context.Context) (stdin io.WriteCloser, stdout io.Reader, stderr io.Reader, cleanup func() error, err error)

	// UploadFile uploads a single file to the remote host via SFTP.
	// The mode parameter sets file permissions (e.g., 0644).
	UploadFile(ctx context.Context, localPath string, remotePath string, mode uint32) error

	// DownloadFile downloads a single file from the remote host via SFTP.
	DownloadFile(ctx context.Context, remotePath string, localPath string) error

	// UploadDirectory recursively uploads a directory to the remote host.
	UploadDirectory(ctx context.Context, localPath string, remotePath string) error

	// DownloadDirectory recursively downloads a directory from the remote host.
	DownloadDirectory(ctx context.Context, remotePath string, localPath string) error

	// SetFilePermissions sets file permissions on the remote host.
	SetFilePermissions(ctx context.Context, remotePath string, mode uint32) error

	// SetFileOwnership sets file ownership on the remote host.
	// Requires sudo privileges for changing ownership.
	SetFileOwnership(ctx context.Context, remotePath string, uid int, gid int) error

	// ComputeChecksum calculates the checksum of a remote file.
	// Uses SHA256 by default.
	ComputeChecksum(ctx context.Context, remotePath string) (string, error)

	// GetConnectionInfo returns information about the current connection.
	GetConnectionInfo() ConnectionInfo
}

// ConnectionInfo contains details about an active SSH connection.
type ConnectionInfo struct {
	// Host is the remote hostname or IP address
	Host string

	// Port is the SSH port number
	Port int

	// User is the SSH username
	User string

	// ConnectedAt is when the connection was established
	ConnectedAt time.Time

	// LastActivity is when the connection was last used
	LastActivity time.Time

	// IsPooled indicates if this connection is from a pool
	IsPooled bool
}

// ExecResult represents the result of a command execution.
type ExecResult struct {
	// Stdout is the standard output from the command
	Stdout string

	// Stderr is the standard error output from the command
	Stderr string

	// ExitCode is the command's exit code
	ExitCode int

	// StartedAt is when the command started executing
	StartedAt time.Time

	// FinishedAt is when the command finished
	FinishedAt time.Time

	// Duration is the total execution time
	Duration time.Duration
}

// FileTransferResult represents the result of a file transfer operation.
type FileTransferResult struct {
	// BytesTransferred is the number of bytes transferred
	BytesTransferred int64

	// Duration is the time taken for the transfer
	Duration time.Duration

	// Checksum is the SHA256 checksum of the transferred file (if verified)
	Checksum string

	// StartedAt is when the transfer started
	StartedAt time.Time

	// FinishedAt is when the transfer completed
	FinishedAt time.Time
}

// TransportError represents an error from the transport layer.
type TransportError struct {
	// Op is the operation that failed (e.g., "connect", "exec", "upload")
	Op string

	// Err is the underlying error
	Err error

	// IsTemporary indicates if the error is temporary and can be retried
	IsTemporary bool

	// IsAuthError indicates if the error is related to authentication
	IsAuthError bool
}

func (e *TransportError) Error() string {
	return e.Op + ": " + e.Err.Error()
}

func (e *TransportError) Unwrap() error {
	return e.Err
}

func (e *TransportError) Temporary() bool {
	return e.IsTemporary
}
