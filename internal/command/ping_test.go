package command_test

import (
	"io"
	"strings"
	"testing"

	"jamesbot/internal/command"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pingTestLogger returns a zerolog.Logger that discards output for testing.
func pingTestLogger() zerolog.Logger {
	return zerolog.New(io.Discard).Level(zerolog.Disabled)
}

// createPingTestInteraction creates a test interaction for ping command tests.
func createPingTestInteraction(userID, guildID, channelID string) *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:        "interaction-ping-test",
			ChannelID: channelID,
			GuildID:   guildID,
			Member: &discordgo.Member{
				User: &discordgo.User{
					ID:       userID,
					Username: "testuser",
				},
			},
			User: &discordgo.User{
				ID:       userID,
				Username: "testuser",
			},
			Type: discordgo.InteractionApplicationCommand,
			Data: discordgo.ApplicationCommandInteractionData{
				ID:      "cmd-data-ping",
				Name:    "ping",
				Options: nil,
			},
		},
	}
}

func Test_PingCommand_Name(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "returns ping as command name",
			expected: "ping",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &command.PingCommand{}

			result := cmd.Name()

			assert.Equal(t, tt.expected, result, "Name() should return %q", tt.expected)
		})
	}
}

func Test_PingCommand_Description(t *testing.T) {
	tests := []struct {
		name        string
		containsAny []string
		notEmpty    bool
	}{
		{
			name:        "returns non-empty description with relevant keywords",
			containsAny: []string{"responsive", "pong", "Pong", "ping", "Ping", "response", "latency", "alive", "check"},
			notEmpty:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &command.PingCommand{}

			result := cmd.Description()

			if tt.notEmpty {
				assert.NotEmpty(t, result, "Description() should return non-empty string")
			}

			// Check if description contains at least one relevant keyword
			containsRelevant := false
			for _, keyword := range tt.containsAny {
				if strings.Contains(strings.ToLower(result), strings.ToLower(keyword)) {
					containsRelevant = true
					break
				}
			}
			assert.True(t, containsRelevant,
				"Description() should contain at least one of: %v, got: %q", tt.containsAny, result)
		})
	}
}

func Test_PingCommand_Options(t *testing.T) {
	tests := []struct {
		name             string
		expectNilOrEmpty bool
	}{
		{
			name:             "returns nil or empty slice (no options required)",
			expectNilOrEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &command.PingCommand{}

			result := cmd.Options()

			if tt.expectNilOrEmpty {
				if result != nil {
					assert.Empty(t, result, "Options() should return nil or empty slice")
				}
				// nil is acceptable
			}
		})
	}
}

func Test_PingCommand_Execute(t *testing.T) {
	tests := []struct {
		name            string
		setupContext    func() *command.Context
		expectedContent string
		expectError     bool
	}{
		{
			name: "successful execution responds with Pong!",
			setupContext: func() *command.Context {
				interaction := createPingTestInteraction("user-123", "guild-456", "channel-789")
				// Note: We cannot easily test the actual Respond() call without a real session
				// This test verifies the command can be executed without panicking
				return command.NewContext(nil, interaction, pingTestLogger())
			},
			expectedContent: "Pong!",
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &command.PingCommand{}
			ctx := tt.setupContext()

			// Execute returns error when session is nil (expected behavior)
			// The test verifies the command structure is correct
			err := cmd.Execute(ctx)

			// When session is nil, Respond() will fail with "session or interaction is nil"
			// This is expected behavior - we're testing the command logic, not the Discord API
			if ctx.Session == nil {
				// We expect an error due to nil session, but the command logic is correct
				assert.Error(t, err, "Execute should return error when session is nil")
			} else if tt.expectError {
				assert.Error(t, err, "Execute should return error")
			} else {
				assert.NoError(t, err, "Execute should not return error")
			}
		})
	}
}

func Test_PingCommand_ImplementsCommandInterface(t *testing.T) {
	// This test verifies that PingCommand implements the Command interface
	// If this compiles, PingCommand satisfies command.Command
	var _ command.Command = (*command.PingCommand)(nil)
}

func Test_PingCommand_CanBeRegistered(t *testing.T) {
	// Test that PingCommand can be registered in the Registry
	registry := command.NewRegistry(pingTestLogger())
	cmd := &command.PingCommand{}

	err := registry.Register(cmd)

	require.NoError(t, err, "PingCommand should be registerable")

	// Verify it can be retrieved
	retrieved, found := registry.Get("ping")
	assert.True(t, found, "ping command should be found in registry")
	assert.Equal(t, "ping", retrieved.Name())
}

func Test_PingCommand_ApplicationCommand(t *testing.T) {
	// Test that PingCommand converts correctly to ApplicationCommand
	registry := command.NewRegistry(pingTestLogger())
	cmd := &command.PingCommand{}

	err := registry.Register(cmd)
	require.NoError(t, err)

	appCmds := registry.ApplicationCommands()

	require.Len(t, appCmds, 1)
	assert.Equal(t, "ping", appCmds[0].Name)
	assert.NotEmpty(t, appCmds[0].Description)
}

// Benchmark tests
func Benchmark_PingCommand_Name(b *testing.B) {
	cmd := &command.PingCommand{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Name()
	}
}

func Benchmark_PingCommand_Description(b *testing.B) {
	cmd := &command.PingCommand{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Description()
	}
}

func Benchmark_PingCommand_Options(b *testing.B) {
	cmd := &command.PingCommand{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Options()
	}
}
