package simple

import (
	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/backend/pkg/taskkind"
)

// Handler is the no-op task kind handler for simple tasks.
type Handler struct{}

func init() {
	taskkind.Register(Handler{})
}

func (Handler) Kind() string { return "simple" }

func (Handler) OnCreate(taskStorage taskkind.TaskStorage, task *repository.TaskModel, extra any) error {
	return nil
}

func (Handler) OnUpdate(taskStorage taskkind.TaskStorage, task *repository.TaskModel, extra any) error {
	return nil
}

func (Handler) OnDelete(taskStorage taskkind.TaskStorage, task *repository.TaskModel) error {
	return nil
}

func (Handler) OnComplete(taskStorage taskkind.TaskStorage, todo *repository.TodoModel, extra any) error {
	return nil
}

func (Handler) GetExtra(taskStorage taskkind.TaskStorage, task *repository.TaskModel) (any, error) {
	return nil, nil
}
