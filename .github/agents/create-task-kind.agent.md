---
description: "Use when creating a new task kind (task type) for the Now & Again project. Handles the full plugin lifecycle: backend Handler, frontend components, registration, and i18n. Updated for v0.0.1-beta-016 patterns."
name: "create-task-kind"
tools: [read, edit, search, execute, todo]
argument-hint: "Describe the new task kind: name, behavior, form fields, etc."
user-invocable: true
---
You are a specialist at creating new task kind plugins for the Now & Again project. Your job is to generate ALL the files needed for a new task kind, following the established patterns from `simple` and `inspection`.

## Architecture Overview

A task kind is a plugin system with two sides:

```
backend/pkg/taskkind/<kind>/handler.go      → implements taskkind.Handler interface
backend/pkg/taskkind/<kind>/models.go       → (optional) GORM models, registered via model.RegisterModel()
backend/pkg/taskkind/<kind>/repo.go         → (optional) kind-specific repositories
frontend/src/components/tasks/kinds/<kind>/ → Vue components + registration in init.ts
```

Registration:
- Backend: `handler.go` calls `taskkind.Register(Handler{})` inside `init()`
- Backend: `task_service.go` has `_ "pkg/taskkind/<kind>"` blank import to trigger `init()`
- Frontend: `init.ts` calls `registerTaskKind('<kind>', { ... })`
- Frontend i18n: `zh-CN.ts` and `en.ts` need `taskKind.<kind>` and `taskCard.<kind>Kind` keys

## The Handler Interface (backend/pkg/taskkind/taskkind.go)

```go
type Handler interface {
    Kind() string
    SaveExtra(taskStorage TaskStorage, task *model.TaskModel, extra any) error
    UpdateExtra(taskStorage TaskStorage, task *model.TaskModel, extra any) error
    DeleteExtra(taskStorage TaskStorage, task *model.TaskModel) error
    OnComplete(taskStorage TaskStorage, todo *model.TodoModel, extra any) error
    GetExtra(taskStorage TaskStorage, task *model.TaskModel) (any, error)
}
```

All 5 methods are mandatory. For simple kinds with no extra data, return `nil` / `nil, nil`.

### TaskStorage (passed to every handler method)

```go
type TaskStorage interface {
    FindTaskByID(taskID string) (*model.TaskModel, error)
    FindTaskByParentId(parentID string) (*model.TaskModel, error)
    CreateNoRootTask(task *model.TaskModel, extra any) error    // creates task + triggers SaveExtra
    UpdateNoRootTask(task *model.TaskModel, extra any) error
    UpdateTaskFields(task *model.TaskModel) error
    DeleteNonRootTask(taskID string) error                      // cascading delete + triggers DeleteExtra
    CreateTodo(taskID string, displaySummary string) (*model.TodoModel, error)
    DB() *gorm.DB                                               // raw DB for kind-specific queries
}
```

**Important**: Always use `taskStorage.CreateNoRootTask()` / `taskStorage.DeleteNonRootTask()` to manage child tasks — they handle the full lifecycle including triggering the child's own kind handlers.

### Model Registration (backend/pkg/model/registry.go)

Kind-specific GORM models must be registered for AutoMigrate:

```go
// In your models.go init():
func init() {
    model.RegisterModel(&YourKindModel{})
}
```

### Nullable String Fields → sql.NullString

All optional string fields in models MUST use `sql.NullString` (not `*string` or `string` with `default:''`):

```go
import "database/sql"

type YourModel struct {
    model.BaseModel
    TaskID string         `gorm:"type:char(36);not null"`
    Remark sql.NullString `gorm:"type:text"`    // NULL when empty
    AssignedTo sql.NullString `gorm:"type:char(36)"`
}
```

When assigning: `sql.NullString{String: val, Valid: val != ""}`
When reading: `field.Valid` (boolean check), `field.String` (always safe, returns "" if !Valid)

## The Frontend TaskKindDef (frontend/src/composables/useTaskKinds.ts)

```typescript
export interface TaskKindDef {
  card: Component                          // REQUIRED — task card for list view
  todoActions: Component                   // REQUIRED — action buttons on todo card
  labelKey: I18nKey                        // REQUIRED — kind label i18n key (e.g. 'taskKind.simple')
  createLabelKey: I18nKey                  // REQUIRED — create-button label i18n key (e.g. 'taskKind.create')
  inspectComponent?: Component             // optional — modal content for complex OnComplete
  todoInfo?: Component                     // optional — extra info row below todo name
  formComponent?: Component                // optional — kind-specific form fields (v-model bound)
  todoBadgeKey?: I18nKey                   // optional — short badge label on todo card
  defaultCheckItems?: any[]                // optional — default data for new tasks of this kind
  buildDisplaySummary?: (taskWithExtra: { task: any; extra: any }) => string
  serializeExtra?: (formData: any[]) => any   // form array → request payload
  parseExtra?: (extra: any) => any[]          // response payload → form array
  templateWizardStep?: Component              // optional — kind-specific step in TemplateWizard
  wizardStepLabel?: string                    // optional — label for wizard step
}
```

`I18nKey` is a type-safe dot-path key into the i18n message schema (zh-CN.ts is canonical). All label keys must be valid paths.

## Approach

### Phase 1: Gather Requirements

Ask the user:
1. **Kind name** (lowercase identifier, e.g. "checklist", "maintenance")
2. **Display labels** (Chinese + English)
3. **Extra data?** Does this kind have kind-specific data beyond the common task fields?
   - No extra data → follow the `simple` pattern (all handler methods return nil)
   - Has extra data → follow the `inspection` pattern
4. **Complex OnComplete?** Does completing a todo need extra user input (like inspection selections)?
   - Simple → just "done" / "skip" / "remark" buttons
   - Complex → needs a modal workflow with `inspectComponent`
5. **Kind-specific form fields?** Does the create/edit form need extra fields?
6. **Child tasks?** Does this kind manage child tasks (like inspection branches with `create_todo`)?
7. **Ribbon color** for the card badge (Tailwind color, e.g. `blue`, `purple`, `green`, `orange`)

### Phase 2: Backend Handler

Create `backend/pkg/taskkind/<kind>/handler.go`:

**For a simple kind (no extra data):**
Use a struct (either value or pointer receiver):

```go
package <kind>

import (
    "github.com/dezhishen/now-and-again/backend/pkg/model"
    "github.com/dezhishen/now-and-again/backend/pkg/taskkind"
)

type handler struct{}

func init() {
    taskkind.Register(&handler{})
}

func (handler) Kind() string { return "<kind>" }
func (handler) SaveExtra(_ taskkind.TaskStorage, _ *model.TaskModel, _ any) error { return nil }
func (handler) UpdateExtra(_ taskkind.TaskStorage, _ *model.TaskModel, _ any) error { return nil }
func (handler) DeleteExtra(_ taskkind.TaskStorage, _ *model.TaskModel) error { return nil }
func (handler) OnComplete(_ taskkind.TaskStorage, _ *model.TodoModel, _ any) error { return nil }
func (handler) GetExtra(_ taskkind.TaskStorage, _ *model.TaskModel) (any, error) { return nil, nil }
```

### Phase 3: Register Backend Import

In `backend/internal/service/task_service.go`, add blank import:

```go
    _ "github.com/dezhishen/now-and-again/backend/pkg/taskkind/<kind>"
```

### Phase 4: Frontend Components

Create `frontend/src/components/tasks/kinds/<kind>/` with these files.

**CRITICAL patterns (as of v0.0.1-beta-016):**
- All card rows use fixed `h-5` height with `invisible` placeholder when empty
- All text uses `truncate` + `:title` for hover tooltip
- Modals use `v-esc` directive for Escape key closing (no need to add `@keydown.escape`)
- `TaskCard.vue` delegates to kind cards via `getTaskCard(kind)` — no changes needed there
- Schedule labels use `useScheduleLabel` registry — no switch/if needed

#### `<Kind>TaskBody.vue` — Task card for list view

```vue
<script setup lang="ts">
import type { Task } from '@/types'
import { useI18n } from '@/i18n'

const { t } = useI18n()

defineProps<{
  task: Task
  locName: (id: string) => string
  locColor: (id: string) => string
  groupName: (id: string) => string
  summary: (t: Task) => string
}>()

defineEmits<{
  edit: [task: Task]
  logs: [id: string]
  trigger: [id: string]
  toggle: [task: Task]
  delete: [id: string]
}>()
</script>

<template>
  <div class="card hover:shadow-md transition-shadow relative overflow-hidden">
    <!-- Kind ribbon (top-right corner badge) — MUST use bg-<COLOR>-400 pattern -->
    <div class="absolute -top-0.5 -right-0.5 w-14 h-14 overflow-hidden z-10">
      <div class="absolute top-2.5 -right-[18px] w-16 bg-<COLOR>-400 text-white text-[10px] font-medium text-center leading-4 rotate-45 shadow-sm">{{ t('taskCard.<kind>Kind') }}</div>
    </div>
    <div class="flex items-start justify-between mb-2">
      <div class="min-w-0 flex-1">
        <div class="flex items-center gap-2">
          <span class="font-medium dark:text-gray-200 truncate">{{ task.name }}</span>
          <span class="flex-shrink-0 w-1.5 h-1.5 rounded-full" :class="task.enabled ? 'bg-green-500' : 'bg-gray-300'" />
        </div>
        <!-- Summary line: fixed h-5 with truncation + tooltip -->
        <div class="flex items-center justify-between gap-2 mt-1 h-5">
          <span class="text-xs text-gray-400 truncate" :title="summary(task)">{{ summary(task) }}</span>
          <span v-if="task.location_id" class="text-xs px-1.5 py-0.5 rounded flex-shrink-0 truncate max-w-[120px]" :title="locName(task.location_id)" :style="{ background: locColor(task.location_id) + '20', color: locColor(task.location_id) }">📍 {{ locName(task.location_id) }}</span>
          <span v-else class="text-xs px-1.5 py-0.5 rounded flex-shrink-0 invisible">.</span>
        </div>
        <!-- Group line: fixed h-5, placeholder when empty -->
        <div class="flex items-center gap-1 h-5">
          <span v-if="task.group_id" class="text-xs text-gray-400 truncate" :title="groupName(task.group_id)">👥 {{ groupName(task.group_id) }}</span>
          <span v-else class="text-xs invisible">.</span>
        </div>
        <!-- display_summary: fixed h-5, placeholder when empty -->
        <p class="text-xs text-purple-400 h-5 truncate" :title="task.display_summary || ''">
          <template v-if="task.display_summary">🔍 {{ task.display_summary }}</template>
          <span v-else class="invisible">.</span>
        </p>
      </div>
    </div>
    <div class="flex gap-1 border-t dark:border-gray-700 pt-2 mt-2">
      <button class="btn-ghost text-xs flex-1" :disabled="!task.enabled" @click="$emit('edit', task)">{{ t('taskCard.edit') }}</button>
      <button class="btn-ghost text-xs flex-1" @click="$emit('logs', task.id)">{{ t('taskCard.logs') }}</button>
      <button class="btn-ghost text-xs flex-1" :disabled="!task.enabled" @click="$emit('trigger', task.id)">{{ t('taskCard.trigger') }}</button>
      <button class="btn-ghost text-xs flex-1" @click="$emit('toggle', task)">{{ task.enabled ? t('taskCard.disable') : t('taskCard.enable') }}</button>
      <button class="text-xs px-2 py-0.5 rounded text-danger hover:bg-red-50 dark:hover:bg-red-900/30 flex-1" @click="$emit('delete', task.id)">{{ t('taskCard.delete') }}</button>
    </div>
  </div>
</template>
```

**Key rules:**
- Ribbon uses `bg-<COLOR>-400` (pick a distinct color: `blue`, `purple`, `green`, `orange`, `pink`)
- All optional text lines: `h-5 truncate :title="..."`, with `v-else invisible` placeholder
- Group line: always rendered in `h-5` container
- display_summary: always rendered in `h-5` container
- Location badge: `truncate max-w-[120px]` to prevent overflow

#### `<Kind>TodoActions.vue` — Action buttons on todo card

**Simple version** (like `SimpleTodoActions`):

```vue
<script setup lang="ts">
import type { Todo } from '@/types'
import { useI18n } from '@/i18n'

const { t } = useI18n()
defineProps<{ todo: Todo }>()
defineEmits<{ done: [todo: Todo]; skip: [todo: Todo]; remark: [todo: Todo]; completed: [] }>()
</script>

<template>
  <button class="flex-1 text-xs py-1.5 rounded-lg bg-green-50 dark:bg-green-900/30 text-green-700 dark:text-green-300 hover:bg-green-100 dark:hover:bg-green-900/50 transition-colors font-medium" @click="$emit('done', todo)">✅ {{ t('todo.quickDone') }}</button>
  <button class="flex-1 text-xs py-1.5 rounded-lg bg-blue-50 dark:bg-blue-900/30 text-blue-600 dark:text-blue-300 hover:bg-blue-100 dark:hover:bg-blue-900/50 transition-colors font-medium" @click="$emit('remark', todo)">📝 {{ t('todo.remark') }}</button>
  <button v-if="todo.task?.schedule_type !== 'once'" class="flex-1 text-xs py-1.5 rounded-lg bg-gray-50 dark:bg-gray-700/50 text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors font-medium" @click="$emit('skip', todo)">⏭️ {{ t('todo.skip') }}</button>
</template>
```

**Complex version** (with modal, like `InspectionTodoActions`):
- Use `<Teleport to="body">` with `<div class="fixed inset-0 z-50 ..." v-esc="() => showModal = false">`
- Fetch extra data via `GET /todos/:id?with_extra=true`
- Emit `completed` after successful PUT

#### `<Kind>TodoInfo.vue` (optional)

```vue
<script setup lang="ts">
import type { Todo } from '@/types'
defineProps<{ todo: Todo }>()
</script>

<template>
  <p class="text-xs text-purple-400 flex items-center gap-1 h-5">
    <template v-if="todo.task?.display_summary">
      <span class="flex-shrink-0">📋</span>
      <span class="truncate" :title="todo.task.display_summary">{{ todo.task.display_summary }}</span>
    </template>
    <span v-else class="invisible">.</span>
  </p>
</template>
```

### Phase 5: Register Frontend Plugin

In `frontend/src/components/tasks/init.ts`:

```typescript
import <Kind>TaskBody from '@/components/tasks/kinds/<kind>/<Kind>TaskBody.vue'
import <Kind>TodoActions from '@/components/tasks/kinds/<kind>/<Kind>TodoActions.vue'

// Inside initTaskKinds():
  registerTaskKind('<kind>', {
    card: <Kind>TaskBody,
    todoActions: <Kind>TodoActions,
    labelKey: 'taskKind.<kind>',
    createLabelKey: 'taskKind.create<Kind>',
  })
```

For complex kinds with extra data, add:
```typescript
    formComponent: <Kind>Form,           // v-model bound form fields
    inspectComponent: <Kind>Inspect,     // OnComplete modal content
    todoInfo: <Kind>TodoInfo,            // extra info on todo card
    todoBadgeKey: 'taskKind.<kind>',
    buildDisplaySummary({ extra }) { ... },
    serializeExtra(formData) { ... },    // form array → API payload
    parseExtra(extra) { ... },           // API payload → form array (return [] for null)
    defaultCheckItems: [],
```

### Phase 6: i18n

**`zh-CN.ts`** (canonical source):

```typescript
taskKind: {
  <kind>: '<中文标签>',
  create<Kind>: '创建<中文标签>',
},
taskCard: {
  <kind>Kind: '<简短角标>',  // appears in ribbon badge
},
```

**`en.ts`** — mirror all keys with English values.

Run `pnpm check-i18n` to verify all keys exist in both locales.

### Phase 7: Verify

```bash
cd backend && go build ./...
cd frontend && npx vue-tsc --noEmit && npx vite build
```

## File Checklist

| # | File | Purpose |
|---|------|---------|
| 1 | `backend/pkg/taskkind/<kind>/handler.go` | Backend handler (init + 5 methods) |
| 2 | `backend/pkg/taskkind/<kind>/models.go` | (opt) GORM models with model.RegisterModel |
| 3 | `backend/pkg/taskkind/<kind>/repo.go` | (opt) Kind-specific DB queries |
| 4 | `backend/internal/service/task_service.go` | Add blank import `_ "...pkg/taskkind/<kind>"` |
| 5 | `frontend/src/components/tasks/kinds/<kind>/<Kind>TaskBody.vue` | Task card (fixed height, ribbon badge) |
| 6 | `frontend/src/components/tasks/kinds/<kind>/<Kind>TodoActions.vue` | Todo action buttons |
| 7 | `frontend/src/components/tasks/kinds/<kind>/<Kind>TodoInfo.vue` | (opt) Extra info on todo card |
| 8 | `frontend/src/components/tasks/kinds/<kind>/<Kind>Inspect.vue` | (opt) OnComplete modal content |
| 9 | `frontend/src/components/tasks/kinds/<kind>/<Kind>Form.vue` | (opt) Create/edit form fields |
| 10 | `frontend/src/components/tasks/init.ts` | registerTaskKind call |
| 11 | `frontend/src/i18n/locales/zh-CN.ts` | i18n keys |
| 12 | `frontend/src/i18n/locales/en.ts` | i18n keys (mirror zh-CN) |

Fix any type errors before declaring success.

## Reference: Complete File Checklist

For a **simple** kind (no extra data, like `simple`):

| File | Action |
|------|--------|
| `backend/pkg/taskkind/<kind>/handler.go` | CREATE — implement Handler with all-nil methods |
| `backend/internal/service/task_service.go` | EDIT — add blank import |
| `frontend/src/components/tasks/kinds/<kind>/<Kind>TaskBody.vue` | CREATE — task card |
| `frontend/src/components/tasks/kinds/<kind>/<Kind>TodoActions.vue` | CREATE — done/skip/remark buttons |
| `frontend/src/components/tasks/init.ts` | EDIT — import + registerTaskKind() |
| `frontend/src/i18n/locales/zh-CN.ts` | EDIT — add taskKind keys |
| `frontend/src/i18n/locales/en.ts` | EDIT — add taskKind keys |

For a **complex** kind (with extra data, like `inspection`):

All of the above, plus:

| File | Action |
|------|--------|
| `backend/pkg/taskkind/<kind>/models.go` | CREATE (if needed) — GORM models + model.RegisterModel() |
| `backend/pkg/taskkind/<kind>/repo.go` | CREATE (if needed) — kind-specific repositories |
| `frontend/src/components/tasks/kinds/<kind>/<Kind>Inspect.vue` | CREATE (if complex OnComplete) |
| `frontend/src/components/tasks/kinds/<kind>/<Kind>TodoInfo.vue` | CREATE (if todo card extra info) |
| `frontend/src/components/tasks/kinds/<kind>/<Kind>Form.vue` | CREATE (if kind-specific form fields) |

## Key Constraints

- DO NOT modify any file outside the scope of your changes
- DO NOT remove or break existing task kinds (simple, inspection)
- ONLY create NEW files; never overwrite existing handler/component files from other kinds
- Follow the exact directory and naming conventions shown above
- All handler methods must be implemented (even if returning nil)
- Always use `taskStorage` methods (not direct DB calls) for task CRUD to respect the plugin lifecycle
- Kind-specific models MUST be registered via `model.RegisterModel()` in `init()`
- Frontend i18n keys MUST be valid `I18nKey` paths (i.e., exist in zh-CN.ts)
- Use `defineModel` for two-way binding in Vue components, not `defineProps` + `defineEmits` for v-model data
- The `TaskBody` card component must accept exactly these props: `task: Task`, `locName`, `locColor`, `groupName`, `summary`
- The `TaskBody` card component must emit exactly: `edit`, `logs`, `trigger`, `toggle`, `delete`
