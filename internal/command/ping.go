package command

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// PingCommand implements a simple ping/pong command.
// It is used to check if the bot is responsive and functioning correctly.
type PingCommand struct{}

// Name returns the command name.
func (c *PingCommand) Name() string {
	return "ping"
}

// Description returns the command description.
func (c *PingCommand) Description() string {
	return "Check if the bot is responsive"
}

// Options returns the command options.
// The ping command does not accept any options.
func (c *PingCommand) Options() []*discordgo.ApplicationCommandOption {
	return nil
}

// Execute runs the ping command.
// It responds with "Pong!" to confirm the bot is responsive.
func (c *PingCommand) Execute(ctx *Context) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	return ctx.Respond("Pong!")
}
