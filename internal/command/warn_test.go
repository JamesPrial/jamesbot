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

// warnTestLogger returns a zerolog.Logger that discards output for testing.
func warnTestLogger() zerolog.Logger {
	return zerolog.New(io.Discard).Level(zerolog.Disabled)
}

// createWarnTestInteraction creates a test interaction for warn command tests.
func createWarnTestInteraction(userID, guildID, channelID string, options []*discordgo.ApplicationCommandInteractionDataOption) *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:        "interaction-warn-test",
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
				ID:      "cmd-data-warn",
				Name:    "warn",
				Options: options,
			},
		},
	}
}

// createWarnOptions creates options for warn command testing.
func createWarnOptions(targetUserID string, reason string) []*discordgo.ApplicationCommandInteractionDataOption {
	return []*discordgo.ApplicationCommandInteractionDataOption{
		{
			Name:  "user",
			Type:  discordgo.ApplicationCommandOptionUser,
			Value: targetUserID,
		},
		{
			Name:  "reason",
			Type:  discordgo.ApplicationCommandOptionString,
			Value: reason,
		},
	}
}

// createWarnInteractionWithResolvedUser creates an interaction with resolved user data.
func createWarnInteractionWithResolvedUser(executorID, targetUserID, guildID, channelID string, reason string, targetIsBot bool) *discordgo.InteractionCreate {
	interaction := createWarnTestInteraction(executorID, guildID, channelID, createWarnOptions(targetUserID, reason))

	// Add resolved user data
	interaction.Interaction.Data = discordgo.ApplicationCommandInteractionData{
		ID:      "cmd-data-warn",
		Name:    "warn",
		Options: createWarnOptions(targetUserID, reason),
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

func Test_WarnCommand_Name(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "returns warn as command name",
			expected: "warn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &command.WarnCommand{}

			result := cmd.Name()

			assert.Equal(t, tt.expected, result, "Name() should return %q", tt.expected)
		})
	}
}

func Test_WarnCommand_Description(t *testing.T) {
	tests := []struct {
		name        string
		containsAny []string
		notEmpty    bool
	}{
		{
			name:        "returns non-empty description with relevant keywords",
			containsAny: []string{"warn", "Warn", "warning", "user", "member", "issue"},
			notEmpty:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &command.WarnCommand{}

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

func Test_WarnCommand_Options(t *testing.T) {
	cmd := &command.WarnCommand{}
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
		assert.True(t, reasonOption.Required, "reason option should be required")
		assert.NotEmpty(t, reasonOption.Description, "reason option should have a description")
	})
}

func Test_WarnCommand_Permissions(t *testing.T) {
	tests := []struct {
		name               string
		expectedPermission int64
	}{
		{
			name:               "requires ModerateMembers permission",
			expectedPermission: discordgo.PermissionModerateMembers,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &command.WarnCommand{}

			result := cmd.Permissions()

			// Check that the required permission is included in the returned permissions
			assert.True(t, result&tt.expectedPermission != 0,
				"Permissions() should include PermissionModerateMembers (0x%X), got 0x%X",
				tt.expectedPermission, result)
		})
	}
}

func Test_WarnCommand_Execute(t *testing.T) {
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
			name: "cannot warn self",
			setupContext: func() *command.Context {
				// Executor and target are the same user
				interaction := createWarnInteractionWithResolvedUser(
					"user-123", "user-123", "guild-456", "channel-789",
					"some reason", false,
				)
				return command.NewContext(nil, interaction, warnTestLogger())
			},
			expectError: true,
			errContains: "yourself",
		},
		{
			name: "cannot warn bot",
			setupContext: func() *command.Context {
				interaction := createWarnInteractionWithResolvedUser(
					"moderator-123", "bot-456", "guild-789", "channel-012",
					"Being a bot", true, // target is a bot
				)
				return command.NewContext(nil, interaction, warnTestLogger())
			},
			expectError: true,
			errContains: "bot",
		},
		{
			name: "valid warn target",
			setupContext: func() *command.Context {
				interaction := createWarnInteractionWithResolvedUser(
					"moderator-123", "target-456", "guild-789", "channel-012",
					"Breaking rules", false,
				)
				return command.NewContext(nil, interaction, warnTestLogger())
			},
			// Will fail due to nil session, but should not fail validation
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &command.WarnCommand{}
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

func Test_WarnCommand_Execute_NilContext(t *testing.T) {
	cmd := &command.WarnCommand{}

	assert.NotPanics(t, func() {
		err := cmd.Execute(nil)
		assert.Error(t, err, "Execute should return error for nil context")
	}, "Execute should not panic with nil context")
}

func Test_WarnCommand_Execute_NilInteraction(t *testing.T) {
	cmd := &command.WarnCommand{}
	ctx := command.NewContext(nil, nil, warnTestLogger())

	assert.NotPanics(t, func() {
		_ = cmd.Execute(ctx)
	}, "Execute should not panic with nil interaction in context")
}

func Test_WarnCommand_Execute_CannotWarnSelf(t *testing.T) {
	cmd := &command.WarnCommand{}

	// Create interaction where executor and target are the same
	interaction := createWarnInteractionWithResolvedUser(
		"same-user-id", "same-user-id", "guild-123", "channel-456",
		"some reason", false,
	)
	ctx := command.NewContext(nil, interaction, warnTestLogger())

	err := cmd.Execute(ctx)

	require.Error(t, err, "Execute should return error when trying to warn self")
	assert.Contains(t, strings.ToLower(err.Error()), "yourself",
		"error message should indicate cannot warn yourself")
}

func Test_WarnCommand_Execute_CannotWarnBot(t *testing.T) {
	cmd := &command.WarnCommand{}

	// Create interaction where target is a bot
	interaction := createWarnInteractionWithResolvedUser(
		"moderator-123", "bot-user-456", "guild-123", "channel-456",
		"some reason", true, // target is a bot
	)
	ctx := command.NewContext(nil, interaction, warnTestLogger())

	err := cmd.Execute(ctx)

	require.Error(t, err, "Execute should return error when trying to warn a bot")
	assert.Contains(t, strings.ToLower(err.Error()), "bot",
		"error message should indicate cannot warn a bot")
}

func Test_WarnCommand_Execute_EmptyReason(t *testing.T) {
	cmd := &command.WarnCommand{}

	// Create interaction with empty reason
	interaction := createWarnInteractionWithResolvedUser(
		"moderator-123", "target-456", "guild-123", "channel-456",
		"", // empty reason
		false,
	)
	ctx := command.NewContext(nil, interaction, warnTestLogger())

	err := cmd.Execute(ctx)

	// Since reason is required, empty reason should return a validation error
	require.Error(t, err, "Execute should return error for empty reason")

	errLower := strings.ToLower(err.Error())
	containsExpected := strings.Contains(errLower, "reason") ||
		strings.Contains(errLower, "required") ||
		strings.Contains(errLower, "empty") ||
		strings.Contains(errLower, "validation")
	assert.True(t, containsExpected,
		"error should indicate reason is required or empty, got: %q", err.Error())
}

func Test_WarnCommand_Execute_ValidReasons(t *testing.T) {
	tests := []struct {
		name   string
		reason string
	}{
		{
			name:   "simple reason",
			reason: "Breaking rules",
		},
		{
			name:   "detailed reason",
			reason: "Repeatedly spamming in general chat despite multiple warnings",
		},
		{
			name:   "reason with special characters",
			reason: "Posting inappropriate content (NSFW) in #general",
		},
		{
			name:   "reason with mentions",
			reason: "Harassing <@123456789> in multiple channels",
		},
		{
			name:   "reason with unicode",
			reason: "Using offensive language in chat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &command.WarnCommand{}

			interaction := createWarnInteractionWithResolvedUser(
				"moderator-123", "target-456", "guild-789", "channel-012",
				tt.reason, false,
			)
			ctx := command.NewContext(nil, interaction, warnTestLogger())

			err := cmd.Execute(ctx)

			// Error is expected due to nil session, but should not be a validation error
			require.Error(t, err, "Execute should return error with nil session")

			// The error should not be about invalid reason (unless it's a session error)
			errLower := strings.ToLower(err.Error())
			isSessionError := strings.Contains(errLower, "session") ||
				strings.Contains(errLower, "nil") ||
				strings.Contains(errLower, "respond")
			isValidationError := strings.Contains(errLower, "reason") &&
				(strings.Contains(errLower, "required") || strings.Contains(errLower, "empty"))

			assert.True(t, isSessionError || !isValidationError,
				"valid reason %q should not produce validation error, got: %q",
				tt.reason, err.Error())
		})
	}
}

func Test_WarnCommand_ImplementsCommandInterface(t *testing.T) {
	// This test verifies that WarnCommand implements the Command interface
	// If this compiles, WarnCommand satisfies command.Command
	var _ command.Command = (*command.WarnCommand)(nil)
}

func Test_WarnCommand_ImplementsPermissionedCommandInterface(t *testing.T) {
	// This test verifies that WarnCommand implements the PermissionedCommand interface
	// If this compiles, WarnCommand satisfies command.PermissionedCommand
	var _ command.PermissionedCommand = (*command.WarnCommand)(nil)
}

func Test_WarnCommand_CanBeRegistered(t *testing.T) {
	// Test that WarnCommand can be registered in the Registry
	registry := command.NewRegistry(warnTestLogger())
	cmd := &command.WarnCommand{}

	err := registry.Register(cmd)

	require.NoError(t, err, "WarnCommand should be registerable")

	// Verify it can be retrieved
	retrieved, found := registry.Get("warn")
	assert.True(t, found, "warn command should be found in registry")
	assert.Equal(t, "warn", retrieved.Name())
}

func Test_WarnCommand_ApplicationCommand(t *testing.T) {
	// Test that WarnCommand converts correctly to ApplicationCommand
	registry := command.NewRegistry(warnTestLogger())
	cmd := &command.WarnCommand{}

	err := registry.Register(cmd)
	require.NoError(t, err)

	appCmds := registry.ApplicationCommands()

	require.Len(t, appCmds, 1)
	assert.Equal(t, "warn", appCmds[0].Name)
	assert.NotEmpty(t, appCmds[0].Description)
	require.NotEmpty(t, appCmds[0].Options, "warn command should have options")

	// Verify user option exists and is required
	var userOption *discordgo.ApplicationCommandOption
	for _, opt := range appCmds[0].Options {
		if opt.Name == "user" {
			userOption = opt
			break
		}
	}
	require.NotNil(t, userOption, "ApplicationCommand should have user option")
	assert.True(t, userOption.Required, "user option should be required in ApplicationCommand")

	// Verify reason option exists and is required
	var reasonOption *discordgo.ApplicationCommandOption
	for _, opt := range appCmds[0].Options {
		if opt.Name == "reason" {
			reasonOption = opt
			break
		}
	}
	require.NotNil(t, reasonOption, "ApplicationCommand should have reason option")
	assert.True(t, reasonOption.Required, "reason option should be required in ApplicationCommand")

	// Verify permissions are set
	require.NotNil(t, appCmds[0].DefaultMemberPermissions,
		"ApplicationCommand should have DefaultMemberPermissions set")
	assert.True(t, *appCmds[0].DefaultMemberPermissions&discordgo.PermissionModerateMembers != 0,
		"DefaultMemberPermissions should include ModerateMembers")
}

func Test_WarnCommand_MultipleWarns(t *testing.T) {
	// Test that multiple warn commands can be executed (tests for statelessness)
	cmd := &command.WarnCommand{}

	targets := []string{"target-1", "target-2", "target-3"}

	for _, targetID := range targets {
		t.Run("warn_"+targetID, func(t *testing.T) {
			interaction := createWarnInteractionWithResolvedUser(
				"moderator-123", targetID, "guild-789", "channel-012",
				"Breaking rules", false,
			)
			ctx := command.NewContext(nil, interaction, warnTestLogger())

			// Should not panic and should not have state leakage
			assert.NotPanics(t, func() {
				_ = cmd.Execute(ctx)
			})
		})
	}
}

// Benchmark tests
func Benchmark_WarnCommand_Name(b *testing.B) {
	cmd := &command.WarnCommand{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Name()
	}
}

func Benchmark_WarnCommand_Description(b *testing.B) {
	cmd := &command.WarnCommand{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Description()
	}
}

func Benchmark_WarnCommand_Options(b *testing.B) {
	cmd := &command.WarnCommand{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Options()
	}
}

func Benchmark_WarnCommand_Permissions(b *testing.B) {
	cmd := &command.WarnCommand{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Permissions()
	}
}
