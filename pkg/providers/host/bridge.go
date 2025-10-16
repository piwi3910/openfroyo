package host

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/openfroyo/openfroyo/pkg/engine"
	"github.com/tetratelabs/wazero/api"
)

// WASMBridge provides a bridge between Go and WASM provider functions.
type WASMBridge struct {
	// module is the WASM module instance.
	module api.Module

	// memory provides access to WASM linear memory.
	memory api.Memory

	// malloc is the memory allocation function exported by WASM.
	malloc api.Function

	// free is the memory deallocation function exported by WASM.
	free api.Function

	// providerInit is the init function exported by the provider.
	providerInit api.Function

	// providerRead is the read function exported by the provider.
	providerRead api.Function

	// providerPlan is the plan function exported by the provider.
	providerPlan api.Function

	// providerApply is the apply function exported by the provider.
	providerApply api.Function

	// providerDestroy is the destroy function exported by the provider.
	providerDestroy api.Function

	// providerValidate is the validate function exported by the provider.
	providerValidate api.Function

	// providerSchema is the schema function exported by the provider.
	providerSchema api.Function

	// providerMetadata is the metadata function exported by the provider.
	providerMetadata api.Function

	// timeout is the default timeout for WASM operations.
	timeout time.Duration
}

// NewWASMBridge creates a new WASM bridge for the given module.
func NewWASMBridge(module api.Module, timeout time.Duration) (*WASMBridge, error) {
	bridge := &WASMBridge{
		module:  module,
		timeout: timeout,
	}

	// Get memory
	bridge.memory = module.Memory()
	if bridge.memory == nil {
		return nil, fmt.Errorf("WASM module does not export memory")
	}

	// Get memory management functions
	bridge.malloc = module.ExportedFunction("malloc")
	if bridge.malloc == nil {
		return nil, fmt.Errorf("WASM module does not export malloc function")
	}

	bridge.free = module.ExportedFunction("free")
	if bridge.free == nil {
		return nil, fmt.Errorf("WASM module does not export free function")
	}

	// Get provider functions (all are required)
	bridge.providerInit = module.ExportedFunction("provider_init")
	if bridge.providerInit == nil {
		return nil, fmt.Errorf("WASM module does not export provider_init function")
	}

	bridge.providerRead = module.ExportedFunction("provider_read")
	if bridge.providerRead == nil {
		return nil, fmt.Errorf("WASM module does not export provider_read function")
	}

	bridge.providerPlan = module.ExportedFunction("provider_plan")
	if bridge.providerPlan == nil {
		return nil, fmt.Errorf("WASM module does not export provider_plan function")
	}

	bridge.providerApply = module.ExportedFunction("provider_apply")
	if bridge.providerApply == nil {
		return nil, fmt.Errorf("WASM module does not export provider_apply function")
	}

	bridge.providerDestroy = module.ExportedFunction("provider_destroy")
	if bridge.providerDestroy == nil {
		return nil, fmt.Errorf("WASM module does not export provider_destroy function")
	}

	bridge.providerValidate = module.ExportedFunction("provider_validate")
	if bridge.providerValidate == nil {
		return nil, fmt.Errorf("WASM module does not export provider_validate function")
	}

	bridge.providerSchema = module.ExportedFunction("provider_schema")
	if bridge.providerSchema == nil {
		return nil, fmt.Errorf("WASM module does not export provider_schema function")
	}

	bridge.providerMetadata = module.ExportedFunction("provider_metadata")
	if bridge.providerMetadata == nil {
		return nil, fmt.Errorf("WASM module does not export provider_metadata function")
	}

	return bridge, nil
}

// Init calls the provider's init function.
func (b *WASMBridge) Init(ctx context.Context, config engine.ProviderConfig) error {
	// Marshal config to JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Call WASM function with timeout
	ctx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	result, err := b.callWASMFunction(ctx, b.providerInit, configJSON)
	if err != nil {
		return fmt.Errorf("provider_init failed: %w", err)
	}

	// Check for error response
	if result != nil {
		var errorResp struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(result, &errorResp); err == nil && errorResp.Error != "" {
			return fmt.Errorf("provider init error: %s", errorResp.Error)
		}
	}

	return nil
}

// Read calls the provider's read function.
func (b *WASMBridge) Read(ctx context.Context, req engine.ReadRequest) (*engine.ReadResponse, error) {
	// Marshal request to JSON
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Call WASM function with timeout
	ctx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	resultJSON, err := b.callWASMFunction(ctx, b.providerRead, reqJSON)
	if err != nil {
		return nil, fmt.Errorf("provider_read failed: %w", err)
	}

	// Parse response
	var resp engine.ReadResponse
	if err := json.Unmarshal(resultJSON, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

// Plan calls the provider's plan function.
func (b *WASMBridge) Plan(ctx context.Context, req engine.PlanRequest) (*engine.PlanResponse, error) {
	// Marshal request to JSON
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Call WASM function with timeout
	ctx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	resultJSON, err := b.callWASMFunction(ctx, b.providerPlan, reqJSON)
	if err != nil {
		return nil, fmt.Errorf("provider_plan failed: %w", err)
	}

	// Parse response
	var resp engine.PlanResponse
	if err := json.Unmarshal(resultJSON, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

// Apply calls the provider's apply function.
func (b *WASMBridge) Apply(ctx context.Context, req engine.ApplyRequest) (*engine.ApplyResponse, error) {
	// Marshal request to JSON
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Call WASM function with timeout
	ctx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	resultJSON, err := b.callWASMFunction(ctx, b.providerApply, reqJSON)
	if err != nil {
		return nil, fmt.Errorf("provider_apply failed: %w", err)
	}

	// Parse response
	var resp engine.ApplyResponse
	if err := json.Unmarshal(resultJSON, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

// Destroy calls the provider's destroy function.
func (b *WASMBridge) Destroy(ctx context.Context, req engine.DestroyRequest) (*engine.DestroyResponse, error) {
	// Marshal request to JSON
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Call WASM function with timeout
	ctx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	resultJSON, err := b.callWASMFunction(ctx, b.providerDestroy, reqJSON)
	if err != nil {
		return nil, fmt.Errorf("provider_destroy failed: %w", err)
	}

	// Parse response
	var resp engine.DestroyResponse
	if err := json.Unmarshal(resultJSON, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

// Validate calls the provider's validate function.
func (b *WASMBridge) Validate(ctx context.Context, config json.RawMessage) error {
	// Call WASM function with timeout
	ctx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	resultJSON, err := b.callWASMFunction(ctx, b.providerValidate, config)
	if err != nil {
		return fmt.Errorf("provider_validate failed: %w", err)
	}

	// Check for validation errors
	var result struct {
		Valid  bool     `json:"valid"`
		Errors []string `json:"errors,omitempty"`
	}
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return fmt.Errorf("failed to unmarshal validation result: %w", err)
	}

	if !result.Valid {
		return fmt.Errorf("validation failed: %v", result.Errors)
	}

	return nil
}

// Schema calls the provider's schema function.
func (b *WASMBridge) Schema(ctx context.Context) (*engine.ProviderSchema, error) {
	// Call WASM function with timeout
	ctx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	resultJSON, err := b.callWASMFunction(ctx, b.providerSchema, nil)
	if err != nil {
		return nil, fmt.Errorf("provider_schema failed: %w", err)
	}

	// Parse response
	var schema engine.ProviderSchema
	if err := json.Unmarshal(resultJSON, &schema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema: %w", err)
	}

	return &schema, nil
}

// Metadata calls the provider's metadata function.
func (b *WASMBridge) Metadata(ctx context.Context) (*engine.ProviderMetadata, error) {
	// Call WASM function with timeout
	ctx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	resultJSON, err := b.callWASMFunction(ctx, b.providerMetadata, nil)
	if err != nil {
		return nil, fmt.Errorf("provider_metadata failed: %w", err)
	}

	// Parse response
	var metadata engine.ProviderMetadata
	if err := json.Unmarshal(resultJSON, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &metadata, nil
}

// callWASMFunction calls a WASM function with JSON input/output.
// Returns the JSON response or an error.
func (b *WASMBridge) callWASMFunction(ctx context.Context, fn api.Function, input []byte) ([]byte, error) {
	// Handle nil input
	var inputPtr, inputLen uint32
	if input != nil && len(input) > 0 {
		// Allocate memory in WASM for input
		ptr, err := b.allocate(ctx, uint32(len(input)))
		if err != nil {
			return nil, fmt.Errorf("failed to allocate WASM memory: %w", err)
		}
		defer b.deallocate(ctx, ptr)

		inputPtr = ptr
		inputLen = uint32(len(input))

		// Write input to WASM memory
		if !b.memory.Write(inputPtr, input) {
			return nil, fmt.Errorf("failed to write input to WASM memory")
		}
	}

	// Call the WASM function
	// Function signature: fn(input_ptr: u32, input_len: u32) -> u64
	// Return value is (output_ptr << 32) | output_len
	results, err := fn.Call(ctx, uint64(inputPtr), uint64(inputLen))
	if err != nil {
		return nil, fmt.Errorf("WASM function call failed: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("WASM function returned no results")
	}

	// Extract output pointer and length from packed uint64
	packed := results[0]
	outputPtr := uint32(packed >> 32)
	outputLen := uint32(packed & 0xFFFFFFFF)

	// Handle empty output
	if outputLen == 0 {
		return []byte("{}"), nil
	}

	// Read output from WASM memory
	output, ok := b.memory.Read(outputPtr, outputLen)
	if !ok {
		return nil, fmt.Errorf("failed to read output from WASM memory")
	}

	// Free the output memory (allocated by WASM)
	if err := b.deallocate(ctx, outputPtr); err != nil {
		// Log but don't fail - output was already read
		// In production, this should be logged properly
		_ = err
	}

	return output, nil
}

// allocate allocates memory in WASM and returns the pointer.
func (b *WASMBridge) allocate(ctx context.Context, size uint32) (uint32, error) {
	results, err := b.malloc.Call(ctx, uint64(size))
	if err != nil {
		return 0, fmt.Errorf("malloc failed: %w", err)
	}

	if len(results) == 0 {
		return 0, fmt.Errorf("malloc returned no results")
	}

	ptr := uint32(results[0])
	if ptr == 0 {
		return 0, fmt.Errorf("malloc returned null pointer")
	}

	return ptr, nil
}

// deallocate frees memory in WASM.
func (b *WASMBridge) deallocate(ctx context.Context, ptr uint32) error {
	_, err := b.free.Call(ctx, uint64(ptr))
	if err != nil {
		return fmt.Errorf("free failed: %w", err)
	}
	return nil
}
