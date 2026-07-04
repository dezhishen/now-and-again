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
	Name       string          `json:"name"`
	Kind       string          `json:"kind"`
	GroupID    string          `json:"group_id,omitempty"`
	LocationID string          `json:"location_id,omitempty"`
	Extra      json.RawMessage `json:"extra,omitempty"`
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
		kind := s.Kind
		if kind == "" {
			kind = "simple"
		}
		child := &model.TaskModel{
			FamilyID:      root.FamilyID,
			GroupID:       sql.NullString{String: s.GroupID, Valid: s.GroupID != ""},
			LocationID:    sql.NullString{String: s.LocationID, Valid: s.LocationID != ""},
			ParentTaskID:  sql.NullString{String: root.ID, Valid: true},
			RootTaskID:    root.ID,
			Name:          s.Name,
			ScheduleType:  "once",
			ScheduleData:  `{"time":"09:00"}`,
			Enabled:       true,
			Kind:          kind,
			CreatedByKind: "chain",
			CreatedBy:     root.CreatedBy,
		}
		if i > 0 {
			child.ParentTaskID = sql.NullString{String: prevTaskID, Valid: true}
		}
		if err := storage.CreateNoRootTask(child, s.Extra); err != nil {
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

	// Ensure root task dispatches to this handler on todo completion (not "simple").
	if root.CreatedByKind == "" || root.CreatedByKind == "simple" {
		root.CreatedByKind = "chain"
		_ = storage.UpdateTaskFields(root)
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

// ─── Extra I/O ────────────────────────────────────────────────────

// stepExtra is the clean JSON output for each chain step (no sql.NullString, snake_case).
type stepExtra struct {
	ID          string `json:"id"`
	TaskID      string `json:"task_id"`
	SortOrder   int    `json:"sort_order"`
	Name        string `json:"name"`
	Kind        string `json:"kind"`
	GroupID     string `json:"group_id,omitempty"`
	LocationID  string `json:"location_id,omitempty"`
	ChildTaskID string `json:"child_task_id"`
	Extra       any    `json:"extra,omitempty"`
}

func (h *handler) GetExtra(storage taskkind.TaskStorage, task *model.TaskModel) (any, error) {
	db := storage.DB()
	var rows []ChainStepModel
	db.Where("task_id = ?", task.ID).Order("sort_order ASC").Find(&rows)

	steps := make([]stepExtra, len(rows))
	for i, s := range rows {
		steps[i] = stepExtra{
			ID:          s.ID,
			TaskID:      s.TaskID,
			SortOrder:   s.SortOrder,
			Name:        s.Name,
			Kind:        s.Kind,
			ChildTaskID: s.ChildTaskID,
		}
		if s.GroupID.Valid {
			steps[i].GroupID = s.GroupID.String
		}
		if s.LocationID.Valid {
			steps[i].LocationID = s.LocationID.String
		}

		// Recursively query child task's extra via its own handler.
		if s.ChildTaskID != "" {
			child, err := storage.FindTaskByID(s.ChildTaskID)
			if err == nil && child != nil {
				if childH := storage.LookupHandler(child.Kind); childH != nil {
					steps[i].Extra, _ = childH.GetExtra(storage, child)
				}
			}
		}
	}
	return map[string]any{"steps": steps}, nil
}

// ─── OnTodo ───────────────────────────────────────────────────────

func (h *handler) OnTodo(storage taskkind.TaskStorage, todo *model.TodoModel, extra any) error {
	if todo.Status != string(types.TodoStatusDone) {
		return nil
	}

	// 1. If this is a non-root chain step, delegate to the real kind handler first.
	//    e.g., inspection step → inspection.OnTodo records results, spawns branch tasks.
	if !todo.Task.IsRoot {
		if realHandler := storage.LookupHandler(todo.Task.Kind); realHandler != nil {
			if err := realHandler.OnTodo(storage, todo, extra); err != nil {
				return err
			}
		}
	}

	// 2. Chain progression: create todo for the next step.
	rootTaskID := todo.Task.RootTaskID
	if rootTaskID == "" {
		rootTaskID = todo.Task.ID
	}

	db := storage.DB()

	var nextStep ChainStepModel
	if todo.Task.IsRoot {
		if err := db.Where("task_id = ? AND sort_order = ?", rootTaskID, 0).First(&nextStep).Error; err != nil {
			return nil // no steps defined
		}
	} else {
		var curStep ChainStepModel
		if err := db.Where("child_task_id = ?", todo.Task.ID).First(&curStep).Error; err != nil {
			return nil // not a chain step
		}
		err := db.Where("task_id = ? AND sort_order = ?", rootTaskID, curStep.SortOrder+1).First(&nextStep).Error
		if err != nil {
			return nil // no more steps — chain complete
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
