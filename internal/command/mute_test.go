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

// muteTestLogger returns a zerolog.Logger that discards output for testing.
func muteTestLogger() zerolog.Logger {
	return zerolog.New(io.Discard).Level(zerolog.Disabled)
}

// createMuteTestInteraction creates a test interaction for mute command tests.
func createMuteTestInteraction(userID, guildID, channelID string, options []*discordgo.ApplicationCommandInteractionDataOption) *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:        "interaction-mute-test",
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
				ID:      "cmd-data-mute",
				Name:    "mute",
				Options: options,
			},
		},
	}
}

// createMuteOptions creates options for mute command testing.
func createMuteOptions(targetUserID string, duration string, reason string, includeReason bool) []*discordgo.ApplicationCommandInteractionDataOption {
	options := []*discordgo.ApplicationCommandInteractionDataOption{
		{
			Name:  "user",
			Type:  discordgo.ApplicationCommandOptionUser,
			Value: targetUserID,
		},
		{
			Name:  "duration",
			Type:  discordgo.ApplicationCommandOptionString,
			Value: duration,
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

// createMuteInteractionWithResolvedUser creates an interaction with resolved user data.
func createMuteInteractionWithResolvedUser(executorID, targetUserID, guildID, channelID string, duration string, reason string, includeReason bool, targetIsBot bool) *discordgo.InteractionCreate {
	interaction := createMuteTestInteraction(executorID, guildID, channelID, createMuteOptions(targetUserID, duration, reason, includeReason))

	// Add resolved user data
	interaction.Interaction.Data = discordgo.ApplicationCommandInteractionData{
		ID:      "cmd-data-mute",
		Name:    "mute",
		Options: createMuteOptions(targetUserID, duration, reason, includeReason),
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

func Test_MuteCommand_Name(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "returns mute as command name",
			expected: "mute",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &command.MuteCommand{}

			result := cmd.Name()

			assert.Equal(t, tt.expected, result, "Name() should return %q", tt.expected)
		})
	}
}

func Test_MuteCommand_Description(t *testing.T) {
	tests := []struct {
		name        string
		containsAny []string
		notEmpty    bool
	}{
		{
			name:        "returns non-empty description with relevant keywords",
			containsAny: []string{"mute", "Mute", "timeout", "user", "member", "silence", "temporarily"},
			notEmpty:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &command.MuteCommand{}

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

func Test_MuteCommand_Options(t *testing.T) {
	cmd := &command.MuteCommand{}
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

	t.Run("has duration option", func(t *testing.T) {
		var durationOption *discordgo.ApplicationCommandOption
		for _, opt := range options {
			if opt.Name == "duration" {
				durationOption = opt
				break
			}
		}

		require.NotNil(t, durationOption, "Options should contain 'duration' option")
		assert.Equal(t, discordgo.ApplicationCommandOptionString, durationOption.Type,
			"duration option should be of type String")
		assert.True(t, durationOption.Required, "duration option should be required")
		assert.NotEmpty(t, durationOption.Description, "duration option should have a description")
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
			// Reason might be required or optional depending on implementation
		}
	})
}

func Test_MuteCommand_Permissions(t *testing.T) {
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
			cmd := &command.MuteCommand{}

			result := cmd.Permissions()

			// Check that the required permission is included in the returned permissions
			assert.True(t, result&tt.expectedPermission != 0,
				"Permissions() should include PermissionModerateMembers (0x%X), got 0x%X",
				tt.expectedPermission, result)
		})
	}
}

func Test_MuteCommand_Execute(t *testing.T) {
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
			name: "cannot mute self",
			setupContext: func() *command.Context {
				// Executor and target are the same user
				interaction := createMuteInteractionWithResolvedUser(
					"user-123", "user-123", "guild-456", "channel-789",
					"1h", "no reason", true, false,
				)
				return command.NewContext(nil, interaction, muteTestLogger())
			},
			expectError: true,
			errContains: "yourself",
		},
		{
			name: "valid mute target with duration",
			setupContext: func() *command.Context {
				interaction := createMuteInteractionWithResolvedUser(
					"moderator-123", "target-456", "guild-789", "channel-012",
					"1h", "Being disruptive", true, false,
				)
				return command.NewContext(nil, interaction, muteTestLogger())
			},
			// Will fail due to nil session, but should not fail validation
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &command.MuteCommand{}
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

func Test_MuteCommand_Execute_NilContext(t *testing.T) {
	cmd := &command.MuteCommand{}

	assert.NotPanics(t, func() {
		err := cmd.Execute(nil)
		assert.Error(t, err, "Execute should return error for nil context")
	}, "Execute should not panic with nil context")
}

func Test_MuteCommand_Execute_NilInteraction(t *testing.T) {
	cmd := &command.MuteCommand{}
	ctx := command.NewContext(nil, nil, muteTestLogger())

	assert.NotPanics(t, func() {
		_ = cmd.Execute(ctx)
	}, "Execute should not panic with nil interaction in context")
}

func Test_MuteCommand_Execute_CannotMuteSelf(t *testing.T) {
	cmd := &command.MuteCommand{}

	// Create interaction where executor and target are the same
	interaction := createMuteInteractionWithResolvedUser(
		"same-user-id", "same-user-id", "guild-123", "channel-456",
		"1h", "some reason", true, false,
	)
	ctx := command.NewContext(nil, interaction, muteTestLogger())

	err := cmd.Execute(ctx)

	require.Error(t, err, "Execute should return error when trying to mute self")
	assert.Contains(t, strings.ToLower(err.Error()), "yourself",
		"error message should indicate cannot mute yourself")
}

func Test_MuteCommand_Execute_InvalidDuration(t *testing.T) {
	tests := []struct {
		name        string
		duration    string
		errContains []string // any of these should be in the error
	}{
		{
			name:        "invalid duration format abc",
			duration:    "abc",
			errContains: []string{"invalid", "duration", "parse", "format"},
		},
		{
			name:        "invalid duration format with wrong unit",
			duration:    "1x",
			errContains: []string{"invalid", "duration", "parse", "format", "unit"},
		},
		{
			name:        "empty duration",
			duration:    "",
			errContains: []string{"invalid", "duration", "required", "empty"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &command.MuteCommand{}

			interaction := createMuteInteractionWithResolvedUser(
				"moderator-123", "target-456", "guild-789", "channel-012",
				tt.duration, "reason", true, false,
			)
			ctx := command.NewContext(nil, interaction, muteTestLogger())

			err := cmd.Execute(ctx)

			require.Error(t, err, "Execute should return error for invalid duration")

			// Check if error contains any of the expected strings
			errLower := strings.ToLower(err.Error())
			containsExpected := false
			for _, expected := range tt.errContains {
				if strings.Contains(errLower, strings.ToLower(expected)) {
					containsExpected = true
					break
				}
			}
			assert.True(t, containsExpected,
				"error should contain one of %v, got: %q", tt.errContains, err.Error())
		})
	}
}

func Test_MuteCommand_Execute_DurationTooShort(t *testing.T) {
	cmd := &command.MuteCommand{}

	// 30 seconds is too short for a timeout (minimum is typically 1 minute)
	interaction := createMuteInteractionWithResolvedUser(
		"moderator-123", "target-456", "guild-789", "channel-012",
		"30s", "reason", true, false,
	)
	ctx := command.NewContext(nil, interaction, muteTestLogger())

	err := cmd.Execute(ctx)

	require.Error(t, err, "Execute should return error for duration too short")

	errLower := strings.ToLower(err.Error())
	containsExpected := strings.Contains(errLower, "minimum") ||
		strings.Contains(errLower, "short") ||
		strings.Contains(errLower, "at least") ||
		strings.Contains(errLower, "too") ||
		strings.Contains(errLower, "invalid")
	assert.True(t, containsExpected,
		"error should indicate duration is too short, got: %q", err.Error())
}

func Test_MuteCommand_Execute_DurationTooLong(t *testing.T) {
	cmd := &command.MuteCommand{}

	// 30 days exceeds Discord's maximum timeout of 28 days
	interaction := createMuteInteractionWithResolvedUser(
		"moderator-123", "target-456", "guild-789", "channel-012",
		"30d", "reason", true, false,
	)
	ctx := command.NewContext(nil, interaction, muteTestLogger())

	err := cmd.Execute(ctx)

	require.Error(t, err, "Execute should return error for duration too long")

	errLower := strings.ToLower(err.Error())
	containsExpected := strings.Contains(errLower, "maximum") ||
		strings.Contains(errLower, "28 day") ||
		strings.Contains(errLower, "28day") ||
		strings.Contains(errLower, "long") ||
		strings.Contains(errLower, "exceed") ||
		strings.Contains(errLower, "too") ||
		strings.Contains(errLower, "invalid")
	assert.True(t, containsExpected,
		"error should indicate duration exceeds maximum (28 days), got: %q", err.Error())
}

func Test_MuteCommand_Execute_ValidDurations(t *testing.T) {
	tests := []struct {
		name     string
		duration string
	}{
		{
			name:     "1 minute",
			duration: "1m",
		},
		{
			name:     "5 minutes",
			duration: "5m",
		},
		{
			name:     "1 hour",
			duration: "1h",
		},
		{
			name:     "24 hours",
			duration: "24h",
		},
		{
			name:     "1 day",
			duration: "1d",
		},
		{
			name:     "7 days",
			duration: "7d",
		},
		{
			name:     "28 days (maximum)",
			duration: "28d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &command.MuteCommand{}

			interaction := createMuteInteractionWithResolvedUser(
				"moderator-123", "target-456", "guild-789", "channel-012",
				tt.duration, "reason", true, false,
			)
			ctx := command.NewContext(nil, interaction, muteTestLogger())

			err := cmd.Execute(ctx)

			// Error is expected due to nil session, but should not be a validation error
			require.Error(t, err, "Execute should return error with nil session")

			// The error should not be about invalid duration
			errLower := strings.ToLower(err.Error())
			assert.False(t, strings.Contains(errLower, "duration"),
				"valid duration %q should not produce duration-related error, got: %q",
				tt.duration, err.Error())
		})
	}
}

func Test_MuteCommand_ImplementsCommandInterface(t *testing.T) {
	// This test verifies that MuteCommand implements the Command interface
	// If this compiles, MuteCommand satisfies command.Command
	var _ command.Command = (*command.MuteCommand)(nil)
}

func Test_MuteCommand_ImplementsPermissionedCommandInterface(t *testing.T) {
	// This test verifies that MuteCommand implements the PermissionedCommand interface
	// If this compiles, MuteCommand satisfies command.PermissionedCommand
	var _ command.PermissionedCommand = (*command.MuteCommand)(nil)
}

func Test_MuteCommand_CanBeRegistered(t *testing.T) {
	// Test that MuteCommand can be registered in the Registry
	registry := command.NewRegistry(muteTestLogger())
	cmd := &command.MuteCommand{}

	err := registry.Register(cmd)

	require.NoError(t, err, "MuteCommand should be registerable")

	// Verify it can be retrieved
	retrieved, found := registry.Get("mute")
	assert.True(t, found, "mute command should be found in registry")
	assert.Equal(t, "mute", retrieved.Name())
}

func Test_MuteCommand_ApplicationCommand(t *testing.T) {
	// Test that MuteCommand converts correctly to ApplicationCommand
	registry := command.NewRegistry(muteTestLogger())
	cmd := &command.MuteCommand{}

	err := registry.Register(cmd)
	require.NoError(t, err)

	appCmds := registry.ApplicationCommands()

	require.Len(t, appCmds, 1)
	assert.Equal(t, "mute", appCmds[0].Name)
	assert.NotEmpty(t, appCmds[0].Description)
	require.NotEmpty(t, appCmds[0].Options, "mute command should have options")

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

	// Verify duration option exists and is required
	var durationOption *discordgo.ApplicationCommandOption
	for _, opt := range appCmds[0].Options {
		if opt.Name == "duration" {
			durationOption = opt
			break
		}
	}
	require.NotNil(t, durationOption, "ApplicationCommand should have duration option")
	assert.True(t, durationOption.Required, "duration option should be required in ApplicationCommand")

	// Verify permissions are set
	require.NotNil(t, appCmds[0].DefaultMemberPermissions,
		"ApplicationCommand should have DefaultMemberPermissions set")
	assert.True(t, *appCmds[0].DefaultMemberPermissions&discordgo.PermissionModerateMembers != 0,
		"DefaultMemberPermissions should include ModerateMembers")
}

// Benchmark tests
func Benchmark_MuteCommand_Name(b *testing.B) {
	cmd := &command.MuteCommand{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Name()
	}
}

func Benchmark_MuteCommand_Description(b *testing.B) {
	cmd := &command.MuteCommand{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Description()
	}
}

func Benchmark_MuteCommand_Options(b *testing.B) {
	cmd := &command.MuteCommand{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Options()
	}
}

func Benchmark_MuteCommand_Permissions(b *testing.B) {
	cmd := &command.MuteCommand{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Permissions()
	}
}
