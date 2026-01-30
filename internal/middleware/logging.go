package middleware

import (
	"time"

	"jamesbot/internal/command"

	"github.com/rs/zerolog"
)

// Logging creates a middleware that logs command executions.
// It records the command name, user ID, guild ID, execution duration,
// and any errors that occur. Successful executions are logged at Info level,
// while failures are logged at Error level.
func Logging(logger zerolog.Logger) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *command.Context) error {
			// Record start time
			start := time.Now()

			// Get command name for logging
			commandName := ""
			if ctx.Interaction != nil {
				commandName = ctx.Interaction.ApplicationCommandData().Name
			}

			// Call the next handler
			err := next(ctx)

			// Calculate duration
			duration := time.Since(start)

			// Build log event with context
			logEvent := logger.With().
				Str("command", commandName).
				Str("user_id", ctx.UserID()).
				Str("guild_id", ctx.GuildID()).
				Dur("duration", duration).
				Logger()

			// Log based on success or failure
			if err != nil {
				logEvent.Error().
					Err(err).
					Msg("command execution failed")
			} else {
				logEvent.Info().
					Msg("command executed successfully")
			}

			return err
		}
	}
}
