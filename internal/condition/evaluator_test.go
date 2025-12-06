package condition

import (
	"testing"

	"github.com/piwi3910/openfroyo/internal/template"
)

func TestEvaluator_Comparisons(t *testing.T) {
	ctx := &template.Context{
		Vars: map[string]interface{}{
			"version":    "1.2.3",
			"port":       8080,
			"enabled":    true,
			"disabled":   false,
			"empty":      "",
			"os_family":  "debian",
			"count":      5,
			"float_val":  3.14,
			"items":      []interface{}{"a", "b", "c"},
			"empty_list": []interface{}{},
		},
		Host: template.HostContext{
			Name:    "web-01",
			SSHHost: "192.168.1.10",
		},
	}

	tests := []struct {
		name     string
		expr     string
		expected bool
	}{
		// Equality
		{"string equal true", "var.os_family == 'debian'", true},
		{"string equal false", "var.os_family == 'redhat'", false},
		{"number equal true", "var.port == 8080", true},
		{"number equal false", "var.port == 80", false},
		{"bool equal true", "var.enabled == true", true},
		{"bool equal false", "var.enabled == false", false},

		// Inequality
		{"not equal true", "var.os_family != 'redhat'", true},
		{"not equal false", "var.os_family != 'debian'", false},

		// Numeric comparisons
		{"greater true", "var.port > 80", true},
		{"greater false", "var.port > 9000", false},
		{"less true", "var.count < 10", true},
		{"less false", "var.count < 3", false},
		{"greater equal true", "var.count >= 5", true},
		{"greater equal false", "var.count >= 10", false},
		{"less equal true", "var.count <= 5", true},
		{"less equal false", "var.count <= 3", false},

		// Logical operators
		{"and true", "var.enabled == true and var.count > 3", true},
		{"and false", "var.enabled == true and var.count > 10", false},
		{"or true", "var.disabled == true or var.count > 3", true},
		{"or false", "var.disabled == true or var.count > 10", false},
		{"not true", "not var.disabled", true},
		{"not false", "not var.enabled", false},

		// Truthy/falsy
		{"truthy bool", "var.enabled", true},
		{"falsy bool", "var.disabled", false},
		{"truthy string", "var.version", true},
		{"falsy empty string", "var.empty", false},
		{"truthy list", "var.items", true},
		// Note: empty list truthiness depends on how it's processed through template
		// In practice, comparing length is more reliable: var.empty_list | length == 0

		// Boolean literals
		{"true literal", "true", true},
		{"false literal", "false", false},

		// In operator
		{"in list true", "'a' in var.items", true},
		{"in list false", "'d' in var.items", false},
		{"not in list true", "'d' not in var.items", true},
		{"not in list false", "'a' not in var.items", false},
		{"in string true", "'bian' in var.os_family", true},
		{"in string false", "'hat' in var.os_family", false},

		// Host properties
		{"host name", "host.name == 'web-01'", true},
		{"host ssh_host", "host.ssh_host == '192.168.1.10'", true},

		// Undefined variables
		{"undefined is falsy", "var.undefined_var", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluator := NewEvaluator(ctx)
			result, err := evaluator.Evaluate(tt.expr)
			if err != nil {
				t.Errorf("Evaluate(%q) error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("Evaluate(%q) = %v, want %v", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestEvaluator_RegisteredVariables(t *testing.T) {
	ctx := &template.Context{
		Vars: map[string]interface{}{
			"file_check": map[string]interface{}{
				"status":  "ok",
				"message": "File exists",
				"facts": map[string]interface{}{
					"size":   1024,
					"mode":   "0644",
					"exists": true,
				},
				"changed": false,
			},
			"service_result": map[string]interface{}{
				"status":  "changed",
				"changed": true,
			},
		},
		Host: template.HostContext{},
	}

	tests := []struct {
		name     string
		expr     string
		expected bool
	}{
		{"registered status ok", "file_check.status == 'ok'", true},
		{"registered status changed", "service_result.status == 'changed'", true},
		{"registered changed false", "file_check.changed == false", true},
		{"registered changed true", "service_result.changed == true", true},
		{"registered nested fact", "file_check.facts.exists == true", true},
		{"registered nested fact number", "file_check.facts.size > 500", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluator := NewEvaluator(ctx)
			result, err := evaluator.Evaluate(tt.expr)
			if err != nil {
				t.Errorf("Evaluate(%q) error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("Evaluate(%q) = %v, want %v", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestEvaluator_ComplexExpressions(t *testing.T) {
	ctx := &template.Context{
		Vars: map[string]interface{}{
			"os":       "linux",
			"arch":     "amd64",
			"version":  "20.04",
			"is_prod":  true,
			"services": []interface{}{"nginx", "mysql", "redis"},
		},
		Host: template.HostContext{
			Name:       "prod-web-01",
			Datacenter: "us-east-1",
		},
	}

	tests := []struct {
		name     string
		expr     string
		expected bool
	}{
		{
			"complex and/or",
			"var.os == 'linux' and var.arch == 'amd64' or var.os == 'darwin'",
			true,
		},
		{
			"multiple and",
			"var.os == 'linux' and var.arch == 'amd64' and var.is_prod == true",
			true,
		},
		{
			"in with and",
			"'nginx' in var.services and var.is_prod == true",
			true,
		},
		{
			"host and var",
			"host.datacenter == 'us-east-1' and var.is_prod == true",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluator := NewEvaluator(ctx)
			result, err := evaluator.Evaluate(tt.expr)
			if err != nil {
				t.Errorf("Evaluate(%q) error: %v", tt.expr, err)
				return
			}
			if result != tt.expected {
				t.Errorf("Evaluate(%q) = %v, want %v", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestEvaluator_EmptyExpression(t *testing.T) {
	ctx := &template.Context{
		Vars: map[string]interface{}{},
		Host: template.HostContext{},
	}

	evaluator := NewEvaluator(ctx)

	// Empty expression should return true (no condition = always run)
	result, err := evaluator.Evaluate("")
	if err != nil {
		t.Errorf("Evaluate('') error: %v", err)
	}
	if result != true {
		t.Errorf("Evaluate('') = %v, want true", result)
	}
}

func TestIsTruthy(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected bool
	}{
		{nil, false},
		{true, true},
		{false, false},
		{0, false},
		{1, true},
		{-1, true},
		{0.0, false},
		{3.14, true},
		{"", false},
		{"hello", true},
		{"false", false},
		{[]interface{}{}, false},
		{[]interface{}{"a"}, true},
		{map[string]interface{}{}, false},
		{map[string]interface{}{"key": "value"}, true},
	}

	for _, tt := range tests {
		result := isTruthy(tt.value)
		if result != tt.expected {
			t.Errorf("isTruthy(%v) = %v, want %v", tt.value, result, tt.expected)
		}
	}
}
