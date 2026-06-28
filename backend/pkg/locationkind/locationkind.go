package locationkind

import "github.com/dezhishen/now-and-again/backend/pkg/model"

// ─── Handler Interface ───────────────────────────────────────────

// Handler defines location-kind-specific behavior and metadata.
// Adding a new location type only requires implementing this interface
// and registering via init(). No if/else needed anywhere.
type Handler interface {
	Kind() string

	// Metadata for the frontend dropdown and display
	Label() string // e.g. "室内", "户外", "学校"
	Icon() string  // e.g. "🏠", "🌳", "🏫"

	// Validate is called before creating/updating a location.
	// Return nil if valid, or an error if the extra data is invalid.
	Validate(loc *model.LocationModel, extra any) error
}

// ─── Registry ────────────────────────────────────────────────────

var registry = map[string]Handler{}

func Register(h Handler) { registry[h.Kind()] = h }

func Get(kind string) Handler { return registry[kind] }

// All returns all registered handlers for frontend dropdown population.
func All() []Handler {
	result := make([]Handler, 0, len(registry))
	for _, h := range registry {
		result = append(result, h)
	}
	return result
}
