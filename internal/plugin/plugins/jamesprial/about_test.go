package jamesprial

import (
	"testing"
)

func TestAboutPluginCommand_Name(t *testing.T) {
	cmd := &AboutPluginCommand{}
	if cmd.Name() != "jamesprial" {
		t.Errorf("Name() = %q, want %q", cmd.Name(), "jamesprial")
	}
}

func TestAboutPluginCommand_Description(t *testing.T) {
	cmd := &AboutPluginCommand{}
	if cmd.Description() == "" {
		t.Error("Description() should not be empty")
	}
}

func TestAboutPluginCommand_Options(t *testing.T) {
	cmd := &AboutPluginCommand{}
	opts := cmd.Options()

	if len(opts) > 0 {
		t.Errorf("Options() should return nil or empty slice, got %d options", len(opts))
	}
}
