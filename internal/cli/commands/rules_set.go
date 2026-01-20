// Package commands provides CLI command implementations for JamesBot.
package commands

import (
	"flag"
	"fmt"
	"strings"

	"jamesbot/internal/api"
)

// RulesSetCommand implements the rules set command for modifying rule settings.
type RulesSetCommand struct {
	endpoint string
}

// NewRulesSetCommand creates a new RulesSetCommand instance.
func NewRulesSetCommand() *RulesSetCommand {
	return &RulesSetCommand{}
}

// Name returns the name of the command.
func (c *RulesSetCommand) Name() string {
	return "set"
}

// Synopsis returns a brief description of the command.
func (c *RulesSetCommand) Synopsis() string {
	return "Set or update a rule"
}

// Usage returns detailed usage information for the command.
func (c *RulesSetCommand) Usage() string {
	var sb strings.Builder
	sb.WriteString("Usage: jamesbot rules set <rule-name> <key> <value> [options]\n\n")
	sb.WriteString("Set or update a server rule configuration.\n\n")
	sb.WriteString("Arguments:\n")
	sb.WriteString("  <rule-name>  Name of the rule to modify\n")
	sb.WriteString("  <key>        Configuration key to set\n")
	sb.WriteString("  <value>      Value to set for the key\n\n")
	sb.WriteString("Options:\n")
	sb.WriteString("  --endpoint <url>    API endpoint (default: http://127.0.0.1:8765)\n")
	sb.WriteString("  -h, --help          Show this help message\n\n")
	sb.WriteString("Examples:\n")
	sb.WriteString("  jamesbot rules set spam-filter enabled true\n")
	sb.WriteString("  jamesbot rules set auto-mod threshold 5\n")
	return sb.String()
}

// SetFlags configures the command-line flags for the rules set command.
func (c *RulesSetCommand) SetFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.endpoint, "endpoint", "http://127.0.0.1:8765", "API endpoint")
}

// Run executes the rules set command.
// It accepts a CLI context with stdout/stderr and command arguments.
func (c *RulesSetCommand) Run(ctx *CLIContext, args []string) int {
	// Get stdout and stderr from context
	stdout := ctx.Stdout
	stderr := ctx.Stderr

	// Validate arguments
	if len(args) < 3 {
		fmt.Fprintf(stderr, "Error: Missing required arguments\n\n")
		fmt.Fprintf(stderr, "%s", c.Usage())
		return 1
	}

	ruleName := args[0]
	key := args[1]
	value := args[2]

	// Use API endpoint from context if provided, otherwise use flag value
	endpoint := c.endpoint
	if ctx.APIEndpoint != "" {
		endpoint = ctx.APIEndpoint
	}

	// Create API client
	client := api.NewClient(endpoint)
	if client == nil {
		fmt.Fprintf(stderr, "Error: Failed to create API client\n")
		return 1
	}

	// Set rule via API
	err := client.SetRule(ruleName, key, value)
	if err != nil {
		// Check if this is a connection error
		if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "connection failed") {
			fmt.Fprintf(stderr, "Error: Cannot connect to bot API at %s\n", endpoint)
			fmt.Fprintf(stderr, "Make sure the bot is running with 'jamesbot serve'\n")
			return 1
		}

		// Other API errors
		fmt.Fprintf(stderr, "Error: Failed to set rule: %v\n", err)
		return 1
	}

	// Success message
	fmt.Fprintf(stdout, "Successfully set %s.%s = %s\n", ruleName, key, value)
	return 0
}
