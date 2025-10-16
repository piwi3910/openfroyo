package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/openfroyo/openfroyo/pkg/engine"
)

// TestProviderMetadata tests that provider metadata is correctly returned
func TestProviderMetadata(t *testing.T) {
	p := &Provider{}
	metadata := p.Metadata()

	if metadata.Name != "linux.pkg" {
		t.Errorf("Expected name 'linux.pkg', got '%s'", metadata.Name)
	}

	if metadata.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", metadata.Version)
	}

	if len(metadata.RequiredCapabilities) == 0 {
		t.Error("Expected required capabilities, got none")
	}

	hasExecCapability := false
	for _, cap := range metadata.RequiredCapabilities {
		if cap == "exec:micro-runner" {
			hasExecCapability = true
			break
		}
	}

	if !hasExecCapability {
		t.Error("Expected exec:micro-runner capability")
	}
}

// TestProviderSchema tests that provider schema is valid
func TestProviderSchema(t *testing.T) {
	p := &Provider{}
	schema, err := p.Schema()

	if err != nil {
		t.Fatalf("Schema() returned error: %v", err)
	}

	if schema == nil {
		t.Fatal("Schema() returned nil")
	}

	if schema.Version != "1.0.0" {
		t.Errorf("Expected schema version '1.0.0', got '%s'", schema.Version)
	}

	if len(schema.ResourceTypes) == 0 {
		t.Fatal("Expected resource types, got none")
	}

	pkgSchema, exists := schema.ResourceTypes["package"]
	if !exists {
		t.Fatal("Expected 'package' resource type")
	}

	if pkgSchema.Name != "package" {
		t.Errorf("Expected resource type name 'package', got '%s'", pkgSchema.Name)
	}

	// Validate JSON schemas can be unmarshaled
	var configSchema map[string]interface{}
	if err := json.Unmarshal(pkgSchema.ConfigSchema, &configSchema); err != nil {
		t.Errorf("ConfigSchema is not valid JSON: %v", err)
	}

	var stateSchema map[string]interface{}
	if err := json.Unmarshal(pkgSchema.StateSchema, &stateSchema); err != nil {
		t.Errorf("StateSchema is not valid JSON: %v", err)
	}
}

// TestProviderInit tests provider initialization
func TestProviderInit(t *testing.T) {
	tests := []struct {
		name        string
		config      engine.ProviderConfig
		expectError bool
	}{
		{
			name: "valid config with capabilities",
			config: engine.ProviderConfig{
				Name:         "linux.pkg",
				Version:      "1.0.0",
				Capabilities: []string{"exec:micro-runner"},
			},
			expectError: false,
		},
		{
			name: "missing required capability",
			config: engine.ProviderConfig{
				Name:         "linux.pkg",
				Version:      "1.0.0",
				Capabilities: []string{},
			},
			expectError: true,
		},
		{
			name: "valid config with custom settings",
			config: engine.ProviderConfig{
				Name:         "linux.pkg",
				Version:      "1.0.0",
				Capabilities: []string{"exec:micro-runner"},
				Config: json.RawMessage(`{
					"default_manager": "apt",
					"update_cache": true,
					"cache_validity_minutes": 120
				}`),
			},
			expectError: false,
		},
		{
			name: "invalid JSON config",
			config: engine.ProviderConfig{
				Name:         "linux.pkg",
				Version:      "1.0.0",
				Capabilities: []string{"exec:micro-runner"},
				Config:       json.RawMessage(`{invalid json`),
			},
			expectError: true,
		},
		{
			name: "invalid package manager in config",
			config: engine.ProviderConfig{
				Name:         "linux.pkg",
				Version:      "1.0.0",
				Capabilities: []string{"exec:micro-runner"},
				Config: json.RawMessage(`{
					"default_manager": "homebrew"
				}`),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{}
			err := p.Init(context.Background(), tt.config)

			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if !tt.expectError && !p.initialized {
				t.Error("Provider should be initialized")
			}
		})
	}
}

// TestValidatePackageConfig tests package configuration validation
func TestValidatePackageConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      PackageConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config - present state",
			config: PackageConfig{
				Package: "nginx",
				State:   "present",
			},
			expectError: false,
		},
		{
			name: "valid config - absent state",
			config: PackageConfig{
				Package: "nginx",
				State:   "absent",
			},
			expectError: false,
		},
		{
			name: "valid config - latest state",
			config: PackageConfig{
				Package: "nginx",
				State:   "latest",
			},
			expectError: false,
		},
		{
			name: "valid config - with version",
			config: PackageConfig{
				Package: "nginx",
				State:   "present",
				Version: "1.18.0",
			},
			expectError: false,
		},
		{
			name: "valid config - with manager",
			config: PackageConfig{
				Package: "nginx",
				State:   "present",
				Manager: "apt",
			},
			expectError: false,
		},
		{
			name: "missing package name",
			config: PackageConfig{
				State: "present",
			},
			expectError: true,
			errorMsg:    "package name is required",
		},
		{
			name: "invalid state",
			config: PackageConfig{
				Package: "nginx",
				State:   "installed",
			},
			expectError: true,
			errorMsg:    "invalid state",
		},
		{
			name: "version with absent state",
			config: PackageConfig{
				Package: "nginx",
				State:   "absent",
				Version: "1.18.0",
			},
			expectError: true,
			errorMsg:    "version cannot be specified when state is absent",
		},
		{
			name: "version with latest state",
			config: PackageConfig{
				Package: "nginx",
				State:   "latest",
				Version: "1.18.0",
			},
			expectError: true,
			errorMsg:    "version cannot be specified when state is latest",
		},
		{
			name: "invalid package manager",
			config: PackageConfig{
				Package: "nginx",
				State:   "present",
				Manager: "homebrew",
			},
			expectError: true,
			errorMsg:    "invalid package manager",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePackageConfig(&tt.config)

			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if tt.expectError && err != nil && tt.errorMsg != "" {
				if err.Error() != tt.errorMsg && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
			}
		})
	}
}

// TestProviderValidate tests the Validate method
func TestProviderValidate(t *testing.T) {
	p := &Provider{}
	ctx := context.Background()

	tests := []struct {
		name        string
		config      string
		expectError bool
	}{
		{
			name:        "valid config",
			config:      `{"package": "nginx", "state": "present"}`,
			expectError: false,
		},
		{
			name:        "invalid JSON",
			config:      `{invalid`,
			expectError: true,
		},
		{
			name:        "missing required field",
			config:      `{"state": "present"}`,
			expectError: true,
		},
		{
			name:        "invalid state",
			config:      `{"package": "nginx", "state": "running"}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.Validate(ctx, json.RawMessage(tt.config))

			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

// TestIsValidPackageManager tests package manager validation
func TestIsValidPackageManager(t *testing.T) {
	tests := []struct {
		manager string
		valid   bool
	}{
		{"apt", true},
		{"dnf", true},
		{"yum", true},
		{"zypper", true},
		{"homebrew", false},
		{"pacman", false},
		{"", false},
		{"APT", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.manager, func(t *testing.T) {
			result := isValidPackageManager(tt.manager)
			if result != tt.valid {
				t.Errorf("isValidPackageManager(%s) = %v, want %v", tt.manager, result, tt.valid)
			}
		})
	}
}

// TestPlanOperation tests the Plan method logic
func TestPlanOperation(t *testing.T) {
	p := &Provider{
		initialized:    true,
		packageManager: "apt",
		config: &ProviderConfig{
			UpdateCache:          true,
			CacheValidityMinutes: 60,
		},
		capabilities: map[string]bool{
			"exec:micro-runner": true,
		},
	}

	ctx := context.Background()

	tests := []struct {
		name              string
		desired           PackageConfig
		actual            *PackageState
		expectedOperation engine.OperationType
		expectedChanges   int
	}{
		{
			name: "install package - not installed",
			desired: PackageConfig{
				Package: "nginx",
				State:   "present",
			},
			actual:            nil,
			expectedOperation: engine.OperationCreate,
			expectedChanges:   1, // installed flag
		},
		{
			name: "package already installed",
			desired: PackageConfig{
				Package: "nginx",
				State:   "present",
			},
			actual: &PackageState{
				Package:   "nginx",
				Installed: true,
				Version:   "1.18.0",
			},
			expectedOperation: engine.OperationNoop,
			expectedChanges:   0,
		},
		{
			name: "remove installed package",
			desired: PackageConfig{
				Package: "nginx",
				State:   "absent",
			},
			actual: &PackageState{
				Package:   "nginx",
				Installed: true,
				Version:   "1.18.0",
			},
			expectedOperation: engine.OperationDelete,
			expectedChanges:   2, // installed flag + version
		},
		{
			name: "upgrade to latest",
			desired: PackageConfig{
				Package: "nginx",
				State:   "latest",
			},
			actual: &PackageState{
				Package:          "nginx",
				Installed:        true,
				Version:          "1.18.0",
				AvailableVersion: "1.20.0",
			},
			expectedOperation: engine.OperationUpdate,
			expectedChanges:   1, // version change
		},
		{
			name: "change version",
			desired: PackageConfig{
				Package: "nginx",
				State:   "present",
				Version: "1.20.0",
			},
			actual: &PackageState{
				Package:   "nginx",
				Installed: true,
				Version:   "1.18.0",
			},
			expectedOperation: engine.OperationUpdate,
			expectedChanges:   1, // version change
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desiredJSON, _ := json.Marshal(tt.desired)
			var actualJSON json.RawMessage
			if tt.actual != nil {
				actualJSON, _ = json.Marshal(tt.actual)
			}

			req := engine.PlanRequest{
				ResourceID:   "test-resource",
				DesiredState: desiredJSON,
				ActualState:  actualJSON,
				Operation:    engine.OperationCreate,
			}

			resp, err := p.Plan(ctx, req)
			if err != nil {
				t.Fatalf("Plan() error: %v", err)
			}

			if resp.Operation != tt.expectedOperation {
				t.Errorf("Expected operation %s, got %s", tt.expectedOperation, resp.Operation)
			}

			if len(resp.Changes) != tt.expectedChanges {
				t.Errorf("Expected %d changes, got %d", tt.expectedChanges, len(resp.Changes))
			}
		})
	}
}

// TestResolvePackageManager tests package manager resolution logic
func TestResolvePackageManager(t *testing.T) {
	tests := []struct {
		name              string
		requested         string
		providerDefault   string
		expectedManager   string
		expectError       bool
	}{
		{
			name:            "use requested manager",
			requested:       "dnf",
			providerDefault: "apt",
			expectedManager: "dnf",
			expectError:     false,
		},
		{
			name:            "use provider default",
			requested:       "",
			providerDefault: "apt",
			expectedManager: "apt",
			expectError:     false,
		},
		{
			name:            "invalid requested manager",
			requested:       "homebrew",
			providerDefault: "apt",
			expectError:     true,
		},
		{
			name:            "no manager specified",
			requested:       "",
			providerDefault: "",
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{
				packageManager: tt.providerDefault,
			}

			manager, err := p.resolvePackageManager(tt.requested)

			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if !tt.expectError && manager != tt.expectedManager {
				t.Errorf("Expected manager '%s', got '%s'", tt.expectedManager, manager)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
