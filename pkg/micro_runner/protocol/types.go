// Package protocol defines the JSON-over-stdio communication protocol
// for the OpenFroyo micro-runner.
package protocol

import (
	"encoding/json"
	"fmt"
	"time"
)

// MessageType represents the type of message in the protocol.
type MessageType string

const (
	// MessageTypeReady indicates the runner is ready to receive commands
	MessageTypeReady MessageType = "READY"
	// MessageTypeCommand indicates a command from the controller
	MessageTypeCommand MessageType = "CMD"
	// MessageTypeEvent indicates a progress event from the runner
	MessageTypeEvent MessageType = "EVENT"
	// MessageTypeDone indicates successful completion
	MessageTypeDone MessageType = "DONE"
	// MessageTypeError indicates an error occurred
	MessageTypeError MessageType = "ERROR"
	// MessageTypeExit indicates the runner is exiting
	MessageTypeExit MessageType = "EXIT"
)

// CommandType represents the type of command to execute.
type CommandType string

const (
	// CommandTypeExec executes a shell command
	CommandTypeExec CommandType = "exec"
	// CommandTypeFileWrite writes content to a file
	CommandTypeFileWrite CommandType = "file.write"
	// CommandTypeFileRead reads content from a file
	CommandTypeFileRead CommandType = "file.read"
	// CommandTypePkgEnsure ensures a package is installed or removed
	CommandTypePkgEnsure CommandType = "pkg.ensure"
	// CommandTypeServiceReload reloads a systemd service
	CommandTypeServiceReload CommandType = "service.reload"
	// CommandTypeSudoersEnsure manages sudoers rules
	CommandTypeSudoersEnsure CommandType = "sudoers.ensure"
	// CommandTypeSSHDHarden applies SSH hardening configuration
	CommandTypeSSHDHarden CommandType = "sshd.harden"
)

// Message is the base message structure for all protocol messages.
type Message struct {
	Type      MessageType     `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data,omitempty"`
}

// ReadyMessage is sent when the runner is ready to receive commands.
type ReadyMessage struct {
	Version  string            `json:"version"`
	Platform string            `json:"platform"`
	Arch     string            `json:"arch"`
	PID      int               `json:"pid"`
	Caps     map[string]bool   `json:"capabilities"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// CommandMessage contains a command to execute.
type CommandMessage struct {
	ID             string            `json:"id"`
	Type           CommandType       `json:"type"`
	IdempotencyKey string            `json:"idempotency_key,omitempty"`
	Timeout        int               `json:"timeout"` // seconds
	Params         json.RawMessage   `json:"params"`
	Signature      string            `json:"signature,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// EventMessage contains progress information during command execution.
type EventMessage struct {
	CommandID string            `json:"command_id"`
	Level     string            `json:"level"` // info, warn, debug
	Message   string            `json:"message"`
	Progress  *ProgressInfo     `json:"progress,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// ProgressInfo contains progress tracking information.
type ProgressInfo struct {
	Current int    `json:"current"`
	Total   int    `json:"total"`
	Unit    string `json:"unit"`
}

// DoneMessage indicates successful command completion.
type DoneMessage struct {
	CommandID string            `json:"command_id"`
	Result    json.RawMessage   `json:"result"`
	Duration  float64           `json:"duration"` // seconds
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// ErrorMessage indicates an error occurred.
type ErrorMessage struct {
	CommandID  string            `json:"command_id,omitempty"`
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Details    map[string]string `json:"details,omitempty"`
	Retryable  bool              `json:"retryable"`
	RetryAfter int               `json:"retry_after,omitempty"` // seconds
}

// ExitMessage is sent before the runner terminates.
type ExitMessage struct {
	Reason        string `json:"reason"`
	ExitCode      int    `json:"exit_code"`
	SelfDeleted   bool   `json:"self_deleted"`
	CommandsTotal int    `json:"commands_total"`
}

// Command parameter structures for each command type

// ExecParams contains parameters for shell command execution.
type ExecParams struct {
	Command     string            `json:"command"`
	Args        []string          `json:"args,omitempty"`
	WorkDir     string            `json:"work_dir,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	Shell       string            `json:"shell,omitempty"` // defaults to /bin/sh
	CaptureOut  bool              `json:"capture_out"`
	CaptureErr  bool              `json:"capture_err"`
	StreamLines bool              `json:"stream_lines"`
}

// ExecResult contains the result of command execution.
type ExecResult struct {
	ExitCode int               `json:"exit_code"`
	Stdout   string            `json:"stdout,omitempty"`
	Stderr   string            `json:"stderr,omitempty"`
	Duration float64           `json:"duration"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// FileWriteParams contains parameters for writing a file.
type FileWriteParams struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Mode    string `json:"mode,omitempty"`    // e.g., "0644"
	Owner   string `json:"owner,omitempty"`   // e.g., "root"
	Group   string `json:"group,omitempty"`   // e.g., "root"
	Backup  bool   `json:"backup,omitempty"`  // create .bak before write
	Create  bool   `json:"create"`            // create if not exists
}

// FileWriteResult contains the result of file write operation.
type FileWriteResult struct {
	BytesWritten int64  `json:"bytes_written"`
	Created      bool   `json:"created"`
	BackupPath   string `json:"backup_path,omitempty"`
	Checksum     string `json:"checksum"` // SHA256
}

// FileReadParams contains parameters for reading a file.
type FileReadParams struct {
	Path     string `json:"path"`
	MaxBytes int64  `json:"max_bytes,omitempty"` // limit read size
}

// FileReadResult contains the result of file read operation.
type FileReadResult struct {
	Content   string `json:"content"`
	Size      int64  `json:"size"`
	Mode      string `json:"mode"`
	Owner     string `json:"owner"`
	Group     string `json:"group"`
	Checksum  string `json:"checksum"` // SHA256
	Truncated bool   `json:"truncated"`
}

// PkgEnsureParams contains parameters for package management.
type PkgEnsureParams struct {
	Name    string   `json:"name"`
	Version string   `json:"version,omitempty"` // empty = latest
	State   string   `json:"state"`             // present, absent, latest
	Manager string   `json:"manager,omitempty"` // apt, dnf, yum, zypper (auto-detect if empty)
	Options []string `json:"options,omitempty"` // additional flags
}

// PkgEnsureResult contains the result of package operation.
type PkgEnsureResult struct {
	Changed         bool   `json:"changed"`
	PreviousVersion string `json:"previous_version,omitempty"`
	InstalledVersion string `json:"installed_version,omitempty"`
	Action          string `json:"action"` // installed, removed, upgraded, already_present
}

// ServiceReloadParams contains parameters for service management.
type ServiceReloadParams struct {
	Name   string `json:"name"`
	Action string `json:"action"` // reload, restart, start, stop, enable, disable
}

// ServiceReloadResult contains the result of service operation.
type ServiceReloadResult struct {
	Changed   bool   `json:"changed"`
	Action    string `json:"action"`
	Status    string `json:"status"`    // active, inactive, failed
	Enabled   bool   `json:"enabled"`
	SubState  string `json:"sub_state"` // running, dead, exited
}

// SudoersEnsureParams contains parameters for sudoers management.
type SudoersEnsureParams struct {
	User     string   `json:"user"`
	Commands []string `json:"commands"` // allowed commands (full paths)
	NoPasswd bool     `json:"no_passwd"`
	State    string   `json:"state"` // present, absent
}

// SudoersEnsureResult contains the result of sudoers operation.
type SudoersEnsureResult struct {
	Changed  bool   `json:"changed"`
	FilePath string `json:"file_path"` // e.g., /etc/sudoers.d/froyo-user
	Action   string `json:"action"`    // created, removed, updated
}

// SSHDHardenParams contains parameters for SSH hardening.
type SSHDHardenParams struct {
	DisablePasswordAuth bool     `json:"disable_password_auth"`
	DisableRootLogin    bool     `json:"disable_root_login"`
	AllowUsers          []string `json:"allow_users,omitempty"`
	Port                int      `json:"port,omitempty"` // custom SSH port
	TestConnection      bool     `json:"test_connection"` // verify before applying
}

// SSHDHardenResult contains the result of SSH hardening.
type SSHDHardenResult struct {
	Changed       bool     `json:"changed"`
	BackupPath    string   `json:"backup_path"`
	ModifiedKeys  []string `json:"modified_keys"`
	ServiceAction string   `json:"service_action"` // reloaded, none
}

// Validation methods

// Validate checks if the message type is valid.
func (mt MessageType) Validate() error {
	switch mt {
	case MessageTypeReady, MessageTypeCommand, MessageTypeEvent,
		MessageTypeDone, MessageTypeError, MessageTypeExit:
		return nil
	default:
		return fmt.Errorf("invalid message type: %s", mt)
	}
}

// Validate checks if the command type is valid.
func (ct CommandType) Validate() error {
	switch ct {
	case CommandTypeExec, CommandTypeFileWrite, CommandTypeFileRead,
		CommandTypePkgEnsure, CommandTypeServiceReload,
		CommandTypeSudoersEnsure, CommandTypeSSHDHarden:
		return nil
	default:
		return fmt.Errorf("invalid command type: %s", ct)
	}
}

// Validate checks if the command message is valid.
func (cmd *CommandMessage) Validate() error {
	if cmd.ID == "" {
		return fmt.Errorf("command ID is required")
	}
	if err := cmd.Type.Validate(); err != nil {
		return err
	}
	if cmd.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	if len(cmd.Params) == 0 {
		return fmt.Errorf("command params are required")
	}
	return nil
}

// Validate checks if the event message is valid.
func (evt *EventMessage) Validate() error {
	if evt.CommandID == "" {
		return fmt.Errorf("command ID is required")
	}
	if evt.Level == "" {
		evt.Level = "info"
	}
	validLevels := map[string]bool{"info": true, "warn": true, "debug": true}
	if !validLevels[evt.Level] {
		return fmt.Errorf("invalid event level: %s", evt.Level)
	}
	return nil
}
