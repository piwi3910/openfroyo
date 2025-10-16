# WASM Provider Host Runtime

This package implements the WASM provider host runtime for OpenFroyo, enabling sandboxed execution of infrastructure providers compiled to WebAssembly.

## Architecture Overview

The WASM provider host runtime consists of several key components:

```
┌──────────────────────────────────────────────────────────┐
│                    Provider Registry                      │
│  ┌────────────────────────────────────────────────────┐  │
│  │  Provider Lifecycle Management                     │  │
│  │  - Load from OCI images or local files            │  │
│  │  - Version resolution (exact, ~, ^, latest)       │  │
│  │  - Provider caching                                │  │
│  └────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────┘
                         │
                         ▼
┌──────────────────────────────────────────────────────────┐
│                  WASM Host Provider                       │
│  ┌────────────────────────────────────────────────────┐  │
│  │  Wazero Runtime                                    │  │
│  │  - Pure Go WASM runtime (no CGO)                  │  │
│  │  - WASI support                                    │  │
│  │  - Memory limit enforcement                        │  │
│  │  - Timeout control                                 │  │
│  └────────────────────────────────────────────────────┘  │
│  ┌────────────────────────────────────────────────────┐  │
│  │  WASM Bridge                                       │  │
│  │  - Go ↔ WASM function calls                       │  │
│  │  - JSON marshaling/unmarshaling                   │  │
│  │  - Memory management (malloc/free)                │  │
│  └────────────────────────────────────────────────────┘  │
│  ┌────────────────────────────────────────────────────┐  │
│  │  Capability Enforcer                               │  │
│  │  - net:outbound - HTTP requests                   │  │
│  │  - fs:temp - Temporary file access                │  │
│  │  - fs:read/write - File system access             │  │
│  │  - env:read - Environment variables               │  │
│  │  - secrets:read - Secret decryption               │  │
│  │  - exec:micro-runner - Delegate to micro-runner   │  │
│  └────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────┘
```

## Components

### 1. Manifest Parser (`manifest.go`)

Parses and validates provider manifest files (YAML).

**Key Features:**
- YAML manifest parsing
- Schema validation
- JSON schema loading and caching
- WASM module checksum verification
- Capability extraction

**Example Manifest:**
```yaml
metadata:
  name: linux-pkg
  version: 1.0.0
  author: OpenFroyo Team
  license: MIT
  description: Linux package management provider
  required_capabilities:
    - exec:micro-runner

schema:
  version: "1.0"
  resource_types:
    package:
      name: package
      description: System package management
      config_schema: '{"type": "object", "properties": {"name": {"type": "string"}}}'
      state_schema: '{"type": "object", "properties": {"installed": {"type": "boolean"}}}'
      capabilities:
        - exec:micro-runner

entrypoint: provider.wasm
checksum: sha256:abc123...
```

### 2. WASM Bridge (`bridge.go`)

Provides the interface between Go and WASM provider functions.

**WASM Provider Contract:**
```
// All functions return packed uint64: (ptr << 32) | len
provider_init(config_ptr: u32, config_len: u32) -> u64
provider_read(request_ptr: u32, request_len: u32) -> u64
provider_plan(request_ptr: u32, request_len: u32) -> u64
provider_apply(request_ptr: u32, request_len: u32) -> u64
provider_destroy(request_ptr: u32, request_len: u32) -> u64
provider_validate(config_ptr: u32, config_len: u32) -> u64
provider_schema() -> u64
provider_metadata() -> u64
```

**Memory Management:**
```
malloc(size: u32) -> u32
free(ptr: u32) -> void
```

### 3. WASM Host Runtime (`host.go`)

Manages WASM module instantiation and execution.

**Configuration:**
```go
config := &WASMHostConfig{
    Timeout:          30 * time.Second,  // Operation timeout
    MemoryLimitPages: 256,                // 16MB memory limit
    TempDir:          "/tmp/providers",   // Temp directory for fs:temp
    MicroRunnerPath:  "/usr/bin/micro-runner",
}
```

**Key Features:**
- Wazero runtime integration
- WASI snapshot preview 1 support
- Memory limit enforcement (configurable pages)
- Operation timeouts
- Host function registration for capabilities

### 4. Capability Enforcer (`capabilities.go`)

Enforces capability-based security for provider operations.

**Supported Capabilities:**

| Capability | Description | Security Features |
|-----------|-------------|-------------------|
| `net:outbound` | HTTP requests | Configurable timeout, URL validation |
| `fs:temp` | Temporary files | Path traversal prevention, sandboxed directory |
| `fs:read` | Read files | Sensitive file filtering |
| `fs:write` | Write files | Sensitive path filtering |
| `env:read` | Environment variables | Sensitive variable filtering |
| `secrets:read` | Decrypt secrets | Pluggable decryption function |
| `exec:micro-runner` | Execute micro-runner | (Future implementation) |

**Security Filters:**
- Sensitive files: `/etc/shadow`, `/root/.ssh/*`, etc.
- Sensitive paths: `/etc`, `/sys`, `/proc`, etc.
- Sensitive env vars: `*PASSWORD`, `*SECRET`, `*TOKEN`, etc.

### 5. Provider Registry (`registry.go`)

Manages provider lifecycle and version resolution.

**Features:**
- Provider registration from files or OCI images
- Version resolution:
  - Exact: `1.0.0`
  - Latest: `latest` or empty
  - Tilde: `~1.0.0` (matches 1.0.x)
  - Caret: `^1.0.0` (matches 1.x.x)
- Provider caching (lazy loading)
- Directory scanning for provider discovery
- Concurrent-safe operations

## Usage Examples

### Register and Load a Provider

```go
import (
    "context"
    "github.com/openfroyo/openfroyo/pkg/providers/host"
    "github.com/openfroyo/openfroyo/pkg/engine"
)

// Create registry
registry := host.NewRegistry("/opt/openfroyo/providers", &host.WASMHostConfig{
    Timeout:          30 * time.Second,
    MemoryLimitPages: 256,
    TempDir:          "/tmp/openfroyo",
})

// Set allowed capabilities
registry.SetAllowedCapabilities([]string{
    string(engine.CapabilityNetOutbound),
    string(engine.CapabilityFSTemp),
    string(engine.CapabilityExecMicroRunner),
})

// Register provider from manifest file
err := registry.RegisterFromPath(ctx, "/opt/openfroyo/providers/linux-pkg/manifest.yaml")
if err != nil {
    return fmt.Errorf("failed to register provider: %w", err)
}

// Get provider (loads if not cached)
provider, err := registry.Get(ctx, "linux-pkg", "1.0.0")
if err != nil {
    return fmt.Errorf("failed to get provider: %w", err)
}

// Initialize provider
err = provider.Init(ctx, engine.ProviderConfig{
    Name:    "linux-pkg",
    Version: "1.0.0",
    Capabilities: []string{
        string(engine.CapabilityExecMicroRunner),
    },
    Timeout: 30 * time.Second,
})
if err != nil {
    return fmt.Errorf("failed to initialize provider: %w", err)
}
```

### Execute Provider Operations

```go
// Read current state
readResp, err := provider.Read(ctx, engine.ReadRequest{
    ResourceID: "package:nginx",
    Config:     json.RawMessage(`{"name": "nginx"}`),
})
if err != nil {
    return fmt.Errorf("read failed: %w", err)
}

// Plan changes
planResp, err := provider.Plan(ctx, engine.PlanRequest{
    ResourceID:   "package:nginx",
    DesiredState: json.RawMessage(`{"name": "nginx", "state": "present"}`),
    ActualState:  readResp.State,
    Operation:    engine.OperationUpdate,
})
if err != nil {
    return fmt.Errorf("plan failed: %w", err)
}

// Apply changes
applyResp, err := provider.Apply(ctx, engine.ApplyRequest{
    ResourceID:     "package:nginx",
    DesiredState:   json.RawMessage(`{"name": "nginx", "state": "present"}`),
    ActualState:    readResp.State,
    Operation:      engine.OperationUpdate,
    PlannedChanges: planResp.Changes,
})
if err != nil {
    return fmt.Errorf("apply failed: %w", err)
}
```

### Cleanup

```go
// Close specific provider
wasmProvider := provider.(*host.WASMHostProvider)
err = wasmProvider.Close(ctx)

// Or close entire registry (closes all loaded providers)
err = registry.Close(ctx)
```

## Provider Development

### WASM Module Requirements

1. **Export Memory:**
   ```rust
   #[no_mangle]
   pub static memory: [u8; 65536] = [0; 65536];
   ```

2. **Implement Memory Management:**
   ```rust
   #[no_mangle]
   pub extern "C" fn malloc(size: u32) -> u32 { /* ... */ }

   #[no_mangle]
   pub extern "C" fn free(ptr: u32) { /* ... */ }
   ```

3. **Implement Provider Functions:**
   ```rust
   #[no_mangle]
   pub extern "C" fn provider_init(config_ptr: u32, config_len: u32) -> u64 {
       // Parse config JSON from memory
       // Initialize provider state
       // Return result as packed (ptr << 32) | len
   }

   #[no_mangle]
   pub extern "C" fn provider_read(req_ptr: u32, req_len: u32) -> u64 {
       // Read current resource state
       // Return JSON response
   }

   // ... implement other functions
   ```

### Host Functions Available to WASM

Providers can call these host functions for capability-based operations:

```rust
// HTTP request (net:outbound)
extern "C" {
    fn http_request(url_ptr: u32, url_len: u32, method_ptr: u32, method_len: u32) -> u64;
}

// Temp file operations (fs:temp)
extern "C" {
    fn write_temp_file(name_ptr: u32, name_len: u32, data_ptr: u32, data_len: u32) -> u32;
    fn read_temp_file(name_ptr: u32, name_len: u32) -> u64;
}

// Secret decryption (secrets:read)
extern "C" {
    fn decrypt_secret(encrypted_ptr: u32, encrypted_len: u32) -> u64;
}
```

## Testing

### Unit Tests

```bash
go test ./pkg/providers/host -v
```

### Benchmarks

```bash
go test ./pkg/providers/host -bench=. -benchmem
```

### Test Coverage

The test suite includes:
- Manifest parsing and validation
- Capability enforcement
- Registry operations (version resolution, caching)
- Sensitive file/env filtering
- Path traversal prevention
- Concurrent operations

## Security Considerations

1. **Sandboxing:** WASM modules run in a sandboxed environment with no direct access to host resources.

2. **Capability-Based:** All privileged operations require explicit capability grants.

3. **Memory Limits:** Configurable memory limits prevent resource exhaustion.

4. **Timeouts:** Operation timeouts prevent infinite loops or hangs.

5. **Input Validation:** All WASM inputs/outputs are validated and sanitized.

6. **Path Security:** Prevents path traversal and access to sensitive system files.

7. **Secret Protection:** Sensitive environment variables and secrets are filtered.

## Performance

- **Memory Overhead:** Base runtime ~2-5MB + provider module size
- **Startup Time:** Provider loading typically <100ms
- **Execution Overhead:** Minimal overhead for WASM execution (near-native speed)
- **Caching:** Loaded providers are cached to avoid repeated instantiation

## Limitations

1. **No CGO:** Uses pure Go wazero runtime (no native code execution)
2. **WASI Preview 1:** Limited WASI support (no sockets, limited file I/O)
3. **Single-threaded:** Each provider instance runs in a single thread
4. **Memory Model:** Simple linear memory model (no shared memory)

## Future Enhancements

- [ ] OCI image support for provider distribution
- [ ] Component Model support (WASI 0.2)
- [ ] Enhanced networking capabilities (socket support)
- [ ] Multi-threading support for providers
- [ ] Provider hot-reloading
- [ ] Enhanced observability (metrics, tracing)
- [ ] Provider dependency management
- [ ] Signed provider verification

## Dependencies

- **wazero** (v1.9.0+): Pure Go WebAssembly runtime
- **gopkg.in/yaml.v3**: YAML parsing
- **github.com/openfroyo/openfroyo/pkg/engine**: Engine interfaces

## License

See project LICENSE file.
