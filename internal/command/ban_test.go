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

// banTestLogger returns a zerolog.Logger that discards output for testing.
func banTestLogger() zerolog.Logger {
	return zerolog.New(io.Discard).Level(zerolog.Disabled)
}

// createBanTestInteraction creates a test interaction for ban command tests.
func createBanTestInteraction(userID, guildID, channelID string, options []*discordgo.ApplicationCommandInteractionDataOption) *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:        "interaction-ban-test",
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
				ID:      "cmd-data-ban",
				Name:    "ban",
				Options: options,
			},
		},
	}
}

// createBanOptions creates options for ban command testing.
func createBanOptions(targetUserID string, deleteDays int64, includeDeleteDays bool, reason string, includeReason bool) []*discordgo.ApplicationCommandInteractionDataOption {
	options := []*discordgo.ApplicationCommandInteractionDataOption{
		{
			Name:  "user",
			Type:  discordgo.ApplicationCommandOptionUser,
			Value: targetUserID,
		},
	}

	if includeDeleteDays {
		options = append(options, &discordgo.ApplicationCommandInteractionDataOption{
			Name:  "delete_days",
			Type:  discordgo.ApplicationCommandOptionInteger,
			Value: float64(deleteDays), // JSON numbers are float64
		})
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

// createBanInteractionWithResolvedUser creates an interaction with resolved user data.
func createBanInteractionWithResolvedUser(executorID, targetUserID, guildID, channelID string, deleteDays int64, includeDeleteDays bool, reason string, includeReason bool, targetIsBot bool) *discordgo.InteractionCreate {
	interaction := createBanTestInteraction(executorID, guildID, channelID, createBanOptions(targetUserID, deleteDays, includeDeleteDays, reason, includeReason))

	// Add resolved user data
	interaction.Interaction.Data = discordgo.ApplicationCommandInteractionData{
		ID:      "cmd-data-ban",
		Name:    "ban",
		Options: createBanOptions(targetUserID, deleteDays, includeDeleteDays, reason, includeReason),
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

func Test_BanCommand_Name(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "returns ban as command name",
			expected: "ban",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &command.BanCommand{}

			result := cmd.Name()

			assert.Equal(t, tt.expected, result, "Name() should return %q", tt.expected)
		})
	}
}

func Test_BanCommand_Description(t *testing.T) {
	tests := []struct {
		name        string
		containsAny []string
		notEmpty    bool
	}{
		{
			name:        "returns non-empty description with relevant keywords",
			containsAny: []string{"ban", "Ban", "user", "member", "server", "permanently"},
			notEmpty:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &command.BanCommand{}

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

func Test_BanCommand_Options(t *testing.T) {
	cmd := &command.BanCommand{}
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

	t.Run("has delete_days option", func(t *testing.T) {
		var deleteDaysOption *discordgo.ApplicationCommandOption
		for _, opt := range options {
			if opt.Name == "delete_days" {
				deleteDaysOption = opt
				break
			}
		}

		require.NotNil(t, deleteDaysOption, "Options should contain 'delete_days' option")
		assert.Equal(t, discordgo.ApplicationCommandOptionInteger, deleteDaysOption.Type,
			"delete_days option should be of type Integer")
		assert.False(t, deleteDaysOption.Required, "delete_days option should be optional")
		assert.NotEmpty(t, deleteDaysOption.Description, "delete_days option should have a description")

		// Check min/max values if specified
		if deleteDaysOption.MinValue != nil {
			assert.GreaterOrEqual(t, *deleteDaysOption.MinValue, float64(0),
				"delete_days min value should be >= 0")
		}
		if deleteDaysOption.MaxValue != 0 {
			assert.LessOrEqual(t, deleteDaysOption.MaxValue, float64(7),
				"delete_days max value should be <= 7")
		}
	})

	t.Run("has reason option", func(t *testing.T) {
		var reasonOption *discordgo.ApplicationCommandOption
		for _, opt := range options {
			if opt.Name == "reason" {
				reasonOption = opt
				break
			}
		}

		// reason option may be present but optional
		if reasonOption != nil {
			assert.Equal(t, discordgo.ApplicationCommandOptionString, reasonOption.Type,
				"reason option should be of type String")
			assert.False(t, reasonOption.Required, "reason option should be optional")
		}
	})
}

func Test_BanCommand_Options_DeleteDaysRange(t *testing.T) {
	cmd := &command.BanCommand{}
	options := cmd.Options()

	var deleteDaysOption *discordgo.ApplicationCommandOption
	for _, opt := range options {
		if opt.Name == "delete_days" {
			deleteDaysOption = opt
			break
		}
	}

	require.NotNil(t, deleteDaysOption, "delete_days option must exist")

	t.Run("delete_days has valid range 0-7", func(t *testing.T) {
		// Discord allows specifying min/max on integer options
		// The valid range should be 0-7 days
		if deleteDaysOption.MinValue != nil {
			assert.Equal(t, float64(0), *deleteDaysOption.MinValue,
				"delete_days min value should be 0")
		}
		if deleteDaysOption.MaxValue != 0 {
			assert.Equal(t, float64(7), deleteDaysOption.MaxValue,
				"delete_days max value should be 7")
		}
	})
}

func Test_BanCommand_Permissions(t *testing.T) {
	tests := []struct {
		name               string
		expectedPermission int64
	}{
		{
			name:               "requires BanMembers permission",
			expectedPermission: discordgo.PermissionBanMembers,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &command.BanCommand{}

			result := cmd.Permissions()

			// Check that the required permission is included in the returned permissions
			assert.True(t, result&tt.expectedPermission != 0,
				"Permissions() should include PermissionBanMembers (0x%X), got 0x%X",
				tt.expectedPermission, result)
		})
	}
}

func Test_BanCommand_Execute(t *testing.T) {
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
			name: "cannot ban self",
			setupContext: func() *command.Context {
				// Executor and target are the same user
				interaction := createBanInteractionWithResolvedUser(
					"user-123", "user-123", "guild-456", "channel-789",
					0, false, "no reason", true, false,
				)
				return command.NewContext(nil, interaction, banTestLogger())
			},
			expectError: true,
			errContains: "yourself",
		},
		{
			name: "valid ban target without delete_days",
			setupContext: func() *command.Context {
				interaction := createBanInteractionWithResolvedUser(
					"moderator-123", "target-456", "guild-789", "channel-012",
					0, false, "Breaking rules", true, false,
				)
				return command.NewContext(nil, interaction, banTestLogger())
			},
			// Will fail due to nil session, but should not fail validation
			expectError: true,
		},
		{
			name: "valid ban target with delete_days",
			setupContext: func() *command.Context {
				interaction := createBanInteractionWithResolvedUser(
					"moderator-123", "target-456", "guild-789", "channel-012",
					7, true, "Spam messages", true, false,
				)
				return command.NewContext(nil, interaction, banTestLogger())
			},
			// Will fail due to nil session, but should not fail validation
			expectError: true,
		},
		{
			name: "valid ban target with delete_days at boundary 0",
			setupContext: func() *command.Context {
				interaction := createBanInteractionWithResolvedUser(
					"moderator-123", "target-456", "guild-789", "channel-012",
					0, true, "", false, false,
				)
				return command.NewContext(nil, interaction, banTestLogger())
			},
			expectError: true,
		},
		{
			name: "valid ban target with delete_days at boundary 7",
			setupContext: func() *command.Context {
				interaction := createBanInteractionWithResolvedUser(
					"moderator-123", "target-456", "guild-789", "channel-012",
					7, true, "", false, false,
				)
				return command.NewContext(nil, interaction, banTestLogger())
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &command.BanCommand{}
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

func Test_BanCommand_Execute_NilContext(t *testing.T) {
	cmd := &command.BanCommand{}

	assert.NotPanics(t, func() {
		err := cmd.Execute(nil)
		assert.Error(t, err, "Execute should return error for nil context")
	}, "Execute should not panic with nil context")
}

func Test_BanCommand_Execute_NilInteraction(t *testing.T) {
	cmd := &command.BanCommand{}
	ctx := command.NewContext(nil, nil, banTestLogger())

	assert.NotPanics(t, func() {
		_ = cmd.Execute(ctx)
	}, "Execute should not panic with nil interaction in context")
}

func Test_BanCommand_Execute_CannotBanSelf(t *testing.T) {
	cmd := &command.BanCommand{}

	// Create interaction where executor and target are the same
	interaction := createBanInteractionWithResolvedUser(
		"same-user-id", "same-user-id", "guild-123", "channel-456",
		0, false, "some reason", true, false,
	)
	ctx := command.NewContext(nil, interaction, banTestLogger())

	err := cmd.Execute(ctx)

	require.Error(t, err, "Execute should return error when trying to ban self")
	assert.Contains(t, strings.ToLower(err.Error()), "yourself",
		"error message should indicate cannot ban yourself")
}

func Test_BanCommand_ImplementsCommandInterface(t *testing.T) {
	// This test verifies that BanCommand implements the Command interface
	// If this compiles, BanCommand satisfies command.Command
	var _ command.Command = (*command.BanCommand)(nil)
}

func Test_BanCommand_ImplementsPermissionedCommandInterface(t *testing.T) {
	// This test verifies that BanCommand implements the PermissionedCommand interface
	// If this compiles, BanCommand satisfies command.PermissionedCommand
	var _ command.PermissionedCommand = (*command.BanCommand)(nil)
}

func Test_BanCommand_CanBeRegistered(t *testing.T) {
	// Test that BanCommand can be registered in the Registry
	registry := command.NewRegistry(banTestLogger())
	cmd := &command.BanCommand{}

	err := registry.Register(cmd)

	require.NoError(t, err, "BanCommand should be registerable")

	// Verify it can be retrieved
	retrieved, found := registry.Get("ban")
	assert.True(t, found, "ban command should be found in registry")
	assert.Equal(t, "ban", retrieved.Name())
}

func Test_BanCommand_ApplicationCommand(t *testing.T) {
	// Test that BanCommand converts correctly to ApplicationCommand
	registry := command.NewRegistry(banTestLogger())
	cmd := &command.BanCommand{}

	err := registry.Register(cmd)
	require.NoError(t, err)

	appCmds := registry.ApplicationCommands()

	require.Len(t, appCmds, 1)
	assert.Equal(t, "ban", appCmds[0].Name)
	assert.NotEmpty(t, appCmds[0].Description)
	require.NotEmpty(t, appCmds[0].Options, "ban command should have options")

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

	// Verify delete_days option exists
	var deleteDaysOption *discordgo.ApplicationCommandOption
	for _, opt := range appCmds[0].Options {
		if opt.Name == "delete_days" {
			deleteDaysOption = opt
			break
		}
	}
	require.NotNil(t, deleteDaysOption, "ApplicationCommand should have delete_days option")
	assert.False(t, deleteDaysOption.Required, "delete_days option should be optional in ApplicationCommand")

	// Verify permissions are set
	require.NotNil(t, appCmds[0].DefaultMemberPermissions,
		"ApplicationCommand should have DefaultMemberPermissions set")
	assert.True(t, *appCmds[0].DefaultMemberPermissions&discordgo.PermissionBanMembers != 0,
		"DefaultMemberPermissions should include BanMembers")
}

// Benchmark tests
func Benchmark_BanCommand_Name(b *testing.B) {
	cmd := &command.BanCommand{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Name()
	}
}

func Benchmark_BanCommand_Description(b *testing.B) {
	cmd := &command.BanCommand{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Description()
	}
}

func Benchmark_BanCommand_Options(b *testing.B) {
	cmd := &command.BanCommand{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Options()
	}
}

func Benchmark_BanCommand_Permissions(b *testing.B) {
	cmd := &command.BanCommand{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Permissions()
	}
}
