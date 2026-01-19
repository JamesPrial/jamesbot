package config

import (
	"fmt"
	"strings"
	"time"

	"jamesbot/pkg/errutil"

	"github.com/spf13/viper"
)

// Load reads and validates configuration from the specified file path.
// It returns a Config struct or an error if loading or validation fails.
//
// The loader supports the following features:
//   - Default values for all configuration options
//   - Environment variable overrides with JAMESBOT_ prefix
//   - Validation of required fields
//
// Environment variables use the pattern JAMESBOT_<SECTION>_<KEY>,
// for example: JAMESBOT_DISCORD_TOKEN, JAMESBOT_LOGGING_LEVEL
func Load(path string) (*Config, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Configure environment variable binding
	v.SetEnvPrefix("JAMESBOT")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Explicitly bind environment variables for keys that may not exist in config file
	v.BindEnv("discord.token", "JAMESBOT_DISCORD_TOKEN")
	v.BindEnv("logging.level", "JAMESBOT_LOGGING_LEVEL")
	v.BindEnv("shutdown.timeout", "JAMESBOT_SHUTDOWN_TIMEOUT")

	// Load configuration file if path is provided
	if path != "" {
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Unmarshal configuration into struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate required fields
	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// setDefaults configures default values for all configuration options.
func setDefaults(v *viper.Viper) {
	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "console")

	// Shutdown defaults
	v.SetDefault("shutdown.timeout", 10*time.Second)

	// Discord defaults
	v.SetDefault("discord.cleanup_on_shutdown", false)
}

// validate checks that all required configuration fields are present and valid.
func validate(cfg *Config) error {
	// Validate Discord token is not empty
	if cfg.Discord.Token == "" {
		return &errutil.ConfigError{
			Key:     "discord.token",
			Message: "token is required but not provided",
		}
	}

	return nil
}
