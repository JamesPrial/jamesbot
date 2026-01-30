package command

import (
	"fmt"
	"strings"
	"time"

	"jamesbot/pkg/errutil"

	"github.com/bwmarrin/discordgo"
)

// MuteCommand implements a command to timeout/mute members in the server.
// It requires the Moderate Members permission to execute.
type MuteCommand struct{}

// Name returns the command name.
func (c *MuteCommand) Name() string {
	return "mute"
}

// Description returns the command description.
func (c *MuteCommand) Description() string {
	return "Timeout a member (1m to 28d)"
}

// Permissions returns the required Discord permissions.
// Users must have the Moderate Members permission to execute this command.
func (c *MuteCommand) Permissions() int64 {
	return discordgo.PermissionModerateMembers
}

// Options returns the command options.
// The mute command accepts a user, a duration, and an optional reason.
func (c *MuteCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionUser,
			Name:        "user",
			Description: "The user to timeout",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "duration",
			Description: "Timeout duration (e.g., 1h, 30m, 1d)",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "reason",
			Description: "The reason for timing out this user",
			Required:    false,
		},
	}
}

// Execute runs the mute command.
// It applies a timeout to the specified user for the given duration.
func (c *MuteCommand) Execute(ctx *Context) error {
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

	// Validate cannot mute self
	if targetUser.ID == ctx.UserID() {
		return errutil.UserFriendlyError{
			UserMessage: "You cannot timeout yourself.",
			Err:         fmt.Errorf("user attempted to mute yourself"),
		}
	}

	// Validate cannot mute bots
	if targetUser.Bot {
		return errutil.UserFriendlyError{
			UserMessage: "You cannot timeout bots.",
			Err:         fmt.Errorf("user attempted to mute a bot"),
		}
	}

	// Get and parse duration
	durationStr := ctx.StringOption("duration")
	if durationStr == "" {
		return errutil.ValidationError{
			Field:   "duration",
			Message: "duration is required",
		}
	}

	// Normalize duration string (support "d" for days)
	durationStr = strings.ToLower(durationStr)
	durationStr = strings.ReplaceAll(durationStr, "d", "h")
	// If it was days, multiply hours by 24
	isDays := strings.Contains(ctx.StringOption("duration"), "d")

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return errutil.UserFriendlyError{
			UserMessage: "Invalid duration format. Use formats like: 1h, 30m, 2d",
			Err:         fmt.Errorf("failed to parse duration %s: %w", durationStr, err),
		}
	}

	// If original input was in days, adjust the duration
	if isDays {
		duration = duration * 24
	}

	// Validate duration is between 1 minute and 28 days
	minDuration := time.Minute
	maxDuration := 28 * 24 * time.Hour

	if duration < minDuration {
		return errutil.ValidationError{
			Field:   "duration",
			Message: "duration must be at least 1 minute",
		}
	}

	if duration > maxDuration {
		return errutil.ValidationError{
			Field:   "duration",
			Message: "duration cannot exceed 28 days",
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
			Err:         fmt.Errorf("mute command used outside of guild"),
		}
	}

	// Check session before making Discord API calls
	if ctx.Session == nil {
		return fmt.Errorf("session cannot be nil")
	}

	// Calculate timeout end time
	timeoutUntil := time.Now().Add(duration)

	// Perform the timeout
	err = ctx.Session.GuildMemberTimeout(guildID, targetUser.ID, &timeoutUntil)
	if err != nil {
		return errutil.UserFriendlyError{
			UserMessage: fmt.Sprintf("Failed to timeout %s. I may lack permissions or the user may have a higher role.", targetUser.Username),
			Err:         fmt.Errorf("failed to timeout user %s: %w", targetUser.ID, err),
		}
	}

	// Respond with success
	successMsg := fmt.Sprintf("Successfully timed out %s#%s for %s. Reason: %s",
		targetUser.Username, targetUser.Discriminator, formatDuration(duration), reason)
	return ctx.RespondEphemeral(successMsg)
}

// formatDuration formats a duration into a human-readable string.
func formatDuration(d time.Duration) string {
	if d >= 24*time.Hour {
		days := d / (24 * time.Hour)
		remainder := d % (24 * time.Hour)
		if remainder == 0 {
			return fmt.Sprintf("%dd", days)
		}
		hours := remainder / time.Hour
		if hours > 0 {
			return fmt.Sprintf("%dd %dh", days, hours)
		}
		return fmt.Sprintf("%dd", days)
	}

	if d >= time.Hour {
		hours := d / time.Hour
		remainder := d % time.Hour
		if remainder == 0 {
			return fmt.Sprintf("%dh", hours)
		}
		minutes := remainder / time.Minute
		if minutes > 0 {
			return fmt.Sprintf("%dh %dm", hours, minutes)
		}
		return fmt.Sprintf("%dh", hours)
	}

	minutes := d / time.Minute
	return fmt.Sprintf("%dm", minutes)
}
