# OpenFroyo Config Package

The `config` package provides CUE configuration parsing and Starlark evaluation for OpenFroyo infrastructure orchestration.

## Features

- **CUE Configuration Parsing**: Load and parse CUE files with full type checking and validation
- **Schema Validation**: Built-in schemas for resources, providers, targets, and dependencies
- **Starlark Integration**: Safe execution of Starlark scripts for procedural logic
- **Error Reporting**: Detailed error messages with file locations and line numbers
- **Configuration Merging**: Combine multiple CUE files into a unified configuration
- **Type Safety**: Strongly typed configuration structures with validation

## Architecture

### Core Components

1. **CUEParser** (`cue_parser.go`)
   - Parses CUE files, directories, and inline content
   - Implements the `engine.Evaluator` interface
   - Extracts resources, dependencies, and workspace configuration
   - Provides detailed validation errors with line numbers

2. **SchemaRegistry** (`schemas.go`)
   - Manages CUE schemas for validation
   - Provides built-in schemas for common patterns
   - Supports custom schema registration
   - Thread-safe schema access

3. **StarlarkEvaluator** (`starlark_eval.go`)
   - Executes Starlark scripts with timeout enforcement
   - Sandboxed execution (no filesystem or network access)
   - Type conversion between Go and Starlark values
   - Built-in helper functions (range, enumerate, zip)

4. **Configuration Types** (`types.go`)
   - Strongly typed configuration structures
   - Validation using struct tags
   - Conversion to engine types

## Usage

### Basic CUE Parsing

```go
package main

import (
    "context"
    "log"

    "github.com/openfroyo/openfroyo/pkg/config"
)

func main() {
    parser := config.NewCUEParser()
    ctx := context.Background()

    // Parse configuration files
    cfg, err := parser.Evaluate(ctx, []string{"config.cue", "resources.cue"})
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Loaded %d resources", len(cfg.Resources))
}
```

### Inline CUE Parsing

```go
content := `
workspace: {
    name: "my-infrastructure"
    version: "1.0"
}

resources: {
    web_server: {
        id: "web"
        type: "linux.pkg"
        name: "nginx"
        config: {
            package: "nginx"
            state: "present"
        }
    }
}
`

parsedConfig, err := parser.ParseInline(ctx, content)
if err != nil {
    log.Fatal(err)
}
```

### Starlark Evaluation

```go
evaluator := config.NewStarlarkEvaluator(30 * time.Second)

script := `
def generate_servers(count):
    servers = []
    for i in range(count):
        servers.append({
            "id": "server_" + str(i),
            "name": "server-" + str(i),
            "port": 8000 + i,
        })
    return servers

result = generate_servers(3)
`

input := map[string]interface{}{
    "count": 3,
}

result, err := evaluator.Evaluate(ctx, script, input)
if err != nil {
    log.Fatal(err)
}

log.Printf("Generated %d servers", len(result.Output["result"].([]interface{})))
```

### Schema Validation

```go
registry := config.NewSchemaRegistry()

resource := config.ResourceConfig{
    ID:   "test",
    Type: "linux.pkg",
    Name: "nginx",
    Config: []byte(`{"package":"nginx","state":"present"}`),
}

err := registry.ValidateResource(ctx, resource)
if err != nil {
    log.Printf("Validation failed: %v", err)
}
```

## CUE Configuration Examples

### Simple Web Server

```cue
workspace: {
    name: "webserver"
    version: "1.0"

    providers: [
        {name: "linux.pkg", version: ">=1.0.0"},
        {name: "linux.service", version: ">=1.0.0"},
    ]
}

resources: {
    nginx_pkg: {
        type: "linux.pkg"
        name: "nginx"
        config: {
            package: "nginx"
            state: "present"
        }
        target: {
            labels: {env: "prod", role: "web"}
        }
    }

    nginx_service: {
        type: "linux.service"
        name: "nginx"
        config: {
            name: "nginx"
            state: "running"
            enabled: true
        }
        dependencies: [
            {resource_id: "nginx_pkg", type: "require"}
        ]
    }
}
```

### Multi-Environment Configuration

```cue
#Environment: {
    name: string
    db_host: string
    replicas: int
}

environments: {
    dev: #Environment & {
        name: "development"
        db_host: "localhost"
        replicas: 1
    }

    prod: #Environment & {
        name: "production"
        db_host: "prod-db.internal"
        replicas: 5
    }
}

selectedEnv: environments.dev
```

## Dependency Types

OpenFroyo supports three types of dependencies:

1. **require**: Hard dependency - target must succeed for dependent to run
2. **notify**: Soft dependency - triggers handlers when target changes
3. **order**: Ordering dependency - ensures execution order without requiring success

Example:

```cue
resources: {
    pkg: {
        id: "pkg"
        type: "linux.pkg"
        // ...
    }

    config: {
        id: "config"
        type: "linux.file"
        dependencies: [
            {resource_id: "pkg", type: "require"}
        ]
    }

    service: {
        id: "service"
        type: "linux.service"
        dependencies: [
            {resource_id: "pkg", type: "require"},
            {resource_id: "config", type: "notify"}  // Restart on config change
        ]
    }
}
```

## Target Selectors

Resources can target specific hosts using multiple methods:

```cue
// Target by labels
target: {
    labels: {env: "prod", role: "web"}
}

// Target specific hosts
target: {
    hosts: ["host1", "host2", "host3"]
}

// Target by selector expression
target: {
    selector: "env=prod,role=web"
}

// Target all hosts
target: {
    all: true
}
```

## Built-in Schemas

The schema registry provides validation for:

- **resource**: Resource definitions with type checking
- **workspace**: Workspace configuration
- **provider**: Provider declarations
- **target**: Target selectors
- **dependency**: Resource dependencies

## Error Handling

All parsing and validation errors include detailed location information:

```go
type ValidationError struct {
    File     string // Source file path
    Line     int    // Line number (1-indexed)
    Column   int    // Column number (1-indexed)
    Path     string // CUE path (e.g., "resources.web_server.config")
    Message  string // Error message
    Severity string // "error", "warning", or "info"
}
```

## Security

### Starlark Sandboxing

Starlark execution is sandboxed for security:

- No filesystem access
- No network access
- Timeout enforcement (default 30 seconds)
- Print statements suppressed
- Only safe built-in functions provided

### CUE Validation

- Schema validation enforces type safety
- Constraint checking prevents invalid configurations
- Circular imports detected and prevented

## Testing

Run the test suite:

```bash
# All tests
go test ./pkg/config/...

# Specific test
go test ./pkg/config/ -run TestCUEParser_ParseInline

# Verbose output
go test ./pkg/config/... -v
```

## Examples

Complete examples are available in:

- `examples/configs/simple_webserver.cue`: Basic Apache setup
- `examples/configs/multi_environment.cue`: Environment-specific configs
- `examples/configs/with_starlark.cue`: Starlark integration
- `examples/configs/complex_dependencies.cue`: Dependency management

## Thread Safety

All types in this package are safe for concurrent use:

- `CUEParser`: Thread-safe CUE parsing
- `SchemaRegistry`: Thread-safe schema access with mutex protection
- `StarlarkEvaluator`: Thread-safe script execution

## Performance Considerations

- **Caching**: Parsed schemas are cached in the registry
- **Incremental Parsing**: Load only changed files
- **Parallel Processing**: Independent resources can be parsed concurrently
- **Memory Management**: Large configurations use streaming where possible

## Integration with OpenFroyo Engine

The `CUEParser` implements the `engine.Evaluator` interface:

```go
type Evaluator interface {
    Evaluate(ctx context.Context, sources []string) (*Config, error)
    Validate(ctx context.Context, config *Config) error
    EvaluateStarlark(ctx context.Context, script string, input map[string]interface{}) (map[string]interface{}, error)
    MergeConfigs(ctx context.Context, configs []*Config) (*Config, error)
}
```

This allows seamless integration with the OpenFroyo execution pipeline.

## Contributing

When adding new features:

1. Update type definitions in `types.go`
2. Add schema definitions in `schemas.go`
3. Update parser logic in `cue_parser.go`
4. Add comprehensive tests
5. Update documentation

## License

Part of the OpenFroyo project. See main project LICENSE for details.
