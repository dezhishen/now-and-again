package scheduler

import (
	"fmt"
	"sort"
	"sync"
)

// Handler defines a schedule type. Implementations map schedule_data
// to gocron job definitions. Register in init().
type Handler interface {
	Code() string
	BuildJob(data map[string]any) *JobDef
	IsOneShot() bool
}

// Registry holds registered schedule handlers, keyed by Code().
type Registry struct {
	mu       sync.RWMutex
	handlers map[string]Handler
}

// NewRegistry creates an empty handler registry.
func NewRegistry() *Registry {
	return &Registry{handlers: make(map[string]Handler)}
}

// Register adds a handler. Panics on duplicate code.
func (r *Registry) Register(h Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	code := h.Code()
	if _, ok := r.handlers[code]; ok {
		panic(fmt.Sprintf("handler %q already registered", code))
	}
	r.handlers[code] = h
}

// Get returns the handler for the given code, or nil.
func (r *Registry) Get(code string) Handler {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.handlers[code]
}

// All returns all registered handlers sorted by code.
func (r *Registry) All() []Handler {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]Handler, 0, len(r.handlers))
	for _, h := range r.handlers {
		list = append(list, h)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Code() < list[j].Code() })
	return list
}

// ─── Default global registry (backward compatible) ──────────────

var defaultRegistry = NewRegistry()

// Register adds a handler to the default global registry.
func Register(h Handler) { defaultRegistry.Register(h) }

// HandlerByCode looks up a handler in the default registry.
func HandlerByCode(code string) Handler { return defaultRegistry.Get(code) }

// AllHandlers returns all handlers from the default registry.
func AllHandlers() []Handler { return defaultRegistry.All() }
