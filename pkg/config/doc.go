// Package config provides CUE configuration parsing and Starlark evaluation
// for OpenFroyo infrastructure orchestration.
//
// # Overview
//
// The config package implements the configuration evaluation phase of OpenFroyo,
// responsible for parsing CUE files, validating schemas, and executing Starlark
// scripts for procedural logic.
//
// # Features
//
//   - CUE configuration parsing from files, directories, and inline content
//   - Schema validation with built-in schemas for resources, providers, and targets
//   - Starlark script execution for procedural configuration logic
//   - Type-safe configuration structures
//   - Error reporting with file locations and line numbers
//   - Configuration merging from multiple sources
//
// # Components
//
// CUEParser: Main parser for CUE configuration files. Implements the engine.Evaluator
// interface for integration with the OpenFroyo engine.
//
// SchemaRegistry: Manages CUE schemas for validation. Provides built-in schemas
// for common configuration patterns and supports custom schema registration.
//
// StarlarkEvaluator: Safe Starlark script execution with timeout enforcement and
// sandboxing. Provides built-in functions and type conversion between Go and Starlark.
//
// # Usage Example
//
//	// Create a new parser
//	parser := config.NewCUEParser()
//
//	// Parse configuration files
//	cfg, err := parser.Evaluate(ctx, []string{"config.cue", "resources.cue"})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Validate against schemas
//	if err := parser.Validate(ctx, cfg); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Execute Starlark for procedural logic
//	input := map[string]interface{}{"count": 5}
//	output, err := parser.EvaluateStarlark(ctx, script, input)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # CUE Configuration Structure
//
// OpenFroyo uses CUE to define infrastructure resources with strong typing
// and validation. A typical configuration includes:
//
//	workspace: {
//	    name: "my-infrastructure"
//	    version: "1.0"
//	    providers: [
//	        {name: "linux.pkg", version: ">=1.0.0"},
//	        {name: "linux.service", version: ">=1.0.0"},
//	    ]
//	}
//
//	resources: {
//	    web_server: {
//	        type: "linux.pkg"
//	        name: "nginx"
//	        config: {
//	            package: "nginx"
//	            state: "present"
//	            version: "latest"
//	        }
//	        target: {
//	            labels: {env: "prod", role: "web"}
//	        }
//	    }
//	}
//
// # Starlark Integration
//
// Starlark scripts can be embedded in CUE configurations for procedural logic:
//
//	# Generate multiple resources programmatically
//	def generate_servers(count):
//	    servers = []
//	    for i in range(count):
//	        servers.append({
//	            "id": "server_" + str(i),
//	            "name": "server-" + str(i),
//	        })
//	    return servers
//
// # Schema Validation
//
// Built-in schemas enforce configuration correctness:
//
//   - Resource schema: Validates resource definitions with required fields
//   - Workspace schema: Validates workspace configuration
//   - Provider schema: Validates provider declarations
//   - Target schema: Validates target selectors
//   - Dependency schema: Validates resource dependencies
//
// Custom schemas can be registered for domain-specific validation.
//
// # Error Handling
//
// All parsing and validation errors include detailed location information:
//
//	ValidationError{
//	    File: "config.cue",
//	    Line: 42,
//	    Column: 5,
//	    Path: "resources.web_server.config",
//	    Message: "field 'package' is required",
//	    Severity: "error",
//	}
//
// # Security
//
// Starlark execution is sandboxed:
//   - No filesystem access
//   - No network access
//   - Timeout enforcement (default 30 seconds)
//   - Print statements suppressed
//   - Only safe built-in functions provided
//
// # Thread Safety
//
// All types in this package are safe for concurrent use.
package config
