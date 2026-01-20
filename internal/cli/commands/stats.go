// Package commands provides CLI command implementations for JamesBot.
package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"jamesbot/internal/api"
)

// StatsCommand implements the stats command for displaying bot statistics.
type StatsCommand struct {
	jsonOutput bool
	endpoint   string
}

// NewStatsCommand creates a new StatsCommand instance.
func NewStatsCommand() *StatsCommand {
	return &StatsCommand{}
}

// Name returns the name of the command.
func (c *StatsCommand) Name() string {
	return "stats"
}

// Synopsis returns a brief description of the command.
func (c *StatsCommand) Synopsis() string {
	return "Display bot statistics"
}

// Usage returns detailed usage information for the command.
func (c *StatsCommand) Usage() string {
	var sb strings.Builder
	sb.WriteString("Usage: jamesbot stats [options]\n\n")
	sb.WriteString("Display statistics about the bot's operation.\n\n")
	sb.WriteString("Options:\n")
	sb.WriteString("  --json              Output stats as JSON instead of human-readable format\n")
	sb.WriteString("  --endpoint <url>    API endpoint (default: http://127.0.0.1:8765)\n")
	sb.WriteString("  -h, --help          Show this help message\n")
	return sb.String()
}

// SetFlags configures the command-line flags for the stats command.
func (c *StatsCommand) SetFlags(fs *flag.FlagSet) {
	fs.BoolVar(&c.jsonOutput, "json", false, "Output stats as JSON")
	fs.StringVar(&c.endpoint, "endpoint", "http://127.0.0.1:8765", "API endpoint")
}

// Run executes the stats command.
// It accepts a CLI context with stdout/stderr and command arguments.
func (c *StatsCommand) Run(ctx *CLIContext, args []string) int {
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

	// Get stats from API
	stats, err := client.GetStats()
	if err != nil {
		// Check if this is a connection error
		if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "connection failed") {
			fmt.Fprintf(stderr, "Error: Cannot connect to bot API at %s\n", endpoint)
			fmt.Fprintf(stderr, "Make sure the bot is running with 'jamesbot serve'\n")
			return 1
		}

		// Other API errors
		fmt.Fprintf(stderr, "Error: Failed to get stats: %v\n", err)
		return 1
	}

	// Handle nil stats
	if stats == nil {
		fmt.Fprintf(stderr, "Error: Received nil stats from API\n")
		return 1
	}

	// Output stats in requested format
	if c.jsonOutput {
		// JSON output
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(stats); err != nil {
			fmt.Fprintf(stderr, "Error: Failed to encode stats as JSON: %v\n", err)
			return 1
		}
	} else {
		// Human-readable output
		fmt.Fprintf(stdout, "Uptime: %s\n", stats.Uptime)
		fmt.Fprintf(stdout, "Commands executed: %d\n", stats.CommandsExecuted)
		fmt.Fprintf(stdout, "Guilds: %d\n", stats.GuildCount)
		fmt.Fprintf(stdout, "Active rules: %d\n", stats.ActiveRules)
	}

	return 0
}
