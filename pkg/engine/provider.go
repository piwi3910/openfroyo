package engine

import (
	"context"
	"encoding/json"
	"time"
)

// Provider is the interface that all resource providers must implement.
// This matches the WASM provider contract defined in the MVP spec.
type Provider interface {
	// Init initializes the provider with configuration.
	// This is called once when the provider is loaded.
	Init(ctx context.Context, config ProviderConfig) error

	// Read retrieves the current state of a resource.
	// Returns the actual state or an error.
	Read(ctx context.Context, req ReadRequest) (*ReadResponse, error)

	// Plan computes the operations needed to reach desired state.
	// Returns the planned operations and changes.
	Plan(ctx context.Context, req PlanRequest) (*PlanResponse, error)

	// Apply executes the planned operations to achieve desired state.
	// Returns the result of the operation and new state.
	Apply(ctx context.Context, req ApplyRequest) (*ApplyResponse, error)

	// Destroy removes the resource completely.
	// Returns the result of the destruction.
	Destroy(ctx context.Context, req DestroyRequest) (*DestroyResponse, error)

	// Validate validates a resource configuration against the provider's schema.
	// Returns validation errors if any.
	Validate(ctx context.Context, config json.RawMessage) error

	// Schema returns the JSON schema for this provider's resources.
	Schema() (*ProviderSchema, error)

	// Metadata returns information about this provider.
	Metadata() ProviderMetadata
}

// ProviderConfig contains provider initialization configuration.
type ProviderConfig struct {
	// Name is the name of the provider.
	Name string `json:"name"`

	// Version is the version of the provider.
	Version string `json:"version"`

	// Config is provider-specific configuration.
	Config json.RawMessage `json:"config,omitempty"`

	// Capabilities are the capabilities granted to this provider.
	Capabilities []string `json:"capabilities,omitempty"`

	// WorkDir is the working directory for the provider.
	WorkDir string `json:"work_dir,omitempty"`

	// Timeout is the default timeout for provider operations.
	Timeout time.Duration `json:"timeout"`
}

// ReadRequest contains the parameters for a Read operation.
type ReadRequest struct {
	// ResourceID is the unique identifier of the resource.
	ResourceID string `json:"resource_id"`

	// Config is the current configuration of the resource.
	Config json.RawMessage `json:"config"`

	// Metadata contains additional request metadata.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ReadResponse contains the result of a Read operation.
type ReadResponse struct {
	// State is the current actual state of the resource.
	State json.RawMessage `json:"state"`

	// Exists indicates whether the resource exists.
	Exists bool `json:"exists"`

	// Metadata contains additional response metadata.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// PlanRequest contains the parameters for a Plan operation.
type PlanRequest struct {
	// ResourceID is the unique identifier of the resource.
	ResourceID string `json:"resource_id"`

	// DesiredState is the desired configuration.
	DesiredState json.RawMessage `json:"desired_state"`

	// ActualState is the current state (from Read).
	ActualState json.RawMessage `json:"actual_state,omitempty"`

	// Operation is the requested operation type.
	Operation OperationType `json:"operation"`

	// Metadata contains additional request metadata.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// PlanResponse contains the result of a Plan operation.
type PlanResponse struct {
	// Operation is the determined operation to perform.
	Operation OperationType `json:"operation"`

	// Changes lists the changes that will be made.
	Changes []Change `json:"changes"`

	// RequiresRecreate indicates if the resource must be recreated.
	RequiresRecreate bool `json:"requires_recreate"`

	// Warnings are non-fatal warnings about the plan.
	Warnings []string `json:"warnings,omitempty"`

	// Metadata contains additional response metadata.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ApplyRequest contains the parameters for an Apply operation.
type ApplyRequest struct {
	// ResourceID is the unique identifier of the resource.
	ResourceID string `json:"resource_id"`

	// DesiredState is the desired configuration.
	DesiredState json.RawMessage `json:"desired_state"`

	// ActualState is the current state before the operation.
	ActualState json.RawMessage `json:"actual_state,omitempty"`

	// Operation is the operation to perform.
	Operation OperationType `json:"operation"`

	// PlannedChanges are the changes from the Plan phase.
	PlannedChanges []Change `json:"planned_changes,omitempty"`

	// Metadata contains additional request metadata.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ApplyResponse contains the result of an Apply operation.
type ApplyResponse struct {
	// NewState is the resulting state after the operation.
	NewState json.RawMessage `json:"new_state"`

	// Output contains any output data from the operation.
	Output json.RawMessage `json:"output,omitempty"`

	// Events are events that occurred during the operation.
	Events []ProviderEvent `json:"events,omitempty"`

	// Metadata contains additional response metadata.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// DestroyRequest contains the parameters for a Destroy operation.
type DestroyRequest struct {
	// ResourceID is the unique identifier of the resource.
	ResourceID string `json:"resource_id"`

	// State is the current state of the resource.
	State json.RawMessage `json:"state"`

	// Metadata contains additional request metadata.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// DestroyResponse contains the result of a Destroy operation.
type DestroyResponse struct {
	// Success indicates whether the destruction was successful.
	Success bool `json:"success"`

	// Events are events that occurred during the operation.
	Events []ProviderEvent `json:"events,omitempty"`

	// Metadata contains additional response metadata.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ProviderEvent represents an event emitted by a provider during execution.
type ProviderEvent struct {
	// Timestamp is when the event occurred.
	Timestamp time.Time `json:"timestamp"`

	// Type is the event type.
	Type string `json:"type"`

	// Message is a human-readable message.
	Message string `json:"message"`

	// Data contains event-specific data.
	Data map[string]interface{} `json:"data,omitempty"`
}

// ProviderSchema defines the schema for a provider's resources.
type ProviderSchema struct {
	// Version is the schema version.
	Version string `json:"version"`

	// ResourceTypes maps resource type names to their schemas.
	ResourceTypes map[string]*ResourceTypeSchema `json:"resource_types"`

	// DataSources maps data source names to their schemas.
	DataSources map[string]*DataSourceSchema `json:"data_sources,omitempty"`
}

// ResourceTypeSchema defines the schema for a resource type.
type ResourceTypeSchema struct {
	// Name is the resource type name.
	Name string `json:"name"`

	// Description describes what this resource type manages.
	Description string `json:"description"`

	// ConfigSchema is the JSON schema for resource configuration.
	ConfigSchema json.RawMessage `json:"config_schema"`

	// StateSchema is the JSON schema for resource state.
	StateSchema json.RawMessage `json:"state_schema"`

	// Capabilities lists the capabilities required by this resource type.
	Capabilities []string `json:"capabilities,omitempty"`
}

// DataSourceSchema defines the schema for a data source.
type DataSourceSchema struct {
	// Name is the data source name.
	Name string `json:"name"`

	// Description describes what data this source provides.
	Description string `json:"description"`

	// ConfigSchema is the JSON schema for data source configuration.
	ConfigSchema json.RawMessage `json:"config_schema"`

	// OutputSchema is the JSON schema for data source output.
	OutputSchema json.RawMessage `json:"output_schema"`
}

// ProviderMetadata contains information about a provider.
type ProviderMetadata struct {
	// Name is the provider name.
	Name string `json:"name"`

	// Version is the provider version.
	Version string `json:"version"`

	// Description describes what this provider does.
	Description string `json:"description"`

	// Author is the provider author/maintainer.
	Author string `json:"author"`

	// License is the provider license.
	License string `json:"license"`

	// Repository is the source repository URL.
	Repository string `json:"repository,omitempty"`

	// Homepage is the provider homepage URL.
	Homepage string `json:"homepage,omitempty"`

	// RequiredCapabilities lists capabilities this provider needs.
	RequiredCapabilities []string `json:"required_capabilities,omitempty"`
}

// ProviderCapability represents a capability that can be granted to providers.
type ProviderCapability string

const (
	// CapabilityNetOutbound allows outbound network connections.
	CapabilityNetOutbound ProviderCapability = "net:outbound"

	// CapabilityNetInbound allows inbound network connections.
	CapabilityNetInbound ProviderCapability = "net:inbound"

	// CapabilityFSTemp allows temporary file system access.
	CapabilityFSTemp ProviderCapability = "fs:temp"

	// CapabilityFSRead allows read-only file system access.
	CapabilityFSRead ProviderCapability = "fs:read"

	// CapabilityFSWrite allows write file system access.
	CapabilityFSWrite ProviderCapability = "fs:write"

	// CapabilityExecMicroRunner allows delegating to the micro-runner.
	CapabilityExecMicroRunner ProviderCapability = "exec:micro-runner"

	// CapabilityEnvRead allows reading environment variables.
	CapabilityEnvRead ProviderCapability = "env:read"

	// CapabilitySecretsRead allows reading secrets.
	CapabilitySecretsRead ProviderCapability = "secrets:read"
)

// ProviderManifest represents the manifest file for a provider plugin.
type ProviderManifest struct {
	// Metadata is provider metadata.
	Metadata ProviderMetadata `json:"metadata"`

	// Schema is the provider schema.
	Schema *ProviderSchema `json:"schema"`

	// Entrypoint is the WASM module entrypoint.
	Entrypoint string `json:"entrypoint"`

	// Checksum is the SHA256 checksum of the WASM module.
	Checksum string `json:"checksum"`

	// SBOM is the software bill of materials.
	SBOM json.RawMessage `json:"sbom,omitempty"`
}
