package inspection

import (
	"fmt"

	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/backend/pkg/taskkind"
)

// Handler implements taskkind.Handler and taskkind.Inspector for inspection tasks.
type Handler struct{}

func (Handler) Kind() string { return "inspection" }

func (Handler) OnComplete(ops *taskkind.Ops, todo *repository.TodoModel, branchName, userID string) error {
	// Base completion log is already recorded by TaskService.
	return nil
}

// OnInspect handles the multi-item inspection submission.
func (Handler) OnInspect(ops *taskkind.Ops, todo *repository.TodoModel, selections []taskkind.Selection, userID string) error {
	for _, sel := range selections {
		ops.Repo.CreateUserLog(todo.TaskID, todo.ID, userID, "inspect:"+sel.Branch,
			fmt.Sprintf("[%s] → %s", sel.Item, sel.Branch))
		createFollowUpTask(ops, todo, sel.Item, sel.Branch, userID)
	}
	ops.Repo.CreateUserLog(todo.TaskID, todo.ID, userID, "inspection",
		fmt.Sprintf("巡检完成: %s", todo.Task.Name))
	return nil
}

func init() {
	taskkind.Register(Handler{})
}
