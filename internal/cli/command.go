// Package cli provides the command-line interface for JamesBot.
package cli

import "flag"

// CLICommand defines the interface for CLI subcommands.
// All CLI commands must implement this interface to be registered and executed.
type CLICommand interface {
	// Name returns the command name as it appears in the CLI.
	// This should be lowercase and contain no spaces.
	Name() string

	// Synopsis returns a brief one-line description of the command.
	// This is displayed in the help output listing all commands.
	Synopsis() string

	// Usage returns detailed usage information for the command.
	// This is displayed when the command is invoked with -h or --help.
	Usage() string

	// SetFlags configures the command's flag set.
	// Commands should define their flags by calling methods on fs.
	SetFlags(fs *flag.FlagSet)

	// Run executes the command with the provided context and arguments.
	// It returns an exit code (0 for success, non-zero for errors).
	Run(ctx *Context, args []string) int
}

// ParentCommand is an optional interface for commands with subcommands.
// Commands that implement this interface can have nested subcommands,
// allowing for hierarchical command structures like "jamesbot rules list".
type ParentCommand interface {
	CLICommand

	// Subcommands returns the list of subcommands for this command.
	// The CLI router will use this to dispatch to nested commands.
	Subcommands() []CLICommand
}
