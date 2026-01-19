// JamesBot is a Discord moderation bot built with Go.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"jamesbot/internal/bot"
	"jamesbot/internal/command"
	"jamesbot/internal/config"
	"jamesbot/internal/middleware"

	"github.com/rs/zerolog"
)

func main() {
	// Load configuration
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		// If config file not found, try loading from environment variables only
		cfg, err = config.Load("")
		if err != nil {
			panic("Failed to load configuration: " + err.Error())
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
	}

	// Register commands
	if err := b.RegisterCommand(&command.PingCommand{}); err != nil {
		logger.Fatal().Err(err).Msg("failed to register ping command")
	}
	if err := b.RegisterCommand(&command.EchoCommand{}); err != nil {
		logger.Fatal().Err(err).Msg("failed to register echo command")
	}
	if err := b.RegisterCommand(&command.KickCommand{}); err != nil {
		logger.Fatal().Err(err).Msg("failed to register kick command")
	}
	if err := b.RegisterCommand(&command.BanCommand{}); err != nil {
		logger.Fatal().Err(err).Msg("failed to register ban command")
	}
	if err := b.RegisterCommand(&command.MuteCommand{}); err != nil {
		logger.Fatal().Err(err).Msg("failed to register mute command")
	}
	if err := b.RegisterCommand(&command.WarnCommand{}); err != nil {
		logger.Fatal().Err(err).Msg("failed to register warn command")
	}

	// Start bot
	ctx := context.Background()
	if err := b.Start(ctx); err != nil {
		logger.Fatal().Err(err).Msg("failed to start bot")
	}

	// Wait for interrupt signal
	logger.Info().Msg("bot is running. Press CTRL-C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	// Graceful shutdown
	logger.Info().Msg("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(ctx, cfg.Shutdown.Timeout)
	defer cancel()

	if err := b.Stop(shutdownCtx); err != nil {
		logger.Error().Err(err).Msg("error during shutdown")
	}

	logger.Info().Msg("shutdown complete")
}
