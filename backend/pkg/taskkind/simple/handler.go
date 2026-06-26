package simple

import (
	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/backend/pkg/taskkind"
)

// Handler is the no-op task kind handler for simple tasks.
type Handler struct{}

func (Handler) Kind() string { return "simple" }

func (Handler) OnComplete(ops *taskkind.Ops, todo *repository.TodoModel, branchName, userID string) error {
	return nil
}

func init() {
	taskkind.Register(Handler{})
}
