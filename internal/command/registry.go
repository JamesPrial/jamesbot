package command

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
)

// Registry manages the collection of registered bot commands.
// It provides thread-safe registration and retrieval of commands.
type Registry struct {
	commands map[string]Command
	mu       sync.RWMutex
	logger   zerolog.Logger
}

// NewRegistry creates a new command registry with the provided logger.
func NewRegistry(logger zerolog.Logger) *Registry {
	return &Registry{
		commands: make(map[string]Command),
		logger:   logger,
	}
}

// Register adds a command to the registry.
// It returns an error if the command is nil or if a command with the same name
// is already registered.
func (r *Registry) Register(cmd Command) error {
	if cmd == nil || reflect.ValueOf(cmd).IsNil() {
		return fmt.Errorf("cannot register nil command")
	}

	name := cmd.Name()
	if name == "" {
		return fmt.Errorf("cannot register command with empty name")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.commands[name]; exists {
		return fmt.Errorf("command %q is already registered", name)
	}

	r.commands[name] = cmd
	r.logger.Debug().Str("command", name).Msg("registered command")

	return nil
}

// Get retrieves a command by name from the registry.
// It returns the command and true if found, or nil and false if not found.
func (r *Registry) Get(name string) (Command, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cmd, exists := r.commands[name]
	return cmd, exists
}

// All returns a slice of all registered commands.
// The returned slice is a copy and can be safely modified by the caller.
func (r *Registry) All() []Command {
	r.mu.RLock()
	defer r.mu.RUnlock()

	commands := make([]Command, 0, len(r.commands))
	for _, cmd := range r.commands {
		commands = append(commands, cmd)
	}

	return commands
}

// ApplicationCommands converts all registered commands to Discord application commands.
// This is used to register commands with Discord's API.
func (r *Registry) ApplicationCommands() []*discordgo.ApplicationCommand {
	r.mu.RLock()
	defer r.mu.RUnlock()

	appCommands := make([]*discordgo.ApplicationCommand, 0, len(r.commands))

	for _, cmd := range r.commands {
		appCmd := &discordgo.ApplicationCommand{
			Name:        cmd.Name(),
			Description: cmd.Description(),
			Options:     cmd.Options(),
		}

		// If the command implements PermissionedCommand, set default member permissions
		if permCmd, ok := cmd.(PermissionedCommand); ok {
			perms := permCmd.Permissions()
			appCmd.DefaultMemberPermissions = &perms
		}

		appCommands = append(appCommands, appCmd)
	}

	return appCommands
}
