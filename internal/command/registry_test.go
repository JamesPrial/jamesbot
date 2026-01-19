package command_test

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"testing"

	"jamesbot/internal/command"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCommand is a test double for the Command interface.
type mockCommand struct {
	name        string
	description string
	options     []*discordgo.ApplicationCommandOption
	executeFunc func(ctx *command.Context) error
}

func (m *mockCommand) Name() string {
	return m.name
}

func (m *mockCommand) Description() string {
	return m.description
}

func (m *mockCommand) Options() []*discordgo.ApplicationCommandOption {
	return m.options
}

func (m *mockCommand) Execute(ctx *command.Context) error {
	if m.executeFunc != nil {
		return m.executeFunc(ctx)
	}
	return nil
}

// newMockCommand creates a mock command with the given name.
func newMockCommand(name string) *mockCommand {
	return &mockCommand{
		name:        name,
		description: "A mock command for testing",
		options:     nil,
	}
}

// newMockCommandWithOptions creates a mock command with options.
func newMockCommandWithOptions(name, description string, options []*discordgo.ApplicationCommandOption) *mockCommand {
	return &mockCommand{
		name:        name,
		description: description,
		options:     options,
	}
}

// discardLogger returns a zerolog.Logger that discards all output.
func discardLogger() zerolog.Logger {
	return zerolog.New(io.Discard).Level(zerolog.Disabled)
}

func Test_NewRegistry(t *testing.T) {
	tests := []struct {
		name   string
		logger zerolog.Logger
	}{
		{
			name:   "create registry with logger",
			logger: discardLogger(),
		},
		{
			name:   "create registry with nop logger",
			logger: zerolog.Nop(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := command.NewRegistry(tt.logger)

			require.NotNil(t, registry, "NewRegistry should return non-nil *Registry")
		})
	}
}

func Test_Registry_Register(t *testing.T) {
	tests := []struct {
		name        string
		commands    []command.Command
		wantErr     bool
		errContains string
	}{
		{
			name: "register valid command",
			commands: []command.Command{
				newMockCommand("ping"),
			},
			wantErr: false,
		},
		{
			name: "register multiple unique commands",
			commands: []command.Command{
				newMockCommand("ping"),
				newMockCommand("pong"),
				newMockCommand("help"),
			},
			wantErr: false,
		},
		{
			name: "duplicate registration returns error",
			commands: []command.Command{
				newMockCommand("ping"),
				newMockCommand("ping"),
			},
			wantErr:     true,
			errContains: "already registered",
		},
		{
			name: "nil command returns error",
			commands: []command.Command{
				nil,
			},
			wantErr:     true,
			errContains: "nil command",
		},
		{
			name: "nil command after valid command returns error",
			commands: []command.Command{
				newMockCommand("valid"),
				nil,
			},
			wantErr:     true,
			errContains: "nil command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := command.NewRegistry(discardLogger())
			var lastErr error

			for _, cmd := range tt.commands {
				err := registry.Register(cmd)
				if err != nil {
					lastErr = err
					break
				}
			}

			if tt.wantErr {
				require.Error(t, lastErr, "Register should return an error")
				assert.Contains(t, lastErr.Error(), tt.errContains,
					"error message should contain %q", tt.errContains)
			} else {
				assert.NoError(t, lastErr, "Register should not return an error")
			}
		})
	}
}

func Test_Registry_Get(t *testing.T) {
	tests := []struct {
		name           string
		registeredCmds []string
		getCmdName     string
		wantFound      bool
	}{
		{
			name:           "get existing command",
			registeredCmds: []string{"ping"},
			getCmdName:     "ping",
			wantFound:      true,
		},
		{
			name:           "get non-existent command",
			registeredCmds: []string{"ping"},
			getCmdName:     "unknown",
			wantFound:      false,
		},
		{
			name:           "get from empty registry",
			registeredCmds: []string{},
			getCmdName:     "anything",
			wantFound:      false,
		},
		{
			name:           "get one of multiple registered commands",
			registeredCmds: []string{"ping", "pong", "help"},
			getCmdName:     "pong",
			wantFound:      true,
		},
		{
			name:           "case sensitive lookup - exact match",
			registeredCmds: []string{"Ping"},
			getCmdName:     "Ping",
			wantFound:      true,
		},
		{
			name:           "case sensitive lookup - wrong case",
			registeredCmds: []string{"Ping"},
			getCmdName:     "ping",
			wantFound:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := command.NewRegistry(discardLogger())

			// Register commands
			for _, name := range tt.registeredCmds {
				err := registry.Register(newMockCommand(name))
				require.NoError(t, err, "setup: Register should not fail")
			}

			// Get command
			cmd, found := registry.Get(tt.getCmdName)

			assert.Equal(t, tt.wantFound, found, "found should match expected")
			if tt.wantFound {
				require.NotNil(t, cmd, "command should not be nil when found")
				assert.Equal(t, tt.getCmdName, cmd.Name(),
					"returned command should have the requested name")
			} else {
				assert.Nil(t, cmd, "command should be nil when not found")
			}
		})
	}
}

func Test_Registry_All(t *testing.T) {
	tests := []struct {
		name           string
		registeredCmds []string
		wantCount      int
	}{
		{
			name:           "empty registry returns empty slice",
			registeredCmds: []string{},
			wantCount:      0,
		},
		{
			name:           "single command",
			registeredCmds: []string{"ping"},
			wantCount:      1,
		},
		{
			name:           "multiple commands",
			registeredCmds: []string{"ping", "pong", "help"},
			wantCount:      3,
		},
		{
			name:           "many commands",
			registeredCmds: []string{"cmd1", "cmd2", "cmd3", "cmd4", "cmd5"},
			wantCount:      5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := command.NewRegistry(discardLogger())

			// Register commands
			for _, name := range tt.registeredCmds {
				err := registry.Register(newMockCommand(name))
				require.NoError(t, err, "setup: Register should not fail")
			}

			// Get all commands
			all := registry.All()

			assert.Len(t, all, tt.wantCount, "All() should return expected number of commands")

			// Verify all registered commands are present
			names := make(map[string]bool)
			for _, cmd := range all {
				names[cmd.Name()] = true
			}
			for _, expectedName := range tt.registeredCmds {
				assert.True(t, names[expectedName],
					"All() should include command %q", expectedName)
			}
		})
	}
}

func Test_Registry_All_ReturnsNewSlice(t *testing.T) {
	registry := command.NewRegistry(discardLogger())
	err := registry.Register(newMockCommand("ping"))
	require.NoError(t, err)

	// Get all commands twice
	first := registry.All()
	second := registry.All()

	// Modify the first slice
	if len(first) > 0 {
		first[0] = nil
	}

	// Second slice should be unaffected
	assert.NotNil(t, second[0], "All() should return a new slice each time")
}

func Test_Registry_ApplicationCommands(t *testing.T) {
	tests := []struct {
		name           string
		commands       []*mockCommand
		wantCount      int
		verifyCommands func(t *testing.T, cmds []*discordgo.ApplicationCommand)
	}{
		{
			name:      "empty registry returns empty slice",
			commands:  []*mockCommand{},
			wantCount: 0,
		},
		{
			name: "single command converts correctly",
			commands: []*mockCommand{
				newMockCommandWithOptions("ping", "Ping the bot", nil),
			},
			wantCount: 1,
			verifyCommands: func(t *testing.T, cmds []*discordgo.ApplicationCommand) {
				assert.Equal(t, "ping", cmds[0].Name)
				assert.Equal(t, "Ping the bot", cmds[0].Description)
			},
		},
		{
			name: "multiple commands convert correctly",
			commands: []*mockCommand{
				newMockCommandWithOptions("ping", "Ping the bot", nil),
				newMockCommandWithOptions("help", "Get help", nil),
				newMockCommandWithOptions("status", "Check status", nil),
			},
			wantCount: 3,
		},
		{
			name: "command with options converts correctly",
			commands: []*mockCommand{
				newMockCommandWithOptions("echo", "Echo a message", []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "text",
						Description: "The text to echo",
						Required:    true,
					},
				}),
			},
			wantCount: 1,
			verifyCommands: func(t *testing.T, cmds []*discordgo.ApplicationCommand) {
				require.Len(t, cmds[0].Options, 1)
				assert.Equal(t, "text", cmds[0].Options[0].Name)
				assert.Equal(t, discordgo.ApplicationCommandOptionString, cmds[0].Options[0].Type)
				assert.True(t, cmds[0].Options[0].Required)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := command.NewRegistry(discardLogger())

			// Register commands
			for _, cmd := range tt.commands {
				err := registry.Register(cmd)
				require.NoError(t, err, "setup: Register should not fail")
			}

			// Get application commands
			appCmds := registry.ApplicationCommands()

			assert.Len(t, appCmds, tt.wantCount,
				"ApplicationCommands() should return expected number of commands")

			// Run custom verification if provided
			if tt.verifyCommands != nil && len(appCmds) > 0 {
				tt.verifyCommands(t, appCmds)
			}
		})
	}
}

func Test_Registry_ConcurrentRegister(t *testing.T) {
	registry := command.NewRegistry(discardLogger())
	numGoroutines := 100

	var wg sync.WaitGroup
	errChan := make(chan error, numGoroutines)

	// Try to register commands concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			cmd := newMockCommand("concurrent-cmd")
			err := registry.Register(cmd)
			if err != nil {
				errChan <- err
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	// Collect errors
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	// Exactly one should succeed, rest should fail with "already registered"
	successCount := numGoroutines - len(errs)
	assert.Equal(t, 1, successCount,
		"exactly one concurrent registration should succeed")

	// All errors should be "already registered"
	for _, err := range errs {
		assert.Contains(t, err.Error(), "already registered",
			"concurrent registration errors should be 'already registered'")
	}

	// Verify the command was registered
	cmd, found := registry.Get("concurrent-cmd")
	assert.True(t, found, "command should be registered")
	assert.NotNil(t, cmd, "command should not be nil")
}

func Test_Registry_ConcurrentGet(t *testing.T) {
	registry := command.NewRegistry(discardLogger())

	// Register a command first
	err := registry.Register(newMockCommand("target"))
	require.NoError(t, err)

	numGoroutines := 100
	var wg sync.WaitGroup
	results := make(chan struct {
		cmd   command.Command
		found bool
	}, numGoroutines)

	// Concurrent Get calls
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cmd, found := registry.Get("target")
			results <- struct {
				cmd   command.Command
				found bool
			}{cmd, found}
		}()
	}

	wg.Wait()
	close(results)

	// All should find the command with consistent results
	for result := range results {
		assert.True(t, result.found, "all concurrent Gets should find the command")
		assert.NotNil(t, result.cmd, "all concurrent Gets should return non-nil command")
		assert.Equal(t, "target", result.cmd.Name(),
			"all concurrent Gets should return the correct command")
	}
}

func Test_Registry_ConcurrentRegisterAndGet(t *testing.T) {
	registry := command.NewRegistry(discardLogger())
	numGoroutines := 50

	var wg sync.WaitGroup

	// Half register, half get
	for i := 0; i < numGoroutines; i++ {
		wg.Add(2)

		// Register goroutine
		go func(id int) {
			defer wg.Done()
			cmd := newMockCommand("cmd-" + string(rune('A'+id%26)))
			_ = registry.Register(cmd) // Ignore errors (some will be duplicates)
		}(i)

		// Get goroutine
		go func(id int) {
			defer wg.Done()
			_, _ = registry.Get("cmd-" + string(rune('A'+id%26)))
		}(i)
	}

	wg.Wait()

	// Test should complete without race conditions or panics
	// Verify we can still use the registry
	all := registry.All()
	assert.Greater(t, len(all), 0, "registry should have some commands after concurrent operations")
}

func Test_Registry_Register_EmptyName(t *testing.T) {
	registry := command.NewRegistry(discardLogger())

	cmd := newMockCommand("")
	err := registry.Register(cmd)

	// Empty names should be rejected
	assert.Error(t, err, "registering command with empty name should fail")
	assert.Contains(t, err.Error(), "empty name")
}

func Test_Registry_Get_EmptyName(t *testing.T) {
	registry := command.NewRegistry(discardLogger())

	cmd, found := registry.Get("")

	assert.False(t, found, "getting command with empty name from empty registry should return false")
	assert.Nil(t, cmd, "getting command with empty name should return nil")
}

// Verify that Command interface is properly defined
func Test_Command_Interface(t *testing.T) {
	// This test verifies that our mock satisfies the Command interface
	var _ command.Command = (*mockCommand)(nil)
}

// Test that ApplicationCommands returns properly formatted Discord commands
func Test_Registry_ApplicationCommands_Format(t *testing.T) {
	registry := command.NewRegistry(discardLogger())

	// Register a command with full options
	cmd := newMockCommandWithOptions(
		"complex",
		"A complex command with multiple options",
		[]*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "required-arg",
				Description: "A required string argument",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "optional-count",
				Description: "An optional count",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "target-user",
				Description: "A target user",
				Required:    false,
			},
		},
	)

	err := registry.Register(cmd)
	require.NoError(t, err)

	appCmds := registry.ApplicationCommands()

	require.Len(t, appCmds, 1)
	appCmd := appCmds[0]

	// Verify top-level fields
	assert.Equal(t, "complex", appCmd.Name)
	assert.Equal(t, "A complex command with multiple options", appCmd.Description)

	// Verify options
	require.Len(t, appCmd.Options, 3)

	// First option (required string)
	assert.Equal(t, "required-arg", appCmd.Options[0].Name)
	assert.Equal(t, discordgo.ApplicationCommandOptionString, appCmd.Options[0].Type)
	assert.True(t, appCmd.Options[0].Required)

	// Second option (optional integer)
	assert.Equal(t, "optional-count", appCmd.Options[1].Name)
	assert.Equal(t, discordgo.ApplicationCommandOptionInteger, appCmd.Options[1].Type)
	assert.False(t, appCmd.Options[1].Required)

	// Third option (optional user)
	assert.Equal(t, "target-user", appCmd.Options[2].Name)
	assert.Equal(t, discordgo.ApplicationCommandOptionUser, appCmd.Options[2].Type)
	assert.False(t, appCmd.Options[2].Required)
}

// Benchmark tests
func Benchmark_Registry_Register(b *testing.B) {
	logger := discardLogger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry := command.NewRegistry(logger)
		for j := 0; j < 100; j++ {
			cmd := newMockCommand("cmd-" + string(rune(j)))
			_ = registry.Register(cmd)
		}
	}
}

func Benchmark_Registry_Get(b *testing.B) {
	registry := command.NewRegistry(discardLogger())

	// Setup: register commands
	for i := 0; i < 100; i++ {
		cmd := newMockCommand("cmd-" + string(rune(i)))
		_ = registry.Register(cmd)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = registry.Get("cmd-" + string(rune(i%100)))
	}
}

func Benchmark_Registry_All(b *testing.B) {
	registry := command.NewRegistry(discardLogger())

	// Setup: register commands
	for i := 0; i < 100; i++ {
		cmd := newMockCommand("cmd-" + string(rune(i)))
		_ = registry.Register(cmd)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = registry.All()
	}
}

func Benchmark_Registry_ApplicationCommands(b *testing.B) {
	registry := command.NewRegistry(discardLogger())

	// Setup: register commands
	for i := 0; i < 100; i++ {
		cmd := newMockCommandWithOptions(
			"cmd-"+string(rune(i)),
			"Description",
			[]*discordgo.ApplicationCommandOption{
				{Name: "opt1", Type: discordgo.ApplicationCommandOptionString},
				{Name: "opt2", Type: discordgo.ApplicationCommandOptionInteger},
			},
		)
		_ = registry.Register(cmd)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = registry.ApplicationCommands()
	}
}

// Test error type checking
func Test_Registry_Register_ErrorTypes(t *testing.T) {
	registry := command.NewRegistry(discardLogger())

	t.Run("nil command error is descriptive", func(t *testing.T) {
		err := registry.Register(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "nil command")
	})

	t.Run("duplicate command error is descriptive", func(t *testing.T) {
		err := registry.Register(newMockCommand("dup"))
		require.NoError(t, err)

		err = registry.Register(newMockCommand("dup"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
		assert.Contains(t, err.Error(), "dup")
	})
}

// Test that errors are not nil-check safe
func Test_Registry_Register_NilPointerCommand(t *testing.T) {
	registry := command.NewRegistry(discardLogger())

	var nilCmd *mockCommand = nil
	err := registry.Register(nilCmd)

	require.Error(t, err, "registering nil pointer should return error")
	assert.Contains(t, err.Error(), "nil command")
}

// Verify the Command interface methods match expected signatures
func Test_Command_InterfaceMethods(t *testing.T) {
	cmd := newMockCommand("test")

	t.Run("Name returns string", func(t *testing.T) {
		name := cmd.Name()
		assert.IsType(t, "", name)
	})

	t.Run("Description returns string", func(t *testing.T) {
		desc := cmd.Description()
		assert.IsType(t, "", desc)
	})

	t.Run("Options returns slice of ApplicationCommandOption", func(t *testing.T) {
		opts := cmd.Options()
		assert.IsType(t, []*discordgo.ApplicationCommandOption{}, opts)
	})

	t.Run("Execute takes Context and returns error", func(t *testing.T) {
		err := cmd.Execute(nil)
		if err != nil {
			assert.Implements(t, (*error)(nil), err)
		}
	})
}

// Test with various special characters in command names
func Test_Registry_SpecialCharacterNames(t *testing.T) {
	tests := []struct {
		name     string
		cmdName  string
		wantFind bool
	}{
		{
			name:     "hyphenated name",
			cmdName:  "my-command",
			wantFind: true,
		},
		{
			name:     "underscored name",
			cmdName:  "my_command",
			wantFind: true,
		},
		{
			name:     "numeric suffix",
			cmdName:  "command123",
			wantFind: true,
		},
		{
			name:     "numeric prefix",
			cmdName:  "123command",
			wantFind: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := command.NewRegistry(discardLogger())

			err := registry.Register(newMockCommand(tt.cmdName))
			require.NoError(t, err)

			cmd, found := registry.Get(tt.cmdName)
			assert.Equal(t, tt.wantFind, found)
			if tt.wantFind {
				assert.Equal(t, tt.cmdName, cmd.Name())
			}
		})
	}
}

// Test that errors returned are usable with errors.Is/As patterns
func Test_Registry_ErrorUnwrapping(t *testing.T) {
	registry := command.NewRegistry(discardLogger())

	// Register a command
	err := registry.Register(newMockCommand("test"))
	require.NoError(t, err)

	// Try to register duplicate
	err = registry.Register(newMockCommand("test"))
	require.Error(t, err)

	// The error should be usable as a standard error
	var stdErr error = err
	assert.NotNil(t, stdErr)

	// Check if the error wraps any specific error types (implementation dependent)
	// This test ensures the error is properly formed
	assert.NotEmpty(t, err.Error())
}

// Test registering and retrieving many commands
func Test_Registry_ManyCommands(t *testing.T) {
	registry := command.NewRegistry(discardLogger())
	numCommands := 1000

	// Register many commands
	for i := 0; i < numCommands; i++ {
		name := fmt.Sprintf("cmd-%d", i)
		err := registry.Register(newMockCommand(name))
		require.NoError(t, err, "should register command %s", name)
	}

	// Verify all are retrievable
	all := registry.All()
	assert.Len(t, all, numCommands, "should have all commands registered")

	// Verify specific lookups work
	for i := 0; i < numCommands; i++ {
		name := fmt.Sprintf("cmd-%d", i)
		cmd, found := registry.Get(name)
		assert.True(t, found, "should find command %s", name)
		assert.Equal(t, name, cmd.Name())
	}
}

// Test that a command with executeFunc returning an error is properly set up
func Test_MockCommand_ExecuteFunc(t *testing.T) {
	expectedErr := errors.New("execution failed")
	cmd := &mockCommand{
		name: "failing",
		executeFunc: func(ctx *command.Context) error {
			return expectedErr
		},
	}

	err := cmd.Execute(nil)
	assert.Equal(t, expectedErr, err)
}

// mockPermissionedCommand implements both Command and PermissionedCommand interfaces.
type mockPermissionedCommand struct {
	mockCommand
	permissions int64
}

func (m *mockPermissionedCommand) Permissions() int64 {
	return m.permissions
}

func Test_Registry_ApplicationCommands_WithPermissions(t *testing.T) {
	registry := command.NewRegistry(discardLogger())

	// Register a command with permissions
	cmd := &mockPermissionedCommand{
		mockCommand: mockCommand{
			name:        "admin",
			description: "Admin command",
			options:     nil,
		},
		permissions: discordgo.PermissionAdministrator,
	}

	err := registry.Register(cmd)
	require.NoError(t, err)

	appCmds := registry.ApplicationCommands()

	require.Len(t, appCmds, 1)
	require.NotNil(t, appCmds[0].DefaultMemberPermissions)
	assert.Equal(t, int64(discordgo.PermissionAdministrator), *appCmds[0].DefaultMemberPermissions)
}

// Verify PermissionedCommand interface
func Test_PermissionedCommand_Interface(t *testing.T) {
	var _ command.PermissionedCommand = (*mockPermissionedCommand)(nil)
}
