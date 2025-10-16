package policy

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/openfroyo/openfroyo/pkg/engine"
	"github.com/rs/zerolog"
)

func TestNewEngine(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	eng, err := NewEngine(logger)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	if eng == nil {
		t.Fatal("Engine is nil")
	}

	// Check that built-in policies are loaded
	policies := eng.ListPolicies()
	if len(policies) == 0 {
		t.Fatal("No built-in policies loaded")
	}

	expectedPolicies := []string{
		"resource-naming",
		"required-labels",
		"state-drift",
		"operation-restrictions",
		"provider-versioning",
	}

	for _, expected := range expectedPolicies {
		found := false
		for _, p := range policies {
			if p.Name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected built-in policy not found: %s", expected)
		}
	}
}

func TestEvaluateResource_NamingPolicy(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	eng, err := NewEngine(logger)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	tests := []struct {
		name           string
		resource       *engine.Resource
		expectAllowed  bool
		expectViolation bool
	}{
		{
			name: "valid resource name",
			resource: &engine.Resource{
				ID:   "test-1",
				Name: "valid-resource-name",
				Labels: map[string]string{
					"env":   "development",
					"owner": "test-team",
				},
			},
			expectAllowed:   true,
			expectViolation: false,
		},
		{
			name: "uppercase in name",
			resource: &engine.Resource{
				ID:   "test-2",
				Name: "Invalid-Name",
				Labels: map[string]string{
					"env":   "development",
					"owner": "test-team",
				},
			},
			expectAllowed:   false,
			expectViolation: true,
		},
		{
			name: "name with underscores",
			resource: &engine.Resource{
				ID:   "test-3",
				Name: "invalid_name",
				Labels: map[string]string{
					"env":   "development",
					"owner": "test-team",
				},
			},
			expectAllowed:   false,
			expectViolation: true,
		},
		{
			name: "name too short",
			resource: &engine.Resource{
				ID:   "test-4",
				Name: "ab",
				Labels: map[string]string{
					"env":   "development",
					"owner": "test-team",
				},
			},
			expectAllowed:   false,
			expectViolation: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := eng.EvaluateResource(context.Background(), tt.resource)
			if err != nil {
				t.Fatalf("Evaluation failed: %v", err)
			}

			if result.Allowed != tt.expectAllowed {
				t.Errorf("Expected allowed=%v, got %v", tt.expectAllowed, result.Allowed)
			}

			hasViolation := len(result.Violations) > 0
			if hasViolation != tt.expectViolation {
				t.Errorf("Expected violation=%v, got %v violations: %+v",
					tt.expectViolation, hasViolation, result.Violations)
			}
		})
	}
}

func TestEvaluateResource_RequiredLabels(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	eng, err := NewEngine(logger)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	tests := []struct {
		name            string
		resource        *engine.Resource
		expectAllowed   bool
		expectViolation bool
	}{
		{
			name: "all required labels present",
			resource: &engine.Resource{
				ID:   "test-1",
				Name: "test-resource",
				Labels: map[string]string{
					"env":   "production",
					"owner": "platform-team",
				},
			},
			expectAllowed:   true,
			expectViolation: false,
		},
		{
			name: "missing env label",
			resource: &engine.Resource{
				ID:   "test-2",
				Name: "test-resource",
				Labels: map[string]string{
					"owner": "platform-team",
				},
			},
			expectAllowed:   false,
			expectViolation: true,
		},
		{
			name: "missing owner label",
			resource: &engine.Resource{
				ID:   "test-3",
				Name: "test-resource",
				Labels: map[string]string{
					"env": "production",
				},
			},
			expectAllowed:   false,
			expectViolation: true,
		},
		{
			name: "invalid env value",
			resource: &engine.Resource{
				ID:   "test-4",
				Name: "test-resource",
				Labels: map[string]string{
					"env":   "invalid",
					"owner": "platform-team",
				},
			},
			expectAllowed:   false,
			expectViolation: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := eng.EvaluateResource(context.Background(), tt.resource)
			if err != nil {
				t.Fatalf("Evaluation failed: %v", err)
			}

			if result.Allowed != tt.expectAllowed {
				t.Errorf("Expected allowed=%v, got %v. Violations: %+v",
					tt.expectAllowed, result.Allowed, result.Violations)
			}

			hasViolation := len(result.Violations) > 0
			if hasViolation != tt.expectViolation {
				t.Errorf("Expected violation=%v, got %v violations: %+v",
					tt.expectViolation, hasViolation, result.Violations)
			}
		})
	}
}

func TestEvaluatePlan_OperationRestrictions(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	eng, err := NewEngine(logger)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	plan := &engine.Plan{
		ID:    "test-plan",
		RunID: "test-run",
		Units: []engine.PlanUnit{
			{
				ID:              "unit-1",
				ResourceID:      "resource-1",
				Operation:       "delete",
				ProviderName:    "linux",
				ProviderVersion: "1.0.0",
			},
		},
	}

	// Test with production context
	// Note: Since we can't directly pass context to the plan,
	// this test demonstrates the structure
	result, err := eng.EvaluatePlan(context.Background(), plan)
	if err != nil {
		t.Fatalf("Evaluation failed: %v", err)
	}

	// The policy should evaluate, but without context it won't trigger
	// the production restriction
	if result == nil {
		t.Fatal("Result is nil")
	}
}

func TestEnableDisablePolicy(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	eng, err := NewEngine(logger)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	policyName := "resource-naming"

	// Disable the policy
	err = eng.DisablePolicy(policyName)
	if err != nil {
		t.Fatalf("Failed to disable policy: %v", err)
	}

	policy, err := eng.GetPolicy(policyName)
	if err != nil {
		t.Fatalf("Failed to get policy: %v", err)
	}

	if policy.Enabled {
		t.Error("Policy should be disabled")
	}

	// Create a resource with invalid name
	resource := &engine.Resource{
		ID:   "test-1",
		Name: "INVALID_NAME",
		Labels: map[string]string{
			"env":   "development",
			"owner": "test-team",
		},
	}

	// Evaluate - should pass because policy is disabled
	result, err := eng.EvaluateResource(context.Background(), resource)
	if err != nil {
		t.Fatalf("Evaluation failed: %v", err)
	}

	// Should have no violations from the disabled policy
	for _, v := range result.Violations {
		if v.Policy == policyName {
			t.Error("Disabled policy should not generate violations")
		}
	}

	// Re-enable the policy
	err = eng.EnablePolicy(policyName)
	if err != nil {
		t.Fatalf("Failed to enable policy: %v", err)
	}

	policy, err = eng.GetPolicy(policyName)
	if err != nil {
		t.Fatalf("Failed to get policy: %v", err)
	}

	if !policy.Enabled {
		t.Error("Policy should be enabled")
	}
}

func TestStateDriftPolicy(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	eng, err := NewEngine(logger)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	desired := map[string]interface{}{
		"version": "1.0.0",
		"enabled": true,
		"config": map[string]interface{}{
			"timeout": 30,
			"retries": 3,
		},
	}

	actual := map[string]interface{}{
		"version": "1.0.1", // Different
		"enabled": true,
		"config": map[string]interface{}{
			"timeout": 60, // Different
			"retries": 3,
		},
	}

	desiredJSON, _ := json.Marshal(desired)
	actualJSON, _ := json.Marshal(actual)

	resource := &engine.Resource{
		ID:     "test-drift",
		Name:   "test-resource",
		Config: desiredJSON,
		State:  actualJSON,
		Labels: map[string]string{
			"env":   "production",
			"owner": "platform-team",
		},
	}

	result, err := eng.EvaluateResource(context.Background(), resource)
	if err != nil {
		t.Fatalf("Evaluation failed: %v", err)
	}

	// Check if drift policy was evaluated
	t.Logf("Violations: %+v", result.Violations)
	t.Logf("Allowed: %v", result.Allowed)
}

func TestReloadPolicies(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	eng, err := NewEngine(logger)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	initialCount := len(eng.ListPolicies())

	// Reload policies
	err = eng.ReloadPolicies(context.Background())
	if err != nil {
		t.Fatalf("Failed to reload policies: %v", err)
	}

	afterReloadCount := len(eng.ListPolicies())

	if initialCount != afterReloadCount {
		t.Errorf("Expected %d policies after reload, got %d", initialCount, afterReloadCount)
	}
}

func TestListPolicies(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	eng, err := NewEngine(logger)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	policies := eng.ListPolicies()

	if len(policies) == 0 {
		t.Fatal("No policies returned")
	}

	// Check that all policies have required fields
	for _, p := range policies {
		if p.Name == "" {
			t.Error("Policy has empty name")
		}
		if p.Rego == "" {
			t.Error("Policy has empty Rego code")
		}
		if p.CreatedAt.IsZero() {
			t.Error("Policy has zero CreatedAt")
		}
	}
}

func TestEvaluateConfig(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	eng, err := NewEngine(logger)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	config := &engine.Config{
		ID:     "test-config",
		Source: "test",
		Resources: []engine.Resource{
			{
				ID:   "resource-1",
				Name: "valid-name",
				Labels: map[string]string{
					"env":   "production",
					"owner": "platform-team",
				},
			},
			{
				ID:   "resource-2",
				Name: "INVALID-NAME", // Uppercase - should violate naming policy
				Labels: map[string]string{
					"env":   "production",
					"owner": "platform-team",
				},
			},
		},
		ParsedAt: time.Now(),
	}

	result, err := eng.Evaluate(context.Background(), config)
	if err != nil {
		t.Fatalf("Evaluation failed: %v", err)
	}

	if result.Allowed {
		t.Error("Expected config to be rejected due to naming violation")
	}

	if len(result.Violations) == 0 {
		t.Error("Expected at least one violation")
	}

	// Check that we got a naming violation
	foundNamingViolation := false
	for _, v := range result.Violations {
		if v.Policy == "resource-naming" {
			foundNamingViolation = true
			break
		}
	}

	if !foundNamingViolation {
		t.Error("Expected a naming policy violation")
	}
}
