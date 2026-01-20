// Package commands provides CLI command implementations for JamesBot.
package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"jamesbot/internal/api"
)

// RulesListCommand implements the rules list command for displaying all server rules.
type RulesListCommand struct {
	jsonOutput bool
	endpoint   string
}

// NewRulesListCommand creates a new RulesListCommand instance.
func NewRulesListCommand() *RulesListCommand {
	return &RulesListCommand{}
}

// Name returns the name of the command.
func (c *RulesListCommand) Name() string {
	return "list"
}

// Synopsis returns a brief description of the command.
func (c *RulesListCommand) Synopsis() string {
	return "List all server rules"
}

// Usage returns detailed usage information for the command.
func (c *RulesListCommand) Usage() string {
	var sb strings.Builder
	sb.WriteString("Usage: jamesbot rules list [options]\n\n")
	sb.WriteString("List all configured server rules.\n\n")
	sb.WriteString("Options:\n")
	sb.WriteString("  --json              Output rules as JSON instead of human-readable format\n")
	sb.WriteString("  --endpoint <url>    API endpoint (default: http://127.0.0.1:8765)\n")
	sb.WriteString("  -h, --help          Show this help message\n")
	return sb.String()
}

// SetFlags configures the command-line flags for the rules list command.
func (c *RulesListCommand) SetFlags(fs *flag.FlagSet) {
	fs.BoolVar(&c.jsonOutput, "json", false, "Output rules as JSON")
	fs.StringVar(&c.endpoint, "endpoint", "http://127.0.0.1:8765", "API endpoint")
}

// Run executes the rules list command.
// It accepts a CLI context with stdout/stderr and command arguments.
func (c *RulesListCommand) Run(ctx *CLIContext, args []string) int {
	// Get stdout and stderr from context
	stdout := ctx.Stdout
	stderr := ctx.Stderr

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

	// Get rules from API
	rules, err := client.ListRules()
	if err != nil {
		// Check if this is a connection error
		if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "connection failed") {
			fmt.Fprintf(stderr, "Error: Cannot connect to bot API at %s\n", endpoint)
			fmt.Fprintf(stderr, "Make sure the bot is running with 'jamesbot serve'\n")
			return 1
		}

		// Other API errors
		fmt.Fprintf(stderr, "Error: Failed to get rules: %v\n", err)
		return 1
	}

	// Handle nil rules
	if rules == nil {
		fmt.Fprintf(stderr, "Error: Received nil rules from API\n")
		return 1
	}

	// Output rules in requested format
	if c.jsonOutput {
		// JSON output
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(rules); err != nil {
			fmt.Fprintf(stderr, "Error: Failed to encode rules as JSON: %v\n", err)
			return 1
		}
	} else {
		// Human-readable table output
		if len(rules) == 0 {
			fmt.Fprintf(stdout, "No rules configured\n")
			return 0
		}

		// Calculate column widths
		maxNameLen := len("Name")
		maxDescLen := len("Description")
		for _, rule := range rules {
			if len(rule.Name) > maxNameLen {
				maxNameLen = len(rule.Name)
			}
			if len(rule.Description) > maxDescLen {
				maxDescLen = len(rule.Description)
			}
		}

		// Print header
		fmt.Fprintf(stdout, "%-*s  %-7s  %-*s\n", maxNameLen, "Name", "Enabled", maxDescLen, "Description")
		fmt.Fprintf(stdout, "%s  %s  %s\n", strings.Repeat("-", maxNameLen), strings.Repeat("-", 7), strings.Repeat("-", maxDescLen))

		// Print rules
		for _, rule := range rules {
			enabledStr := "false"
			if rule.Enabled {
				enabledStr = "true"
			}
			fmt.Fprintf(stdout, "%-*s  %-7s  %-*s\n", maxNameLen, rule.Name, enabledStr, maxDescLen, rule.Description)
		}
	}

	return 0
}
