package policy

import (
	"time"
)

// GetBuiltinPolicies returns all built-in policies.
func GetBuiltinPolicies() []Policy {
	return []Policy{
		resourceNamingPolicy(),
		requiredLabelsPolicy(),
		stateDriftPolicy(),
		operationRestrictionsPolicy(),
		providerVersioningPolicy(),
	}
}

// resourceNamingPolicy enforces resource naming conventions.
func resourceNamingPolicy() Policy {
	return Policy{
		Name:        "resource-naming",
		Description: "Enforces resource naming conventions (lowercase, alphanumeric, hyphens only)",
		Severity:    SeverityError,
		Enabled:     true,
		Tags:        []string{"naming", "conventions"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Rego: `package openfroyo.policies.naming

import rego.v1

# Resource naming must follow conventions
deny contains violation if {
	input.resource
	resource := input.resource

	# Check if name exists
	not resource.name
	violation := {
		"message": sprintf("Resource %s must have a name", [resource.id]),
		"severity": "error",
		"resource": resource.id,
	}
}

deny contains violation if {
	input.resource
	resource := input.resource
	name := resource.name

	# Name must be lowercase
	lower(name) != name
	violation := {
		"message": sprintf("Resource name '%s' must be lowercase", [name]),
		"severity": "error",
		"resource": resource.id,
	}
}

deny contains violation if {
	input.resource
	resource := input.resource
	name := resource.name

	# Name must match pattern: alphanumeric and hyphens only
	not regex.match("^[a-z0-9-]+$", name)
	violation := {
		"message": sprintf("Resource name '%s' must contain only lowercase letters, numbers, and hyphens", [name]),
		"severity": "error",
		"resource": resource.id,
	}
}

deny contains violation if {
	input.resource
	resource := input.resource
	name := resource.name

	# Name must not start or end with hyphen
	regex.match("^-.*", name)
	violation := {
		"message": sprintf("Resource name '%s' must not start with a hyphen", [name]),
		"severity": "error",
		"resource": resource.id,
	}
}

deny contains violation if {
	input.resource
	resource := input.resource
	name := resource.name

	regex.match(".*-$", name)
	violation := {
		"message": sprintf("Resource name '%s' must not end with a hyphen", [name]),
		"severity": "error",
		"resource": resource.id,
	}
}

deny contains violation if {
	input.resource
	resource := input.resource
	name := resource.name

	# Name must be between 3 and 63 characters
	count(name) < 3
	violation := {
		"message": sprintf("Resource name '%s' must be at least 3 characters long", [name]),
		"severity": "error",
		"resource": resource.id,
	}
}

deny contains violation if {
	input.resource
	resource := input.resource
	name := resource.name

	count(name) > 63
	violation := {
		"message": sprintf("Resource name '%s' must not exceed 63 characters", [name]),
		"severity": "error",
		"resource": resource.id,
	}
}`,
	}
}

// requiredLabelsPolicy ensures critical labels are present.
func requiredLabelsPolicy() Policy {
	return Policy{
		Name:        "required-labels",
		Description: "Ensures critical labels (env, owner) are present on all resources",
		Severity:    SeverityError,
		Enabled:     true,
		Tags:        []string{"labels", "metadata"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Rego: `package openfroyo.policies.labels

import rego.v1

required_labels := ["env", "owner"]

deny contains violation if {
	input.resource
	resource := input.resource

	# Check if labels exist
	not resource.labels
	violation := {
		"message": sprintf("Resource %s must have labels", [resource.id]),
		"severity": "error",
		"resource": resource.id,
	}
}

deny contains violation if {
	input.resource
	resource := input.resource
	some label in required_labels

	# Check if required label is present
	not resource.labels[label]
	violation := {
		"message": sprintf("Resource %s missing required label: %s", [resource.id, label]),
		"severity": "error",
		"resource": resource.id,
	}
}

deny contains violation if {
	input.resource
	resource := input.resource
	some label in required_labels

	# Check if required label has a value
	resource.labels[label] == ""
	violation := {
		"message": sprintf("Resource %s has empty required label: %s", [resource.id, label]),
		"severity": "error",
		"resource": resource.id,
	}
}

# Validate environment label values
deny contains violation if {
	input.resource
	resource := input.resource
	env := resource.labels.env

	# Must be one of the allowed environments
	not env in ["development", "staging", "production", "test"]
	violation := {
		"message": sprintf("Resource %s has invalid env label: %s (must be development, staging, production, or test)", [resource.id, env]),
		"severity": "error",
		"resource": resource.id,
	}
}`,
	}
}

// stateDriftPolicy enforces maximum acceptable drift threshold.
func stateDriftPolicy() Policy {
	return Policy{
		Name:        "state-drift",
		Description: "Enforces maximum acceptable state drift threshold",
		Severity:    SeverityWarning,
		Enabled:     true,
		Tags:        []string{"drift", "compliance"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Rego: `package openfroyo.policies.drift

import rego.v1

# Maximum drift threshold (percentage)
max_drift_threshold := 10

deny contains violation if {
	input.desired_state
	input.actual_state

	# Calculate drift percentage (simplified - counts differing fields)
	desired := input.desired_state
	actual := input.actual_state

	# Count total fields in desired state
	total_fields := count(object.keys(desired))

	# Count differing fields
	different_fields := count([k |
		some k in object.keys(desired)
		desired[k] != actual[k]
	])

	# Calculate drift percentage
	drift_percentage := (different_fields / total_fields) * 100
	drift_percentage > max_drift_threshold

	violation := {
		"message": sprintf("State drift of %.1f%% exceeds maximum threshold of %d%%", [drift_percentage, max_drift_threshold]),
		"severity": "warning",
		"resource": input.resource.id,
	}
}`,
	}
}

// operationRestrictionsPolicy prevents destructive operations in production.
func operationRestrictionsPolicy() Policy {
	return Policy{
		Name:        "operation-restrictions",
		Description: "Prevents destructive operations in production without approval",
		Severity:    SeverityCritical,
		Enabled:     true,
		Tags:        []string{"operations", "safety", "production"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Rego: `package openfroyo.policies.operations

import rego.v1

# Destructive operations that require approval
destructive_operations := ["delete", "recreate"]

deny contains violation if {
	input.context
	context := input.context

	# Check if operation is destructive
	some op in destructive_operations
	context.operation == op

	# Check if environment is production
	context.environment == "production"

	# Check if this is not a dry run
	not context.dry_run

	violation := {
		"message": sprintf("Destructive operation '%s' is not allowed in production environment", [op]),
		"severity": "critical",
		"resource": input.resource.id,
	}
}

# Prevent deletion of resources with critical label
deny contains violation if {
	input.resource
	input.context
	resource := input.resource
	context := input.context

	context.operation == "delete"
	resource.labels.critical == "true"

	violation := {
		"message": sprintf("Cannot delete resource %s marked as critical", [resource.id]),
		"severity": "critical",
		"resource": resource.id,
	}
}

# Warn about batch operations
deny contains violation if {
	input.plan
	plan := input.plan

	# Count resources to be deleted
	delete_count := count([u |
		some u in plan.units
		u.operation == "delete"
	])

	# Warn if more than 5 resources will be deleted
	delete_count > 5

	violation := {
		"message": sprintf("Plan will delete %d resources - please review carefully", [delete_count]),
		"severity": "warning",
	}
}`,
	}
}

// providerVersioningPolicy enforces minimum provider versions.
func providerVersioningPolicy() Policy {
	return Policy{
		Name:        "provider-versioning",
		Description: "Enforces minimum provider versions for security and compatibility",
		Severity:    SeverityWarning,
		Enabled:     true,
		Tags:        []string{"providers", "versioning", "security"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Rego: `package openfroyo.policies.providers

import rego.v1

# Minimum provider versions
min_provider_versions := {
	"linux": "1.0.0",
	"docker": "1.0.0",
	"kubernetes": "1.0.0",
}

deny contains violation if {
	input.plan
	plan := input.plan
	some unit in plan.units

	# Check if provider version is specified
	not unit.provider_version
	violation := {
		"message": sprintf("Plan unit %s does not specify provider version", [unit.id]),
		"severity": "warning",
		"resource": unit.resource_id,
	}
}

deny contains violation if {
	input.plan
	plan := input.plan
	some unit in plan.units

	# Get minimum version for this provider
	min_version := min_provider_versions[unit.provider_name]
	min_version

	# Simple version comparison (assumes semantic versioning)
	unit.provider_version < min_version

	violation := {
		"message": sprintf("Provider %s version %s is below minimum required version %s",
			[unit.provider_name, unit.provider_version, min_version]),
		"severity": "warning",
		"resource": unit.resource_id,
	}
}

# Warn about beta/alpha versions in production
deny contains violation if {
	input.plan
	input.context
	plan := input.plan
	context := input.context
	some unit in plan.units

	context.environment == "production"
	regex.match("(alpha|beta|rc)", unit.provider_version)

	violation := {
		"message": sprintf("Provider %s version %s is pre-release and should not be used in production",
			[unit.provider_name, unit.provider_version]),
		"severity": "warning",
		"resource": unit.resource_id,
	}
}`,
	}
}
