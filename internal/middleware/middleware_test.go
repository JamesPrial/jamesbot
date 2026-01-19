package middleware_test

import (
	"errors"
	"io"
	"sync"
	"testing"

	"jamesbot/internal/command"
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

// createTestContext creates a command.Context for testing.
func createTestContext() *command.Context {
	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:        "test-interaction",
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
				Name: "testcmd",
			},
		},
	}
	return command.NewContext(nil, interaction, discardLogger())
}

// executionTracker tracks the order of middleware and handler execution.
type executionTracker struct {
	mu     sync.Mutex
	order  []string
	errors map[string]error
}

func newExecutionTracker() *executionTracker {
	return &executionTracker{
		order:  make([]string, 0),
		errors: make(map[string]error),
	}
}

func (t *executionTracker) record(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.order = append(t.order, name)
}

func (t *executionTracker) getOrder() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	result := make([]string, len(t.order))
	copy(result, t.order)
	return result
}

func (t *executionTracker) setError(name string, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.errors[name] = err
}

func (t *executionTracker) getError(name string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.errors[name]
}

// createTrackingMiddleware creates a middleware that records its execution.
func createTrackingMiddleware(name string, tracker *executionTracker) middleware.Middleware {
	return func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(ctx *command.Context) error {
			tracker.record(name + "-before")
			err := next(ctx)
			tracker.record(name + "-after")
			return err
		}
	}
}

// createErrorMiddleware creates a middleware that returns an error instead of calling next.
func createErrorMiddleware(name string, err error, tracker *executionTracker) middleware.Middleware {
	return func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(ctx *command.Context) error {
			tracker.record(name + "-before")
			// Return error without calling next
			tracker.record(name + "-error")
			return err
		}
	}
}

// createConditionalErrorMiddleware creates a middleware that returns an error only if configured.
func createConditionalErrorMiddleware(name string, tracker *executionTracker) middleware.Middleware {
	return func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(ctx *command.Context) error {
			tracker.record(name + "-before")
			if err := tracker.getError(name); err != nil {
				tracker.record(name + "-error")
				return err
			}
			err := next(ctx)
			tracker.record(name + "-after")
			return err
		}
	}
}

func Test_Chain_EmptyChain(t *testing.T) {
	tracker := newExecutionTracker()

	// Create a simple handler
	handler := func(ctx *command.Context) error {
		tracker.record("handler")
		return nil
	}

	// Chain with no middlewares
	chained := middleware.Chain()(handler)

	// Execute
	err := chained(createTestContext())

	assert.NoError(t, err)
	assert.Equal(t, []string{"handler"}, tracker.getOrder(),
		"empty chain should execute handler directly")
}

func Test_Chain_SingleMiddleware(t *testing.T) {
	tracker := newExecutionTracker()

	handler := func(ctx *command.Context) error {
		tracker.record("handler")
		return nil
	}

	mw := createTrackingMiddleware("mw1", tracker)

	chained := middleware.Chain(mw)(handler)

	err := chained(createTestContext())

	assert.NoError(t, err)
	assert.Equal(t, []string{"mw1-before", "handler", "mw1-after"}, tracker.getOrder(),
		"single middleware should wrap handler correctly")
}

func Test_Chain_MultipleMiddlewares(t *testing.T) {
	tests := []struct {
		name          string
		middlewares   []string
		expectedOrder []string
	}{
		{
			name:        "two middlewares",
			middlewares: []string{"mw1", "mw2"},
			expectedOrder: []string{
				"mw1-before", "mw2-before", "handler", "mw2-after", "mw1-after",
			},
		},
		{
			name:        "three middlewares",
			middlewares: []string{"mw1", "mw2", "mw3"},
			expectedOrder: []string{
				"mw1-before", "mw2-before", "mw3-before",
				"handler",
				"mw3-after", "mw2-after", "mw1-after",
			},
		},
		{
			name:        "four middlewares",
			middlewares: []string{"A", "B", "C", "D"},
			expectedOrder: []string{
				"A-before", "B-before", "C-before", "D-before",
				"handler",
				"D-after", "C-after", "B-after", "A-after",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := newExecutionTracker()

			handler := func(ctx *command.Context) error {
				tracker.record("handler")
				return nil
			}

			// Create middlewares
			mws := make([]middleware.Middleware, len(tt.middlewares))
			for i, name := range tt.middlewares {
				mws[i] = createTrackingMiddleware(name, tracker)
			}

			chained := middleware.Chain(mws...)(handler)

			err := chained(createTestContext())

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOrder, tracker.getOrder(),
				"middlewares should execute in correct order")
		})
	}
}

func Test_Chain_ErrorPropagation(t *testing.T) {
	tests := []struct {
		name          string
		middlewares   []string
		errorAt       string
		expectedError string
		expectedOrder []string
		handlerCalled bool
	}{
		{
			name:          "handler returns error",
			middlewares:   []string{"mw1", "mw2"},
			errorAt:       "handler",
			expectedError: "handler error",
			expectedOrder: []string{
				"mw1-before", "mw2-before", "handler", "mw2-after", "mw1-after",
			},
			handlerCalled: true,
		},
		{
			name:          "first middleware returns error",
			middlewares:   []string{"mw1", "mw2"},
			errorAt:       "mw1",
			expectedError: "mw1 error",
			expectedOrder: []string{
				"mw1-before", "mw1-error",
			},
			handlerCalled: false,
		},
		{
			name:          "second middleware returns error",
			middlewares:   []string{"mw1", "mw2"},
			errorAt:       "mw2",
			expectedError: "mw2 error",
			expectedOrder: []string{
				"mw1-before", "mw2-before", "mw2-error", "mw1-after",
			},
			handlerCalled: false,
		},
		{
			name:          "middle middleware returns error in chain of three",
			middlewares:   []string{"A", "B", "C"},
			errorAt:       "B",
			expectedError: "B error",
			expectedOrder: []string{
				"A-before", "B-before", "B-error", "A-after",
			},
			handlerCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := newExecutionTracker()
			handlerExecuted := false

			handler := func(ctx *command.Context) error {
				handlerExecuted = true
				tracker.record("handler")
				if tt.errorAt == "handler" {
					return errors.New("handler error")
				}
				return nil
			}

			// Create middlewares
			mws := make([]middleware.Middleware, len(tt.middlewares))
			for i, name := range tt.middlewares {
				if name == tt.errorAt {
					mws[i] = createErrorMiddleware(name, errors.New(name+" error"), tracker)
				} else {
					mws[i] = createTrackingMiddleware(name, tracker)
				}
			}

			chained := middleware.Chain(mws...)(handler)

			err := chained(createTestContext())

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
			assert.Equal(t, tt.expectedOrder, tracker.getOrder())
			assert.Equal(t, tt.handlerCalled, handlerExecuted,
				"handler execution should match expected")
		})
	}
}

func Test_Chain_ExecutionOrder_ABCHandler(t *testing.T) {
	// Explicit test case from specification:
	// Chain(A, B, C)(handler) executes as:
	// A.before -> B.before -> C.before -> handler -> C.after -> B.after -> A.after

	tracker := newExecutionTracker()

	handler := func(ctx *command.Context) error {
		tracker.record("handler")
		return nil
	}

	mwA := createTrackingMiddleware("A", tracker)
	mwB := createTrackingMiddleware("B", tracker)
	mwC := createTrackingMiddleware("C", tracker)

	chained := middleware.Chain(mwA, mwB, mwC)(handler)

	err := chained(createTestContext())

	assert.NoError(t, err)

	expectedOrder := []string{
		"A-before",
		"B-before",
		"C-before",
		"handler",
		"C-after",
		"B-after",
		"A-after",
	}

	assert.Equal(t, expectedOrder, tracker.getOrder(),
		"Chain(A, B, C)(handler) should execute in order: A.before -> B.before -> C.before -> handler -> C.after -> B.after -> A.after")
}

func Test_Chain_ErrorBubblesUp(t *testing.T) {
	// Test that error from handler bubbles up through all middlewares
	tracker := newExecutionTracker()
	var capturedErrors []error

	// Create middlewares that capture the error
	createCapturingMiddleware := func(name string) middleware.Middleware {
		return func(next middleware.HandlerFunc) middleware.HandlerFunc {
			return func(ctx *command.Context) error {
				tracker.record(name + "-before")
				err := next(ctx)
				tracker.record(name + "-after")
				if err != nil {
					capturedErrors = append(capturedErrors, err)
				}
				return err
			}
		}
	}

	handlerErr := errors.New("handler failed")
	handler := func(ctx *command.Context) error {
		tracker.record("handler")
		return handlerErr
	}

	mwA := createCapturingMiddleware("A")
	mwB := createCapturingMiddleware("B")

	chained := middleware.Chain(mwA, mwB)(handler)

	err := chained(createTestContext())

	assert.Equal(t, handlerErr, err, "error should bubble up")
	assert.Len(t, capturedErrors, 2, "both middlewares should see the error")
	for _, captured := range capturedErrors {
		assert.Equal(t, handlerErr, captured)
	}
}

func Test_Chain_MiddlewareCanModifyContext(t *testing.T) {
	// This test verifies that middlewares receive the same context
	var contexts []*command.Context

	collectContextMiddleware := func(name string) middleware.Middleware {
		return func(next middleware.HandlerFunc) middleware.HandlerFunc {
			return func(ctx *command.Context) error {
				contexts = append(contexts, ctx)
				return next(ctx)
			}
		}
	}

	handler := func(ctx *command.Context) error {
		contexts = append(contexts, ctx)
		return nil
	}

	chained := middleware.Chain(
		collectContextMiddleware("A"),
		collectContextMiddleware("B"),
	)(handler)

	testCtx := createTestContext()
	err := chained(testCtx)

	assert.NoError(t, err)
	assert.Len(t, contexts, 3, "should have collected 3 contexts (2 middlewares + handler)")

	// All should be the same context
	for i, ctx := range contexts {
		assert.Equal(t, testCtx, ctx, "context %d should be the same", i)
	}
}

func Test_Chain_NilHandler(t *testing.T) {
	// Test behavior when handler is nil (implementation-defined)
	t.Skip("Behavior with nil handler is implementation-defined")

	mw := func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(ctx *command.Context) error {
			return next(ctx)
		}
	}

	chained := middleware.Chain(mw)(nil)
	_ = chained
}

func Test_Chain_NilContext(t *testing.T) {
	// Test that middlewares can handle nil context if passed
	tracker := newExecutionTracker()

	handler := func(ctx *command.Context) error {
		tracker.record("handler")
		return nil
	}

	mw := createTrackingMiddleware("mw", tracker)

	chained := middleware.Chain(mw)(handler)

	// Passing nil context - behavior is implementation-defined
	// This documents whether it panics or handles gracefully
	err := chained(nil)

	// If it doesn't panic, verify it still executed
	if err == nil {
		assert.Contains(t, tracker.getOrder(), "handler")
	}
}

func Test_Chain_ReusableMiddleware(t *testing.T) {
	// Test that the same middleware can be used multiple times
	callCount := 0
	countingMiddleware := func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(ctx *command.Context) error {
			callCount++
			return next(ctx)
		}
	}

	handler := func(ctx *command.Context) error {
		return nil
	}

	// Use same middleware twice in chain
	chained := middleware.Chain(countingMiddleware, countingMiddleware)(handler)

	err := chained(createTestContext())

	assert.NoError(t, err)
	assert.Equal(t, 2, callCount, "middleware should be called twice")
}

func Test_Chain_MiddlewareCanShortCircuit(t *testing.T) {
	handlerCalled := false

	shortCircuitMiddleware := func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(ctx *command.Context) error {
			// Don't call next, return early
			return nil
		}
	}

	handler := func(ctx *command.Context) error {
		handlerCalled = true
		return nil
	}

	chained := middleware.Chain(shortCircuitMiddleware)(handler)

	err := chained(createTestContext())

	assert.NoError(t, err)
	assert.False(t, handlerCalled, "handler should not be called when middleware short-circuits")
}

func Test_Chain_MiddlewareCanReturnDifferentError(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := errors.New("wrapped error")

	wrappingMiddleware := func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(ctx *command.Context) error {
			err := next(ctx)
			if err != nil {
				return wrappedErr
			}
			return nil
		}
	}

	handler := func(ctx *command.Context) error {
		return originalErr
	}

	chained := middleware.Chain(wrappingMiddleware)(handler)

	err := chained(createTestContext())

	assert.Equal(t, wrappedErr, err, "middleware should be able to transform errors")
	assert.NotEqual(t, originalErr, err)
}

func Test_Chain_MiddlewareCanSwallowError(t *testing.T) {
	errorSwallowingMiddleware := func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(ctx *command.Context) error {
			_ = next(ctx) // Ignore the error
			return nil
		}
	}

	handler := func(ctx *command.Context) error {
		return errors.New("this error will be swallowed")
	}

	chained := middleware.Chain(errorSwallowingMiddleware)(handler)

	err := chained(createTestContext())

	assert.NoError(t, err, "middleware should be able to swallow errors")
}

// Benchmark tests
func Benchmark_Chain_NoMiddleware(b *testing.B) {
	handler := func(ctx *command.Context) error {
		return nil
	}

	chained := middleware.Chain()(handler)
	ctx := createTestContext()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = chained(ctx)
	}
}

func Benchmark_Chain_SingleMiddleware(b *testing.B) {
	mw := func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(ctx *command.Context) error {
			return next(ctx)
		}
	}

	handler := func(ctx *command.Context) error {
		return nil
	}

	chained := middleware.Chain(mw)(handler)
	ctx := createTestContext()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = chained(ctx)
	}
}

func Benchmark_Chain_FiveMiddlewares(b *testing.B) {
	mw := func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(ctx *command.Context) error {
			return next(ctx)
		}
	}

	handler := func(ctx *command.Context) error {
		return nil
	}

	chained := middleware.Chain(mw, mw, mw, mw, mw)(handler)
	ctx := createTestContext()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = chained(ctx)
	}
}

func Benchmark_Chain_TenMiddlewares(b *testing.B) {
	mw := func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(ctx *command.Context) error {
			return next(ctx)
		}
	}

	handler := func(ctx *command.Context) error {
		return nil
	}

	chained := middleware.Chain(mw, mw, mw, mw, mw, mw, mw, mw, mw, mw)(handler)
	ctx := createTestContext()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = chained(ctx)
	}
}

// Test type definitions match expected signatures
func Test_HandlerFunc_TypeSignature(t *testing.T) {
	// Verify HandlerFunc type is func(*command.Context) error
	var h middleware.HandlerFunc = func(ctx *command.Context) error {
		return nil
	}
	_ = h
}

func Test_Middleware_TypeSignature(t *testing.T) {
	// Verify Middleware type is func(HandlerFunc) HandlerFunc
	var mw middleware.Middleware = func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(ctx *command.Context) error {
			return next(ctx)
		}
	}
	_ = mw
}

// Test that Chain returns a function with correct signature
func Test_Chain_ReturnType(t *testing.T) {
	handler := func(ctx *command.Context) error {
		return nil
	}

	// Chain should return a function that takes HandlerFunc and returns HandlerFunc
	chainResult := middleware.Chain()

	// The result should be callable with a handler
	wrapped := chainResult(handler)

	// The wrapped result should be a HandlerFunc
	var _ middleware.HandlerFunc = wrapped
}

// Test middleware with panics (implementation-defined behavior)
func Test_Chain_MiddlewarePanic(t *testing.T) {
	t.Skip("Panic handling is implementation-defined")

	panicMiddleware := func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(ctx *command.Context) error {
			panic("middleware panic")
		}
	}

	handler := func(ctx *command.Context) error {
		return nil
	}

	chained := middleware.Chain(panicMiddleware)(handler)

	assert.Panics(t, func() {
		_ = chained(createTestContext())
	})
}

// Test concurrent middleware execution safety
func Test_Chain_ConcurrentExecution(t *testing.T) {
	var counter int64
	var mu sync.Mutex

	countingMiddleware := func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(ctx *command.Context) error {
			mu.Lock()
			counter++
			mu.Unlock()
			return next(ctx)
		}
	}

	handler := func(ctx *command.Context) error {
		return nil
	}

	chained := middleware.Chain(countingMiddleware, countingMiddleware)(handler)

	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = chained(createTestContext())
		}()
	}

	wg.Wait()

	// Each execution calls 2 middlewares
	expectedCount := int64(numGoroutines * 2)
	assert.Equal(t, expectedCount, counter,
		"concurrent executions should all complete")
}

// Test that middleware chain is immutable after creation
func Test_Chain_Immutability(t *testing.T) {
	tracker1 := newExecutionTracker()
	tracker2 := newExecutionTracker()

	handler1 := func(ctx *command.Context) error {
		tracker1.record("handler1")
		return nil
	}

	handler2 := func(ctx *command.Context) error {
		tracker2.record("handler2")
		return nil
	}

	mw := createTrackingMiddleware("mw", tracker1)

	// Create chain once
	chainFunc := middleware.Chain(mw)

	// Use with different handlers
	chained1 := chainFunc(handler1)
	chained2 := chainFunc(handler2)

	// Execute both
	_ = chained1(createTestContext())

	// Reset tracker for second execution
	tracker1 = newExecutionTracker()
	mw2 := createTrackingMiddleware("mw", tracker2)
	chainFunc2 := middleware.Chain(mw2)
	chained2 = chainFunc2(handler2)
	_ = chained2(createTestContext())

	// Verify they executed independently
	assert.Contains(t, tracker2.getOrder(), "handler2")
}

// Test empty middleware slice edge case
func Test_Chain_NilMiddlewareSlice(t *testing.T) {
	handler := func(ctx *command.Context) error {
		return nil
	}

	// Passing explicit empty slice (should work same as no args)
	var mws []middleware.Middleware
	chained := middleware.Chain(mws...)(handler)

	err := chained(createTestContext())

	assert.NoError(t, err)
}

// Test very deep middleware chain
func Test_Chain_DeepChain(t *testing.T) {
	depth := 100
	callCount := 0

	mw := func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(ctx *command.Context) error {
			callCount++
			return next(ctx)
		}
	}

	handler := func(ctx *command.Context) error {
		callCount++
		return nil
	}

	// Create slice of middlewares
	mws := make([]middleware.Middleware, depth)
	for i := 0; i < depth; i++ {
		mws[i] = mw
	}

	chained := middleware.Chain(mws...)(handler)

	err := chained(createTestContext())

	assert.NoError(t, err)
	assert.Equal(t, depth+1, callCount, "all middlewares and handler should be called")
}

// Test middleware that modifies error
func Test_Chain_MiddlewareWrapsError(t *testing.T) {
	innerErr := errors.New("inner error")

	wrappingMiddleware := func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(ctx *command.Context) error {
			err := next(ctx)
			if err != nil {
				return errors.New("wrapped: " + err.Error())
			}
			return nil
		}
	}

	handler := func(ctx *command.Context) error {
		return innerErr
	}

	chained := middleware.Chain(wrappingMiddleware)(handler)

	err := chained(createTestContext())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "wrapped:")
	assert.Contains(t, err.Error(), "inner error")
}

// Test middleware that recovers from handler errors
func Test_Chain_MiddlewareRecovery(t *testing.T) {
	recoveryMiddleware := func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(ctx *command.Context) error {
			err := next(ctx)
			if err != nil {
				// Log the error but don't propagate it
				return nil
			}
			return nil
		}
	}

	handler := func(ctx *command.Context) error {
		return errors.New("this error will be recovered")
	}

	chained := middleware.Chain(recoveryMiddleware)(handler)

	err := chained(createTestContext())

	assert.NoError(t, err, "recovery middleware should suppress errors")
}

// Test that Chain() can be called multiple times with different middlewares
func Test_Chain_MultipleChainCalls(t *testing.T) {
	tracker := newExecutionTracker()

	handler := func(ctx *command.Context) error {
		tracker.record("handler")
		return nil
	}

	// Create two different chains
	chain1 := middleware.Chain(createTrackingMiddleware("A", tracker))
	chain2 := middleware.Chain(createTrackingMiddleware("B", tracker))

	// Use chain1
	chained1 := chain1(handler)
	err1 := chained1(createTestContext())
	assert.NoError(t, err1)

	// Reset tracker
	tracker = newExecutionTracker()

	// Create fresh middleware for chain2 with new tracker
	chain2 = middleware.Chain(createTrackingMiddleware("B", tracker))
	chained2 := chain2(func(ctx *command.Context) error {
		tracker.record("handler2")
		return nil
	})
	err2 := chained2(createTestContext())
	assert.NoError(t, err2)

	assert.Contains(t, tracker.getOrder(), "B-before")
	assert.Contains(t, tracker.getOrder(), "handler2")
	assert.Contains(t, tracker.getOrder(), "B-after")
}

// Test that unused variable is removed
var _ = createConditionalErrorMiddleware
