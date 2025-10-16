package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/openfroyo/openfroyo/pkg/engine"
)

func TestCUEParser_ParseInline(t *testing.T) {
	parser := NewCUEParser()
	ctx := context.Background()

	tests := []struct {
		name      string
		content   string
		wantErr   bool
		errCount  int
		checkFunc func(*testing.T, *ParsedConfig)
	}{
		{
			name: "valid simple config",
			content: `
workspace: {
	name: "test"
	version: "1.0"
}

resources: {
	test_resource: {
		id: "test_res"
		type: "linux.pkg"
		name: "nginx"
		config: {
			package: "nginx"
			state: "present"
		}
	}
}
`,
			wantErr: false,
			checkFunc: func(t *testing.T, pc *ParsedConfig) {
				if pc.Workspace.Name != "test" {
					t.Errorf("expected workspace name 'test', got %s", pc.Workspace.Name)
				}
				if len(pc.Resources) != 1 {
					t.Errorf("expected 1 resource, got %d", len(pc.Resources))
				}
				if len(pc.Resources) > 0 && pc.Resources[0].Type != "linux.pkg" {
					t.Errorf("expected resource type 'linux.pkg', got %s", pc.Resources[0].Type)
				}
			},
		},
		{
			name: "invalid CUE syntax",
			content: `
workspace: {
	name: "test"
	invalid syntax here
}
`,
			wantErr:  true,
			errCount: 1,
		},
		{
			name: "missing required field",
			content: `
resources: {
	test_resource: {
		type: "linux.pkg"
		config: {}
	}
}
`,
			wantErr:  true,
			errCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc, err := parser.ParseInline(ctx, tt.content)

			if tt.wantErr {
				if err == nil && len(pc.Errors) == 0 {
					t.Errorf("expected error, got none")
				}
				if tt.errCount > 0 && len(pc.Errors) != tt.errCount {
					t.Errorf("expected %d errors, got %d", tt.errCount, len(pc.Errors))
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(pc.Errors) > 0 {
					t.Errorf("unexpected validation errors: %v", pc.Errors)
				}
				if tt.checkFunc != nil {
					tt.checkFunc(t, pc)
				}
			}
		})
	}
}

func TestCUEParser_ParseFile(t *testing.T) {
	parser := NewCUEParser()
	ctx := context.Background()

	// Create temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.cue")

	content := `
workspace: {
	name: "filetest"
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
		labels: {
			env: "test"
		}
	}
}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	pc, err := parser.Parse(ctx, []string{testFile})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(pc.Errors) > 0 {
		t.Fatalf("unexpected validation errors: %v", pc.Errors)
	}

	if pc.Workspace.Name != "filetest" {
		t.Errorf("expected workspace name 'filetest', got %s", pc.Workspace.Name)
	}

	if len(pc.Resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(pc.Resources))
	}

	res := pc.Resources[0]
	if res.ID != "web" {
		t.Errorf("expected resource ID 'web', got %s", res.ID)
	}
	if res.Labels["env"] != "test" {
		t.Errorf("expected label env='test', got %s", res.Labels["env"])
	}
}

func TestCUEParser_Evaluate(t *testing.T) {
	parser := NewCUEParser()
	ctx := context.Background()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "config.cue")

	content := `
workspace: {
	name: "integration"
	version: "1.0"
	providers: [{
		name: "linux.pkg"
		version: ">=1.0.0"
	}]
}

resources: {
	app: {
		id: "app"
		type: "linux.pkg"
		name: "myapp"
		config: {
			package: "myapp"
			state: "present"
		}
	}
}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cfg, err := parser.Evaluate(ctx, []string{testFile})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg == nil {
		t.Fatal("expected non-nil config")
	}

	if len(cfg.Resources) != 1 {
		t.Errorf("expected 1 resource, got %d", len(cfg.Resources))
	}
}

func TestCUEParser_MergeConfigs(t *testing.T) {
	parser := NewCUEParser()
	ctx := context.Background()

	tmpDir := t.TempDir()

	// Create two config files
	file1 := filepath.Join(tmpDir, "config1.cue")
	file2 := filepath.Join(tmpDir, "config2.cue")

	content1 := `
workspace: {name: "merge1", version: "1.0"}
resources: {
	res1: {
		id: "res1"
		type: "linux.pkg"
		name: "pkg1"
		config: {package: "pkg1", state: "present"}
	}
}
`

	content2 := `
workspace: {name: "merge2", version: "1.0"}
resources: {
	res2: {
		id: "res2"
		type: "linux.pkg"
		name: "pkg2"
		config: {package: "pkg2", state: "present"}
	}
}
`

	if err := os.WriteFile(file1, []byte(content1), 0644); err != nil {
		t.Fatalf("failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte(content2), 0644); err != nil {
		t.Fatalf("failed to create file2: %v", err)
	}

	cfg1, err := parser.Evaluate(ctx, []string{file1})
	if err != nil {
		t.Fatalf("failed to evaluate config1: %v", err)
	}

	cfg2, err := parser.Evaluate(ctx, []string{file2})
	if err != nil {
		t.Fatalf("failed to evaluate config2: %v", err)
	}

	merged, err := parser.MergeConfigs(ctx, []*engine.Config{cfg1, cfg2})
	if err != nil {
		t.Fatalf("failed to merge configs: %v", err)
	}

	if len(merged.Resources) != 2 {
		t.Errorf("expected 2 resources in merged config, got %d", len(merged.Resources))
	}
}

func TestCUEParser_Dependencies(t *testing.T) {
	parser := NewCUEParser()
	ctx := context.Background()

	content := `
workspace: {name: "deps", version: "1.0"}

resources: {
	pkg: {
		id: "pkg"
		type: "linux.pkg"
		name: "nginx"
		config: {package: "nginx", state: "present"}
	}

	svc: {
		id: "svc"
		type: "linux.service"
		name: "nginx"
		config: {name: "nginx", state: "running"}
		dependencies: [
			{resource_id: "pkg", type: "require"}
		]
	}
}
`

	pc, err := parser.ParseInline(ctx, content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(pc.Errors) > 0 {
		t.Fatalf("unexpected validation errors: %v", pc.Errors)
	}

	// Find service resource
	var svcRes *ResourceConfig
	for i := range pc.Resources {
		if pc.Resources[i].ID == "svc" {
			svcRes = &pc.Resources[i]
			break
		}
	}

	if svcRes == nil {
		t.Fatal("service resource not found")
	}

	if len(svcRes.Dependencies) != 1 {
		t.Errorf("expected 1 dependency, got %d", len(svcRes.Dependencies))
	}

	if len(svcRes.Dependencies) > 0 {
		dep := svcRes.Dependencies[0]
		if dep.ResourceID != "pkg" {
			t.Errorf("expected dependency on 'pkg', got %s", dep.ResourceID)
		}
		if dep.Type != engine.DependencyRequire {
			t.Errorf("expected require dependency, got %s", dep.Type)
		}
	}
}

func TestCUEParser_TargetSelectors(t *testing.T) {
	parser := NewCUEParser()
	ctx := context.Background()

	content := `
workspace: {name: "targets", version: "1.0"}

resources: {
	res_labels: {
		id: "res1"
		type: "linux.pkg"
		name: "pkg1"
		config: {package: "pkg1", state: "present"}
		target: {
			labels: {env: "prod", role: "web"}
		}
	}

	res_hosts: {
		id: "res2"
		type: "linux.pkg"
		name: "pkg2"
		config: {package: "pkg2", state: "present"}
		target: {
			hosts: ["host1", "host2"]
		}
	}

	res_all: {
		id: "res3"
		type: "linux.pkg"
		name: "pkg3"
		config: {package: "pkg3", state: "present"}
		target: {
			all: true
		}
	}
}
`

	pc, err := parser.ParseInline(ctx, content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(pc.Errors) > 0 {
		t.Fatalf("unexpected validation errors: %v", pc.Errors)
	}

	if len(pc.Resources) != 3 {
		t.Fatalf("expected 3 resources, got %d", len(pc.Resources))
	}

	// Check label-based target
	res1 := pc.Resources[0]
	if len(res1.Target.Labels) != 2 {
		t.Errorf("expected 2 target labels, got %d", len(res1.Target.Labels))
	}

	// Check host-based target
	res2 := pc.Resources[1]
	if len(res2.Target.Hosts) != 2 {
		t.Errorf("expected 2 target hosts, got %d", len(res2.Target.Hosts))
	}

	// Check all-targets
	res3 := pc.Resources[2]
	if !res3.Target.All {
		t.Error("expected target.all to be true")
	}
}
