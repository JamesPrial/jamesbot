// Package commands provides CLI command implementations for JamesBot.
package commands

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"jamesbot/internal/bot"
	"jamesbot/internal/command"
	"jamesbot/internal/config"
	"jamesbot/internal/control"
	"jamesbot/internal/middleware"
	"jamesbot/internal/plugin"
	"jamesbot/internal/plugin/plugins/jamesprial"

	"github.com/rs/zerolog"
)

// CLIContext represents the execution context for CLI commands.
// This is a local type to avoid import cycles with the cli package.
type CLIContext struct {
	Stdout      io.Writer
	Stderr      io.Writer
	Config      *config.Config
	APIEndpoint string
}

// ServeCommand implements the serve command for starting the Discord bot.
type ServeCommand struct {
	configPath string
	apiPort    int
}

// NewServeCommand creates a new ServeCommand instance.
func NewServeCommand() *ServeCommand {
	return &ServeCommand{}
}

// Name returns the name of the command.
func (c *ServeCommand) Name() string {
	return "serve"
}

// Synopsis returns a brief description of the command.
func (c *ServeCommand) Synopsis() string {
	return "Start the Discord bot server"
}

// Usage returns detailed usage information for the command.
func (c *ServeCommand) Usage() string {
	var sb strings.Builder
	sb.WriteString("Usage: jamesbot serve [options]\n\n")
	sb.WriteString("Start the Discord bot server and connect to Discord.\n\n")
	sb.WriteString("Options:\n")
	sb.WriteString("  -c, --config <path>  Path to config file (default: config/config.yaml)\n")
	sb.WriteString("  --api-port <port>    Control API port (default: 8765)\n")
	sb.WriteString("  -h, --help           Show this help message\n")
	return sb.String()
}

// SetFlags configures the command-line flags for the serve command.
func (c *ServeCommand) SetFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.configPath, "c", "config/config.yaml", "Path to config file")
	fs.StringVar(&c.configPath, "config", "config/config.yaml", "Path to config file")
	fs.IntVar(&c.apiPort, "api-port", 8765, "Control API port")
}

// Run executes the serve command.
// It accepts a CLI context with stdout/stderr and command arguments.
func (c *ServeCommand) Run(ctx *CLIContext, args []string) int {
	// Get stderr from context
	stderr := ctx.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	// Load configuration
	cfg, err := config.Load(c.configPath)
	if err != nil {
		// If config file not found, try loading from environment variables only
		cfg, err = config.Load("")
		if err != nil {
			fmt.Fprintf(stderr, "Error: Failed to load configuration: %v\n", err)
			return 1
		}
	}

	// Create logger
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	// Configure log level
	level, err := zerolog.ParseLevel(cfg.Logging.Level)
	if err != nil {
		logger.Warn().
			Str("level", cfg.Logging.Level).
			Msg("invalid log level, using info")
		level = zerolog.InfoLevel
	}
	logger = logger.Level(level)

	// Create bot with middleware
	b, err := bot.New(cfg, logger,
		bot.WithMiddleware(
			middleware.Recovery(logger),
			middleware.Logging(logger),
		),
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create bot")
		return 1
	}

	// Register core commands
	if err := c.registerCommands(b, logger); err != nil {
		logger.Fatal().Err(err).Msg("failed to register commands")
		return 1
	}

	// Load plugins
	pluginLoader := c.loadPlugins(logger)
	defer pluginLoader.ShutdownAll()

	// Register plugin commands
	for _, cmd := range pluginLoader.Commands() {
		if err := b.RegisterCommand(cmd); err != nil {
			logger.Warn().
				Str("command", cmd.Name()).
				Err(err).
				Msg("failed to register plugin command")
		} else {
			logger.Debug().
				Str("command", cmd.Name()).
				Msg("registered plugin command")
		}
	}

	// Start bot
	botCtx := context.Background()
	if err := b.Start(botCtx); err != nil {
		logger.Fatal().Err(err).Msg("failed to start bot")
		return 1
	}

	// Start control API server
	controlServer := control.NewServer(c.apiPort, b, logger)
	if err := controlServer.Start(); err != nil {
		logger.Fatal().Err(err).Msg("failed to start control API server")
		return 1
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := controlServer.Stop(shutdownCtx); err != nil {
			logger.Error().Err(err).Msg("error stopping control API server")
		}
	}()

	// Wait for interrupt signal
	logger.Info().Msg("bot is running. Press CTRL-C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	signal.Stop(stop) // Clean up signal handler

	// Graceful shutdown
	logger.Info().Msg("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(botCtx, cfg.Shutdown.Timeout)
	defer cancel()

	if err := b.Stop(shutdownCtx); err != nil {
		logger.Error().Err(err).Msg("error during shutdown")
		return 1
	}

	logger.Info().Msg("shutdown complete")
	return 0
}

// registerCommands registers all bot commands with the bot instance.
func (c *ServeCommand) registerCommands(b *bot.Bot, logger zerolog.Logger) error {
	commands := []command.Command{
		&command.PingCommand{},
		&command.EchoCommand{},
		&command.KickCommand{},
		&command.BanCommand{},
		&command.MuteCommand{},
		&command.WarnCommand{},
	}

	for _, cmd := range commands {
		if err := b.RegisterCommand(cmd); err != nil {
			return fmt.Errorf("failed to register %s command: %w", cmd.Name(), err)
		}
		logger.Debug().Str("command", cmd.Name()).Msg("registered command")
	}

	return nil
}

// loadPlugins initializes and loads all plugins.
func (c *ServeCommand) loadPlugins(logger zerolog.Logger) *plugin.Loader {
	registry := plugin.NewRegistry(logger)
	loader := plugin.NewLoader(registry, logger)

	// Load the JamesPrial plugin
	plugins := []plugin.Plugin{
		jamesprial.New(),
	}

	for _, p := range plugins {
		if err := loader.Load(p); err != nil {
			logger.Error().
				Err(err).
				Str("plugin", p.Name()).
				Msg("failed to load plugin")
		} else {
			logger.Info().
				Str("plugin", p.Name()).
				Str("version", p.Version()).
				Msg("loaded plugin")
		}
	}

	return loader
}
