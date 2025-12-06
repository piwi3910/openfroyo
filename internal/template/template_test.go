package template

import (
	"reflect"
	"testing"
)

func TestProcess(t *testing.T) {
	ctx := &Context{
		Vars: map[string]interface{}{
			"name":    "nginx",
			"port":    8080,
			"enabled": true,
			"config": map[string]interface{}{
				"path": "/etc/nginx",
				"file": "nginx.conf",
			},
		},
		Host: HostContext{
			Name:       "web-01",
			SSHHost:    "192.168.1.10",
			SSHPort:    22,
			SSHUser:    "admin",
			Datacenter: "us-east-1",
			AgentID:    "agent-123",
			Mode:       "ssh",
		},
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple variable",
			input:    "Service: {{ var.name }}",
			expected: "Service: nginx",
		},
		{
			name:     "variable with vars prefix",
			input:    "Port: {{ vars.port }}",
			expected: "Port: 8080",
		},
		{
			name:     "host property",
			input:    "Host: {{ host.name }}",
			expected: "Host: web-01",
		},
		{
			name:     "host ssh_host",
			input:    "IP: {{ host.ssh_host }}",
			expected: "IP: 192.168.1.10",
		},
		{
			name:     "multiple variables",
			input:    "{{ var.name }} on {{ host.name }}:{{ vars.port }}",
			expected: "nginx on web-01:8080",
		},
		{
			name:     "nested variable",
			input:    "Config: {{ var.config.path }}/{{ var.config.file }}",
			expected: "Config: /etc/nginx/nginx.conf",
		},
		{
			name:     "no template",
			input:    "Plain text without templates",
			expected: "Plain text without templates",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Process(tt.input, ctx)
			if err != nil {
				t.Errorf("Process failed: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestProcessWithFilters(t *testing.T) {
	ctx := &Context{
		Vars: map[string]interface{}{
			"name":     "nginx",
			"empty":    "",
			"spaced":   "  hello world  ",
			"sentence": "hello world",
		},
		Host: HostContext{
			Name: "web-01",
		},
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "default filter with value",
			input:    "{{ var.name | default('apache') }}",
			expected: "nginx",
		},
		{
			name:     "default filter with empty",
			input:    "{{ var.empty | default('fallback') }}",
			expected: "fallback",
		},
		{
			name:     "upper filter",
			input:    "{{ var.name | upper }}",
			expected: "NGINX",
		},
		{
			name:     "lower filter",
			input:    "{{ var.sentence | lower }}",
			expected: "hello world",
		},
		{
			name:     "trim filter",
			input:    "{{ var.spaced | trim }}",
			expected: "hello world",
		},
		{
			name:     "replace filter",
			input:    "{{ var.sentence | replace('world', 'universe') }}",
			expected: "hello universe",
		},
		{
			name:     "length filter string",
			input:    "{{ var.name | length }}",
			expected: "5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Process(tt.input, ctx)
			if err != nil {
				t.Errorf("Process failed: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestProcessVars(t *testing.T) {
	ctx := &Context{
		Vars: map[string]interface{}{
			"base_path": "/opt/app",
			"version":   "1.0.0",
		},
		Host: HostContext{
			Name: "web-01",
		},
	}

	vars := map[string]interface{}{
		"install_path": "{{ var.base_path }}/{{ var.version }}",
		"host_name":    "{{ host.name }}",
		"static":       "no template",
		"number":       42,
		"nested": map[string]interface{}{
			"path": "{{ var.base_path }}/config",
		},
	}

	result, err := ProcessVars(vars, ctx)
	if err != nil {
		t.Fatalf("ProcessVars failed: %v", err)
	}

	if result["install_path"] != "/opt/app/1.0.0" {
		t.Errorf("Expected '/opt/app/1.0.0', got '%v'", result["install_path"])
	}

	if result["host_name"] != "web-01" {
		t.Errorf("Expected 'web-01', got '%v'", result["host_name"])
	}

	if result["static"] != "no template" {
		t.Errorf("Expected 'no template', got '%v'", result["static"])
	}

	if result["number"] != 42 {
		t.Errorf("Expected 42, got '%v'", result["number"])
	}

	nested, ok := result["nested"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected nested to be a map")
	}
	if nested["path"] != "/opt/app/config" {
		t.Errorf("Expected '/opt/app/config', got '%v'", nested["path"])
	}
}

func TestArrayFilters(t *testing.T) {
	ctx := &Context{
		Vars: map[string]interface{}{
			"items": []interface{}{"a", "b", "c"},
		},
		Host: HostContext{},
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "join filter",
			input:    "{{ var.items | join(',') }}",
			expected: "a,b,c",
		},
		{
			name:     "first filter",
			input:    "{{ var.items | first }}",
			expected: "a",
		},
		{
			name:     "last filter",
			input:    "{{ var.items | last }}",
			expected: "c",
		},
		{
			name:     "length filter array",
			input:    "{{ var.items | length }}",
			expected: "3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Process(tt.input, ctx)
			if err != nil {
				t.Errorf("Process failed: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestSplitFilter(t *testing.T) {
	ctx := &Context{
		Vars: map[string]interface{}{
			"csv": "a,b,c",
		},
		Host: HostContext{},
	}

	// Test split returns array
	result, err := evaluateExpression("var.csv | split(',')", ctx)
	if err != nil {
		t.Fatalf("evaluateExpression failed: %v", err)
	}

	arr, ok := result.([]string)
	if !ok {
		t.Fatalf("Expected []string, got %T", result)
	}

	expected := []string{"a", "b", "c"}
	if !reflect.DeepEqual(arr, expected) {
		t.Errorf("Expected %v, got %v", expected, arr)
	}
}

func TestParseArgs(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{
			input:    "'hello'",
			expected: []string{"hello"},
		},
		{
			input:    "'hello', 'world'",
			expected: []string{"hello", "world"},
		},
		{
			input:    "\"hello\", \"world\"",
			expected: []string{"hello", "world"},
		},
		{
			input:    "'a,b', 'c'",
			expected: []string{"a,b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseArgs(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestHostProperties(t *testing.T) {
	ctx := &Context{
		Vars: map[string]interface{}{},
		Host: HostContext{
			Name:       "web-01",
			SSHHost:    "192.168.1.10",
			SSHPort:    22,
			SSHUser:    "admin",
			Datacenter: "us-east-1",
			AgentID:    "agent-123",
			Mode:       "ssh",
		},
	}

	tests := []struct {
		property string
		expected interface{}
	}{
		{"name", "web-01"},
		{"ssh_host", "192.168.1.10"},
		{"ssh_port", 22},
		{"ssh_user", "admin"},
		{"datacenter", "us-east-1"},
		{"agent_id", "agent-123"},
		{"mode", "ssh"},
	}

	for _, tt := range tests {
		t.Run(tt.property, func(t *testing.T) {
			result, err := getHostProperty(ctx.Host, tt.property)
			if err != nil {
				t.Errorf("getHostProperty failed: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}

	// Test invalid property
	_, err := getHostProperty(ctx.Host, "invalid")
	if err == nil {
		t.Error("Expected error for invalid property")
	}
}
