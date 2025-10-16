package policy

import (
	"encoding/json"
	"time"

	"github.com/openfroyo/openfroyo/pkg/engine"
)

// Severity represents the severity level of a policy violation.
type Severity string

const (
	// SeverityInfo is for informational messages.
	SeverityInfo Severity = "info"

	// SeverityWarning is for warnings that should be reviewed.
	SeverityWarning Severity = "warning"

	// SeverityError is for errors that should block operations.
	SeverityError Severity = "error"

	// SeverityCritical is for critical violations that must be addressed immediately.
	SeverityCritical Severity = "critical"
)

// Policy represents a policy rule with its Rego code.
type Policy struct {
	// Name is the unique name of the policy.
	Name string `json:"name"`

	// Description provides a human-readable description.
	Description string `json:"description"`

	// Rego contains the Rego policy code.
	Rego string `json:"rego"`

	// Severity is the default severity for violations.
	Severity Severity `json:"severity"`

	// Enabled indicates if the policy is active.
	Enabled bool `json:"enabled"`

	// Tags are labels for organizing policies.
	Tags []string `json:"tags,omitempty"`

	// Metadata contains additional policy metadata.
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// CreatedAt is when the policy was created.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the policy was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// PolicyViolation represents a single policy violation.
type PolicyViolation struct {
	// Policy is the name of the policy that was violated.
	Policy string `json:"policy"`

	// Resource is the resource ID that violated the policy.
	Resource string `json:"resource,omitempty"`

	// Message is a human-readable violation message.
	Message string `json:"message"`

	// Severity is the violation severity level.
	Severity Severity `json:"severity"`

	// Details contains additional violation details.
	Details map[string]interface{} `json:"details,omitempty"`

	// Remediation provides suggested fixes.
	Remediation string `json:"remediation,omitempty"`

	// DetectedAt is when the violation was detected.
	DetectedAt time.Time `json:"detected_at"`
}

// PolicyResult represents the result of policy evaluation.
type PolicyResult struct {
	// Allowed indicates if the operation is allowed.
	Allowed bool `json:"allowed"`

	// Violations lists all policy violations.
	Violations []PolicyViolation `json:"violations,omitempty"`

	// Warnings lists policy warnings that don't block operations.
	Warnings []PolicyViolation `json:"warnings,omitempty"`

	// EvaluatedAt is when the policy was evaluated.
	EvaluatedAt time.Time `json:"evaluated_at"`

	// EvaluatedPolicies lists the names of policies that were evaluated.
	EvaluatedPolicies []string `json:"evaluated_policies"`

	// Duration is how long the evaluation took.
	Duration time.Duration `json:"duration"`

	// Context contains evaluation context information.
	Context *PolicyContext `json:"context,omitempty"`
}

// PolicyInput represents the input data for policy evaluation.
type PolicyInput struct {
	// Resource is the resource being evaluated.
	Resource *engine.Resource `json:"resource,omitempty"`

	// DesiredState is the desired state from configuration.
	DesiredState json.RawMessage `json:"desired_state,omitempty"`

	// ActualState is the actual state from facts.
	ActualState json.RawMessage `json:"actual_state,omitempty"`

	// Plan is the execution plan being evaluated.
	Plan *engine.Plan `json:"plan,omitempty"`

	// Context provides additional evaluation context.
	Context *PolicyContext `json:"context"`
}

// PolicyContext provides context information for policy evaluation.
type PolicyContext struct {
	// User is the user performing the operation.
	User string `json:"user,omitempty"`

	// Environment is the environment (e.g., "production", "staging").
	Environment string `json:"environment,omitempty"`

	// Timestamp is when the evaluation is occurring.
	Timestamp time.Time `json:"timestamp"`

	// Operation is the operation being performed (e.g., "create", "update", "delete").
	Operation string `json:"operation,omitempty"`

	// Target contains information about the target system.
	Target *engine.TargetInfo `json:"target,omitempty"`

	// DryRun indicates if this is a dry-run evaluation.
	DryRun bool `json:"dry_run"`

	// Metadata contains additional context metadata.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// PolicyBundle represents a collection of related policies.
type PolicyBundle struct {
	// Name is the unique name of the bundle.
	Name string `json:"name"`

	// Version is the bundle version.
	Version string `json:"version"`

	// Description provides a human-readable description.
	Description string `json:"description"`

	// Policies are the policies in this bundle.
	Policies []Policy `json:"policies"`

	// CreatedAt is when the bundle was created.
	CreatedAt time.Time `json:"created_at"`
}

// DriftPolicy represents a policy for drift detection and remediation.
type DriftPolicy struct {
	// ResourcePattern is a regex pattern for matching resource IDs.
	ResourcePattern string `json:"resource_pattern"`

	// MaxDriftThreshold is the maximum acceptable drift percentage.
	MaxDriftThreshold float64 `json:"max_drift_threshold"`

	// AutoRemediate indicates if drift should be auto-remediated.
	AutoRemediate bool `json:"auto_remediate"`

	// RemediationDelay is the delay before auto-remediation.
	RemediationDelay time.Duration `json:"remediation_delay,omitempty"`

	// NotificationChannels lists channels to notify on drift.
	NotificationChannels []string `json:"notification_channels,omitempty"`
}

// ValidationError represents a policy validation error.
type ValidationError struct {
	// Field is the field that failed validation.
	Field string `json:"field"`

	// Message describes the validation error.
	Message string `json:"message"`

	// Value is the invalid value.
	Value interface{} `json:"value,omitempty"`
}

// PolicyReport represents a comprehensive policy evaluation report.
type PolicyReport struct {
	// ID is the unique identifier for this report.
	ID string `json:"id"`

	// GeneratedAt is when the report was generated.
	GeneratedAt time.Time `json:"generated_at"`

	// Results are the policy evaluation results.
	Results []*PolicyResult `json:"results"`

	// Summary provides aggregate statistics.
	Summary *PolicySummary `json:"summary"`

	// Recommendations lists recommended actions.
	Recommendations []string `json:"recommendations,omitempty"`
}

// PolicySummary provides aggregate statistics for policy evaluation.
type PolicySummary struct {
	// TotalPolicies is the total number of policies evaluated.
	TotalPolicies int `json:"total_policies"`

	// TotalViolations is the total number of violations.
	TotalViolations int `json:"total_violations"`

	// ViolationsBySeverity breaks down violations by severity.
	ViolationsBySeverity map[Severity]int `json:"violations_by_severity"`

	// TotalWarnings is the total number of warnings.
	TotalWarnings int `json:"total_warnings"`

	// AllowedOperations is the number of allowed operations.
	AllowedOperations int `json:"allowed_operations"`

	// BlockedOperations is the number of blocked operations.
	BlockedOperations int `json:"blocked_operations"`

	// EvaluationDuration is the total evaluation time.
	EvaluationDuration time.Duration `json:"evaluation_duration"`
}
