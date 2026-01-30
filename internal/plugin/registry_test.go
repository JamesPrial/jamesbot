package plugin

import (
	"os"
	"testing"

	"github.com/rs/zerolog"
)

// mockPlugin is a simple plugin for testing.
type mockPlugin struct {
	name        string
	description string
	version     string
}

func (m *mockPlugin) Name() string        { return m.name }
func (m *mockPlugin) Description() string { return m.description }
func (m *mockPlugin) Version() string     { return m.version }

func newMockPlugin(name string) *mockPlugin {
	return &mockPlugin{
		name:        name,
		description: "Mock plugin for testing",
		version:     "1.0.0",
	}
}

func testLogger() zerolog.Logger {
	return zerolog.New(os.Stderr).Level(zerolog.Disabled)
}

func TestRegistry_Register(t *testing.T) {
	tests := []struct {
		name      string
		plugins   []*mockPlugin
		wantErr   bool
		wantCount int
	}{
		{
			name:      "register single plugin",
			plugins:   []*mockPlugin{newMockPlugin("test")},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name:      "register multiple plugins",
			plugins:   []*mockPlugin{newMockPlugin("test1"), newMockPlugin("test2")},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:      "register duplicate plugin",
			plugins:   []*mockPlugin{newMockPlugin("test"), newMockPlugin("test")},
			wantErr:   true,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRegistry(testLogger())

			var lastErr error
			for _, p := range tt.plugins {
				if err := r.Register(p); err != nil {
					lastErr = err
				}
			}

			if (lastErr != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", lastErr, tt.wantErr)
			}

			if r.Count() != tt.wantCount {
				t.Errorf("Count() = %d, want %d", r.Count(), tt.wantCount)
			}
		})
	}
}

func TestRegistry_Get(t *testing.T) {
	r := NewRegistry(testLogger())
	p := newMockPlugin("test")
	_ = r.Register(p)

	tests := []struct {
		name       string
		pluginName string
		wantFound  bool
	}{
		{
			name:       "get existing plugin",
			pluginName: "test",
			wantFound:  true,
		},
		{
			name:       "get non-existing plugin",
			pluginName: "nonexistent",
			wantFound:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.Get(tt.pluginName)
			if (got != nil) != tt.wantFound {
				t.Errorf("Get() found = %v, wantFound %v", got != nil, tt.wantFound)
			}
		})
	}
}

func TestRegistry_All(t *testing.T) {
	r := NewRegistry(testLogger())
	_ = r.Register(newMockPlugin("a"))
	_ = r.Register(newMockPlugin("b"))
	_ = r.Register(newMockPlugin("c"))

	all := r.All()

	if len(all) != 3 {
		t.Errorf("All() returned %d plugins, want 3", len(all))
	}

	// Verify order is maintained
	names := make([]string, len(all))
	for i, p := range all {
		names[i] = p.Name()
	}

	if names[0] != "a" || names[1] != "b" || names[2] != "c" {
		t.Errorf("All() order = %v, want [a, b, c]", names)
	}
}

func TestRegistry_Names(t *testing.T) {
	r := NewRegistry(testLogger())
	_ = r.Register(newMockPlugin("first"))
	_ = r.Register(newMockPlugin("second"))

	names := r.Names()

	if len(names) != 2 {
		t.Errorf("Names() returned %d names, want 2", len(names))
	}

	if names[0] != "first" || names[1] != "second" {
		t.Errorf("Names() = %v, want [first, second]", names)
	}
}

func TestRegistry_Unregister(t *testing.T) {
	r := NewRegistry(testLogger())
	_ = r.Register(newMockPlugin("test"))

	// Unregister existing
	if err := r.Unregister("test"); err != nil {
		t.Errorf("Unregister() error = %v", err)
	}

	if r.Count() != 0 {
		t.Errorf("Count() after unregister = %d, want 0", r.Count())
	}

	// Unregister non-existing
	if err := r.Unregister("nonexistent"); err == nil {
		t.Error("Unregister() should error for non-existing plugin")
	}
}

func TestRegistry_Info(t *testing.T) {
	r := NewRegistry(testLogger())
	_ = r.Register(newMockPlugin("test"))

	infos := r.Info()

	if len(infos) != 1 {
		t.Errorf("Info() returned %d items, want 1", len(infos))
	}

	if infos[0].Name != "test" {
		t.Errorf("Info()[0].Name = %q, want %q", infos[0].Name, "test")
	}
}
