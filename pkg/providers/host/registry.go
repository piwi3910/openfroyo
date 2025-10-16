package host

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/openfroyo/openfroyo/pkg/engine"
	"gopkg.in/yaml.v3"
)

// Registry implements the ProviderRegistry interface for WASM providers.
type Registry struct {
	// mu protects the registry state.
	mu sync.RWMutex

	// providers maps provider key (name@version) to provider instance.
	providers map[string]*WASMHostProvider

	// manifests maps provider key to manifest.
	manifests map[string]*Manifest

	// wasmModules maps provider key to WASM module bytes.
	wasmModules map[string][]byte

	// loader is the manifest loader.
	loader *ManifestLoader

	// hostConfig is the default host configuration.
	hostConfig *WASMHostConfig

	// allowedCapabilities is the set of capabilities allowed in this registry.
	allowedCapabilities map[string]bool
}

// NewRegistry creates a new provider registry.
func NewRegistry(baseDir string, hostConfig *WASMHostConfig) *Registry {
	if hostConfig == nil {
		hostConfig = &WASMHostConfig{
			Timeout:          30,
			MemoryLimitPages: 256,
			TempDir:          os.TempDir(),
		}
	}

	return &Registry{
		providers:           make(map[string]*WASMHostProvider),
		manifests:           make(map[string]*Manifest),
		wasmModules:         make(map[string][]byte),
		loader:              NewManifestLoader(baseDir),
		hostConfig:          hostConfig,
		allowedCapabilities: make(map[string]bool),
	}
}

// SetAllowedCapabilities sets the capabilities allowed in this registry.
func (r *Registry) SetAllowedCapabilities(capabilities []string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.allowedCapabilities = make(map[string]bool)
	for _, cap := range capabilities {
		r.allowedCapabilities[cap] = true
	}
}

// Register registers a provider plugin.
func (r *Registry) Register(ctx context.Context, manifest *engine.ProviderManifest, wasmModule []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Convert manifest to YAML bytes
	manifestBytes, err := yaml.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	// Parse manifest
	parsedManifest, err := r.loader.LoadFromBytes(manifestBytes, wasmModule)
	if err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Build provider key
	key := buildProviderKey(manifest.Metadata.Name, manifest.Metadata.Version)

	// Check if already registered
	if _, exists := r.providers[key]; exists {
		return fmt.Errorf("provider %s already registered", key)
	}

	// Validate capabilities
	if err := r.ValidateCapabilities(ctx, parsedManifest.GetCapabilities()); err != nil {
		return fmt.Errorf("capability validation failed: %w", err)
	}

	// Store manifest and WASM module
	r.manifests[key] = parsedManifest
	r.wasmModules[key] = wasmModule

	return nil
}

// RegisterFromPath registers a provider from a manifest file and WASM module.
func (r *Registry) RegisterFromPath(ctx context.Context, manifestPath string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Load manifest
	manifest, err := r.loader.LoadFromFile(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	// Read WASM module
	wasmModule, err := os.ReadFile(manifest.WasmPath)
	if err != nil {
		return fmt.Errorf("failed to read WASM module: %w", err)
	}

	// Verify checksum if provided
	if manifest.Raw.Checksum != "" {
		if err := manifest.VerifyChecksum(wasmModule); err != nil {
			return fmt.Errorf("checksum verification failed: %w", err)
		}
	}

	// Build provider key
	key := buildProviderKey(manifest.Raw.Metadata.Name, manifest.Raw.Metadata.Version)

	// Check if already registered
	if _, exists := r.providers[key]; exists {
		return fmt.Errorf("provider %s already registered", key)
	}

	// Validate capabilities
	if err := r.ValidateCapabilities(ctx, manifest.GetCapabilities()); err != nil {
		return fmt.Errorf("capability validation failed: %w", err)
	}

	// Store manifest and WASM module
	r.manifests[key] = manifest
	r.wasmModules[key] = wasmModule

	return nil
}

// Get retrieves a provider by name and version.
func (r *Registry) Get(ctx context.Context, name, version string) (engine.Provider, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Resolve version
	key, err := r.resolveVersion(name, version)
	if err != nil {
		return nil, err
	}

	// Check if provider is already loaded
	if provider, exists := r.providers[key]; exists {
		return provider, nil
	}

	// Load provider
	manifest, exists := r.manifests[key]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", key)
	}

	wasmModule, exists := r.wasmModules[key]
	if !exists {
		return nil, fmt.Errorf("WASM module for provider %s not found", key)
	}

	// Create provider instance
	provider, err := NewWASMHostProvider(ctx, manifest, wasmModule, r.hostConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	// Cache provider
	r.providers[key] = provider

	return provider, nil
}

// List lists all registered providers.
func (r *Registry) List(ctx context.Context) ([]engine.ProviderMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metadata := make([]engine.ProviderMetadata, 0, len(r.manifests))
	for _, manifest := range r.manifests {
		metadata = append(metadata, manifest.Raw.Metadata)
	}

	return metadata, nil
}

// Unregister removes a provider from the registry.
func (r *Registry) Unregister(ctx context.Context, name, version string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := buildProviderKey(name, version)

	// Close provider if loaded
	if provider, exists := r.providers[key]; exists {
		if err := provider.Close(ctx); err != nil {
			return fmt.Errorf("failed to close provider: %w", err)
		}
		delete(r.providers, key)
	}

	// Remove manifest and WASM module
	delete(r.manifests, key)
	delete(r.wasmModules, key)

	return nil
}

// ValidateCapabilities validates that requested capabilities are allowed.
func (r *Registry) ValidateCapabilities(ctx context.Context, capabilities []string) error {
	// If no allowed capabilities are set, allow all
	if len(r.allowedCapabilities) == 0 {
		return nil
	}

	var denied []string
	for _, cap := range capabilities {
		if !r.allowedCapabilities[cap] {
			denied = append(denied, cap)
		}
	}

	if len(denied) > 0 {
		return fmt.Errorf("capabilities not allowed: %v", denied)
	}

	return nil
}

// Close closes all loaded providers and releases resources.
func (r *Registry) Close(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var errors []error
	for key, provider := range r.providers {
		if err := provider.Close(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to close provider %s: %w", key, err))
		}
	}

	// Clear all maps
	r.providers = make(map[string]*WASMHostProvider)
	r.manifests = make(map[string]*Manifest)
	r.wasmModules = make(map[string][]byte)

	if len(errors) > 0 {
		return fmt.Errorf("errors closing providers: %v", errors)
	}

	return nil
}

// resolveVersion resolves a version constraint to an exact version.
// Supports:
// - Exact version: "1.0.0"
// - Latest: "latest" or ""
// - Tilde range: "~1.0.0" (matches 1.0.x)
// - Caret range: "^1.0.0" (matches 1.x.x)
func (r *Registry) resolveVersion(name, version string) (string, error) {
	// Handle latest
	if version == "" || version == "latest" {
		return r.findLatestVersion(name)
	}

	// Handle tilde range (~1.0.0)
	if strings.HasPrefix(version, "~") {
		return r.findTildeVersion(name, version[1:])
	}

	// Handle caret range (^1.0.0)
	if strings.HasPrefix(version, "^") {
		return r.findCaretVersion(name, version[1:])
	}

	// Exact version
	key := buildProviderKey(name, version)
	if _, exists := r.manifests[key]; !exists {
		return "", fmt.Errorf("provider %s not found", key)
	}

	return key, nil
}

// findLatestVersion finds the latest version of a provider.
func (r *Registry) findLatestVersion(name string) (string, error) {
	var latest string
	for key := range r.manifests {
		if strings.HasPrefix(key, name+"@") {
			// Simple string comparison - in production, use semantic versioning
			if latest == "" || key > latest {
				latest = key
			}
		}
	}

	if latest == "" {
		return "", fmt.Errorf("provider %s not found", name)
	}

	return latest, nil
}

// findTildeVersion finds a version matching the tilde constraint.
func (r *Registry) findTildeVersion(name, version string) (string, error) {
	// Extract major.minor
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid version format: %s", version)
	}

	prefix := name + "@" + parts[0] + "." + parts[1]

	var match string
	for key := range r.manifests {
		if strings.HasPrefix(key, prefix) {
			if match == "" || key > match {
				match = key
			}
		}
	}

	if match == "" {
		return "", fmt.Errorf("no version matching ~%s found for provider %s", version, name)
	}

	return match, nil
}

// findCaretVersion finds a version matching the caret constraint.
func (r *Registry) findCaretVersion(name, version string) (string, error) {
	// Extract major
	parts := strings.Split(version, ".")
	if len(parts) < 1 {
		return "", fmt.Errorf("invalid version format: %s", version)
	}

	prefix := name + "@" + parts[0]

	var match string
	for key := range r.manifests {
		if strings.HasPrefix(key, prefix) {
			if match == "" || key > match {
				match = key
			}
		}
	}

	if match == "" {
		return "", fmt.Errorf("no version matching ^%s found for provider %s", version, name)
	}

	return match, nil
}

// buildProviderKey builds a unique key for a provider.
func buildProviderKey(name, version string) string {
	return name + "@" + version
}

// ScanDirectory scans a directory for provider manifests and registers them.
func (r *Registry) ScanDirectory(ctx context.Context, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Look for manifest.yaml in subdirectory
			manifestPath := filepath.Join(dir, entry.Name(), "manifest.yaml")
			if _, err := os.Stat(manifestPath); err == nil {
				if err := r.RegisterFromPath(ctx, manifestPath); err != nil {
					// Log error but continue
					fmt.Printf("Warning: failed to register provider from %s: %v\n", manifestPath, err)
				}
			}
		}
	}

	return nil
}

// GetProviderInfo returns information about a registered provider.
func (r *Registry) GetProviderInfo(name, version string) (*engine.ProviderMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := buildProviderKey(name, version)
	manifest, exists := r.manifests[key]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", key)
	}

	metadata := manifest.Raw.Metadata
	return &metadata, nil
}
