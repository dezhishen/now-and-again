package indoor

import (
	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/backend/pkg/locationkind"
)

type Handler struct{}

func (Handler) Kind() string  { return "indoor" }
func (Handler) Label() string { return "室内" }
func (Handler) Icon() string  { return "🏠" }

func (Handler) Validate(loc *repository.LocationModel, extra any) error {
	// Indoor locations need no extra validation for now
	return nil
}

func init() {
	locationkind.Register(Handler{})
}
