package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"jamesbot/internal/command"
	"jamesbot/internal/handler"
	"jamesbot/internal/middleware"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCommand implements command.Command for testing.
type mockCommand struct {
	name        string
	description string
	options     []*discordgo.ApplicationCommandOption
	executeFunc func(ctx *command.Context) error
	executed    bool
	executedCtx *command.Context
}

func (m *mockCommand) Name() string        { return m.name }
func (m *mockCommand) Description() string { return m.description }
func (m *mockCommand) Options() []*discordgo.ApplicationCommandOption {
	return m.options
}
func (m *mockCommand) Execute(ctx *command.Context) error {
	m.executed = true
	m.executedCtx = ctx
	if m.executeFunc != nil {
		return m.executeFunc(ctx)
	}
	return nil
}

// newMockCommand creates a mock command with the given name.
func newMockCommand(name string) *mockCommand {
	return &mockCommand{
		name:        name,
		description: "Mock command for testing",
	}
}

// interactionLogCapture provides log capture for interaction handler tests.
type interactionLogCapture struct {
	buf *bytes.Buffer
}

func newInteractionLogCapture() *interactionLogCapture {
	return &interactionLogCapture{buf: &bytes.Buffer{}}
}

func (lc *interactionLogCapture) logger() zerolog.Logger {
	return zerolog.New(lc.buf).With().Timestamp().Logger()
}

func (lc *interactionLogCapture) entries() []map[string]interface{} {
	var entries []map[string]interface{}
	lines := bytes.Split(lc.buf.Bytes(), []byte("\n"))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		var entry map[string]interface{}
		if err := json.Unmarshal(line, &entry); err == nil {
			entries = append(entries, entry)
		}
	}
	return entries
}

func (lc *interactionLogCapture) contains(s string) bool {
	return bytes.Contains(lc.buf.Bytes(), []byte(s))
}

func (lc *interactionLogCapture) containsLevel(level string) bool {
	for _, entry := range lc.entries() {
		if l, ok := entry["level"].(string); ok && l == level {
			return true
		}
	}
	return false
}

// createTestInteraction creates a discordgo.InteractionCreate for testing.
func createTestInteraction(cmdName string, interactionType discordgo.InteractionType) *discordgo.InteractionCreate {
	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:        "test-interaction-id",
			ChannelID: "test-channel",
			GuildID:   "test-guild",
			Member: &discordgo.Member{
				User: &discordgo.User{
					ID:       "test-user",
					Username: "testuser",
				},
			},
			Type: interactionType,
		},
	}

	if interactionType == discordgo.InteractionApplicationCommand {
		interaction.Interaction.Data = discordgo.ApplicationCommandInteractionData{
			ID:   "cmd-data-id",
			Name: cmdName,
		}
	} else if interactionType == discordgo.InteractionMessageComponent {
		interaction.Interaction.Data = discordgo.MessageComponentInteractionData{
			CustomID: "button-click",
		}
	}

	return interaction
}

// createTestRegistry creates a command registry with the given commands.
func createTestRegistry(logger zerolog.Logger, commands ...*mockCommand) *command.Registry {
	registry := command.NewRegistry(logger)
	for _, cmd := range commands {
		_ = registry.Register(cmd)
	}
	return registry
}

// noopMiddleware returns a middleware that does nothing (pass-through).
func noopMiddleware() middleware.Middleware {
	return func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return next
	}
}

func Test_NewInteractionHandler(t *testing.T) {
	tests := []struct {
		name       string
		registry   *command.Registry
		middleware middleware.Middleware
		logger     zerolog.Logger
	}{
		{
			name:       "create handler with all parameters",
			registry:   command.NewRegistry(zerolog.Nop()),
			middleware: noopMiddleware(),
			logger:     zerolog.Nop(),
		},
		{
			name:       "create handler with nil middleware",
			registry:   command.NewRegistry(zerolog.Nop()),
			middleware: nil,
			logger:     zerolog.Nop(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := handler.NewInteractionHandler(tt.registry, tt.middleware, tt.logger)

			require.NotNil(t, h, "NewInteractionHandler should return non-nil *InteractionHandler")
		})
	}
}

func Test_InteractionHandler_Handle_ValidCommand(t *testing.T) {
	capture := newInteractionLogCapture()
	logger := capture.logger()

	pingCmd := newMockCommand("ping")
	registry := createTestRegistry(logger, pingCmd)

	h := handler.NewInteractionHandler(registry, noopMiddleware(), logger)

	interaction := createTestInteraction("ping", discordgo.InteractionApplicationCommand)
	h.Handle(nil, interaction)

	assert.True(t, pingCmd.executed, "registered command should be executed")
	assert.NotNil(t, pingCmd.executedCtx, "command should receive context")
}

func Test_InteractionHandler_Handle_UnknownCommand(t *testing.T) {
	capture := newInteractionLogCapture()
	logger := capture.logger()

	// Register "ping" but try to execute "foo"
	pingCmd := newMockCommand("ping")
	registry := createTestRegistry(logger, pingCmd)

	h := handler.NewInteractionHandler(registry, noopMiddleware(), logger)

	interaction := createTestInteraction("foo", discordgo.InteractionApplicationCommand)

	// Should not panic
	assert.NotPanics(t, func() {
		h.Handle(nil, interaction)
	}, "Handle should not panic with unknown command")

	// ping command should NOT be executed
	assert.False(t, pingCmd.executed, "registered command should not be executed for wrong name")

	// Should log a warning
	assert.True(t, capture.containsLevel("warn") || capture.contains("unknown") || capture.contains("not found"),
		"should log warning for unknown command")
}

func Test_InteractionHandler_Handle_NonCommandInteraction(t *testing.T) {
	capture := newInteractionLogCapture()
	logger := capture.logger()

	pingCmd := newMockCommand("ping")
	registry := createTestRegistry(logger, pingCmd)

	h := handler.NewInteractionHandler(registry, noopMiddleware(), logger)

	tests := []struct {
		name            string
		interactionType discordgo.InteractionType
	}{
		{
			name:            "message component interaction (button click)",
			interactionType: discordgo.InteractionMessageComponent,
		},
		{
			name:            "autocomplete interaction",
			interactionType: discordgo.InteractionApplicationCommandAutocomplete,
		},
		{
			name:            "modal submit interaction",
			interactionType: discordgo.InteractionModalSubmit,
		},
		{
			name:            "ping interaction",
			interactionType: discordgo.InteractionPing,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pingCmd.executed = false // Reset
			interaction := createTestInteraction("ping", tt.interactionType)

			// Should not panic
			assert.NotPanics(t, func() {
				h.Handle(nil, interaction)
			}, "Handle should not panic with non-command interaction")

			// Command should NOT be executed for non-command interactions
			assert.False(t, pingCmd.executed,
				"command should not be executed for non-command interaction type")
		})
	}
}

func Test_InteractionHandler_Handle_NilInteraction(t *testing.T) {
	capture := newInteractionLogCapture()
	logger := capture.logger()

	pingCmd := newMockCommand("ping")
	registry := createTestRegistry(logger, pingCmd)

	h := handler.NewInteractionHandler(registry, noopMiddleware(), logger)

	// Should not panic with nil interaction
	assert.NotPanics(t, func() {
		h.Handle(nil, nil)
	}, "Handle should not panic with nil interaction")

	assert.False(t, pingCmd.executed, "command should not be executed with nil interaction")
}

func Test_InteractionHandler_Handle_NilInteractionData(t *testing.T) {
	capture := newInteractionLogCapture()
	logger := capture.logger()

	pingCmd := newMockCommand("ping")
	registry := createTestRegistry(logger, pingCmd)

	h := handler.NewInteractionHandler(registry, noopMiddleware(), logger)

	// Create interaction with application command type but no data
	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:        "test-id",
			ChannelID: "test-channel",
			GuildID:   "test-guild",
			Type:      discordgo.InteractionApplicationCommand,
			// Data is nil/empty
		},
	}

	// Should not panic with nil/empty data
	assert.NotPanics(t, func() {
		h.Handle(nil, interaction)
	}, "Handle should not panic with nil interaction data")

	assert.False(t, pingCmd.executed, "command should not be executed with nil data")
}

func Test_InteractionHandler_Handle_MiddlewareExecution(t *testing.T) {
	capture := newInteractionLogCapture()
	logger := capture.logger()

	middlewareExecuted := false
	testMiddleware := func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(ctx *command.Context) error {
			middlewareExecuted = true
			return next(ctx)
		}
	}

	pingCmd := newMockCommand("ping")
	registry := createTestRegistry(logger, pingCmd)

	h := handler.NewInteractionHandler(registry, testMiddleware, logger)

	interaction := createTestInteraction("ping", discordgo.InteractionApplicationCommand)
	h.Handle(nil, interaction)

	assert.True(t, middlewareExecuted, "middleware should be executed")
	assert.True(t, pingCmd.executed, "command should be executed after middleware")
}

func Test_InteractionHandler_Handle_MiddlewareCanBlockCommand(t *testing.T) {
	capture := newInteractionLogCapture()
	logger := capture.logger()

	blockingMiddleware := func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(ctx *command.Context) error {
			// Don't call next - block execution
			return errors.New("blocked by middleware")
		}
	}

	pingCmd := newMockCommand("ping")
	registry := createTestRegistry(logger, pingCmd)

	h := handler.NewInteractionHandler(registry, blockingMiddleware, logger)

	interaction := createTestInteraction("ping", discordgo.InteractionApplicationCommand)
	h.Handle(nil, interaction)

	assert.False(t, pingCmd.executed,
		"command should not be executed when middleware blocks")
}

func Test_InteractionHandler_Handle_CommandError(t *testing.T) {
	capture := newInteractionLogCapture()
	logger := capture.logger()

	cmdError := errors.New("command execution failed")
	failingCmd := newMockCommand("failing")
	failingCmd.executeFunc = func(ctx *command.Context) error {
		return cmdError
	}

	registry := createTestRegistry(logger, failingCmd)

	h := handler.NewInteractionHandler(registry, noopMiddleware(), logger)

	interaction := createTestInteraction("failing", discordgo.InteractionApplicationCommand)

	// Should not panic when command returns error
	assert.NotPanics(t, func() {
		h.Handle(nil, interaction)
	}, "Handle should not panic when command returns error")

	assert.True(t, failingCmd.executed, "command should still be executed")
}

func Test_InteractionHandler_Handle_MultipleCommands(t *testing.T) {
	capture := newInteractionLogCapture()
	logger := capture.logger()

	pingCmd := newMockCommand("ping")
	helpCmd := newMockCommand("help")
	infoCmd := newMockCommand("info")

	registry := createTestRegistry(logger, pingCmd, helpCmd, infoCmd)

	h := handler.NewInteractionHandler(registry, noopMiddleware(), logger)

	// Execute ping
	interaction := createTestInteraction("ping", discordgo.InteractionApplicationCommand)
	h.Handle(nil, interaction)

	assert.True(t, pingCmd.executed, "ping command should be executed")
	assert.False(t, helpCmd.executed, "help command should not be executed")
	assert.False(t, infoCmd.executed, "info command should not be executed")

	// Reset and execute help
	pingCmd.executed = false
	interaction = createTestInteraction("help", discordgo.InteractionApplicationCommand)
	h.Handle(nil, interaction)

	assert.False(t, pingCmd.executed, "ping command should not be executed")
	assert.True(t, helpCmd.executed, "help command should be executed")
	assert.False(t, infoCmd.executed, "info command should not be executed")
}

func Test_InteractionHandler_Handle_ContextContainsInteraction(t *testing.T) {
	logger := zerolog.Nop()

	pingCmd := newMockCommand("ping")
	registry := createTestRegistry(logger, pingCmd)

	h := handler.NewInteractionHandler(registry, noopMiddleware(), logger)

	interaction := createTestInteraction("ping", discordgo.InteractionApplicationCommand)
	h.Handle(nil, interaction)

	require.NotNil(t, pingCmd.executedCtx, "command should receive context")
	assert.Equal(t, interaction, pingCmd.executedCtx.Interaction,
		"context should contain the original interaction")
}

func Test_InteractionHandler_Handle_EmptyCommandName(t *testing.T) {
	capture := newInteractionLogCapture()
	logger := capture.logger()

	pingCmd := newMockCommand("ping")
	registry := createTestRegistry(logger, pingCmd)

	h := handler.NewInteractionHandler(registry, noopMiddleware(), logger)

	// Create interaction with empty command name
	interaction := createTestInteraction("", discordgo.InteractionApplicationCommand)

	// Should not panic
	assert.NotPanics(t, func() {
		h.Handle(nil, interaction)
	}, "Handle should not panic with empty command name")

	assert.False(t, pingCmd.executed, "command should not be executed with empty name")
}

func Test_InteractionHandler_Handle_ChainedMiddleware(t *testing.T) {
	logger := zerolog.Nop()

	executionOrder := []string{}

	mw1 := func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(ctx *command.Context) error {
			executionOrder = append(executionOrder, "mw1-before")
			err := next(ctx)
			executionOrder = append(executionOrder, "mw1-after")
			return err
		}
	}

	mw2 := func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(ctx *command.Context) error {
			executionOrder = append(executionOrder, "mw2-before")
			err := next(ctx)
			executionOrder = append(executionOrder, "mw2-after")
			return err
		}
	}

	chainedMiddleware := middleware.Chain(mw1, mw2)

	pingCmd := newMockCommand("ping")
	pingCmd.executeFunc = func(ctx *command.Context) error {
		executionOrder = append(executionOrder, "command")
		return nil
	}

	registry := createTestRegistry(logger, pingCmd)

	h := handler.NewInteractionHandler(registry, chainedMiddleware, logger)

	interaction := createTestInteraction("ping", discordgo.InteractionApplicationCommand)
	h.Handle(nil, interaction)

	expected := []string{"mw1-before", "mw2-before", "command", "mw2-after", "mw1-after"}
	assert.Equal(t, expected, executionOrder,
		"middleware should execute in correct order")
}

func Test_InteractionHandler_Handle_NilMiddleware(t *testing.T) {
	logger := zerolog.Nop()

	pingCmd := newMockCommand("ping")
	registry := createTestRegistry(logger, pingCmd)

	// Create handler with nil middleware
	h := handler.NewInteractionHandler(registry, nil, logger)

	interaction := createTestInteraction("ping", discordgo.InteractionApplicationCommand)

	// Should not panic with nil middleware
	assert.NotPanics(t, func() {
		h.Handle(nil, interaction)
	}, "Handle should not panic with nil middleware")

	// Command should still be executed
	assert.True(t, pingCmd.executed, "command should be executed even with nil middleware")
}

// Benchmark tests

func Benchmark_InteractionHandler_Handle(b *testing.B) {
	logger := zerolog.Nop()
	pingCmd := newMockCommand("ping")
	registry := createTestRegistry(logger, pingCmd)
	h := handler.NewInteractionHandler(registry, noopMiddleware(), logger)
	interaction := createTestInteraction("ping", discordgo.InteractionApplicationCommand)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pingCmd.executed = false
		h.Handle(nil, interaction)
	}
}

func Benchmark_InteractionHandler_Handle_WithMiddleware(b *testing.B) {
	logger := zerolog.Nop()

	mw := func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(ctx *command.Context) error {
			return next(ctx)
		}
	}
	chainedMiddleware := middleware.Chain(mw, mw, mw)

	pingCmd := newMockCommand("ping")
	registry := createTestRegistry(logger, pingCmd)
	h := handler.NewInteractionHandler(registry, chainedMiddleware, logger)
	interaction := createTestInteraction("ping", discordgo.InteractionApplicationCommand)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pingCmd.executed = false
		h.Handle(nil, interaction)
	}
}

func Benchmark_InteractionHandler_Handle_UnknownCommand(b *testing.B) {
	logger := zerolog.Nop()
	pingCmd := newMockCommand("ping")
	registry := createTestRegistry(logger, pingCmd)
	h := handler.NewInteractionHandler(registry, noopMiddleware(), logger)
	interaction := createTestInteraction("unknown", discordgo.InteractionApplicationCommand)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Handle(nil, interaction)
	}
}
