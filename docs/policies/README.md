# OpenFroyo Policy System

The OpenFroyo policy system uses Open Policy Agent (OPA) and the Rego policy language to enforce governance, compliance, and best practices for infrastructure management.

## Table of Contents

- [Overview](#overview)
- [Built-in Policies](#built-in-policies)
- [Custom Policies](#custom-policies)
- [Policy Evaluation](#policy-evaluation)
- [Examples](#examples)
- [Best Practices](#best-practices)

## Overview

Policies in OpenFroyo are evaluated at multiple points in the workflow:

1. **Configuration Validation** - Before planning
2. **Plan Evaluation** - Before execution
3. **Resource Evaluation** - During create/update operations
4. **Drift Detection** - After state comparison

### Severity Levels

Policies can generate violations at four severity levels:

- `info` - Informational messages
- `warning` - Issues that should be reviewed but don't block operations
- `error` - Issues that block operations
- `critical` - Severe issues requiring immediate attention

## Built-in Policies

OpenFroyo includes five built-in policies that are enabled by default:

### 1. Resource Naming (`resource-naming`)

**Purpose**: Enforces consistent resource naming conventions

**Rules**:
- Names must be lowercase
- Only alphanumeric characters and hyphens allowed
- Must not start or end with a hyphen
- Length between 3 and 63 characters

**Severity**: `error`

**Example violation**:
```
Resource name 'Invalid_Name' must contain only lowercase letters, numbers, and hyphens
```

### 2. Required Labels (`required-labels`)

**Purpose**: Ensures critical labels are present on all resources

**Required labels**:
- `env` - Environment (must be: development, staging, production, or test)
- `owner` - Resource owner/team

**Severity**: `error`

**Example violation**:
```
Resource web-server missing required label: owner
```

### 3. State Drift (`state-drift`)

**Purpose**: Enforces maximum acceptable state drift threshold

**Rules**:
- Maximum drift threshold: 10%
- Calculated based on differing fields between desired and actual state

**Severity**: `warning`

**Example violation**:
```
State drift of 15.5% exceeds maximum threshold of 10%
```

### 4. Operation Restrictions (`operation-restrictions`)

**Purpose**: Prevents destructive operations in production

**Rules**:
- Blocks `delete` and `recreate` operations in production
- Prevents deletion of resources marked as `critical: "true"`
- Warns when more than 5 resources will be deleted

**Severity**: `critical`

**Example violation**:
```
Destructive operation 'delete' is not allowed in production environment
```

### 5. Provider Versioning (`provider-versioning`)

**Purpose**: Enforces minimum provider versions for security

**Rules**:
- Warns if provider version is not specified
- Enforces minimum versions for known providers
- Warns about pre-release versions in production

**Severity**: `warning`

**Example violation**:
```
Provider linux version 0.9.0 is below minimum required version 1.0.0
```

## Custom Policies

### Writing Custom Policies

Custom policies are written in Rego. Here's the basic structure:

```rego
package custom.policies.myteam

import rego.v1

deny contains violation if {
    input.resource
    resource := input.resource

    # Your policy logic here
    condition_that_should_deny

    violation := {
        "message": "Description of what went wrong",
        "severity": "error",
        "resource": resource.id,
    }
}
```

### Available Input Data

Policies receive different input based on the evaluation context:

**Resource Evaluation**:
```json
{
  "resource": {
    "id": "resource-id",
    "name": "resource-name",
    "type": "resource-type",
    "config": {},
    "state": {},
    "labels": {},
    "annotations": {}
  },
  "desired_state": {},
  "actual_state": {},
  "context": {
    "user": "username",
    "environment": "production",
    "timestamp": "2024-01-01T00:00:00Z",
    "operation": "create",
    "dry_run": false
  }
}
```

**Plan Evaluation**:
```json
{
  "plan": {
    "id": "plan-id",
    "units": [
      {
        "id": "unit-id",
        "resource_id": "resource-id",
        "operation": "create",
        "provider_name": "linux",
        "provider_version": "1.0.0"
      }
    ]
  },
  "context": {
    "user": "username",
    "environment": "production"
  }
}
```

## Policy Evaluation

### Using the Policy Engine

```go
import (
    "context"
    "github.com/openfroyo/openfroyo/pkg/policy"
    "github.com/rs/zerolog"
)

// Create engine
logger := zerolog.New(os.Stdout)
engine, err := policy.NewEngine(logger)
if err != nil {
    log.Fatal(err)
}

// Evaluate a resource
resource := &engine.Resource{
    ID:   "web-server",
    Name: "web-server-prod",
    Labels: map[string]string{
        "env":   "production",
        "owner": "platform-team",
    },
}

result, err := engine.EvaluateResource(ctx, resource)
if err != nil {
    log.Fatal(err)
}

if !result.Allowed {
    for _, violation := range result.Violations {
        fmt.Printf("[%s] %s: %s\n",
            violation.Severity,
            violation.Policy,
            violation.Message)
    }
}
```

### Loading Custom Policies

```go
// Load from files or directories
paths := []string{
    "/etc/froyo/policies",
    "/opt/policies/custom.rego",
}

err = engine.LoadPolicies(ctx, paths)
if err != nil {
    log.Fatal(err)
}
```

### Hot Reload

```go
loader := policy.NewLoader(logger)
err = loader.Watch(ctx, paths, func(policies []policy.Policy) error {
    return engine.LoadPolicies(ctx, paths)
})
```

### Managing Policies

```go
// List all policies
policies := engine.ListPolicies()

// Get specific policy
policy, err := engine.GetPolicy("resource-naming")

// Disable a policy
err = engine.DisablePolicy("state-drift")

// Enable a policy
err = engine.EnablePolicy("state-drift")
```

## Examples

### Example 1: Backup Label Requirement

Require all production resources to have a backup schedule:

```rego
package custom.policies.backup

import rego.v1

deny contains violation if {
    input.resource
    resource := input.resource

    # Check if production environment
    resource.labels.env == "production"

    # Verify backup label exists
    not resource.labels.backup

    violation := {
        "message": sprintf("Production resource %s must have a backup label", [resource.id]),
        "severity": "error",
        "resource": resource.id,
    }
}

deny contains violation if {
    input.resource
    resource := input.resource

    resource.labels.env == "production"
    resource.labels.backup != ""

    # Validate backup schedule format
    not regex.match(`^(daily|weekly|monthly)$`, resource.labels.backup)

    violation := {
        "message": sprintf("Invalid backup schedule '%s' for resource %s (must be: daily, weekly, or monthly)",
            [resource.labels.backup, resource.id]),
        "severity": "error",
        "resource": resource.id,
    }
}
```

### Example 2: Cost Center Requirement

Require cost center labels for tracking:

```rego
package custom.policies.costcenter

import rego.v1

deny contains violation if {
    input.resource
    resource := input.resource

    # Require cost-center label
    not resource.labels["cost-center"]

    violation := {
        "message": sprintf("Resource %s must have a cost-center label for billing", [resource.id]),
        "severity": "warning",
        "resource": resource.id,
    }
}

deny contains violation if {
    input.resource
    resource := input.resource

    cost_center := resource.labels["cost-center"]
    cost_center != ""

    # Validate cost center format (e.g., CC-12345)
    not regex.match(`^CC-\d{5}$`, cost_center)

    violation := {
        "message": sprintf("Invalid cost-center format '%s' (must be CC-XXXXX)", [cost_center]),
        "severity": "warning",
        "resource": resource.id,
    }
}
```

### Example 3: Security Group Restrictions

Prevent overly permissive security configurations:

```rego
package custom.policies.security

import rego.v1

deny contains violation if {
    input.resource
    resource := input.resource

    # Check if this is a security-related resource
    contains(resource.type, "security")

    # Parse config
    config := json.unmarshal(resource.config)

    # Check for wildcard access
    some rule in config.rules
    rule.source == "0.0.0.0/0"
    rule.port in [22, 3389]  # SSH or RDP

    violation := {
        "message": sprintf("Resource %s allows public access to sensitive port %d",
            [resource.id, rule.port]),
        "severity": "critical",
        "resource": resource.id,
    }
}
```

### Example 4: Team-Based Authorization

Ensure teams can only modify their own resources:

```rego
package custom.policies.authorization

import rego.v1

deny contains violation if {
    input.resource
    input.context

    resource := input.resource
    context := input.context

    # Check if operation is destructive
    context.operation in ["delete", "recreate"]

    # Get user's team from context
    user_team := context.metadata.team

    # Get resource owner
    resource_owner := resource.labels.owner

    # Verify team matches
    user_team != resource_owner

    violation := {
        "message": sprintf("User from team '%s' cannot perform '%s' on resource owned by '%s'",
            [user_team, context.operation, resource_owner]),
        "severity": "critical",
        "resource": resource.id,
    }
}
```

### Example 5: Compliance Tags

Ensure compliance-related tags are present:

```rego
package custom.policies.compliance

import rego.v1

# Required compliance tags
compliance_tags := ["data-classification", "retention-period", "encryption"]

deny contains violation if {
    input.resource
    resource := input.resource
    some tag in compliance_tags

    # Check if tag is present
    not resource.labels[tag]

    violation := {
        "message": sprintf("Resource %s missing required compliance tag: %s",
            [resource.id, tag]),
        "severity": "error",
        "resource": resource.id,
    }
}

# Validate data classification values
deny contains violation if {
    input.resource
    resource := input.resource

    classification := resource.labels["data-classification"]
    classification != ""

    # Must be one of the allowed classifications
    not classification in ["public", "internal", "confidential", "restricted"]

    violation := {
        "message": sprintf("Invalid data-classification '%s' (must be: public, internal, confidential, or restricted)",
            [classification]),
        "severity": "error",
        "resource": resource.id,
    }
}
```

## Best Practices

### 1. Policy Organization

- Group related policies in packages (e.g., `custom.policies.security`)
- Use descriptive policy names
- Include comments explaining policy intent
- Version your policy files

### 2. Performance

- Keep policies focused and specific
- Avoid complex computations in policies
- Use helper functions for reusable logic
- Test policies with realistic data volumes

### 3. Severity Guidelines

- `info` - FYI messages, recommendations
- `warning` - Issues that should be fixed but don't block deployment
- `error` - Issues that should block deployment
- `critical` - Security or compliance issues

### 4. Error Messages

- Be specific about what's wrong
- Provide actionable remediation steps
- Include resource IDs for context
- Use clear, non-technical language when possible

### 5. Testing

- Write unit tests for custom policies
- Test with both valid and invalid inputs
- Verify performance with large datasets
- Test edge cases and boundary conditions

### 6. Documentation

- Document all custom policies
- Explain the business rationale
- Provide examples of compliant and non-compliant resources
- Keep documentation updated with policy changes

## Policy File Structure

### Rego File (.rego)

```rego
# Policy description goes here
# Can span multiple lines
package custom.policies.myteam

import rego.v1

# Helper functions
is_production if {
    input.context.environment == "production"
}

# Policy rules
deny contains violation if {
    # Conditions
    # ...

    violation := {
        "message": "Error message",
        "severity": "error",
        "resource": resource.id,
    }
}
```

### JSON Policy Definition (.json)

```json
{
  "name": "my-custom-policy",
  "description": "Enforces custom business rules",
  "rego": "package custom.policies.myteam\n...",
  "severity": "error",
  "enabled": true,
  "tags": ["custom", "business-rules"],
  "metadata": {
    "owner": "platform-team",
    "version": "1.0.0"
  }
}
```

## Policy Bundle Structure

```json
{
  "name": "team-policies",
  "version": "1.2.0",
  "description": "Policies for the platform team",
  "policies": [
    {
      "name": "policy1",
      "description": "First policy",
      "rego": "package ...",
      "severity": "error",
      "enabled": true
    },
    {
      "name": "policy2",
      "description": "Second policy",
      "rego": "package ...",
      "severity": "warning",
      "enabled": true
    }
  ]
}
```

## Troubleshooting

### Policy Not Triggering

1. Check if policy is enabled: `engine.GetPolicy("policy-name")`
2. Verify policy syntax is valid
3. Check input data structure matches policy expectations
4. Enable debug logging to see evaluation details

### Performance Issues

1. Profile policy evaluation time
2. Simplify complex policy logic
3. Use prepared queries (done automatically)
4. Consider caching frequently accessed data

### Unexpected Violations

1. Review input data structure
2. Check severity levels
3. Verify policy logic with test cases
4. Enable verbose logging

## Additional Resources

- [OPA Documentation](https://www.openpolicyagent.org/docs/latest/)
- [Rego Language Reference](https://www.openpolicyagent.org/docs/latest/policy-language/)
- [OPA Playground](https://play.openpolicyagent.org/) - Test policies online
- OpenFroyo Policy Package Documentation (see `pkg/policy/doc.go`)
