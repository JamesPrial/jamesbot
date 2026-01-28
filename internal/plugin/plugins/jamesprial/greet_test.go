package jamesprial

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestGreetCommand_Name(t *testing.T) {
	cmd := &GreetCommand{}
	if cmd.Name() != "greet" {
		t.Errorf("Name() = %q, want %q", cmd.Name(), "greet")
	}
}

func TestGreetCommand_Description(t *testing.T) {
	cmd := &GreetCommand{}
	if cmd.Description() == "" {
		t.Error("Description() should not be empty")
	}
}

func TestGreetCommand_Options(t *testing.T) {
	cmd := &GreetCommand{}
	opts := cmd.Options()

	if len(opts) != 2 {
		t.Errorf("Options() returned %d options, want 2", len(opts))
		return
	}

	// Check user option
	userOpt := opts[0]
	if userOpt.Name != "user" {
		t.Errorf("Options()[0].Name = %q, want %q", userOpt.Name, "user")
	}
	if userOpt.Type != discordgo.ApplicationCommandOptionUser {
		t.Errorf("Options()[0].Type = %d, want %d", userOpt.Type, discordgo.ApplicationCommandOptionUser)
	}
	if userOpt.Required {
		t.Error("User option should not be required")
	}

	// Check message option
	msgOpt := opts[1]
	if msgOpt.Name != "message" {
		t.Errorf("Options()[1].Name = %q, want %q", msgOpt.Name, "message")
	}
	if msgOpt.Type != discordgo.ApplicationCommandOptionString {
		t.Errorf("Options()[1].Type = %d, want %d", msgOpt.Type, discordgo.ApplicationCommandOptionString)
	}
	if msgOpt.Required {
		t.Error("Message option should not be required")
	}
}
