package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"jamesbot/internal/config"
	"jamesbot/pkg/errutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a temporary config file
func createTempConfigFile(t *testing.T, content string) string {
	t.Helper()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	err := os.WriteFile(configPath, []byte(content), 0644)
	require.NoError(t, err, "failed to create temp config file")

	return configPath
}

// Helper to clear environment variables for clean test state
func clearEnvVars(t *testing.T) {
	t.Helper()

	envVars := []string{
		"JAMESBOT_DISCORD_TOKEN",
		"JAMESBOT_LOGGING_LEVEL",
		"JAMESBOT_SHUTDOWN_TIMEOUT",
	}

	for _, env := range envVars {
		if val, exists := os.LookupEnv(env); exists {
			t.Cleanup(func() {
				os.Setenv(env, val)
			})
			os.Unsetenv(env)
		}
	}
}

func Test_Load_ValidConfigFile(t *testing.T) {
	clearEnvVars(t)

	tests := []struct {
		name           string
		configContent  string
		expectedToken  string
		expectedLevel  string
		expectedPrefix string
	}{
		{
			name: "valid config with all fields",
			configContent: `
discord:
  token: "test-token-12345"
logging:
  level: "debug"
bot:
  prefix: "!"
`,
			expectedToken:  "test-token-12345",
			expectedLevel:  "debug",
			expectedPrefix: "!",
		},
		{
			name: "valid config with token only",
			configContent: `
discord:
  token: "minimal-token"
`,
			expectedToken: "minimal-token",
			expectedLevel: "info", // default value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := createTempConfigFile(t, tt.configContent)

			cfg, err := config.Load(configPath)

			require.NoError(t, err)
			require.NotNil(t, cfg)
			assert.Equal(t, tt.expectedToken, cfg.Discord.Token)
			if tt.expectedLevel != "" {
				assert.Equal(t, tt.expectedLevel, cfg.Logging.Level)
			}
		})
	}
}

func Test_Load_MissingFile(t *testing.T) {
	clearEnvVars(t)

	nonExistentPath := "/path/that/does/not/exist/config.yaml"

	cfg, err := config.Load(nonExistentPath)

	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, strings.ToLower(err.Error()), "read config",
		"error should mention 'read config'")
}

func Test_Load_MissingToken(t *testing.T) {
	clearEnvVars(t)

	tests := []struct {
		name          string
		configContent string
	}{
		{
			name: "config without discord section",
			configContent: `
logging:
  level: "info"
`,
		},
		{
			name: "config with empty token",
			configContent: `
discord:
  token: ""
`,
		},
		{
			name: "config with discord section but no token field",
			configContent: `
discord:
  guild_id: "123456"
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := createTempConfigFile(t, tt.configContent)

			cfg, err := config.Load(configPath)

			require.Error(t, err)
			assert.Nil(t, cfg)

			// Check that error is a ConfigError for discord.token
			var configErr *errutil.ConfigError
			if assert.ErrorAs(t, err, &configErr) {
				assert.Equal(t, "discord.token", configErr.Key,
					"ConfigError should be for discord.token")
			}
		})
	}
}

func Test_Load_EnvVarOverride(t *testing.T) {
	tests := []struct {
		name          string
		configContent string
		envVar        string
		envValue      string
		expectedToken string
	}{
		{
			name: "env var overrides file value",
			configContent: `
discord:
  token: "file-token"
`,
			envVar:        "JAMESBOT_DISCORD_TOKEN",
			envValue:      "env-token-override",
			expectedToken: "env-token-override",
		},
		{
			name: "env var provides token when file has none",
			configContent: `
logging:
  level: "debug"
`,
			envVar:        "JAMESBOT_DISCORD_TOKEN",
			envValue:      "env-token-only",
			expectedToken: "env-token-only",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvVars(t)
			configPath := createTempConfigFile(t, tt.configContent)

			// Set environment variable
			os.Setenv(tt.envVar, tt.envValue)
			t.Cleanup(func() {
				os.Unsetenv(tt.envVar)
			})

			cfg, err := config.Load(configPath)

			require.NoError(t, err)
			require.NotNil(t, cfg)
			assert.Equal(t, tt.expectedToken, cfg.Discord.Token)
		})
	}
}

func Test_Load_DefaultValues(t *testing.T) {
	clearEnvVars(t)

	// Minimal valid config - only required token
	configContent := `
discord:
  token: "test-token"
`
	configPath := createTempConfigFile(t, configContent)

	cfg, err := config.Load(configPath)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify default values
	assert.Equal(t, "info", cfg.Logging.Level,
		"default logging level should be 'info'")
	assert.Equal(t, 10*time.Second, cfg.Shutdown.Timeout,
		"default shutdown timeout should be 10s")
}

func Test_Load_InvalidYAML(t *testing.T) {
	clearEnvVars(t)

	tests := []struct {
		name          string
		configContent string
	}{
		{
			name: "malformed YAML with bad indentation",
			configContent: `
discord:
token: "test"
  invalid: true
`,
		},
		{
			name:          "completely invalid YAML",
			configContent: `{{{not valid yaml at all:::`,
		},
		{
			name:          "YAML with tabs instead of spaces",
			configContent: "discord:\n\ttoken: \"test\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := createTempConfigFile(t, tt.configContent)

			cfg, err := config.Load(configPath)

			require.Error(t, err)
			assert.Nil(t, cfg)
		})
	}
}

func Test_Load_EmptyConfigFileWithTokenOnly(t *testing.T) {
	clearEnvVars(t)

	// Minimal valid config - token only
	configContent := `
discord:
  token: "only-token"
`
	configPath := createTempConfigFile(t, configContent)

	cfg, err := config.Load(configPath)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Token should be set
	assert.Equal(t, "only-token", cfg.Discord.Token)

	// All defaults should be applied
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, 10*time.Second, cfg.Shutdown.Timeout)
}

func Test_Load_EnvVarTakesPrecedenceOverFileValue(t *testing.T) {
	clearEnvVars(t)

	configContent := `
discord:
  token: "file-value-token"
`
	configPath := createTempConfigFile(t, configContent)

	// Set environment variable with different value
	envToken := "environment-variable-token"
	os.Setenv("JAMESBOT_DISCORD_TOKEN", envToken)
	t.Cleanup(func() {
		os.Unsetenv("JAMESBOT_DISCORD_TOKEN")
	})

	cfg, err := config.Load(configPath)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Environment variable should take precedence
	assert.Equal(t, envToken, cfg.Discord.Token,
		"environment variable should override file value")
	assert.NotEqual(t, "file-value-token", cfg.Discord.Token,
		"file value should not be used when env var is set")
}

func Test_Load_VariousValidConfigurations(t *testing.T) {
	clearEnvVars(t)

	tests := []struct {
		name          string
		configContent string
		checkFunc     func(t *testing.T, cfg *config.Config)
	}{
		{
			name: "custom logging level",
			configContent: `
discord:
  token: "test-token"
logging:
  level: "debug"
`,
			checkFunc: func(t *testing.T, cfg *config.Config) {
				assert.Equal(t, "debug", cfg.Logging.Level)
			},
		},
		{
			name: "custom shutdown timeout",
			configContent: `
discord:
  token: "test-token"
shutdown:
  timeout: 30s
`,
			checkFunc: func(t *testing.T, cfg *config.Config) {
				assert.Equal(t, 30*time.Second, cfg.Shutdown.Timeout)
			},
		},
		{
			name: "warning logging level",
			configContent: `
discord:
  token: "test-token"
logging:
  level: "warn"
`,
			checkFunc: func(t *testing.T, cfg *config.Config) {
				assert.Equal(t, "warn", cfg.Logging.Level)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := createTempConfigFile(t, tt.configContent)

			cfg, err := config.Load(configPath)

			require.NoError(t, err)
			require.NotNil(t, cfg)
			tt.checkFunc(t, cfg)
		})
	}
}

func Test_Load_FilePermissions(t *testing.T) {
	clearEnvVars(t)

	// Create a config file with restricted permissions (unreadable)
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	content := `
discord:
  token: "test-token"
`
	err := os.WriteFile(configPath, []byte(content), 0000)
	require.NoError(t, err, "failed to create temp config file")

	// Ensure cleanup restores permissions so temp dir can be cleaned
	t.Cleanup(func() {
		os.Chmod(configPath, 0644)
	})

	cfg, loadErr := config.Load(configPath)

	// Should fail to read the file
	require.Error(t, loadErr)
	assert.Nil(t, cfg)
}

func Test_Load_EmptyFile(t *testing.T) {
	clearEnvVars(t)

	configPath := createTempConfigFile(t, "")

	cfg, err := config.Load(configPath)

	// Empty file should fail because token is required
	require.Error(t, err)
	assert.Nil(t, cfg)
}

func Test_Load_ConfigStructure(t *testing.T) {
	clearEnvVars(t)

	// Test that the config structure has expected fields accessible
	configContent := `
discord:
  token: "test-token"
  guild_id: "123456789"
logging:
  level: "info"
  format: "json"
shutdown:
  timeout: 15s
`
	configPath := createTempConfigFile(t, configContent)

	cfg, err := config.Load(configPath)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify structure is accessible
	assert.NotNil(t, cfg.Discord)
	assert.NotNil(t, cfg.Logging)
	assert.NotNil(t, cfg.Shutdown)
}
