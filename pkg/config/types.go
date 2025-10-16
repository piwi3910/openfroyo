package config

import (
	"encoding/json"
	"time"

	"github.com/openfroyo/openfroyo/pkg/engine"
)

// ResourceConfig represents a resource configuration from CUE.
type ResourceConfig struct {
	// ID is the unique identifier for this resource (e.g., "web_server_pkg").
	ID string `json:"id" validate:"required"`

	// Type is the resource type (e.g., "linux.pkg", "linux.service").
	Type string `json:"type" validate:"required"`

	// Name is the human-readable name.
	Name string `json:"name" validate:"required"`

	// Config is the resource-specific configuration.
	Config json.RawMessage `json:"config" validate:"required"`

	// Labels are key-value pairs for organizing and selecting resources.
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations are additional metadata.
	Annotations map[string]string `json:"annotations,omitempty"`

	// Dependencies lists the dependencies for this resource.
	Dependencies []DependencyConfig `json:"dependencies,omitempty"`

	// Target specifies which hosts/targets this resource applies to.
	Target TargetSelector `json:"target,omitempty"`

	// Provider overrides the provider name and version for this resource.
	Provider *ProviderOverride `json:"provider,omitempty"`
}

// DependencyConfig represents a dependency relationship between resources.
type DependencyConfig struct {
	// ResourceID is the ID of the resource this depends on.
	ResourceID string `json:"resource_id" validate:"required"`

	// Type is the dependency type (require, notify, order).
	Type engine.DependencyType `json:"type" validate:"required,oneof=require notify order"`
}

// ProviderOverride allows overriding provider details for a specific resource.
type ProviderOverride struct {
	// Name is the provider name (e.g., "linux.pkg").
	Name string `json:"name" validate:"required"`

	// Version is the provider version constraint (e.g., ">=1.0.0").
	Version string `json:"version,omitempty"`
}

// TargetSelector specifies which targets a resource applies to.
type TargetSelector struct {
	// Hosts lists specific host IDs or patterns.
	Hosts []string `json:"hosts,omitempty"`

	// Labels matches targets with these labels.
	Labels map[string]string `json:"labels,omitempty"`

	// Selector is a label selector expression (e.g., "env=prod,role=web").
	Selector string `json:"selector,omitempty"`

	// All indicates this resource applies to all targets.
	All bool `json:"all,omitempty"`
}

// ProviderConfig represents provider configuration from CUE.
type ProviderConfig struct {
	// Name is the provider name (e.g., "linux.pkg").
	Name string `json:"name" validate:"required"`

	// Version is the provider version or constraint.
	Version string `json:"version,omitempty"`

	// Source is where to fetch the provider (OCI registry URL).
	Source string `json:"source,omitempty"`

	// Config is provider-specific configuration.
	Config json.RawMessage `json:"config,omitempty"`

	// Capabilities are the capabilities this provider requires.
	Capabilities []string `json:"capabilities,omitempty"`
}

// WorkspaceConfig represents the workspace configuration.
type WorkspaceConfig struct {
	// Name is the workspace name.
	Name string `json:"name" validate:"required"`

	// Version is the configuration version.
	Version string `json:"version,omitempty"`

	// Providers lists the providers used in this workspace.
	Providers []ProviderConfig `json:"providers,omitempty"`

	// Variables are workspace-level variables.
	Variables map[string]interface{} `json:"variables,omitempty"`

	// Backend configures state storage.
	Backend *BackendConfig `json:"backend,omitempty"`

	// Policy configures policy enforcement.
	Policy *PolicyConfig `json:"policy,omitempty"`

	// Metadata contains additional workspace metadata.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// BackendConfig configures state storage backend.
type BackendConfig struct {
	// Type is the backend type (solo, cluster).
	Type string `json:"type" validate:"required,oneof=solo cluster"`

	// Path is the local path for solo backend.
	Path string `json:"path,omitempty"`

	// Config is backend-specific configuration.
	Config json.RawMessage `json:"config,omitempty"`
}

// PolicyConfig configures policy enforcement.
type PolicyConfig struct {
	// Enabled indicates if policy enforcement is enabled.
	Enabled bool `json:"enabled"`

	// Paths lists policy file paths.
	Paths []string `json:"paths,omitempty"`

	// Mode is the enforcement mode (advisory, enforcing).
	Mode string `json:"mode,omitempty" validate:"omitempty,oneof=advisory enforcing"`

	// OnViolation specifies the action on violation (warn, fail).
	OnViolation string `json:"on_violation,omitempty" validate:"omitempty,oneof=warn fail"`
}

// ParsedConfig represents the fully parsed configuration from CUE.
type ParsedConfig struct {
	// Workspace is the workspace configuration.
	Workspace WorkspaceConfig `json:"workspace"`

	// Resources are all resources defined in the configuration.
	Resources []ResourceConfig `json:"resources"`

	// SourceFiles are the CUE files that were parsed.
	SourceFiles []string `json:"source_files"`

	// ParsedAt is when the configuration was parsed.
	ParsedAt time.Time `json:"parsed_at"`

	// Errors lists any validation errors.
	Errors []ValidationError `json:"errors,omitempty"`
}

// ValidationError represents a validation error with location information.
type ValidationError struct {
	// File is the source file path.
	File string `json:"file,omitempty"`

	// Line is the line number (1-indexed).
	Line int `json:"line,omitempty"`

	// Column is the column number (1-indexed).
	Column int `json:"column,omitempty"`

	// Path is the CUE path to the error (e.g., "resources.web_server.config").
	Path string `json:"path,omitempty"`

	// Message is the error message.
	Message string `json:"message"`

	// Severity is the error severity (error, warning, info).
	Severity string `json:"severity" validate:"required,oneof=error warning info"`
}

// ConfigSource represents a source of CUE configuration.
type ConfigSource struct {
	// Type is the source type (file, directory, inline).
	Type string `json:"type" validate:"required,oneof=file directory inline"`

	// Path is the file or directory path.
	Path string `json:"path,omitempty"`

	// Content is the inline CUE content.
	Content string `json:"content,omitempty"`
}

// MergeOptions controls how multiple configurations are merged.
type MergeOptions struct {
	// AllowConflicts allows conflicting values (last wins).
	AllowConflicts bool `json:"allow_conflicts"`

	// IncludePaths filters which paths to merge.
	IncludePaths []string `json:"include_paths,omitempty"`

	// ExcludePaths filters which paths to exclude from merge.
	ExcludePaths []string `json:"exclude_paths,omitempty"`
}

// EvaluateOptions controls CUE evaluation behavior.
type EvaluateOptions struct {
	// Package is the CUE package to evaluate.
	Package string `json:"package,omitempty"`

	// Tags are CUE build tags (e.g., "env=prod").
	Tags []string `json:"tags,omitempty"`

	// Concrete requires all values to be concrete (no unresolved references).
	Concrete bool `json:"concrete"`

	// ValidateSchemas enables schema validation during evaluation.
	ValidateSchemas bool `json:"validate_schemas"`

	// AllowStarlark enables Starlark function execution.
	AllowStarlark bool `json:"allow_starlark"`

	// StarlarkTimeout is the timeout for Starlark execution.
	StarlarkTimeout time.Duration `json:"starlark_timeout,omitempty"`
}

// StarlarkContext provides context for Starlark execution.
type StarlarkContext struct {
	// Input is the input data passed to Starlark.
	Input map[string]interface{} `json:"input,omitempty"`

	// Timeout is the execution timeout.
	Timeout time.Duration `json:"timeout"`

	// AllowedModules lists allowed Starlark modules.
	AllowedModules []string `json:"allowed_modules,omitempty"`

	// Builtins are additional built-in functions to provide.
	Builtins map[string]interface{} `json:"builtins,omitempty"`
}

// StarlarkResult represents the result of Starlark execution.
type StarlarkResult struct {
	// Output is the output data from Starlark.
	Output map[string]interface{} `json:"output,omitempty"`

	// ExecutionTime is how long the script took to execute.
	ExecutionTime time.Duration `json:"execution_time"`

	// Error is any error that occurred.
	Error string `json:"error,omitempty"`
}

// ToEngineConfig converts ParsedConfig to engine.Config.
func (pc *ParsedConfig) ToEngineConfig() *engine.Config {
	resources := make([]engine.Resource, len(pc.Resources))
	for i, rc := range pc.Resources {
		resources[i] = engine.Resource{
			ID:           rc.ID,
			Type:         rc.Type,
			Name:         rc.Name,
			Config:       rc.Config,
			Labels:       rc.Labels,
			Annotations:  rc.Annotations,
			Dependencies: toDependencyIDs(rc.Dependencies),
			Status:       engine.ResourceStatusUnknown,
			CreatedAt:    pc.ParsedAt,
			UpdatedAt:    pc.ParsedAt,
			Version:      1,
		}
	}

	return &engine.Config{
		ID:        pc.Workspace.Name,
		Source:    formatSourceFiles(pc.SourceFiles),
		ParsedAt:  pc.ParsedAt,
		Resources: resources,
		Variables: pc.Workspace.Variables,
		Metadata:  pc.Workspace.Metadata,
	}
}

// toDependencyIDs converts DependencyConfig slice to string slice.
func toDependencyIDs(deps []DependencyConfig) []string {
	ids := make([]string, len(deps))
	for i, dep := range deps {
		ids[i] = dep.ResourceID
	}
	return ids
}

// formatSourceFiles formats source files for display.
func formatSourceFiles(files []string) string {
	if len(files) == 0 {
		return "inline"
	}
	if len(files) == 1 {
		return files[0]
	}
	return files[0] + " (+" + string(rune(len(files)-1)) + " more)"
}
