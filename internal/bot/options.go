// Package bot provides the core bot implementation for JamesBot.
package bot

import "jamesbot/internal/middleware"

// Option is a functional option for configuring the Bot.
// Functional options allow for flexible and extensible bot configuration
// while maintaining a clean API.
type Option func(*Bot)

// WithMiddleware adds middleware to the bot's command execution chain.
// Middleware will be applied in the order provided, with earlier middleware
// wrapping later middleware and the command handler.
//
// Example:
//
//	bot, err := bot.New(cfg, logger,
//	    bot.WithMiddleware(
//	        middleware.Recovery(logger),
//	        middleware.Logging(logger),
//	    ),
//	)
func WithMiddleware(mw ...middleware.Middleware) Option {
	return func(b *Bot) {
		b.middlewares = append(b.middlewares, mw...)
	}
}
