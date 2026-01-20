package cli

import (
	"io"
	"os"

	"jamesbot/internal/config"
)

// Context provides execution context and resources for CLI commands.
// It wraps standard I/O streams, configuration, and API connection details
// to provide convenient access to command execution resources.
type Context struct {
	// Stdout is the standard output stream for command output.
	Stdout io.Writer

	// Stderr is the standard error stream for error messages.
	Stderr io.Writer

	// Config is the application configuration.
	// This may be nil if config loading failed or is not needed.
	Config *config.Config

	// APIEndpoint is the base URL for API requests.
	// This is used by commands that need to communicate with the bot's API.
	APIEndpoint string
}

// NewContext creates a new CLI context with the provided components.
// If stdout or stderr are nil, they default to os.Stdout and os.Stderr respectively.
func NewContext(stdout, stderr io.Writer, cfg *config.Config, apiEndpoint string) *Context {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}

	return &Context{
		Stdout:      stdout,
		Stderr:      stderr,
		Config:      cfg,
		APIEndpoint: apiEndpoint,
	}
}
