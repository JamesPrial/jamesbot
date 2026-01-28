package jamesprial

import (
	"fmt"

	"github.com/bwmarrin/discordgo"

	"jamesbot/internal/command"
)

// GreetCommand sends a personalized greeting.
type GreetCommand struct{}

// Name returns the command name.
func (c *GreetCommand) Name() string {
	return "greet"
}

// Description returns the command description.
func (c *GreetCommand) Description() string {
	return "Send a personalized greeting to a user"
}

// Options returns the command options.
func (c *GreetCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionUser,
			Name:        "user",
			Description: "The user to greet",
			Required:    false,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "message",
			Description: "Custom greeting message",
			Required:    false,
		},
	}
}

// Execute runs the greet command.
func (c *GreetCommand) Execute(ctx *command.Context) error {
	// Get the user to greet (defaults to command invoker)
	targetUser := ctx.UserOption("user")

	// Determine the target user ID
	var targetID string
	if targetUser != nil {
		targetID = targetUser.ID
	} else {
		// Fall back to command invoker
		targetID = ctx.UserID()
		if targetID == "" {
			return fmt.Errorf("unable to determine target user")
		}
	}

	// Get custom message or use default
	message := ctx.StringOption("message")
	if message == "" {
		message = "Hello"
	}

	greeting := fmt.Sprintf("%s, <@%s>! Welcome to the server!", message, targetID)

	return ctx.Respond(greeting)
}

// Ensure GreetCommand implements Command interface.
var _ command.Command = (*GreetCommand)(nil)
