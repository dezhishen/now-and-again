package simple

import (
	"github.com/dezhishen/now-and-again/backend/pkg/model"
	"github.com/dezhishen/now-and-again/backend/pkg/taskkind"
)

// Handler is the no-op task kind handler for simple tasks.
type Handler struct{}

func init() {
	taskkind.Register(Handler{})
}

func (Handler) Kind() string { return "simple" }

func (Handler) SaveExtra(_ taskkind.TaskStorage, _ *model.TaskModel, _ any) error   { return nil }
func (Handler) UpdateExtra(_ taskkind.TaskStorage, _ *model.TaskModel, _ any) error { return nil }
func (Handler) DeleteExtra(_ taskkind.TaskStorage, _ *model.TaskModel) error        { return nil }

func (Handler) OnTodo(_ taskkind.TaskStorage, _ *model.TodoModel, _ any) error {
	return nil
}

func (Handler) GetExtra(taskStorage taskkind.TaskStorage, task *model.TaskModel) (any, error) {
	return nil, nil
}
