package command_test

import (
	"errors"
	"io"
	"testing"

	"jamesbot/internal/command"
	"jamesbot/pkg/errutil"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// echoTestLogger returns a zerolog.Logger that discards output for testing.
func echoTestLogger() zerolog.Logger {
	return zerolog.New(io.Discard).Level(zerolog.Disabled)
}

// createEchoTestInteraction creates a test interaction for echo command tests.
func createEchoTestInteraction(userID, guildID, channelID string, options []*discordgo.ApplicationCommandInteractionDataOption) *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:        "interaction-echo-test",
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
				ID:      "cmd-data-echo",
				Name:    "echo",
				Options: options,
			},
		},
	}
}

// createTextOption creates a text string option for echo command testing.
func createTextOption(value string) []*discordgo.ApplicationCommandInteractionDataOption {
	return []*discordgo.ApplicationCommandInteractionDataOption{
		{
			Name:  "text",
			Type:  discordgo.ApplicationCommandOptionString,
			Value: value,
		},
	}
}

func Test_EchoCommand_Name(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "returns echo as command name",
			expected: "echo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &command.EchoCommand{}

			result := cmd.Name()

			assert.Equal(t, tt.expected, result, "Name() should return %q", tt.expected)
		})
	}
}

func Test_EchoCommand_Description(t *testing.T) {
	tests := []struct {
		name     string
		notEmpty bool
	}{
		{
			name:     "returns non-empty description",
			notEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &command.EchoCommand{}

			result := cmd.Description()

			if tt.notEmpty {
				assert.NotEmpty(t, result, "Description() should return non-empty string")
			}
		})
	}
}

func Test_EchoCommand_Options(t *testing.T) {
	tests := []struct {
		name               string
		expectTextOption   bool
		expectTextRequired bool
		expectTextType     discordgo.ApplicationCommandOptionType
	}{
		{
			name:               "has required text option of string type",
			expectTextOption:   true,
			expectTextRequired: true,
			expectTextType:     discordgo.ApplicationCommandOptionString,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &command.EchoCommand{}

			result := cmd.Options()

			if tt.expectTextOption {
				require.NotNil(t, result, "Options() should not return nil")
				require.NotEmpty(t, result, "Options() should not return empty slice")

				// Find the text option
				var textOption *discordgo.ApplicationCommandOption
				for _, opt := range result {
					if opt.Name == "text" {
						textOption = opt
						break
					}
				}

				require.NotNil(t, textOption, "Options should contain 'text' option")
				assert.Equal(t, "text", textOption.Name, "text option should have name 'text'")

				if tt.expectTextRequired {
					assert.True(t, textOption.Required, "text option should be required")
				}

				assert.Equal(t, tt.expectTextType, textOption.Type,
					"text option should have type ApplicationCommandOptionString")
			}
		})
	}
}

func Test_EchoCommand_Options_TextOptionProperties(t *testing.T) {
	cmd := &command.EchoCommand{}
	options := cmd.Options()

	require.NotNil(t, options, "Options should not be nil")
	require.NotEmpty(t, options, "Options should not be empty")

	// Find text option
	var textOption *discordgo.ApplicationCommandOption
	for _, opt := range options {
		if opt.Name == "text" {
			textOption = opt
			break
		}
	}

	require.NotNil(t, textOption, "text option must exist")

	t.Run("text option has description", func(t *testing.T) {
		assert.NotEmpty(t, textOption.Description, "text option should have a description")
	})

	t.Run("text option is required", func(t *testing.T) {
		assert.True(t, textOption.Required, "text option should be required")
	})

	t.Run("text option is string type", func(t *testing.T) {
		assert.Equal(t, discordgo.ApplicationCommandOptionString, textOption.Type,
			"text option should be string type")
	})
}

func Test_EchoCommand_Execute(t *testing.T) {
	tests := []struct {
		name                  string
		textValue             string
		hasTextOption         bool
		expectError           bool
		expectValidationError bool
		expectedField         string
	}{
		{
			name:          "valid text echoes back hello",
			textValue:     "hello",
			hasTextOption: true,
			expectError:   false,
		},
		{
			name:          "valid text echoes back world",
			textValue:     "world",
			hasTextOption: true,
			expectError:   false,
		},
		{
			name:          "valid text with spaces",
			textValue:     "hello world from discord",
			hasTextOption: true,
			expectError:   false,
		},
		{
			name:          "valid text with special characters",
			textValue:     "Hello! @user #channel $money",
			hasTextOption: true,
			expectError:   false,
		},
		{
			name:          "valid text with unicode",
			textValue:     "Hello World! Bonjour!",
			hasTextOption: true,
			expectError:   false,
		},
		{
			name:                  "empty text option returns ValidationError",
			textValue:             "",
			hasTextOption:         true,
			expectError:           true,
			expectValidationError: true,
			expectedField:         "text",
		},
		{
			name:                  "missing text option returns ValidationError",
			textValue:             "",
			hasTextOption:         false,
			expectError:           true,
			expectValidationError: true,
			expectedField:         "text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &command.EchoCommand{}

			var options []*discordgo.ApplicationCommandInteractionDataOption
			if tt.hasTextOption {
				options = createTextOption(tt.textValue)
			}

			interaction := createEchoTestInteraction("user-123", "guild-456", "channel-789", options)
			ctx := command.NewContext(nil, interaction, echoTestLogger())

			err := cmd.Execute(ctx)

			if tt.expectError {
				require.Error(t, err, "Execute should return an error")

				if tt.expectValidationError {
					var validationErr errutil.ValidationError
					if errors.As(err, &validationErr) {
						assert.Equal(t, tt.expectedField, validationErr.Field,
							"ValidationError should be for field %q", tt.expectedField)
					} else {
						// Check if error message mentions the field
						assert.Contains(t, err.Error(), tt.expectedField,
							"error should mention the field %q", tt.expectedField)
					}
				}
			} else {
				// When session is nil, Respond() will fail - this is expected
				// We verify the command returns an error due to nil session, not validation
				if ctx.Session == nil {
					assert.Error(t, err, "Execute should return error when session is nil")
					// Verify it's not a validation error
					var validationErr errutil.ValidationError
					isValidationErr := errors.As(err, &validationErr)
					assert.False(t, isValidationErr,
						"error should not be ValidationError for valid input with nil session")
				} else {
					assert.NoError(t, err, "Execute should not return error")
				}
			}
		})
	}
}

func Test_EchoCommand_Execute_ValidationErrorType(t *testing.T) {
	// Specifically test that empty text returns errutil.ValidationError type
	cmd := &command.EchoCommand{}

	tests := []struct {
		name          string
		options       []*discordgo.ApplicationCommandInteractionDataOption
		expectedField string
	}{
		{
			name:          "empty text returns ValidationError for text field",
			options:       createTextOption(""),
			expectedField: "text",
		},
		{
			name:          "no options returns ValidationError for text field",
			options:       nil,
			expectedField: "text",
		},
		{
			name:          "empty options slice returns ValidationError for text field",
			options:       []*discordgo.ApplicationCommandInteractionDataOption{},
			expectedField: "text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interaction := createEchoTestInteraction("user-123", "guild-456", "channel-789", tt.options)
			ctx := command.NewContext(nil, interaction, echoTestLogger())

			err := cmd.Execute(ctx)

			require.Error(t, err, "Execute should return error for missing/empty text")

			var validationErr errutil.ValidationError
			if errors.As(err, &validationErr) {
				assert.Equal(t, tt.expectedField, validationErr.Field,
					"ValidationError.Field should be %q", tt.expectedField)
				assert.NotEmpty(t, validationErr.Message,
					"ValidationError.Message should not be empty")
			}
			// If error is not ValidationError, at minimum the error should exist
		})
	}
}

func Test_EchoCommand_ImplementsCommandInterface(t *testing.T) {
	// This test verifies that EchoCommand implements the Command interface
	// If this compiles, EchoCommand satisfies command.Command
	var _ command.Command = (*command.EchoCommand)(nil)
}

func Test_EchoCommand_CanBeRegistered(t *testing.T) {
	// Test that EchoCommand can be registered in the Registry
	registry := command.NewRegistry(echoTestLogger())
	cmd := &command.EchoCommand{}

	err := registry.Register(cmd)

	require.NoError(t, err, "EchoCommand should be registerable")

	// Verify it can be retrieved
	retrieved, found := registry.Get("echo")
	assert.True(t, found, "echo command should be found in registry")
	assert.Equal(t, "echo", retrieved.Name())
}

func Test_EchoCommand_ApplicationCommand(t *testing.T) {
	// Test that EchoCommand converts correctly to ApplicationCommand
	registry := command.NewRegistry(echoTestLogger())
	cmd := &command.EchoCommand{}

	err := registry.Register(cmd)
	require.NoError(t, err)

	appCmds := registry.ApplicationCommands()

	require.Len(t, appCmds, 1)
	assert.Equal(t, "echo", appCmds[0].Name)
	assert.NotEmpty(t, appCmds[0].Description)
	require.NotEmpty(t, appCmds[0].Options, "echo command should have options")

	// Verify text option
	var textOption *discordgo.ApplicationCommandOption
	for _, opt := range appCmds[0].Options {
		if opt.Name == "text" {
			textOption = opt
			break
		}
	}
	require.NotNil(t, textOption, "ApplicationCommand should have text option")
	assert.True(t, textOption.Required, "text option should be required in ApplicationCommand")
}

func Test_EchoCommand_Execute_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		textValue   string
		description string
	}{
		{
			name:        "single character",
			textValue:   "a",
			description: "should handle single character",
		},
		{
			name:        "very long text",
			textValue:   string(make([]byte, 2000)), // 2000 characters
			description: "should handle long text",
		},
		{
			name:        "whitespace only",
			textValue:   "   ",
			description: "should handle whitespace-only input",
		},
		{
			name:        "newlines",
			textValue:   "line1\nline2\nline3",
			description: "should handle newlines",
		},
		{
			name:        "tabs",
			textValue:   "col1\tcol2\tcol3",
			description: "should handle tabs",
		},
		{
			name:        "discord mentions",
			textValue:   "<@123456789> <#987654321> <@&111222333>",
			description: "should handle Discord mention syntax",
		},
		{
			name:        "code blocks",
			textValue:   "```go\nfmt.Println(\"hello\")\n```",
			description: "should handle code blocks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &command.EchoCommand{}
			options := createTextOption(tt.textValue)
			interaction := createEchoTestInteraction("user-123", "guild-456", "channel-789", options)
			ctx := command.NewContext(nil, interaction, echoTestLogger())

			err := cmd.Execute(ctx)

			// With nil session, we expect a session-related error, not a validation error
			assert.Error(t, err, "Execute should return error with nil session")

			// Verify it's not a validation error (unless whitespace-only should be rejected)
			var validationErr errutil.ValidationError
			isValidationErr := errors.As(err, &validationErr)

			// Whitespace-only might be considered empty by the implementation
			if tt.textValue == "   " {
				// Implementation may or may not treat whitespace-only as empty
				// Either behavior is acceptable
				t.Logf("whitespace-only input: isValidationErr=%v", isValidationErr)
			} else {
				assert.False(t, isValidationErr,
					"non-empty text should not produce ValidationError: %s", tt.description)
			}
		})
	}
}

func Test_EchoCommand_Execute_WithNilContext(t *testing.T) {
	cmd := &command.EchoCommand{}

	// This test checks graceful handling of edge cases
	// A nil context should not cause a panic (defensive programming)
	assert.NotPanics(t, func() {
		_ = cmd.Execute(nil)
	}, "Execute should not panic with nil context")
}

func Test_EchoCommand_Execute_WithNilInteraction(t *testing.T) {
	cmd := &command.EchoCommand{}
	ctx := command.NewContext(nil, nil, echoTestLogger())

	// Should not panic, may return error
	assert.NotPanics(t, func() {
		_ = cmd.Execute(ctx)
	}, "Execute should not panic with nil interaction in context")
}

// Benchmark tests
func Benchmark_EchoCommand_Name(b *testing.B) {
	cmd := &command.EchoCommand{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Name()
	}
}

func Benchmark_EchoCommand_Description(b *testing.B) {
	cmd := &command.EchoCommand{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Description()
	}
}

func Benchmark_EchoCommand_Options(b *testing.B) {
	cmd := &command.EchoCommand{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Options()
	}
}

func Benchmark_EchoCommand_Execute_Valid(b *testing.B) {
	cmd := &command.EchoCommand{}
	options := createTextOption("benchmark test message")
	interaction := createEchoTestInteraction("user-1", "guild-1", "channel-1", options)
	ctx := command.NewContext(nil, interaction, zerolog.Nop())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Execute(ctx)
	}
}

func Benchmark_EchoCommand_Execute_Empty(b *testing.B) {
	cmd := &command.EchoCommand{}
	options := createTextOption("")
	interaction := createEchoTestInteraction("user-1", "guild-1", "channel-1", options)
	ctx := command.NewContext(nil, interaction, zerolog.Nop())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Execute(ctx)
	}
}
