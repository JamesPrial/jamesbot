// Package errutil provides custom error types for JamesBot.
package errutil

import "fmt"

// ConfigError represents a configuration-related error.
// It is returned when there are problems with configuration values.
type ConfigError struct {
	Key     string
	Message string
}

// Error implements the error interface for ConfigError.
func (e ConfigError) Error() string {
	return fmt.Sprintf("config error for %s: %s", e.Key, e.Message)
}

// ValidationError represents an input validation failure.
// It is returned when user input or command arguments fail validation.
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface for ValidationError.
func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error for %s: %s", e.Field, e.Message)
}

// CommandError wraps errors that occur during command execution.
// It provides context about which command failed while preserving the underlying error.
type CommandError struct {
	Command string
	Err     error
}

// Error implements the error interface for CommandError.
func (e CommandError) Error() string {
	return fmt.Sprintf("command %s failed: %v", e.Command, e.Err)
}

// Unwrap returns the underlying error, supporting errors.Unwrap.
func (e CommandError) Unwrap() error {
	return e.Err
}

// UserFriendlyError wraps an internal error with a user-friendly message.
// The Error() method returns the internal error for logging,
// while UserMessage contains the message suitable for display to Discord users.
type UserFriendlyError struct {
	UserMessage string
	Err         error
}

// Error implements the error interface for UserFriendlyError.
// It returns the internal error message for logging purposes.
func (e UserFriendlyError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.UserMessage
}

// Unwrap returns the underlying error, supporting errors.Unwrap.
func (e UserFriendlyError) Unwrap() error {
	return e.Err
}

// PermissionError represents a missing permission error.
// It is returned when a user or bot lacks the required permissions for an operation.
type PermissionError struct {
	Permission string
}

// Error implements the error interface for PermissionError.
func (e PermissionError) Error() string {
	return fmt.Sprintf("missing permission: %s", e.Permission)
}
