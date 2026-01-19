package command

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
)

// Context provides command execution context and helper methods.
// It wraps the Discord session, interaction, and logger to provide
// convenient access to command execution resources.
type Context struct {
	// Session is the Discord session for API interactions.
	Session *discordgo.Session

	// Interaction contains the interaction data from Discord.
	Interaction *discordgo.InteractionCreate

	// Logger is a structured logger for command execution.
	Logger zerolog.Logger
}

// NewContext creates a new command context with the provided components.
// The logger will be enhanced with contextual fields for the command execution.
func NewContext(s *discordgo.Session, i *discordgo.InteractionCreate, logger zerolog.Logger) *Context {
	if i == nil {
		return &Context{
			Session:     s,
			Interaction: nil,
			Logger:      logger,
		}
	}

	// Enhance logger with context
	contextLogger := logger.With().
		Str("guild_id", guildIDFromInteraction(i)).
		Str("channel_id", channelIDFromInteraction(i)).
		Str("user_id", userIDFromInteraction(i)).
		Logger()

	return &Context{
		Session:     s,
		Interaction: i,
		Logger:      contextLogger,
	}
}

// Respond sends a response message to the interaction.
// This creates a public response visible to all users in the channel.
func (c *Context) Respond(content string) error {
	if c.Session == nil || c.Interaction == nil {
		return fmt.Errorf("cannot respond: session or interaction is nil")
	}

	return c.Session.InteractionRespond(c.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
}

// RespondEphemeral sends an ephemeral response message to the interaction.
// This creates a private response visible only to the user who invoked the command.
func (c *Context) RespondEphemeral(content string) error {
	if c.Session == nil || c.Interaction == nil {
		return fmt.Errorf("cannot respond: session or interaction is nil")
	}

	return c.Session.InteractionRespond(c.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

// RespondEmbed sends an embed response to the interaction.
// This creates a public response with a rich embed visible to all users.
func (c *Context) RespondEmbed(embed *discordgo.MessageEmbed) error {
	if c.Session == nil || c.Interaction == nil {
		return fmt.Errorf("cannot respond: session or interaction is nil")
	}

	if embed == nil {
		return fmt.Errorf("embed cannot be nil")
	}

	return c.Session.InteractionRespond(c.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// StringOption retrieves a string option value by name.
// Returns an empty string if the option is not found or has no value.
func (c *Context) StringOption(name string) string {
	if c.Interaction == nil || c.Interaction.ApplicationCommandData().Options == nil {
		return ""
	}

	for _, opt := range c.Interaction.ApplicationCommandData().Options {
		if opt.Name == name && opt.Type == discordgo.ApplicationCommandOptionString {
			return opt.StringValue()
		}
	}

	return ""
}

// IntOption retrieves an integer option value by name.
// Returns 0 if the option is not found or has no value.
func (c *Context) IntOption(name string) int64 {
	if c.Interaction == nil || c.Interaction.ApplicationCommandData().Options == nil {
		return 0
	}

	for _, opt := range c.Interaction.ApplicationCommandData().Options {
		if opt.Name == name && opt.Type == discordgo.ApplicationCommandOptionInteger {
			return opt.IntValue()
		}
	}

	return 0
}

// UserOption retrieves a user option value by name.
// Returns nil if the option is not found or has no value.
func (c *Context) UserOption(name string) *discordgo.User {
	if c.Interaction == nil || c.Interaction.ApplicationCommandData().Options == nil {
		return nil
	}

	for _, opt := range c.Interaction.ApplicationCommandData().Options {
		if opt.Name == name && opt.Type == discordgo.ApplicationCommandOptionUser {
			// First, try to get from resolved data (works without session)
			userID := opt.Value.(string)
			if c.Interaction.ApplicationCommandData().Resolved != nil {
				if user, ok := c.Interaction.ApplicationCommandData().Resolved.Users[userID]; ok {
					return user
				}
			}

			// Fallback to UserValue (requires session)
			return opt.UserValue(c.Session)
		}
	}

	return nil
}

// BoolOption retrieves a boolean option value by name.
// Returns false if the option is not found or has no value.
func (c *Context) BoolOption(name string) bool {
	if c.Interaction == nil || c.Interaction.ApplicationCommandData().Options == nil {
		return false
	}

	for _, opt := range c.Interaction.ApplicationCommandData().Options {
		if opt.Name == name && opt.Type == discordgo.ApplicationCommandOptionBoolean {
			return opt.BoolValue()
		}
	}

	return false
}

// UserID returns the ID of the user who invoked the command.
// Returns an empty string if the interaction is nil.
func (c *Context) UserID() string {
	return userIDFromInteraction(c.Interaction)
}

// GuildID returns the ID of the guild where the command was invoked.
// Returns an empty string if the interaction is nil or not in a guild.
func (c *Context) GuildID() string {
	return guildIDFromInteraction(c.Interaction)
}

// ChannelID returns the ID of the channel where the command was invoked.
// Returns an empty string if the interaction is nil.
func (c *Context) ChannelID() string {
	return channelIDFromInteraction(c.Interaction)
}

// userIDFromInteraction safely extracts the user ID from an interaction.
func userIDFromInteraction(i *discordgo.InteractionCreate) string {
	if i == nil {
		return ""
	}

	if i.Member != nil && i.Member.User != nil {
		return i.Member.User.ID
	}

	if i.User != nil {
		return i.User.ID
	}

	return ""
}

// guildIDFromInteraction safely extracts the guild ID from an interaction.
func guildIDFromInteraction(i *discordgo.InteractionCreate) string {
	if i == nil {
		return ""
	}

	return i.GuildID
}

// channelIDFromInteraction safely extracts the channel ID from an interaction.
func channelIDFromInteraction(i *discordgo.InteractionCreate) string {
	if i == nil {
		return ""
	}

	return i.ChannelID
}
