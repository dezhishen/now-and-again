---
description: "Use when creating a new task kind (task type) for the Now & Again project. Handles the full plugin lifecycle: backend Handler, frontend components, registration, and i18n."
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
- Frontend i18n: `zh-CN.ts` and `en.ts` need `taskKind.<kind>` and `taskKind.create<Kind>` keys

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
    UpdateNoRootTask(task *model.TaskModel) error
    DeleteNonRootTask(taskID string) error                      // cascading delete + triggers DeleteExtra
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
   - Simple → just a "done" button, like `SimpleTodoActions`
   - Complex → needs a modal workflow, like `InspectionTodoActions` + `InspectionInspect`
5. **Kind-specific form fields?** Does the create/edit form need extra fields beyond name/schedule/location/group?
6. **Child tasks?** Does this kind manage child tasks (like inspection branches with `create_todo`)?

### Phase 2: Backend Handler

Create `backend/pkg/taskkind/<kind>/handler.go`:

**For a simple kind (no extra data):**

```go
package <kind>

import (
    "github.com/dezhishen/now-and-again/backend/pkg/model"
    "github.com/dezhishen/now-and-again/backend/pkg/taskkind"
)

type Handler struct{}

func init() {
    taskkind.Register(Handler{})
}

func (Handler) Kind() string { return "<kind>" }

func (Handler) SaveExtra(taskStorage taskkind.TaskStorage, task *model.TaskModel, extra any) error {
    return nil
}

func (Handler) UpdateExtra(taskStorage taskkind.TaskStorage, task *model.TaskModel, extra any) error {
    return nil
}

func (Handler) DeleteExtra(taskStorage taskkind.TaskStorage, task *model.TaskModel) error {
    return nil
}

func (Handler) OnComplete(taskStorage taskkind.TaskStorage, todo *model.TodoModel, extra any) error {
    return nil
}

func (Handler) GetExtra(taskStorage taskkind.TaskStorage, task *model.TaskModel) (any, error) {
    return nil, nil
}
```

**For a complex kind (with extra data/models):**

Use a struct (pointer or value receiver, both work — `inspection` uses `&handler{}`):

```go
package <kind>

type handler struct{}

func init() {
    taskkind.Register(&handler{})
}

func (handler) Kind() string { return "<kind>" }
```

Implement each method:
- **SaveExtra**: Parse `extra any` (marshal→unmarshal pattern, see inspection), persist kind-specific data via `taskStorage.DB()`. Use the same DB handle as the caller (could be transactional).
- **UpdateExtra**: Load existing data, diff with new data, apply creates/updates/deletes. Handle child task management via `taskStorage.CreateNoRootTask()` / `taskStorage.DeleteNonRootTask()`.
- **DeleteExtra**: Clean up all kind-specific data. If there are child tasks, call `taskStorage.DeleteNonRootTask()` for each (it cascades and triggers child DeleteExtra handlers).
- **OnComplete**: Called when a todo is marked done. `extra` comes from `CompleteTodoRequest.extra`. Use `taskStorage.DB()` for audit/log writes.
- **GetExtra**: Return kind-specific data for task detail/todo views. Called by `GetTaskWithExtra` and `GetTodoWithExtra`.

**If the kind manages child tasks:**
- When creating child tasks, use `taskStorage.CreateNoRootTask(childTask, childExtra)` — it creates the task AND triggers the child's `SaveExtra`.
- When deleting child tasks, use `taskStorage.DeleteNonRootTask(taskID)` — it cascades to grandchildren AND triggers each task's `DeleteExtra`.
- Never call `h.SaveExtra()` / `h.DeleteExtra()` directly on a different kind's handler — use TaskStorage methods.

**If you have kind-specific GORM models**, create `backend/pkg/taskkind/<kind>/models.go`:

```go
package <kind>

import "github.com/dezhishen/now-and-again/backend/pkg/model"

type YourModel struct {
    model.BaseModel
    TaskID   string `gorm:"index;type:char(36);not null"`
    // ... kind-specific fields
}

func (YourModel) TableName() string { return "your_table_name" }

func init() {
    model.RegisterModel(&YourModel{})
}
```

**If you have kind-specific repositories**, create `backend/pkg/taskkind/<kind>/repo.go`:

```go
package <kind>

import "gorm.io/gorm"

type YourRepo struct{ db *gorm.DB }

func NewYourRepo(db *gorm.DB) *YourRepo { return &YourRepo{db} }

func (r *YourRepo) FindByTaskID(taskID string) ([]YourModel, error) { ... }
func (r *YourRepo) Create(m *YourModel) error { ... }
// etc.
```

### Phase 3: Register Backend Import

Add the blank import to `backend/internal/service/task_service.go`.

Find the existing blank imports block:
```go
import (
    // ... other imports ...
    _ "github.com/dezhishen/now-and-again/backend/pkg/taskkind/inspection"
    _ "github.com/dezhishen/now-and-again/backend/pkg/taskkind/simple"
)
```

Add your new import:
```go
    _ "github.com/dezhishen/now-and-again/backend/pkg/taskkind/<kind>"
```

### Phase 4: Frontend Components

Create the directory `frontend/src/components/tasks/kinds/<kind>/` with these files:

#### `<Kind>TaskBody.vue` — Full task card for list view

Follow the exact pattern from `SimpleTaskBody.vue` / `InspectionTaskBody.vue`:

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
    <!-- Kind ribbon — pick a distinctive color -->
    <div class="absolute -top-0.5 -right-0.5 w-14 h-14 overflow-hidden z-10">
      <div class="absolute top-2.5 -right-[18px] w-16 bg-<COLOR>-400 text-white text-[10px] font-medium text-center leading-4 rotate-45 shadow-sm">{{ t('taskCard.<kind>Kind') }}</div>
    </div>
    <!-- task name, enabled dot, summary, location badge, group, display_summary -->
    <div class="flex items-start justify-between mb-2">
      <div class="min-w-0 flex-1">
        <div class="flex items-center gap-2">
          <span class="font-medium dark:text-gray-200 truncate">{{ task.name }}</span>
          <span class="flex-shrink-0 w-1.5 h-1.5 rounded-full" :class="task.enabled ? 'bg-green-500' : 'bg-gray-300'" />
        </div>
        <div class="flex items-center justify-between gap-2 mt-1 h-5">
          <span class="text-xs text-gray-400 truncate">{{ summary(task) }}</span>
          <span v-if="task.location_id" class="text-xs px-1.5 py-0.5 rounded flex-shrink-0" :style="{ background: locColor(task.location_id) + '20', color: locColor(task.location_id) }">
            📍 {{ locName(task.location_id) }}
          </span>
        </div>
        <div v-if="task.group_id" class="flex items-center gap-1 mt-1">
          <span class="text-xs text-gray-400">👥 {{ groupName(task.group_id) }}</span>
        </div>
        <p v-if="task.display_summary" class="text-xs text-purple-400 mt-1">🔍 {{ task.display_summary }}</p>
        <p v-else class="text-xs mt-1 invisible">.</p>
      </div>
    </div>
    <!-- Action buttons row -->
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

The key differences from `SimpleTaskBody`: the ribbon color and the `t('taskCard.<kind>Kind')` key.

#### `<Kind>TodoActions.vue` — Action buttons on todo card

**Simple version** (just mark done/skip, like `SimpleTodoActions`):

```vue
<script setup lang="ts">
import type { Todo } from '@/types'
import { useI18n } from '@/i18n'

const { t } = useI18n()

defineProps<{ todo: Todo }>()
defineEmits<{ done: [todo: Todo]; skip: [todo: Todo]; remark: [todo: Todo] }>()
</script>

<template>
  <button class="flex-1 text-xs py-1.5 rounded-lg bg-green-50 dark:bg-green-900/30 text-green-700 dark:text-green-300 hover:bg-green-100 dark:hover:bg-green-900/50 transition-colors font-medium" @click="$emit('done', todo)">✅ {{ t('todo.quickDone') }}</button>
  <button class="flex-1 text-xs py-1.5 rounded-lg bg-blue-50 dark:bg-blue-900/30 text-blue-600 dark:text-blue-300 hover:bg-blue-100 dark:hover:bg-blue-900/50 transition-colors font-medium" @click="$emit('remark', todo)">📝 {{ t('todo.remark') }}</button>
  <button v-if="todo.task?.schedule_type !== 'once'" class="flex-1 text-xs py-1.5 rounded-lg bg-gray-50 dark:bg-gray-700/50 text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors font-medium" @click="$emit('skip', todo)">⏭️ {{ t('todo.skip') }}</button>
</template>
```

**Complex version** (with modal for inspection, like `InspectionTodoActions`):

The pattern is:
1. A primary button that opens a modal (e.g., "🔍 巡检")
2. On open, fetch `/todos/:id?with_extra=true` to get full task + extra data
3. In the modal, render the `inspectComponent` (or your custom content)
4. Use `defineModel` for two-way binding of selections/data within the modal
5. On submit, call `PUT /todos/:id` with `{ todo: { status: 'done' }, extra: { ... } }`
6. Emit `completed` event on success so the parent refreshes

Key patterns:
- Use `<Teleport to="body">` for the modal
- Use `defineModel<...>('selections', ...)` for complex state shared with the inspect component
- Show loading spinner while fetching extra data

#### `<Kind>TodoInfo.vue` (optional) — Extra info below todo name

```vue
<script setup lang="ts">
import type { Todo } from '@/types'

defineProps<{ todo: Todo }>()
</script>

<template>
  <p v-if="todo.task?.display_summary" class="text-xs text-purple-400 flex items-center gap-1">
    <span>📋</span>
    <span>{{ todo.task.display_summary }}</span>
  </p>
</template>
```

#### `<Kind>Inspect.vue` (optional) — Modal content for complex completion

Use `defineModel` for two-way binding:

```vue
<script setup lang="ts">
const model = defineModel<{ task: any; extra: any } | null>('task', { required: true })
const selections = defineModel<Record<string, ...>>('selections', { default: () => ({}) })
</script>
```

The parent (`TodoActions`) owns `selections` and passes it via `v-model:selections`. The inspect component reads from `model.extra` and populates `selections`.

#### `<Kind>Form.vue` (optional) — Kind-specific form fields in task create/edit

This component receives `v-model` (via `defineModel`) with the kind-specific data array. The main form (`TaskView.vue`) uses:

```vue
<component :is="getFormComponent(kind)" v-model="checkItems" ... />
```

The form component pattern (see `TaskFormCheckItems.vue`):

```vue
<script setup lang="ts">
const checkItems = defineModel<any[]>({ required: true })
// ... manipulate checkItems directly
</script>
```

For child task editing within form components, use `SubTaskEditor`:
```vue
<SubTaskEditor v-model="branch.branch_task" :groups="groups" :locations="locations" />
```

### Phase 5: Register Frontend Plugin

In `frontend/src/components/tasks/init.ts`:

1. Add imports for all your components
2. Add `registerTaskKind('<kind>', { ... })` inside `initTaskKinds()`

Example for a simple kind:

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

Example for a complex kind (with extra data, like inspection):

```typescript
  registerTaskKind('<kind>', {
    card: <Kind>TaskBody,
    inspectComponent: <Kind>Inspect,
    todoActions: <Kind>TodoActions,
    todoInfo: <Kind>TodoInfo,
    formComponent: <Kind>Form,
    todoBadgeKey: 'taskKind.<kind>',
    labelKey: 'taskKind.<kind>',
    createLabelKey: 'taskKind.create<Kind>',
    buildDisplaySummary({ extra }) {
      // Generate display_summary string from extra data
      // Return '' if no summary
    },
    serializeExtra(formData) {
      // Transform form array → API request payload
    },
    parseExtra(extra) {
      // Transform API response payload → form array
      // Return [] if null/undefined
    },
    defaultCheckItems: [],
  })
```

### Phase 6: i18n

Add entries to both locale files:

**`frontend/src/i18n/locales/zh-CN.ts`** — in the `taskKind` section:

```typescript
  taskKind: {
    simple: '任务',
    inspect: '巡检',
    <kind>: '<中文标签>',
    create: '创建任务',
    createInspect: '创建巡检',
    create<Kind>: '创建<中文标签>',
  },
```

If you need a ribbon label, add a `taskCard` key:

```typescript
  taskCard: {
    simpleKind: '任务',
    inspectKind: '巡检',
    <kind>Kind: '<简短标签>',
    // ... existing keys
  },
```

**`frontend/src/i18n/locales/en.ts`** — mirror the same keys with English values.

The i18n keys are type-checked at compile time via `I18nKey`. Chinese locale is the canonical source of truth — English must have all the same keys.

### Phase 7: Verify

Run compilation:

```bash
cd backend && go build ./...
cd frontend && npx vue-tsc --noEmit
```

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
