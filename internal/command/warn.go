package command

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"jamesbot/pkg/errutil"
)

// WarnCommand implements a command to warn members.
// It sends a direct message to the user with the warning.
// It requires the Moderate Members permission to execute.
type WarnCommand struct{}

// Name returns the command name.
func (c *WarnCommand) Name() string {
	return "warn"
}

// Description returns the command description.
func (c *WarnCommand) Description() string {
	return "Warn a member"
}

// Permissions returns the required Discord permissions.
// Users must have the Moderate Members permission to execute this command.
func (c *WarnCommand) Permissions() int64 {
	return discordgo.PermissionModerateMembers
}

// Options returns the command options.
// The warn command accepts a user and a required reason.
func (c *WarnCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionUser,
			Name:        "user",
			Description: "The user to warn",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "reason",
			Description: "The reason for warning this user",
			Required:    true,
		},
	}
}

// Execute runs the warn command.
// It sends a DM to the target user and confirms the warning to the moderator.
func (c *WarnCommand) Execute(ctx *Context) error {
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

	// Validate cannot warn self
	if targetUser.ID == ctx.UserID() {
		return errutil.UserFriendlyError{
			UserMessage: "You cannot warn yourself.",
			Err:         fmt.Errorf("user attempted to warn yourself"),
		}
	}

	// Validate cannot warn bots
	if targetUser.Bot {
		return errutil.UserFriendlyError{
			UserMessage: "You cannot warn bots.",
			Err:         fmt.Errorf("user attempted to warn a bot"),
		}
	}

	// Get required reason
	reason := ctx.StringOption("reason")
	if reason == "" {
		return errutil.ValidationError{
			Field:   "reason",
			Message: "reason is required",
		}
	}

	// Get guild ID for context
	guildID := ctx.GuildID()
	if guildID == "" {
		return errutil.UserFriendlyError{
			UserMessage: "This command can only be used in a server.",
			Err:         fmt.Errorf("warn command used outside of guild"),
		}
	}

	// Check session before making Discord API calls
	if ctx.Session == nil {
		return fmt.Errorf("session cannot be nil")
	}

	// Get guild name for the warning message
	guild, err := ctx.Session.Guild(guildID)
	var guildName string
	if err == nil && guild != nil {
		guildName = guild.Name
	} else {
		guildName = "this server"
	}

	// Attempt to send a DM to the user
	dmChannel, err := ctx.Session.UserChannelCreate(targetUser.ID)
	dmSent := false
	if err == nil && dmChannel != nil {
		warningMsg := fmt.Sprintf("You have been warned in %s.\nReason: %s", guildName, reason)
		_, err = ctx.Session.ChannelMessageSend(dmChannel.ID, warningMsg)
		if err == nil {
			dmSent = true
		}
	}

	// Respond with confirmation
	var responseMsg string
	if dmSent {
		responseMsg = fmt.Sprintf("Successfully warned %s#%s. They have been notified via DM.\nReason: %s",
			targetUser.Username, targetUser.Discriminator, reason)
	} else {
		responseMsg = fmt.Sprintf("Successfully warned %s#%s. (Unable to send DM - user may have DMs disabled)\nReason: %s",
			targetUser.Username, targetUser.Discriminator, reason)
	}

	return ctx.RespondEphemeral(responseMsg)
}
