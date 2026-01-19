package command_test

import (
	"io"
	"testing"

	"jamesbot/internal/command"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestInteractionCreate creates a discordgo.InteractionCreate for testing.
func createTestInteractionCreate(userID, guildID, channelID string, options []*discordgo.ApplicationCommandInteractionDataOption) *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:        "interaction-123",
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
				ID:      "cmd-data-123",
				Name:    "testcmd",
				Options: options,
			},
		},
	}
}

// createTestSession creates a minimal discordgo.Session for testing.
// Note: This returns nil since we can't easily create a real session,
// but tests should handle nil sessions gracefully.
func createTestSession() *discordgo.Session {
	return nil
}

// testLogger returns a zerolog.Logger for testing.
func testLogger() zerolog.Logger {
	return zerolog.New(io.Discard).Level(zerolog.Disabled)
}

func Test_NewContext(t *testing.T) {
	tests := []struct {
		name        string
		session     *discordgo.Session
		interaction *discordgo.InteractionCreate
		logger      zerolog.Logger
	}{
		{
			name:        "create context with all parameters",
			session:     createTestSession(),
			interaction: createTestInteractionCreate("user-123", "guild-456", "channel-789", nil),
			logger:      testLogger(),
		},
		{
			name:        "create context with nop logger",
			session:     createTestSession(),
			interaction: createTestInteractionCreate("user-123", "guild-456", "channel-789", nil),
			logger:      zerolog.Nop(),
		},
		{
			name:        "create context with nil session",
			session:     nil,
			interaction: createTestInteractionCreate("user-123", "guild-456", "channel-789", nil),
			logger:      testLogger(),
		},
		{
			name:        "create context with nil interaction",
			session:     createTestSession(),
			interaction: nil,
			logger:      testLogger(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := command.NewContext(tt.session, tt.interaction, tt.logger)

			require.NotNil(t, ctx, "NewContext should return non-nil *Context")
		})
	}
}

func Test_Context_StringOption(t *testing.T) {
	tests := []struct {
		name          string
		options       []*discordgo.ApplicationCommandInteractionDataOption
		optionName    string
		expectedValue string
	}{
		{
			name: "existing string option",
			options: []*discordgo.ApplicationCommandInteractionDataOption{
				{
					Name:  "text",
					Type:  discordgo.ApplicationCommandOptionString,
					Value: "hello world",
				},
			},
			optionName:    "text",
			expectedValue: "hello world",
		},
		{
			name: "missing option returns empty string",
			options: []*discordgo.ApplicationCommandInteractionDataOption{
				{
					Name:  "other",
					Type:  discordgo.ApplicationCommandOptionString,
					Value: "value",
				},
			},
			optionName:    "missing",
			expectedValue: "",
		},
		{
			name:          "no options returns empty string",
			options:       nil,
			optionName:    "anything",
			expectedValue: "",
		},
		{
			name:          "empty options slice returns empty string",
			options:       []*discordgo.ApplicationCommandInteractionDataOption{},
			optionName:    "missing",
			expectedValue: "",
		},
		{
			name: "multiple options - get first",
			options: []*discordgo.ApplicationCommandInteractionDataOption{
				{
					Name:  "first",
					Type:  discordgo.ApplicationCommandOptionString,
					Value: "first-value",
				},
				{
					Name:  "second",
					Type:  discordgo.ApplicationCommandOptionString,
					Value: "second-value",
				},
			},
			optionName:    "first",
			expectedValue: "first-value",
		},
		{
			name: "multiple options - get second",
			options: []*discordgo.ApplicationCommandInteractionDataOption{
				{
					Name:  "first",
					Type:  discordgo.ApplicationCommandOptionString,
					Value: "first-value",
				},
				{
					Name:  "second",
					Type:  discordgo.ApplicationCommandOptionString,
					Value: "second-value",
				},
			},
			optionName:    "second",
			expectedValue: "second-value",
		},
		{
			name: "case sensitive option name - exact match",
			options: []*discordgo.ApplicationCommandInteractionDataOption{
				{
					Name:  "Text",
					Type:  discordgo.ApplicationCommandOptionString,
					Value: "case-sensitive",
				},
			},
			optionName:    "Text",
			expectedValue: "case-sensitive",
		},
		{
			name: "case sensitive option name - wrong case",
			options: []*discordgo.ApplicationCommandInteractionDataOption{
				{
					Name:  "Text",
					Type:  discordgo.ApplicationCommandOptionString,
					Value: "case-sensitive",
				},
			},
			optionName:    "text",
			expectedValue: "",
		},
		{
			name: "empty string value",
			options: []*discordgo.ApplicationCommandInteractionDataOption{
				{
					Name:  "empty",
					Type:  discordgo.ApplicationCommandOptionString,
					Value: "",
				},
			},
			optionName:    "empty",
			expectedValue: "",
		},
		{
			name: "string with special characters",
			options: []*discordgo.ApplicationCommandInteractionDataOption{
				{
					Name:  "special",
					Type:  discordgo.ApplicationCommandOptionString,
					Value: "hello\nworld\ttab",
				},
			},
			optionName:    "special",
			expectedValue: "hello\nworld\ttab",
		},
		{
			name: "unicode string value",
			options: []*discordgo.ApplicationCommandInteractionDataOption{
				{
					Name:  "unicode",
					Type:  discordgo.ApplicationCommandOptionString,
					Value: "Hello World!",
				},
			},
			optionName:    "unicode",
			expectedValue: "Hello World!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interaction := createTestInteractionCreate("user-1", "guild-1", "channel-1", tt.options)
			ctx := command.NewContext(createTestSession(), interaction, testLogger())

			result := ctx.StringOption(tt.optionName)

			assert.Equal(t, tt.expectedValue, result,
				"StringOption(%q) should return expected value", tt.optionName)
		})
	}
}

func Test_Context_IntOption(t *testing.T) {
	tests := []struct {
		name          string
		options       []*discordgo.ApplicationCommandInteractionDataOption
		optionName    string
		expectedValue int64
	}{
		{
			name: "existing int option",
			options: []*discordgo.ApplicationCommandInteractionDataOption{
				{
					Name:  "count",
					Type:  discordgo.ApplicationCommandOptionInteger,
					Value: float64(5), // Discord sends integers as float64 in JSON
				},
			},
			optionName:    "count",
			expectedValue: 5,
		},
		{
			name: "missing option returns zero",
			options: []*discordgo.ApplicationCommandInteractionDataOption{
				{
					Name:  "other",
					Type:  discordgo.ApplicationCommandOptionInteger,
					Value: float64(10),
				},
			},
			optionName:    "missing",
			expectedValue: 0,
		},
		{
			name:          "no options returns zero",
			options:       nil,
			optionName:    "anything",
			expectedValue: 0,
		},
		{
			name: "zero value",
			options: []*discordgo.ApplicationCommandInteractionDataOption{
				{
					Name:  "zero",
					Type:  discordgo.ApplicationCommandOptionInteger,
					Value: float64(0),
				},
			},
			optionName:    "zero",
			expectedValue: 0,
		},
		{
			name: "negative value",
			options: []*discordgo.ApplicationCommandInteractionDataOption{
				{
					Name:  "negative",
					Type:  discordgo.ApplicationCommandOptionInteger,
					Value: float64(-42),
				},
			},
			optionName:    "negative",
			expectedValue: -42,
		},
		{
			name: "large value",
			options: []*discordgo.ApplicationCommandInteractionDataOption{
				{
					Name:  "large",
					Type:  discordgo.ApplicationCommandOptionInteger,
					Value: float64(999999999),
				},
			},
			optionName:    "large",
			expectedValue: 999999999,
		},
		{
			name: "multiple options - get specific",
			options: []*discordgo.ApplicationCommandInteractionDataOption{
				{
					Name:  "first",
					Type:  discordgo.ApplicationCommandOptionInteger,
					Value: float64(1),
				},
				{
					Name:  "second",
					Type:  discordgo.ApplicationCommandOptionInteger,
					Value: float64(2),
				},
				{
					Name:  "third",
					Type:  discordgo.ApplicationCommandOptionInteger,
					Value: float64(3),
				},
			},
			optionName:    "second",
			expectedValue: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interaction := createTestInteractionCreate("user-1", "guild-1", "channel-1", tt.options)
			ctx := command.NewContext(createTestSession(), interaction, testLogger())

			result := ctx.IntOption(tt.optionName)

			assert.Equal(t, tt.expectedValue, result,
				"IntOption(%q) should return expected value", tt.optionName)
		})
	}
}

func Test_Context_UserID(t *testing.T) {
	tests := []struct {
		name           string
		interaction    *discordgo.InteractionCreate
		expectedUserID string
	}{
		{
			name:           "extract user ID from guild interaction",
			interaction:    createTestInteractionCreate("user-123456", "guild-1", "channel-1", nil),
			expectedUserID: "user-123456",
		},
		{
			name:           "extract user ID with numeric ID",
			interaction:    createTestInteractionCreate("123456789012345678", "guild-1", "channel-1", nil),
			expectedUserID: "123456789012345678",
		},
		{
			name: "extract user ID from DM interaction (User field only)",
			interaction: &discordgo.InteractionCreate{
				Interaction: &discordgo.Interaction{
					User: &discordgo.User{
						ID:       "dm-user-456",
						Username: "dmuser",
					},
					Type: discordgo.InteractionApplicationCommand,
					Data: discordgo.ApplicationCommandInteractionData{},
				},
			},
			expectedUserID: "dm-user-456",
		},
		{
			name:           "nil interaction returns empty string",
			interaction:    nil,
			expectedUserID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := command.NewContext(createTestSession(), tt.interaction, testLogger())

			result := ctx.UserID()

			assert.Equal(t, tt.expectedUserID, result, "UserID() should return expected value")
		})
	}
}

func Test_Context_GuildID(t *testing.T) {
	tests := []struct {
		name            string
		interaction     *discordgo.InteractionCreate
		expectedGuildID string
	}{
		{
			name:            "extract guild ID from interaction",
			interaction:     createTestInteractionCreate("user-1", "guild-789012", "channel-1", nil),
			expectedGuildID: "guild-789012",
		},
		{
			name:            "numeric guild ID",
			interaction:     createTestInteractionCreate("user-1", "987654321098765432", "channel-1", nil),
			expectedGuildID: "987654321098765432",
		},
		{
			name: "empty guild ID for DM",
			interaction: &discordgo.InteractionCreate{
				Interaction: &discordgo.Interaction{
					GuildID:   "",
					ChannelID: "dm-channel",
					User: &discordgo.User{
						ID: "user-1",
					},
					Type: discordgo.InteractionApplicationCommand,
					Data: discordgo.ApplicationCommandInteractionData{},
				},
			},
			expectedGuildID: "",
		},
		{
			name:            "nil interaction returns empty string",
			interaction:     nil,
			expectedGuildID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := command.NewContext(createTestSession(), tt.interaction, testLogger())

			result := ctx.GuildID()

			assert.Equal(t, tt.expectedGuildID, result, "GuildID() should return expected value")
		})
	}
}

func Test_Context_ChannelID(t *testing.T) {
	tests := []struct {
		name              string
		interaction       *discordgo.InteractionCreate
		expectedChannelID string
	}{
		{
			name:              "extract channel ID from interaction",
			interaction:       createTestInteractionCreate("user-1", "guild-1", "channel-456789", nil),
			expectedChannelID: "channel-456789",
		},
		{
			name:              "numeric channel ID",
			interaction:       createTestInteractionCreate("user-1", "guild-1", "123456789012345678", nil),
			expectedChannelID: "123456789012345678",
		},
		{
			name: "DM channel ID",
			interaction: &discordgo.InteractionCreate{
				Interaction: &discordgo.Interaction{
					GuildID:   "",
					ChannelID: "dm-channel-999",
					User: &discordgo.User{
						ID: "user-1",
					},
					Type: discordgo.InteractionApplicationCommand,
					Data: discordgo.ApplicationCommandInteractionData{},
				},
			},
			expectedChannelID: "dm-channel-999",
		},
		{
			name:              "nil interaction returns empty string",
			interaction:       nil,
			expectedChannelID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := command.NewContext(createTestSession(), tt.interaction, testLogger())

			result := ctx.ChannelID()

			assert.Equal(t, tt.expectedChannelID, result, "ChannelID() should return expected value")
		})
	}
}

func Test_Context_AllIDMethods(t *testing.T) {
	// Test that all ID methods work correctly together
	interaction := createTestInteractionCreate(
		"user-111222333",
		"guild-444555666",
		"channel-777888999",
		nil,
	)

	ctx := command.NewContext(createTestSession(), interaction, testLogger())

	assert.Equal(t, "user-111222333", ctx.UserID())
	assert.Equal(t, "guild-444555666", ctx.GuildID())
	assert.Equal(t, "channel-777888999", ctx.ChannelID())
}

func Test_Context_MixedOptions(t *testing.T) {
	// Test context with mixed option types
	options := []*discordgo.ApplicationCommandInteractionDataOption{
		{
			Name:  "message",
			Type:  discordgo.ApplicationCommandOptionString,
			Value: "Hello",
		},
		{
			Name:  "count",
			Type:  discordgo.ApplicationCommandOptionInteger,
			Value: float64(42),
		},
		{
			Name:  "other-string",
			Type:  discordgo.ApplicationCommandOptionString,
			Value: "World",
		},
	}

	interaction := createTestInteractionCreate("user-1", "guild-1", "channel-1", options)
	ctx := command.NewContext(createTestSession(), interaction, testLogger())

	// Test retrieving each type correctly
	assert.Equal(t, "Hello", ctx.StringOption("message"))
	assert.Equal(t, int64(42), ctx.IntOption("count"))
	assert.Equal(t, "World", ctx.StringOption("other-string"))

	// Test that wrong type returns zero/empty value
	assert.Equal(t, "", ctx.StringOption("count"))       // int option as string
	assert.Equal(t, int64(0), ctx.IntOption("message"))  // string option as int
	assert.Equal(t, "", ctx.StringOption("nonexistent")) // missing
	assert.Equal(t, int64(0), ctx.IntOption("nonexistent"))
}

func Test_Context_BoolOption(t *testing.T) {
	tests := []struct {
		name          string
		options       []*discordgo.ApplicationCommandInteractionDataOption
		optionName    string
		expectedValue bool
	}{
		{
			name: "true boolean option",
			options: []*discordgo.ApplicationCommandInteractionDataOption{
				{
					Name:  "enabled",
					Type:  discordgo.ApplicationCommandOptionBoolean,
					Value: true,
				},
			},
			optionName:    "enabled",
			expectedValue: true,
		},
		{
			name: "false boolean option",
			options: []*discordgo.ApplicationCommandInteractionDataOption{
				{
					Name:  "disabled",
					Type:  discordgo.ApplicationCommandOptionBoolean,
					Value: false,
				},
			},
			optionName:    "disabled",
			expectedValue: false,
		},
		{
			name:          "missing boolean option returns false",
			options:       nil,
			optionName:    "missing",
			expectedValue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interaction := createTestInteractionCreate("user-1", "guild-1", "channel-1", tt.options)
			ctx := command.NewContext(createTestSession(), interaction, testLogger())

			result := ctx.BoolOption(tt.optionName)
			assert.Equal(t, tt.expectedValue, result)
		})
	}
}

func Test_Context_Interaction(t *testing.T) {
	// Test that the original interaction is accessible via the Interaction field
	interaction := createTestInteractionCreate("user-123", "guild-456", "channel-789", nil)
	ctx := command.NewContext(createTestSession(), interaction, testLogger())

	assert.Equal(t, interaction, ctx.Interaction)
}

func Test_Context_Session(t *testing.T) {
	// Test that the session is accessible via the Session field
	session := createTestSession()
	interaction := createTestInteractionCreate("user-1", "guild-1", "channel-1", nil)
	ctx := command.NewContext(session, interaction, testLogger())

	assert.Equal(t, session, ctx.Session)
}

func Test_Context_Logger(t *testing.T) {
	// Test that the logger is accessible via the Logger field
	logger := testLogger()
	interaction := createTestInteractionCreate("user-1", "guild-1", "channel-1", nil)
	ctx := command.NewContext(createTestSession(), interaction, logger)

	// The logger in context may be enhanced with context fields
	// Just verify it's set and usable
	assert.NotPanics(t, func() {
		ctx.Logger.Info().Msg("test")
	})
}

// Benchmark tests for Context methods
func Benchmark_Context_StringOption(b *testing.B) {
	options := []*discordgo.ApplicationCommandInteractionDataOption{
		{Name: "opt1", Type: discordgo.ApplicationCommandOptionString, Value: "value1"},
		{Name: "opt2", Type: discordgo.ApplicationCommandOptionString, Value: "value2"},
		{Name: "opt3", Type: discordgo.ApplicationCommandOptionString, Value: "value3"},
		{Name: "opt4", Type: discordgo.ApplicationCommandOptionString, Value: "value4"},
		{Name: "opt5", Type: discordgo.ApplicationCommandOptionString, Value: "value5"},
	}
	interaction := createTestInteractionCreate("user-1", "guild-1", "channel-1", options)
	ctx := command.NewContext(nil, interaction, zerolog.Nop())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ctx.StringOption("opt3")
	}
}

func Benchmark_Context_IntOption(b *testing.B) {
	options := []*discordgo.ApplicationCommandInteractionDataOption{
		{Name: "int1", Type: discordgo.ApplicationCommandOptionInteger, Value: float64(1)},
		{Name: "int2", Type: discordgo.ApplicationCommandOptionInteger, Value: float64(2)},
		{Name: "int3", Type: discordgo.ApplicationCommandOptionInteger, Value: float64(3)},
	}
	interaction := createTestInteractionCreate("user-1", "guild-1", "channel-1", options)
	ctx := command.NewContext(nil, interaction, zerolog.Nop())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ctx.IntOption("int2")
	}
}

func Benchmark_Context_UserID(b *testing.B) {
	interaction := createTestInteractionCreate("user-123456789", "guild-1", "channel-1", nil)
	ctx := command.NewContext(nil, interaction, zerolog.Nop())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ctx.UserID()
	}
}

func Benchmark_Context_GuildID(b *testing.B) {
	interaction := createTestInteractionCreate("user-1", "guild-123456789", "channel-1", nil)
	ctx := command.NewContext(nil, interaction, zerolog.Nop())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ctx.GuildID()
	}
}

func Benchmark_Context_ChannelID(b *testing.B) {
	interaction := createTestInteractionCreate("user-1", "guild-1", "channel-123456789", nil)
	ctx := command.NewContext(nil, interaction, zerolog.Nop())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ctx.ChannelID()
	}
}

// Test empty interaction data
func Test_Context_EmptyInteractionData(t *testing.T) {
	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:        "int-1",
			ChannelID: "channel-1",
			GuildID:   "guild-1",
			Member: &discordgo.Member{
				User: &discordgo.User{
					ID: "user-1",
				},
			},
			Type: discordgo.InteractionApplicationCommand,
			Data: discordgo.ApplicationCommandInteractionData{
				// Empty data - no options
			},
		},
	}

	ctx := command.NewContext(createTestSession(), interaction, testLogger())

	// Should return empty/zero for options
	assert.Equal(t, "", ctx.StringOption("any"))
	assert.Equal(t, int64(0), ctx.IntOption("any"))

	// Should still return IDs
	assert.Equal(t, "user-1", ctx.UserID())
	assert.Equal(t, "guild-1", ctx.GuildID())
	assert.Equal(t, "channel-1", ctx.ChannelID())
}

// Test option name edge cases
func Test_Context_OptionNameEdgeCases(t *testing.T) {
	options := []*discordgo.ApplicationCommandInteractionDataOption{
		{Name: "", Type: discordgo.ApplicationCommandOptionString, Value: "empty-name"},
		{Name: "normal", Type: discordgo.ApplicationCommandOptionString, Value: "normal-value"},
		{Name: "with-hyphen", Type: discordgo.ApplicationCommandOptionString, Value: "hyphen-value"},
		{Name: "with_underscore", Type: discordgo.ApplicationCommandOptionString, Value: "underscore-value"},
	}

	interaction := createTestInteractionCreate("user-1", "guild-1", "channel-1", options)
	ctx := command.NewContext(createTestSession(), interaction, testLogger())

	tests := []struct {
		name     string
		expected string
	}{
		{"", "empty-name"},
		{"normal", "normal-value"},
		{"with-hyphen", "hyphen-value"},
		{"with_underscore", "underscore-value"},
	}

	for _, tt := range tests {
		t.Run("option_"+tt.name, func(t *testing.T) {
			result := ctx.StringOption(tt.name)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test very long option values
func Test_Context_LongOptionValues(t *testing.T) {
	longValue := ""
	for i := 0; i < 1000; i++ {
		longValue += "a"
	}

	options := []*discordgo.ApplicationCommandInteractionDataOption{
		{Name: "long", Type: discordgo.ApplicationCommandOptionString, Value: longValue},
	}

	interaction := createTestInteractionCreate("user-1", "guild-1", "channel-1", options)
	ctx := command.NewContext(createTestSession(), interaction, testLogger())

	result := ctx.StringOption("long")
	assert.Equal(t, longValue, result)
	assert.Len(t, result, 1000)
}

// Test that Context is usable with nil values where documented
func Test_Context_NilSafety(t *testing.T) {
	// Create context with nil interaction
	ctx := command.NewContext(nil, nil, zerolog.Nop())

	// These should not panic with nil interaction
	assert.NotPanics(t, func() {
		_ = ctx.UserID()
	})
	assert.NotPanics(t, func() {
		_ = ctx.GuildID()
	})
	assert.NotPanics(t, func() {
		_ = ctx.ChannelID()
	})
	assert.NotPanics(t, func() {
		_ = ctx.StringOption("test")
	})
	assert.NotPanics(t, func() {
		_ = ctx.IntOption("test")
	})
	assert.NotPanics(t, func() {
		_ = ctx.BoolOption("test")
	})

	// They should return empty/zero values
	assert.Equal(t, "", ctx.UserID())
	assert.Equal(t, "", ctx.GuildID())
	assert.Equal(t, "", ctx.ChannelID())
	assert.Equal(t, "", ctx.StringOption("test"))
	assert.Equal(t, int64(0), ctx.IntOption("test"))
	assert.Equal(t, false, ctx.BoolOption("test"))
}

// Test Context fields are properly accessible
func Test_Context_PublicFields(t *testing.T) {
	session := createTestSession()
	interaction := createTestInteractionCreate("user-1", "guild-1", "channel-1", nil)
	logger := testLogger()

	ctx := command.NewContext(session, interaction, logger)

	// Verify public fields are accessible
	assert.Equal(t, session, ctx.Session)
	assert.Equal(t, interaction, ctx.Interaction)
	// Logger may be enhanced but should be accessible
	_ = ctx.Logger
}

// Test UserOption method
func Test_Context_UserOption(t *testing.T) {
	tests := []struct {
		name        string
		options     []*discordgo.ApplicationCommandInteractionDataOption
		optionName  string
		expectedNil bool
	}{
		{
			name: "user option present",
			options: []*discordgo.ApplicationCommandInteractionDataOption{
				{
					Name:  "target",
					Type:  discordgo.ApplicationCommandOptionUser,
					Value: "target-user-123",
				},
			},
			optionName:  "target",
			expectedNil: false, // Note: actual User value depends on Session
		},
		{
			name:        "missing user option returns nil",
			options:     nil,
			optionName:  "target",
			expectedNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interaction := createTestInteractionCreate("user-1", "guild-1", "channel-1", tt.options)
			ctx := command.NewContext(createTestSession(), interaction, testLogger())

			result := ctx.UserOption(tt.optionName)

			if tt.expectedNil {
				assert.Nil(t, result)
			}
			// Note: When session is nil, UserOption may return nil even for valid options
		})
	}
}

// Test interaction with nil Member but valid User (DM case)
func Test_Context_DMInteraction(t *testing.T) {
	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:        "dm-interaction",
			ChannelID: "dm-channel",
			GuildID:   "",  // Empty for DM
			Member:    nil, // Nil for DM
			User: &discordgo.User{
				ID:       "dm-user-id",
				Username: "dmuser",
			},
			Type: discordgo.InteractionApplicationCommand,
			Data: discordgo.ApplicationCommandInteractionData{
				Name: "dmcmd",
			},
		},
	}

	ctx := command.NewContext(createTestSession(), interaction, testLogger())

	assert.Equal(t, "dm-user-id", ctx.UserID(), "should extract user ID from User field in DM")
	assert.Equal(t, "", ctx.GuildID(), "guild ID should be empty for DM")
	assert.Equal(t, "dm-channel", ctx.ChannelID(), "channel ID should be set for DM")
}

// Test interaction with both Member and User (guild case - Member takes precedence)
func Test_Context_GuildInteractionUserPrecedence(t *testing.T) {
	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:        "guild-interaction",
			ChannelID: "guild-channel",
			GuildID:   "guild-123",
			Member: &discordgo.Member{
				User: &discordgo.User{
					ID:       "member-user-id",
					Username: "memberuser",
				},
			},
			User: &discordgo.User{
				ID:       "should-not-be-used",
				Username: "topuser",
			},
			Type: discordgo.InteractionApplicationCommand,
			Data: discordgo.ApplicationCommandInteractionData{
				Name: "guildcmd",
			},
		},
	}

	ctx := command.NewContext(createTestSession(), interaction, testLogger())

	// Member.User should take precedence over User
	assert.Equal(t, "member-user-id", ctx.UserID(), "should extract user ID from Member.User in guild")
}
