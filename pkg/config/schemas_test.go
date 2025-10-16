package config

import (
	"context"
	"testing"
)

func TestSchemaRegistry_RegisterAndGet(t *testing.T) {
	sr := NewSchemaRegistry()

	customSchema := `
#CustomType: {
	field1: string
	field2: int
}
`

	err := sr.RegisterSchema("custom", customSchema)
	if err != nil {
		t.Fatalf("failed to register schema: %v", err)
	}

	schema, ok := sr.GetSchema("custom")
	if !ok {
		t.Fatal("expected to find custom schema")
	}

	if schema.Err() != nil {
		t.Errorf("schema has errors: %v", schema.Err())
	}
}

func TestSchemaRegistry_BuiltInSchemas(t *testing.T) {
	sr := NewSchemaRegistry()

	builtins := []string{
		"resource",
		"workspace",
		"provider",
		"target",
		"dependency",
	}

	for _, name := range builtins {
		t.Run(name, func(t *testing.T) {
			schema, ok := sr.GetSchema(name)
			if !ok {
				t.Fatalf("built-in schema %s not found", name)
			}

			if schema.Err() != nil {
				t.Errorf("built-in schema %s has errors: %v", name, schema.Err())
			}
		})
	}
}

func TestSchemaRegistry_ValidateResource(t *testing.T) {
	sr := NewSchemaRegistry()
	ctx := context.Background()

	tests := []struct {
		name     string
		resource ResourceConfig
		wantErr  bool
	}{
		{
			name: "valid resource",
			resource: ResourceConfig{
				ID:   "test_resource",
				Type: "linux.pkg",
				Name: "nginx",
				Config: []byte(`{"package":"nginx","state":"present"}`),
			},
			wantErr: false,
		},
		{
			name: "invalid resource - bad ID",
			resource: ResourceConfig{
				ID:   "invalid id with spaces",
				Type: "linux.pkg",
				Name: "nginx",
				Config: []byte(`{"package":"nginx"}`),
			},
			wantErr: true,
		},
		{
			name: "invalid resource - bad type",
			resource: ResourceConfig{
				ID:   "test",
				Type: "InvalidType",
				Name: "nginx",
				Config: []byte(`{"package":"nginx"}`),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sr.ValidateResource(ctx, tt.resource)

			if tt.wantErr {
				if err == nil {
					t.Error("expected validation error, got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestSchemaRegistry_ValidateWorkspace(t *testing.T) {
	sr := NewSchemaRegistry()
	ctx := context.Background()

	tests := []struct {
		name      string
		workspace WorkspaceConfig
		wantErr   bool
	}{
		{
			name: "valid workspace",
			workspace: WorkspaceConfig{
				Name:    "test-workspace",
				Version: "1.0",
				Backend: &BackendConfig{
					Type: "solo",
					Path: "./data",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid workspace - bad name",
			workspace: WorkspaceConfig{
				Name:    "invalid name!",
				Version: "1.0",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sr.ValidateWorkspace(ctx, tt.workspace)

			if tt.wantErr {
				if err == nil {
					t.Error("expected validation error, got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestSchemaRegistry_ValidateTarget(t *testing.T) {
	sr := NewSchemaRegistry()
	ctx := context.Background()

	tests := []struct {
		name    string
		target  TargetSelector
		wantErr bool
	}{
		{
			name: "valid target with labels",
			target: TargetSelector{
				Labels: map[string]string{
					"env":  "prod",
					"role": "web",
				},
			},
			wantErr: false,
		},
		{
			name: "valid target with hosts",
			target: TargetSelector{
				Hosts: []string{"host1", "host2"},
			},
			wantErr: false,
		},
		{
			name: "valid target with all",
			target: TargetSelector{
				All: true,
			},
			wantErr: false,
		},
		{
			name:    "invalid target - no targeting method",
			target:  TargetSelector{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sr.ValidateTarget(ctx, tt.target)

			if tt.wantErr {
				if err == nil {
					t.Error("expected validation error, got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestSchemaRegistry_ValidateDependency(t *testing.T) {
	sr := NewSchemaRegistry()
	ctx := context.Background()

	tests := []struct {
		name       string
		dependency DependencyConfig
		wantErr    bool
	}{
		{
			name: "valid require dependency",
			dependency: DependencyConfig{
				ResourceID: "resource1",
				Type:       "require",
			},
			wantErr: false,
		},
		{
			name: "valid notify dependency",
			dependency: DependencyConfig{
				ResourceID: "resource2",
				Type:       "notify",
			},
			wantErr: false,
		},
		{
			name: "valid order dependency",
			dependency: DependencyConfig{
				ResourceID: "resource3",
				Type:       "order",
			},
			wantErr: false,
		},
		{
			name: "invalid dependency - bad resource ID",
			dependency: DependencyConfig{
				ResourceID: "invalid id!",
				Type:       "require",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sr.ValidateDependency(ctx, tt.dependency)

			if tt.wantErr {
				if err == nil {
					t.Error("expected validation error, got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestSchemaRegistry_ListSchemas(t *testing.T) {
	sr := NewSchemaRegistry()

	schemas := sr.ListSchemas()

	if len(schemas) < 5 {
		t.Errorf("expected at least 5 schemas, got %d", len(schemas))
	}

	// Check for built-in schemas
	expectedSchemas := map[string]bool{
		"resource":   false,
		"workspace":  false,
		"provider":   false,
		"target":     false,
		"dependency": false,
	}

	for _, schema := range schemas {
		if _, exists := expectedSchemas[schema]; exists {
			expectedSchemas[schema] = true
		}
	}

	for name, found := range expectedSchemas {
		if !found {
			t.Errorf("expected built-in schema %s not found", name)
		}
	}
}

func TestSchemaRegistry_InvalidSchema(t *testing.T) {
	sr := NewSchemaRegistry()

	invalidSchema := `
this is not valid CUE syntax
`

	err := sr.RegisterSchema("invalid", invalidSchema)
	if err == nil {
		t.Error("expected error when registering invalid schema")
	}
}
