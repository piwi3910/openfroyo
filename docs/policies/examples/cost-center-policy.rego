# Cost Center Policy
# Requires cost-center labels for billing and tracking

package custom.policies.costcenter

import rego.v1

# Deny if cost-center label is missing
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

# Deny if cost-center format is invalid
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
