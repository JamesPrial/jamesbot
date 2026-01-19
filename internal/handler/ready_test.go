package handler_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"jamesbot/internal/handler"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// logCapture provides utilities for capturing and parsing zerolog output.
type logCapture struct {
	buf *bytes.Buffer
}

// newLogCapture creates a new log capture buffer.
func newLogCapture() *logCapture {
	return &logCapture{buf: &bytes.Buffer{}}
}

// logger returns a zerolog.Logger that writes to the capture buffer.
func (lc *logCapture) logger() zerolog.Logger {
	return zerolog.New(lc.buf).With().Timestamp().Logger()
}

// entries parses all log entries from the buffer.
func (lc *logCapture) entries() []map[string]interface{} {
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

// lastEntry returns the last log entry.
func (lc *logCapture) lastEntry() map[string]interface{} {
	entries := lc.entries()
	if len(entries) == 0 {
		return nil
	}
	return entries[len(entries)-1]
}

// contains checks if the log output contains a specific string.
func (lc *logCapture) contains(s string) bool {
	return bytes.Contains(lc.buf.Bytes(), []byte(s))
}

// createTestReadyEvent creates a discordgo.Ready event for testing.
func createTestReadyEvent(username string, guilds []*discordgo.Guild) *discordgo.Ready {
	return &discordgo.Ready{
		User: &discordgo.User{
			ID:       "bot-user-id",
			Username: username,
		},
		Guilds: guilds,
	}
}

// createTestGuilds creates a slice of test guilds.
func createTestGuilds(count int) []*discordgo.Guild {
	guilds := make([]*discordgo.Guild, count)
	for i := 0; i < count; i++ {
		guilds[i] = &discordgo.Guild{
			ID:   "guild-" + string(rune('0'+i)),
			Name: "Test Guild " + string(rune('0'+i)),
		}
	}
	return guilds
}

func Test_NewReadyHandler(t *testing.T) {
	tests := []struct {
		name   string
		logger zerolog.Logger
	}{
		{
			name:   "create handler with logger",
			logger: zerolog.Nop(),
		},
		{
			name:   "create handler with disabled logger",
			logger: zerolog.New(nil).Level(zerolog.Disabled),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := handler.NewReadyHandler(tt.logger)

			require.NotNil(t, h, "NewReadyHandler should return non-nil *ReadyHandler")
		})
	}
}

func Test_ReadyHandler_Handle_WithGuilds(t *testing.T) {
	tests := []struct {
		name           string
		username       string
		guildCount     int
		expectedFields map[string]interface{}
	}{
		{
			name:       "ready event with 5 guilds",
			username:   "JamesBot",
			guildCount: 5,
			expectedFields: map[string]interface{}{
				"username":    "JamesBot",
				"guild_count": float64(5), // JSON numbers are float64
			},
		},
		{
			name:       "ready event with 0 guilds",
			username:   "TestBot",
			guildCount: 0,
			expectedFields: map[string]interface{}{
				"username":    "TestBot",
				"guild_count": float64(0),
			},
		},
		{
			name:       "ready event with 1 guild",
			username:   "SingleGuildBot",
			guildCount: 1,
			expectedFields: map[string]interface{}{
				"username":    "SingleGuildBot",
				"guild_count": float64(1),
			},
		},
		{
			name:       "ready event with many guilds",
			username:   "PopularBot",
			guildCount: 100,
			expectedFields: map[string]interface{}{
				"username":    "PopularBot",
				"guild_count": float64(100),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capture := newLogCapture()
			h := handler.NewReadyHandler(capture.logger())

			guilds := createTestGuilds(tt.guildCount)
			ready := createTestReadyEvent(tt.username, guilds)

			// Execute the handler
			h.Handle(nil, ready)

			// Verify log output
			entry := capture.lastEntry()
			require.NotNil(t, entry, "should have logged an entry")

			// Check expected fields
			for key, expectedValue := range tt.expectedFields {
				actualValue, exists := entry[key]
				assert.True(t, exists, "log entry should contain field %q", key)
				assert.Equal(t, expectedValue, actualValue,
					"log field %q should have expected value", key)
			}

			// Check message indicates ready state
			if msg, ok := entry["message"].(string); ok {
				assert.True(t, capture.contains("ready") || capture.contains("Ready"),
					"log message should indicate bot is ready, got: %s", msg)
			}
		})
	}
}

func Test_ReadyHandler_Handle_LogsUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
	}{
		{
			name:     "simple username",
			username: "JamesBot",
		},
		{
			name:     "username with numbers",
			username: "Bot123",
		},
		{
			name:     "username with special characters",
			username: "Test_Bot-Dev",
		},
		{
			name:     "empty username",
			username: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capture := newLogCapture()
			h := handler.NewReadyHandler(capture.logger())

			ready := createTestReadyEvent(tt.username, createTestGuilds(1))
			h.Handle(nil, ready)

			entry := capture.lastEntry()
			require.NotNil(t, entry, "should have logged an entry")

			username, exists := entry["username"]
			assert.True(t, exists, "log should contain username field")
			assert.Equal(t, tt.username, username, "logged username should match")
		})
	}
}

func Test_ReadyHandler_Handle_LogsGuildCount(t *testing.T) {
	tests := []struct {
		name               string
		guildCount         int
		expectedGuildCount float64
	}{
		{
			name:               "zero guilds",
			guildCount:         0,
			expectedGuildCount: 0,
		},
		{
			name:               "one guild",
			guildCount:         1,
			expectedGuildCount: 1,
		},
		{
			name:               "five guilds",
			guildCount:         5,
			expectedGuildCount: 5,
		},
		{
			name:               "large number of guilds",
			guildCount:         1000,
			expectedGuildCount: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capture := newLogCapture()
			h := handler.NewReadyHandler(capture.logger())

			ready := createTestReadyEvent("TestBot", createTestGuilds(tt.guildCount))
			h.Handle(nil, ready)

			entry := capture.lastEntry()
			require.NotNil(t, entry, "should have logged an entry")

			guildCount, exists := entry["guild_count"]
			assert.True(t, exists, "log should contain guild_count field")
			assert.Equal(t, tt.expectedGuildCount, guildCount,
				"logged guild_count should match")
		})
	}
}

func Test_ReadyHandler_Handle_NilReady(t *testing.T) {
	capture := newLogCapture()
	h := handler.NewReadyHandler(capture.logger())

	// Should not panic with nil ready event
	assert.NotPanics(t, func() {
		h.Handle(nil, nil)
	}, "Handle should not panic with nil ready event")
}

func Test_ReadyHandler_Handle_NilUser(t *testing.T) {
	capture := newLogCapture()
	h := handler.NewReadyHandler(capture.logger())

	ready := &discordgo.Ready{
		User:   nil,
		Guilds: createTestGuilds(3),
	}

	// Should not panic with nil user
	assert.NotPanics(t, func() {
		h.Handle(nil, ready)
	}, "Handle should not panic with nil user in ready event")
}

func Test_ReadyHandler_Handle_NilGuilds(t *testing.T) {
	capture := newLogCapture()
	h := handler.NewReadyHandler(capture.logger())

	ready := &discordgo.Ready{
		User: &discordgo.User{
			ID:       "bot-id",
			Username: "TestBot",
		},
		Guilds: nil,
	}

	// Should not panic with nil guilds
	assert.NotPanics(t, func() {
		h.Handle(nil, ready)
	}, "Handle should not panic with nil guilds in ready event")

	// Should log 0 guild count
	entry := capture.lastEntry()
	if entry != nil {
		guildCount, exists := entry["guild_count"]
		if exists {
			assert.Equal(t, float64(0), guildCount,
				"nil guilds should be logged as 0")
		}
	}
}

func Test_ReadyHandler_Handle_SessionNotUsed(t *testing.T) {
	// The ready handler should work regardless of session state
	// since it only logs the ready event data
	capture := newLogCapture()
	h := handler.NewReadyHandler(capture.logger())

	ready := createTestReadyEvent("TestBot", createTestGuilds(2))

	// With nil session
	assert.NotPanics(t, func() {
		h.Handle(nil, ready)
	}, "Handle should work with nil session")

	assert.True(t, capture.contains("TestBot"),
		"should log even with nil session")
}

func Test_ReadyHandler_Handle_LogLevel(t *testing.T) {
	capture := newLogCapture()
	h := handler.NewReadyHandler(capture.logger())

	ready := createTestReadyEvent("TestBot", createTestGuilds(1))
	h.Handle(nil, ready)

	entry := capture.lastEntry()
	require.NotNil(t, entry, "should have logged an entry")

	// Ready event should typically be logged at info level
	level, exists := entry["level"]
	if exists {
		assert.Equal(t, "info", level,
			"ready event should be logged at info level")
	}
}

// Benchmark tests

func Benchmark_ReadyHandler_Handle(b *testing.B) {
	logger := zerolog.Nop()
	h := handler.NewReadyHandler(logger)
	ready := createTestReadyEvent("BenchBot", createTestGuilds(10))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Handle(nil, ready)
	}
}

func Benchmark_ReadyHandler_Handle_ManyGuilds(b *testing.B) {
	logger := zerolog.Nop()
	h := handler.NewReadyHandler(logger)
	ready := createTestReadyEvent("BenchBot", createTestGuilds(1000))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Handle(nil, ready)
	}
}
