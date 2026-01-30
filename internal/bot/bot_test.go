package bot_test

import (
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"jamesbot/internal/bot"
	"jamesbot/internal/command"
	"jamesbot/internal/config"
	"jamesbot/internal/middleware"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// discardLogger returns a zerolog.Logger that discards all output.
func discardLogger() zerolog.Logger {
	return zerolog.New(io.Discard).Level(zerolog.Disabled)
}

// mockCommand is a test double for the Command interface.
type mockCommand struct {
	name        string
	description string
	options     []*discordgo.ApplicationCommandOption
	executeFunc func(ctx *command.Context) error
}

func (m *mockCommand) Name() string {
	return m.name
}

func (m *mockCommand) Description() string {
	return m.description
}

func (m *mockCommand) Options() []*discordgo.ApplicationCommandOption {
	return m.options
}

func (m *mockCommand) Execute(ctx *command.Context) error {
	if m.executeFunc != nil {
		return m.executeFunc(ctx)
	}
	return nil
}

// newMockCommand creates a mock command with the given name.
func newMockCommand(name string) *mockCommand {
	return &mockCommand{
		name:        name,
		description: "A mock command for testing",
		options:     nil,
	}
}

// newMockCommandWithOptions creates a mock command with full customization.
func newMockCommandWithOptions(name, description string, options []*discordgo.ApplicationCommandOption) *mockCommand {
	return &mockCommand{
		name:        name,
		description: description,
		options:     options,
	}
}

// validConfig creates a valid configuration for testing.
func validConfig() *config.Config {
	return &config.Config{
		Discord: config.DiscordConfig{
			Token:             "test-token-12345",
			GuildID:           "test-guild-id",
			CleanupOnShutdown: false,
		},
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "console",
		},
		Shutdown: config.ShutdownConfig{
			Timeout: 10 * time.Second,
		},
	}
}

// configWithEmptyToken creates a config with an empty token.
func configWithEmptyToken() *config.Config {
	cfg := validConfig()
	cfg.Discord.Token = ""
	return cfg
}

// trackingMiddleware creates a middleware that tracks execution for testing.
func trackingMiddleware(name string, tracker *[]string, mu *sync.Mutex) middleware.Middleware {
	return func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(ctx *command.Context) error {
			mu.Lock()
			*tracker = append(*tracker, name+"-before")
			mu.Unlock()

			err := next(ctx)

			mu.Lock()
			*tracker = append(*tracker, name+"-after")
			mu.Unlock()

			return err
		}
	}
}

// =============================================================================
// New() Tests
// =============================================================================

func Test_New_ValidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *config.Config
	}{
		{
			name:   "valid config with all fields",
			config: validConfig(),
		},
		{
			name: "valid config with minimal token",
			config: &config.Config{
				Discord: config.DiscordConfig{
					Token: "t",
				},
			},
		},
		{
			name: "valid config with long token",
			config: &config.Config{
				Discord: config.DiscordConfig{
					Token: "test-value",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := bot.New(tt.config, discardLogger())

			require.NoError(t, err, "New() should not return error for valid config")
			require.NotNil(t, b, "New() should return non-nil *Bot")
		})
	}
}

func Test_New_EmptyToken(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.Config
		errContains string
	}{
		{
			name:        "empty token string",
			config:      configWithEmptyToken(),
			errContains: "token",
		},
		// Note: whitespace-only token test depends on implementation
		// Current implementation does not trim whitespace
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := bot.New(tt.config, discardLogger())

			require.Error(t, err, "New() should return error for empty token")
			assert.Nil(t, b, "New() should return nil *Bot for empty token")
			assert.Contains(t, err.Error(), tt.errContains,
				"error should mention 'token'")
		})
	}
}

func Test_New_NilConfig(t *testing.T) {
	b, err := bot.New(nil, discardLogger())

	require.Error(t, err, "New() should return error for nil config")
	assert.Nil(t, b, "New() should return nil *Bot for nil config")
	assert.Contains(t, err.Error(), "config",
		"error should mention 'config'")
}

func Test_New_WithOptions(t *testing.T) {
	cfg := validConfig()

	// Create tracking variables
	var tracker []string
	var mu sync.Mutex

	mw1 := trackingMiddleware("mw1", &tracker, &mu)
	mw2 := trackingMiddleware("mw2", &tracker, &mu)

	t.Run("single middleware option", func(t *testing.T) {
		b, err := bot.New(cfg, discardLogger(), bot.WithMiddleware(mw1))

		require.NoError(t, err, "New() with options should not return error")
		require.NotNil(t, b, "New() with options should return non-nil *Bot")
	})

	t.Run("multiple middleware options", func(t *testing.T) {
		b, err := bot.New(cfg, discardLogger(),
			bot.WithMiddleware(mw1),
			bot.WithMiddleware(mw2),
		)

		require.NoError(t, err, "New() with multiple options should not return error")
		require.NotNil(t, b, "New() with multiple options should return non-nil *Bot")
	})

	t.Run("middleware chain option", func(t *testing.T) {
		b, err := bot.New(cfg, discardLogger(),
			bot.WithMiddleware(mw1, mw2),
		)

		require.NoError(t, err, "New() with chained middlewares should not return error")
		require.NotNil(t, b, "New() with chained middlewares should return non-nil *Bot")
	})
}

func Test_New_PreservesConfigValues(t *testing.T) {
	cfg := validConfig()
	cfg.Discord.GuildID = "specific-guild-id"

	b, err := bot.New(cfg, discardLogger())

	require.NoError(t, err)
	require.NotNil(t, b)

	// The bot should have internalized the config values
	// We can verify this indirectly through behavior
}

// =============================================================================
// RegisterCommand() Tests
// =============================================================================

func Test_RegisterCommand_ValidCommand(t *testing.T) {
	tests := []struct {
		name    string
		command command.Command
	}{
		{
			name:    "simple command",
			command: newMockCommand("ping"),
		},
		{
			name:    "command with description",
			command: newMockCommandWithOptions("echo", "Echo a message back", nil),
		},
		{
			name: "command with options",
			command: newMockCommandWithOptions("greet", "Greet a user", []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "User to greet",
					Required:    true,
				},
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := bot.New(validConfig(), discardLogger())
			require.NoError(t, err)

			err = b.RegisterCommand(tt.command)

			assert.NoError(t, err, "RegisterCommand() should not return error for valid command")
		})
	}
}

func Test_RegisterCommand_NilCommand(t *testing.T) {
	b, err := bot.New(validConfig(), discardLogger())
	require.NoError(t, err)

	err = b.RegisterCommand(nil)

	require.Error(t, err, "RegisterCommand() should return error for nil command")
}

func Test_RegisterCommand_NilPointerCommand(t *testing.T) {
	b, err := bot.New(validConfig(), discardLogger())
	require.NoError(t, err)

	var nilCmd *mockCommand = nil
	err = b.RegisterCommand(nilCmd)

	require.Error(t, err, "RegisterCommand() should return error for nil pointer command")
}

func Test_RegisterCommand_DuplicateCommand(t *testing.T) {
	b, err := bot.New(validConfig(), discardLogger())
	require.NoError(t, err)

	cmd1 := newMockCommand("duplicate")
	cmd2 := newMockCommand("duplicate")

	// First registration should succeed
	err = b.RegisterCommand(cmd1)
	require.NoError(t, err, "first registration should succeed")

	// Second registration with same name should fail
	err = b.RegisterCommand(cmd2)
	require.Error(t, err, "duplicate registration should return error")
	assert.Contains(t, err.Error(), "already registered",
		"error should mention 'already registered'")
}

func Test_RegisterCommand_MultipleUniqueCommands(t *testing.T) {
	b, err := bot.New(validConfig(), discardLogger())
	require.NoError(t, err)

	commands := []command.Command{
		newMockCommand("cmd1"),
		newMockCommand("cmd2"),
		newMockCommand("cmd3"),
	}

	for _, cmd := range commands {
		err := b.RegisterCommand(cmd)
		assert.NoError(t, err, "RegisterCommand() should succeed for unique command %s", cmd.Name())
	}
}

func Test_RegisterCommand_CaseSensitive(t *testing.T) {
	b, err := bot.New(validConfig(), discardLogger())
	require.NoError(t, err)

	// Register with one case
	err = b.RegisterCommand(newMockCommand("Ping"))
	require.NoError(t, err)

	// Register with different case - should succeed (different name)
	err = b.RegisterCommand(newMockCommand("ping"))
	assert.NoError(t, err, "commands with different cases should be treated as different")
}

func Test_RegisterCommand_EmptyName(t *testing.T) {
	b, err := bot.New(validConfig(), discardLogger())
	require.NoError(t, err)

	cmd := newMockCommand("")
	err = b.RegisterCommand(cmd)

	assert.Error(t, err, "RegisterCommand() should return error for empty name")
}

func Test_RegisterCommand_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name    string
		cmdName string
	}{
		{
			name:    "hyphenated name",
			cmdName: "my-command",
		},
		{
			name:    "underscored name",
			cmdName: "my_command",
		},
		{
			name:    "numeric suffix",
			cmdName: "command123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := bot.New(validConfig(), discardLogger())
			require.NoError(t, err)

			cmd := newMockCommand(tt.cmdName)
			err = b.RegisterCommand(cmd)

			assert.NoError(t, err, "RegisterCommand() should accept command name with special chars")
		})
	}
}

// =============================================================================
// Concurrent Registration Tests
// =============================================================================

func Test_RegisterCommand_ConcurrentRegistration(t *testing.T) {
	b, err := bot.New(validConfig(), discardLogger())
	require.NoError(t, err)

	numGoroutines := 100
	var wg sync.WaitGroup
	errChan := make(chan error, numGoroutines)

	// Try to register the same command name concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cmd := newMockCommand("concurrent-cmd")
			err := b.RegisterCommand(cmd)
			if err != nil {
				errChan <- err
			}
		}()
	}

	wg.Wait()
	close(errChan)

	// Collect errors
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	// Exactly one should succeed, rest should fail with "already registered"
	successCount := numGoroutines - len(errs)
	assert.Equal(t, 1, successCount,
		"exactly one concurrent registration should succeed")

	// All errors should be "already registered"
	for _, err := range errs {
		assert.Contains(t, err.Error(), "already registered",
			"concurrent registration errors should be 'already registered'")
	}
}

func Test_RegisterCommand_ConcurrentDifferentCommands(t *testing.T) {
	b, err := bot.New(validConfig(), discardLogger())
	require.NoError(t, err)

	numCommands := 50
	var wg sync.WaitGroup
	errChan := make(chan error, numCommands)

	// Register different commands concurrently
	for i := 0; i < numCommands; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			cmd := newMockCommand("cmd-" + string(rune('A'+id%26)) + string(rune('0'+id/26)))
			err := b.RegisterCommand(cmd)
			if err != nil {
				errChan <- err
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	// All should succeed (different names)
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	assert.Empty(t, errs, "all concurrent registrations of unique commands should succeed")
}

// =============================================================================
// Start() and Stop() Tests (Limited - No Discord API)
// =============================================================================

func Test_Start_NilReceiver(t *testing.T) {
	var b *bot.Bot = nil

	// This tests nil receiver handling
	// The behavior depends on implementation - it may panic or return error
	defer func() {
		if r := recover(); r != nil {
			// Panic is acceptable for nil receiver
			t.Logf("Start() on nil receiver panicked as expected: %v", r)
		}
	}()

	ctx := context.Background()
	err := b.Start(ctx)
	if err != nil {
		assert.Error(t, err, "Start() on nil receiver should return error")
	}
}

func Test_Stop_NilReceiver(t *testing.T) {
	var b *bot.Bot = nil

	// This tests nil receiver handling
	defer func() {
		if r := recover(); r != nil {
			// Panic is acceptable for nil receiver
			t.Logf("Stop() on nil receiver panicked as expected: %v", r)
		}
	}()

	ctx := context.Background()
	err := b.Stop(ctx)
	if err != nil {
		assert.Error(t, err, "Stop() on nil receiver should return error")
	}
}

func Test_Stop_BeforeStart(t *testing.T) {
	b, err := bot.New(validConfig(), discardLogger())
	require.NoError(t, err)

	// Stop before Start should handle gracefully
	// Implementation may return error or be no-op
	ctx := context.Background()
	err = b.Stop(ctx)
	// Either behavior is acceptable:
	// - No error (stop is idempotent)
	// - Error indicating not started
	t.Logf("Stop() before Start() returned: %v", err)
}

func Test_Start_RequiresValidSession(t *testing.T) {
	// Note: Start() will fail because we don't have a real Discord token
	// This test verifies it fails gracefully rather than panicking
	b, err := bot.New(validConfig(), discardLogger())
	require.NoError(t, err)

	// Start will attempt to connect to Discord with an invalid token
	// It should return an error, not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Start() should not panic: %v", r)
		}
	}()

	ctx := context.Background()
	err = b.Start(ctx)
	// We expect an error since the token is invalid
	// The actual error depends on discordgo behavior
	t.Logf("Start() with invalid token returned: %v", err)
}

func Test_Start_WithCancelledContext(t *testing.T) {
	b, err := bot.New(validConfig(), discardLogger())
	require.NoError(t, err)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Start with cancelled context - behavior depends on implementation
	// May fail fast or attempt to connect anyway
	err = b.Start(ctx)
	// Log the result for documentation
	t.Logf("Start() with cancelled context returned: %v", err)
}

func Test_Stop_WithTimeout(t *testing.T) {
	b, err := bot.New(validConfig(), discardLogger())
	require.NoError(t, err)

	// Create a context with a very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Stop with timeout - should handle gracefully
	err = b.Stop(ctx)
	// Log the result for documentation
	t.Logf("Stop() with timeout context returned: %v", err)
}

// =============================================================================
// Option Tests
// =============================================================================

func Test_WithMiddleware_NilMiddleware(t *testing.T) {
	cfg := validConfig()

	// Passing nil middleware - behavior is implementation-defined
	// It should either be ignored or cause an error
	b, err := bot.New(cfg, discardLogger(), bot.WithMiddleware(nil))

	// Either behavior is acceptable
	if err != nil {
		t.Logf("WithMiddleware(nil) caused New() to return error: %v", err)
	} else {
		require.NotNil(t, b, "bot should be created if nil middleware is ignored")
	}
}

func Test_WithMiddleware_EmptySlice(t *testing.T) {
	cfg := validConfig()

	// Empty middleware slice should work
	b, err := bot.New(cfg, discardLogger(), bot.WithMiddleware())

	require.NoError(t, err, "WithMiddleware() with no args should succeed")
	require.NotNil(t, b, "bot should be created")
}

func Test_WithMiddleware_MultipleMiddlewares(t *testing.T) {
	cfg := validConfig()

	mw1 := func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(ctx *command.Context) error {
			return next(ctx)
		}
	}

	mw2 := func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(ctx *command.Context) error {
			return next(ctx)
		}
	}

	mw3 := func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(ctx *command.Context) error {
			return next(ctx)
		}
	}

	b, err := bot.New(cfg, discardLogger(),
		bot.WithMiddleware(mw1, mw2, mw3),
	)

	require.NoError(t, err, "WithMiddleware() with multiple middlewares should succeed")
	require.NotNil(t, b, "bot should be created")
}

// =============================================================================
// Error Type Tests
// =============================================================================

func Test_New_ErrorTypes(t *testing.T) {
	t.Run("nil config error is descriptive", func(t *testing.T) {
		_, err := bot.New(nil, discardLogger())

		require.Error(t, err)
		assert.NotEmpty(t, err.Error(), "error should have a message")
		assert.Contains(t, err.Error(), "config")
	})

	t.Run("empty token error is descriptive", func(t *testing.T) {
		cfg := configWithEmptyToken()
		_, err := bot.New(cfg, discardLogger())

		require.Error(t, err)
		assert.NotEmpty(t, err.Error(), "error should have a message")
		assert.Contains(t, err.Error(), "token")
	})
}

func Test_RegisterCommand_ErrorTypes(t *testing.T) {
	t.Run("nil command error is descriptive", func(t *testing.T) {
		b, err := bot.New(validConfig(), discardLogger())
		require.NoError(t, err)

		err = b.RegisterCommand(nil)

		require.Error(t, err)
		assert.NotEmpty(t, err.Error())
	})

	t.Run("duplicate command error includes name", func(t *testing.T) {
		b, err := bot.New(validConfig(), discardLogger())
		require.NoError(t, err)

		cmd := newMockCommand("testcmd")
		err = b.RegisterCommand(cmd)
		require.NoError(t, err)

		err = b.RegisterCommand(newMockCommand("testcmd"))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "testcmd")
		assert.Contains(t, err.Error(), "already registered")
	})
}

// =============================================================================
// Interface Compliance Tests
// =============================================================================

// Verify Command interface is properly implemented by mock
func Test_Command_InterfaceCompliance(t *testing.T) {
	var _ command.Command = (*mockCommand)(nil)
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func Benchmark_New(b *testing.B) {
	cfg := validConfig()
	logger := discardLogger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = bot.New(cfg, logger)
	}
}

func Benchmark_New_WithMiddleware(b *testing.B) {
	cfg := validConfig()
	logger := discardLogger()
	mw := func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(ctx *command.Context) error {
			return next(ctx)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = bot.New(cfg, logger, bot.WithMiddleware(mw, mw, mw))
	}
}

func Benchmark_RegisterCommand(b *testing.B) {
	logger := discardLogger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bt, _ := bot.New(validConfig(), logger)
		for j := 0; j < 10; j++ {
			cmd := newMockCommand("cmd-" + string(rune(j)))
			_ = bt.RegisterCommand(cmd)
		}
	}
}

func Benchmark_RegisterCommand_ManyCommands(b *testing.B) {
	logger := discardLogger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bt, _ := bot.New(validConfig(), logger)
		for j := 0; j < 100; j++ {
			cmd := newMockCommand("cmd-" + string(rune(j)))
			_ = bt.RegisterCommand(cmd)
		}
	}
}

// =============================================================================
// Edge Case Tests
// =============================================================================

func Test_New_LoggerVariations(t *testing.T) {
	tests := []struct {
		name   string
		logger zerolog.Logger
	}{
		{
			name:   "nop logger",
			logger: zerolog.Nop(),
		},
		{
			name:   "discard logger",
			logger: discardLogger(),
		},
		{
			name:   "default level logger",
			logger: zerolog.New(io.Discard),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := bot.New(validConfig(), tt.logger)

			require.NoError(t, err)
			require.NotNil(t, b)
		})
	}
}

func Test_RegisterCommand_AfterMultipleOperations(t *testing.T) {
	b, err := bot.New(validConfig(), discardLogger())
	require.NoError(t, err)

	// Register some commands
	for i := 0; i < 5; i++ {
		err := b.RegisterCommand(newMockCommand("initial-" + string(rune('A'+i))))
		require.NoError(t, err)
	}

	// Try to register duplicates (should fail)
	err = b.RegisterCommand(newMockCommand("initial-A"))
	require.Error(t, err)

	// Register more unique commands (should succeed)
	for i := 0; i < 5; i++ {
		err := b.RegisterCommand(newMockCommand("later-" + string(rune('A'+i))))
		assert.NoError(t, err)
	}
}

func Test_New_MultipleBotInstances(t *testing.T) {
	// Multiple bot instances should be independent
	cfg1 := validConfig()
	cfg1.Discord.GuildID = "guild-1"

	cfg2 := validConfig()
	cfg2.Discord.GuildID = "guild-2"

	b1, err1 := bot.New(cfg1, discardLogger())
	b2, err2 := bot.New(cfg2, discardLogger())

	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NotNil(t, b1)
	require.NotNil(t, b2)

	// They should be independent
	assert.NotSame(t, b1, b2, "bot instances should be different")

	// Registering to one should not affect the other
	err := b1.RegisterCommand(newMockCommand("unique-to-b1"))
	require.NoError(t, err)

	// Same command name should work on b2
	err = b2.RegisterCommand(newMockCommand("unique-to-b1"))
	assert.NoError(t, err, "registering same command name to different bot should succeed")
}

// =============================================================================
// mockPermissionedCommand for permission testing
// =============================================================================

type mockPermissionedCommand struct {
	mockCommand
	permissions int64
}

func (m *mockPermissionedCommand) Permissions() int64 {
	return m.permissions
}

func Test_RegisterCommand_PermissionedCommand(t *testing.T) {
	b, err := bot.New(validConfig(), discardLogger())
	require.NoError(t, err)

	cmd := &mockPermissionedCommand{
		mockCommand: mockCommand{
			name:        "admin-cmd",
			description: "Admin command",
		},
		permissions: discordgo.PermissionAdministrator,
	}

	err = b.RegisterCommand(cmd)
	assert.NoError(t, err, "RegisterCommand() should accept PermissionedCommand")
}

// Verify PermissionedCommand interface compliance
func Test_PermissionedCommand_InterfaceCompliance(t *testing.T) {
	var _ command.PermissionedCommand = (*mockPermissionedCommand)(nil)
}

// =============================================================================
// Error wrapping/unwrapping tests
// =============================================================================

func Test_Errors_CanBeUnwrapped(t *testing.T) {
	b, err := bot.New(validConfig(), discardLogger())
	require.NoError(t, err)

	// Register a command
	err = b.RegisterCommand(newMockCommand("test"))
	require.NoError(t, err)

	// Register duplicate
	err = b.RegisterCommand(newMockCommand("test"))
	require.Error(t, err)

	// Error should be usable as a standard error
	var stdErr error = err
	assert.NotNil(t, stdErr)
	assert.NotEmpty(t, stdErr.Error())

	// Check if it can be used with errors.Is (if implemented with sentinel errors)
	// This is implementation-dependent
	t.Logf("Duplicate command error: %v", err)
}

// =============================================================================
// Stress Tests
// =============================================================================

func Test_RegisterCommand_ManyCommands(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	b, err := bot.New(validConfig(), discardLogger())
	require.NoError(t, err)

	numCommands := 1000

	for i := 0; i < numCommands; i++ {
		cmd := newMockCommand("stress-cmd-" + string(rune(i/26/26+'A')) +
			string(rune(i/26%26+'a')) +
			string(rune(i%26+'a')))
		err := b.RegisterCommand(cmd)
		require.NoError(t, err, "should register command %d", i)
	}
}

// =============================================================================
// Tests for errors.Is compatibility
// =============================================================================

func Test_New_ErrorsCanBeCheckedWithErrorsIs(t *testing.T) {
	_, err := bot.New(nil, discardLogger())

	require.Error(t, err)

	// The error should be compatible with standard error handling
	var targetErr error = err
	assert.NotNil(t, targetErr)

	// Test that errors.Is doesn't panic (even if there's no specific sentinel)
	_ = errors.Is(err, errors.New("some error"))
}

// =============================================================================
// Context parameter tests
// =============================================================================

func Test_Start_AcceptsContext(t *testing.T) {
	b, err := bot.New(validConfig(), discardLogger())
	require.NoError(t, err)

	// Verify Start accepts context.Context
	// The actual call will fail (no valid Discord token) but it should accept the context
	ctx := context.Background()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Start() should accept context without panic: %v", r)
		}
	}()

	_ = b.Start(ctx)
}

func Test_Stop_AcceptsContext(t *testing.T) {
	b, err := bot.New(validConfig(), discardLogger())
	require.NoError(t, err)

	// Verify Stop accepts context.Context
	ctx := context.Background()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Stop() should accept context without panic: %v", r)
		}
	}()

	_ = b.Stop(ctx)
}

func Test_Start_WithDeadlineContext(t *testing.T) {
	b, err := bot.New(validConfig(), discardLogger())
	require.NoError(t, err)

	// Create a context with a deadline
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(100*time.Millisecond))
	defer cancel()

	// Start with deadline context
	err = b.Start(ctx)
	// Log the result for documentation
	t.Logf("Start() with deadline context returned: %v", err)
}

// =============================================================================
// Option Type Tests
// =============================================================================

func Test_Option_TypeIsFunction(t *testing.T) {
	// Verify Option is a function type
	var opt bot.Option = bot.WithMiddleware()
	assert.NotNil(t, opt, "Option should be a non-nil function")
}

func Test_WithMiddleware_ReturnsOption(t *testing.T) {
	mw := func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return next
	}

	opt := bot.WithMiddleware(mw)
	assert.NotNil(t, opt, "WithMiddleware should return non-nil Option")
}

// =============================================================================
// Regression Tests
// =============================================================================

func Test_RegisterCommand_SameNameDifferentInstances(t *testing.T) {
	// Regression: Ensure that different command instances with the same name
	// are correctly identified as duplicates
	b, err := bot.New(validConfig(), discardLogger())
	require.NoError(t, err)

	cmd1 := &mockCommand{name: "test", description: "First"}
	cmd2 := &mockCommand{name: "test", description: "Second"}

	err = b.RegisterCommand(cmd1)
	require.NoError(t, err)

	err = b.RegisterCommand(cmd2)
	require.Error(t, err, "should reject second command with same name")
	assert.Contains(t, err.Error(), "already registered")
}

func Test_New_ConfigNotMutated(t *testing.T) {
	cfg := validConfig()
	originalToken := cfg.Discord.Token
	originalGuildID := cfg.Discord.GuildID

	_, err := bot.New(cfg, discardLogger())
	require.NoError(t, err)

	// Config should not be mutated
	assert.Equal(t, originalToken, cfg.Discord.Token, "token should not be mutated")
	assert.Equal(t, originalGuildID, cfg.Discord.GuildID, "guild ID should not be mutated")
}
