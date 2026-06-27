package simple

import (
	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/backend/pkg/taskkind"
)

// Handler is the no-op task kind handler for simple tasks.
type Handler struct{}

func (Handler) Kind() string { return "simple" }

func (Handler) OnComplete(ops *taskkind.Ops, todo *repository.TodoModel, extra any, branchName, userID string) error {
	return nil
}

func (Handler) OnCreate(ops *taskkind.Ops, task *repository.TaskTemplateModel, extra any) error {
	return nil
}

func (Handler) OnUpdate(ops *taskkind.Ops, task *repository.TaskTemplateModel, extra any) error {
	return nil
}

func (Handler) GetExtra(ops *taskkind.Ops, task *repository.TaskTemplateModel) (any, error) {
	return nil, nil
}

func (Handler) OnDelete(ops *taskkind.Ops, task *repository.TaskTemplateModel) error {
	return nil
}

func init() {
	taskkind.Register(Handler{})
}
