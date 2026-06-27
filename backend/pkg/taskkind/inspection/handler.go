package inspection

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dezhishen/now-and-again/backend/internal/repository"
	"github.com/dezhishen/now-and-again/backend/pkg/scheduler"
	"github.com/dezhishen/now-and-again/backend/pkg/taskkind"
	"github.com/dezhishen/now-and-again/backend/pkg/types"
)

// Handler implements taskkind.Handler and taskkind.Inspector for inspection tasks.
type Handler struct{}

func (Handler) Kind() string { return "inspection" }

// GetExtra returns check_items + children for the detail view.
func (Handler) GetExtra(ops *taskkind.Ops, task *repository.TaskTemplateModel) (any, error) {
	// Reload with full relations
	full, err := ops.Repo.FindTaskFull(task.ID)
	if err != nil {
		return nil, err
	}
	extra := struct {
		CheckItems []types.CheckItemDTO `json:"check_items"`
		Children   []types.TaskTemplate `json:"children"`
	}{}
	for _, ci := range full.CheckItems {
		branches := make([]types.CheckItemBranchDTO, 0, len(ci.Branches))
		for _, b := range ci.Branches {
			dto := types.CheckItemBranchDTO{
				ID: b.ID, Name: b.Name,
				CreateTodo:   b.CreateTodo,
				BranchTaskID: b.BranchTaskID,
				SortOrder:    b.SortOrder,
			}
			if b.BranchTask != nil {
				dto.TodoName = b.BranchTask.Name
				dto.LocationID = b.BranchTask.LocationID
			}
			branches = append(branches, dto)
		}
		extra.CheckItems = append(extra.CheckItems, types.CheckItemDTO{
			ID: ci.ID, Name: ci.Name, SortOrder: ci.SortOrder,
			Branches: branches,
		})
	}
	for _, c := range full.Children {
		var data any
		json.Unmarshal([]byte(c.ScheduleData), &data)
		extra.Children = append(extra.Children, types.TaskTemplate{
			ID: c.ID, Name: c.Name, Kind: c.Kind,
			ParentTaskID: c.ParentTaskID, IsRoot: c.IsRoot,
			ScheduleType: c.ScheduleType, ScheduleData: data,
			Enabled: c.Enabled,
		})
	}
	return extra, nil
}

func (Handler) OnComplete(ops *taskkind.Ops, todo *repository.TodoModel, extra any, branchName, userID string) error {
	// Try different types that might carry inspection selections
	switch s := extra.(type) {
	case []taskkind.Selection:
		if len(s) == 0 {
			return nil
		}
		return doInspect(ops, todo, s, userID)
	case []types.InspectionSelection:
		if len(s) == 0 {
			return nil
		}
		selections := make([]taskkind.Selection, len(s))
		for i, sel := range s {
			selections[i] = taskkind.Selection{Item: sel.Item, Branch: sel.Branch}
		}
		return doInspect(ops, todo, selections, userID)
	}
	return nil
}

// OnInspect handles the multi-item inspection submission (called via the dedicated endpoint).
func (Handler) OnInspect(ops *taskkind.Ops, todo *repository.TodoModel, selections []taskkind.Selection, userID string) error {
	return doInspect(ops, todo, selections, userID)
}

func doInspect(ops *taskkind.Ops, todo *repository.TodoModel, selections []taskkind.Selection, userID string) error {
	for _, sel := range selections {
		// Persist to inspection_results table
		result := &repository.InspectionResultModel{
			TaskID:     todo.TaskID,
			TodoID:     todo.ID,
			FamilyID:   todo.FamilyID,
			ItemName:   sel.Item,
			BranchName: sel.Branch,
			CreatedBy:  userID,
		}
		if err := ops.Repo.CreateInspectionResult(result); err != nil {
			return fmt.Errorf("save inspection result: %w", err)
		}

		ops.Repo.CreateUserLog(todo.TaskID, todo.ID, userID, "inspection",
			fmt.Sprintf("[%s] → %s", sel.Item, sel.Branch))

		// Find the pre-created branch task
		branchTask, err := ops.Repo.FindBranchTask(todo.TaskID, sel.Item, sel.Branch)
		if err != nil || branchTask == nil || branchTask.ID == "" {
			continue // no follow-up task needed (e.g. "正常" branch)
		}

		// Enable the branch task
		branchTask.Enabled = true
		if err := ops.Repo.UpdateTask(branchTask); err != nil {
			return fmt.Errorf("enable branch task: %w", err)
		}

		// Create todo immediately
		followTodo := &repository.TodoModel{
			TaskID:   branchTask.ID,
			FamilyID: todo.FamilyID,
			Status:   "pending",
			DueStart: time.Now(),
			DueDate:  time.Now().Add(24 * time.Hour),
		}
		if err := ops.Repo.CreateTodo(followTodo); err != nil {
			return fmt.Errorf("create follow-up todo: %w", err)
		}

		// Register with scheduler
		ops.Scheduler.RegisterJob(&scheduler.JobBuilder{
			TaskID:       branchTask.ID,
			ScheduleType: "once",
			ScheduleData: branchTask.ScheduleData,
			Callback:     func(taskID string, triggeredAt time.Time) error { return nil },
			AfterFire:    func(taskID string) { ops.Repo.DisableTask(taskID) },
		})

		ops.Repo.CreateUserLog(todo.TaskID, todo.ID, userID, "follow_up",
			fmt.Sprintf("巡检跟进 → 启用任务「%s」(%s)", branchTask.Name, branchTask.ID))
	}

	ops.Repo.CreateUserLog(todo.TaskID, todo.ID, userID, "inspection",
		fmt.Sprintf("巡检完成: %s", todo.Task.Name))

	// Disable one-shot inspection after completion
	if todo.Task.ScheduleType == "once" {
		ops.Repo.DisableTask(todo.TaskID)
		ops.Scheduler.RemoveJob(todo.TaskID)
	}
	return nil
}

func init() {
	taskkind.Register(Handler{})
}

// OnCreate persists check_items, branches, and child tasks for a new inspection.
func (Handler) OnCreate(ops *taskkind.Ops, task *repository.TaskTemplateModel, extra any) error {
	return persistCheckItems(ops, task, extra)
}

// OnUpdate clears old sub-tables and recreates from the updated check items.
func (Handler) OnUpdate(ops *taskkind.Ops, task *repository.TaskTemplateModel, extra any) error {
	ops.Repo.DeleteChildren(task.ID)
	ops.Repo.DeleteCheckItemsByTask(task.ID)
	return persistCheckItems(ops, task, extra)
}

func persistCheckItems(ops *taskkind.Ops, task *repository.TaskTemplateModel, extra any) error {
	// Extract check_items from the generic extra map
	var items []types.CheckItemDTO
	switch e := extra.(type) {
	case map[string]interface{}:
		if raw, ok := e["check_items"]; ok {
			// Re-marshal through JSON for type conversion
			b, _ := json.Marshal(raw)
			json.Unmarshal(b, &items)
		}
	case []types.CheckItemDTO:
		items = e
	case []interface{}:
		b, _ := json.Marshal(e)
		json.Unmarshal(b, &items)
	}
	if len(items) == 0 {
		// Update summary for empty case
		ops.Repo.UpdateDisplaySummary(task.ID, "")
		return nil
	}

	// Build display summary: "3个检查项"
	totalBranches := 0
	for _, dto := range items {
		totalBranches += len(dto.Branches)
	}
	summary := fmt.Sprintf("%d个检查项", len(items))
	ops.Repo.UpdateDisplaySummary(task.ID, summary)

	for _, dto := range items {
		item := &repository.CheckItemModel{
			TaskID:    task.ID,
			Name:      dto.Name,
			SortOrder: dto.SortOrder,
		}
		if err := ops.Repo.CreateCheckItem(item); err != nil {
			return fmt.Errorf("create check item: %w", err)
		}
		for _, b := range dto.Branches {
			var branchTaskID string
			if b.CreateTodo {
				todoName := b.TodoName
				if todoName == "" {
					todoName = dto.Name + " - " + b.Name
				}
				child := &repository.TaskTemplateModel{
					FamilyID:     task.FamilyID,
					IsRoot:       false,
					ParentTaskID: task.ID,
					LocationID:   b.LocationID,
					Name:         todoName,
					ScheduleType: "once",
					ScheduleData: `{"time":"00:00"}`,
					Enabled:      false,
					Kind:         "simple",
					CreatedBy:    task.CreatedBy,
				}
				if err := ops.Repo.CreateTask(child); err != nil {
					return fmt.Errorf("create branch task: %w", err)
				}
				branchTaskID = child.ID
			}
			cb := &repository.CheckItemBranchModel{
				CheckItemID:  item.ID,
				Name:         b.Name,
				CreateTodo:   b.CreateTodo,
				BranchTaskID: branchTaskID,
				SortOrder:    b.SortOrder,
			}
			if err := ops.Repo.CreateCheckItemBranch(cb); err != nil {
				return fmt.Errorf("create check item branch: %w", err)
			}
		}
	}
	return nil
}

// OnDelete cleans up check_items, branches, and child tasks when an inspection is deleted.
func (Handler) OnDelete(ops *taskkind.Ops, task *repository.TaskTemplateModel) error {
	if err := ops.Repo.DeleteChildren(task.ID); err != nil {
		return fmt.Errorf("delete children: %w", err)
	}
	if err := ops.Repo.DeleteCheckItemsByTask(task.ID); err != nil {
		return fmt.Errorf("delete check items: %w", err)
	}
	return nil
}
