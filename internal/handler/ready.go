// Package handler provides Discord event handlers for JamesBot.
package handler

import (
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
)

// ReadyHandler handles the Discord Ready event.
// It logs information about the bot's connection to Discord.
type ReadyHandler struct {
	logger zerolog.Logger
}

// NewReadyHandler creates a new ready event handler with the provided logger.
func NewReadyHandler(logger zerolog.Logger) *ReadyHandler {
	return &ReadyHandler{
		logger: logger,
	}
}

// Handle processes the Ready event from Discord.
// It logs the bot's username, discriminator, and guild count.
func (h *ReadyHandler) Handle(s *discordgo.Session, r *discordgo.Ready) {
	if r == nil || r.User == nil {
		h.logger.Warn().Msg("received ready event with nil data")
		return
	}

	guildCount := 0
	if r.Guilds != nil {
		guildCount = len(r.Guilds)
	}

	h.logger.Info().
		Str("username", r.User.Username).
		Str("discriminator", r.User.Discriminator).
		Int("guild_count", guildCount).
		Msg("bot ready")
}
