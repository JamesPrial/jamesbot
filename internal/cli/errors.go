package cli

// Exit codes for CLI commands.
// These are returned by command Run() methods and used as process exit codes.
const (
	// ExitSuccess indicates successful command execution.
	ExitSuccess = 0

	// ExitUsageError indicates invalid command usage or arguments.
	// This is returned when flags are incorrect or required arguments are missing.
	ExitUsageError = 1

	// ExitConnectionFail indicates a network connection failure.
	// This is returned when the command cannot establish a connection to required services.
	ExitConnectionFail = 2

	// ExitAPIError indicates an API error response.
	// This is returned when the API returns an error or unexpected response.
	ExitAPIError = 3

	// ExitConfigError indicates a configuration error.
	// This is returned when configuration is invalid or cannot be loaded.
	ExitConfigError = 4

	// ExitFailure indicates a general failure.
	// This is used for unspecified errors during command execution.
	ExitFailure = 1
)
