package config

import (
	"context"
	"encoding/json"
	"testing"
)

// TestConciseSyntax tests the concise provider namespace syntax.
func TestConciseSyntax(t *testing.T) {
	parser := NewCUEParser()

	// Test basic concise syntax
	cueContent := `
package test

workspace: {
	name: "test"
	version: "1.0.0"
}

// Concise syntax
linux: pkg: {
	nginx: {}
	postgresql: {
		state: "present"
		version: "14.5"
	}
	apache2: {
		state: "absent"
	}
}
`

	parsedConfig, err := parser.ParseInline(context.Background(), cueContent)
	if err != nil {
		t.Fatalf("Failed to parse CUE: %v", err)
	}

	if len(parsedConfig.Errors) > 0 {
		t.Fatalf("Parse errors: %v", parsedConfig.Errors)
	}

	// Should have 3 resources from concise syntax
	if len(parsedConfig.Resources) != 3 {
		t.Fatalf("Expected 3 resources, got %d", len(parsedConfig.Resources))
	}

	// Test nginx resource (defaults)
	nginxResource := findResource(parsedConfig.Resources, "linux-pkg-nginx")
	if nginxResource == nil {
		t.Fatal("nginx resource not found")
	}

	if nginxResource.Type != "linux.pkg::pkg" {
		t.Errorf("Expected type 'linux.pkg::pkg', got '%s'", nginxResource.Type)
	}

	var nginxConfig map[string]interface{}
	if err := json.Unmarshal(nginxResource.Config, &nginxConfig); err != nil {
		t.Fatalf("Failed to unmarshal nginx config: %v", err)
	}

	if nginxConfig["package"] != "nginx" {
		t.Errorf("Expected package 'nginx', got '%v'", nginxConfig["package"])
	}

	if nginxConfig["state"] != "present" {
		t.Errorf("Expected state 'present', got '%v'", nginxConfig["state"])
	}

	// Test postgresql resource (with version)
	postgresResource := findResource(parsedConfig.Resources, "linux-pkg-postgresql")
	if postgresResource == nil {
		t.Fatal("postgresql resource not found")
	}

	var postgresConfig map[string]interface{}
	if err := json.Unmarshal(postgresResource.Config, &postgresConfig); err != nil {
		t.Fatalf("Failed to unmarshal postgresql config: %v", err)
	}

	if postgresConfig["package"] != "postgresql" {
		t.Errorf("Expected package 'postgresql', got '%v'", postgresConfig["package"])
	}

	if postgresConfig["version"] != "14.5" {
		t.Errorf("Expected version '14.5', got '%v'", postgresConfig["version"])
	}

	// Test apache2 resource (absent state)
	apacheResource := findResource(parsedConfig.Resources, "linux-pkg-apache2")
	if apacheResource == nil {
		t.Fatal("apache2 resource not found")
	}

	var apacheConfig map[string]interface{}
	if err := json.Unmarshal(apacheResource.Config, &apacheConfig); err != nil {
		t.Fatalf("Failed to unmarshal apache2 config: %v", err)
	}

	if apacheConfig["state"] != "absent" {
		t.Errorf("Expected state 'absent', got '%v'", apacheConfig["state"])
	}
}

// TestConciseAndVerboseMixed tests that both syntaxes can be used together.
func TestConciseAndVerboseMixed(t *testing.T) {
	parser := NewCUEParser()

	cueContent := `
package test

workspace: {
	name: "test"
	version: "1.0.0"
}

// Concise syntax
linux: pkg: nginx: {}

// Verbose syntax
resources: custom_nginx: {
	id: "nginx-custom"
	type: "linux.pkg::package"
	name: "nginx-custom"
	config: {
		package: "nginx"
		state: "latest"
	}
}
`

	parsedConfig, err := parser.ParseInline(context.Background(), cueContent)
	if err != nil {
		t.Fatalf("Failed to parse CUE: %v", err)
	}

	if len(parsedConfig.Errors) > 0 {
		t.Fatalf("Parse errors: %v", parsedConfig.Errors)
	}

	// Should have 2 resources: 1 from concise, 1 from verbose
	if len(parsedConfig.Resources) != 2 {
		t.Fatalf("Expected 2 resources, got %d", len(parsedConfig.Resources))
	}

	// Check both exist
	if findResource(parsedConfig.Resources, "linux-pkg-nginx") == nil {
		t.Error("Concise nginx resource not found")
	}

	if findResource(parsedConfig.Resources, "nginx-custom") == nil {
		t.Error("Verbose nginx resource not found")
	}
}

// TestMultiplePackages tests multiple packages in concise syntax.
func TestMultiplePackages(t *testing.T) {
	parser := NewCUEParser()

	cueContent := `
package test

workspace: {
	name: "test"
	version: "1.0.0"
}

linux: pkg: {
	nginx: {}
	postgresql: {}
	redis: {}
	curl: {}
	wget: {}
}
`

	parsedConfig, err := parser.ParseInline(context.Background(), cueContent)
	if err != nil {
		t.Fatalf("Failed to parse CUE: %v", err)
	}

	if len(parsedConfig.Errors) > 0 {
		t.Fatalf("Parse errors: %v", parsedConfig.Errors)
	}

	// Should have 5 resources
	if len(parsedConfig.Resources) != 5 {
		t.Fatalf("Expected 5 resources, got %d", len(parsedConfig.Resources))
	}

	expectedPackages := []string{"nginx", "postgresql", "redis", "curl", "wget"}
	for _, pkg := range expectedPackages {
		resourceID := "linux-pkg-" + pkg
		if findResource(parsedConfig.Resources, resourceID) == nil {
			t.Errorf("Package '%s' not found", pkg)
		}
	}
}

// Helper function to find a resource by ID
func findResource(resources []ResourceConfig, id string) *ResourceConfig {
	for i := range resources {
		if resources[i].ID == id {
			return &resources[i]
		}
	}
	return nil
}
