package host

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/openfroyo/openfroyo/pkg/engine"
	"gopkg.in/yaml.v3"
)

// Manifest represents a parsed provider manifest with loaded schemas.
type Manifest struct {
	// Raw is the raw manifest data from the YAML file.
	Raw *engine.ProviderManifest

	// Schemas maps resource type names to their parsed JSON schemas.
	Schemas map[string]*ResourceSchemas

	// Path is the file path where the manifest was loaded from.
	Path string

	// WasmPath is the path to the WASM module.
	WasmPath string

	// Verified indicates if the WASM module checksum has been verified.
	Verified bool
}

// ResourceSchemas contains the parsed JSON schemas for a resource type.
type ResourceSchemas struct {
	// Name is the resource type name.
	Name string

	// ConfigSchema is the parsed config schema.
	ConfigSchema map[string]interface{}

	// StateSchema is the parsed state schema.
	StateSchema map[string]interface{}

	// Capabilities are the required capabilities for this resource type.
	Capabilities []string
}

// ManifestLoader loads and parses provider manifests.
type ManifestLoader struct {
	// BaseDir is the base directory for resolving relative paths.
	BaseDir string
}

// NewManifestLoader creates a new manifest loader.
func NewManifestLoader(baseDir string) *ManifestLoader {
	return &ManifestLoader{
		BaseDir: baseDir,
	}
}

// LoadFromFile loads a manifest from a YAML file.
func (m *ManifestLoader) LoadFromFile(path string) (*Manifest, error) {
	// Read the manifest file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	// Parse YAML
	var raw engine.ProviderManifest
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse manifest YAML: %w", err)
	}

	// Validate manifest
	if err := m.validateManifest(&raw); err != nil {
		return nil, fmt.Errorf("invalid manifest: %w", err)
	}

	manifest := &Manifest{
		Raw:     &raw,
		Path:    path,
		Schemas: make(map[string]*ResourceSchemas),
	}

	// Resolve WASM module path
	if err := m.resolveWasmPath(manifest); err != nil {
		return nil, fmt.Errorf("failed to resolve WASM path: %w", err)
	}

	// Load and parse schemas
	if err := m.loadSchemas(manifest); err != nil {
		return nil, fmt.Errorf("failed to load schemas: %w", err)
	}

	return manifest, nil
}

// LoadFromBytes loads a manifest from raw bytes.
func (m *ManifestLoader) LoadFromBytes(data []byte, wasmModule []byte) (*Manifest, error) {
	// Parse YAML
	var raw engine.ProviderManifest
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse manifest YAML: %w", err)
	}

	// Validate manifest
	if err := m.validateManifest(&raw); err != nil {
		return nil, fmt.Errorf("invalid manifest: %w", err)
	}

	manifest := &Manifest{
		Raw:     &raw,
		Schemas: make(map[string]*ResourceSchemas),
	}

	// Verify WASM module checksum if provided
	if raw.Checksum != "" {
		hash := sha256.Sum256(wasmModule)
		computedChecksum := hex.EncodeToString(hash[:])
		if computedChecksum != raw.Checksum {
			return nil, fmt.Errorf("WASM module checksum mismatch: expected %s, got %s",
				raw.Checksum, computedChecksum)
		}
		manifest.Verified = true
	}

	// Load and parse schemas
	if err := m.loadSchemas(manifest); err != nil {
		return nil, fmt.Errorf("failed to load schemas: %w", err)
	}

	return manifest, nil
}

// validateManifest validates the basic structure of a manifest.
func (m *ManifestLoader) validateManifest(manifest *engine.ProviderManifest) error {
	// Check required metadata fields
	if manifest.Metadata.Name == "" {
		return fmt.Errorf("provider name is required")
	}
	if manifest.Metadata.Version == "" {
		return fmt.Errorf("provider version is required")
	}
	if manifest.Metadata.Author == "" {
		return fmt.Errorf("provider author is required")
	}
	if manifest.Metadata.License == "" {
		return fmt.Errorf("provider license is required")
	}

	// Check entrypoint
	if manifest.Entrypoint == "" {
		return fmt.Errorf("entrypoint is required")
	}

	// Check schema
	if manifest.Schema == nil {
		return fmt.Errorf("schema is required")
	}
	if manifest.Schema.Version == "" {
		return fmt.Errorf("schema version is required")
	}
	if len(manifest.Schema.ResourceTypes) == 0 {
		return fmt.Errorf("at least one resource type is required")
	}

	// Validate each resource type
	for name, rt := range manifest.Schema.ResourceTypes {
		if rt.Name != name {
			return fmt.Errorf("resource type name mismatch: key=%s, name=%s", name, rt.Name)
		}
		if rt.Description == "" {
			return fmt.Errorf("resource type %s: description is required", name)
		}
		if len(rt.ConfigSchema) == 0 {
			return fmt.Errorf("resource type %s: config schema is required", name)
		}
		if len(rt.StateSchema) == 0 {
			return fmt.Errorf("resource type %s: state schema is required", name)
		}
	}

	return nil
}

// resolveWasmPath resolves the path to the WASM module.
func (m *ManifestLoader) resolveWasmPath(manifest *Manifest) error {
	// If entrypoint is absolute, use it directly
	if filepath.IsAbs(manifest.Raw.Entrypoint) {
		manifest.WasmPath = manifest.Raw.Entrypoint
		return nil
	}

	// Resolve relative to manifest path
	if manifest.Path != "" {
		manifestDir := filepath.Dir(manifest.Path)
		manifest.WasmPath = filepath.Join(manifestDir, manifest.Raw.Entrypoint)
	} else {
		// Resolve relative to base directory
		manifest.WasmPath = filepath.Join(m.BaseDir, manifest.Raw.Entrypoint)
	}

	// Verify the WASM file exists
	if _, err := os.Stat(manifest.WasmPath); err != nil {
		return fmt.Errorf("WASM module not found at %s: %w", manifest.WasmPath, err)
	}

	return nil
}

// loadSchemas loads and parses JSON schemas from the manifest.
func (m *ManifestLoader) loadSchemas(manifest *Manifest) error {
	for name, rt := range manifest.Raw.Schema.ResourceTypes {
		schemas := &ResourceSchemas{
			Name:         name,
			Capabilities: rt.Capabilities,
		}

		// Parse config schema
		if len(rt.ConfigSchema) > 0 {
			var configSchema map[string]interface{}
			if err := json.Unmarshal(rt.ConfigSchema, &configSchema); err != nil {
				return fmt.Errorf("failed to parse config schema for %s: %w", name, err)
			}
			schemas.ConfigSchema = configSchema
		}

		// Parse state schema
		if len(rt.StateSchema) > 0 {
			var stateSchema map[string]interface{}
			if err := json.Unmarshal(rt.StateSchema, &stateSchema); err != nil {
				return fmt.Errorf("failed to parse state schema for %s: %w", name, err)
			}
			schemas.StateSchema = stateSchema
		}

		manifest.Schemas[name] = schemas
	}

	return nil
}

// VerifyChecksum verifies the WASM module checksum against the manifest.
func (m *Manifest) VerifyChecksum(wasmModule []byte) error {
	if m.Raw.Checksum == "" {
		return fmt.Errorf("no checksum in manifest")
	}

	hash := sha256.Sum256(wasmModule)
	computedChecksum := hex.EncodeToString(hash[:])

	if computedChecksum != m.Raw.Checksum {
		return fmt.Errorf("WASM module checksum mismatch: expected %s, got %s",
			m.Raw.Checksum, computedChecksum)
	}

	m.Verified = true
	return nil
}

// GetCapabilities returns all capabilities required by this provider.
func (m *Manifest) GetCapabilities() []string {
	capSet := make(map[string]bool)

	// Add metadata capabilities
	for _, cap := range m.Raw.Metadata.RequiredCapabilities {
		capSet[cap] = true
	}

	// Add resource type capabilities
	for _, schemas := range m.Schemas {
		for _, cap := range schemas.Capabilities {
			capSet[cap] = true
		}
	}

	// Convert to slice
	caps := make([]string, 0, len(capSet))
	for cap := range capSet {
		caps = append(caps, cap)
	}

	return caps
}

// GetResourceTypes returns a list of all resource type names.
func (m *Manifest) GetResourceTypes() []string {
	types := make([]string, 0, len(m.Schemas))
	for name := range m.Schemas {
		types = append(types, name)
	}
	return types
}

// GetResourceSchema returns the schema for a specific resource type.
func (m *Manifest) GetResourceSchema(resourceType string) (*ResourceSchemas, error) {
	schema, ok := m.Schemas[resourceType]
	if !ok {
		return nil, fmt.Errorf("resource type %s not found in manifest", resourceType)
	}
	return schema, nil
}
