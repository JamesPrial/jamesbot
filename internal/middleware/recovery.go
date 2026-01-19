package middleware

import (
	"fmt"
	"runtime/debug"

	"github.com/rs/zerolog"

	"jamesbot/internal/command"
	"jamesbot/pkg/errutil"
)

// Recovery creates a middleware that recovers from panics during command execution.
// When a panic occurs, it logs the panic with a stack trace and returns a
// user-friendly error message. This prevents the bot from crashing when
// a command handler panics.
func Recovery(logger zerolog.Logger) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *command.Context) (err error) {
			// Use defer/recover to catch panics
			defer func() {
				if r := recover(); r != nil {
					// Get stack trace
					stack := debug.Stack()

					// Log the panic with stack trace
					logger.Error().
						Interface("panic", r).
						Bytes("stack", stack).
						Str("command", getCommandName(ctx)).
						Str("user_id", ctx.UserID()).
						Str("guild_id", ctx.GuildID()).
						Msg("panic recovered in command handler")

					// Return a user-friendly error
					err = errutil.UserFriendlyError{
						UserMessage: "An unexpected error occurred. The issue has been logged.",
						Err:         fmt.Errorf("panic recovered: %v", r),
					}
				}
			}()

			// Call the next handler
			return next(ctx)
		}
	}
}

// getCommandName safely extracts the command name from context.
func getCommandName(ctx *command.Context) string {
	if ctx == nil || ctx.Interaction == nil {
		return ""
	}
	return ctx.Interaction.ApplicationCommandData().Name
}
