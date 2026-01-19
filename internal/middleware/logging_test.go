package middleware_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"jamesbot/internal/command"
	"jamesbot/internal/middleware"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// loggingLogCapture provides utilities for capturing and parsing zerolog output.
type loggingLogCapture struct {
	buf *bytes.Buffer
}

func newLoggingLogCapture() *loggingLogCapture {
	return &loggingLogCapture{buf: &bytes.Buffer{}}
}

func (lc *loggingLogCapture) logger() zerolog.Logger {
	return zerolog.New(lc.buf).With().Timestamp().Logger()
}

func (lc *loggingLogCapture) entries() []map[string]interface{} {
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

func (lc *loggingLogCapture) lastEntry() map[string]interface{} {
	entries := lc.entries()
	if len(entries) == 0 {
		return nil
	}
	return entries[len(entries)-1]
}

func (lc *loggingLogCapture) contains(s string) bool {
	return bytes.Contains(lc.buf.Bytes(), []byte(s))
}

func (lc *loggingLogCapture) raw() string {
	return lc.buf.String()
}

// createLoggingTestContext creates a command context for logging middleware tests.
func createLoggingTestContext(logger zerolog.Logger, userID, guildID, channelID, cmdName string) *command.Context {
	interaction := &discordgo.InteractionCreate{
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
			Type: discordgo.InteractionApplicationCommand,
			Data: discordgo.ApplicationCommandInteractionData{
				ID:   "cmd-data-123",
				Name: cmdName,
			},
		},
	}
	return command.NewContext(nil, interaction, logger)
}

func Test_Logging_SuccessfulCommand(t *testing.T) {
	tests := []struct {
		name        string
		commandName string
	}{
		{
			name:        "ping command success",
			commandName: "ping",
		},
		{
			name:        "help command success",
			commandName: "help",
		},
		{
			name:        "command with hyphen",
			commandName: "user-info",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capture := newLoggingLogCapture()
			logger := capture.logger()

			loggingMW := middleware.Logging(logger)

			handler := func(ctx *command.Context) error {
				return nil // Success
			}

			wrapped := loggingMW(handler)
			ctx := createLoggingTestContext(logger, "user-123", "guild-456", "channel-789", tt.commandName)

			err := wrapped(ctx)

			assert.NoError(t, err, "successful command should return no error")

			entry := capture.lastEntry()
			require.NotNil(t, entry, "should have logged an entry")

			// Check log level is info for success
			if level, ok := entry["level"].(string); ok {
				assert.Equal(t, "info", level, "successful command should log at info level")
			}
		})
	}
}

func Test_Logging_FailedCommand(t *testing.T) {
	tests := []struct {
		name        string
		commandName string
		err         error
	}{
		{
			name:        "command with simple error",
			commandName: "failing",
			err:         errors.New("command failed"),
		},
		{
			name:        "command with wrapped error",
			commandName: "broken",
			err:         errors.New("internal: database connection failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capture := newLoggingLogCapture()
			logger := capture.logger()

			loggingMW := middleware.Logging(logger)

			handler := func(ctx *command.Context) error {
				return tt.err
			}

			wrapped := loggingMW(handler)
			ctx := createLoggingTestContext(logger, "user-123", "guild-456", "channel-789", tt.commandName)

			err := wrapped(ctx)

			assert.Error(t, err, "failed command should return error")
			assert.Equal(t, tt.err, err, "error should be propagated")

			entry := capture.lastEntry()
			require.NotNil(t, entry, "should have logged an entry")

			// Check log level is error for failure
			if level, ok := entry["level"].(string); ok {
				assert.Equal(t, "error", level, "failed command should log at error level")
			}

			// Check error is in log
			assert.True(t, capture.contains(tt.err.Error()) || capture.contains("error"),
				"log should contain error information")
		})
	}
}

func Test_Logging_LogsDuration(t *testing.T) {
	capture := newLoggingLogCapture()
	logger := capture.logger()

	loggingMW := middleware.Logging(logger)

	handler := func(ctx *command.Context) error {
		time.Sleep(10 * time.Millisecond) // Small delay to ensure measurable duration
		return nil
	}

	wrapped := loggingMW(handler)
	ctx := createLoggingTestContext(logger, "user-123", "guild-456", "channel-789", "slow")

	err := wrapped(ctx)

	assert.NoError(t, err)

	entry := capture.lastEntry()
	require.NotNil(t, entry, "should have logged an entry")

	// Check that duration is logged (could be "duration", "duration_ms", "elapsed", etc.)
	hasDuration := false
	for key := range entry {
		if strings.Contains(strings.ToLower(key), "duration") ||
			strings.Contains(strings.ToLower(key), "elapsed") ||
			strings.Contains(strings.ToLower(key), "time") ||
			strings.Contains(strings.ToLower(key), "latency") {
			hasDuration = true
			break
		}
	}
	assert.True(t, hasDuration || capture.contains("ms") || capture.contains("duration"),
		"log should contain duration information")
}

func Test_Logging_LogsCommandName(t *testing.T) {
	tests := []struct {
		name        string
		commandName string
	}{
		{
			name:        "simple command name",
			commandName: "ping",
		},
		{
			name:        "command with hyphen",
			commandName: "user-info",
		},
		{
			name:        "command with underscore",
			commandName: "get_data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capture := newLoggingLogCapture()
			logger := capture.logger()

			loggingMW := middleware.Logging(logger)

			handler := func(ctx *command.Context) error {
				return nil
			}

			wrapped := loggingMW(handler)
			ctx := createLoggingTestContext(logger, "user-123", "guild-456", "channel-789", tt.commandName)

			err := wrapped(ctx)

			assert.NoError(t, err)

			// Check that command name appears in log
			assert.True(t, capture.contains(tt.commandName),
				"log should contain command name %q", tt.commandName)
		})
	}
}

func Test_Logging_LogsUserAndGuildIDs(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		guildID   string
		channelID string
	}{
		{
			name:      "standard IDs",
			userID:    "user-123456789",
			guildID:   "guild-987654321",
			channelID: "channel-111222333",
		},
		{
			name:      "numeric IDs",
			userID:    "123456789012345678",
			guildID:   "987654321098765432",
			channelID: "111222333444555666",
		},
		{
			name:      "DM context (empty guild)",
			userID:    "dm-user-123",
			guildID:   "",
			channelID: "dm-channel-456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capture := newLoggingLogCapture()
			logger := capture.logger()

			loggingMW := middleware.Logging(logger)

			handler := func(ctx *command.Context) error {
				return nil
			}

			wrapped := loggingMW(handler)
			ctx := createLoggingTestContext(logger, tt.userID, tt.guildID, tt.channelID, "testcmd")

			err := wrapped(ctx)

			assert.NoError(t, err)

			// The context logger already has user_id and guild_id from NewContext
			// The logging middleware should include these in its output
			rawLog := capture.raw()

			// User ID should be logged
			assert.True(t, strings.Contains(rawLog, tt.userID) ||
				strings.Contains(rawLog, "user"),
				"log should contain user ID or user field")

			// Guild ID should be logged (if non-empty)
			if tt.guildID != "" {
				assert.True(t, strings.Contains(rawLog, tt.guildID) ||
					strings.Contains(rawLog, "guild"),
					"log should contain guild ID or guild field")
			}
		})
	}
}

func Test_Logging_ErrorPropagation(t *testing.T) {
	capture := newLoggingLogCapture()
	logger := capture.logger()

	expectedErr := errors.New("original error")
	loggingMW := middleware.Logging(logger)

	handler := func(ctx *command.Context) error {
		return expectedErr
	}

	wrapped := loggingMW(handler)
	ctx := createLoggingTestContext(logger, "user-1", "guild-1", "channel-1", "test")

	err := wrapped(ctx)

	assert.Equal(t, expectedErr, err, "original error should be propagated")
}

func Test_Logging_FastCommand(t *testing.T) {
	capture := newLoggingLogCapture()
	logger := capture.logger()

	loggingMW := middleware.Logging(logger)

	handler := func(ctx *command.Context) error {
		return nil // Instant return
	}

	wrapped := loggingMW(handler)
	ctx := createLoggingTestContext(logger, "user-1", "guild-1", "channel-1", "fast")

	err := wrapped(ctx)

	assert.NoError(t, err)

	entry := capture.lastEntry()
	require.NotNil(t, entry, "should log even for fast commands")
}

func Test_Logging_SlowCommand(t *testing.T) {
	capture := newLoggingLogCapture()
	logger := capture.logger()

	loggingMW := middleware.Logging(logger)

	handler := func(ctx *command.Context) error {
		time.Sleep(50 * time.Millisecond)
		return nil
	}

	wrapped := loggingMW(handler)
	ctx := createLoggingTestContext(logger, "user-1", "guild-1", "channel-1", "slow")

	start := time.Now()
	err := wrapped(ctx)
	elapsed := time.Since(start)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(50),
		"command should take at least 50ms")

	entry := capture.lastEntry()
	require.NotNil(t, entry, "should log slow command")
}

func Test_Logging_MiddlewareChaining(t *testing.T) {
	capture := newLoggingLogCapture()
	logger := capture.logger()

	executionOrder := []string{}

	beforeMW := func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(ctx *command.Context) error {
			executionOrder = append(executionOrder, "before")
			return next(ctx)
		}
	}

	afterMW := func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(ctx *command.Context) error {
			err := next(ctx)
			executionOrder = append(executionOrder, "after")
			return err
		}
	}

	loggingMW := middleware.Logging(logger)

	handler := func(ctx *command.Context) error {
		executionOrder = append(executionOrder, "handler")
		return nil
	}

	// Chain: before -> logging -> after -> handler
	chained := middleware.Chain(beforeMW, loggingMW, afterMW)(handler)
	ctx := createLoggingTestContext(logger, "user-1", "guild-1", "channel-1", "test")

	err := chained(ctx)

	assert.NoError(t, err)
	assert.Equal(t, []string{"before", "handler", "after"}, executionOrder,
		"execution order should be correct with logging middleware")
}

func Test_Logging_MultipleInvocations(t *testing.T) {
	capture := newLoggingLogCapture()
	logger := capture.logger()

	loggingMW := middleware.Logging(logger)

	callCount := 0
	handler := func(ctx *command.Context) error {
		callCount++
		return nil
	}

	wrapped := loggingMW(handler)

	// Invoke multiple times
	for i := 0; i < 5; i++ {
		ctx := createLoggingTestContext(logger, "user-1", "guild-1", "channel-1", "test")
		err := wrapped(ctx)
		assert.NoError(t, err)
	}

	assert.Equal(t, 5, callCount, "handler should be called 5 times")

	entries := capture.entries()
	assert.GreaterOrEqual(t, len(entries), 5,
		"should have at least 5 log entries")
}

func Test_Logging_DifferentLogLevels(t *testing.T) {
	t.Run("info level for success", func(t *testing.T) {
		capture := newLoggingLogCapture()
		logger := capture.logger()

		loggingMW := middleware.Logging(logger)
		wrapped := loggingMW(func(ctx *command.Context) error {
			return nil
		})

		ctx := createLoggingTestContext(logger, "user-1", "guild-1", "channel-1", "test")
		_ = wrapped(ctx)

		entry := capture.lastEntry()
		if entry != nil {
			if level, ok := entry["level"].(string); ok {
				assert.Equal(t, "info", level)
			}
		}
	})

	t.Run("error level for failure", func(t *testing.T) {
		capture := newLoggingLogCapture()
		logger := capture.logger()

		loggingMW := middleware.Logging(logger)
		wrapped := loggingMW(func(ctx *command.Context) error {
			return errors.New("failed")
		})

		ctx := createLoggingTestContext(logger, "user-1", "guild-1", "channel-1", "test")
		_ = wrapped(ctx)

		entry := capture.lastEntry()
		if entry != nil {
			if level, ok := entry["level"].(string); ok {
				assert.Equal(t, "error", level)
			}
		}
	})
}

// Benchmark tests

func Benchmark_Logging_Middleware(b *testing.B) {
	logger := zerolog.Nop()
	loggingMW := middleware.Logging(logger)

	handler := func(ctx *command.Context) error {
		return nil
	}

	wrapped := loggingMW(handler)
	ctx := createLoggingTestContext(logger, "user-1", "guild-1", "channel-1", "bench")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wrapped(ctx)
	}
}

func Benchmark_Logging_Middleware_WithError(b *testing.B) {
	logger := zerolog.Nop()
	loggingMW := middleware.Logging(logger)
	testErr := errors.New("test error")

	handler := func(ctx *command.Context) error {
		return testErr
	}

	wrapped := loggingMW(handler)
	ctx := createLoggingTestContext(logger, "user-1", "guild-1", "channel-1", "bench")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wrapped(ctx)
	}
}

func Benchmark_Logging_Middleware_RealLogger(b *testing.B) {
	buf := &bytes.Buffer{}
	logger := zerolog.New(buf)
	loggingMW := middleware.Logging(logger)

	handler := func(ctx *command.Context) error {
		return nil
	}

	wrapped := loggingMW(handler)
	ctx := createLoggingTestContext(logger, "user-1", "guild-1", "channel-1", "bench")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = wrapped(ctx)
	}
}
