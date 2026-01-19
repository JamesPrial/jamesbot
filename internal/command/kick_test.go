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

// kickTestLogger returns a zerolog.Logger that discards output for testing.
func kickTestLogger() zerolog.Logger {
	return zerolog.New(io.Discard).Level(zerolog.Disabled)
}

// createKickTestInteraction creates a test interaction for kick command tests.
func createKickTestInteraction(userID, guildID, channelID string, options []*discordgo.ApplicationCommandInteractionDataOption) *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:        "interaction-kick-test",
			ChannelID: channelID,
			GuildID:   guildID,
			Member: &discordgo.Member{
				User: &discordgo.User{
					ID:       userID,
					Username: "moderator",
				},
			},
			User: &discordgo.User{
				ID:       userID,
				Username: "moderator",
			},
			Type: discordgo.InteractionApplicationCommand,
			Data: discordgo.ApplicationCommandInteractionData{
				ID:      "cmd-data-kick",
				Name:    "kick",
				Options: options,
			},
		},
	}
}

// createKickOptions creates options for kick command testing.
func createKickOptions(targetUserID string, reason string, includeReason bool) []*discordgo.ApplicationCommandInteractionDataOption {
	options := []*discordgo.ApplicationCommandInteractionDataOption{
		{
			Name:  "user",
			Type:  discordgo.ApplicationCommandOptionUser,
			Value: targetUserID,
		},
	}

	if includeReason {
		options = append(options, &discordgo.ApplicationCommandInteractionDataOption{
			Name:  "reason",
			Type:  discordgo.ApplicationCommandOptionString,
			Value: reason,
		})
	}

	return options
}

// createKickInteractionWithResolvedUser creates an interaction with resolved user data.
func createKickInteractionWithResolvedUser(executorID, targetUserID, guildID, channelID string, reason string, includeReason bool, targetIsBot bool) *discordgo.InteractionCreate {
	interaction := createKickTestInteraction(executorID, guildID, channelID, createKickOptions(targetUserID, reason, includeReason))

	// Add resolved user data
	interaction.Interaction.Data = discordgo.ApplicationCommandInteractionData{
		ID:      "cmd-data-kick",
		Name:    "kick",
		Options: createKickOptions(targetUserID, reason, includeReason),
		Resolved: &discordgo.ApplicationCommandInteractionDataResolved{
			Users: map[string]*discordgo.User{
				targetUserID: {
					ID:       targetUserID,
					Username: "targetuser",
					Bot:      targetIsBot,
				},
			},
		},
	}

	return interaction
}

func Test_KickCommand_Name(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "returns kick as command name",
			expected: "kick",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &command.KickCommand{}

			result := cmd.Name()

			assert.Equal(t, tt.expected, result, "Name() should return %q", tt.expected)
		})
	}
}

func Test_KickCommand_Description(t *testing.T) {
	tests := []struct {
		name        string
		containsAny []string
		notEmpty    bool
	}{
		{
			name:        "returns non-empty description with relevant keywords",
			containsAny: []string{"kick", "Kick", "user", "member", "remove", "server"},
			notEmpty:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &command.KickCommand{}

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

func Test_KickCommand_Options(t *testing.T) {
	cmd := &command.KickCommand{}
	options := cmd.Options()

	require.NotNil(t, options, "Options() should not return nil")
	require.NotEmpty(t, options, "Options() should not return empty slice")

	t.Run("has user option", func(t *testing.T) {
		var userOption *discordgo.ApplicationCommandOption
		for _, opt := range options {
			if opt.Name == "user" {
				userOption = opt
				break
			}
		}

		require.NotNil(t, userOption, "Options should contain 'user' option")
		assert.Equal(t, discordgo.ApplicationCommandOptionUser, userOption.Type,
			"user option should be of type User")
		assert.True(t, userOption.Required, "user option should be required")
		assert.NotEmpty(t, userOption.Description, "user option should have a description")
	})

	t.Run("has reason option", func(t *testing.T) {
		var reasonOption *discordgo.ApplicationCommandOption
		for _, opt := range options {
			if opt.Name == "reason" {
				reasonOption = opt
				break
			}
		}

		require.NotNil(t, reasonOption, "Options should contain 'reason' option")
		assert.Equal(t, discordgo.ApplicationCommandOptionString, reasonOption.Type,
			"reason option should be of type String")
		assert.False(t, reasonOption.Required, "reason option should be optional")
		assert.NotEmpty(t, reasonOption.Description, "reason option should have a description")
	})
}

func Test_KickCommand_Permissions(t *testing.T) {
	tests := []struct {
		name               string
		expectedPermission int64
	}{
		{
			name:               "requires KickMembers permission",
			expectedPermission: discordgo.PermissionKickMembers,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &command.KickCommand{}

			result := cmd.Permissions()

			// Check that the required permission is included in the returned permissions
			assert.True(t, result&tt.expectedPermission != 0,
				"Permissions() should include PermissionKickMembers (0x%X), got 0x%X",
				tt.expectedPermission, result)
		})
	}
}

func Test_KickCommand_Execute(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func() *command.Context
		expectError    bool
		errContains    string
		shouldNotPanic bool
	}{
		{
			name: "nil context returns error without panic",
			setupContext: func() *command.Context {
				return nil
			},
			expectError:    true,
			shouldNotPanic: true,
		},
		{
			name: "cannot kick self",
			setupContext: func() *command.Context {
				// Executor and target are the same user
				interaction := createKickInteractionWithResolvedUser(
					"user-123", "user-123", "guild-456", "channel-789",
					"no reason", true, false,
				)
				return command.NewContext(nil, interaction, kickTestLogger())
			},
			expectError: true,
			errContains: "yourself",
		},
		{
			name: "valid kick target with reason",
			setupContext: func() *command.Context {
				interaction := createKickInteractionWithResolvedUser(
					"moderator-123", "target-456", "guild-789", "channel-012",
					"Breaking rules", true, false,
				)
				return command.NewContext(nil, interaction, kickTestLogger())
			},
			// Will fail due to nil session, but should not fail validation
			expectError: true,
		},
		{
			name: "valid kick target without reason",
			setupContext: func() *command.Context {
				interaction := createKickInteractionWithResolvedUser(
					"moderator-123", "target-456", "guild-789", "channel-012",
					"", false, false,
				)
				return command.NewContext(nil, interaction, kickTestLogger())
			},
			// Will fail due to nil session, but should not fail validation
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &command.KickCommand{}
			ctx := tt.setupContext()

			if tt.shouldNotPanic {
				assert.NotPanics(t, func() {
					_ = cmd.Execute(ctx)
				}, "Execute should not panic")
			}

			err := cmd.Execute(ctx)

			if tt.expectError {
				require.Error(t, err, "Execute should return an error")
				if tt.errContains != "" {
					assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tt.errContains),
						"error should contain %q", tt.errContains)
				}
			}
		})
	}
}

func Test_KickCommand_Execute_NilContext(t *testing.T) {
	cmd := &command.KickCommand{}

	assert.NotPanics(t, func() {
		err := cmd.Execute(nil)
		assert.Error(t, err, "Execute should return error for nil context")
	}, "Execute should not panic with nil context")
}

func Test_KickCommand_Execute_NilInteraction(t *testing.T) {
	cmd := &command.KickCommand{}
	ctx := command.NewContext(nil, nil, kickTestLogger())

	assert.NotPanics(t, func() {
		_ = cmd.Execute(ctx)
	}, "Execute should not panic with nil interaction in context")
}

func Test_KickCommand_Execute_CannotKickSelf(t *testing.T) {
	cmd := &command.KickCommand{}

	// Create interaction where executor and target are the same
	interaction := createKickInteractionWithResolvedUser(
		"same-user-id", "same-user-id", "guild-123", "channel-456",
		"some reason", true, false,
	)
	ctx := command.NewContext(nil, interaction, kickTestLogger())

	err := cmd.Execute(ctx)

	require.Error(t, err, "Execute should return error when trying to kick self")
	assert.Contains(t, strings.ToLower(err.Error()), "yourself",
		"error message should indicate cannot kick yourself")
}

func Test_KickCommand_ImplementsCommandInterface(t *testing.T) {
	// This test verifies that KickCommand implements the Command interface
	// If this compiles, KickCommand satisfies command.Command
	var _ command.Command = (*command.KickCommand)(nil)
}

func Test_KickCommand_ImplementsPermissionedCommandInterface(t *testing.T) {
	// This test verifies that KickCommand implements the PermissionedCommand interface
	// If this compiles, KickCommand satisfies command.PermissionedCommand
	var _ command.PermissionedCommand = (*command.KickCommand)(nil)
}

func Test_KickCommand_CanBeRegistered(t *testing.T) {
	// Test that KickCommand can be registered in the Registry
	registry := command.NewRegistry(kickTestLogger())
	cmd := &command.KickCommand{}

	err := registry.Register(cmd)

	require.NoError(t, err, "KickCommand should be registerable")

	// Verify it can be retrieved
	retrieved, found := registry.Get("kick")
	assert.True(t, found, "kick command should be found in registry")
	assert.Equal(t, "kick", retrieved.Name())
}

func Test_KickCommand_ApplicationCommand(t *testing.T) {
	// Test that KickCommand converts correctly to ApplicationCommand
	registry := command.NewRegistry(kickTestLogger())
	cmd := &command.KickCommand{}

	err := registry.Register(cmd)
	require.NoError(t, err)

	appCmds := registry.ApplicationCommands()

	require.Len(t, appCmds, 1)
	assert.Equal(t, "kick", appCmds[0].Name)
	assert.NotEmpty(t, appCmds[0].Description)
	require.NotEmpty(t, appCmds[0].Options, "kick command should have options")

	// Verify user option exists
	var userOption *discordgo.ApplicationCommandOption
	for _, opt := range appCmds[0].Options {
		if opt.Name == "user" {
			userOption = opt
			break
		}
	}
	require.NotNil(t, userOption, "ApplicationCommand should have user option")
	assert.True(t, userOption.Required, "user option should be required in ApplicationCommand")

	// Verify permissions are set
	require.NotNil(t, appCmds[0].DefaultMemberPermissions,
		"ApplicationCommand should have DefaultMemberPermissions set")
	assert.True(t, *appCmds[0].DefaultMemberPermissions&discordgo.PermissionKickMembers != 0,
		"DefaultMemberPermissions should include KickMembers")
}

// Benchmark tests
func Benchmark_KickCommand_Name(b *testing.B) {
	cmd := &command.KickCommand{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Name()
	}
}

func Benchmark_KickCommand_Description(b *testing.B) {
	cmd := &command.KickCommand{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Description()
	}
}

func Benchmark_KickCommand_Options(b *testing.B) {
	cmd := &command.KickCommand{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Options()
	}
}

func Benchmark_KickCommand_Permissions(b *testing.B) {
	cmd := &command.KickCommand{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Permissions()
	}
}
