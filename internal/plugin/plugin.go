// Package plugin provides an extensible plugin system for JamesBot.
// Plugins can register commands, middleware, and event handlers.
package plugin

import (
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"

	"jamesbot/internal/command"
	"jamesbot/internal/middleware"
)

// Plugin defines the interface that all plugins must implement.
type Plugin interface {
	// Name returns the unique identifier for the plugin.
	Name() string

	// Description returns a human-readable description of what the plugin does.
	Description() string

	// Version returns the plugin version string.
	Version() string
}

// CommandProvider is implemented by plugins that provide Discord commands.
type CommandProvider interface {
	Plugin
	// Commands returns the Discord commands this plugin provides.
	Commands() []command.Command
}

// MiddlewareProvider is implemented by plugins that provide middleware.
type MiddlewareProvider interface {
	Plugin
	// Middleware returns the middleware this plugin provides.
	// Middleware is executed in the order returned.
	Middleware() []middleware.Middleware
}

// EventHandlerProvider is implemented by plugins that need to handle Discord events.
type EventHandlerProvider interface {
	Plugin
	// EventHandlers returns functions to be registered with the Discord session.
	// Each handler is registered via session.AddHandler().
	EventHandlers() []interface{}
}

// Initializable is implemented by plugins that need initialization.
type Initializable interface {
	Plugin
	// Init is called when the plugin is loaded. Return an error to prevent loading.
	Init(ctx *InitContext) error
}

// Shutdownable is implemented by plugins that need cleanup on shutdown.
type Shutdownable interface {
	Plugin
	// Shutdown is called when the bot is shutting down.
	Shutdown() error
}

// InitContext provides context for plugin initialization.
type InitContext struct {
	// Logger is a zerolog logger scoped to the plugin.
	Logger zerolog.Logger

	// Session is the Discord session (may be nil during early init).
	Session *discordgo.Session

	// Config provides plugin-specific configuration values.
	Config map[string]interface{}
}

// Info contains metadata about a loaded plugin.
type Info struct {
	Name        string
	Description string
	Version     string
	Enabled     bool
	Commands    []string
	Middleware  int
	Handlers    int
}

// InfoFromPlugin extracts Info from a Plugin instance.
func InfoFromPlugin(p Plugin) Info {
	info := Info{
		Name:        p.Name(),
		Description: p.Description(),
		Version:     p.Version(),
		Enabled:     true,
	}

	if cp, ok := p.(CommandProvider); ok {
		for _, cmd := range cp.Commands() {
			info.Commands = append(info.Commands, cmd.Name())
		}
	}

	if mp, ok := p.(MiddlewareProvider); ok {
		info.Middleware = len(mp.Middleware())
	}

	if eh, ok := p.(EventHandlerProvider); ok {
		info.Handlers = len(eh.EventHandlers())
	}

	return info
}
