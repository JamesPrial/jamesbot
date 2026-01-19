package command

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"jamesbot/pkg/errutil"
)

// EchoCommand repeats the user's input back to them.
// This command is useful for testing user input handling and validation.
type EchoCommand struct{}

// Name returns the command name.
func (c *EchoCommand) Name() string {
	return "echo"
}

// Description returns the command description.
func (c *EchoCommand) Description() string {
	return "Repeat your message back to you"
}

// Options returns the command options.
// The echo command requires a text parameter.
func (c *EchoCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "text",
			Description: "The text to echo back",
			Required:    true,
		},
	}
}

// Execute runs the echo command.
// It retrieves the text option and echoes it back to the user.
// Returns a ValidationError if the text is empty.
func (c *EchoCommand) Execute(ctx *Context) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	text := ctx.StringOption("text")
	if text == "" {
		return &errutil.ValidationError{Field: "text", Message: "text cannot be empty"}
	}
	return ctx.Respond(text)
}
