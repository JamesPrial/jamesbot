// Package config provides configuration management for JamesBot.
package config

import "time"

// Config represents the complete configuration for JamesBot.
type Config struct {
	Discord  DiscordConfig  `mapstructure:"discord"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Shutdown ShutdownConfig `mapstructure:"shutdown"`
}

// DiscordConfig contains Discord-specific configuration.
type DiscordConfig struct {
	// Token is the Discord bot token used for authentication.
	Token string `mapstructure:"token"`

	// GuildID is the Discord server (guild) ID where the bot operates.
	GuildID string `mapstructure:"guild_id"`

	// CleanupOnShutdown determines whether to remove registered commands on shutdown.
	CleanupOnShutdown bool `mapstructure:"cleanup_on_shutdown"`
}

// LoggingConfig contains logging configuration.
type LoggingConfig struct {
	// Level is the minimum log level (debug, info, warn, error, fatal, panic).
	Level string `mapstructure:"level"`

	// Format is the log output format (console, json).
	Format string `mapstructure:"format"`
}

// ShutdownConfig contains graceful shutdown configuration.
type ShutdownConfig struct {
	// Timeout is the maximum duration to wait for graceful shutdown.
	Timeout time.Duration `mapstructure:"timeout"`
}
