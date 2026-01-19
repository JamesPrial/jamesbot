// Package command provides the command framework for JamesBot.
package command

import "github.com/bwmarrin/discordgo"

// Command defines the interface for bot commands.
// All bot commands must implement this interface to be registered and executed.
type Command interface {
	// Name returns the command name as it will appear in Discord.
	// This should be lowercase and contain no spaces.
	Name() string

	// Description returns a brief description of what the command does.
	// This is displayed to users in Discord's command picker.
	Description() string

	// Options returns the command's application command options.
	// This defines the arguments and parameters the command accepts.
	Options() []*discordgo.ApplicationCommandOption

	// Execute runs the command with the provided context.
	// It should return an error if the command execution fails.
	Execute(ctx *Context) error
}

// PermissionedCommand is an optional interface that commands can implement
// to specify required Discord permissions.
// If a command implements this interface, the bot should verify that
// the user has the required permissions before executing the command.
type PermissionedCommand interface {
	Command

	// Permissions returns the required Discord permissions as a bitfield.
	// Use discordgo.Permission* constants to construct this value.
	Permissions() int64
}
