// Package bot provides the core bot implementation for JamesBot.
package bot

import (
	"context"
	"fmt"

	"jamesbot/internal/command"
	"jamesbot/internal/config"
	"jamesbot/internal/handler"
	"jamesbot/internal/middleware"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
)

// Bot represents the JamesBot Discord bot instance.
// It manages the Discord session, command registry, and event handlers.
type Bot struct {
	session     *discordgo.Session
	registry    *command.Registry
	config      *config.Config
	logger      zerolog.Logger
	middlewares []middleware.Middleware

	interactionHandler *handler.InteractionHandler
	readyHandler       *handler.ReadyHandler
}

// New creates a new Bot instance with the provided configuration and logger.
// It validates the configuration, creates a Discord session, and sets up handlers.
// Functional options can be provided to customize the bot's behavior.
//
// Returns an error if the configuration is invalid or if the Discord session
// cannot be created.
func New(cfg *config.Config, logger zerolog.Logger, opts ...Option) (*Bot, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Validate that Discord token is not empty
	if cfg.Discord.Token == "" {
		return nil, fmt.Errorf("discord token cannot be empty")
	}

	// Create Discord session
	session, err := discordgo.New("Bot " + cfg.Discord.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create discord session: %w", err)
	}

	// Set Discord intents
	session.Identify.Intents = discordgo.IntentsGuilds

	// Create bot instance
	bot := &Bot{
		session:     session,
		registry:    command.NewRegistry(logger),
		config:      cfg,
		logger:      logger,
		middlewares: make([]middleware.Middleware, 0),
	}

	// Apply functional options
	for _, opt := range opts {
		opt(bot)
	}

	// Create handlers
	bot.readyHandler = handler.NewReadyHandler(logger)

	// Create middleware chain
	var combinedMiddleware middleware.Middleware
	if len(bot.middlewares) > 0 {
		combinedMiddleware = middleware.Chain(bot.middlewares...)
	}

	bot.interactionHandler = handler.NewInteractionHandler(
		bot.registry,
		combinedMiddleware,
		logger,
	)

	return bot, nil
}

// RegisterCommand registers a command with the bot's command registry.
// The command will be available for execution once the bot is started.
//
// Returns an error if the command is nil or if a command with the same name
// is already registered.
func (b *Bot) RegisterCommand(cmd command.Command) error {
	if b == nil {
		return fmt.Errorf("bot cannot be nil")
	}
	return b.registry.Register(cmd)
}

// Start starts the bot and connects to Discord.
// It registers event handlers, opens the Discord session, and registers
// slash commands with Discord's API.
//
// The context parameter is currently unused but is included for future
// support of graceful startup cancellation.
func (b *Bot) Start(ctx context.Context) error {
	if b == nil {
		return fmt.Errorf("bot cannot be nil")
	}

	// Add event handlers
	b.session.AddHandler(b.readyHandler.Handle)
	b.session.AddHandler(b.interactionHandler.Handle)

	// Open Discord session
	if err := b.session.Open(); err != nil {
		return fmt.Errorf("failed to open discord session: %w", err)
	}

	b.logger.Info().Msg("discord session opened")

	// Register slash commands with Discord
	appCommands := b.registry.ApplicationCommands()

	guildID := b.config.Discord.GuildID
	if guildID != "" {
		b.logger.Info().
			Str("guild_id", guildID).
			Int("command_count", len(appCommands)).
			Msg("registering guild-specific commands")
	} else {
		b.logger.Info().
			Int("command_count", len(appCommands)).
			Msg("registering global commands")
	}

	for _, appCmd := range appCommands {
		_, err := b.session.ApplicationCommandCreate(
			b.session.State.User.ID,
			guildID,
			appCmd,
		)
		if err != nil {
			return fmt.Errorf("failed to register command %q: %w", appCmd.Name, err)
		}

		b.logger.Debug().
			Str("command", appCmd.Name).
			Msg("registered command")
	}

	b.logger.Info().Msg("bot started successfully")

	return nil
}

// Stop gracefully stops the bot and disconnects from Discord.
// If the configuration specifies cleanup on shutdown, it will remove
// all registered slash commands from Discord.
//
// The context parameter can be used to set a deadline for the shutdown process.
func (b *Bot) Stop(ctx context.Context) error {
	if b == nil {
		return fmt.Errorf("bot cannot be nil")
	}

	b.logger.Info().Msg("stopping bot")

	// Cleanup slash commands if configured
	if b.config.Discord.CleanupOnShutdown {
		b.logger.Info().Msg("cleaning up slash commands")

		guildID := b.config.Discord.GuildID
		commands, err := b.session.ApplicationCommands(b.session.State.User.ID, guildID)
		if err != nil {
			b.logger.Error().
				Err(err).
				Msg("failed to retrieve commands for cleanup")
		} else {
			for _, cmd := range commands {
				err := b.session.ApplicationCommandDelete(
					b.session.State.User.ID,
					guildID,
					cmd.ID,
				)
				if err != nil {
					b.logger.Error().
						Err(err).
						Str("command", cmd.Name).
						Msg("failed to delete command")
				} else {
					b.logger.Debug().
						Str("command", cmd.Name).
						Msg("deleted command")
				}
			}
		}
	}

	// Close Discord session
	if err := b.session.Close(); err != nil {
		return fmt.Errorf("failed to close discord session: %w", err)
	}

	b.logger.Info().Msg("bot stopped")

	return nil
}
