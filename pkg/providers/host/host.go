package host

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/openfroyo/openfroyo/pkg/engine"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// WASMHostProvider implements the Provider interface using a WASM module.
type WASMHostProvider struct {
	// manifest is the parsed provider manifest.
	manifest *Manifest

	// runtime is the wazero runtime.
	runtime wazero.Runtime

	// module is the compiled WASM module.
	module api.Module

	// bridge is the WASM bridge for calling provider functions.
	bridge *WASMBridge

	// enforcer is the capability enforcer.
	enforcer *CapabilityEnforcer

	// config is the provider configuration.
	config engine.ProviderConfig

	// initialized indicates if the provider has been initialized.
	initialized bool

	// timeout is the default timeout for operations.
	timeout time.Duration

	// memoryLimitPages is the maximum memory limit in pages (64KB each).
	memoryLimitPages uint32
}

// WASMHostConfig contains configuration for the WASM host.
type WASMHostConfig struct {
	// Timeout is the default timeout for WASM operations.
	Timeout time.Duration

	// MemoryLimitPages is the maximum memory limit in pages (64KB each).
	// Default is 256 pages (16MB).
	MemoryLimitPages uint32

	// TempDir is the temporary directory for fs:temp capability.
	TempDir string

	// MicroRunnerPath is the path to the micro-runner executable.
	MicroRunnerPath string
}

// NewWASMHostProvider creates a new WASM host provider from a manifest and WASM module.
func NewWASMHostProvider(ctx context.Context, manifest *Manifest, wasmModule []byte, hostConfig *WASMHostConfig) (*WASMHostProvider, error) {
	// Set default config values
	if hostConfig == nil {
		hostConfig = &WASMHostConfig{
			Timeout:          30 * time.Second,
			MemoryLimitPages: 256, // 16MB
			TempDir:          os.TempDir(),
		}
	}

	if hostConfig.Timeout == 0 {
		hostConfig.Timeout = 30 * time.Second
	}

	if hostConfig.MemoryLimitPages == 0 {
		hostConfig.MemoryLimitPages = 256
	}

	// Create capability enforcer
	capabilities := manifest.GetCapabilities()
	enforcer := NewCapabilityEnforcer(capabilities, hostConfig.TempDir, hostConfig.MicroRunnerPath)

	// Create wazero runtime with configuration
	runtimeConfig := wazero.NewRuntimeConfig().
		WithMemoryLimitPages(hostConfig.MemoryLimitPages).
		WithCloseOnContextDone(true)

	runtime := wazero.NewRuntimeWithConfig(ctx, runtimeConfig)

	// Instantiate WASI
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, runtime); err != nil {
		runtime.Close(ctx)
		return nil, fmt.Errorf("failed to instantiate WASI: %w", err)
	}

	// Create host function builder
	builder := runtime.NewHostModuleBuilder("env")

	// Register host functions for capabilities
	if err := registerHostFunctions(builder, enforcer); err != nil {
		runtime.Close(ctx)
		return nil, fmt.Errorf("failed to register host functions: %w", err)
	}

	// Instantiate host module
	if _, err := builder.Instantiate(ctx); err != nil {
		runtime.Close(ctx)
		return nil, fmt.Errorf("failed to instantiate host module: %w", err)
	}

	// Compile and instantiate WASM module
	module, err := runtime.Instantiate(ctx, wasmModule)
	if err != nil {
		runtime.Close(ctx)
		return nil, fmt.Errorf("failed to instantiate WASM module: %w", err)
	}

	// Create bridge
	bridge, err := NewWASMBridge(module, hostConfig.Timeout)
	if err != nil {
		module.Close(ctx)
		runtime.Close(ctx)
		return nil, fmt.Errorf("failed to create WASM bridge: %w", err)
	}

	provider := &WASMHostProvider{
		manifest:         manifest,
		runtime:          runtime,
		module:           module,
		bridge:           bridge,
		enforcer:         enforcer,
		timeout:          hostConfig.Timeout,
		memoryLimitPages: hostConfig.MemoryLimitPages,
	}

	return provider, nil
}

// registerHostFunctions registers host functions that WASM can call.
func registerHostFunctions(builder wazero.HostModuleBuilder, enforcer *CapabilityEnforcer) error {
	// Register net:outbound capability function
	builder.NewFunctionBuilder().
		WithFunc(func(ctx context.Context, mod api.Module, urlPtr, urlLen, methodPtr, methodLen uint32) uint64 {
			// Read URL from memory
			urlBytes, ok := mod.Memory().Read(urlPtr, urlLen)
			if !ok {
				return packError("failed to read URL from memory")
			}

			// Read method from memory
			methodBytes, ok := mod.Memory().Read(methodPtr, methodLen)
			if !ok {
				return packError("failed to read method from memory")
			}

			url := string(urlBytes)
			method := string(methodBytes)

			// Perform HTTP request
			resp, err := enforcer.HTTPRequest(ctx, method, url, nil)
			if err != nil {
				return packError(err.Error())
			}
			defer resp.Body.Close()

			// Return status code
			return uint64(resp.StatusCode)
		}).
		Export("http_request")

	// Register fs:temp capability functions
	builder.NewFunctionBuilder().
		WithFunc(func(ctx context.Context, mod api.Module, namePtr, nameLen, dataPtr, dataLen uint32) uint32 {
			// Read name from memory
			nameBytes, ok := mod.Memory().Read(namePtr, nameLen)
			if !ok {
				return 1 // Error
			}

			// Read data from memory
			dataBytes, ok := mod.Memory().Read(dataPtr, dataLen)
			if !ok {
				return 1 // Error
			}

			name := string(nameBytes)

			// Write temp file
			if err := enforcer.WriteTempFile(name, dataBytes); err != nil {
				return 1 // Error
			}

			return 0 // Success
		}).
		Export("write_temp_file")

	builder.NewFunctionBuilder().
		WithFunc(func(ctx context.Context, mod api.Module, namePtr, nameLen uint32) uint64 {
			// Read name from memory
			nameBytes, ok := mod.Memory().Read(namePtr, nameLen)
			if !ok {
				return packError("failed to read name from memory")
			}

			name := string(nameBytes)

			// Read temp file
			data, err := enforcer.ReadTempFile(name)
			if err != nil {
				return packError(err.Error())
			}

			// Allocate memory for result (this is simplified - in production,
			// we'd need a proper memory management strategy)
			return uint64(len(data))
		}).
		Export("read_temp_file")

	// Register secrets:read capability function
	builder.NewFunctionBuilder().
		WithFunc(func(ctx context.Context, mod api.Module, encryptedPtr, encryptedLen uint32) uint64 {
			// Read encrypted secret from memory
			encryptedBytes, ok := mod.Memory().Read(encryptedPtr, encryptedLen)
			if !ok {
				return packError("failed to read encrypted secret from memory")
			}

			encrypted := string(encryptedBytes)

			// Decrypt secret
			decrypted, err := enforcer.DecryptSecret(encrypted)
			if err != nil {
				return packError(err.Error())
			}

			// Return decrypted length (simplified)
			return uint64(len(decrypted))
		}).
		Export("decrypt_secret")

	return nil
}

// packError packs an error message into a uint64 return value.
// Format: error_code (upper 32 bits) | length (lower 32 bits)
// error_code = 1 for errors, 0 for success
func packError(msg string) uint64 {
	errorCode := uint64(1) << 32
	return errorCode | uint64(len(msg))
}

// Init initializes the provider with configuration.
func (p *WASMHostProvider) Init(ctx context.Context, config engine.ProviderConfig) error {
	if p.initialized {
		return fmt.Errorf("provider already initialized")
	}

	// Validate capabilities
	if err := p.enforcer.ValidateCapabilities(config.Capabilities); err != nil {
		return fmt.Errorf("capability validation failed: %w", err)
	}

	// Set config
	p.config = config

	// Call provider init
	if err := p.bridge.Init(ctx, config); err != nil {
		return fmt.Errorf("provider initialization failed: %w", err)
	}

	p.initialized = true
	return nil
}

// Read retrieves the current state of a resource.
func (p *WASMHostProvider) Read(ctx context.Context, req engine.ReadRequest) (*engine.ReadResponse, error) {
	if !p.initialized {
		return nil, fmt.Errorf("provider not initialized")
	}

	return p.bridge.Read(ctx, req)
}

// Plan computes the operations needed to reach desired state.
func (p *WASMHostProvider) Plan(ctx context.Context, req engine.PlanRequest) (*engine.PlanResponse, error) {
	if !p.initialized {
		return nil, fmt.Errorf("provider not initialized")
	}

	return p.bridge.Plan(ctx, req)
}

// Apply executes the planned operations to achieve desired state.
func (p *WASMHostProvider) Apply(ctx context.Context, req engine.ApplyRequest) (*engine.ApplyResponse, error) {
	if !p.initialized {
		return nil, fmt.Errorf("provider not initialized")
	}

	return p.bridge.Apply(ctx, req)
}

// Destroy removes the resource completely.
func (p *WASMHostProvider) Destroy(ctx context.Context, req engine.DestroyRequest) (*engine.DestroyResponse, error) {
	if !p.initialized {
		return nil, fmt.Errorf("provider not initialized")
	}

	return p.bridge.Destroy(ctx, req)
}

// Validate validates a resource configuration against the provider's schema.
func (p *WASMHostProvider) Validate(ctx context.Context, config json.RawMessage) error {
	if !p.initialized {
		return fmt.Errorf("provider not initialized")
	}

	return p.bridge.Validate(ctx, config)
}

// Schema returns the JSON schema for this provider's resources.
func (p *WASMHostProvider) Schema() (*engine.ProviderSchema, error) {
	if !p.initialized {
		return nil, fmt.Errorf("provider not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	return p.bridge.Schema(ctx)
}

// Metadata returns information about this provider.
func (p *WASMHostProvider) Metadata() engine.ProviderMetadata {
	return p.manifest.Raw.Metadata
}

// Close closes the provider and releases resources.
func (p *WASMHostProvider) Close(ctx context.Context) error {
	// Clean up temporary files
	if err := p.enforcer.Cleanup(); err != nil {
		// Log but don't fail
		_ = err
	}

	// Close module
	if p.module != nil {
		if err := p.module.Close(ctx); err != nil {
			return fmt.Errorf("failed to close WASM module: %w", err)
		}
	}

	// Close runtime
	if p.runtime != nil {
		if err := p.runtime.Close(ctx); err != nil {
			return fmt.Errorf("failed to close WASM runtime: %w", err)
		}
	}

	return nil
}

// GetManifest returns the provider manifest.
func (p *WASMHostProvider) GetManifest() *Manifest {
	return p.manifest
}

// GetCapabilities returns the granted capabilities.
func (p *WASMHostProvider) GetCapabilities() []string {
	return p.manifest.GetCapabilities()
}

// IsInitialized returns true if the provider has been initialized.
func (p *WASMHostProvider) IsInitialized() bool {
	return p.initialized
}
