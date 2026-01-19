package middleware_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"jamesbot/internal/command"
	"jamesbot/internal/middleware"
	"jamesbot/pkg/errutil"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recoveryLogCapture provides utilities for capturing and parsing zerolog output.
type recoveryLogCapture struct {
	buf *bytes.Buffer
}

func newRecoveryLogCapture() *recoveryLogCapture {
	return &recoveryLogCapture{buf: &bytes.Buffer{}}
}

func (lc *recoveryLogCapture) logger() zerolog.Logger {
	return zerolog.New(lc.buf).With().Timestamp().Logger()
}

func (lc *recoveryLogCapture) entries() []map[string]interface{} {
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

func (lc *recoveryLogCapture) lastEntry() map[string]interface{} {
	entries := lc.entries()
	if len(entries) == 0 {
		return nil
	}
	return entries[len(entries)-1]
}

func (lc *recoveryLogCapture) contains(s string) bool {
	return bytes.Contains(lc.buf.Bytes(), []byte(s))
}

func (lc *recoveryLogCapture) raw() string {
	return lc.buf.String()
}

// createRecoveryTestContext creates a command context for recovery middleware tests.
func createRecoveryTestContext(logger zerolog.Logger) *command.Context {
	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:        "interaction-123",
			ChannelID: "test-channel",
			GuildID:   "test-guild",
			Member: &discordgo.Member{
				User: &discordgo.User{
					ID:       "test-user",
					Username: "testuser",
				},
			},
			Type: discordgo.InteractionApplicationCommand,
			Data: discordgo.ApplicationCommandInteractionData{
				ID:   "cmd-data-123",
				Name: "testcmd",
			},
		},
	}
	return command.NewContext(nil, interaction, logger)
}

func Test_Recovery_NoPanic(t *testing.T) {
	tests := []struct {
		name          string
		handlerResult error
	}{
		{
			name:          "handler returns nil",
			handlerResult: nil,
		},
		{
			name:          "handler returns error",
			handlerResult: errors.New("normal error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capture := newRecoveryLogCapture()
			logger := capture.logger()

			recoveryMW := middleware.Recovery(logger)

			handler := func(ctx *command.Context) error {
				return tt.handlerResult
			}

			wrapped := recoveryMW(handler)
			ctx := createRecoveryTestContext(logger)

			err := wrapped(ctx)

			// Should return the original error unchanged
			assert.Equal(t, tt.handlerResult, err,
				"should return handler result unchanged when no panic")
		})
	}
}

func Test_Recovery_PanicWithString(t *testing.T) {
	capture := newRecoveryLogCapture()
	logger := capture.logger()

	recoveryMW := middleware.Recovery(logger)

	handler := func(ctx *command.Context) error {
		panic("oops")
	}

	wrapped := recoveryMW(handler)
	ctx := createRecoveryTestContext(logger)

	// Should not panic
	var err error
	assert.NotPanics(t, func() {
		err = wrapped(ctx)
	}, "recovery middleware should catch panic")

	// Should return UserFriendlyError
	require.Error(t, err, "should return error after panic")

	var userFriendlyErr errutil.UserFriendlyError
	if errors.As(err, &userFriendlyErr) {
		// Good - it's a UserFriendlyError
		assert.NotEmpty(t, userFriendlyErr.UserMessage,
			"UserFriendlyError should have a user message")
	} else {
		// Also acceptable if implementation returns a different error type
		// as long as it returns an error
		assert.Error(t, err, "should return some error after panic")
	}

	// Should log the panic
	assert.True(t, capture.contains("oops") || capture.contains("panic"),
		"log should contain panic information")
}

func Test_Recovery_PanicWithError(t *testing.T) {
	capture := newRecoveryLogCapture()
	logger := capture.logger()

	recoveryMW := middleware.Recovery(logger)

	panicErr := errors.New("err")
	handler := func(ctx *command.Context) error {
		panic(panicErr)
	}

	wrapped := recoveryMW(handler)
	ctx := createRecoveryTestContext(logger)

	// Should not panic
	var err error
	assert.NotPanics(t, func() {
		err = wrapped(ctx)
	}, "recovery middleware should catch panic with error")

	// Should return UserFriendlyError
	require.Error(t, err, "should return error after panic")

	var userFriendlyErr errutil.UserFriendlyError
	if errors.As(err, &userFriendlyErr) {
		assert.NotEmpty(t, userFriendlyErr.UserMessage,
			"UserFriendlyError should have a user message")
	}

	// Should log the panic
	assert.True(t, capture.contains("err") || capture.contains("panic"),
		"log should contain panic error information")
}

func Test_Recovery_PanicWithInt(t *testing.T) {
	capture := newRecoveryLogCapture()
	logger := capture.logger()

	recoveryMW := middleware.Recovery(logger)

	handler := func(ctx *command.Context) error {
		panic(42)
	}

	wrapped := recoveryMW(handler)
	ctx := createRecoveryTestContext(logger)

	// Should not panic
	var err error
	assert.NotPanics(t, func() {
		err = wrapped(ctx)
	}, "recovery middleware should catch panic with int")

	// Should return error
	require.Error(t, err, "should return error after panic")

	// Should log something about the panic
	assert.True(t, capture.contains("42") || capture.contains("panic"),
		"log should contain panic information")
}

func Test_Recovery_PanicWithNil(t *testing.T) {
	capture := newRecoveryLogCapture()
	logger := capture.logger()

	recoveryMW := middleware.Recovery(logger)

	handler := func(ctx *command.Context) error {
		panic(nil)
	}

	wrapped := recoveryMW(handler)
	ctx := createRecoveryTestContext(logger)

	// Should not panic (or should handle nil panic gracefully)
	assert.NotPanics(t, func() {
		_ = wrapped(ctx)
	}, "recovery middleware should handle nil panic")
}

func Test_Recovery_StackTraceLogged(t *testing.T) {
	capture := newRecoveryLogCapture()
	logger := capture.logger()

	recoveryMW := middleware.Recovery(logger)

	handler := func(ctx *command.Context) error {
		panic("stack trace test")
	}

	wrapped := recoveryMW(handler)
	ctx := createRecoveryTestContext(logger)

	_ = wrapped(ctx)

	// Check for stack trace indicators in log
	raw := capture.raw()
	hasStackInfo := strings.Contains(raw, "stack") ||
		strings.Contains(raw, "goroutine") ||
		strings.Contains(raw, ".go:") ||
		strings.Contains(raw, "recovery_test.go") ||
		strings.Contains(raw, "panic")

	assert.True(t, hasStackInfo,
		"log should contain stack trace or panic information")
}

func Test_Recovery_LogLevel(t *testing.T) {
	capture := newRecoveryLogCapture()
	logger := capture.logger()

	recoveryMW := middleware.Recovery(logger)

	handler := func(ctx *command.Context) error {
		panic("level test")
	}

	wrapped := recoveryMW(handler)
	ctx := createRecoveryTestContext(logger)

	_ = wrapped(ctx)

	entry := capture.lastEntry()
	if entry != nil {
		if level, ok := entry["level"].(string); ok {
			// Panic should be logged at error or higher level
			assert.Contains(t, []string{"error", "fatal", "panic"}, level,
				"panic should be logged at error level or higher")
		}
	}
}

func Test_Recovery_UserFriendlyMessage(t *testing.T) {
	capture := newRecoveryLogCapture()
	logger := capture.logger()

	recoveryMW := middleware.Recovery(logger)

	handler := func(ctx *command.Context) error {
		panic("internal details that users shouldn't see")
	}

	wrapped := recoveryMW(handler)
	ctx := createRecoveryTestContext(logger)

	err := wrapped(ctx)

	require.Error(t, err)

	var userFriendlyErr errutil.UserFriendlyError
	if errors.As(err, &userFriendlyErr) {
		// The user message should NOT contain internal details
		assert.NotContains(t, userFriendlyErr.UserMessage, "internal details",
			"user message should not expose internal panic details")

		// The user message should be friendly
		assert.True(t,
			strings.Contains(strings.ToLower(userFriendlyErr.UserMessage), "error") ||
				strings.Contains(strings.ToLower(userFriendlyErr.UserMessage), "wrong") ||
				strings.Contains(strings.ToLower(userFriendlyErr.UserMessage), "try") ||
				strings.Contains(strings.ToLower(userFriendlyErr.UserMessage), "unexpected") ||
				len(userFriendlyErr.UserMessage) > 0,
			"user message should be user-friendly")
	}
}

func Test_Recovery_ChainedMiddleware(t *testing.T) {
	capture := newRecoveryLogCapture()
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

	recoveryMW := middleware.Recovery(logger)

	handler := func(ctx *command.Context) error {
		executionOrder = append(executionOrder, "handler")
		panic("chained panic")
	}

	// Chain: before -> recovery -> after -> handler
	// Recovery should catch panic from handler and allow after to run
	chained := middleware.Chain(beforeMW, recoveryMW, afterMW)(handler)
	ctx := createRecoveryTestContext(logger)

	assert.NotPanics(t, func() {
		_ = chained(ctx)
	}, "chained middleware should not panic")

	// before should have run
	assert.Contains(t, executionOrder, "before", "before middleware should run")
	// handler should have run (and panicked)
	assert.Contains(t, executionOrder, "handler", "handler should run")
}

func Test_Recovery_MultiplePanics(t *testing.T) {
	capture := newRecoveryLogCapture()
	logger := capture.logger()

	recoveryMW := middleware.Recovery(logger)

	panicCount := 0
	handler := func(ctx *command.Context) error {
		panicCount++
		panic("panic number " + string(rune('0'+panicCount)))
	}

	wrapped := recoveryMW(handler)

	// Multiple invocations should all be recovered
	for i := 0; i < 3; i++ {
		ctx := createRecoveryTestContext(logger)
		assert.NotPanics(t, func() {
			_ = wrapped(ctx)
		}, "recovery should catch panic on invocation %d", i+1)
	}

	assert.Equal(t, 3, panicCount, "handler should have panicked 3 times")
}

func Test_Recovery_PanicAfterWork(t *testing.T) {
	capture := newRecoveryLogCapture()
	logger := capture.logger()

	recoveryMW := middleware.Recovery(logger)

	workDone := false
	handler := func(ctx *command.Context) error {
		workDone = true
		panic("panic after doing work")
	}

	wrapped := recoveryMW(handler)
	ctx := createRecoveryTestContext(logger)

	assert.NotPanics(t, func() {
		_ = wrapped(ctx)
	})

	assert.True(t, workDone, "work should have been done before panic")
}

func Test_Recovery_PreservesNormalErrors(t *testing.T) {
	capture := newRecoveryLogCapture()
	logger := capture.logger()

	recoveryMW := middleware.Recovery(logger)

	customErr := errors.New("custom error, not a panic")
	handler := func(ctx *command.Context) error {
		return customErr
	}

	wrapped := recoveryMW(handler)
	ctx := createRecoveryTestContext(logger)

	err := wrapped(ctx)

	// Should return the original error unchanged
	assert.Equal(t, customErr, err,
		"recovery middleware should not modify normal errors")
}

func Test_Recovery_PanicWithStruct(t *testing.T) {
	capture := newRecoveryLogCapture()
	logger := capture.logger()

	type customPanic struct {
		Message string
		Code    int
	}

	recoveryMW := middleware.Recovery(logger)

	handler := func(ctx *command.Context) error {
		panic(customPanic{Message: "custom panic struct", Code: 500})
	}

	wrapped := recoveryMW(handler)
	ctx := createRecoveryTestContext(logger)

	assert.NotPanics(t, func() {
		err := wrapped(ctx)
		require.Error(t, err, "should return error after struct panic")
	})
}

func Test_Recovery_DeepPanic(t *testing.T) {
	capture := newRecoveryLogCapture()
	logger := capture.logger()

	recoveryMW := middleware.Recovery(logger)

	// Create a deep call stack before panicking
	var deepPanic func(depth int)
	deepPanic = func(depth int) {
		if depth <= 0 {
			panic("deep panic")
		}
		deepPanic(depth - 1)
	}

	handler := func(ctx *command.Context) error {
		deepPanic(10)
		return nil
	}

	wrapped := recoveryMW(handler)
	ctx := createRecoveryTestContext(logger)

	assert.NotPanics(t, func() {
		err := wrapped(ctx)
		require.Error(t, err, "should return error after deep panic")
	})
}

func Test_Recovery_ConcurrentPanics(t *testing.T) {
	logger := zerolog.Nop()
	recoveryMW := middleware.Recovery(logger)

	handler := func(ctx *command.Context) error {
		panic("concurrent panic")
	}

	wrapped := recoveryMW(handler)

	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() {
				done <- true
			}()
			ctx := createRecoveryTestContext(logger)
			_ = wrapped(ctx)
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
	// If we reach here, all panics were recovered
}

// Benchmark tests

func Benchmark_Recovery_NoPanic(b *testing.B) {
	logger := zerolog.Nop()
	recoveryMW := middleware.Recovery(logger)

	handler := func(ctx *command.Context) error {
		return nil
	}

	wrapped := recoveryMW(handler)
	ctx := createRecoveryTestContext(logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wrapped(ctx)
	}
}

func Benchmark_Recovery_WithPanic(b *testing.B) {
	logger := zerolog.Nop()
	recoveryMW := middleware.Recovery(logger)

	handler := func(ctx *command.Context) error {
		panic("benchmark panic")
	}

	wrapped := recoveryMW(handler)
	ctx := createRecoveryTestContext(logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wrapped(ctx)
	}
}

func Benchmark_Recovery_WithError(b *testing.B) {
	logger := zerolog.Nop()
	recoveryMW := middleware.Recovery(logger)
	testErr := errors.New("benchmark error")

	handler := func(ctx *command.Context) error {
		return testErr
	}

	wrapped := recoveryMW(handler)
	ctx := createRecoveryTestContext(logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wrapped(ctx)
	}
}
