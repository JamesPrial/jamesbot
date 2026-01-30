package command

import (
	"fmt"

	"jamesbot/pkg/errutil"

	"github.com/bwmarrin/discordgo"
)

// KickCommand implements a command to kick members from the server.
// It requires the Kick Members permission to execute.
type KickCommand struct{}

// Name returns the command name.
func (c *KickCommand) Name() string {
	return "kick"
}

// Description returns the command description.
func (c *KickCommand) Description() string {
	return "Kick a member from the server"
}

// Permissions returns the required Discord permissions.
// Users must have the Kick Members permission to execute this command.
func (c *KickCommand) Permissions() int64 {
	return discordgo.PermissionKickMembers
}

// Options returns the command options.
// The kick command accepts a user and an optional reason.
func (c *KickCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionUser,
			Name:        "user",
			Description: "The user to kick",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "reason",
			Description: "The reason for kicking this user",
			Required:    false,
		},
	}
}

// Execute runs the kick command.
// It kicks the specified user from the server with an optional reason.
func (c *KickCommand) Execute(ctx *Context) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}

	// Get the target user
	targetUser := ctx.UserOption("user")
	if targetUser == nil {
		return errutil.ValidationError{
			Field:   "user",
			Message: "user is required",
		}
	}

	// Validate cannot kick self
	if targetUser.ID == ctx.UserID() {
		return errutil.UserFriendlyError{
			UserMessage: "You cannot kick yourself.",
			Err:         fmt.Errorf("user attempted to kick yourself"),
		}
	}

	// Validate cannot kick bots
	if targetUser.Bot {
		return errutil.UserFriendlyError{
			UserMessage: "You cannot kick bots.",
			Err:         fmt.Errorf("user attempted to kick a bot"),
		}
	}

	// Get optional reason
	reason := ctx.StringOption("reason")
	if reason == "" {
		reason = "No reason provided"
	}

	// Get guild ID
	guildID := ctx.GuildID()
	if guildID == "" {
		return errutil.UserFriendlyError{
			UserMessage: "This command can only be used in a server.",
			Err:         fmt.Errorf("kick command used outside of guild"),
		}
	}

	// Check session before making Discord API calls
	if ctx.Session == nil {
		return fmt.Errorf("session cannot be nil")
	}

	// Perform the kick
	err := ctx.Session.GuildMemberDeleteWithReason(guildID, targetUser.ID, reason)
	if err != nil {
		return errutil.UserFriendlyError{
			UserMessage: fmt.Sprintf("Failed to kick %s. I may lack permissions or the user may have a higher role.", targetUser.Username),
			Err:         fmt.Errorf("failed to kick user %s: %w", targetUser.ID, err),
		}
	}

	// Respond with success
	successMsg := fmt.Sprintf("Successfully kicked %s#%s. Reason: %s", targetUser.Username, targetUser.Discriminator, reason)
	return ctx.RespondEphemeral(successMsg)
}
