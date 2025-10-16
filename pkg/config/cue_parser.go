package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/errors"
	"cuelang.org/go/cue/load"
	"github.com/go-playground/validator/v10"
	"github.com/openfroyo/openfroyo/pkg/engine"
)

// CUEParser parses and validates CUE configuration files.
type CUEParser struct {
	ctx              *cue.Context
	schemaRegistry   *SchemaRegistry
	starlarkEvaluator *StarlarkEvaluator
	validator        *validator.Validate
}

// NewCUEParser creates a new CUE parser.
func NewCUEParser() *CUEParser {
	return &CUEParser{
		ctx:              cuecontext.New(),
		schemaRegistry:   NewSchemaRegistry(),
		starlarkEvaluator: NewStarlarkEvaluator(30 * time.Second),
		validator:        validator.New(),
	}
}

// Evaluate parses CUE configuration files and returns the parsed configuration.
// This implements the engine.Evaluator interface.
func (cp *CUEParser) Evaluate(ctx context.Context, sources []string) (*engine.Config, error) {
	// Parse all sources
	parsedConfig, err := cp.Parse(ctx, sources)
	if err != nil {
		return nil, err
	}

	// Check for validation errors
	if len(parsedConfig.Errors) > 0 {
		return nil, fmt.Errorf("validation errors: %v", parsedConfig.Errors)
	}

	// Convert to engine.Config
	return parsedConfig.ToEngineConfig(), nil
}

// Validate validates a configuration against schemas and policies.
// This implements the engine.Evaluator interface.
func (cp *CUEParser) Validate(ctx context.Context, config *engine.Config) error {
	// Validate each resource
	for _, resource := range config.Resources {
		rc := ResourceConfig{
			ID:           resource.ID,
			Type:         resource.Type,
			Name:         resource.Name,
			Config:       resource.Config,
			Labels:       resource.Labels,
			Annotations:  resource.Annotations,
		}

		if err := cp.validator.Struct(rc); err != nil {
			return fmt.Errorf("resource %s validation failed: %w", resource.ID, err)
		}
	}

	return nil
}

// EvaluateStarlark executes Starlark scripts for procedural logic.
// This implements the engine.Evaluator interface.
func (cp *CUEParser) EvaluateStarlark(ctx context.Context, script string, input map[string]interface{}) (map[string]interface{}, error) {
	result, err := cp.starlarkEvaluator.Evaluate(ctx, script, input)
	if err != nil {
		return nil, err
	}

	if result.Error != "" {
		return nil, fmt.Errorf("starlark error: %s", result.Error)
	}

	return result.Output, nil
}

// MergeConfigs merges multiple configurations into a single configuration.
// This implements the engine.Evaluator interface.
func (cp *CUEParser) MergeConfigs(ctx context.Context, configs []*engine.Config) (*engine.Config, error) {
	if len(configs) == 0 {
		return nil, fmt.Errorf("no configs to merge")
	}

	if len(configs) == 1 {
		return configs[0], nil
	}

	// Start with the first config
	merged := &engine.Config{
		ID:        configs[0].ID,
		Source:    "merged",
		ParsedAt:  time.Now(),
		Resources: make([]engine.Resource, 0),
		Variables: make(map[string]interface{}),
		Metadata:  make(map[string]interface{}),
	}

	// Track resources by ID to detect duplicates
	resourceMap := make(map[string]engine.Resource)

	// Merge all configs
	for _, cfg := range configs {
		// Merge resources
		for _, res := range cfg.Resources {
			if existing, exists := resourceMap[res.ID]; exists {
				return nil, fmt.Errorf("duplicate resource ID %s in configs %s and %s", res.ID, existing.Name, res.Name)
			}
			resourceMap[res.ID] = res
		}

		// Merge variables
		for k, v := range cfg.Variables {
			merged.Variables[k] = v
		}

		// Merge metadata
		for k, v := range cfg.Metadata {
			merged.Metadata[k] = v
		}
	}

	// Convert map to slice
	for _, res := range resourceMap {
		merged.Resources = append(merged.Resources, res)
	}

	return merged, nil
}

// Parse parses CUE configuration from the given sources.
func (cp *CUEParser) Parse(ctx context.Context, sources []string) (*ParsedConfig, error) {
	if len(sources) == 0 {
		return nil, fmt.Errorf("no sources provided")
	}

	var cueValue cue.Value
	var sourceFiles []string
	var parseErrors []ValidationError

	// Determine if sources are files or directories
	for _, source := range sources {
		info, err := os.Stat(source)
		if err != nil {
			return nil, fmt.Errorf("failed to stat source %s: %w", source, err)
		}

		if info.IsDir() {
			// Load directory as CUE package
			val, files, errs := cp.loadDirectory(source)
			if len(errs) > 0 {
				parseErrors = append(parseErrors, errs...)
			}
			if val.Exists() {
				if cueValue.Exists() {
					cueValue = cueValue.Unify(val)
				} else {
					cueValue = val
				}
			}
			sourceFiles = append(sourceFiles, files...)
		} else {
			// Load single file
			val, errs := cp.loadFile(source)
			if len(errs) > 0 {
				parseErrors = append(parseErrors, errs...)
			}
			if val.Exists() {
				if cueValue.Exists() {
					cueValue = cueValue.Unify(val)
				} else {
					cueValue = val
				}
			}
			sourceFiles = append(sourceFiles, source)
		}
	}

	// Check for parse errors
	if len(parseErrors) > 0 {
		return &ParsedConfig{
			SourceFiles: sourceFiles,
			ParsedAt:    time.Now(),
			Errors:      parseErrors,
		}, nil
	}

	// Validate the unified value
	if err := cueValue.Err(); err != nil {
		parseErrors = append(parseErrors, cp.convertCUEErrors(err)...)
		return &ParsedConfig{
			SourceFiles: sourceFiles,
			ParsedAt:    time.Now(),
			Errors:      parseErrors,
		}, nil
	}

	// Extract configuration
	parsedConfig, err := cp.extractConfig(cueValue, sourceFiles)
	if err != nil {
		return nil, fmt.Errorf("failed to extract config: %w", err)
	}

	return parsedConfig, nil
}

// loadDirectory loads a directory as a CUE package.
func (cp *CUEParser) loadDirectory(dir string) (cue.Value, []string, []ValidationError) {
	// Load the package
	buildInstances := load.Instances([]string{dir}, nil)
	if len(buildInstances) == 0 {
		return cue.Value{}, nil, []ValidationError{{
			File:     dir,
			Message:  "no CUE files found",
			Severity: "error",
		}}
	}

	inst := buildInstances[0]
	if inst.Err != nil {
		return cue.Value{}, nil, cp.convertCUEErrors(inst.Err)
	}

	val := cp.ctx.BuildInstance(inst)
	if err := val.Err(); err != nil {
		return cue.Value{}, nil, cp.convertCUEErrors(err)
	}

	// Get list of files
	var files []string
	for _, file := range inst.Files {
		if file.Filename != "" {
			files = append(files, file.Filename)
		}
	}

	return val, files, nil
}

// loadFile loads a single CUE file.
func (cp *CUEParser) loadFile(path string) (cue.Value, []ValidationError) {
	content, err := os.ReadFile(path)
	if err != nil {
		return cue.Value{}, []ValidationError{{
			File:     path,
			Message:  fmt.Sprintf("failed to read file: %v", err),
			Severity: "error",
		}}
	}

	val := cp.ctx.CompileString(string(content), cue.Filename(path))
	if err := val.Err(); err != nil {
		return cue.Value{}, cp.convertCUEErrors(err)
	}

	return val, nil
}

// extractConfig extracts the configuration from a CUE value.
func (cp *CUEParser) extractConfig(val cue.Value, sourceFiles []string) (*ParsedConfig, error) {
	parsedConfig := &ParsedConfig{
		SourceFiles: sourceFiles,
		ParsedAt:    time.Now(),
	}

	// Extract workspace configuration
	workspaceVal := val.LookupPath(cue.ParsePath("workspace"))
	if workspaceVal.Exists() {
		var workspace WorkspaceConfig
		if err := workspaceVal.Decode(&workspace); err != nil {
			parsedConfig.Errors = append(parsedConfig.Errors, ValidationError{
				Path:     "workspace",
				Message:  fmt.Sprintf("failed to decode workspace: %v", err),
				Severity: "error",
			})
		} else {
			parsedConfig.Workspace = workspace
		}
	}

	// Extract resources
	resourcesVal := val.LookupPath(cue.ParsePath("resources"))
	if resourcesVal.Exists() {
		// Resources can be either a map or a list
		if resourcesVal.Kind() == cue.StructKind {
			// Map of resources
			iter, err := resourcesVal.Fields(cue.All())
			if err != nil {
				parsedConfig.Errors = append(parsedConfig.Errors, ValidationError{
					Path:     "resources",
					Message:  fmt.Sprintf("failed to iterate resources: %v", err),
					Severity: "error",
				})
			} else {
				for iter.Next() {
					resource, err := cp.extractResource(iter.Selector().String(), iter.Value())
					if err != nil {
						parsedConfig.Errors = append(parsedConfig.Errors, ValidationError{
							Path:     fmt.Sprintf("resources.%s", iter.Selector()),
							Message:  err.Error(),
							Severity: "error",
						})
					} else {
						parsedConfig.Resources = append(parsedConfig.Resources, resource)
					}
				}
			}
		} else if resourcesVal.Kind() == cue.ListKind {
			// List of resources
			list, err := resourcesVal.List()
			if err != nil {
				parsedConfig.Errors = append(parsedConfig.Errors, ValidationError{
					Path:     "resources",
					Message:  fmt.Sprintf("failed to list resources: %v", err),
					Severity: "error",
				})
			} else {
				idx := 0
				for list.Next() {
					resource, err := cp.extractResource("", list.Value())
					if err != nil {
						parsedConfig.Errors = append(parsedConfig.Errors, ValidationError{
							Path:     fmt.Sprintf("resources[%d]", idx),
							Message:  err.Error(),
							Severity: "error",
						})
					} else {
						parsedConfig.Resources = append(parsedConfig.Resources, resource)
					}
					idx++
				}
			}
		}
	}

	return parsedConfig, nil
}

// extractResource extracts a resource configuration from a CUE value.
func (cp *CUEParser) extractResource(id string, val cue.Value) (ResourceConfig, error) {
	var resource ResourceConfig

	// Decode the resource
	if err := val.Decode(&resource); err != nil {
		return resource, fmt.Errorf("failed to decode resource: %w", err)
	}

	// If ID is provided as key and not in value, use the key
	if resource.ID == "" && id != "" {
		resource.ID = id
	}

	// Validate using struct tags
	if err := cp.validator.Struct(resource); err != nil {
		return resource, fmt.Errorf("validation failed: %w", err)
	}

	return resource, nil
}

// convertCUEErrors converts CUE errors to ValidationError slice.
func (cp *CUEParser) convertCUEErrors(err error) []ValidationError {
	var validationErrors []ValidationError

	// Handle CUE error types
	errs := errors.Errors(err)
	for _, e := range errs {
		pos := errors.Positions(e)
		var file string
		var line, column int

		if len(pos) > 0 {
			file = pos[0].Filename()
			line = pos[0].Line()
			column = pos[0].Column()
		}

		validationErrors = append(validationErrors, ValidationError{
			File:     file,
			Line:     line,
			Column:   column,
			Message:  errors.Details(e, nil),
			Severity: "error",
		})
	}

	return validationErrors
}

// ParseInline parses inline CUE content.
func (cp *CUEParser) ParseInline(ctx context.Context, content string) (*ParsedConfig, error) {
	val := cp.ctx.CompileString(content)
	if err := val.Err(); err != nil {
		return &ParsedConfig{
			SourceFiles: []string{"inline"},
			ParsedAt:    time.Now(),
			Errors:      cp.convertCUEErrors(err),
		}, nil
	}

	return cp.extractConfig(val, []string{"inline"})
}

// ValidateWithSchema validates a CUE value against a schema.
func (cp *CUEParser) ValidateWithSchema(ctx context.Context, data interface{}, schemaName string) error {
	return cp.schemaRegistry.ValidateAgainstSchema(ctx, schemaName, data)
}

// GetSchemaRegistry returns the schema registry.
func (cp *CUEParser) GetSchemaRegistry() *SchemaRegistry {
	return cp.schemaRegistry
}

// ExtractValue extracts a specific path from a CUE configuration.
func (cp *CUEParser) ExtractValue(val cue.Value, path string) (interface{}, error) {
	v := val.LookupPath(cue.ParsePath(path))
	if !v.Exists() {
		return nil, fmt.Errorf("path %s not found", path)
	}

	// Try to decode to JSON first
	var result interface{}
	if err := v.Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode value at %s: %w", path, err)
	}

	return result, nil
}

// MergeValues merges two CUE values.
func (cp *CUEParser) MergeValues(val1, val2 cue.Value) (cue.Value, error) {
	merged := val1.Unify(val2)
	if err := merged.Err(); err != nil {
		return cue.Value{}, fmt.Errorf("failed to merge values: %w", err)
	}
	return merged, nil
}

// ExportJSON exports a CUE value to JSON.
func (cp *CUEParser) ExportJSON(val cue.Value) ([]byte, error) {
	var data interface{}
	if err := val.Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode value: %w", err)
	}

	return json.MarshalIndent(data, "", "  ")
}

// LoadFromDirectory loads all CUE files from a directory.
func (cp *CUEParser) LoadFromDirectory(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".cue") {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return files, nil
}
