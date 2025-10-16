// Package engine provides the core types and interfaces for the OpenFroyo orchestration engine.
// It defines the 6-phase execution workflow: Config -> Facts -> Plan -> Apply -> Result -> Drift.
package engine

import (
	"errors"
	"fmt"
)

// ErrorClass represents the classification of an error for retry and recovery logic.
type ErrorClass string

const (
	// ErrorClassTransient indicates a temporary failure that may succeed on retry.
	// Examples: network timeouts, temporary service unavailability.
	ErrorClassTransient ErrorClass = "transient"

	// ErrorClassThrottled indicates rate limiting or quota exhaustion.
	// Should be retried with exponential backoff.
	ErrorClassThrottled ErrorClass = "throttled"

	// ErrorClassConflict indicates a resource state conflict.
	// Examples: concurrent modifications, optimistic locking failures.
	ErrorClassConflict ErrorClass = "conflict"

	// ErrorClassPermanent indicates a non-recoverable error.
	// Examples: invalid configuration, permission denied, resource not found.
	ErrorClassPermanent ErrorClass = "permanent"
)

// EngineError represents a classified error with context.
// nolint:revive // EngineError is intentionally named to distinguish from standard errors
type EngineError struct {
	// Class is the error classification for retry logic.
	Class ErrorClass `json:"class"`

	// Message is the human-readable error message.
	Message string `json:"message"`

	// Code is an optional error code for programmatic handling.
	Code string `json:"code,omitempty"`

	// Resource is the resource ID that caused the error, if applicable.
	Resource string `json:"resource,omitempty"`

	// Operation is the operation being performed when the error occurred.
	Operation string `json:"operation,omitempty"`

	// Err is the underlying error that caused this error.
	Err error `json:"-"`

	// Details contains additional context-specific information.
	Details map[string]interface{} `json:"details,omitempty"`
}

// Error implements the error interface.
func (e *EngineError) Error() string {
	if e.Resource != "" && e.Operation != "" {
		return fmt.Sprintf("[%s] %s (resource=%s, operation=%s): %s",
			e.Class, e.Message, e.Resource, e.Operation, e.unwrapMessage())
	}
	if e.Resource != "" {
		return fmt.Sprintf("[%s] %s (resource=%s): %s",
			e.Class, e.Message, e.Resource, e.unwrapMessage())
	}
	return fmt.Sprintf("[%s] %s: %s", e.Class, e.Message, e.unwrapMessage())
}

// Unwrap returns the underlying error for error chain inspection.
func (e *EngineError) Unwrap() error {
	return e.Err
}

// unwrapMessage returns the error message from the underlying error chain.
func (e *EngineError) unwrapMessage() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return ""
}

// Is implements error equality checking for errors.Is.
func (e *EngineError) Is(target error) bool {
	t, ok := target.(*EngineError)
	if !ok {
		return false
	}
	return e.Class == t.Class && e.Code == t.Code
}

// NewTransientError creates a new transient error.
func NewTransientError(message string, err error) *EngineError {
	return &EngineError{
		Class:   ErrorClassTransient,
		Message: message,
		Err:     err,
	}
}

// NewThrottledError creates a new throttled error.
func NewThrottledError(message string, err error) *EngineError {
	return &EngineError{
		Class:   ErrorClassThrottled,
		Message: message,
		Err:     err,
	}
}

// NewConflictError creates a new conflict error.
func NewConflictError(message string, err error) *EngineError {
	return &EngineError{
		Class:   ErrorClassConflict,
		Message: message,
		Err:     err,
	}
}

// NewPermanentError creates a new permanent error.
func NewPermanentError(message string, err error) *EngineError {
	return &EngineError{
		Class:   ErrorClassPermanent,
		Message: message,
		Err:     err,
	}
}

// WithResource adds resource context to an error.
func (e *EngineError) WithResource(resourceID string) *EngineError {
	e.Resource = resourceID
	return e
}

// WithOperation adds operation context to an error.
func (e *EngineError) WithOperation(operation string) *EngineError {
	e.Operation = operation
	return e
}

// WithCode adds an error code to an error.
func (e *EngineError) WithCode(code string) *EngineError {
	e.Code = code
	return e
}

// WithDetail adds a detail field to the error context.
func (e *EngineError) WithDetail(key string, value interface{}) *EngineError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// IsTransient returns true if the error is classified as transient.
func IsTransient(err error) bool {
	var e *EngineError
	if errors.As(err, &e) {
		return e.Class == ErrorClassTransient
	}
	return false
}

// IsThrottled returns true if the error is classified as throttled.
func IsThrottled(err error) bool {
	var e *EngineError
	if errors.As(err, &e) {
		return e.Class == ErrorClassThrottled
	}
	return false
}

// IsConflict returns true if the error is classified as a conflict.
func IsConflict(err error) bool {
	var e *EngineError
	if errors.As(err, &e) {
		return e.Class == ErrorClassConflict
	}
	return false
}

// IsPermanent returns true if the error is classified as permanent.
func IsPermanent(err error) bool {
	var e *EngineError
	if errors.As(err, &e) {
		return e.Class == ErrorClassPermanent
	}
	return false
}

// IsRetryable returns true if the error can be retried.
// Transient, throttled, and conflict errors are retryable.
func IsRetryable(err error) bool {
	return IsTransient(err) || IsThrottled(err) || IsConflict(err)
}

// Common error codes.
const (
	ErrCodeValidation       = "VALIDATION_ERROR"
	ErrCodeNotFound         = "NOT_FOUND"
	ErrCodeAlreadyExists    = "ALREADY_EXISTS"
	ErrCodePermissionDenied = "PERMISSION_DENIED"
	ErrCodeTimeout          = "TIMEOUT"
	ErrCodeRateLimited      = "RATE_LIMITED"
	ErrCodeConflict         = "CONFLICT"
	ErrCodeInternal         = "INTERNAL_ERROR"
	ErrCodeProviderFailed   = "PROVIDER_FAILED"
	ErrCodeDependencyFailed = "DEPENDENCY_FAILED"
)
