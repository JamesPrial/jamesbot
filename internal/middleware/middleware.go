// Package middleware provides middleware support for command execution.
package middleware

import "jamesbot/internal/command"

// HandlerFunc is a function that handles command execution.
// It takes a command context and returns an error if the execution fails.
type HandlerFunc func(ctx *command.Context) error

// Middleware is a function that wraps a HandlerFunc to add additional behavior.
// Middleware can perform actions before and after the handler executes,
// modify the context, handle errors, or short-circuit execution.
type Middleware func(next HandlerFunc) HandlerFunc

// Chain combines multiple middlewares into a single middleware.
// Middlewares are executed in the order they are provided, with the first
// middleware in the slice being the outermost wrapper.
//
// Example:
//
//	combined := Chain(LoggingMiddleware, AuthMiddleware, RateLimitMiddleware)
//	handler := combined(commandHandler)
//
// This will execute in order: Logging -> Auth -> RateLimit -> commandHandler
func Chain(middlewares ...Middleware) Middleware {
	return func(final HandlerFunc) HandlerFunc {
		// Build the handler chain in reverse order
		handler := final
		for i := len(middlewares) - 1; i >= 0; i-- {
			handler = middlewares[i](handler)
		}
		return handler
	}
}
