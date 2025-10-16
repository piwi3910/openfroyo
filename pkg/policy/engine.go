package policy

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/storage"
	"github.com/open-policy-agent/opa/storage/inmem"
	"github.com/openfroyo/openfroyo/pkg/engine"
	"github.com/rs/zerolog"
)

// Engine implements the PolicyEngine interface from pkg/engine/interfaces.go.
type Engine struct {
	mu           sync.RWMutex
	policies     map[string]*compiledPolicy
	store        storage.Store
	logger       zerolog.Logger
	compiler     *ast.Compiler
	builtinPolicies []Policy
}

// compiledPolicy represents a compiled Rego policy.
type compiledPolicy struct {
	policy   *Policy
	module   *ast.Module
	query    rego.PreparedEvalQuery
	compiled time.Time
}

// NewEngine creates a new policy engine.
func NewEngine(logger zerolog.Logger) (*Engine, error) {
	store := inmem.New()

	e := &Engine{
		policies:     make(map[string]*compiledPolicy),
		store:        store,
		logger:       logger.With().Str("component", "policy-engine").Logger(),
		builtinPolicies: GetBuiltinPolicies(),
	}

	// Load built-in policies
	if err := e.loadBuiltinPolicies(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to load built-in policies: %w", err)
	}

	return e, nil
}

// Evaluate evaluates policies against a configuration.
func (e *Engine) Evaluate(ctx context.Context, config *engine.Config) (*engine.PolicyResult, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var allViolations []engine.PolicyViolation
	var warnings []string
	evaluatedPolicies := make([]string, 0, len(e.policies))

	for _, cp := range e.policies {
		if !cp.policy.Enabled {
			continue
		}

		evaluatedPolicies = append(evaluatedPolicies, cp.policy.Name)

		// Evaluate each resource in the config
		for i := range config.Resources {
			input := &PolicyInput{
				Resource: &config.Resources[i],
				Context: &PolicyContext{
					Timestamp: time.Now(),
					Operation: "validate",
				},
			}

			violations, err := e.evaluatePolicy(ctx, cp, input)
			if err != nil {
				e.logger.Error().Err(err).
					Str("policy", cp.policy.Name).
					Str("resource", config.Resources[i].ID).
					Msg("Policy evaluation failed")
				warnings = append(warnings, fmt.Sprintf("Policy %s evaluation failed: %v", cp.policy.Name, err))
				continue
			}

			allViolations = append(allViolations, violations...)
		}
	}

	// Determine if allowed based on violations
	allowed := true
	for i := range allViolations {
		if allViolations[i].Severity == string(SeverityError) || allViolations[i].Severity == string(SeverityCritical) {
			allowed = false
			break
		}
	}

	return &engine.PolicyResult{
		Allowed:    allowed,
		Violations: allViolations,
		Warnings:   warnings,
		EvaluatedAt: time.Now(),
	}, nil
}

// EvaluatePlan evaluates policies against a plan.
func (e *Engine) EvaluatePlan(ctx context.Context, plan *engine.Plan) (*engine.PolicyResult, error) {
	startTime := time.Now()
	e.mu.RLock()
	defer e.mu.RUnlock()

	var allViolations []engine.PolicyViolation
	var warnings []string
	evaluatedPolicies := make([]string, 0, len(e.policies))

	for _, cp := range e.policies {
		if !cp.policy.Enabled {
			continue
		}

		evaluatedPolicies = append(evaluatedPolicies, cp.policy.Name)

		input := &PolicyInput{
			Plan: plan,
			Context: &PolicyContext{
				Timestamp: time.Now(),
				Operation: "plan",
			},
		}

		violations, err := e.evaluatePolicy(ctx, cp, input)
		if err != nil {
			e.logger.Error().Err(err).
				Str("policy", cp.policy.Name).
				Str("plan", plan.ID).
				Msg("Policy evaluation failed")
			warnings = append(warnings, fmt.Sprintf("Policy %s evaluation failed: %v", cp.policy.Name, err))
			continue
		}

		allViolations = append(allViolations, violations...)
	}

	// Determine if allowed
	allowed := true
	for i := range allViolations {
		if allViolations[i].Severity == string(SeverityError) || allViolations[i].Severity == string(SeverityCritical) {
			allowed = false
			break
		}
	}

	duration := time.Since(startTime)
	e.logger.Debug().
		Str("plan_id", plan.ID).
		Int("violations", len(allViolations)).
		Dur("duration", duration).
		Msg("Plan policy evaluation completed")

	return &engine.PolicyResult{
		Allowed:    allowed,
		Violations: allViolations,
		Warnings:   warnings,
		EvaluatedAt: time.Now(),
	}, nil
}

// EvaluateResource evaluates policies against a single resource.
func (e *Engine) EvaluateResource(ctx context.Context, resource *engine.Resource) (*engine.PolicyResult, error) {
	startTime := time.Now()
	e.mu.RLock()
	defer e.mu.RUnlock()

	var allViolations []engine.PolicyViolation
	var warnings []string
	evaluatedPolicies := make([]string, 0, len(e.policies))

	for _, cp := range e.policies {
		if !cp.policy.Enabled {
			continue
		}

		evaluatedPolicies = append(evaluatedPolicies, cp.policy.Name)

		input := &PolicyInput{
			Resource: resource,
			Context: &PolicyContext{
				Timestamp: time.Now(),
				Operation: "validate",
			},
		}

		violations, err := e.evaluatePolicy(ctx, cp, input)
		if err != nil {
			e.logger.Error().Err(err).
				Str("policy", cp.policy.Name).
				Str("resource", resource.ID).
				Msg("Policy evaluation failed")
			warnings = append(warnings, fmt.Sprintf("Policy %s evaluation failed: %v", cp.policy.Name, err))
			continue
		}

		allViolations = append(allViolations, violations...)
	}

	// Determine if allowed
	allowed := true
	for i := range allViolations {
		if allViolations[i].Severity == string(SeverityError) || allViolations[i].Severity == string(SeverityCritical) {
			allowed = false
			break
		}
	}

	duration := time.Since(startTime)
	e.logger.Debug().
		Str("resource_id", resource.ID).
		Int("violations", len(allViolations)).
		Dur("duration", duration).
		Msg("Resource policy evaluation completed")

	return &engine.PolicyResult{
		Allowed:    allowed,
		Violations: allViolations,
		Warnings:   warnings,
		EvaluatedAt: time.Now(),
	}, nil
}

// LoadPolicies loads policy files.
func (e *Engine) LoadPolicies(ctx context.Context, paths []string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	loader := NewLoader(e.logger)
	policies, err := loader.LoadFromPaths(ctx, paths)
	if err != nil {
		return fmt.Errorf("failed to load policies: %w", err)
	}

	// Compile and store policies
	for i := range policies {
		if err := e.compileAndStorePolicy(ctx, &policies[i]); err != nil {
			e.logger.Error().Err(err).
				Str("policy", policies[i].Name).
				Msg("Failed to compile policy")
			return fmt.Errorf("failed to compile policy %s: %w", policies[i].Name, err)
		}
	}

	e.logger.Info().
		Int("count", len(policies)).
		Msg("Policies loaded successfully")

	return nil
}

// evaluatePolicy evaluates a single compiled policy.
func (e *Engine) evaluatePolicy(ctx context.Context, cp *compiledPolicy, input *PolicyInput) ([]engine.PolicyViolation, error) {
	// Build the query to get all deny violations from the policy package
	// Extract package name from the policy
	packageName := extractPackageName(cp.policy.Rego)

	// Create a query specifically for deny results
	query := fmt.Sprintf("data.%s.deny", packageName)

	r := rego.New(
		rego.Module(cp.policy.Name, cp.policy.Rego),
		rego.Query(query),
		rego.Input(input),
	)

	results, err := r.Eval(ctx)
	if err != nil {
		return nil, fmt.Errorf("policy evaluation error: %w", err)
	}

	var violations []engine.PolicyViolation

	// Process results
	for _, result := range results {
		if len(result.Expressions) > 0 {
			// The result should be a set of violations
			if denySet, ok := result.Expressions[0].Value.([]interface{}); ok {
				for _, d := range denySet {
					violation := e.createViolation(cp.policy, d, input)
					violations = append(violations, violation)
				}
			}
		}
	}

	return violations, nil
}

// extractPackageName extracts the package name from Rego code.
func extractPackageName(rego string) string {
	lines := strings.Split(rego, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "package ") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}
	return "openfroyo.policies"
}

// createViolation creates a PolicyViolation from policy result.
func (e *Engine) createViolation(policy *Policy, result interface{}, input *PolicyInput) engine.PolicyViolation {
	violation := engine.PolicyViolation{
		Policy:   policy.Name,
		Severity: string(policy.Severity),
	}

	if input.Resource != nil {
		violation.ResourceID = input.Resource.ID
	}

	// Extract message from result
	switch v := result.(type) {
	case string:
		violation.Message = v
	case map[string]interface{}:
		if msg, ok := v["message"].(string); ok {
			violation.Message = msg
		}
		if sev, ok := v["severity"].(string); ok {
			violation.Severity = sev
		}
		if res, ok := v["resource"].(string); ok {
			violation.ResourceID = res
		}
	default:
		violation.Message = fmt.Sprintf("%v", result)
	}

	return violation
}

// compileAndStorePolicy compiles a policy and stores it.
func (e *Engine) compileAndStorePolicy(ctx context.Context, policy *Policy) error {
	// Parse the Rego module
	module, err := ast.ParseModule(policy.Name, policy.Rego)
	if err != nil {
		return fmt.Errorf("failed to parse policy: %w", err)
	}

	// Create a new Rego query
	r := rego.New(
		rego.Module(policy.Name, policy.Rego),
		rego.Store(e.store),
		rego.Query("data"),
	)

	// Prepare the query for reuse
	query, err := r.PrepareForEval(ctx)
	if err != nil {
		return fmt.Errorf("failed to prepare query: %w", err)
	}

	e.policies[policy.Name] = &compiledPolicy{
		policy:   policy,
		module:   module,
		query:    query,
		compiled: time.Now(),
	}

	e.logger.Debug().
		Str("policy", policy.Name).
		Msg("Policy compiled successfully")

	return nil
}

// loadBuiltinPolicies loads the built-in policies.
func (e *Engine) loadBuiltinPolicies(ctx context.Context) error {
	for i := range e.builtinPolicies {
		if err := e.compileAndStorePolicy(ctx, &e.builtinPolicies[i]); err != nil {
			return fmt.Errorf("failed to compile built-in policy %s: %w", e.builtinPolicies[i].Name, err)
		}
	}

	e.logger.Info().
		Int("count", len(e.builtinPolicies)).
		Msg("Built-in policies loaded")

	return nil
}

// GetPolicy returns a policy by name.
func (e *Engine) GetPolicy(name string) (*Policy, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	cp, exists := e.policies[name]
	if !exists {
		return nil, fmt.Errorf("policy not found: %s", name)
	}

	return cp.policy, nil
}

// ListPolicies returns all loaded policies.
func (e *Engine) ListPolicies() []Policy {
	e.mu.RLock()
	defer e.mu.RUnlock()

	policies := make([]Policy, 0, len(e.policies))
	for _, cp := range e.policies {
		policies = append(policies, *cp.policy)
	}

	return policies
}

// ReloadPolicies reloads all policies.
func (e *Engine) ReloadPolicies(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Clear existing policies
	e.policies = make(map[string]*compiledPolicy)

	// Reload built-in policies
	return e.loadBuiltinPolicies(ctx)
}

// EnablePolicy enables a policy by name.
func (e *Engine) EnablePolicy(name string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	cp, exists := e.policies[name]
	if !exists {
		return fmt.Errorf("policy not found: %s", name)
	}

	cp.policy.Enabled = true
	e.logger.Info().Str("policy", name).Msg("Policy enabled")

	return nil
}

// DisablePolicy disables a policy by name.
func (e *Engine) DisablePolicy(name string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	cp, exists := e.policies[name]
	if !exists {
		return fmt.Errorf("policy not found: %s", name)
	}

	cp.policy.Enabled = false
	e.logger.Info().Str("policy", name).Msg("Policy disabled")

	return nil
}
