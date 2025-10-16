package config

import (
	"context"
	"fmt"
	"sync"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
)

// SchemaRegistry manages CUE schemas for validation.
type SchemaRegistry struct {
	ctx     *cue.Context
	schemas map[string]cue.Value
	mu      sync.RWMutex
}

// NewSchemaRegistry creates a new schema registry with built-in schemas.
func NewSchemaRegistry() *SchemaRegistry {
	ctx := cuecontext.New()
	sr := &SchemaRegistry{
		ctx:     ctx,
		schemas: make(map[string]cue.Value),
	}

	// Register built-in schemas
	sr.registerBuiltInSchemas()

	return sr
}

// registerBuiltInSchemas registers all built-in schemas.
func (sr *SchemaRegistry) registerBuiltInSchemas() {
	// Register resource schema
	sr.RegisterSchema("resource", builtinResourceSchema)

	// Register workspace schema
	sr.RegisterSchema("workspace", builtinWorkspaceSchema)

	// Register provider schema
	sr.RegisterSchema("provider", builtinProviderSchema)

	// Register target schema
	sr.RegisterSchema("target", builtinTargetSchema)

	// Register dependency schema
	sr.RegisterSchema("dependency", builtinDependencySchema)
}

// RegisterSchema registers a CUE schema with the given name.
func (sr *SchemaRegistry) RegisterSchema(name, schema string) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	val := sr.ctx.CompileString(schema)
	if err := val.Err(); err != nil {
		return fmt.Errorf("failed to compile schema %s: %w", name, err)
	}

	sr.schemas[name] = val
	return nil
}

// GetSchema retrieves a schema by name.
func (sr *SchemaRegistry) GetSchema(name string) (cue.Value, bool) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	val, ok := sr.schemas[name]
	return val, ok
}

// ValidateAgainstSchema validates data against a named schema.
func (sr *SchemaRegistry) ValidateAgainstSchema(ctx context.Context, schemaName string, data interface{}) error {
	schema, ok := sr.GetSchema(schemaName)
	if !ok {
		return fmt.Errorf("schema %s not found", schemaName)
	}

	// Convert data to CUE value
	dataVal := sr.ctx.Encode(data)
	if err := dataVal.Err(); err != nil {
		return fmt.Errorf("failed to encode data: %w", err)
	}

	// Unify with schema (validates)
	unified := schema.Unify(dataVal)
	if err := unified.Validate(cue.Concrete(true)); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}

// ListSchemas returns all registered schema names.
func (sr *SchemaRegistry) ListSchemas() []string {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	names := make([]string, 0, len(sr.schemas))
	for name := range sr.schemas {
		names = append(names, name)
	}
	return names
}

// Built-in schema definitions

const builtinResourceSchema = `
// Resource schema for OpenFroyo resource definitions
#Resource: {
	// ID is the unique identifier for this resource
	id: string & =~"^[a-zA-Z0-9_-]+$"

	// Type is the resource type (e.g., "linux.pkg", "linux.service")
	type: string & =~"^[a-z0-9]+\\.[a-z0-9_]+$"

	// Name is the human-readable name
	name: string

	// Config is the resource-specific configuration
	config: {...}

	// Labels are key-value pairs for organizing resources
	labels?: {[string]: string}

	// Annotations are additional metadata
	annotations?: {[string]: string}

	// Dependencies lists resource dependencies
	dependencies?: [...#Dependency]

	// Target specifies which hosts this applies to
	target?: #Target

	// Provider overrides the provider
	provider?: {
		name:    string
		version?: string
	}
}
`

const builtinWorkspaceSchema = `
// Workspace schema for OpenFroyo workspace configuration
#Workspace: {
	// Name is the workspace name
	name: string & =~"^[a-zA-Z0-9_-]+$"

	// Version is the configuration version
	version?: string

	// Providers lists the providers used
	providers?: [...#Provider]

	// Variables are workspace-level variables
	variables?: {[string]: _}

	// Backend configures state storage
	backend?: {
		type: "solo" | "cluster"
		path?: string
		config?: {...}
	}

	// Policy configures policy enforcement
	policy?: {
		enabled: bool
		paths?: [...string]
		mode?: "advisory" | "enforcing"
		on_violation?: "warn" | "fail"
	}

	// Metadata contains additional workspace metadata
	metadata?: {[string]: _}
}
`

const builtinProviderSchema = `
// Provider schema for OpenFroyo provider configuration
#Provider: {
	// Name is the provider name
	name: string & =~"^[a-z0-9]+\\.[a-z0-9_]+$"

	// Version is the provider version constraint
	version?: string

	// Source is the OCI registry URL
	source?: string

	// Config is provider-specific configuration
	config?: {...}

	// Capabilities are required capabilities
	capabilities?: [...string]
}
`

const builtinTargetSchema = `
// Target schema for target selector configuration
#Target: {
	// Hosts lists specific host IDs or patterns
	hosts?: [...string]

	// Labels matches targets with these labels
	labels?: {[string]: string}

	// Selector is a label selector expression
	selector?: string

	// All indicates this applies to all targets
	all?: bool

	// At least one targeting method must be specified
	_hasTargeting: (hosts != _|_ | labels != _|_ | selector != _|_ | all == true)
	if !_hasTargeting {
		// Force a constraint error if no targeting specified
		_error: "target must specify at least one of: hosts, labels, selector, or all"
	}
}
`

const builtinDependencySchema = `
// Dependency schema for resource dependencies
#Dependency: {
	// ResourceID is the ID of the resource this depends on
	resource_id: string & =~"^[a-zA-Z0-9_-]+$"

	// Type is the dependency type
	type: "require" | "notify" | "order"
}
`

// ValidateResource validates a resource configuration against the resource schema.
func (sr *SchemaRegistry) ValidateResource(ctx context.Context, resource ResourceConfig) error {
	return sr.ValidateAgainstSchema(ctx, "resource", resource)
}

// ValidateWorkspace validates a workspace configuration against the workspace schema.
func (sr *SchemaRegistry) ValidateWorkspace(ctx context.Context, workspace WorkspaceConfig) error {
	return sr.ValidateAgainstSchema(ctx, "workspace", workspace)
}

// ValidateProvider validates a provider configuration against the provider schema.
func (sr *SchemaRegistry) ValidateProvider(ctx context.Context, provider ProviderConfig) error {
	return sr.ValidateAgainstSchema(ctx, "provider", provider)
}

// ValidateTarget validates a target selector against the target schema.
func (sr *SchemaRegistry) ValidateTarget(ctx context.Context, target TargetSelector) error {
	return sr.ValidateAgainstSchema(ctx, "target", target)
}

// ValidateDependency validates a dependency configuration against the dependency schema.
func (sr *SchemaRegistry) ValidateDependency(ctx context.Context, dependency DependencyConfig) error {
	return sr.ValidateAgainstSchema(ctx, "dependency", dependency)
}
