package config

import (
	"context"
	"testing"
	"time"
)

func TestStarlarkEvaluator_Evaluate(t *testing.T) {
	evaluator := NewStarlarkEvaluator(5 * time.Second)
	ctx := context.Background()

	tests := []struct {
		name      string
		script    string
		input     map[string]interface{}
		checkFunc func(*testing.T, *StarlarkResult)
		wantErr   bool
	}{
		{
			name: "simple arithmetic",
			script: `
result = 2 + 2
`,
			input: nil,
			checkFunc: func(t *testing.T, sr *StarlarkResult) {
				if sr.Output["result"] != int64(4) {
					t.Errorf("expected result=4, got %v", sr.Output["result"])
				}
			},
			wantErr: false,
		},
		{
			name: "use input variables",
			script: `
doubled = count * 2
`,
			input: map[string]interface{}{
				"count": 5,
			},
			checkFunc: func(t *testing.T, sr *StarlarkResult) {
				if sr.Output["doubled"] != int64(10) {
					t.Errorf("expected doubled=10, got %v", sr.Output["doubled"])
				}
			},
			wantErr: false,
		},
		{
			name: "generate list with function",
			script: `
def make_list(n):
    result = []
    for i in range(n):
        result.append(i * 2)
    return result

output = make_list(5)
`,
			input: nil,
			checkFunc: func(t *testing.T, sr *StarlarkResult) {
				output, ok := sr.Output["output"].([]interface{})
				if !ok {
					t.Fatalf("expected output to be a list, got %T", sr.Output["output"])
				}
				if len(output) != 5 {
					t.Errorf("expected list of length 5, got %d", len(output))
				}
				if output[0] != int64(0) || output[4] != int64(8) {
					t.Errorf("unexpected list values: %v", output)
				}
			},
			wantErr: false,
		},
		{
			name: "generate dict with function",
			script: `
def make_servers(count):
    servers = {}
    for i in range(count):
        servers["server_" + str(i)] = {
            "id": i,
            "name": "server-" + str(i),
            "port": 8000 + i,
        }
    return servers

result = make_servers(3)
`,
			input: nil,
			checkFunc: func(t *testing.T, sr *StarlarkResult) {
				result, ok := sr.Output["result"].(map[string]interface{})
				if !ok {
					t.Fatalf("expected result to be a dict, got %T", sr.Output["result"])
				}
				if len(result) != 3 {
					t.Errorf("expected dict with 3 keys, got %d", len(result))
				}

				server0, ok := result["server_0"].(map[string]interface{})
				if !ok {
					t.Fatalf("expected server_0 to be a dict")
				}
				if server0["name"] != "server-0" {
					t.Errorf("expected server_0.name='server-0', got %v", server0["name"])
				}
			},
			wantErr: false,
		},
		{
			name: "list comprehension",
			script: `
result = [i * 2 for i in range(1, 6)]
`,
			input: nil,
			checkFunc: func(t *testing.T, sr *StarlarkResult) {
				result, ok := sr.Output["result"].([]interface{})
				if !ok {
					t.Fatalf("expected result to be a list")
				}
				if len(result) != 5 {
					t.Errorf("expected list of length 5, got %d", len(result))
				}
			},
			wantErr: false,
		},
		{
			name: "dict comprehension",
			script: `
items = ["a", "b", "c"]
result = {i: val for i, val in enumerate(items)}
`,
			input: nil,
			checkFunc: func(t *testing.T, sr *StarlarkResult) {
				result, ok := sr.Output["result"].(map[string]interface{})
				if !ok {
					t.Fatalf("expected result to be a dict")
				}
				if len(result) != 3 {
					t.Errorf("expected dict with 3 keys, got %d", len(result))
				}
			},
			wantErr: false,
		},
		{
			name: "syntax error",
			script: `
invalid syntax here
`,
			input:   nil,
			wantErr: true,
		},
		{
			name: "runtime error",
			script: `
result = undefined_variable
`,
			input:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.Evaluate(ctx, tt.script, tt.input)

			if tt.wantErr {
				if err == nil && result.Error == "" {
					t.Errorf("expected error, got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result.Error != "" {
					t.Errorf("unexpected result error: %s", result.Error)
				}
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}

			// Check execution time is recorded
			if result != nil && result.ExecutionTime == 0 {
				t.Error("expected non-zero execution time")
			}
		})
	}
}

func TestStarlarkEvaluator_Timeout(t *testing.T) {
	evaluator := NewStarlarkEvaluator(100 * time.Millisecond)
	ctx := context.Background()

	// Script that takes too long
	script := `
def slow_function():
    result = 0
    for i in range(10000000):
        result = result + i
    return result

output = slow_function()
`

	result, err := evaluator.Evaluate(ctx, script, nil)
	if err == nil {
		t.Error("expected timeout error")
	}

	if result != nil && result.Error == "" {
		t.Error("expected timeout error in result")
	}
}

func TestStarlarkEvaluator_TypeConversion(t *testing.T) {
	evaluator := NewStarlarkEvaluator(5 * time.Second)
	ctx := context.Background()

	tests := []struct {
		name      string
		input     map[string]interface{}
		script    string
		checkFunc func(*testing.T, *StarlarkResult)
	}{
		{
			name: "bool conversion",
			input: map[string]interface{}{
				"enabled": true,
			},
			script: `
result = enabled and True
`,
			checkFunc: func(t *testing.T, sr *StarlarkResult) {
				if sr.Output["result"] != true {
					t.Errorf("expected result=true, got %v", sr.Output["result"])
				}
			},
		},
		{
			name: "int conversion",
			input: map[string]interface{}{
				"count": 42,
			},
			script: `
result = count + 8
`,
			checkFunc: func(t *testing.T, sr *StarlarkResult) {
				if sr.Output["result"] != int64(50) {
					t.Errorf("expected result=50, got %v", sr.Output["result"])
				}
			},
		},
		{
			name: "float conversion",
			input: map[string]interface{}{
				"price": 19.99,
			},
			script: `
result = price * 2
`,
			checkFunc: func(t *testing.T, sr *StarlarkResult) {
				result, ok := sr.Output["result"].(float64)
				if !ok {
					t.Fatalf("expected result to be float64")
				}
				expected := 19.99 * 2
				if result != expected {
					t.Errorf("expected result=%.2f, got %.2f", expected, result)
				}
			},
		},
		{
			name: "string conversion",
			input: map[string]interface{}{
				"name": "test",
			},
			script: `
result = name + "-suffix"
`,
			checkFunc: func(t *testing.T, sr *StarlarkResult) {
				if sr.Output["result"] != "test-suffix" {
					t.Errorf("expected result='test-suffix', got %v", sr.Output["result"])
				}
			},
		},
		{
			name: "list conversion",
			input: map[string]interface{}{
				"items": []interface{}{"a", "b", "c"},
			},
			script: `
result = len(items)
`,
			checkFunc: func(t *testing.T, sr *StarlarkResult) {
				if sr.Output["result"] != int64(3) {
					t.Errorf("expected result=3, got %v", sr.Output["result"])
				}
			},
		},
		{
			name: "dict conversion",
			input: map[string]interface{}{
				"config": map[string]interface{}{
					"host": "localhost",
					"port": 8080,
				},
			},
			script: `
result = config["host"] + ":" + str(config["port"])
`,
			checkFunc: func(t *testing.T, sr *StarlarkResult) {
				if sr.Output["result"] != "localhost:8080" {
					t.Errorf("expected result='localhost:8080', got %v", sr.Output["result"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.Evaluate(ctx, tt.script, tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Error != "" {
				t.Fatalf("unexpected result error: %s", result.Error)
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

func TestStarlarkEvaluator_Security(t *testing.T) {
	evaluator := NewStarlarkEvaluator(5 * time.Second)
	ctx := context.Background()

	// Attempt to use print (should be suppressed)
	script := `
print("this should not appear")
result = "done"
`

	result, err := evaluator.Evaluate(ctx, script, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Output["result"] != "done" {
		t.Errorf("expected result='done', got %v", result.Output["result"])
	}
}
