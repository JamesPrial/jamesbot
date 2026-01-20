package cli

import (
	"flag"
	"fmt"
	"io"
	"os"

	"jamesbot/internal/cli/commands"
)

const (
	// Version is the current version of JamesBot.
	Version = "1.1.0"

	// AppName is the application name displayed in help text.
	AppName = "jamesbot"
)

// Run is the main entry point for the CLI application.
// It handles command routing, help, and version display.
// Returns an exit code suitable for os.Exit().
func Run(args []string, stdout, stderr io.Writer) int {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}

	// Handle no arguments - print usage
	if len(args) == 0 {
		printUsage(stdout)
		return ExitSuccess
	}

	// Handle global flags and special commands
	firstArg := args[0]
	switch firstArg {
	case "-h", "--help", "help":
		printUsage(stdout)
		return ExitSuccess
	case "-v", "--version", "version":
		fmt.Fprintf(stdout, "%s version %s\n", AppName, Version)
		return ExitSuccess
	}

	// Get the command registry
	commands := getCommands()

	// Find and execute the command
	cmdName := args[0]
	cmd, ok := commands[cmdName]
	if !ok {
		fmt.Fprintf(stderr, "Error: unknown command %q\n\n", cmdName)
		printUsage(stderr)
		return ExitUsageError
	}

	// Check if this is a parent command with subcommands
	if parentCmd, isParent := cmd.(ParentCommand); isParent {
		return runParentCommand(parentCmd, args[1:], stdout, stderr)
	}

	// Execute the command
	return runCommand(cmd, args[1:], stdout, stderr)
}

// runCommand executes a single command with its arguments.
func runCommand(cmd CLICommand, args []string, stdout, stderr io.Writer) int {
	// Create flag set for the command
	fs := flag.NewFlagSet(cmd.Name(), flag.ContinueOnError)
	fs.SetOutput(stderr)

	// Configure command flags
	cmd.SetFlags(fs)

	// Add standard help flag
	helpFlag := fs.Bool("h", false, "Show help")

	// Parse flags
	if err := fs.Parse(args); err != nil {
		return ExitUsageError
	}

	// Handle help flag
	if *helpFlag {
		fmt.Fprintf(stdout, "%s\n", cmd.Usage())
		return ExitSuccess
	}

	// Create execution context
	ctx := NewContext(stdout, stderr, nil, "")

	// Execute the command
	return cmd.Run(ctx, fs.Args())
}

// runParentCommand handles execution of parent commands with subcommands.
func runParentCommand(parent ParentCommand, args []string, stdout, stderr io.Writer) int {
	// If no subcommand specified, print parent command usage
	if len(args) == 0 {
		fmt.Fprintf(stdout, "%s\n", parent.Usage())
		return ExitSuccess
	}

	// Handle help for parent command
	if args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		fmt.Fprintf(stdout, "%s\n", parent.Usage())
		return ExitSuccess
	}

	// Find the subcommand
	subcommands := parent.Subcommands()
	subcmdName := args[0]

	for _, subcmd := range subcommands {
		if subcmd.Name() == subcmdName {
			return runCommand(subcmd, args[1:], stdout, stderr)
		}
	}

	// Subcommand not found
	fmt.Fprintf(stderr, "Error: unknown subcommand %q for %q\n\n", subcmdName, parent.Name())
	fmt.Fprintf(stderr, "%s\n", parent.Usage())
	return ExitUsageError
}

// printUsage prints the general usage information.
func printUsage(w io.Writer) {
	fmt.Fprintf(w, "Usage: %s <command> [options]\n\n", AppName)
	fmt.Fprintf(w, "JamesBot - Discord moderation bot\n\n")
	fmt.Fprintf(w, "Commands:\n")

	commands := getCommands()
	for _, name := range []string{"serve", "stats", "rules"} {
		if cmd, ok := commands[name]; ok {
			fmt.Fprintf(w, "  %-12s %s\n", name, cmd.Synopsis())
		}
	}

	fmt.Fprintf(w, "\nGlobal Options:\n")
	fmt.Fprintf(w, "  -h, --help     Show help\n")
	fmt.Fprintf(w, "  -v, --version  Show version\n")
	fmt.Fprintf(w, "\nUse \"%s <command> -h\" for more information about a command.\n", AppName)
}

// getCommands returns the map of available commands.
// This is the command registry for the CLI.
func getCommands() map[string]CLICommand {
	return map[string]CLICommand{
		"serve": newServeCommandAdapter(),
		"stats": newStatsCommandAdapter(),
		"rules": newRulesCommandAdapter(),
	}
}

// serveCommandAdapter adapts commands.ServeCommand to the CLICommand interface.
type serveCommandAdapter struct {
	cmd *commands.ServeCommand
}

func newServeCommandAdapter() *serveCommandAdapter {
	return &serveCommandAdapter{
		cmd: commands.NewServeCommand(),
	}
}

func (a *serveCommandAdapter) Name() string {
	return a.cmd.Name()
}

func (a *serveCommandAdapter) Synopsis() string {
	return a.cmd.Synopsis()
}

func (a *serveCommandAdapter) Usage() string {
	return a.cmd.Usage()
}

func (a *serveCommandAdapter) SetFlags(fs *flag.FlagSet) {
	a.cmd.SetFlags(fs)
}

func (a *serveCommandAdapter) Run(ctx *Context, args []string) int {
	// Convert cli.Context to commands.CLIContext
	cmdCtx := &commands.CLIContext{
		Stdout:      ctx.Stdout,
		Stderr:      ctx.Stderr,
		Config:      ctx.Config,
		APIEndpoint: ctx.APIEndpoint,
	}
	return a.cmd.Run(cmdCtx, args)
}

// statsCommandAdapter adapts commands.StatsCommand to the CLICommand interface.
type statsCommandAdapter struct {
	cmd *commands.StatsCommand
}

func newStatsCommandAdapter() *statsCommandAdapter {
	return &statsCommandAdapter{
		cmd: commands.NewStatsCommand(),
	}
}

func (a *statsCommandAdapter) Name() string {
	return a.cmd.Name()
}

func (a *statsCommandAdapter) Synopsis() string {
	return a.cmd.Synopsis()
}

func (a *statsCommandAdapter) Usage() string {
	return a.cmd.Usage()
}

func (a *statsCommandAdapter) SetFlags(fs *flag.FlagSet) {
	a.cmd.SetFlags(fs)
}

func (a *statsCommandAdapter) Run(ctx *Context, args []string) int {
	// Convert cli.Context to commands.CLIContext
	cmdCtx := &commands.CLIContext{
		Stdout:      ctx.Stdout,
		Stderr:      ctx.Stderr,
		Config:      ctx.Config,
		APIEndpoint: ctx.APIEndpoint,
	}
	return a.cmd.Run(cmdCtx, args)
}

// rulesCommandAdapter adapts commands.RulesCommand to the CLICommand interface.
// This adapter also implements ParentCommand for subcommand routing.
type rulesCommandAdapter struct {
	cmd *commands.RulesCommand
}

func newRulesCommandAdapter() *rulesCommandAdapter {
	return &rulesCommandAdapter{
		cmd: commands.NewRulesCommand(),
	}
}

func (a *rulesCommandAdapter) Name() string {
	return a.cmd.Name()
}

func (a *rulesCommandAdapter) Synopsis() string {
	return a.cmd.Synopsis()
}

func (a *rulesCommandAdapter) Usage() string {
	return a.cmd.Usage()
}

func (a *rulesCommandAdapter) SetFlags(fs *flag.FlagSet) {
	a.cmd.SetFlags(fs)
}

func (a *rulesCommandAdapter) Run(ctx *Context, args []string) int {
	// Convert cli.Context to commands.CLIContext
	cmdCtx := &commands.CLIContext{
		Stdout:      ctx.Stdout,
		Stderr:      ctx.Stderr,
		Config:      ctx.Config,
		APIEndpoint: ctx.APIEndpoint,
	}
	return a.cmd.Run(cmdCtx, args)
}

func (a *rulesCommandAdapter) Subcommands() []CLICommand {
	return []CLICommand{
		newRulesListCommandAdapter(),
		newRulesSetCommandAdapter(),
	}
}

// rulesListCommandAdapter adapts commands.RulesListCommand to the CLICommand interface.
type rulesListCommandAdapter struct {
	cmd *commands.RulesListCommand
}

func newRulesListCommandAdapter() *rulesListCommandAdapter {
	return &rulesListCommandAdapter{
		cmd: commands.NewRulesListCommand(),
	}
}

func (a *rulesListCommandAdapter) Name() string {
	return a.cmd.Name()
}

func (a *rulesListCommandAdapter) Synopsis() string {
	return a.cmd.Synopsis()
}

func (a *rulesListCommandAdapter) Usage() string {
	return a.cmd.Usage()
}

func (a *rulesListCommandAdapter) SetFlags(fs *flag.FlagSet) {
	a.cmd.SetFlags(fs)
}

func (a *rulesListCommandAdapter) Run(ctx *Context, args []string) int {
	// Convert cli.Context to commands.CLIContext
	cmdCtx := &commands.CLIContext{
		Stdout:      ctx.Stdout,
		Stderr:      ctx.Stderr,
		Config:      ctx.Config,
		APIEndpoint: ctx.APIEndpoint,
	}
	return a.cmd.Run(cmdCtx, args)
}

// rulesSetCommandAdapter adapts commands.RulesSetCommand to the CLICommand interface.
type rulesSetCommandAdapter struct {
	cmd *commands.RulesSetCommand
}

func newRulesSetCommandAdapter() *rulesSetCommandAdapter {
	return &rulesSetCommandAdapter{
		cmd: commands.NewRulesSetCommand(),
	}
}

func (a *rulesSetCommandAdapter) Name() string {
	return a.cmd.Name()
}

func (a *rulesSetCommandAdapter) Synopsis() string {
	return a.cmd.Synopsis()
}

func (a *rulesSetCommandAdapter) Usage() string {
	return a.cmd.Usage()
}

func (a *rulesSetCommandAdapter) SetFlags(fs *flag.FlagSet) {
	a.cmd.SetFlags(fs)
}

func (a *rulesSetCommandAdapter) Run(ctx *Context, args []string) int {
	// Convert cli.Context to commands.CLIContext
	cmdCtx := &commands.CLIContext{
		Stdout:      ctx.Stdout,
		Stderr:      ctx.Stderr,
		Config:      ctx.Config,
		APIEndpoint: ctx.APIEndpoint,
	}
	return a.cmd.Run(cmdCtx, args)
}
