package chain

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/dezhishen/now-and-again/backend/pkg/model"
	"github.com/dezhishen/now-and-again/backend/pkg/taskkind"
	"github.com/dezhishen/now-and-again/backend/pkg/types"
)

type handler struct{}

func init() {
	taskkind.Register(&handler{})
}

func (handler) Kind() string { return "chain" }

// ─── Input ────────────────────────────────────────────────────────

type chainStepInput struct {
	Name       string `json:"name"`
	Kind       string `json:"kind"`
	GroupID    string `json:"group_id,omitempty"`
	LocationID string `json:"location_id,omitempty"`
}

type chainExtraInput struct {
	Steps []chainStepInput `json:"steps,omitempty"`
}

func parseSteps(extra any) ([]chainStepInput, error) {
	if extra == nil {
		return nil, nil
	}
	data, err := json.Marshal(extra)
	if err != nil {
		return nil, err
	}
	var ce chainExtraInput
	if err := json.Unmarshal(data, &ce); err != nil {
		return nil, err
	}
	return ce.Steps, nil
}

// ─── Lifecycle ────────────────────────────────────────────────────

func (h *handler) SaveExtra(storage taskkind.TaskStorage, root *model.TaskModel, extra any) error {
	steps, err := parseSteps(extra)
	if err != nil {
		return fmt.Errorf("chain: parse extra: %w", err)
	}
	if len(steps) == 0 {
		return nil
	}

	db := storage.DB()
	// Clear old steps and child tasks.
	var oldSteps []ChainStepModel
	db.Where("task_id = ?", root.ID).Find(&oldSteps)
	for _, s := range oldSteps {
		if s.ChildTaskID != "" {
			storage.DeleteNonRootTask(s.ChildTaskID)
		}
	}
	db.Where("task_id = ?", root.ID).Delete(&ChainStepModel{})

	// Create child tasks immediately and link them.
	var prevTaskID string
	for i, s := range steps {
		child := &model.TaskModel{
			FamilyID:     root.FamilyID,
			GroupID:      sql.NullString{String: s.GroupID, Valid: s.GroupID != ""},
			LocationID:   sql.NullString{String: s.LocationID, Valid: s.LocationID != ""},
			ParentTaskID: sql.NullString{String: root.ID, Valid: true},
			RootTaskID:   root.ID,
			Name:         s.Name,
			ScheduleType: "once",
			ScheduleData: `{"time":"09:00"}`,
			Enabled:      true,
			Kind:         "chain",
			CreatedBy:    root.CreatedBy,
		}
		if i > 0 {
			child.ParentTaskID = sql.NullString{String: prevTaskID, Valid: true}
		}
		if err := storage.CreateNoRootTask(child, nil); err != nil {
			return fmt.Errorf("chain: create child task %d: %w", i, err)
		}
		prevTaskID = child.ID

		step := &ChainStepModel{
			TaskID:      root.ID,
			SortOrder:   i,
			Name:        s.Name,
			Kind:        s.Kind,
			GroupID:     sql.NullString{String: s.GroupID, Valid: s.GroupID != ""},
			LocationID:  sql.NullString{String: s.LocationID, Valid: s.LocationID != ""},
			ChildTaskID: child.ID,
		}
		if err := db.Create(step).Error; err != nil {
			return fmt.Errorf("chain: save step %d: %w", i, err)
		}
	}
	return nil
}

func (h *handler) UpdateExtra(storage taskkind.TaskStorage, task *model.TaskModel, extra any) error {
	return h.SaveExtra(storage, task, extra)
}

func (h *handler) DeleteExtra(storage taskkind.TaskStorage, task *model.TaskModel) error {
	db := storage.DB()
	var steps []ChainStepModel
	db.Where("task_id = ?", task.ID).Find(&steps)
	for _, s := range steps {
		if s.ChildTaskID != "" {
			_ = storage.DeleteNonRootTask(s.ChildTaskID)
		}
	}
	db.Where("task_id = ?", task.ID).Delete(&ChainStepModel{})
	return nil
}

func (h *handler) GetExtra(storage taskkind.TaskStorage, task *model.TaskModel) (any, error) {
	db := storage.DB()
	var steps []ChainStepModel
	db.Where("task_id = ?", task.ID).Order("sort_order ASC").Find(&steps)
	return map[string]any{"steps": steps}, nil
}

// ─── OnTodo ───────────────────────────────────────────────────────

func (h *handler) OnTodo(storage taskkind.TaskStorage, todo *model.TodoModel, extra any) error {
	if todo.Status != string(types.TodoStatusDone) {
		return nil
	}

	rootTaskID := todo.Task.RootTaskID
	if rootTaskID == "" {
		rootTaskID = todo.Task.ID
	}

	db := storage.DB()

	// Find next step: if this is the root, find first step (SortOrder=0).
	// Otherwise, find the current step's sibling (SortOrder+1).
	var nextStep ChainStepModel
	if todo.Task.IsRoot {
		if err := db.Where("task_id = ? AND sort_order = ?", rootTaskID, 0).First(&nextStep).Error; err != nil {
			return nil // no steps defined
		}
	} else {
		var curStep ChainStepModel
		if err := db.Where("child_task_id = ?", todo.Task.ID).First(&curStep).Error; err != nil {
			return nil // not a chain step (e.g., old data)
		}
		err := db.Where("task_id = ? AND sort_order = ?", rootTaskID, curStep.SortOrder+1).First(&nextStep).Error
		if err != nil {
			return nil // no more steps
		}
	}

	displaySummary := ""
	if todo.Remark.Valid && todo.Remark.String != "" {
		displaySummary = "上一步备注: " + todo.Remark.String
	}
	if _, err := storage.CreateTodo(nextStep.ChildTaskID, displaySummary); err != nil {
		return fmt.Errorf("chain: create todo: %w", err)
	}
	return nil
}
