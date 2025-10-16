# Compliance Policy
# Ensures compliance-related tags are present and valid

package custom.policies.compliance

import rego.v1

# Required compliance tags
compliance_tags := ["data-classification", "retention-period", "encryption"]

# Deny if required compliance tag is missing
deny contains violation if {
	input.resource
	resource := input.resource
	some tag in compliance_tags

	# Check if tag is present
	not resource.labels[tag]

	violation := {
		"message": sprintf("Resource %s missing required compliance tag: %s", [resource.id, tag]),
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
		"message": sprintf(
			"Invalid data-classification '%s' (must be: public, internal, confidential, or restricted)",
			[classification],
		),
		"severity": "error",
		"resource": resource.id,
	}
}

# Validate retention period format
deny contains violation if {
	input.resource
	resource := input.resource

	retention := resource.labels["retention-period"]
	retention != ""

	# Format should be like "7d", "30d", "1y", etc.
	not regex.match(`^\d+[dmy]$`, retention)

	violation := {
		"message": sprintf(
			"Invalid retention-period format '%s' (must be number followed by d/m/y, e.g., 30d, 1y)",
			[retention],
		),
		"severity": "error",
		"resource": resource.id,
	}
}

# Validate encryption setting
deny contains violation if {
	input.resource
	resource := input.resource

	encryption := resource.labels.encryption
	encryption != ""

	# Must be "enabled" or "disabled"
	not encryption in ["enabled", "disabled"]

	violation := {
		"message": sprintf(
			"Invalid encryption value '%s' (must be: enabled or disabled)",
			[encryption],
		),
		"severity": "error",
		"resource": resource.id,
	}
}
