package plugin

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"

	"jamesbot/internal/command"
	"jamesbot/internal/middleware"
)

// Loader handles plugin initialization and integration with the bot.
type Loader struct {
	registry *Registry
	logger   zerolog.Logger
	session  *discordgo.Session
	configs  map[string]map[string]interface{}
}

// NewLoader creates a new plugin loader.
func NewLoader(registry *Registry, logger zerolog.Logger) *Loader {
	return &Loader{
		registry: registry,
		logger:   logger.With().Str("component", "plugin-loader").Logger(),
		configs:  make(map[string]map[string]interface{}),
	}
}

// SetSession sets the Discord session for plugins that need it during init.
func (l *Loader) SetSession(session *discordgo.Session) {
	l.session = session
}

// SetPluginConfig sets configuration for a specific plugin.
func (l *Loader) SetPluginConfig(pluginName string, config map[string]interface{}) {
	l.configs[pluginName] = config
}

// Load registers and initializes a plugin.
// If the plugin implements Initializable, its Init method is called.
func (l *Loader) Load(p Plugin) error {
	// Register with registry
	if err := l.registry.Register(p); err != nil {
		return fmt.Errorf("failed to register plugin %q: %w", p.Name(), err)
	}

	// Initialize if applicable
	if init, ok := p.(Initializable); ok {
		ctx := &InitContext{
			Logger:  l.logger.With().Str("plugin", p.Name()).Logger(),
			Session: l.session,
			Config:  l.configs[p.Name()],
		}

		if err := init.Init(ctx); err != nil {
			// Unregister on init failure
			_ = l.registry.Unregister(p.Name())
			return fmt.Errorf("failed to initialize plugin %q: %w", p.Name(), err)
		}

		l.logger.Debug().
			Str("plugin", p.Name()).
			Msg("plugin initialized")
	}

	return nil
}

// LoadAll loads multiple plugins. Returns the first error encountered.
func (l *Loader) LoadAll(plugins ...Plugin) error {
	for _, p := range plugins {
		if err := l.Load(p); err != nil {
			return err
		}
	}
	return nil
}

// Commands collects all commands from loaded CommandProvider plugins.
func (l *Loader) Commands() []command.Command {
	var commands []command.Command

	for _, p := range l.registry.All() {
		if cp, ok := p.(CommandProvider); ok {
			commands = append(commands, cp.Commands()...)
		}
	}

	return commands
}

// Middleware collects all middleware from loaded MiddlewareProvider plugins.
// Middleware is returned in plugin registration order.
func (l *Loader) Middleware() []middleware.Middleware {
	var mws []middleware.Middleware

	for _, p := range l.registry.All() {
		if mp, ok := p.(MiddlewareProvider); ok {
			mws = append(mws, mp.Middleware()...)
		}
	}

	return mws
}

// EventHandlers collects all event handlers from loaded EventHandlerProvider plugins.
func (l *Loader) EventHandlers() []interface{} {
	var handlers []interface{}

	for _, p := range l.registry.All() {
		if eh, ok := p.(EventHandlerProvider); ok {
			handlers = append(handlers, eh.EventHandlers()...)
		}
	}

	return handlers
}

// RegisterEventHandlers registers all plugin event handlers with the Discord session.
func (l *Loader) RegisterEventHandlers(session *discordgo.Session) {
	for _, h := range l.EventHandlers() {
		session.AddHandler(h)
	}
}

// ShutdownAll calls Shutdown on all plugins that implement Shutdownable.
// Errors are logged but do not stop other plugins from shutting down.
func (l *Loader) ShutdownAll() {
	for _, p := range l.registry.All() {
		if s, ok := p.(Shutdownable); ok {
			if err := s.Shutdown(); err != nil {
				l.logger.Error().
					Err(err).
					Str("plugin", p.Name()).
					Msg("plugin shutdown error")
			} else {
				l.logger.Debug().
					Str("plugin", p.Name()).
					Msg("plugin shutdown complete")
			}
		}
	}
}

// Info returns information about all loaded plugins.
func (l *Loader) Info() []Info {
	return l.registry.Info()
}
