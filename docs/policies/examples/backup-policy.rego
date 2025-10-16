# Backup Policy
# Requires all production resources to have a backup label with a valid schedule

package custom.policies.backup

import rego.v1

# Deny if production resource lacks backup label
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

# Deny if backup schedule format is invalid
deny contains violation if {
	input.resource
	resource := input.resource

	resource.labels.env == "production"
	resource.labels.backup != ""

	# Validate backup schedule format
	not regex.match(`^(daily|weekly|monthly)$`, resource.labels.backup)

	violation := {
		"message": sprintf(
			"Invalid backup schedule '%s' for resource %s (must be: daily, weekly, or monthly)",
			[resource.labels.backup, resource.id],
		),
		"severity": "error",
		"resource": resource.id,
	}
}
