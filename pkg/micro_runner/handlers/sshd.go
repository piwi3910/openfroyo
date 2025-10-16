package handlers

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/openfroyo/openfroyo/pkg/micro_runner/protocol"
)

// SSHDHardenHandler handles SSH hardening operations.
type SSHDHardenHandler struct{}

// Handle applies SSH hardening configuration.
func (h *SSHDHardenHandler) Handle(ctx context.Context, params *protocol.SSHDHardenParams, eventCh chan<- *protocol.EventMessage) (*protocol.SSHDHardenResult, error) {
	sshdConfigPath := "/etc/ssh/sshd_config"

	result := &protocol.SSHDHardenResult{
		ModifiedKeys: []string{},
	}

	// Create backup
	backupPath := sshdConfigPath + ".bak"
	if err := copyFile(sshdConfigPath, backupPath); err != nil {
		return nil, fmt.Errorf("failed to create backup: %w", err)
	}
	result.BackupPath = backupPath

	// Read current config
	config, err := h.readSSHDConfig(sshdConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read sshd_config: %w", err)
	}

	originalConfig := make(map[string]string)
	for k, v := range config {
		originalConfig[k] = v
	}

	// Apply hardening settings
	if params.DisablePasswordAuth {
		config["PasswordAuthentication"] = "no"
		result.ModifiedKeys = append(result.ModifiedKeys, "PasswordAuthentication")
	}

	if params.DisableRootLogin {
		config["PermitRootLogin"] = "no"
		result.ModifiedKeys = append(result.ModifiedKeys, "PermitRootLogin")
	}

	if len(params.AllowUsers) > 0 {
		config["AllowUsers"] = strings.Join(params.AllowUsers, " ")
		result.ModifiedKeys = append(result.ModifiedKeys, "AllowUsers")
	}

	if params.Port > 0 {
		config["Port"] = fmt.Sprintf("%d", params.Port)
		result.ModifiedKeys = append(result.ModifiedKeys, "Port")
	}

	// Check if anything changed
	changed := false
	for _, key := range result.ModifiedKeys {
		if originalConfig[key] != config[key] {
			changed = true
			break
		}
	}

	if !changed {
		result.Changed = false
		result.ServiceAction = "none"
		return result, nil
	}

	// Write new config
	if err := h.writeSSHDConfig(sshdConfigPath, config); err != nil {
		// Restore backup
		copyFile(backupPath, sshdConfigPath)
		return nil, fmt.Errorf("failed to write sshd_config: %w", err)
	}

	// Test connection if requested
	if params.TestConnection {
		if err := h.testSSHDConfig(ctx); err != nil {
			// Restore backup
			copyFile(backupPath, sshdConfigPath)
			return nil, fmt.Errorf("sshd config test failed: %w", err)
		}
	}

	// Reload sshd service
	if err := h.reloadSSHD(ctx); err != nil {
		// Restore backup
		copyFile(backupPath, sshdConfigPath)
		return nil, fmt.Errorf("failed to reload sshd: %w", err)
	}

	result.Changed = true
	result.ServiceAction = "reloaded"

	return result, nil
}

func (h *SSHDHardenHandler) readSSHDConfig(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := make(map[string]string)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 {
			key := parts[0]
			value := strings.Join(parts[1:], " ")
			config[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return config, nil
}

func (h *SSHDHardenHandler) writeSSHDConfig(path string, config map[string]string) error {
	// Read original file to preserve comments and structure
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	processedKeys := make(map[string]bool)

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Keep comments and empty lines
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			lines = append(lines, line)
			continue
		}

		// Parse key
		parts := strings.Fields(trimmed)
		if len(parts) >= 1 {
			key := parts[0]
			if newValue, ok := config[key]; ok {
				// Replace with new value
				lines = append(lines, fmt.Sprintf("%s %s", key, newValue))
				processedKeys[key] = true
			} else {
				// Keep original line
				lines = append(lines, line)
			}
		}
	}
	file.Close()

	if err := scanner.Err(); err != nil {
		return err
	}

	// Add any new keys that weren't in the original file
	for key, value := range config {
		if !processedKeys[key] {
			lines = append(lines, fmt.Sprintf("%s %s", key, value))
		}
	}

	// Write back
	content := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(path, []byte(content), 0644)
}

func (h *SSHDHardenHandler) testSSHDConfig(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "sshd", "-t")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sshd -t failed: %w", err)
	}
	return nil
}

func (h *SSHDHardenHandler) reloadSSHD(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "systemctl", "reload", "sshd")
	if err := cmd.Run(); err != nil {
		// Try ssh service name
		cmd = exec.CommandContext(ctx, "systemctl", "reload", "ssh")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to reload sshd/ssh service: %w", err)
		}
	}
	return nil
}
