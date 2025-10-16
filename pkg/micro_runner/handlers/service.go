package handlers

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/openfroyo/openfroyo/pkg/micro_runner/protocol"
)

// ServiceReloadHandler handles systemd service operations.
type ServiceReloadHandler struct{}

// Handle manages a systemd service.
func (h *ServiceReloadHandler) Handle(ctx context.Context, params *protocol.ServiceReloadParams, eventCh chan<- *protocol.EventMessage) (*protocol.ServiceReloadResult, error) {
	if params.Name == "" {
		return nil, fmt.Errorf("service name is required")
	}

	// Get current service status before action
	beforeStatus, beforeEnabled, _, err := h.getServiceStatus(ctx, params.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get service status: %w", err)
	}

	result := &protocol.ServiceReloadResult{}

	// Execute the requested action
	switch params.Action {
	case "reload":
		if err := h.reloadService(ctx, params.Name); err != nil {
			return nil, err
		}
		result.Action = "reloaded"

	case "restart":
		if err := h.restartService(ctx, params.Name); err != nil {
			return nil, err
		}
		result.Action = "restarted"

	case "start":
		if beforeStatus == "active" {
			result.Changed = false
			result.Action = "already_started"
		} else {
			if err := h.startService(ctx, params.Name); err != nil {
				return nil, err
			}
			result.Action = "started"
			result.Changed = true
		}

	case "stop":
		if beforeStatus == "inactive" {
			result.Changed = false
			result.Action = "already_stopped"
		} else {
			if err := h.stopService(ctx, params.Name); err != nil {
				return nil, err
			}
			result.Action = "stopped"
			result.Changed = true
		}

	case "enable":
		if beforeEnabled {
			result.Changed = false
			result.Action = "already_enabled"
		} else {
			if err := h.enableService(ctx, params.Name); err != nil {
				return nil, err
			}
			result.Action = "enabled"
			result.Changed = true
		}

	case "disable":
		if !beforeEnabled {
			result.Changed = false
			result.Action = "already_disabled"
		} else {
			if err := h.disableService(ctx, params.Name); err != nil {
				return nil, err
			}
			result.Action = "disabled"
			result.Changed = true
		}

	default:
		return nil, fmt.Errorf("invalid action: %s", params.Action)
	}

	// Get status after action
	afterStatus, afterEnabled, afterSubState, err := h.getServiceStatus(ctx, params.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get service status after action: %w", err)
	}

	result.Status = afterStatus
	result.Enabled = afterEnabled
	result.SubState = afterSubState

	// Determine if there was a change for reload/restart
	if params.Action == "reload" || params.Action == "restart" {
		result.Changed = true
	}

	return result, nil
}

func (h *ServiceReloadHandler) getServiceStatus(ctx context.Context, name string) (string, bool, string, error) {
	// Check if service is active
	statusCmd := exec.CommandContext(ctx, "systemctl", "is-active", name)
	statusOut, _ := statusCmd.Output()
	status := strings.TrimSpace(string(statusOut))

	// Check if service is enabled
	enabledCmd := exec.CommandContext(ctx, "systemctl", "is-enabled", name)
	enabledOut, _ := enabledCmd.Output()
	enabled := strings.TrimSpace(string(enabledOut)) == "enabled"

	// Get detailed status including substate
	showCmd := exec.CommandContext(ctx, "systemctl", "show", name, "--property=SubState", "--value")
	showOut, _ := showCmd.Output()
	subState := strings.TrimSpace(string(showOut))

	return status, enabled, subState, nil
}

func (h *ServiceReloadHandler) reloadService(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "systemctl", "reload", name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to reload service: %w", err)
	}
	return nil
}

func (h *ServiceReloadHandler) restartService(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "systemctl", "restart", name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to restart service: %w", err)
	}
	return nil
}

func (h *ServiceReloadHandler) startService(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "systemctl", "start", name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}
	return nil
}

func (h *ServiceReloadHandler) stopService(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "systemctl", "stop", name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}
	return nil
}

func (h *ServiceReloadHandler) enableService(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "systemctl", "enable", name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to enable service: %w", err)
	}
	return nil
}

func (h *ServiceReloadHandler) disableService(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "systemctl", "disable", name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to disable service: %w", err)
	}
	return nil
}
