// Package jamesprial provides an example plugin for JamesBot.
// This plugin demonstrates the plugin system with a greeting command.
package jamesprial

import (
	"github.com/rs/zerolog"

	"jamesbot/internal/command"
	"jamesbot/internal/plugin"
)

const (
	PluginName    = "jamesprial"
	PluginVersion = "1.0.0"
)

// Plugin implements the JamesPrial example plugin.
type Plugin struct {
	logger zerolog.Logger
}

// New creates a new JamesPrial plugin instance.
func New() *Plugin {
	return &Plugin{}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return PluginName
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Example plugin demonstrating the JamesBot plugin system"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return PluginVersion
}

// Init initializes the plugin.
func (p *Plugin) Init(ctx *plugin.InitContext) error {
	p.logger = ctx.Logger
	p.logger.Info().Msg("JamesPrial plugin initialized")
	return nil
}

// Commands returns the commands provided by this plugin.
func (p *Plugin) Commands() []command.Command {
	return []command.Command{
		&GreetCommand{},
		&AboutPluginCommand{},
	}
}

// Shutdown cleans up plugin resources.
func (p *Plugin) Shutdown() error {
	p.logger.Info().Msg("JamesPrial plugin shutting down")
	return nil
}

// Ensure Plugin implements the required interfaces.
var (
	_ plugin.Plugin          = (*Plugin)(nil)
	_ plugin.CommandProvider = (*Plugin)(nil)
	_ plugin.Initializable   = (*Plugin)(nil)
	_ plugin.Shutdownable    = (*Plugin)(nil)
)
