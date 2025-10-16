// Package main implements the linux.pkg provider for OpenFroyo.
// This provider manages Linux packages across multiple package managers
// (apt, dnf, yum, zypper) and compiles to WASM for secure, portable execution.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openfroyo/openfroyo/pkg/engine"
)

// Provider implements the engine.Provider interface for Linux package management.
type Provider struct {
	config        *ProviderConfig
	capabilities  map[string]bool
	initialized   bool
	packageManager string
}

// ProviderConfig holds provider-specific configuration.
type ProviderConfig struct {
	// DefaultManager specifies the default package manager to use.
	// If empty, auto-detection will be used.
	DefaultManager string `json:"default_manager,omitempty"`

	// UpdateCache determines if package cache should be updated before operations.
	UpdateCache bool `json:"update_cache,omitempty"`

	// CacheValidityMinutes specifies how long the package cache is considered fresh.
	CacheValidityMinutes int `json:"cache_validity_minutes,omitempty"`
}

// PackageConfig represents the desired configuration for a package resource.
type PackageConfig struct {
	// Package is the name of the package to manage.
	Package string `json:"package"`

	// State is the desired state (present, absent, latest).
	State string `json:"state"`

	// Version is the specific version to install (optional).
	// If empty and state is "present", the latest available version is installed.
	Version string `json:"version,omitempty"`

	// Repository is a specific repository to use (optional).
	Repository string `json:"repository,omitempty"`

	// Manager is the package manager to use (optional, auto-detected if empty).
	Manager string `json:"manager,omitempty"`

	// Options are additional package manager specific options.
	Options []string `json:"options,omitempty"`
}

// PackageState represents the current state of a package.
type PackageState struct {
	// Package is the name of the package.
	Package string `json:"package"`

	// Installed indicates whether the package is installed.
	Installed bool `json:"installed"`

	// Version is the currently installed version (empty if not installed).
	Version string `json:"version,omitempty"`

	// Manager is the package manager managing this package.
	Manager string `json:"manager"`

	// AvailableVersion is the latest available version.
	AvailableVersion string `json:"available_version,omitempty"`

	// Repository is the repository the package is from.
	Repository string `json:"repository,omitempty"`
}

// Init initializes the provider with configuration.
func (p *Provider) Init(ctx context.Context, config engine.ProviderConfig) error {
	// Parse provider-specific configuration
	p.config = &ProviderConfig{
		UpdateCache: true,
		CacheValidityMinutes: 60,
	}

	if len(config.Config) > 0 {
		if err := json.Unmarshal(config.Config, p.config); err != nil {
			return fmt.Errorf("failed to parse provider config: %w", err)
		}
	}

	// Store capabilities
	p.capabilities = make(map[string]bool)
	for _, cap := range config.Capabilities {
		p.capabilities[cap] = true
	}

	// Verify we have the required exec:micro-runner capability
	if !p.capabilities["exec:micro-runner"] {
		return fmt.Errorf("provider requires exec:micro-runner capability")
	}

	// Detect or validate package manager
	if p.config.DefaultManager != "" {
		if !isValidPackageManager(p.config.DefaultManager) {
			return fmt.Errorf("invalid package manager: %s", p.config.DefaultManager)
		}
		p.packageManager = p.config.DefaultManager
	}

	p.initialized = true
	return nil
}

// Read retrieves the current state of a package resource.
func (p *Provider) Read(ctx context.Context, req engine.ReadRequest) (*engine.ReadResponse, error) {
	if !p.initialized {
		return nil, fmt.Errorf("provider not initialized")
	}

	// Parse the resource configuration
	var config PackageConfig
	if err := json.Unmarshal(req.Config, &config); err != nil {
		return nil, fmt.Errorf("failed to parse resource config: %w", err)
	}

	// Validate configuration
	if err := validatePackageConfig(&config); err != nil {
		return nil, err
	}

	// Determine package manager to use
	manager, err := p.resolvePackageManager(config.Manager)
	if err != nil {
		return nil, err
	}

	// Get current package state via micro-runner
	state, err := p.getPackageState(ctx, manager, config.Package)
	if err != nil {
		return nil, fmt.Errorf("failed to read package state: %w", err)
	}

	// Marshal state
	stateJSON, err := json.Marshal(state)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal state: %w", err)
	}

	return &engine.ReadResponse{
		State:  stateJSON,
		Exists: state.Installed,
		Metadata: map[string]interface{}{
			"manager": manager,
			"package": config.Package,
		},
	}, nil
}

// Plan computes the operations needed to reach desired state.
func (p *Provider) Plan(ctx context.Context, req engine.PlanRequest) (*engine.PlanResponse, error) {
	if !p.initialized {
		return nil, fmt.Errorf("provider not initialized")
	}

	// Parse desired and actual states
	var desired PackageConfig
	if err := json.Unmarshal(req.DesiredState, &desired); err != nil {
		return nil, fmt.Errorf("failed to parse desired state: %w", err)
	}

	var actual PackageState
	actualExists := len(req.ActualState) > 0
	if actualExists {
		if err := json.Unmarshal(req.ActualState, &actual); err != nil {
			return nil, fmt.Errorf("failed to parse actual state: %w", err)
		}
	}

	// Determine the operation and changes
	operation := engine.OperationNoop
	var changes []engine.Change
	var warnings []string

	switch desired.State {
	case "present":
		if !actualExists || !actual.Installed {
			// Package needs to be installed
			operation = engine.OperationCreate
			changes = append(changes, engine.Change{
				Path:   ".installed",
				Before: false,
				After:  true,
				Action: engine.ChangeActionAdd,
			})
			if desired.Version != "" {
				changes = append(changes, engine.Change{
					Path:   ".version",
					Before: nil,
					After:  desired.Version,
					Action: engine.ChangeActionAdd,
				})
			}
		} else if desired.Version != "" && actual.Version != desired.Version {
			// Package version needs to be changed
			operation = engine.OperationUpdate
			changes = append(changes, engine.Change{
				Path:   ".version",
				Before: actual.Version,
				After:  desired.Version,
				Action: engine.ChangeActionModify,
			})
		}

	case "absent":
		if actualExists && actual.Installed {
			// Package needs to be removed
			operation = engine.OperationDelete
			changes = append(changes, engine.Change{
				Path:   ".installed",
				Before: true,
				After:  false,
				Action: engine.ChangeActionRemove,
			})
			if actual.Version != "" {
				changes = append(changes, engine.Change{
					Path:   ".version",
					Before: actual.Version,
					After:  nil,
					Action: engine.ChangeActionRemove,
				})
			}
		}

	case "latest":
		if !actualExists || !actual.Installed {
			// Package needs to be installed
			operation = engine.OperationCreate
			changes = append(changes, engine.Change{
				Path:   ".installed",
				Before: false,
				After:  true,
				Action: engine.ChangeActionAdd,
			})
			changes = append(changes, engine.Change{
				Path:   ".version",
				Before: nil,
				After:  "latest",
				Action: engine.ChangeActionAdd,
			})
		} else if actual.AvailableVersion != "" && actual.Version != actual.AvailableVersion {
			// Package needs to be upgraded
			operation = engine.OperationUpdate
			changes = append(changes, engine.Change{
				Path:   ".version",
				Before: actual.Version,
				After:  actual.AvailableVersion,
				Action: engine.ChangeActionModify,
			})
		}
	}

	return &engine.PlanResponse{
		Operation:        operation,
		Changes:          changes,
		RequiresRecreate: false,
		Warnings:         warnings,
		Metadata: map[string]interface{}{
			"package": desired.Package,
			"state":   desired.State,
		},
	}, nil
}

// Apply executes the planned operations to achieve desired state.
func (p *Provider) Apply(ctx context.Context, req engine.ApplyRequest) (*engine.ApplyResponse, error) {
	if !p.initialized {
		return nil, fmt.Errorf("provider not initialized")
	}

	// Parse desired state
	var desired PackageConfig
	if err := json.Unmarshal(req.DesiredState, &desired); err != nil {
		return nil, fmt.Errorf("failed to parse desired state: %w", err)
	}

	// Determine package manager
	manager, err := p.resolvePackageManager(desired.Manager)
	if err != nil {
		return nil, err
	}

	// Execute the operation via micro-runner
	var events []engine.ProviderEvent

	switch req.Operation {
	case engine.OperationCreate:
		events, err = p.installPackage(ctx, manager, &desired)
		if err != nil {
			return nil, fmt.Errorf("failed to install package: %w", err)
		}

	case engine.OperationUpdate:
		events, err = p.updatePackage(ctx, manager, &desired)
		if err != nil {
			return nil, fmt.Errorf("failed to update package: %w", err)
		}

	case engine.OperationDelete:
		events, err = p.removePackage(ctx, manager, &desired)
		if err != nil {
			return nil, fmt.Errorf("failed to remove package: %w", err)
		}

	case engine.OperationNoop:
		// No operation needed

	default:
		return nil, fmt.Errorf("unsupported operation: %s", req.Operation)
	}

	// Read the new state
	newState, err := p.getPackageState(ctx, manager, desired.Package)
	if err != nil {
		return nil, fmt.Errorf("failed to read new state: %w", err)
	}

	newStateJSON, err := json.Marshal(newState)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal new state: %w", err)
	}

	return &engine.ApplyResponse{
		NewState: newStateJSON,
		Events:   events,
		Metadata: map[string]interface{}{
			"manager": manager,
			"package": desired.Package,
		},
	}, nil
}

// Destroy removes the package completely.
func (p *Provider) Destroy(ctx context.Context, req engine.DestroyRequest) (*engine.DestroyResponse, error) {
	if !p.initialized {
		return nil, fmt.Errorf("provider not initialized")
	}

	// Parse the current state
	var state PackageState
	if err := json.Unmarshal(req.State, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state: %w", err)
	}

	// If package is not installed, nothing to do
	if !state.Installed {
		return &engine.DestroyResponse{
			Success: true,
		}, nil
	}

	// Create a config for removal
	config := PackageConfig{
		Package: state.Package,
		State:   "absent",
		Manager: state.Manager,
	}

	// Remove the package
	events, err := p.removePackage(ctx, state.Manager, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to remove package: %w", err)
	}

	return &engine.DestroyResponse{
		Success: true,
		Events:  events,
	}, nil
}

// Validate validates a resource configuration against the provider's schema.
func (p *Provider) Validate(ctx context.Context, config json.RawMessage) error {
	var pkgConfig PackageConfig
	if err := json.Unmarshal(config, &pkgConfig); err != nil {
		return fmt.Errorf("invalid configuration format: %w", err)
	}

	return validatePackageConfig(&pkgConfig)
}

// Schema returns the JSON schema for this provider's resources.
func (p *Provider) Schema() (*engine.ProviderSchema, error) {
	// Load the JSON schema from embedded file
	// In a real implementation, this would be embedded using go:embed
	schema := &engine.ProviderSchema{
		Version: "1.0.0",
		ResourceTypes: map[string]*engine.ResourceTypeSchema{
			"package": {
				Name:         "package",
				Description:  "Manages Linux packages across multiple package managers",
				ConfigSchema: json.RawMessage(packageConfigSchema),
				StateSchema:  json.RawMessage(packageStateSchema),
				Capabilities: []string{"exec:micro-runner"},
			},
		},
	}

	return schema, nil
}

// Metadata returns information about this provider.
func (p *Provider) Metadata() engine.ProviderMetadata {
	return engine.ProviderMetadata{
		Name:        "linux.pkg",
		Version:     "1.0.0",
		Description: "Linux package management provider supporting apt, dnf, yum, and zypper",
		Author:      "OpenFroyo",
		License:     "Apache-2.0",
		Repository:  "https://github.com/openfroyo/openfroyo",
		Homepage:    "https://openfroyo.io",
		RequiredCapabilities: []string{
			"exec:micro-runner",
		},
	}
}

// Helper methods

// resolvePackageManager determines which package manager to use.
func (p *Provider) resolvePackageManager(requested string) (string, error) {
	if requested != "" {
		if !isValidPackageManager(requested) {
			return "", fmt.Errorf("invalid package manager: %s", requested)
		}
		return requested, nil
	}

	if p.packageManager != "" {
		return p.packageManager, nil
	}

	// Auto-detection would be done via micro-runner
	// For now, return an error requiring explicit configuration
	return "", fmt.Errorf("package manager must be specified or configured")
}

// getPackageState retrieves the current state of a package via micro-runner.
func (p *Provider) getPackageState(ctx context.Context, manager, packageName string) (*PackageState, error) {
	// This would communicate with the micro-runner via the exec capability
	// For WASM compilation, this is a placeholder that represents the interface
	// The actual implementation would use WASI host functions to call micro-runner

	// The micro-runner would execute commands like:
	// - apt: dpkg-query -W -f='${Version}' packageName
	// - dnf/yum: rpm -q --queryformat '%{VERSION}-%{RELEASE}' packageName
	// - zypper: rpm -q --queryformat '%{VERSION}-%{RELEASE}' packageName

	state := &PackageState{
		Package: packageName,
		Manager: manager,
	}

	// Placeholder: In real implementation, this would call micro-runner
	// via WASI host function exposed by the WASM runtime

	return state, nil
}

// installPackage installs a package via micro-runner.
func (p *Provider) installPackage(ctx context.Context, manager string, config *PackageConfig) ([]engine.ProviderEvent, error) {
	// This would send a pkg.ensure command to the micro-runner with:
	// - name: config.Package
	// - version: config.Version
	// - state: "present"
	// - manager: manager
	// - options: config.Options

	var events []engine.ProviderEvent

	// Placeholder: In real implementation, this would communicate with micro-runner

	return events, nil
}

// updatePackage updates a package to a different version.
func (p *Provider) updatePackage(ctx context.Context, manager string, config *PackageConfig) ([]engine.ProviderEvent, error) {
	// For state "latest", this would use the upgrade/update command
	// For a specific version, this might require remove + install

	var events []engine.ProviderEvent

	// Placeholder: In real implementation, this would communicate with micro-runner

	return events, nil
}

// removePackage removes a package via micro-runner.
func (p *Provider) removePackage(ctx context.Context, manager string, config *PackageConfig) ([]engine.ProviderEvent, error) {
	// This would send a pkg.ensure command to the micro-runner with:
	// - name: config.Package
	// - state: "absent"
	// - manager: manager

	var events []engine.ProviderEvent

	// Placeholder: In real implementation, this would communicate with micro-runner

	return events, nil
}

// validatePackageConfig validates a package configuration.
func validatePackageConfig(config *PackageConfig) error {
	if config.Package == "" {
		return fmt.Errorf("package name is required")
	}

	if config.State == "" {
		config.State = "present"
	}

	validStates := map[string]bool{
		"present": true,
		"absent":  true,
		"latest":  true,
	}

	if !validStates[config.State] {
		return fmt.Errorf("invalid state: %s (must be present, absent, or latest)", config.State)
	}

	if config.Manager != "" && !isValidPackageManager(config.Manager) {
		return fmt.Errorf("invalid package manager: %s", config.Manager)
	}

	if config.State == "absent" && config.Version != "" {
		return fmt.Errorf("version cannot be specified when state is absent")
	}

	if config.State == "latest" && config.Version != "" {
		return fmt.Errorf("version cannot be specified when state is latest")
	}

	return nil
}

// isValidPackageManager checks if a package manager is supported.
func isValidPackageManager(manager string) bool {
	validManagers := map[string]bool{
		"apt":    true,
		"dnf":    true,
		"yum":    true,
		"zypper": true,
	}
	return validManagers[manager]
}

// JSON Schema definitions (would normally be in separate files)
const packageConfigSchema = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["package"],
  "properties": {
    "package": {
      "type": "string",
      "description": "Name of the package to manage",
      "minLength": 1
    },
    "state": {
      "type": "string",
      "description": "Desired state of the package",
      "enum": ["present", "absent", "latest"],
      "default": "present"
    },
    "version": {
      "type": "string",
      "description": "Specific version to install (optional)"
    },
    "repository": {
      "type": "string",
      "description": "Specific repository to use (optional)"
    },
    "manager": {
      "type": "string",
      "description": "Package manager to use (auto-detected if not specified)",
      "enum": ["apt", "dnf", "yum", "zypper"]
    },
    "options": {
      "type": "array",
      "description": "Additional package manager options",
      "items": {
        "type": "string"
      }
    }
  }
}`

const packageStateSchema = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "package": {
      "type": "string",
      "description": "Name of the package"
    },
    "installed": {
      "type": "boolean",
      "description": "Whether the package is installed"
    },
    "version": {
      "type": "string",
      "description": "Currently installed version"
    },
    "manager": {
      "type": "string",
      "description": "Package manager managing this package",
      "enum": ["apt", "dnf", "yum", "zypper"]
    },
    "available_version": {
      "type": "string",
      "description": "Latest available version"
    },
    "repository": {
      "type": "string",
      "description": "Repository the package is from"
    }
  }
}`

// Main function required for WASM module
// In a WASM environment, this would export the provider interface
func main() {
	// WASM module initialization
	// The WASM host would call the exported functions
}
