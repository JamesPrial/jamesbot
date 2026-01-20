// Package commands provides CLI command implementations for JamesBot.
package commands

import (
	"flag"
	"strings"
)

// RulesCommand is a parent command for rule management.
// It acts as a container for subcommands like list and set.
type RulesCommand struct{}

// NewRulesCommand creates a new RulesCommand instance.
func NewRulesCommand() *RulesCommand {
	return &RulesCommand{}
}

// Name returns the name of the command.
func (c *RulesCommand) Name() string {
	return "rules"
}

// Synopsis returns a brief description of the command.
func (c *RulesCommand) Synopsis() string {
	return "Manage server rules"
}

// Usage returns detailed usage information for the command.
func (c *RulesCommand) Usage() string {
	var sb strings.Builder
	sb.WriteString("Usage: jamesbot rules <subcommand> [options]\n\n")
	sb.WriteString("Manage server rules and rule configurations.\n\n")
	sb.WriteString("Subcommands:\n")
	sb.WriteString("  list   List all server rules\n")
	sb.WriteString("  set    Set or update a rule\n\n")
	sb.WriteString("Use \"jamesbot rules <subcommand> -h\" for more information about a subcommand.\n")
	return sb.String()
}

// SetFlags configures the command-line flags for the rules command.
// Parent commands typically don't have their own flags.
func (c *RulesCommand) SetFlags(fs *flag.FlagSet) {
	// No flags for parent command
}

// Run executes the rules command.
// When invoked without a subcommand, it prints usage information.
func (c *RulesCommand) Run(ctx *CLIContext, args []string) int {
	// This method should not be called directly when the command is properly
	// registered as a ParentCommand, but we provide a fallback implementation.
	stdout := ctx.Stdout
	stdout.Write([]byte(c.Usage()))
	return 0
}
