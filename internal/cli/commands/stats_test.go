package commands_test

import (
	"bytes"
	"encoding/json"
	"flag"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"jamesbot/internal/cli/commands"
	"jamesbot/internal/control"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: StatsCommand uses commands.CLIContext instead of cli.Context
// to avoid import cycles. The cli package provides an adapter.

// Test_StatsCommand_Name verifies the command returns "stats" as its name.
func Test_StatsCommand_Name(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "returns stats as command name",
			expected: "stats",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &commands.StatsCommand{}

			result := cmd.Name()

			assert.Equal(t, tt.expected, result, "Name() should return %q", tt.expected)
		})
	}
}

// Test_StatsCommand_Synopsis verifies the command returns a non-empty synopsis.
func Test_StatsCommand_Synopsis(t *testing.T) {
	tests := []struct {
		name     string
		notEmpty bool
	}{
		{
			name:     "returns non-empty synopsis",
			notEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &commands.StatsCommand{}

			result := cmd.Synopsis()

			if tt.notEmpty {
				assert.NotEmpty(t, result, "Synopsis() should return non-empty string")
			}
		})
	}
}

// Test_StatsCommand_Usage verifies the usage string contains "stats".
func Test_StatsCommand_Usage(t *testing.T) {
	tests := []struct {
		name           string
		expectContains []string
		expectNotEmpty bool
	}{
		{
			name:           "returns usage containing stats",
			expectContains: []string{"stats"},
			expectNotEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &commands.StatsCommand{}

			result := cmd.Usage()

			if tt.expectNotEmpty {
				assert.NotEmpty(t, result, "Usage() should return non-empty string")
			}

			resultLower := strings.ToLower(result)
			for _, expected := range tt.expectContains {
				assert.Contains(t, resultLower, strings.ToLower(expected),
					"Usage() should contain %q", expected)
			}
		})
	}
}

// Test_StatsCommand_SetFlags verifies the command registers --json and --endpoint flags.
func Test_StatsCommand_SetFlags(t *testing.T) {
	tests := []struct {
		name          string
		expectedFlags []string
	}{
		{
			name:          "registers json and endpoint flags",
			expectedFlags: []string{"json", "endpoint"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &commands.StatsCommand{}
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			fs.SetOutput(&bytes.Buffer{}) // Suppress flag output

			cmd.SetFlags(fs)

			for _, flagName := range tt.expectedFlags {
				f := fs.Lookup(flagName)
				require.NotNil(t, f, "SetFlags should register --%s flag", flagName)
			}
		})
	}
}

// Test_StatsCommand_SetFlags_JSONFlagDefault verifies the json flag defaults to false.
func Test_StatsCommand_SetFlags_JSONFlagDefault(t *testing.T) {
	tests := []struct {
		name          string
		parseArgs     []string
		expectedValue string
	}{
		{
			name:          "json flag defaults to false",
			parseArgs:     []string{},
			expectedValue: "false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &commands.StatsCommand{}
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			fs.SetOutput(&bytes.Buffer{})

			cmd.SetFlags(fs)

			err := fs.Parse(tt.parseArgs)
			require.NoError(t, err, "Flag parsing should succeed")

			jsonFlag := fs.Lookup("json")
			require.NotNil(t, jsonFlag)
			assert.Equal(t, tt.expectedValue, jsonFlag.Value.String(),
				"json flag should default to %q", tt.expectedValue)
		})
	}
}

// Test_StatsCommand_SetFlags_EndpointFlagDefault verifies the endpoint flag default.
func Test_StatsCommand_SetFlags_EndpointFlagDefault(t *testing.T) {
	tests := []struct {
		name             string
		parseArgs        []string
		expectedContains string
	}{
		{
			name:             "endpoint flag has a default value",
			parseArgs:        []string{},
			expectedContains: "127.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &commands.StatsCommand{}
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			fs.SetOutput(&bytes.Buffer{})

			cmd.SetFlags(fs)

			err := fs.Parse(tt.parseArgs)
			require.NoError(t, err, "Flag parsing should succeed")

			endpointFlag := fs.Lookup("endpoint")
			require.NotNil(t, endpointFlag)
			assert.Contains(t, endpointFlag.Value.String(), tt.expectedContains,
				"endpoint flag should contain %q", tt.expectedContains)
		})
	}
}

// Test_StatsCommand_SetFlags_JSONFlagCustomValue verifies custom json flag value.
func Test_StatsCommand_SetFlags_JSONFlagCustomValue(t *testing.T) {
	tests := []struct {
		name          string
		parseArgs     []string
		expectedValue string
	}{
		{
			name:          "json flag accepts true value",
			parseArgs:     []string{"--json"},
			expectedValue: "true",
		},
		{
			name:          "json flag with explicit true",
			parseArgs:     []string{"--json=true"},
			expectedValue: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &commands.StatsCommand{}
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			fs.SetOutput(&bytes.Buffer{})

			cmd.SetFlags(fs)

			err := fs.Parse(tt.parseArgs)
			require.NoError(t, err, "Flag parsing should succeed")

			jsonFlag := fs.Lookup("json")
			require.NotNil(t, jsonFlag)
			assert.Equal(t, tt.expectedValue, jsonFlag.Value.String(),
				"json flag should be set to %q", tt.expectedValue)
		})
	}
}

// Test_StatsCommand_SetFlags_EndpointFlagCustomValue verifies custom endpoint flag value.
func Test_StatsCommand_SetFlags_EndpointFlagCustomValue(t *testing.T) {
	tests := []struct {
		name          string
		parseArgs     []string
		expectedValue string
	}{
		{
			name:          "endpoint flag accepts custom value",
			parseArgs:     []string{"--endpoint", "http://custom:9999"},
			expectedValue: "http://custom:9999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &commands.StatsCommand{}
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			fs.SetOutput(&bytes.Buffer{})

			cmd.SetFlags(fs)

			err := fs.Parse(tt.parseArgs)
			require.NoError(t, err, "Flag parsing should succeed")

			endpointFlag := fs.Lookup("endpoint")
			require.NotNil(t, endpointFlag)
			assert.Equal(t, tt.expectedValue, endpointFlag.Value.String(),
				"endpoint flag should be set to %q", tt.expectedValue)
		})
	}
}

// Test_StatsCommand_Run_HumanReadable tests human-readable output scenarios.
func Test_StatsCommand_Run_HumanReadable(t *testing.T) {
	tests := []struct {
		name           string
		stats          control.Stats
		expectExitCode int
		expectContains []string
	}{
		{
			name: "successful stats displays human readable output",
			stats: control.Stats{
				Uptime:           "1h30m0s",
				StartTime:        1704067200,
				CommandsExecuted: 42,
				GuildCount:       5,
				ActiveRules:      3,
			},
			expectExitCode: 0,
			expectContains: []string{"uptime", "commands", "guilds"},
		},
		{
			name: "displays all stat fields",
			stats: control.Stats{
				Uptime:           "2h0m0s",
				StartTime:        1704067200,
				CommandsExecuted: 100,
				GuildCount:       10,
				ActiveRules:      7,
			},
			expectExitCode: 0,
			expectContains: []string{"uptime", "commands", "guilds"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/stats" {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(tt.stats)
					return
				}
				http.NotFound(w, r)
			}))
			defer server.Close()

			cmd := &commands.StatsCommand{}
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			fs.SetOutput(stderr)

			cmd.SetFlags(fs)

			// Parse with the mock server endpoint
			err := fs.Parse([]string{"--endpoint", server.URL})
			require.NoError(t, err, "Flag parsing should succeed")

			ctx := &commands.CLIContext{
				Stdout:      stdout,
				Stderr:      stderr,
				APIEndpoint: server.URL,
			}

			exitCode := cmd.Run(ctx, fs.Args())

			assert.Equal(t, tt.expectExitCode, exitCode,
				"Run() should return exit code %d", tt.expectExitCode)

			outputLower := strings.ToLower(stdout.String())
			for _, expected := range tt.expectContains {
				assert.Contains(t, outputLower, strings.ToLower(expected),
					"stdout should contain %q", expected)
			}
		})
	}
}

// Test_StatsCommand_Run_FormatsDuration verifies duration formatting.
func Test_StatsCommand_Run_FormatsDuration(t *testing.T) {
	tests := []struct {
		name           string
		uptime         string
		expectContains string
	}{
		{
			name:           "formats duration 2h30m",
			uptime:         "2h30m0s",
			expectContains: "2h30m",
		},
		{
			name:           "formats short duration",
			uptime:         "5m30s",
			expectContains: "5m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			stats := control.Stats{
				Uptime:           tt.uptime,
				StartTime:        1704067200,
				CommandsExecuted: 10,
				GuildCount:       2,
				ActiveRules:      1,
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/stats" {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(stats)
					return
				}
				http.NotFound(w, r)
			}))
			defer server.Close()

			cmd := &commands.StatsCommand{}
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			fs.SetOutput(stderr)

			cmd.SetFlags(fs)

			err := fs.Parse([]string{"--endpoint", server.URL})
			require.NoError(t, err, "Flag parsing should succeed")

			ctx := &commands.CLIContext{
				Stdout:      stdout,
				Stderr:      stderr,
				APIEndpoint: server.URL,
			}

			exitCode := cmd.Run(ctx, fs.Args())

			assert.Equal(t, 0, exitCode, "Run() should return exit code 0")
			assert.Contains(t, stdout.String(), tt.expectContains,
				"stdout should contain formatted duration %q", tt.expectContains)
		})
	}
}

// Test_StatsCommand_Run_JSONOutput tests JSON output format.
func Test_StatsCommand_Run_JSONOutput(t *testing.T) {
	tests := []struct {
		name           string
		stats          control.Stats
		expectExitCode int
	}{
		{
			name: "json flag outputs valid JSON",
			stats: control.Stats{
				Uptime:           "1h0m0s",
				StartTime:        1704067200,
				CommandsExecuted: 50,
				GuildCount:       3,
				ActiveRules:      2,
			},
			expectExitCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/stats" {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(tt.stats)
					return
				}
				http.NotFound(w, r)
			}))
			defer server.Close()

			cmd := &commands.StatsCommand{}
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			fs.SetOutput(stderr)

			cmd.SetFlags(fs)

			err := fs.Parse([]string{"--json", "--endpoint", server.URL})
			require.NoError(t, err, "Flag parsing should succeed")

			ctx := &commands.CLIContext{
				Stdout:      stdout,
				Stderr:      stderr,
				APIEndpoint: server.URL,
			}

			exitCode := cmd.Run(ctx, fs.Args())

			assert.Equal(t, tt.expectExitCode, exitCode,
				"Run() should return exit code %d", tt.expectExitCode)

			// Verify output is valid JSON
			var output map[string]interface{}
			err = json.Unmarshal(stdout.Bytes(), &output)
			assert.NoError(t, err, "stdout should be valid JSON")
		})
	}
}

// Test_StatsCommand_Run_JSONContainsFields verifies JSON output contains expected fields.
func Test_StatsCommand_Run_JSONContainsFields(t *testing.T) {
	tests := []struct {
		name           string
		stats          control.Stats
		expectedFields []string
	}{
		{
			name: "JSON output contains all required fields",
			stats: control.Stats{
				Uptime:           "1h0m0s",
				StartTime:        1704067200,
				CommandsExecuted: 50,
				GuildCount:       3,
				ActiveRules:      2,
			},
			expectedFields: []string{"uptime", "commands_executed", "guild_count"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/stats" {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(tt.stats)
					return
				}
				http.NotFound(w, r)
			}))
			defer server.Close()

			cmd := &commands.StatsCommand{}
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			fs.SetOutput(stderr)

			cmd.SetFlags(fs)

			err := fs.Parse([]string{"--json", "--endpoint", server.URL})
			require.NoError(t, err, "Flag parsing should succeed")

			ctx := &commands.CLIContext{
				Stdout:      stdout,
				Stderr:      stderr,
				APIEndpoint: server.URL,
			}

			exitCode := cmd.Run(ctx, fs.Args())
			assert.Equal(t, 0, exitCode, "Run() should return exit code 0")

			// Parse JSON and verify fields
			var output map[string]interface{}
			err = json.Unmarshal(stdout.Bytes(), &output)
			require.NoError(t, err, "stdout should be valid JSON")

			for _, field := range tt.expectedFields {
				_, exists := output[field]
				assert.True(t, exists, "JSON output should contain field %q", field)
			}
		})
	}
}

// Test_StatsCommand_Run_BotNotRunning tests error handling when bot is not running.
func Test_StatsCommand_Run_BotNotRunning(t *testing.T) {
	tests := []struct {
		name                string
		expectExitCode      int
		expectStderrContain []string
	}{
		{
			name:                "bot not running returns error",
			expectExitCode:      1,
			expectStderrContain: []string{"cannot connect", "connection"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use an endpoint that will fail to connect
			invalidEndpoint := "http://localhost:1" // Port 1 should not have a server

			cmd := &commands.StatsCommand{}
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			fs.SetOutput(stderr)

			cmd.SetFlags(fs)

			err := fs.Parse([]string{"--endpoint", invalidEndpoint})
			require.NoError(t, err, "Flag parsing should succeed")

			ctx := &commands.CLIContext{
				Stdout:      stdout,
				Stderr:      stderr,
				APIEndpoint: invalidEndpoint,
			}

			exitCode := cmd.Run(ctx, fs.Args())

			assert.Equal(t, tt.expectExitCode, exitCode,
				"Run() should return exit code %d when bot not running", tt.expectExitCode)

			// Check stderr contains one of the expected substrings
			stderrLower := strings.ToLower(stderr.String() + stdout.String())
			containsExpected := false
			for _, expected := range tt.expectStderrContain {
				if strings.Contains(stderrLower, strings.ToLower(expected)) {
					containsExpected = true
					break
				}
			}
			assert.True(t, containsExpected,
				"output should contain one of: %v", tt.expectStderrContain)
		})
	}
}

// Test_StatsCommand_Run_ServerError tests error handling when server returns error.
func Test_StatsCommand_Run_ServerError(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		expectExitCode int
	}{
		{
			name:           "server 500 error returns exit 1",
			statusCode:     500,
			expectExitCode: 1,
		},
		{
			name:           "server 404 error returns exit 1",
			statusCode:     404,
			expectExitCode: 1,
		},
		{
			name:           "server 503 error returns exit 1",
			statusCode:     503,
			expectExitCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server that returns error
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte("Internal Server Error"))
			}))
			defer server.Close()

			cmd := &commands.StatsCommand{}
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			fs.SetOutput(stderr)

			cmd.SetFlags(fs)

			err := fs.Parse([]string{"--endpoint", server.URL})
			require.NoError(t, err, "Flag parsing should succeed")

			ctx := &commands.CLIContext{
				Stdout:      stdout,
				Stderr:      stderr,
				APIEndpoint: server.URL,
			}

			exitCode := cmd.Run(ctx, fs.Args())

			assert.Equal(t, tt.expectExitCode, exitCode,
				"Run() should return exit code %d on server error", tt.expectExitCode)

			// Verify error message is present
			combinedOutput := strings.ToLower(stdout.String() + stderr.String())
			assert.True(t,
				strings.Contains(combinedOutput, "error") ||
					strings.Contains(combinedOutput, "failed") ||
					strings.Contains(combinedOutput, "status"),
				"output should contain error information")
		})
	}
}

// Test_StatsCommand_Run_InvalidJSON tests error handling for invalid JSON response.
func Test_StatsCommand_Run_InvalidJSON(t *testing.T) {
	tests := []struct {
		name           string
		response       string
		expectExitCode int
	}{
		{
			name:           "invalid JSON response returns exit 1",
			response:       "not valid json",
			expectExitCode: 1,
		},
		{
			name:           "empty response returns exit 1",
			response:       "",
			expectExitCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server returning invalid JSON
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/stats" {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(tt.response))
					return
				}
				http.NotFound(w, r)
			}))
			defer server.Close()

			cmd := &commands.StatsCommand{}
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			fs.SetOutput(stderr)

			cmd.SetFlags(fs)

			err := fs.Parse([]string{"--endpoint", server.URL})
			require.NoError(t, err, "Flag parsing should succeed")

			ctx := &commands.CLIContext{
				Stdout:      stdout,
				Stderr:      stderr,
				APIEndpoint: server.URL,
			}

			exitCode := cmd.Run(ctx, fs.Args())

			assert.Equal(t, tt.expectExitCode, exitCode,
				"Run() should return exit code %d on invalid JSON", tt.expectExitCode)
		})
	}
}

// Test_StatsCommand_ImplementsCLICommand verifies the command has required methods.
func Test_StatsCommand_ImplementsCLICommand(t *testing.T) {
	cmd := &commands.StatsCommand{}

	// Verify all methods are callable
	assert.NotEmpty(t, cmd.Name(), "Name() should be callable")
	assert.NotEmpty(t, cmd.Synopsis(), "Synopsis() should be callable")
	assert.NotEmpty(t, cmd.Usage(), "Usage() should be callable")

	// SetFlags should be callable without panic
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cmd.SetFlags(fs)

	// Run should be callable with CLIContext
	ctx := &commands.CLIContext{
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}
	// Note: We don't actually call Run here as it would try to connect
	assert.NotNil(t, ctx, "CLIContext should be constructible")
}

// Test_StatsCommand_SetFlags_FlagDescriptions verifies flags have descriptions.
func Test_StatsCommand_SetFlags_FlagDescriptions(t *testing.T) {
	cmd := &commands.StatsCommand{}
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(&bytes.Buffer{})

	cmd.SetFlags(fs)

	// Check that flags have usage descriptions
	jsonFlag := fs.Lookup("json")
	if jsonFlag != nil {
		assert.NotEmpty(t, jsonFlag.Usage, "--json flag should have a usage description")
	}

	endpointFlag := fs.Lookup("endpoint")
	if endpointFlag != nil {
		assert.NotEmpty(t, endpointFlag.Usage, "--endpoint flag should have a usage description")
	}
}

// Benchmark tests

func Benchmark_StatsCommand_Name(b *testing.B) {
	cmd := &commands.StatsCommand{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Name()
	}
}

func Benchmark_StatsCommand_Synopsis(b *testing.B) {
	cmd := &commands.StatsCommand{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Synopsis()
	}
}

func Benchmark_StatsCommand_Usage(b *testing.B) {
	cmd := &commands.StatsCommand{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Usage()
	}
}

func Benchmark_StatsCommand_SetFlags(b *testing.B) {
	cmd := &commands.StatsCommand{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fs := flag.NewFlagSet("bench", flag.ContinueOnError)
		cmd.SetFlags(fs)
	}
}

func Benchmark_StatsCommand_Run_WithMockServer(b *testing.B) {
	stats := control.Stats{
		Uptime:           "1h0m0s",
		StartTime:        1704067200,
		CommandsExecuted: 50,
		GuildCount:       3,
		ActiveRules:      2,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/stats" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(stats)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	cmd := &commands.StatsCommand{}
	fs := flag.NewFlagSet("bench", flag.ContinueOnError)
	fs.SetOutput(&bytes.Buffer{})
	cmd.SetFlags(fs)
	fs.Parse([]string{"--endpoint", server.URL})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		ctx := &commands.CLIContext{
			Stdout:      stdout,
			Stderr:      stderr,
			APIEndpoint: server.URL,
		}
		cmd.Run(ctx, nil)
	}
}
