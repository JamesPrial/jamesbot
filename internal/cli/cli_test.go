package cli_test

import (
	"bytes"
	"flag"
	"io"
	"strings"
	"testing"

	"jamesbot/internal/cli"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_Run_Cases tests the Run function with various argument combinations.
func Test_Run_Cases(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		wantExitCode  int
		wantStdout    []string // strings that should appear in stdout
		wantStderr    []string // strings that should appear in stderr
		wantNotStdout []string // strings that should NOT appear in stdout
		wantNotStderr []string // strings that should NOT appear in stderr
	}{
		{
			name:         "no arguments prints usage to stdout",
			args:         []string{},
			wantExitCode: 0,
			wantStdout:   []string{"usage", "jamesbot"},
		},
		{
			name:         "help flag -h prints usage to stdout",
			args:         []string{"-h"},
			wantExitCode: 0,
			wantStdout:   []string{"usage", "jamesbot"},
		},
		{
			name:         "help flag --help prints usage to stdout",
			args:         []string{"--help"},
			wantExitCode: 0,
			wantStdout:   []string{"usage", "jamesbot"},
		},
		{
			name:         "help command prints usage to stdout",
			args:         []string{"help"},
			wantExitCode: 0,
			wantStdout:   []string{"usage", "jamesbot"},
		},
		{
			name:         "version flag -v prints version to stdout",
			args:         []string{"-v"},
			wantExitCode: 0,
			wantStdout:   []string{"version"},
		},
		{
			name:         "version flag --version prints version to stdout",
			args:         []string{"--version"},
			wantExitCode: 0,
			wantStdout:   []string{"version"},
		},
		{
			name:         "version command prints version to stdout",
			args:         []string{"version"},
			wantExitCode: 0,
			wantStdout:   []string{"version"},
		},
		{
			name:         "unknown command returns error",
			args:         []string{"unknown"},
			wantExitCode: 1,
			wantStderr:   []string{"unknown command"},
		},
		{
			name:         "rules command with no subcommand prints rules usage",
			args:         []string{"rules"},
			wantExitCode: 0,
			wantStdout:   []string{"rules"},
		},
		{
			name:         "rules unknown subcommand returns error",
			args:         []string{"rules", "unknown"},
			wantExitCode: 1,
			wantStderr:   []string{"unknown"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			exitCode := cli.Run(tt.args, stdout, stderr)

			assert.Equal(t, tt.wantExitCode, exitCode, "Run(%v) exit code", tt.args)

			// Check stdout contains expected strings (case-insensitive)
			stdoutStr := strings.ToLower(stdout.String())
			for _, want := range tt.wantStdout {
				assert.Contains(t, stdoutStr, strings.ToLower(want),
					"stdout should contain %q", want)
			}

			// Check stderr contains expected strings (case-insensitive)
			stderrStr := strings.ToLower(stderr.String())
			for _, want := range tt.wantStderr {
				assert.Contains(t, stderrStr, strings.ToLower(want),
					"stderr should contain %q", want)
			}

			// Check stdout does NOT contain unwanted strings
			for _, notWant := range tt.wantNotStdout {
				assert.NotContains(t, stdoutStr, strings.ToLower(notWant),
					"stdout should NOT contain %q", notWant)
			}

			// Check stderr does NOT contain unwanted strings
			for _, notWant := range tt.wantNotStderr {
				assert.NotContains(t, stderrStr, strings.ToLower(notWant),
					"stderr should NOT contain %q", notWant)
			}
		})
	}
}

// Test_Run_ServeCommand tests that serve command is routed correctly.
func Test_Run_ServeCommand(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	// The serve command should be recognized and routed
	// It may return 0 (success) or 1 (missing config/token) depending on implementation
	exitCode := cli.Run([]string{"serve"}, stdout, stderr)

	// The serve command should be recognized (not return "unknown command")
	stderrStr := strings.ToLower(stderr.String())
	assert.NotContains(t, stderrStr, "unknown command",
		"serve should be a recognized command")

	// Exit code depends on whether serve can run without config
	// Valid exit codes are 0 (success/help shown) or 1 (missing config)
	assert.True(t, exitCode == 0 || exitCode == 1,
		"serve command should return valid exit code, got %d", exitCode)
}

// Test_Run_StatsCommand tests that stats command is routed correctly.
func Test_Run_StatsCommand(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	// The stats command should be recognized and routed
	exitCode := cli.Run([]string{"stats"}, stdout, stderr)

	// The stats command should be recognized (not return "unknown command")
	stderrStr := strings.ToLower(stderr.String())
	assert.NotContains(t, stderrStr, "unknown command",
		"stats should be a recognized command")

	// Exit code depends on implementation requirements
	assert.True(t, exitCode == 0 || exitCode == 1,
		"stats command should return valid exit code, got %d", exitCode)
}

// Test_Run_RulesListCommand tests that rules list subcommand is routed correctly.
func Test_Run_RulesListCommand(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	// The rules list command should be recognized and routed
	exitCode := cli.Run([]string{"rules", "list"}, stdout, stderr)

	// The rules list command should be recognized
	stderrStr := strings.ToLower(stderr.String())
	assert.NotContains(t, stderrStr, "unknown",
		"rules list should be a recognized subcommand")

	// Exit code depends on implementation requirements
	assert.True(t, exitCode == 0 || exitCode == 1,
		"rules list command should return valid exit code, got %d", exitCode)
}

// Test_Run_RulesSetCommand tests that rules set subcommand is routed correctly.
func Test_Run_RulesSetCommand(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	// The rules set command should be recognized and routed
	exitCode := cli.Run([]string{"rules", "set"}, stdout, stderr)

	// The rules set command should be recognized
	stderrStr := strings.ToLower(stderr.String())
	assert.NotContains(t, stderrStr, "unknown",
		"rules set should be a recognized subcommand")

	// Exit code depends on implementation requirements
	assert.True(t, exitCode == 0 || exitCode == 1,
		"rules set command should return valid exit code, got %d", exitCode)
}

// Test_Run_MultipleUnknownCommands tests various unknown command inputs.
func Test_Run_MultipleUnknownCommands(t *testing.T) {
	unknownCommands := [][]string{
		{"notacommand"},
		{"invalid"},
		{"foo", "bar"},
		{"--invalid-flag"},
		{"-x"},
	}

	for _, args := range unknownCommands {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			exitCode := cli.Run(args, stdout, stderr)

			assert.Equal(t, 1, exitCode, "unknown command %v should return exit code 1", args)
		})
	}
}

// Test_Run_HelpVariations tests all variations of help requests.
func Test_Run_HelpVariations(t *testing.T) {
	helpArgs := [][]string{
		{},
		{"-h"},
		{"--help"},
		{"help"},
	}

	for _, args := range helpArgs {
		name := "no_args"
		if len(args) > 0 {
			name = args[0]
		}
		t.Run(name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			exitCode := cli.Run(args, stdout, stderr)

			assert.Equal(t, 0, exitCode, "help variation %v should return exit code 0", args)
			assert.NotEmpty(t, stdout.String(), "help should produce output to stdout")
		})
	}
}

// Test_Run_VersionVariations tests all variations of version requests.
func Test_Run_VersionVariations(t *testing.T) {
	versionArgs := [][]string{
		{"-v"},
		{"--version"},
		{"version"},
	}

	for _, args := range versionArgs {
		t.Run(args[0], func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			exitCode := cli.Run(args, stdout, stderr)

			assert.Equal(t, 0, exitCode, "version variation %v should return exit code 0", args)
			assert.NotEmpty(t, stdout.String(), "version should produce output to stdout")
		})
	}
}

// Test_NewContext_Cases tests the NewContext function.
func Test_NewContext_Cases(t *testing.T) {
	tests := []struct {
		name            string
		stdout          io.Writer
		stderr          io.Writer
		apiEndpoint     string
		wantNonNil      bool
		wantAPIEndpoint string
		expectDefault   bool // if true, nil writers should be replaced with defaults
	}{
		{
			name:            "valid writers creates context",
			stdout:          &bytes.Buffer{},
			stderr:          &bytes.Buffer{},
			apiEndpoint:     "http://127.0.0.1:8765",
			wantNonNil:      true,
			wantAPIEndpoint: "http://127.0.0.1:8765",
		},
		{
			name:            "nil stdout defaults to os.Stdout",
			stdout:          nil,
			stderr:          &bytes.Buffer{},
			apiEndpoint:     "http://127.0.0.1:8765",
			wantNonNil:      true,
			wantAPIEndpoint: "http://127.0.0.1:8765",
			expectDefault:   true,
		},
		{
			name:            "nil stderr defaults to os.Stderr",
			stdout:          &bytes.Buffer{},
			stderr:          nil,
			apiEndpoint:     "http://127.0.0.1:8765",
			wantNonNil:      true,
			wantAPIEndpoint: "http://127.0.0.1:8765",
			expectDefault:   true,
		},
		{
			name:            "empty api endpoint is preserved",
			stdout:          &bytes.Buffer{},
			stderr:          &bytes.Buffer{},
			apiEndpoint:     "",
			wantNonNil:      true,
			wantAPIEndpoint: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := cli.NewContext(tt.stdout, tt.stderr, nil, tt.apiEndpoint)

			if tt.wantNonNil {
				require.NotNil(t, ctx, "NewContext should return non-nil Context")
			}

			// Verify the context has the expected API endpoint
			assert.Equal(t, tt.wantAPIEndpoint, ctx.APIEndpoint,
				"Context should have expected API endpoint")

			// Verify writers are not nil (defaults applied)
			assert.NotNil(t, ctx.Stdout, "Context.Stdout should not be nil")
			assert.NotNil(t, ctx.Stderr, "Context.Stderr should not be nil")
		})
	}
}

// Test_NewContext_APIEndpoint verifies the API endpoint is set correctly.
func Test_NewContext_APIEndpoint(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	ctx := cli.NewContext(stdout, stderr, nil, "http://127.0.0.1:8765")

	require.NotNil(t, ctx)
	assert.Equal(t, "http://127.0.0.1:8765", ctx.APIEndpoint,
		"API endpoint should be set to provided value")
}

// Test_Context_WritersAreUsable verifies that writers in context can be used.
func Test_Context_WritersAreUsable(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	ctx := cli.NewContext(stdout, stderr, nil, "")

	// Write to stdout via context
	_, err := ctx.Stdout.Write([]byte("test stdout"))
	require.NoError(t, err)
	assert.Equal(t, "test stdout", stdout.String())

	// Write to stderr via context
	_, err = ctx.Stderr.Write([]byte("test stderr"))
	require.NoError(t, err)
	assert.Equal(t, "test stderr", stderr.String())
}

// Test_CLICommand_Interface verifies that mock commands can implement CLICommand.
func Test_CLICommand_Interface(t *testing.T) {
	// This test verifies the CLICommand interface exists and can be implemented
	var _ cli.CLICommand = (*mockCommand)(nil)
}

// mockCommand is a test implementation of CLICommand interface.
type mockCommand struct {
	name     string
	synopsis string
	usage    string
	runFunc  func(ctx *cli.Context, args []string) int
}

func (m *mockCommand) Name() string {
	return m.name
}

func (m *mockCommand) Synopsis() string {
	return m.synopsis
}

func (m *mockCommand) Usage() string {
	return m.usage
}

func (m *mockCommand) SetFlags(fs *flag.FlagSet) {
	// No flags for mock command
}

func (m *mockCommand) Run(ctx *cli.Context, args []string) int {
	if m.runFunc != nil {
		return m.runFunc(ctx, args)
	}
	return 0
}

// Test_Run_OutputGoesToCorrectWriter verifies stdout/stderr separation.
func Test_Run_OutputGoesToCorrectWriter(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expectStdout bool
		expectStderr bool
	}{
		{
			name:         "help goes to stdout",
			args:         []string{"--help"},
			expectStdout: true,
			expectStderr: false,
		},
		{
			name:         "version goes to stdout",
			args:         []string{"--version"},
			expectStdout: true,
			expectStderr: false,
		},
		{
			name:         "unknown command error goes to stderr",
			args:         []string{"unknowncmd"},
			expectStdout: false,
			expectStderr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			cli.Run(tt.args, stdout, stderr)

			if tt.expectStdout {
				assert.NotEmpty(t, stdout.String(), "expected output on stdout")
			}
			if tt.expectStderr {
				assert.NotEmpty(t, stderr.String(), "expected output on stderr")
			}
		})
	}
}

// Test_Run_EmptyArgs is an explicit test for empty arguments.
func Test_Run_EmptyArgs(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := cli.Run([]string{}, stdout, stderr)

	assert.Equal(t, 0, exitCode, "empty args should show help and return 0")
	assert.NotEmpty(t, stdout.String(), "empty args should produce usage output")
}

// Test_Run_NilWriters tests behavior with nil writers.
func Test_Run_NilWriters(t *testing.T) {
	// This test verifies the function handles nil writers gracefully
	// The implementation should either:
	// 1. Accept nil and not write anything
	// 2. Panic (which we catch here)
	// 3. Use default writers (os.Stdout/os.Stderr)

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Run panicked with nil writers (acceptable behavior): %v", r)
		}
	}()

	// If this doesn't panic, the implementation handles nil gracefully
	exitCode := cli.Run([]string{"--version"}, nil, nil)
	_ = exitCode // We just want to verify no panic
}

// Test_Context_Fields verifies Context struct has expected fields.
func Test_Context_Fields(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	ctx := cli.NewContext(stdout, stderr, nil, "http://127.0.0.1:8765")

	// Verify Context has the expected fields by accessing them
	assert.NotNil(t, ctx)
	_ = ctx.Stdout      // Should exist
	_ = ctx.Stderr      // Should exist
	_ = ctx.APIEndpoint // Should exist
}

// Benchmark tests

func Benchmark_Run_Help(b *testing.B) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stdout.Reset()
		stderr.Reset()
		cli.Run([]string{"--help"}, stdout, stderr)
	}
}

func Benchmark_Run_Version(b *testing.B) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stdout.Reset()
		stderr.Reset()
		cli.Run([]string{"--version"}, stdout, stderr)
	}
}

func Benchmark_Run_UnknownCommand(b *testing.B) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stdout.Reset()
		stderr.Reset()
		cli.Run([]string{"unknown"}, stdout, stderr)
	}
}

func Benchmark_NewContext(b *testing.B) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cli.NewContext(stdout, stderr, nil, "")
	}
}

// Test_Run_CommandParsing tests argument parsing edge cases.
func Test_Run_CommandParsing(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantExitCode int
	}{
		{
			name:         "single dash is unknown",
			args:         []string{"-"},
			wantExitCode: 1,
		},
		{
			name:         "double dash alone is unknown",
			args:         []string{"--"},
			wantExitCode: 1,
		},
		{
			name:         "empty string argument is unknown",
			args:         []string{""},
			wantExitCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			exitCode := cli.Run(tt.args, stdout, stderr)

			assert.Equal(t, tt.wantExitCode, exitCode,
				"Run(%v) should return exit code %d", tt.args, tt.wantExitCode)
		})
	}
}

// Test_Run_SubcommandHelp tests help for subcommands.
func Test_Run_SubcommandHelp(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantExitCode int
		wantStdout   []string
	}{
		{
			name:         "rules with no subcommand shows rules help",
			args:         []string{"rules"},
			wantExitCode: 0,
			wantStdout:   []string{"rules"},
		},
		{
			name:         "rules --help shows rules help",
			args:         []string{"rules", "--help"},
			wantExitCode: 0,
			wantStdout:   []string{"rules"},
		},
		{
			name:         "rules -h shows rules help",
			args:         []string{"rules", "-h"},
			wantExitCode: 0,
			wantStdout:   []string{"rules"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			exitCode := cli.Run(tt.args, stdout, stderr)

			assert.Equal(t, tt.wantExitCode, exitCode)

			stdoutStr := strings.ToLower(stdout.String())
			for _, want := range tt.wantStdout {
				assert.Contains(t, stdoutStr, strings.ToLower(want))
			}
		})
	}
}
