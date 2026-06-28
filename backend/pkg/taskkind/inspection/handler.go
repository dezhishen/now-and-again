package inspection

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dezhishen/now-and-again/backend/pkg/model"
	"github.com/dezhishen/now-and-again/backend/pkg/taskkind"
	"github.com/dezhishen/now-and-again/backend/pkg/types"
)

// handler implements taskkind.Handler for inspection tasks.
type handler struct{}

func init() {
	taskkind.Register(&handler{})
}

func (handler) Kind() string { return "inspection" }

type extraData struct {
	CheckItems []types.CheckItemDTO `json:"check_items"`
}

// Lifecycle — called by taskService for every task.
// Only handles extra CRUD; task fields (e.g. display_summary) are the
// caller's responsibility.
func (h *handler) OnCreate(taskStorage taskkind.TaskStorage, task *model.TaskModel, extra any) error {
	items, err := parseCheckItems(extra)
	if err != nil {
		return fmt.Errorf("parse check items: %w", err)
	}
	return h.saveCheckItems(taskStorage, task, items)
}

func (h *handler) OnUpdate(taskStorage taskkind.TaskStorage, task *model.TaskModel, extra any) error {
	db := taskStorage.DB()
	// Delete child tasks referenced by branch_task_id
	db.Where("id IN (?)",
		db.Model(&CheckItemBranchModel{}).Select("branch_task_id").
			Where("branch_task_id != '' AND check_item_id IN (?)",
				db.Model(&CheckItemModel{}).Select("id").Where("task_id = ?", task.ID),
			),
	).Delete(&model.TaskModel{})
	// Delete old branches and check items
	if err := NewCheckItemBranchRepo(db).DeleteCheckItemBranchesByTask(task.ID); err != nil {
		return fmt.Errorf("delete old branches: %w", err)
	}
	if err := NewCheckItemRepo(db).DeleteCheckItemsByTask(task.ID); err != nil {
		return fmt.Errorf("delete old check items: %w", err)
	}
	items, err := parseCheckItems(extra)
	if err != nil {
		return fmt.Errorf("parse check items: %w", err)
	}
	return h.saveCheckItems(taskStorage, task, items)
}

func (h *handler) OnDelete(taskStorage taskkind.TaskStorage, task *model.TaskModel) error {
	db := taskStorage.DB()
	// Delete child tasks referenced by branch_task_id
	db.Where("id IN (?)",
		db.Model(&CheckItemBranchModel{}).Select("branch_task_id").
			Where("branch_task_id != '' AND check_item_id IN (?)",
				db.Model(&CheckItemModel{}).Select("id").Where("task_id = ?", task.ID),
			),
	).Delete(&model.TaskModel{})
	if err := NewCheckItemBranchRepo(db).DeleteCheckItemBranchesByTask(task.ID); err != nil {
		return fmt.Errorf("delete branches: %w", err)
	}
	if err := NewCheckItemRepo(db).DeleteCheckItemsByTask(task.ID); err != nil {
		return fmt.Errorf("delete check items: %w", err)
	}
	return nil
}

func (h *handler) OnComplete(taskStorage taskkind.TaskStorage, todo *model.TodoModel, extra any) error {
	selections, err := parseSelections(extra)
	if err != nil {
		return fmt.Errorf("parse extra: %w", err)
	}
	if len(selections) == 0 {
		return nil
	}

	// Load check items with branches for this task
	checkItems, err := NewCheckItemRepo(taskStorage.DB()).FindCheckItemsByTask(todo.TaskID)
	if err != nil {
		return fmt.Errorf("load check items: %w", err)
	}

	var details []string
	for _, sel := range selections {
		ci := findCheckItemByID(checkItems, sel.ItemID)
		if ci == nil {
			continue
		}
		branch := findBranchByID(ci.Branches, sel.BranchID)
		if branch == nil {
			continue
		}

		// Record inspection result
		itemName := sel.ItemName
		if itemName == "" {
			itemName = ci.Name
		}
		branchName := sel.BranchName
		if branchName == "" {
			branchName = branch.Name
		}
		details = append(details, itemName+" → "+branchName)

		result := &InspectionResultModel{
			TaskID:     todo.TaskID,
			TodoID:     todo.ID,
			FamilyID:   todo.FamilyID,
			ItemName:   itemName,
			BranchName: branchName,
			CreatedBy:  todo.CompletedBy,
		}
		NewCheckItemRepo(taskStorage.DB()).CreateInspectionResult(result)

		// If branch has create_todo, spawn or trigger child task
		if branch.CreateTodo {
			h.ensureBranchTask(taskStorage, todo, branch)
		}
	}

	// Write a detailed log so users can trace why branch todos were generated.
	if len(details) > 0 {
		taskStorage.DB().Create(&model.TaskLogModel{
			TaskID:     todo.TaskID,
			TodoID:     todo.ID,
			Status:     "done",
			Message:    fmt.Sprintf("巡检结果: %s", strings.Join(details, ", ")),
			LogType:    "user",
			OperatorID: todo.CompletedBy,
		})
	}

	return nil
}

// GetExtra returns check_items + children for the detail view.
func (h *handler) GetExtra(taskStorage taskkind.TaskStorage, task *model.TaskModel) (any, error) {
	extraData, err := h.getExtraData(taskStorage, task)
	if err != nil {
		return nil, fmt.Errorf("get extra data: %w", err)
	}
	return extraData, nil
}

// GetExtra returns kind-specific data for the task detail page.
// e.g. for inspection: check_items + children

func (h *handler) getExtraData(taskStorage taskkind.TaskStorage, task *model.TaskModel) (*extraData, error) {
	// Load check items and branches
	checkItems, error := NewCheckItemRepo(taskStorage.DB()).FindCheckItemsByTask(task.ID)
	if error != nil {
		return nil, fmt.Errorf("load check items: %w", error)
	}
	extraData := &extraData{CheckItems: make([]types.CheckItemDTO, 0, len(checkItems))}
	for _, ci := range checkItems {
		branches := make([]types.CheckItemBranchDTO, 0, len(ci.Branches))
		for _, b := range ci.Branches {
			dto := types.CheckItemBranchDTO{
				ID:           b.ID,
				Name:         b.Name,
				CreateTodo:   b.CreateTodo,
				BranchTaskID: b.BranchTaskID,
				SortOrder:    b.SortOrder,
			}
			// FindTheTaskOfBranche
			if b.BranchTaskID != "" {
				branchTask, err := taskStorage.FindTaskByID(b.BranchTaskID)
				if err == nil && branchTask != nil {
					dto.BranchTask = &types.TaskWithExtra{
						Task: types.TaskFromModel(branchTask),
					}
				}
			}

			branches = append(branches, dto)
		}
		extraData.CheckItems = append(extraData.CheckItems, types.CheckItemDTO{
			ID: ci.ID, Name: ci.Name, SortOrder: ci.SortOrder,
			Branches: branches,
		})
	}
	return extraData, nil
}

// ─── Helpers ─────────────────────────────────────────────────────

func parseCheckItems(extra any) ([]types.CheckItemDTO, error) {
	if extra == nil {
		return nil, nil
	}
	data, err := json.Marshal(extra)
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		CheckItems []types.CheckItemDTO `json:"check_items"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}
	return wrapper.CheckItems, nil
}

func (h *handler) saveCheckItems(taskStorage taskkind.TaskStorage, task *model.TaskModel, items []types.CheckItemDTO) error {
	// Use the same DB (possibly transactional) as the task repo.
	db := taskStorage.DB()
	checkItemRepo := NewCheckItemRepo(db)
	checkItemBranchRepo := NewCheckItemBranchRepo(db)

	for i, item := range items {
		ci := &CheckItemModel{
			TaskID:    task.ID,
			Name:      item.Name,
			SortOrder: i,
		}
		if err := checkItemRepo.CreateCheckItem(ci); err != nil {
			return fmt.Errorf("create check item %s: %w", item.Name, err)
		}
		for j, b := range item.Branches {
			branch := &CheckItemBranchModel{
				CheckItemID: ci.ID,
				Name:        b.Name,
				CreateTodo:  b.CreateTodo,
				SortOrder:   j,
			}
			// Create child task if create_todo is set and task params provided.
			if b.CreateTodo && b.BranchTask != nil && b.BranchTask.Task != nil && b.BranchTask.Task.Name != "" {
				kind := b.BranchTask.Task.Kind
				if kind == "" {
					kind = "simple"
				}
				scheduleType := b.BranchTask.Task.ScheduleType
				if scheduleType == "" {
					scheduleType = "once"
				}
				childTask := &model.TaskModel{
					FamilyID:     task.FamilyID,
					GroupID:      b.BranchTask.Task.GroupID,
					LocationID:   b.BranchTask.Task.LocationID,
					ParentTaskID: task.ID,
					IsRoot:       false,
					Name:         b.BranchTask.Task.Name,
					ScheduleType: scheduleType,
					Enabled:      true,
					Kind:         kind,
					CreatedBy:    task.CreatedBy,
				}
				if err := taskStorage.CreateNoRootTask(childTask); err != nil {
					return fmt.Errorf("create branch task %s: %w", b.BranchTask.Task.Name, err)
				}
				branch.BranchTaskID = childTask.ID
			}
			if err := checkItemBranchRepo.CreateCheckItemBranch(branch); err != nil {
				return fmt.Errorf("create branch %s: %w", b.Name, err)
			}
		}
	}
	return nil
}

func parseSelections(extra any) ([]taskkind.Selection, error) {
	if extra == nil {
		return nil, nil
	}
	data, err := json.Marshal(extra)
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		Selections []taskkind.Selection `json:"selections"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}
	return wrapper.Selections, nil
}

func findCheckItemByID(items []CheckItemModel, id string) *CheckItemModel {
	for i := range items {
		if items[i].ID == id {
			return &items[i]
		}
	}
	return nil
}

func findBranchByID(branches []CheckItemBranchModel, id string) *CheckItemBranchModel {
	for i := range branches {
		if branches[i].ID == id {
			return &branches[i]
		}
	}
	return nil
}

func findCheckItemByName(items []CheckItemModel, name string) *CheckItemModel {
	for i := range items {
		if items[i].Name == name {
			return &items[i]
		}
	}
	return nil
}

func findBranchByName(branches []CheckItemBranchModel, name string) *CheckItemBranchModel {
	for i := range branches {
		if branches[i].Name == name {
			return &branches[i]
		}
	}
	return nil
}

func (h *handler) ensureBranchTask(taskStorage taskkind.TaskStorage, todo *model.TodoModel, branch *CheckItemBranchModel) {
	if branch.BranchTaskID == "" {
		return
	}
	branchTask, err := taskStorage.FindTaskByID(branch.BranchTaskID)
	if err != nil || branchTask == nil {
		return
	}
	now := time.Now()
	db := taskStorage.DB()
	branchTodo := &model.TodoModel{
		TaskID:     branchTask.ID,
		FamilyID:   branchTask.FamilyID,
		LocationID: branchTask.LocationID,
		DueStart:   now,
		DueDate:    now.Add(24 * time.Hour),
		Status:     "pending",
	}
	if err := db.Create(branchTodo).Error; err != nil {
		return
	}
	// Log the auto-generated branch todo so it appears in task logs.
	db.Create(&model.TaskLogModel{
		TaskID:  branchTask.ID,
		TodoID:  branchTodo.ID,
		Status:  "generated",
		Message: fmt.Sprintf("巡检分支「%s」自动生成待办: %s", branch.Name, branchTask.Name),
		LogType: "system",
	})
}
