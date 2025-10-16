package policy

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestLoadFromFile_Rego(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	loader := NewLoader(logger)

	// Create a temporary .rego file
	tmpDir := t.TempDir()
	policyFile := filepath.Join(tmpDir, "test-policy.rego")

	regoContent := `package test.policy

# Test policy for validation

deny[msg] {
	input.resource.name == "invalid"
	msg := "Invalid resource name"
}`

	err := os.WriteFile(policyFile, []byte(regoContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	policy, err := loader.loadFromFile(context.Background(), policyFile)
	if err != nil {
		t.Fatalf("Failed to load policy: %v", err)
	}

	if policy.Name != "test-policy" {
		t.Errorf("Expected name 'test-policy', got '%s'", policy.Name)
	}

	if policy.Rego != regoContent {
		t.Error("Rego content doesn't match")
	}

	if !policy.Enabled {
		t.Error("Policy should be enabled by default")
	}
}

func TestLoadFromFile_JSON(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	loader := NewLoader(logger)

	tmpDir := t.TempDir()
	policyFile := filepath.Join(tmpDir, "test-policy.json")

	policy := Policy{
		Name:        "test-json-policy",
		Description: "A test policy",
		Rego:        "package test\ndeny[msg] { false }",
		Severity:    SeverityError,
		Enabled:     true,
		Tags:        []string{"test"},
	}

	data, err := json.Marshal(policy)
	if err != nil {
		t.Fatalf("Failed to marshal policy: %v", err)
	}

	err = os.WriteFile(policyFile, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	loaded, err := loader.loadFromFile(context.Background(), policyFile)
	if err != nil {
		t.Fatalf("Failed to load policy: %v", err)
	}

	if loaded.Name != policy.Name {
		t.Errorf("Expected name '%s', got '%s'", policy.Name, loaded.Name)
	}

	if loaded.Description != policy.Description {
		t.Errorf("Expected description '%s', got '%s'", policy.Description, loaded.Description)
	}

	if loaded.Severity != policy.Severity {
		t.Errorf("Expected severity '%s', got '%s'", policy.Severity, loaded.Severity)
	}
}

func TestLoadFromDirectory(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	loader := NewLoader(logger)

	tmpDir := t.TempDir()

	// Create multiple policy files
	policies := map[string]string{
		"policy1.rego": `package policy1
deny[msg] { false }`,
		"policy2.rego": `package policy2
deny[msg] { false }`,
		"policy3.rego": `package policy3
deny[msg] { false }`,
	}

	for filename, content := range policies {
		path := filepath.Join(tmpDir, filename)
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}
	}

	// Also create a non-policy file that should be ignored
	err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Test"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	loaded, err := loader.loadFromDirectory(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Failed to load directory: %v", err)
	}

	if len(loaded) != len(policies) {
		t.Errorf("Expected %d policies, got %d", len(policies), len(loaded))
	}
}

func TestLoadFromDirectory_Recursive(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	loader := NewLoader(logger)

	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	err := os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create policies in both directories
	err = os.WriteFile(filepath.Join(tmpDir, "policy1.rego"), []byte("package p1\ndeny[msg] { false }"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	err = os.WriteFile(filepath.Join(subDir, "policy2.rego"), []byte("package p2\ndeny[msg] { false }"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	loaded, err := loader.loadFromDirectory(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Failed to load directory: %v", err)
	}

	if len(loaded) != 2 {
		t.Errorf("Expected 2 policies (including subdirectory), got %d", len(loaded))
	}
}

func TestLoadFromPaths(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	loader := NewLoader(logger)

	tmpDir := t.TempDir()

	// Create a directory with policies
	dir1 := filepath.Join(tmpDir, "dir1")
	err := os.Mkdir(dir1, 0755)
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	err = os.WriteFile(filepath.Join(dir1, "policy1.rego"), []byte("package p1\ndeny[msg] { false }"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create a single policy file
	file1 := filepath.Join(tmpDir, "policy2.rego")
	err = os.WriteFile(file1, []byte("package p2\ndeny[msg] { false }"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	paths := []string{dir1, file1}
	loaded, err := loader.LoadFromPaths(context.Background(), paths)
	if err != nil {
		t.Fatalf("Failed to load paths: %v", err)
	}

	if len(loaded) != 2 {
		t.Errorf("Expected 2 policies, got %d", len(loaded))
	}
}

func TestLoadBundle(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	loader := NewLoader(logger)

	tmpDir := t.TempDir()
	bundleFile := filepath.Join(tmpDir, "bundle.json")

	bundle := PolicyBundle{
		Name:        "test-bundle",
		Version:     "1.0.0",
		Description: "Test policy bundle",
		Policies: []Policy{
			{
				Name:        "policy1",
				Description: "First policy",
				Rego:        "package p1\ndeny[msg] { false }",
				Severity:    SeverityError,
				Enabled:     true,
			},
			{
				Name:        "policy2",
				Description: "Second policy",
				Rego:        "package p2\ndeny[msg] { false }",
				Severity:    SeverityWarning,
				Enabled:     true,
			},
		},
		CreatedAt: time.Now(),
	}

	data, err := json.Marshal(bundle)
	if err != nil {
		t.Fatalf("Failed to marshal bundle: %v", err)
	}

	err = os.WriteFile(bundleFile, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write bundle file: %v", err)
	}

	loaded, err := loader.LoadBundle(context.Background(), bundleFile)
	if err != nil {
		t.Fatalf("Failed to load bundle: %v", err)
	}

	if loaded.Name != bundle.Name {
		t.Errorf("Expected bundle name '%s', got '%s'", bundle.Name, loaded.Name)
	}

	if loaded.Version != bundle.Version {
		t.Errorf("Expected version '%s', got '%s'", bundle.Version, loaded.Version)
	}

	if len(loaded.Policies) != len(bundle.Policies) {
		t.Errorf("Expected %d policies, got %d", len(bundle.Policies), len(loaded.Policies))
	}
}

func TestExtractDescription(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	loader := NewLoader(logger)

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name: "single line comment",
			content: `# This is a test policy
package test`,
			expected: "This is a test policy",
		},
		{
			name: "multi line comments",
			content: `# This is a test policy
# that spans multiple lines
package test`,
			expected: "This is a test policy that spans multiple lines",
		},
		{
			name: "no comments",
			content: `package test
deny[msg] { false }`,
			expected: "",
		},
		{
			name: "comments with empty lines",
			content: `# First line
#
# Second line
package test`,
			expected: "First line Second line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := loader.extractDescription(tt.content)
			if result != tt.expected {
				t.Errorf("Expected description '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestClearCache(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	loader := NewLoader(logger)

	tmpDir := t.TempDir()
	policyFile := filepath.Join(tmpDir, "test.rego")
	err := os.WriteFile(policyFile, []byte("package test\ndeny[msg] { false }"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Load a policy to populate cache
	_, err = loader.loadFromFile(context.Background(), policyFile)
	if err != nil {
		t.Fatalf("Failed to load policy: %v", err)
	}

	// Cache should have one entry
	if len(loader.cache) != 1 {
		t.Errorf("Expected 1 cache entry, got %d", len(loader.cache))
	}

	// Clear cache
	loader.ClearCache()

	// Cache should be empty
	if len(loader.cache) != 0 {
		t.Errorf("Expected 0 cache entries after clear, got %d", len(loader.cache))
	}
}

func TestLoadFromFile_UnsupportedType(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	loader := NewLoader(logger)

	tmpDir := t.TempDir()
	policyFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(policyFile, []byte("not a policy"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err = loader.loadFromFile(context.Background(), policyFile)
	if err == nil {
		t.Error("Expected error for unsupported file type")
	}
}

func TestLoadFromFile_InvalidJSON(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	loader := NewLoader(logger)

	tmpDir := t.TempDir()
	policyFile := filepath.Join(tmpDir, "test.json")
	err := os.WriteFile(policyFile, []byte("invalid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err = loader.loadFromFile(context.Background(), policyFile)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestLoadFromPath_NonExistent(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	loader := NewLoader(logger)

	_, err := loader.loadFromPath(context.Background(), "/nonexistent/path")
	if err == nil {
		t.Error("Expected error for non-existent path")
	}
}
