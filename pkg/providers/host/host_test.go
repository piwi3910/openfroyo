package host

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/openfroyo/openfroyo/pkg/engine"
)

// TestManifestLoader tests the manifest loader functionality.
func TestManifestLoader(t *testing.T) {
	t.Run("LoadFromBytes", func(t *testing.T) {
		manifestYAML := `
metadata:
  name: test-provider
  version: 1.0.0
  author: Test Author
  license: MIT
  description: Test provider
  required_capabilities:
    - net:outbound
    - fs:temp

schema:
  version: "1.0"
  resource_types:
    test_resource:
      name: test_resource
      description: A test resource
      config_schema: '{"type": "object", "properties": {"name": {"type": "string"}}}'
      state_schema: '{"type": "object", "properties": {"status": {"type": "string"}}}'
      capabilities:
        - net:outbound

entrypoint: test.wasm
checksum: ""
`

		loader := NewManifestLoader("/tmp")
		wasmModule := []byte("fake wasm module")

		manifest, err := loader.LoadFromBytes([]byte(manifestYAML), wasmModule)
		if err != nil {
			t.Fatalf("Failed to load manifest: %v", err)
		}

		if manifest.Raw.Metadata.Name != "test-provider" {
			t.Errorf("Expected name 'test-provider', got '%s'", manifest.Raw.Metadata.Name)
		}

		if manifest.Raw.Metadata.Version != "1.0.0" {
			t.Errorf("Expected version '1.0.0', got '%s'", manifest.Raw.Metadata.Version)
		}

		caps := manifest.GetCapabilities()
		if len(caps) == 0 {
			t.Error("Expected capabilities, got none")
		}

		resourceTypes := manifest.GetResourceTypes()
		if len(resourceTypes) != 1 || resourceTypes[0] != "test_resource" {
			t.Errorf("Expected 1 resource type 'test_resource', got %v", resourceTypes)
		}
	})

	t.Run("ValidateManifest", func(t *testing.T) {
		tests := []struct {
			name        string
			manifest    *engine.ProviderManifest
			expectError bool
		}{
			{
				name: "Valid manifest",
				manifest: &engine.ProviderManifest{
					Metadata: engine.ProviderMetadata{
						Name:    "test",
						Version: "1.0.0",
						Author:  "Test",
						License: "MIT",
					},
					Schema: &engine.ProviderSchema{
						Version: "1.0",
						ResourceTypes: map[string]*engine.ResourceTypeSchema{
							"test": {
								Name:         "test",
								Description:  "Test resource",
								ConfigSchema: json.RawMessage(`{}`),
								StateSchema:  json.RawMessage(`{}`),
							},
						},
					},
					Entrypoint: "test.wasm",
				},
				expectError: false,
			},
			{
				name: "Missing name",
				manifest: &engine.ProviderManifest{
					Metadata: engine.ProviderMetadata{
						Version: "1.0.0",
						Author:  "Test",
						License: "MIT",
					},
					Schema: &engine.ProviderSchema{
						Version: "1.0",
						ResourceTypes: map[string]*engine.ResourceTypeSchema{
							"test": {
								Name:         "test",
								Description:  "Test",
								ConfigSchema: json.RawMessage(`{}`),
								StateSchema:  json.RawMessage(`{}`),
							},
						},
					},
					Entrypoint: "test.wasm",
				},
				expectError: true,
			},
			{
				name: "Missing entrypoint",
				manifest: &engine.ProviderManifest{
					Metadata: engine.ProviderMetadata{
						Name:    "test",
						Version: "1.0.0",
						Author:  "Test",
						License: "MIT",
					},
					Schema: &engine.ProviderSchema{
						Version: "1.0",
						ResourceTypes: map[string]*engine.ResourceTypeSchema{
							"test": {
								Name:         "test",
								Description:  "Test",
								ConfigSchema: json.RawMessage(`{}`),
								StateSchema:  json.RawMessage(`{}`),
							},
						},
					},
				},
				expectError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				loader := NewManifestLoader("/tmp")
				err := loader.validateManifest(tt.manifest)

				if tt.expectError && err == nil {
					t.Error("Expected error, got none")
				}
				if !tt.expectError && err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			})
		}
	})
}

// TestCapabilityEnforcer tests the capability enforcement.
func TestCapabilityEnforcer(t *testing.T) {
	tempDir := t.TempDir()

	enforcer := NewCapabilityEnforcer(
		[]string{string(engine.CapabilityFSTemp), string(engine.CapabilityNetOutbound)},
		tempDir,
		"",
	)

	t.Run("HasCapability", func(t *testing.T) {
		if !enforcer.HasCapability(engine.CapabilityFSTemp) {
			t.Error("Expected fs:temp capability to be granted")
		}

		if enforcer.HasCapability(engine.CapabilitySecretsRead) {
			t.Error("Expected secrets:read capability to NOT be granted")
		}
	})

	t.Run("ValidateCapabilities", func(t *testing.T) {
		// Valid capabilities
		err := enforcer.ValidateCapabilities([]string{
			string(engine.CapabilityFSTemp),
			string(engine.CapabilityNetOutbound),
		})
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Invalid capability
		err = enforcer.ValidateCapabilities([]string{
			string(engine.CapabilitySecretsRead),
		})
		if err == nil {
			t.Error("Expected error for missing capability")
		}
	})

	t.Run("TempFileOperations", func(t *testing.T) {
		// Write temp file
		testData := []byte("test data")
		err := enforcer.WriteTempFile("test.txt", testData)
		if err != nil {
			t.Fatalf("Failed to write temp file: %v", err)
		}

		// Read temp file
		data, err := enforcer.ReadTempFile("test.txt")
		if err != nil {
			t.Fatalf("Failed to read temp file: %v", err)
		}

		if string(data) != string(testData) {
			t.Errorf("Expected data '%s', got '%s'", testData, data)
		}

		// List temp files
		files, err := enforcer.ListTempFiles()
		if err != nil {
			t.Fatalf("Failed to list temp files: %v", err)
		}

		if len(files) != 1 || files[0] != "test.txt" {
			t.Errorf("Expected 1 file 'test.txt', got %v", files)
		}

		// Delete temp file
		err = enforcer.DeleteTempFile("test.txt")
		if err != nil {
			t.Fatalf("Failed to delete temp file: %v", err)
		}

		// Verify deleted
		files, err = enforcer.ListTempFiles()
		if err != nil {
			t.Fatalf("Failed to list temp files: %v", err)
		}

		if len(files) != 0 {
			t.Errorf("Expected 0 files, got %v", files)
		}
	})

	t.Run("PathTraversalPrevention", func(t *testing.T) {
		// Attempt path traversal
		err := enforcer.WriteTempFile("../etc/passwd", []byte("malicious"))
		if err == nil {
			t.Error("Expected error for path traversal attempt")
		}
	})

	t.Run("HTTPRequest", func(t *testing.T) {
		ctx := context.Background()

		// Valid request (will fail to connect but should pass capability check)
		_, err := enforcer.HTTPRequest(ctx, "GET", "http://localhost:9999", nil)
		// We expect a connection error, not a capability error
		if err != nil && err.Error() == "capability net:outbound not granted" {
			t.Error("HTTP request capability check failed incorrectly")
		}
	})

	t.Run("DeniedCapability", func(t *testing.T) {
		// Try to decrypt secret without permission
		_, err := enforcer.DecryptSecret("encrypted")
		if err == nil {
			t.Error("Expected error for denied capability")
		}
		if err != nil && err.Error() != "capability secrets:read not granted" {
			t.Errorf("Expected capability error, got: %v", err)
		}
	})
}

// TestRegistry tests the provider registry.
func TestRegistry(t *testing.T) {
	tempDir := t.TempDir()

	registry := NewRegistry(tempDir, &WASMHostConfig{
		Timeout:          10 * time.Second,
		MemoryLimitPages: 256,
		TempDir:          tempDir,
	})

	t.Run("SetAllowedCapabilities", func(t *testing.T) {
		capabilities := []string{
			string(engine.CapabilityNetOutbound),
			string(engine.CapabilityFSTemp),
		}
		registry.SetAllowedCapabilities(capabilities)

		// Valid capabilities
		err := registry.ValidateCapabilities(context.Background(), capabilities)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Invalid capability
		err = registry.ValidateCapabilities(context.Background(), []string{
			string(engine.CapabilitySecretsRead),
		})
		if err == nil {
			t.Error("Expected error for disallowed capability")
		}
	})

	t.Run("BuildProviderKey", func(t *testing.T) {
		key := buildProviderKey("test", "1.0.0")
		if key != "test@1.0.0" {
			t.Errorf("Expected 'test@1.0.0', got '%s'", key)
		}
	})

	t.Run("VersionResolution", func(t *testing.T) {
		// Create test manifests
		testManifests := map[string]*Manifest{
			"test@1.0.0": {
				Raw: &engine.ProviderManifest{
					Metadata: engine.ProviderMetadata{
						Name:    "test",
						Version: "1.0.0",
					},
				},
			},
			"test@1.0.1": {
				Raw: &engine.ProviderManifest{
					Metadata: engine.ProviderMetadata{
						Name:    "test",
						Version: "1.0.1",
					},
				},
			},
			"test@1.1.0": {
				Raw: &engine.ProviderManifest{
					Metadata: engine.ProviderMetadata{
						Name:    "test",
						Version: "1.1.0",
					},
				},
			},
		}

		registry.manifests = testManifests

		// Test exact version
		key, err := registry.resolveVersion("test", "1.0.0")
		if err != nil {
			t.Errorf("Failed to resolve exact version: %v", err)
		}
		if key != "test@1.0.0" {
			t.Errorf("Expected 'test@1.0.0', got '%s'", key)
		}

		// Test latest
		key, err = registry.resolveVersion("test", "latest")
		if err != nil {
			t.Errorf("Failed to resolve latest version: %v", err)
		}
		if key != "test@1.1.0" {
			t.Errorf("Expected 'test@1.1.0', got '%s'", key)
		}

		// Test tilde range
		key, err = registry.resolveVersion("test", "~1.0.0")
		if err != nil {
			t.Errorf("Failed to resolve tilde version: %v", err)
		}
		if key != "test@1.0.1" {
			t.Errorf("Expected 'test@1.0.1', got '%s'", key)
		}

		// Test not found
		_, err = registry.resolveVersion("nonexistent", "1.0.0")
		if err == nil {
			t.Error("Expected error for non-existent provider")
		}
	})
}

// TestManifestFromFile tests loading a manifest from a file.
func TestManifestFromFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create test manifest file
	manifestYAML := `
metadata:
  name: file-provider
  version: 1.0.0
  author: Test Author
  license: MIT
  description: Test provider from file
  required_capabilities:
    - net:outbound

schema:
  version: "1.0"
  resource_types:
    test_resource:
      name: test_resource
      description: A test resource
      config_schema: '{"type": "object"}'
      state_schema: '{"type": "object"}'

entrypoint: test.wasm
checksum: ""
`

	manifestPath := filepath.Join(tempDir, "manifest.yaml")
	err := os.WriteFile(manifestPath, []byte(manifestYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to write manifest file: %v", err)
	}

	// Create fake WASM file
	wasmPath := filepath.Join(tempDir, "test.wasm")
	err = os.WriteFile(wasmPath, []byte("fake wasm"), 0644)
	if err != nil {
		t.Fatalf("Failed to write WASM file: %v", err)
	}

	// Load manifest
	loader := NewManifestLoader(tempDir)
	manifest, err := loader.LoadFromFile(manifestPath)
	if err != nil {
		t.Fatalf("Failed to load manifest from file: %v", err)
	}

	if manifest.Raw.Metadata.Name != "file-provider" {
		t.Errorf("Expected name 'file-provider', got '%s'", manifest.Raw.Metadata.Name)
	}

	if manifest.WasmPath != wasmPath {
		t.Errorf("Expected WASM path '%s', got '%s'", wasmPath, manifest.WasmPath)
	}
}

// TestSensitiveFiltering tests sensitive file and env var filtering.
func TestSensitiveFiltering(t *testing.T) {
	enforcer := NewCapabilityEnforcer(
		[]string{string(engine.CapabilityFSRead), string(engine.CapabilityEnvRead)},
		os.TempDir(),
		"",
	)

	t.Run("SensitiveFiles", func(t *testing.T) {
		sensitivePaths := []string{
			"/etc/shadow",
			"/etc/passwd",
			"/root/.ssh/id_rsa",
		}

		for _, path := range sensitivePaths {
			_, err := enforcer.ReadFile(path)
			if err == nil {
				t.Errorf("Expected error for sensitive file: %s", path)
			}
		}
	})

	t.Run("SensitiveEnvVars", func(t *testing.T) {
		sensitiveVars := []string{
			"AWS_SECRET_ACCESS_KEY",
			"GITHUB_TOKEN",
			"DATABASE_PASSWORD",
			"API_KEY",
		}

		for _, varName := range sensitiveVars {
			_, err := enforcer.ReadEnv(varName)
			if err == nil {
				t.Errorf("Expected error for sensitive env var: %s", varName)
			}
		}
	})
}

// BenchmarkCapabilityCheck benchmarks capability checking.
func BenchmarkCapabilityCheck(b *testing.B) {
	enforcer := NewCapabilityEnforcer(
		[]string{string(engine.CapabilityFSTemp), string(engine.CapabilityNetOutbound)},
		os.TempDir(),
		"",
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = enforcer.HasCapability(engine.CapabilityFSTemp)
	}
}

// BenchmarkTempFileWrite benchmarks temporary file writing.
func BenchmarkTempFileWrite(b *testing.B) {
	tempDir := b.TempDir()
	enforcer := NewCapabilityEnforcer(
		[]string{string(engine.CapabilityFSTemp)},
		tempDir,
		"",
	)

	testData := []byte("test data for benchmarking")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filename := filepath.Join(tempDir, "bench.txt")
		_ = enforcer.WriteTempFile(filename, testData)
	}

	b.StopTimer()
	_ = enforcer.Cleanup()
}
