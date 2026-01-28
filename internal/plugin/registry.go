package plugin

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog"
)

// Registry manages plugin registration and lifecycle.
type Registry struct {
	mu      sync.RWMutex
	plugins map[string]Plugin
	order   []string // Maintains registration order
	logger  zerolog.Logger
}

// NewRegistry creates a new plugin registry.
func NewRegistry(logger zerolog.Logger) *Registry {
	return &Registry{
		plugins: make(map[string]Plugin),
		order:   make([]string, 0),
		logger:  logger.With().Str("component", "plugin-registry").Logger(),
	}
}

// Register adds a plugin to the registry.
// Returns an error if a plugin with the same name already exists.
func (r *Registry) Register(p Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := p.Name()
	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("plugin %q already registered", name)
	}

	r.plugins[name] = p
	r.order = append(r.order, name)

	r.logger.Info().
		Str("plugin", name).
		Str("version", p.Version()).
		Msg("plugin registered")

	return nil
}

// Get returns a plugin by name, or nil if not found.
func (r *Registry) Get(name string) Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.plugins[name]
}

// All returns all registered plugins in registration order.
func (r *Registry) All() []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Plugin, 0, len(r.order))
	for _, name := range r.order {
		if p, ok := r.plugins[name]; ok {
			result = append(result, p)
		}
	}
	return result
}

// Names returns the names of all registered plugins in registration order.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]string, len(r.order))
	copy(result, r.order)
	return result
}

// Count returns the number of registered plugins.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.plugins)
}

// Unregister removes a plugin from the registry.
// Returns an error if the plugin is not found.
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[name]; !exists {
		return fmt.Errorf("plugin %q not found", name)
	}

	delete(r.plugins, name)

	// Remove from order slice
	for i, n := range r.order {
		if n == name {
			r.order = append(r.order[:i], r.order[i+1:]...)
			break
		}
	}

	r.logger.Info().
		Str("plugin", name).
		Msg("plugin unregistered")

	return nil
}

// Info returns metadata about all registered plugins.
func (r *Registry) Info() []Info {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Info, 0, len(r.order))
	for _, name := range r.order {
		if p, ok := r.plugins[name]; ok {
			result = append(result, InfoFromPlugin(p))
		}
	}
	return result
}
