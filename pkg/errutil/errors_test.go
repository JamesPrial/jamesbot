package errutil_test

import (
	"errors"
	"io"
	"testing"

	"jamesbot/pkg/errutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ConfigError_Error(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		message  string
		expected string
	}{
		{
			name:     "error message format with key and message",
			key:      "token",
			message:  "required",
			expected: "config error for token: required",
		},
		{
			name:     "empty key still formats correctly",
			key:      "",
			message:  "msg",
			expected: "config error for : msg",
		},
		{
			name:     "empty message",
			key:      "database",
			message:  "",
			expected: "config error for database: ",
		},
		{
			name:     "both empty",
			key:      "",
			message:  "",
			expected: "config error for : ",
		},
		{
			name:     "key with dots",
			key:      "discord.token",
			message:  "is required",
			expected: "config error for discord.token: is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errutil.ConfigError{
				Key:     tt.key,
				Message: tt.message,
			}

			got := err.Error()

			assert.Equal(t, tt.expected, got)
		})
	}
}

func Test_ConfigError_ImplementsError(t *testing.T) {
	var _ error = errutil.ConfigError{}
}

func Test_ValidationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		message  string
		expected string
	}{
		{
			name:     "error message format with field and message",
			field:    "user",
			message:  "invalid",
			expected: "validation error for user: invalid",
		},
		{
			name:     "empty field",
			field:    "",
			message:  "cannot be empty",
			expected: "validation error for : cannot be empty",
		},
		{
			name:     "empty message",
			field:    "email",
			message:  "",
			expected: "validation error for email: ",
		},
		{
			name:     "field with special characters",
			field:    "user.email",
			message:  "invalid format",
			expected: "validation error for user.email: invalid format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errutil.ValidationError{
				Field:   tt.field,
				Message: tt.message,
			}

			got := err.Error()

			assert.Equal(t, tt.expected, got)
		})
	}
}

func Test_ValidationError_ImplementsError(t *testing.T) {
	var _ error = errutil.ValidationError{}
}

func Test_CommandError_Error(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		err      error
		expected string
	}{
		{
			name:     "error message with wrapped io.EOF",
			command:  "ping",
			err:      io.EOF,
			expected: "command ping failed: EOF",
		},
		{
			name:     "error message with custom error",
			command:  "ban",
			err:      errors.New("user not found"),
			expected: "command ban failed: user not found",
		},
		{
			name:     "empty command name",
			command:  "",
			err:      io.EOF,
			expected: "command  failed: EOF",
		},
		{
			name:     "nil inner error",
			command:  "kick",
			err:      nil,
			expected: "command kick failed: <nil>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errutil.CommandError{
				Command: tt.command,
				Err:     tt.err,
			}

			got := err.Error()

			assert.Equal(t, tt.expected, got)
		})
	}
}

func Test_CommandError_Unwrap(t *testing.T) {
	tests := []struct {
		name        string
		innerErr    error
		expectedErr error
	}{
		{
			name:        "unwrap returns io.EOF",
			innerErr:    io.EOF,
			expectedErr: io.EOF,
		},
		{
			name:        "unwrap returns custom error",
			innerErr:    errors.New("custom error"),
			expectedErr: errors.New("custom error"),
		},
		{
			name:        "unwrap returns nil for nil inner error",
			innerErr:    nil,
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errutil.CommandError{
				Command: "test",
				Err:     tt.innerErr,
			}

			got := err.Unwrap()

			if tt.expectedErr == nil {
				assert.Nil(t, got)
			} else {
				assert.Equal(t, tt.expectedErr.Error(), got.Error())
			}
		})
	}
}

func Test_CommandError_ErrorsIs(t *testing.T) {
	innerErr := io.EOF
	cmdErr := errutil.CommandError{
		Command: "ping",
		Err:     innerErr,
	}

	assert.True(t, errors.Is(cmdErr, io.EOF), "errors.Is should find wrapped io.EOF")
	assert.False(t, errors.Is(cmdErr, io.ErrUnexpectedEOF), "errors.Is should not find unrelated error")
}

func Test_CommandError_ImplementsError(t *testing.T) {
	var _ error = errutil.CommandError{}
}

func Test_UserFriendlyError_Error(t *testing.T) {
	tests := []struct {
		name        string
		userMessage string
		innerErr    error
		expected    string
	}{
		{
			name:        "error returns internal message not user message",
			userMessage: "oops",
			innerErr:    io.EOF,
			expected:    "EOF",
		},
		{
			name:        "error returns custom internal error message",
			userMessage: "Something went wrong",
			innerErr:    errors.New("database connection failed"),
			expected:    "database connection failed",
		},
		{
			name:        "nil inner error",
			userMessage: "try again",
			innerErr:    nil,
			expected:    "try again",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errutil.UserFriendlyError{
				UserMessage: tt.userMessage,
				Err:         tt.innerErr,
			}

			got := err.Error()

			assert.Equal(t, tt.expected, got)
		})
	}
}

func Test_UserFriendlyError_Unwrap(t *testing.T) {
	tests := []struct {
		name        string
		innerErr    error
		expectedErr error
	}{
		{
			name:        "unwrap returns io.EOF",
			innerErr:    io.EOF,
			expectedErr: io.EOF,
		},
		{
			name:        "unwrap returns custom error",
			innerErr:    errors.New("internal error"),
			expectedErr: errors.New("internal error"),
		},
		{
			name:        "unwrap returns nil for nil inner error",
			innerErr:    nil,
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errutil.UserFriendlyError{
				UserMessage: "user message",
				Err:         tt.innerErr,
			}

			got := err.Unwrap()

			if tt.expectedErr == nil {
				assert.Nil(t, got)
			} else {
				assert.Equal(t, tt.expectedErr.Error(), got.Error())
			}
		})
	}
}

func Test_UserFriendlyError_UserMessageAccessible(t *testing.T) {
	err := errutil.UserFriendlyError{
		UserMessage: "Please try again later",
		Err:         errors.New("internal"),
	}

	// The UserMessage field should be accessible for displaying to users
	assert.Equal(t, "Please try again later", err.UserMessage)
}

func Test_UserFriendlyError_ErrorsIs(t *testing.T) {
	innerErr := io.EOF
	ufErr := errutil.UserFriendlyError{
		UserMessage: "oops",
		Err:         innerErr,
	}

	assert.True(t, errors.Is(ufErr, io.EOF), "errors.Is should find wrapped io.EOF")
	assert.False(t, errors.Is(ufErr, io.ErrUnexpectedEOF), "errors.Is should not find unrelated error")
}

func Test_UserFriendlyError_ImplementsError(t *testing.T) {
	var _ error = errutil.UserFriendlyError{}
}

func Test_PermissionError_Error(t *testing.T) {
	tests := []struct {
		name       string
		permission string
		expected   string
	}{
		{
			name:       "missing permission message",
			permission: "Kick Members",
			expected:   "missing permission: Kick Members",
		},
		{
			name:       "empty permission",
			permission: "",
			expected:   "missing permission: ",
		},
		{
			name:       "administrator permission",
			permission: "Administrator",
			expected:   "missing permission: Administrator",
		},
		{
			name:       "multiple words permission",
			permission: "Manage Guild Members",
			expected:   "missing permission: Manage Guild Members",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errutil.PermissionError{
				Permission: tt.permission,
			}

			got := err.Error()

			assert.Equal(t, tt.expected, got)
		})
	}
}

func Test_PermissionError_ImplementsError(t *testing.T) {
	var _ error = errutil.PermissionError{}
}

// Test that all error types can be used with errors.As
func Test_ErrorTypes_ErrorsAs(t *testing.T) {
	t.Run("ConfigError can be extracted with errors.As", func(t *testing.T) {
		original := &errutil.ConfigError{Key: "test", Message: "msg"}
		var target *errutil.ConfigError
		require.True(t, errors.As(original, &target))
		assert.Equal(t, "test", target.Key)
		assert.Equal(t, "msg", target.Message)
	})

	t.Run("ValidationError can be extracted with errors.As", func(t *testing.T) {
		original := &errutil.ValidationError{Field: "email", Message: "invalid"}
		var target *errutil.ValidationError
		require.True(t, errors.As(original, &target))
		assert.Equal(t, "email", target.Field)
		assert.Equal(t, "invalid", target.Message)
	})

	t.Run("CommandError can be extracted with errors.As", func(t *testing.T) {
		original := &errutil.CommandError{Command: "ping", Err: io.EOF}
		var target *errutil.CommandError
		require.True(t, errors.As(original, &target))
		assert.Equal(t, "ping", target.Command)
		assert.Equal(t, io.EOF, target.Err)
	})

	t.Run("UserFriendlyError can be extracted with errors.As", func(t *testing.T) {
		original := &errutil.UserFriendlyError{UserMessage: "oops", Err: io.EOF}
		var target *errutil.UserFriendlyError
		require.True(t, errors.As(original, &target))
		assert.Equal(t, "oops", target.UserMessage)
		assert.Equal(t, io.EOF, target.Err)
	})

	t.Run("PermissionError can be extracted with errors.As", func(t *testing.T) {
		original := &errutil.PermissionError{Permission: "Admin"}
		var target *errutil.PermissionError
		require.True(t, errors.As(original, &target))
		assert.Equal(t, "Admin", target.Permission)
	})
}
