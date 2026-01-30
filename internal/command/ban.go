package command

import (
	"fmt"

	"jamesbot/pkg/errutil"

	"github.com/bwmarrin/discordgo"
)

// BanCommand implements a command to ban members from the server.
// It requires the Ban Members permission to execute.
type BanCommand struct{}

// Name returns the command name.
func (c *BanCommand) Name() string {
	return "ban"
}

// Description returns the command description.
func (c *BanCommand) Description() string {
	return "Ban a member from the server"
}

// Permissions returns the required Discord permissions.
// Users must have the Ban Members permission to execute this command.
func (c *BanCommand) Permissions() int64 {
	return discordgo.PermissionBanMembers
}

// Options returns the command options.
// The ban command accepts a user, an optional reason, and optional message deletion days.
func (c *BanCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionUser,
			Name:        "user",
			Description: "The user to ban",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "reason",
			Description: "The reason for banning this user",
			Required:    false,
		},
		{
			Type:        discordgo.ApplicationCommandOptionInteger,
			Name:        "delete_days",
			Description: "Number of days of messages to delete (0-7)",
			Required:    false,
			MinValue:    func() *float64 { v := 0.0; return &v }(),
			MaxValue:    7.0,
		},
	}
}

// Execute runs the ban command.
// It bans the specified user from the server with an optional reason and message deletion.
func (c *BanCommand) Execute(ctx *Context) error {
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

	// Validate cannot ban self
	if targetUser.ID == ctx.UserID() {
		return errutil.UserFriendlyError{
			UserMessage: "You cannot ban yourself.",
			Err:         fmt.Errorf("user attempted to ban yourself"),
		}
	}

	// Validate cannot ban bots
	if targetUser.Bot {
		return errutil.UserFriendlyError{
			UserMessage: "You cannot ban bots.",
			Err:         fmt.Errorf("user attempted to ban a bot"),
		}
	}

	// Get optional reason
	reason := ctx.StringOption("reason")
	if reason == "" {
		reason = "No reason provided"
	}

	// Get optional delete days (defaults to 0)
	deleteDays := int(ctx.IntOption("delete_days"))
	if deleteDays < 0 || deleteDays > 7 {
		return errutil.ValidationError{
			Field:   "delete_days",
			Message: "delete_days must be between 0 and 7",
		}
	}

	// Get guild ID
	guildID := ctx.GuildID()
	if guildID == "" {
		return errutil.UserFriendlyError{
			UserMessage: "This command can only be used in a server.",
			Err:         fmt.Errorf("ban command used outside of guild"),
		}
	}

	// Check session before making Discord API calls
	if ctx.Session == nil {
		return fmt.Errorf("session cannot be nil")
	}

	// Perform the ban
	err := ctx.Session.GuildBanCreateWithReason(guildID, targetUser.ID, reason, deleteDays)
	if err != nil {
		return errutil.UserFriendlyError{
			UserMessage: fmt.Sprintf("Failed to ban %s. I may lack permissions or the user may have a higher role.", targetUser.Username),
			Err:         fmt.Errorf("failed to ban user %s: %w", targetUser.ID, err),
		}
	}

	// Respond with success
	successMsg := fmt.Sprintf("Successfully banned %s#%s. Reason: %s", targetUser.Username, targetUser.Discriminator, reason)
	if deleteDays > 0 {
		successMsg += fmt.Sprintf(" (Deleted %d days of messages)", deleteDays)
	}
	return ctx.RespondEphemeral(successMsg)
}
