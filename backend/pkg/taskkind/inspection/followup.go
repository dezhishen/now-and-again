package inspection

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/backend/pkg/scheduler"
	"github.com/dezhishen/now-and-again/backend/pkg/taskkind"
)

// ─── Follow-up Task ──────────────────────────────────────────────

type branchConfig struct {
	Name       string `json:"name"`
	CreateTodo bool   `json:"create_todo"`
	TodoName   string `json:"todo_name"`
	GroupID    string `json:"group_id"`
}

type checkItemConfig struct {
	Name     string         `json:"name"`
	Branches []branchConfig `json:"branches"`
}

func createFollowUpTask(ops *taskkind.Ops, todo *repository.TodoModel, itemName, branchName, userID string) error {
	items, err := parseCheckItems(todo.Task.CheckItems)
	if err != nil {
		return fmt.Errorf("parse check items: %w", err)
	}
	for _, item := range items {
		if item.Name != itemName {
			continue
		}
		for _, b := range item.Branches {
			if b.Name == branchName && b.CreateTodo {
				name := b.TodoName
				if name == "" {
					name = fmt.Sprintf("%s - %s - %s", todo.Task.Name, itemName, branchName)
				}
				name = strings.Replace(name, "{name}", todo.Task.Name, -1)
				return createOneTimeTask(ops, todo, name, b.GroupID, userID)
			}
		}
	}
	return nil
}

func createOneTimeTask(ops *taskkind.Ops, todo *repository.TodoModel, name, groupID, userID string) error {
	followTask := &repository.TaskTemplateModel{
		FamilyID:     todo.FamilyID,
		GroupID:      groupID,
		LocationID:   todo.LocationID,
		Name:         name,
		ScheduleType: "once",
		ScheduleData: `{"time":"` + time.Now().Format("15:04") + `"}`,
		Enabled:      true,
		Kind:         "simple",
		CreatedBy:    userID,
	}
	if err := ops.Repo.CreateTask(followTask); err != nil {
		return fmt.Errorf("create follow-up task: %w", err)
	}

	ops.Repo.CreateUserLog(todo.TaskID, todo.ID, userID, "follow_up",
		fmt.Sprintf("巡检跟进 → 创建任务「%s」(%s)", name, followTask.ID))
	ops.Repo.CreateUserLog(followTask.ID, "", userID, "created",
		fmt.Sprintf("从巡检「%s」创建", todo.Task.Name))

	// Create the first todo immediately — don't wait for the scheduler.
	followTodo := &repository.TodoModel{
		TaskID:     followTask.ID,
		FamilyID:   todo.FamilyID,
		LocationID: todo.LocationID,
		Status:     "pending",
		DueStart:   time.Now(),
		DueDate:    time.Now().Add(24 * time.Hour),
	}
	if err := ops.Repo.CreateTodo(followTodo); err != nil {
		return fmt.Errorf("create follow-up todo: %w", err)
	}

	// Register with scheduler solely for auto-disable after the one-shot fires.
	ops.Scheduler.RegisterJob(&scheduler.JobBuilder{
		TaskID:       followTask.ID,
		ScheduleType: "once",
		ScheduleData: followTask.ScheduleData,
		Callback:     func(taskID string, triggeredAt time.Time) error { return nil },
		AfterFire: func(taskID string) {
			ops.Repo.DisableTask(taskID)
		},
	})
	return nil
}

func parseCheckItems(raw string) ([]checkItemConfig, error) {
	var items []checkItemConfig
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil, err
	}
	return items, nil
}
