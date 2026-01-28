package jamesprial

import (
	"os"
	"testing"

	"github.com/rs/zerolog"

	"jamesbot/internal/plugin"
)

func testLogger() zerolog.Logger {
	return zerolog.New(os.Stderr).Level(zerolog.Disabled)
}

func TestPlugin_Metadata(t *testing.T) {
	p := New()

	if p.Name() != PluginName {
		t.Errorf("Name() = %q, want %q", p.Name(), PluginName)
	}

	if p.Version() != PluginVersion {
		t.Errorf("Version() = %q, want %q", p.Version(), PluginVersion)
	}

	if p.Description() == "" {
		t.Error("Description() should not be empty")
	}
}

func TestPlugin_Init(t *testing.T) {
	p := New()

	ctx := &plugin.InitContext{
		Logger: testLogger(),
	}

	err := p.Init(ctx)
	if err != nil {
		t.Errorf("Init() error = %v", err)
	}
}

func TestPlugin_Commands(t *testing.T) {
	p := New()

	commands := p.Commands()

	if len(commands) != 2 {
		t.Errorf("Commands() returned %d commands, want 2", len(commands))
	}

	// Check command names
	names := make(map[string]bool)
	for _, cmd := range commands {
		names[cmd.Name()] = true
	}

	if !names["greet"] {
		t.Error("Commands() should include 'greet' command")
	}

	if !names["jamesprial"] {
		t.Error("Commands() should include 'jamesprial' command")
	}
}

func TestPlugin_Shutdown(t *testing.T) {
	p := New()

	// Initialize first
	ctx := &plugin.InitContext{
		Logger: testLogger(),
	}
	_ = p.Init(ctx)

	err := p.Shutdown()
	if err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}
}

func TestPlugin_Interfaces(t *testing.T) {
	p := New()

	// Verify interface implementations
	var _ plugin.Plugin = p
	var _ plugin.CommandProvider = p
	var _ plugin.Initializable = p
	var _ plugin.Shutdownable = p
}
