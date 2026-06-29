package tasktemplate

import (
	"fmt"
	"sync"
)

// Registry is a thread-safe provider registry.
type Registry struct {
	mu       sync.RWMutex
	handlers map[string]Provider
}

// NewRegistry creates a ready-to-use Registry.
func NewRegistry() *Registry {
	return &Registry{handlers: make(map[string]Provider)}
}

// Register adds a provider. Panics on duplicate code.
func (r *Registry) Register(p Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	code := p.Code()
	if _, ok := r.handlers[code]; ok {
		panic(fmt.Sprintf("tasktemplate provider %q already registered", code))
	}
	r.handlers[code] = p
}

// Get returns the provider registered under code, or nil.
func (r *Registry) Get(code string) Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.handlers[code]
}

// All returns every registered provider in no particular order.
func (r *Registry) All() []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]Provider, 0, len(r.handlers))
	for _, h := range r.handlers {
		result = append(result, h)
	}
	return result
}

// ─── package-level convenience ──────────────────────────────────

var defaultRegistry = NewRegistry()

// Register adds p to the default package registry.
func Register(p Provider) { defaultRegistry.Register(p) }

// GetProvider returns the registered provider or nil.
func GetProvider(code string) Provider { return defaultRegistry.Get(code) }

// AllProviders returns all registered providers.
func AllProviders() []Provider { return defaultRegistry.All() }
