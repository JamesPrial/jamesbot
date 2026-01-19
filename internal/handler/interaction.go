package handler

import (
	"errors"
	"jamesbot/internal/command"
	"jamesbot/internal/middleware"
	"jamesbot/pkg/errutil"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
)

// InteractionHandler handles Discord interaction events.
// It processes application commands by looking them up in the registry
// and executing them through the middleware chain.
type InteractionHandler struct {
	registry   *command.Registry
	middleware middleware.Middleware
	logger     zerolog.Logger
}

// NewInteractionHandler creates a new interaction handler with the provided components.
// The middleware parameter can be nil if no middleware is needed.
func NewInteractionHandler(registry *command.Registry, mw middleware.Middleware, logger zerolog.Logger) *InteractionHandler {
	return &InteractionHandler{
		registry:   registry,
		middleware: mw,
		logger:     logger,
	}
}

// Handle processes interaction events from Discord.
// It currently supports ApplicationCommand interactions and routes them to
// the appropriate command handler.
func (h *InteractionHandler) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i == nil {
		h.logger.Warn().Msg("received nil interaction")
		return
	}

	// Only handle application command interactions
	if i.Type != discordgo.InteractionApplicationCommand {
		h.logger.Debug().
			Int("type", int(i.Type)).
			Msg("ignoring non-command interaction")
		return
	}

	// Get command name from interaction
	if i.Data == nil {
		h.logger.Warn().Msg("received application command interaction with nil data")
		return
	}
	data := i.ApplicationCommandData()
	commandName := data.Name

	// Look up command in registry
	cmd, exists := h.registry.Get(commandName)
	if !exists {
		h.logger.Error().
			Str("command", commandName).
			Msg("command not found in registry")

		// Respond with error message
		ctx := command.NewContext(s, i, h.logger)
		_ = ctx.RespondEphemeral("Command not found. This might be a configuration issue.")
		return
	}

	// Create command context
	ctx := command.NewContext(s, i, h.logger)

	// Create the base handler that executes the command
	handler := middleware.HandlerFunc(func(ctx *command.Context) error {
		return cmd.Execute(ctx)
	})

	// Wrap with middleware if provided
	if h.middleware != nil {
		handler = h.middleware(handler)
	}

	// Execute the command through the middleware chain
	if err := handler(ctx); err != nil {
		h.handleError(ctx, err)
	}
}

// handleError processes errors from command execution.
// It extracts user-friendly messages when available and logs the full error.
func (h *InteractionHandler) handleError(ctx *command.Context, err error) {
	if err == nil {
		return
	}

	// Log the error
	h.logger.Error().
		Err(err).
		Str("command", ctx.Interaction.ApplicationCommandData().Name).
		Str("user_id", ctx.UserID()).
		Str("guild_id", ctx.GuildID()).
		Msg("command execution failed")

	// Extract user message from UserFriendlyError if present
	userMessage := "An error occurred while executing the command."
	var userFriendlyErr errutil.UserFriendlyError
	if errors.As(err, &userFriendlyErr) {
		if userFriendlyErr.UserMessage != "" {
			userMessage = userFriendlyErr.UserMessage
		}
	}

	// Respond to the user with an ephemeral message
	if respondErr := ctx.RespondEphemeral(userMessage); respondErr != nil {
		h.logger.Error().
			Err(respondErr).
			Msg("failed to send error response to user")
	}
}
