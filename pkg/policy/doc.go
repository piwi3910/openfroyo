// Package policy provides Open Policy Agent (OPA) integration for OpenFroyo.
//
// This package implements policy enforcement for infrastructure configurations,
// plans, and operations using the Rego policy language. It includes built-in
// policies for common governance requirements and supports custom policy loading.
//
// # Architecture
//
// The policy system consists of four main components:
//
//  1. Engine - Compiles and evaluates Rego policies
//  2. Loader - Loads policies from files, directories, and bundles
//  3. Types - Data structures for policies, violations, and results
//  4. Built-in Policies - Pre-defined policies for common requirements
//
// # Usage
//
// Creating a policy engine:
//
//	logger := zerolog.New(os.Stdout)
//	engine, err := policy.NewEngine(logger)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Evaluating a resource:
//
//	resource := &engine.Resource{
//	    ID:   "web-server",
//	    Name: "web-server-prod",
//	    Labels: map[string]string{
//	        "env":   "production",
//	        "owner": "platform-team",
//	    },
//	}
//
//	result, err := engine.EvaluateResource(ctx, resource)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	if !result.Allowed {
//	    for _, violation := range result.Violations {
//	        fmt.Printf("Policy %s violated: %s\n", violation.Policy, violation.Message)
//	    }
//	}
//
// Loading custom policies:
//
//	paths := []string{
//	    "/etc/froyo/policies",
//	    "/opt/policies/custom.rego",
//	}
//
//	err = engine.LoadPolicies(ctx, paths)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Built-in Policies
//
// The following policies are included by default:
//
//  1. resource-naming - Enforces resource naming conventions
//  2. required-labels - Ensures critical labels (env, owner) are present
//  3. state-drift - Maximum acceptable drift threshold
//  4. operation-restrictions - Prevents destructive operations in production
//  5. provider-versioning - Enforces minimum provider versions
//
// # Custom Policies
//
// Custom policies can be written in Rego and loaded from files:
//
//	package custom.policies.backup
//
//	import rego.v1
//
//	deny contains violation if {
//	    input.resource
//	    resource := input.resource
//
//	    # Require backup label for production resources
//	    resource.labels.env == "production"
//	    not resource.labels.backup
//
//	    violation := {
//	        "message": "Production resources must have a backup label",
//	        "severity": "error",
//	        "resource": resource.id,
//	    }
//	}
//
// # Policy Evaluation Points
//
// Policies are evaluated at multiple points in the OpenFroyo workflow:
//
//  1. Configuration validation - Before planning
//  2. Plan evaluation - Before execution
//  3. Resource evaluation - During create/update operations
//  4. Drift detection - After state comparison
//
// # Severity Levels
//
// Violations have four severity levels:
//
//  - info: Informational messages
//  - warning: Issues that should be reviewed but don't block operations
//  - error: Issues that block operations
//  - critical: Severe issues requiring immediate attention
//
// # Hot Reload
//
// The loader supports watching policy files for changes and reloading automatically:
//
//	loader := policy.NewLoader(logger)
//	err = loader.Watch(ctx, paths, func(policies []policy.Policy) error {
//	    return engine.LoadPolicies(ctx, paths)
//	})
//
// # Performance
//
// Policies are compiled once and reused for multiple evaluations. The engine
// uses OPA's PreparedEvalQuery for optimal performance. Caching is implemented
// at both the loader and engine levels.
//
// # Context Injection
//
// Policy evaluations can include context information:
//
//  - User: Who initiated the operation
//  - Environment: Target environment (production, staging, etc.)
//  - Operation: Type of operation (create, update, delete)
//  - Timestamp: When the evaluation occurred
//  - Dry run: Whether this is a dry-run evaluation
//
// This context allows policies to make environment-aware decisions.
package policy
