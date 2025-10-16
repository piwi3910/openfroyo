package handlers

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/openfroyo/openfroyo/pkg/micro_runner/protocol"
)

// SudoersEnsureHandler handles sudoers rule management.
type SudoersEnsureHandler struct{}

// Handle manages sudoers rules for a user.
func (h *SudoersEnsureHandler) Handle(ctx context.Context, params *protocol.SudoersEnsureParams, eventCh chan<- *protocol.EventMessage) (*protocol.SudoersEnsureResult, error) {
	if params.User == "" {
		return nil, fmt.Errorf("user is required")
	}

	sudoersDir := "/etc/sudoers.d"
	fileName := fmt.Sprintf("froyo-%s", params.User)
	filePath := filepath.Join(sudoersDir, fileName)

	result := &protocol.SudoersEnsureResult{
		FilePath: filePath,
	}

	// Check if file exists
	_, err := os.Stat(filePath)
	fileExists := err == nil

	switch params.State {
	case "present":
		// Build sudoers rule
		rule := h.buildSudoersRule(params.User, params.Commands, params.NoPasswd)

		if fileExists {
			// Check if content needs updating
			existing, err := os.ReadFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read existing file: %w", err)
			}
			if string(existing) == rule {
				result.Changed = false
				result.Action = "already_present"
				return result, nil
			}
			result.Action = "updated"
		} else {
			result.Action = "created"
		}

		// Write the rule
		if err := os.WriteFile(filePath, []byte(rule), 0440); err != nil {
			return nil, fmt.Errorf("failed to write sudoers file: %w", err)
		}

		// Validate sudoers syntax
		if err := h.validateSudoers(ctx, filePath); err != nil {
			// Remove invalid file
			os.Remove(filePath)
			return nil, fmt.Errorf("invalid sudoers syntax: %w", err)
		}

		result.Changed = true

	case "absent":
		if !fileExists {
			result.Changed = false
			result.Action = "already_absent"
		} else {
			if err := os.Remove(filePath); err != nil {
				return nil, fmt.Errorf("failed to remove sudoers file: %w", err)
			}
			result.Changed = true
			result.Action = "removed"
		}

	default:
		return nil, fmt.Errorf("invalid state: %s", params.State)
	}

	return result, nil
}

func (h *SudoersEnsureHandler) buildSudoersRule(user string, commands []string, noPasswd bool) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# Managed by OpenFroyo micro-runner\n"))
	b.WriteString(fmt.Sprintf("# User: %s\n", user))

	passwd := "PASSWD"
	if noPasswd {
		passwd = "NOPASSWD"
	}

	if len(commands) == 0 {
		// No commands specified = all commands
		b.WriteString(fmt.Sprintf("%s ALL=(%s) ALL\n", user, passwd))
	} else {
		// Specific commands
		cmdList := strings.Join(commands, ", ")
		b.WriteString(fmt.Sprintf("%s ALL=(%s) %s\n", user, passwd, cmdList))
	}

	return b.String()
}

func (h *SudoersEnsureHandler) validateSudoers(ctx context.Context, filePath string) error {
	// Use visudo to validate syntax
	cmd := exec.CommandContext(ctx, "visudo", "-c", "-f", filePath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("visudo validation failed: %w", err)
	}
	return nil
}
