package inspection

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dezhishen/now-and-again/backend/pkg/model"
	"github.com/dezhishen/now-and-again/backend/pkg/taskkind"
	"github.com/dezhishen/now-and-again/backend/pkg/types"
	"gorm.io/gorm"
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
func (h *handler) SaveExtra(taskStorage taskkind.TaskStorage, task *model.TaskModel, extra any) error {
	items, err := parseCheckItems(extra)
	if err != nil {
		return fmt.Errorf("parse check items: %w", err)
	}
	return h.saveCheckItems(taskStorage, task, items)
}

func (h *handler) UpdateExtra(taskStorage taskkind.TaskStorage, task *model.TaskModel, extra any) error {
	if extra == nil {
		return nil // field-only update, keep existing extra data
	}
	items, err := parseCheckItems(extra)
	if err != nil {
		return fmt.Errorf("parse check items: %w", err)
	}

	db := taskStorage.DB()
	checkItemRepo := NewCheckItemRepo(db)
	checkItemBranchRepo := NewCheckItemBranchRepo(db)

	// Load existing items with branches
	oldItems, err := checkItemRepo.FindCheckItemsByTask(task.ID)
	if err != nil {
		return fmt.Errorf("load existing items: %w", err)
	}

	// Build maps of old data by ID
	oldItemMap := make(map[string]CheckItemModel, len(oldItems))
	oldBranchMap := make(map[string]CheckItemBranchModel)
	for _, oi := range oldItems {
		oldItemMap[oi.ID] = oi
		for _, ob := range oi.Branches {
			oldBranchMap[ob.ID] = ob
		}
	}

	keptItemIDs := make(map[string]bool)
	keptBranchIDs := make(map[string]bool)

	for i, item := range items {
		if item.ID != "" {
			if old, ok := oldItemMap[item.ID]; ok {
				// Update existing item
				keptItemIDs[item.ID] = true
				if old.Name != item.Name || old.SortOrder != i {
					checkItemRepo.UpdateCheckItem(&CheckItemModel{
						BaseModel: model.BaseModel{ID: item.ID},
						Name:      item.Name,
						SortOrder: i,
					})
				}
				// Granular branch diff
				if err := h.updateBranches(taskStorage, db, checkItemBranchRepo, item.ID, old.Branches, item.Branches, &keptBranchIDs, task); err != nil {
					return fmt.Errorf("update branches for item %s: %w", item.ID, err)
				}
				continue
			}
		}
		// New item → create with all branches
		ci := &CheckItemModel{TaskID: task.ID, Name: item.Name, SortOrder: i}
		if err := checkItemRepo.CreateCheckItem(ci); err != nil {
			return fmt.Errorf("create check item: %w", err)
		}
		keptItemIDs[ci.ID] = true
		for j, b := range item.Branches {
			if err := h.createBranch(taskStorage, checkItemBranchRepo, ci.ID, j, b, task); err != nil {
				return fmt.Errorf("create branch: %w", err)
			}
		}
	}

	// Delete old items not in new data
	for _, oi := range oldItems {
		if keptItemIDs[oi.ID] {
			continue
		}
		// Delete branch tasks and branches first
		for _, ob := range oi.Branches {
			if ob.BranchTaskID != "" {
				taskStorage.DeleteNonRootTask(ob.BranchTaskID)
			}
			checkItemBranchRepo.DeleteCheckItemBranch(ob.ID)
		}
		checkItemRepo.DeleteCheckItem(oi.ID)
	}

	return nil
}

// updateBranches diffs old vs new branches for a single check item.
func (h *handler) updateBranches(
	taskStorage taskkind.TaskStorage,
	db *gorm.DB,
	branchRepo *CheckItemBranchRepo,
	itemID string,
	oldBranches []CheckItemBranchModel,
	newBranches []types.CheckItemBranchDTO,
	keptBranchIDs *map[string]bool,
	parentTask *model.TaskModel,
) error {
	oldMap := make(map[string]CheckItemBranchModel, len(oldBranches))
	for _, ob := range oldBranches {
		oldMap[ob.ID] = ob
	}

	for j, nb := range newBranches {
		if nb.ID != "" {
			if old, ok := oldMap[nb.ID]; ok {
				// Update existing branch
				(*keptBranchIDs)[nb.ID] = true
				branchTaskID := old.BranchTaskID
				// Handle child task: create/update/delete based on create_todo
				if nb.CreateTodo && nb.BranchTask != nil && nb.BranchTask.Task != nil && nb.BranchTask.Task.Name != "" {
					if branchTaskID != "" {
						// Update existing child task (may delete+recreate if kind changed)
						newID, err := h.updateChildTask(taskStorage, branchTaskID, nb.BranchTask)
						if err != nil {
							return err
						}
						branchTaskID = newID
					} else {
						// Create new child task
						branchTaskID = h.createChildTask(taskStorage, nb.BranchTask, parentTask)
					}
				} else if !nb.CreateTodo && branchTaskID != "" {
					// create_todo was unchecked → delete child task
					taskStorage.DeleteNonRootTask(branchTaskID)
					branchTaskID = ""
				}
				branchRepo.UpdateCheckItemBranch(&CheckItemBranchModel{
					BaseModel:    model.BaseModel{ID: nb.ID},
					Name:         nb.Name,
					CreateTodo:   nb.CreateTodo,
					BranchTaskID: branchTaskID,
					SortOrder:    j,
				})
				continue
			}
		}
		// New branch → create
		if err := h.createBranch(taskStorage, branchRepo, itemID, j, nb, parentTask); err != nil {
			return err
		}
	}

	// Delete old branches not in new data
	for _, ob := range oldBranches {
		if (*keptBranchIDs)[ob.ID] {
			continue
		}
		if ob.BranchTaskID != "" {
			taskStorage.DeleteNonRootTask(ob.BranchTaskID)
		}
		branchRepo.DeleteCheckItemBranch(ob.ID)
	}
	return nil
}

// createBranch creates a new branch (and its child task if create_todo).
// Returns error so callers can detect DB failures instead of silently losing data.
func (h *handler) createBranch(taskStorage taskkind.TaskStorage, branchRepo *CheckItemBranchRepo, itemID string, sortOrder int, b types.CheckItemBranchDTO, parentTask *model.TaskModel) error {
	branch := &CheckItemBranchModel{
		CheckItemID: itemID,
		Name:        b.Name,
		CreateTodo:  b.CreateTodo,
		SortOrder:   sortOrder,
	}
	if b.CreateTodo && b.BranchTask != nil && b.BranchTask.Task != nil && b.BranchTask.Task.Name != "" {
		branch.BranchTaskID = h.createChildTask(taskStorage, b.BranchTask, parentTask)
	}
	return branchRepo.CreateCheckItemBranch(branch)
}

// createChildTask creates a non-root task and returns its ID.
func (h *handler) createChildTask(taskStorage taskkind.TaskStorage, bt *types.TaskWithExtra, parent *model.TaskModel) string {
	kind := bt.Task.Kind
	if kind == "" {
		kind = "simple"
	}
	scheduleType := bt.Task.ScheduleType
	if scheduleType == "" {
		scheduleType = "once"
	}
	dataJSON, _ := json.Marshal(bt.Task.ScheduleData)
	child := &model.TaskModel{
		FamilyID:     parent.FamilyID,
		GroupID:      bt.Task.GroupID,
		LocationID:   bt.Task.LocationID,
		ParentTaskID: parent.ID,
		RootTaskID:   parent.RootTaskID,
		Name:         bt.Task.Name,
		ScheduleType: scheduleType,
		ScheduleData: string(dataJSON),
		Enabled:      true,
		Kind:         kind,
		CreatedBy:    parent.CreatedBy,
	}
	taskStorage.CreateNoRootTask(child, bt.Extra)
	return child.ID
}

// updateChildTask updates an existing child task. If the task kind changed,
// it deletes the old task and creates a new one so the new kind's SaveExtra runs.
func (h *handler) updateChildTask(taskStorage taskkind.TaskStorage, taskID string, bt *types.TaskWithExtra) (string, error) {
	child, err := taskStorage.FindTaskByID(taskID)
	if err != nil || child == nil {
		return "", err
	}

	newKind := bt.Task.Kind
	if newKind == "" {
		newKind = "simple"
	}

	// Kind changed → delete old and create new to trigger proper lifecycle
	if child.Kind != newKind {
		if err := taskStorage.DeleteNonRootTask(taskID); err != nil {
			return "", fmt.Errorf("delete old child task: %w", err)
		}
		parent, _ := taskStorage.FindTaskByID(child.ParentTaskID)
		if parent == nil {
			parent = &model.TaskModel{FamilyID: child.FamilyID, RootTaskID: child.RootTaskID}
		}
		return h.createChildTask(taskStorage, bt, parent), nil
	}

	// Same kind — field update with extra data
	if bt.Task.Name != "" {
		child.Name = bt.Task.Name
	}
	if bt.Task.ScheduleType != "" {
		child.ScheduleType = bt.Task.ScheduleType
	}
	if bt.Task.ScheduleData != nil {
		dataJSON, _ := json.Marshal(bt.Task.ScheduleData)
		child.ScheduleData = string(dataJSON)
	}
	if bt.Task.GroupID != "" {
		child.GroupID = bt.Task.GroupID
	}
	if bt.Task.LocationID != "" {
		child.LocationID = bt.Task.LocationID
	}
	child.Enabled = true
	if err := taskStorage.UpdateNoRootTask(child, bt.Extra); err != nil {
		return "", fmt.Errorf("update child task: %w", err)
	}
	return taskID, nil
}

func (h *handler) DeleteExtra(taskStorage taskkind.TaskStorage, task *model.TaskModel) error {
	db := taskStorage.DB()
	// Delete child tasks via full lifecycle (triggers their DeleteExtra recursively).
	var childIDs []string
	db.Model(&CheckItemBranchModel{}).
		Select("branch_task_id").
		Where("branch_task_id != '' AND check_item_id IN (?)",
			db.Model(&CheckItemModel{}).Select("id").Where("task_id = ?", task.ID),
		).Pluck("branch_task_id", &childIDs)
	for _, id := range childIDs {
		if err := taskStorage.DeleteNonRootTask(id); err != nil {
			return fmt.Errorf("delete child task %s: %w", id, err)
		}
	}
	// Clean up plugin-specific data.
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
				branch.BranchTaskID = h.createChildTask(taskStorage, b.BranchTask, task)
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
	branchTodo, err := taskStorage.CreateTodo(branchTask.ID, todo.Remark)
	if err != nil {
		return
	}
	// Log the auto-generated branch todo so it appears in task logs.
	taskStorage.DB().Create(&model.TaskLogModel{
		TaskID:  branchTask.ID,
		TodoID:  branchTodo.ID,
		Status:  "generated",
		Message: fmt.Sprintf("巡检分支「%s」自动生成待办: %s", branch.Name, branchTask.Name),
		LogType: "system",
	})
}
