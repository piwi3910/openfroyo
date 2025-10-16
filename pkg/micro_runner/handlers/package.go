package handlers

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/openfroyo/openfroyo/pkg/micro_runner/protocol"
)

// PkgEnsureHandler handles package management operations.
type PkgEnsureHandler struct{}

// Handle ensures a package is in the desired state.
func (h *PkgEnsureHandler) Handle(ctx context.Context, params *protocol.PkgEnsureParams, eventCh chan<- *protocol.EventMessage) (*protocol.PkgEnsureResult, error) {
	if params.Name == "" {
		return nil, fmt.Errorf("package name is required")
	}

	// Detect package manager if not specified
	manager := params.Manager
	if manager == "" {
		var err error
		manager, err = detectPackageManager()
		if err != nil {
			return nil, fmt.Errorf("failed to detect package manager: %w", err)
		}
	}

	result := &protocol.PkgEnsureResult{}

	// Check current package status
	installed, currentVersion, err := h.isPackageInstalled(ctx, manager, params.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check package status: %w", err)
	}

	result.PreviousVersion = currentVersion

	// Handle based on desired state
	switch params.State {
	case "present":
		if installed {
			result.Changed = false
			result.Action = "already_present"
			result.InstalledVersion = currentVersion
		} else {
			if err := h.installPackage(ctx, manager, params.Name, params.Version, params.Options); err != nil {
				return nil, fmt.Errorf("failed to install package: %w", err)
			}
			result.Changed = true
			result.Action = "installed"
			// Get new version
			_, newVersion, _ := h.isPackageInstalled(ctx, manager, params.Name)
			result.InstalledVersion = newVersion
		}

	case "absent":
		if !installed {
			result.Changed = false
			result.Action = "already_absent"
		} else {
			if err := h.removePackage(ctx, manager, params.Name, params.Options); err != nil {
				return nil, fmt.Errorf("failed to remove package: %w", err)
			}
			result.Changed = true
			result.Action = "removed"
		}

	case "latest":
		if !installed {
			if err := h.installPackage(ctx, manager, params.Name, "", params.Options); err != nil {
				return nil, fmt.Errorf("failed to install package: %w", err)
			}
			result.Changed = true
			result.Action = "installed"
		} else {
			if err := h.upgradePackage(ctx, manager, params.Name, params.Options); err != nil {
				return nil, fmt.Errorf("failed to upgrade package: %w", err)
			}
			result.Changed = true
			result.Action = "upgraded"
		}
		_, newVersion, _ := h.isPackageInstalled(ctx, manager, params.Name)
		result.InstalledVersion = newVersion

	default:
		return nil, fmt.Errorf("invalid state: %s", params.State)
	}

	return result, nil
}

func (h *PkgEnsureHandler) isPackageInstalled(ctx context.Context, manager, name string) (bool, string, error) {
	var cmd *exec.Cmd

	switch manager {
	case "apt":
		cmd = exec.CommandContext(ctx, "dpkg-query", "-W", "-f=${Version}", name)
	case "dnf", "yum":
		cmd = exec.CommandContext(ctx, "rpm", "-q", "--queryformat", "%{VERSION}-%{RELEASE}", name)
	case "zypper":
		cmd = exec.CommandContext(ctx, "rpm", "-q", "--queryformat", "%{VERSION}-%{RELEASE}", name)
	default:
		return false, "", fmt.Errorf("unsupported package manager: %s", manager)
	}

	output, err := cmd.Output()
	if err != nil {
		return false, "", nil // Package not installed
	}

	version := strings.TrimSpace(string(output))
	return true, version, nil
}

func (h *PkgEnsureHandler) installPackage(ctx context.Context, manager, name, version string, options []string) error {
	pkgSpec := name
	if version != "" {
		switch manager {
		case "apt":
			pkgSpec = fmt.Sprintf("%s=%s", name, version)
		case "dnf", "yum":
			pkgSpec = fmt.Sprintf("%s-%s", name, version)
		}
	}

	var args []string
	switch manager {
	case "apt":
		args = append([]string{"install", "-y"}, options...)
		args = append(args, pkgSpec)
	case "dnf", "yum":
		args = append([]string{"install", "-y"}, options...)
		args = append(args, pkgSpec)
	case "zypper":
		args = append([]string{"install", "-y"}, options...)
		args = append(args, pkgSpec)
	default:
		return fmt.Errorf("unsupported package manager: %s", manager)
	}

	cmd := exec.CommandContext(ctx, manager, args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}

func (h *PkgEnsureHandler) removePackage(ctx context.Context, manager, name string, options []string) error {
	var args []string
	switch manager {
	case "apt":
		args = append([]string{"remove", "-y"}, options...)
		args = append(args, name)
	case "dnf", "yum":
		args = append([]string{"remove", "-y"}, options...)
		args = append(args, name)
	case "zypper":
		args = append([]string{"remove", "-y"}, options...)
		args = append(args, name)
	default:
		return fmt.Errorf("unsupported package manager: %s", manager)
	}

	cmd := exec.CommandContext(ctx, manager, args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}

func (h *PkgEnsureHandler) upgradePackage(ctx context.Context, manager, name string, options []string) error {
	var args []string
	switch manager {
	case "apt":
		args = append([]string{"upgrade", "-y"}, options...)
		args = append(args, name)
	case "dnf", "yum":
		args = append([]string{"upgrade", "-y"}, options...)
		args = append(args, name)
	case "zypper":
		args = append([]string{"update", "-y"}, options...)
		args = append(args, name)
	default:
		return fmt.Errorf("unsupported package manager: %s", manager)
	}

	cmd := exec.CommandContext(ctx, manager, args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}

func detectPackageManager() (string, error) {
	managers := []string{"apt", "dnf", "yum", "zypper"}
	for _, mgr := range managers {
		if _, err := exec.LookPath(mgr); err == nil {
			return mgr, nil
		}
	}
	return "", fmt.Errorf("no supported package manager found")
}
