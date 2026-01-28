package plugin

import (
	"errors"
	"testing"

	"github.com/bwmarrin/discordgo"

	"jamesbot/internal/command"
	"jamesbot/internal/middleware"
)

// mockCommandPlugin provides commands.
type mockCommandPlugin struct {
	mockPlugin
	commands []command.Command
}

func (m *mockCommandPlugin) Commands() []command.Command {
	return m.commands
}

// mockMiddlewarePlugin provides middleware.
type mockMiddlewarePlugin struct {
	mockPlugin
	middlewares []middleware.Middleware
}

func (m *mockMiddlewarePlugin) Middleware() []middleware.Middleware {
	return m.middlewares
}

// mockEventPlugin provides event handlers.
type mockEventPlugin struct {
	mockPlugin
	handlers []interface{}
}

func (m *mockEventPlugin) EventHandlers() []interface{} {
	return m.handlers
}

// mockInitializablePlugin requires initialization.
type mockInitializablePlugin struct {
	mockPlugin
	initError error
	initCalled bool
}

func (m *mockInitializablePlugin) Init(ctx *InitContext) error {
	m.initCalled = true
	return m.initError
}

// mockShutdownablePlugin requires shutdown.
type mockShutdownablePlugin struct {
	mockPlugin
	shutdownError error
	shutdownCalled bool
}

func (m *mockShutdownablePlugin) Shutdown() error {
	m.shutdownCalled = true
	return m.shutdownError
}

// mockCommand is a simple command for testing.
type mockCommand struct {
	name string
}

func (m *mockCommand) Name() string                                           { return m.name }
func (m *mockCommand) Description() string                                    { return "test" }
func (m *mockCommand) Options() []*discordgo.ApplicationCommandOption         { return nil }
func (m *mockCommand) Execute(ctx *command.Context) error                     { return nil }

func TestLoader_Load(t *testing.T) {
	tests := []struct {
		name    string
		plugin  Plugin
		wantErr bool
	}{
		{
			name:    "load simple plugin",
			plugin:  newMockPlugin("test"),
			wantErr: false,
		},
		{
			name: "load initializable plugin success",
			plugin: &mockInitializablePlugin{
				mockPlugin: mockPlugin{name: "init-success", description: "test", version: "1.0.0"},
				initError:  nil,
			},
			wantErr: false,
		},
		{
			name: "load initializable plugin failure",
			plugin: &mockInitializablePlugin{
				mockPlugin: mockPlugin{name: "init-fail", description: "test", version: "1.0.0"},
				initError:  errors.New("init error"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRegistry(testLogger())
			l := NewLoader(r, testLogger())

			err := l.Load(tt.plugin)

			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}

			// On failure, plugin should not be registered
			if tt.wantErr && r.Get(tt.plugin.Name()) != nil {
				t.Error("Failed plugin should not remain in registry")
			}
		})
	}
}

func TestLoader_LoadAll(t *testing.T) {
	r := NewRegistry(testLogger())
	l := NewLoader(r, testLogger())

	plugins := []Plugin{
		newMockPlugin("a"),
		newMockPlugin("b"),
		newMockPlugin("c"),
	}

	err := l.LoadAll(plugins...)

	if err != nil {
		t.Errorf("LoadAll() error = %v", err)
	}

	if r.Count() != 3 {
		t.Errorf("LoadAll() registered %d plugins, want 3", r.Count())
	}
}

func TestLoader_Commands(t *testing.T) {
	r := NewRegistry(testLogger())
	l := NewLoader(r, testLogger())

	// Plugin with commands
	cmdPlugin := &mockCommandPlugin{
		mockPlugin: mockPlugin{name: "cmd-plugin", description: "test", version: "1.0.0"},
		commands: []command.Command{
			&mockCommand{name: "cmd1"},
			&mockCommand{name: "cmd2"},
		},
	}

	// Plugin without commands
	simplePlugin := newMockPlugin("simple")

	_ = l.Load(cmdPlugin)
	_ = l.Load(simplePlugin)

	commands := l.Commands()

	if len(commands) != 2 {
		t.Errorf("Commands() returned %d commands, want 2", len(commands))
	}
}

func TestLoader_Middleware(t *testing.T) {
	r := NewRegistry(testLogger())
	l := NewLoader(r, testLogger())

	mw := func(next middleware.HandlerFunc) middleware.HandlerFunc { return next }

	mwPlugin := &mockMiddlewarePlugin{
		mockPlugin:  mockPlugin{name: "mw-plugin", description: "test", version: "1.0.0"},
		middlewares: []middleware.Middleware{mw, mw},
	}

	_ = l.Load(mwPlugin)

	middlewares := l.Middleware()

	if len(middlewares) != 2 {
		t.Errorf("Middleware() returned %d middlewares, want 2", len(middlewares))
	}
}

func TestLoader_EventHandlers(t *testing.T) {
	r := NewRegistry(testLogger())
	l := NewLoader(r, testLogger())

	handler := func(s *discordgo.Session, r *discordgo.Ready) {}

	eventPlugin := &mockEventPlugin{
		mockPlugin: mockPlugin{name: "event-plugin", description: "test", version: "1.0.0"},
		handlers:   []interface{}{handler},
	}

	_ = l.Load(eventPlugin)

	handlers := l.EventHandlers()

	if len(handlers) != 1 {
		t.Errorf("EventHandlers() returned %d handlers, want 1", len(handlers))
	}
}

func TestLoader_ShutdownAll(t *testing.T) {
	r := NewRegistry(testLogger())
	l := NewLoader(r, testLogger())

	shutdownPlugin := &mockShutdownablePlugin{
		mockPlugin: mockPlugin{name: "shutdown-plugin", description: "test", version: "1.0.0"},
	}

	_ = l.Load(shutdownPlugin)
	l.ShutdownAll()

	if !shutdownPlugin.shutdownCalled {
		t.Error("Shutdown() was not called on plugin")
	}
}

func TestLoader_Info(t *testing.T) {
	r := NewRegistry(testLogger())
	l := NewLoader(r, testLogger())

	_ = l.Load(newMockPlugin("test"))

	infos := l.Info()

	if len(infos) != 1 {
		t.Errorf("Info() returned %d items, want 1", len(infos))
	}
}
