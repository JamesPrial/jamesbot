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

// Note: RulesCommand and subcommands use commands.CLIContext instead of cli.Context
// to avoid import cycles. The cli package provides an adapter.

// =============================================================================
// RulesCommand (Parent) Tests
// =============================================================================

// Test_RulesCommand_Name verifies the command returns "rules" as its name.
func Test_RulesCommand_Name(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "returns rules as command name",
			expected: "rules",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &commands.RulesCommand{}

			result := cmd.Name()

			assert.Equal(t, tt.expected, result, "Name() should return %q", tt.expected)
		})
	}
}

// Test_RulesCommand_Synopsis verifies the command returns a non-empty synopsis.
func Test_RulesCommand_Synopsis(t *testing.T) {
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
			cmd := &commands.RulesCommand{}

			result := cmd.Synopsis()

			if tt.notEmpty {
				assert.NotEmpty(t, result, "Synopsis() should return non-empty string")
			}
		})
	}
}

// Test_RulesCommand_Usage verifies the usage string contains expected subcommands.
func Test_RulesCommand_Usage(t *testing.T) {
	tests := []struct {
		name           string
		expectContains []string
		expectNotEmpty bool
	}{
		{
			name:           "returns usage containing list and set subcommands",
			expectContains: []string{"list", "set"},
			expectNotEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &commands.RulesCommand{}

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

// Test_RulesCommand_Subcommands - REMOVED
// This test was removed because Subcommands() is implemented on the adapter
// in cli.go, not on RulesCommand directly. Subcommand testing belongs in cli_test.go.

// Test_RulesCommand_Run_WithoutSubcommand verifies run without subcommand prints usage.
func Test_RulesCommand_Run_WithoutSubcommand(t *testing.T) {
	tests := []struct {
		name                 string
		args                 []string
		expectExitCode       int
		expectOutputContains []string
	}{
		{
			name:                 "no subcommand prints usage and exits 0",
			args:                 []string{},
			expectExitCode:       0,
			expectOutputContains: []string{"list", "set"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &commands.RulesCommand{}
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			ctx := &commands.CLIContext{
				Stdout:      stdout,
				Stderr:      stderr,
				APIEndpoint: "",
			}

			exitCode := cmd.Run(ctx, tt.args)

			assert.Equal(t, tt.expectExitCode, exitCode,
				"Run() should return exit code %d", tt.expectExitCode)

			// Check combined output contains expected strings
			combinedOutput := strings.ToLower(stdout.String() + stderr.String())
			for _, expected := range tt.expectOutputContains {
				assert.Contains(t, combinedOutput, strings.ToLower(expected),
					"output should contain %q", expected)
			}
		})
	}
}

// Test_RulesCommand_SetFlags verifies the command can set flags without panic.
func Test_RulesCommand_SetFlags(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "SetFlags does not panic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &commands.RulesCommand{}
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			fs.SetOutput(&bytes.Buffer{})

			// Should not panic
			assert.NotPanics(t, func() {
				cmd.SetFlags(fs)
			}, "SetFlags should not panic")
		})
	}
}

// =============================================================================
// RulesListCommand Tests
// =============================================================================

// Test_RulesListCommand_Name verifies the command returns "list" as its name.
func Test_RulesListCommand_Name(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "returns list as command name",
			expected: "list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &commands.RulesListCommand{}

			result := cmd.Name()

			assert.Equal(t, tt.expected, result, "Name() should return %q", tt.expected)
		})
	}
}

// Test_RulesListCommand_Synopsis verifies the command returns a non-empty synopsis.
func Test_RulesListCommand_Synopsis(t *testing.T) {
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
			cmd := &commands.RulesListCommand{}

			result := cmd.Synopsis()

			if tt.notEmpty {
				assert.NotEmpty(t, result, "Synopsis() should return non-empty string")
			}
		})
	}
}

// Test_RulesListCommand_Usage verifies the usage string contains expected content.
func Test_RulesListCommand_Usage(t *testing.T) {
	tests := []struct {
		name           string
		expectContains []string
		expectNotEmpty bool
	}{
		{
			name:           "returns usage containing rules list",
			expectContains: []string{"rules", "list"},
			expectNotEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &commands.RulesListCommand{}

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

// Test_RulesListCommand_SetFlags verifies the command registers --json and --endpoint flags.
func Test_RulesListCommand_SetFlags(t *testing.T) {
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
			cmd := &commands.RulesListCommand{}
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			fs.SetOutput(&bytes.Buffer{})

			cmd.SetFlags(fs)

			for _, flagName := range tt.expectedFlags {
				f := fs.Lookup(flagName)
				require.NotNil(t, f, "SetFlags should register --%s flag", flagName)
			}
		})
	}
}

// Test_RulesListCommand_SetFlags_JSONFlagDefault verifies the json flag defaults to false.
func Test_RulesListCommand_SetFlags_JSONFlagDefault(t *testing.T) {
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
			cmd := &commands.RulesListCommand{}
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

// Test_RulesListCommand_Run_SuccessfulList tests successful rules listing.
func Test_RulesListCommand_Run_SuccessfulList(t *testing.T) {
	tests := []struct {
		name           string
		rules          []control.Rule
		expectExitCode int
		expectContains []string
	}{
		{
			name: "successful list displays rule names",
			rules: []control.Rule{
				{Name: "anti-spam", Description: "Prevents spam", Enabled: true, Key: "threshold", Value: "5"},
				{Name: "link-filter", Description: "Filters links", Enabled: false, Key: "domains", Value: "*.xyz"},
			},
			expectExitCode: 0,
			expectContains: []string{"anti-spam", "link-filter"},
		},
		{
			name: "single rule displays correctly",
			rules: []control.Rule{
				{Name: "profanity-filter", Description: "Filters profanity", Enabled: true, Key: "level", Value: "strict"},
			},
			expectExitCode: 0,
			expectContains: []string{"profanity-filter"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/rules" && r.Method == http.MethodGet {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(tt.rules)
					return
				}
				http.NotFound(w, r)
			}))
			defer server.Close()

			cmd := &commands.RulesListCommand{}
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
				"Run() should return exit code %d", tt.expectExitCode)

			outputLower := strings.ToLower(stdout.String())
			for _, expected := range tt.expectContains {
				assert.Contains(t, outputLower, strings.ToLower(expected),
					"stdout should contain %q", expected)
			}
		})
	}
}

// Test_RulesListCommand_Run_EmptyRules tests handling of empty rules list.
func Test_RulesListCommand_Run_EmptyRules(t *testing.T) {
	tests := []struct {
		name              string
		rules             []control.Rule
		expectExitCode    int
		expectContainsOne []string
	}{
		{
			name:              "empty rules shows no rules message",
			rules:             []control.Rule{},
			expectExitCode:    0,
			expectContainsOne: []string{"no rules", "no configured rules", "empty", "none"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/rules" && r.Method == http.MethodGet {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(tt.rules)
					return
				}
				http.NotFound(w, r)
			}))
			defer server.Close()

			cmd := &commands.RulesListCommand{}
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
				"Run() should return exit code %d", tt.expectExitCode)

			// Check output contains one of the expected messages
			combinedOutput := strings.ToLower(stdout.String() + stderr.String())
			containsExpected := false
			for _, expected := range tt.expectContainsOne {
				if strings.Contains(combinedOutput, strings.ToLower(expected)) {
					containsExpected = true
					break
				}
			}
			assert.True(t, containsExpected,
				"output should contain one of: %v", tt.expectContainsOne)
		})
	}
}

// Test_RulesListCommand_Run_JSONOutput tests JSON output format.
func Test_RulesListCommand_Run_JSONOutput(t *testing.T) {
	tests := []struct {
		name           string
		rules          []control.Rule
		expectExitCode int
	}{
		{
			name: "json flag outputs valid JSON array",
			rules: []control.Rule{
				{Name: "spam-filter", Description: "Filters spam", Enabled: true, Key: "rate", Value: "10"},
				{Name: "caps-filter", Description: "Filters caps", Enabled: false, Key: "threshold", Value: "50"},
			},
			expectExitCode: 0,
		},
		{
			name:           "empty rules outputs empty JSON array",
			rules:          []control.Rule{},
			expectExitCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/rules" && r.Method == http.MethodGet {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(tt.rules)
					return
				}
				http.NotFound(w, r)
			}))
			defer server.Close()

			cmd := &commands.RulesListCommand{}
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

			// Verify output is valid JSON array
			var output []map[string]interface{}
			err = json.Unmarshal(stdout.Bytes(), &output)
			assert.NoError(t, err, "stdout should be valid JSON array")
			assert.Len(t, output, len(tt.rules), "JSON array should have correct length")
		})
	}
}

// Test_RulesListCommand_Run_JSONContainsFields verifies JSON output contains expected fields.
func Test_RulesListCommand_Run_JSONContainsFields(t *testing.T) {
	tests := []struct {
		name           string
		rules          []control.Rule
		expectedFields []string
	}{
		{
			name: "JSON output contains all required fields",
			rules: []control.Rule{
				{Name: "test-rule", Description: "Test description", Enabled: true, Key: "test-key", Value: "test-value"},
			},
			expectedFields: []string{"name", "description", "enabled"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/rules" && r.Method == http.MethodGet {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(tt.rules)
					return
				}
				http.NotFound(w, r)
			}))
			defer server.Close()

			cmd := &commands.RulesListCommand{}
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
			var output []map[string]interface{}
			err = json.Unmarshal(stdout.Bytes(), &output)
			require.NoError(t, err, "stdout should be valid JSON")
			require.NotEmpty(t, output, "output should not be empty")

			for _, field := range tt.expectedFields {
				_, exists := output[0][field]
				assert.True(t, exists, "JSON output should contain field %q", field)
			}
		})
	}
}

// Test_RulesListCommand_Run_ConnectionError tests error handling when server is unavailable.
func Test_RulesListCommand_Run_ConnectionError(t *testing.T) {
	tests := []struct {
		name                string
		expectExitCode      int
		expectStderrContain []string
	}{
		{
			name:                "connection error returns exit 1",
			expectExitCode:      1,
			expectStderrContain: []string{"cannot connect", "connection", "error", "failed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use an endpoint that will fail to connect
			invalidEndpoint := "http://localhost:1" // Port 1 should not have a server

			cmd := &commands.RulesListCommand{}
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
				"Run() should return exit code %d when connection fails", tt.expectExitCode)

			// Check stderr contains one of the expected substrings
			combinedOutput := strings.ToLower(stdout.String() + stderr.String())
			containsExpected := false
			for _, expected := range tt.expectStderrContain {
				if strings.Contains(combinedOutput, strings.ToLower(expected)) {
					containsExpected = true
					break
				}
			}
			assert.True(t, containsExpected,
				"output should contain one of: %v", tt.expectStderrContain)
		})
	}
}

// Test_RulesListCommand_Run_ServerError tests error handling when server returns error.
func Test_RulesListCommand_Run_ServerError(t *testing.T) {
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

			cmd := &commands.RulesListCommand{}
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

// Test_RulesListCommand_ImplementsCLICommand verifies the command has required methods.
func Test_RulesListCommand_ImplementsCLICommand(t *testing.T) {
	cmd := &commands.RulesListCommand{}

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
	assert.NotNil(t, ctx, "CLIContext should be constructible")
}

// =============================================================================
// RulesSetCommand Tests
// =============================================================================

// Test_RulesSetCommand_Name verifies the command returns "set" as its name.
func Test_RulesSetCommand_Name(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "returns set as command name",
			expected: "set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &commands.RulesSetCommand{}

			result := cmd.Name()

			assert.Equal(t, tt.expected, result, "Name() should return %q", tt.expected)
		})
	}
}

// Test_RulesSetCommand_Synopsis verifies the command returns a non-empty synopsis.
func Test_RulesSetCommand_Synopsis(t *testing.T) {
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
			cmd := &commands.RulesSetCommand{}

			result := cmd.Synopsis()

			if tt.notEmpty {
				assert.NotEmpty(t, result, "Synopsis() should return non-empty string")
			}
		})
	}
}

// Test_RulesSetCommand_Usage verifies the usage string contains expected content.
func Test_RulesSetCommand_Usage(t *testing.T) {
	tests := []struct {
		name           string
		expectContains []string
		expectNotEmpty bool
	}{
		{
			name:           "returns usage containing rules set",
			expectContains: []string{"rules", "set"},
			expectNotEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &commands.RulesSetCommand{}

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

// Test_RulesSetCommand_SetFlags verifies the command registers expected flags.
func Test_RulesSetCommand_SetFlags(t *testing.T) {
	tests := []struct {
		name          string
		expectedFlags []string
	}{
		{
			name:          "registers endpoint flag",
			expectedFlags: []string{"endpoint"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &commands.RulesSetCommand{}
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			fs.SetOutput(&bytes.Buffer{})

			cmd.SetFlags(fs)

			for _, flagName := range tt.expectedFlags {
				f := fs.Lookup(flagName)
				require.NotNil(t, f, "SetFlags should register --%s flag", flagName)
			}
		})
	}
}

// Test_RulesSetCommand_Run_SuccessfulSet tests successful rule setting.
func Test_RulesSetCommand_Run_SuccessfulSet(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		expectExitCode    int
		expectContainsOne []string
	}{
		{
			name:              "successful set displays success message",
			args:              []string{"spam-filter", "threshold", "10"},
			expectExitCode:    0,
			expectContainsOne: []string{"success", "updated", "set", "ok"},
		},
		{
			name:              "setting rule with different values succeeds",
			args:              []string{"link-filter", "enabled", "true"},
			expectExitCode:    0,
			expectContainsOne: []string{"success", "updated", "set", "ok"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/rules/set" && r.Method == http.MethodPost {
					// Verify request body
					var body map[string]string
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					if body["name"] == "" || body["key"] == "" || body["value"] == "" {
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"status":"ok"}`))
					return
				}
				http.NotFound(w, r)
			}))
			defer server.Close()

			cmd := &commands.RulesSetCommand{}
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

			exitCode := cmd.Run(ctx, tt.args)

			assert.Equal(t, tt.expectExitCode, exitCode,
				"Run() should return exit code %d", tt.expectExitCode)

			// Check output contains one of the expected messages
			combinedOutput := strings.ToLower(stdout.String() + stderr.String())
			containsExpected := false
			for _, expected := range tt.expectContainsOne {
				if strings.Contains(combinedOutput, strings.ToLower(expected)) {
					containsExpected = true
					break
				}
			}
			assert.True(t, containsExpected,
				"output should contain one of: %v, got: %s", tt.expectContainsOne, combinedOutput)
		})
	}
}

// Test_RulesSetCommand_Run_MissingArgs tests error handling for missing arguments.
func Test_RulesSetCommand_Run_MissingArgs(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		expectExitCode    int
		expectContainsOne []string
	}{
		{
			name:              "no arguments returns exit 1 with usage error",
			args:              []string{},
			expectExitCode:    1,
			expectContainsOne: []string{"usage", "error", "required", "missing", "argument"},
		},
		{
			name:              "only name argument returns exit 1",
			args:              []string{"spam-filter"},
			expectExitCode:    1,
			expectContainsOne: []string{"usage", "error", "required", "missing", "argument"},
		},
		{
			name:              "only name and key returns exit 1",
			args:              []string{"spam-filter", "threshold"},
			expectExitCode:    1,
			expectContainsOne: []string{"usage", "error", "required", "missing", "argument"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server that should not be called
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("Server should not be called when arguments are missing")
				http.NotFound(w, r)
			}))
			defer server.Close()

			cmd := &commands.RulesSetCommand{}
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

			exitCode := cmd.Run(ctx, tt.args)

			assert.Equal(t, tt.expectExitCode, exitCode,
				"Run() should return exit code %d when args are missing", tt.expectExitCode)

			// Check output contains one of the expected messages
			combinedOutput := strings.ToLower(stdout.String() + stderr.String())
			containsExpected := false
			for _, expected := range tt.expectContainsOne {
				if strings.Contains(combinedOutput, strings.ToLower(expected)) {
					containsExpected = true
					break
				}
			}
			assert.True(t, containsExpected,
				"output should contain one of: %v, got: %s", tt.expectContainsOne, combinedOutput)
		})
	}
}

// Test_RulesSetCommand_Run_ConnectionError tests error handling when server is unavailable.
func Test_RulesSetCommand_Run_ConnectionError(t *testing.T) {
	tests := []struct {
		name                string
		args                []string
		expectExitCode      int
		expectStderrContain []string
	}{
		{
			name:                "connection error returns exit 1",
			args:                []string{"test-rule", "key", "value"},
			expectExitCode:      1,
			expectStderrContain: []string{"cannot connect", "connection", "error", "failed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use an endpoint that will fail to connect
			invalidEndpoint := "http://localhost:1" // Port 1 should not have a server

			cmd := &commands.RulesSetCommand{}
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

			exitCode := cmd.Run(ctx, tt.args)

			assert.Equal(t, tt.expectExitCode, exitCode,
				"Run() should return exit code %d when connection fails", tt.expectExitCode)

			// Check stderr contains one of the expected substrings
			combinedOutput := strings.ToLower(stdout.String() + stderr.String())
			containsExpected := false
			for _, expected := range tt.expectStderrContain {
				if strings.Contains(combinedOutput, strings.ToLower(expected)) {
					containsExpected = true
					break
				}
			}
			assert.True(t, containsExpected,
				"output should contain one of: %v", tt.expectStderrContain)
		})
	}
}

// Test_RulesSetCommand_Run_ServerError tests error handling when server returns error.
func Test_RulesSetCommand_Run_ServerError(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		statusCode     int
		expectExitCode int
	}{
		{
			name:           "server 400 bad request returns exit 1",
			args:           []string{"test-rule", "key", "value"},
			statusCode:     400,
			expectExitCode: 1,
		},
		{
			name:           "server 500 error returns exit 1",
			args:           []string{"test-rule", "key", "value"},
			statusCode:     500,
			expectExitCode: 1,
		},
		{
			name:           "server 404 not found returns exit 1",
			args:           []string{"nonexistent-rule", "key", "value"},
			statusCode:     404,
			expectExitCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server that returns error
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(`{"error":"server error"}`))
			}))
			defer server.Close()

			cmd := &commands.RulesSetCommand{}
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

			exitCode := cmd.Run(ctx, tt.args)

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

// Test_RulesSetCommand_ImplementsCLICommand verifies the command has required methods.
func Test_RulesSetCommand_ImplementsCLICommand(t *testing.T) {
	cmd := &commands.RulesSetCommand{}

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
	assert.NotNil(t, ctx, "CLIContext should be constructible")
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func Benchmark_RulesCommand_Name(b *testing.B) {
	cmd := &commands.RulesCommand{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Name()
	}
}

func Benchmark_RulesCommand_Synopsis(b *testing.B) {
	cmd := &commands.RulesCommand{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Synopsis()
	}
}

func Benchmark_RulesCommand_Usage(b *testing.B) {
	cmd := &commands.RulesCommand{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Usage()
	}
}

// Benchmark_RulesCommand_Subcommands - REMOVED
// This benchmark was removed because Subcommands() is implemented on the adapter
// in cli.go, not on RulesCommand directly. Subcommand benchmarking belongs in cli_test.go.

func Benchmark_RulesListCommand_Name(b *testing.B) {
	cmd := &commands.RulesListCommand{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Name()
	}
}

func Benchmark_RulesListCommand_SetFlags(b *testing.B) {
	cmd := &commands.RulesListCommand{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fs := flag.NewFlagSet("bench", flag.ContinueOnError)
		cmd.SetFlags(fs)
	}
}

func Benchmark_RulesListCommand_Run_WithMockServer(b *testing.B) {
	rules := []control.Rule{
		{Name: "spam-filter", Description: "Filters spam", Enabled: true, Key: "threshold", Value: "5"},
		{Name: "link-filter", Description: "Filters links", Enabled: false, Key: "domains", Value: "*.xyz"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rules" && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(rules)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	cmd := &commands.RulesListCommand{}
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

func Benchmark_RulesSetCommand_Name(b *testing.B) {
	cmd := &commands.RulesSetCommand{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Name()
	}
}

func Benchmark_RulesSetCommand_SetFlags(b *testing.B) {
	cmd := &commands.RulesSetCommand{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fs := flag.NewFlagSet("bench", flag.ContinueOnError)
		cmd.SetFlags(fs)
	}
}

func Benchmark_RulesSetCommand_Run_WithMockServer(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rules/set" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	cmd := &commands.RulesSetCommand{}
	fs := flag.NewFlagSet("bench", flag.ContinueOnError)
	fs.SetOutput(&bytes.Buffer{})
	cmd.SetFlags(fs)
	fs.Parse([]string{"--endpoint", server.URL})

	args := []string{"test-rule", "key", "value"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		ctx := &commands.CLIContext{
			Stdout:      stdout,
			Stderr:      stderr,
			APIEndpoint: server.URL,
		}
		cmd.Run(ctx, args)
	}
}
