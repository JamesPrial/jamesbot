package jamesprial

import (
	"fmt"

	"github.com/bwmarrin/discordgo"

	"jamesbot/internal/command"
)

// AboutPluginCommand shows information about the JamesPrial plugin.
type AboutPluginCommand struct{}

// Name returns the command name.
func (c *AboutPluginCommand) Name() string {
	return "jamesprial"
}

// Description returns the command description.
func (c *AboutPluginCommand) Description() string {
	return "Show information about the JamesPrial plugin"
}

// Options returns the command options.
func (c *AboutPluginCommand) Options() []*discordgo.ApplicationCommandOption {
	return nil
}

// Execute runs the about command.
func (c *AboutPluginCommand) Execute(ctx *command.Context) error {
	embed := &discordgo.MessageEmbed{
		Title:       "JamesPrial Plugin",
		Description: "An example plugin demonstrating the JamesBot plugin system.",
		Color:       0x5865F2, // Discord blurple
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Version",
				Value:  PluginVersion,
				Inline: true,
			},
			{
				Name:   "Author",
				Value:  "JamesPrial",
				Inline: true,
			},
			{
				Name:   "Commands",
				Value:  fmt.Sprintf("`/greet` - Send a greeting\n`/%s` - Show this info", PluginName),
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "JamesBot Plugin System",
		},
	}

	return ctx.RespondEmbed(embed)
}

// Ensure AboutPluginCommand implements Command interface.
var _ command.Command = (*AboutPluginCommand)(nil)
