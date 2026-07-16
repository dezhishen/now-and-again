<script setup lang="ts">
import {computed, inject, onMounted, ref, type Ref, watch} from 'vue'
import { useI18n } from '@/i18n'
import { api } from '@/api/client'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import TaskCard from '@/components/tasks/TaskCard.vue'
import { useToast } from '@/composables/useToast'
import { useLoading } from '@/composables/useLoading'
import { useErrorHandler } from '@/composables/useErrorHandler'
import ErrorDisplay from '@/components/ErrorDisplay.vue'
import { getDefaultCheckItems, getTaskKinds, getFormComponent, buildDisplaySummary, serializeExtra, parseExtra } from '@/composables/useTaskKinds'
import { scheduleLabel, getScheduleTypes } from '@/components/schedule/registry'
import { useConfirm } from '@/composables/useConfirm'
import { initTaskKinds } from '@/components/tasks/init'
import TemplatePickerDialog from '@/components/taskTemplate/TemplatePickerDialog.vue'
import type { Task, FamilyGroup, TaskTemplate } from '@/types'

// Initialize plugin task kinds
initTaskKinds()

const toast = useToast()
const { error, setError, clearError } = useErrorHandler()

// Reload on tab activation
const refreshKey = inject<Ref<string>>('refreshKey', ref(''))
watch(refreshKey, (newVal) => {
  if (newVal === 'tasks') withLoading(async () => { await loadTasks(); await loadGroups(); await loadLocations() })
})

const tasks = ref<Task[]>([])
const groups = ref<FamilyGroup[]>([])
const locations = ref<{ id: string; name: string; color: string; floor_plan_id?: string }[]>([])
const { loading, withLoading } = useLoading()

// Log modal
const showLogs = ref(false)
const logTaskId = ref('')
const logTaskName = ref('')
const showSystemLogs = ref(false)
const logsLoading = ref(false)
const logPage = ref(1)
const logTotal = ref(0)
const logSearch = ref('')
const LOG_PAGE_SIZE = 10
const logs = ref<{ id: string; task_id: string; task_name?: string; status: string; message?: string; log_type: string; operator_id?: string; created_at: string }[]>([])

const logTotalPages = computed(() => Math.max(1, Math.ceil(logTotal.value / LOG_PAGE_SIZE)))

// Task form — main panel owns common fields + save/cancel; sub-component provides extra fields only
const showTaskForm = ref(false)
const taskName = ref('')
const taskSchedule = ref('daily')
const taskTime = ref('09:00')
const taskDate = ref('')
const taskDays = ref<number[]>([])
const taskYearDay = ref(1)
const taskDuration = ref('')
const taskGroupID = ref('')
const taskLocationID = ref('')
const taskKind = ref('simple')
const checkItems = ref<any[]>([])   // extra data bound to kind-specific formComponent
const editingTask = ref<Task | null>(null)
const saving = ref(false)

// ── Template picker ───────────────────────────────────────────────
const showTemplatePicker = ref(false)

const { t, td } = useI18n()

const SCHEDULE_TYPES = getScheduleTypes()

const WEEKDAYS = ['一', '二', '三', '四', '五', '六', '日']
const MONTH_DAYS = Array.from({ length: 31 }, (_, i) => i + 1)
const DURATION_OPTIONS = [
  { value: '15m', label: '15分钟' },
  { value: '30m', label: '30分钟' },
  { value: '1h', label: '1小时' },
  { value: '2h', label: '2小时' },
  { value: '4h', label: '4小时' },
  { value: '8h', label: '8小时' },
  { value: '1d', label: '1天' },
]

onMounted(() => {
  withLoading(async () => {
    await loadTasks()
    await loadGroups()
    await loadLocations()
  })
})

// Filters — three-state: 'unset' (no filter) / 'true' (only) / 'false' (hide).
// Default unset → no query param → backend returns all (no filter).
const archivedFilter = ref<'unset' | 'true' | 'false'>('unset')
const disabledFilter = ref<'unset' | 'true' | 'false'>('unset')
const filteredTasks = computed(() => tasks.value)

// Cycle: unset → true → false → unset
function cycleFilter(v: 'unset' | 'true' | 'false'): 'unset' | 'true' | 'false' {
  if (v === 'unset') return 'true'
  if (v === 'true') return 'false'
  return 'unset'
}

// Reload when filters change
watch([archivedFilter, disabledFilter], () => {
  withLoading(loadTasks)
})

const allKinds = computed(() => getTaskKinds())

async function loadLocations() {
  try {
    locations.value = await api.get<any[]>('/locations')
  } catch { locations.value = [] }
}

async function loadGroups() {
  try { groups.value = await api.get<FamilyGroup[]>('/groups') } catch { groups.value = [] }
}

async function loadTasks() {
  try {
    const params = new URLSearchParams()
    if (archivedFilter.value !== 'unset') params.set('archived', archivedFilter.value)
    if (disabledFilter.value !== 'unset') params.set('disabled', disabledFilter.value)
    const qs = params.toString()
    tasks.value = await api.get<Task[]>('/tasks' + (qs ? '?' + qs : ''))
  } catch { tasks.value = [] }
}

function buildScheduleData(): any {
  switch (taskSchedule.value) {
    case 'once': return { date: taskDate.value, time: taskTime.value }
    case 'daily': return { time: taskTime.value }
    case 'weekly': return { days: taskDays.value, time: taskTime.value }
    case 'monthly': return { days: taskDays.value, time: taskTime.value }
    case 'yearly': return { day: taskYearDay.value, days: taskDays.value, time: taskTime.value }
    case 'interval': return { days: taskDays.value[0] || 1, time: taskTime.value }
  }
}

function openCreate(kind?: string) {
  kind = kind || 'simple'
  editingTask.value = null
  taskName.value = ''; taskSchedule.value = 'daily'; taskTime.value = '09:00'
  taskDate.value = ''; taskDays.value = []; taskYearDay.value = 1; taskDuration.value = ''; taskGroupID.value = ''; taskLocationID.value = ''
  taskKind.value = kind
  checkItems.value = getDefaultCheckItems(kind) ? [...getDefaultCheckItems(kind)!] : []
  showTaskForm.value = true
}

/** Called when user confirms a template → pre-fill the task form. */
function applyTemplate(tmpl: TaskTemplate, taskDefaults: any, extraSchema: any) {
  showTemplatePicker.value = false

  editingTask.value = null
  taskName.value = taskDefaults?.name || ''
  taskSchedule.value = taskDefaults?.schedule_type || 'daily'
  taskDuration.value = taskDefaults?.duration || ''
  taskGroupID.value = taskDefaults?.group_id || ''
  taskLocationID.value = taskDefaults?.location_id || ''
  taskKind.value = tmpl.kind

  const sd = taskDefaults?.schedule_data || {}
  taskTime.value = sd.time || '09:00'
  taskDate.value = sd.date || ''
  // Normalize: templates store singular "day", direct form uses plural "days" array
  if (sd.days) {
    taskDays.value = sd.days
  } else if (sd.day !== undefined && taskSchedule.value !== 'yearly') {
    taskDays.value = [sd.day]
  } else {
    taskDays.value = []
  }
  if (sd.day !== undefined) taskYearDay.value = sd.day

  // Populate kind-specific extra fields from template extra_schema.
  const parsedExtra = parseExtra(tmpl.kind, extraSchema)
  if (parsedExtra && parsedExtra.length > 0) {
    checkItems.value = JSON.parse(JSON.stringify(parsedExtra))
  } else {
    checkItems.value = getDefaultCheckItems(tmpl.kind) ? [...getDefaultCheckItems(tmpl.kind)!] : []
  }

  showTaskForm.value = true
}

async function openEdit(task: Task) {
  editingTask.value = task
  taskName.value = task.name
  taskSchedule.value = task.schedule_type
  taskGroupID.value = task.group_id || ''
  taskLocationID.value = task.location_id || ''
  taskKind.value = task.kind || 'simple'
  checkItems.value = []

  if (getFormComponent(task.kind || 'simple')) {
    try {
      const res = await api.get<{ extra: any }>('/tasks/' + task.id + '?with_extra=true')
      checkItems.value = parseExtra(task.kind || 'simple', res.extra)
    } catch { /* non-critical */ }
  }

  showTaskForm.value = true
  const data = task.schedule_data || {}
  taskTime.value = data.time || '09:00'
  taskDate.value = data.date || ''
  taskDays.value = data.days || []
  taskYearDay.value = data.day || 1
  taskDuration.value = task.duration || ''
}

async function saveTask() {
  if (saving.value) return
  saving.value = true
  const data = buildScheduleData()
  const taskFields: any = {
    name: taskName.value,
    schedule_type: taskSchedule.value,
    schedule_data: data,
    duration: taskDuration.value || undefined,
    kind: taskKind.value,
    enabled: editingTask.value?.enabled ?? true,
  }
  if (taskGroupID.value) taskFields.group_id = taskGroupID.value
  if (taskLocationID.value) taskFields.location_id = taskLocationID.value

  const body: any = { task: taskFields }
  const extra = serializeExtra(taskKind.value, checkItems.value)
  if (extra !== undefined) {
    body.extra = extra
  }
  const displaySummary = buildDisplaySummary(taskKind.value, { task: taskFields, extra })
  if (displaySummary) {
    taskFields.display_summary = displaySummary
  }

  try {
    if (editingTask.value) {
      await api.put('/tasks/' + editingTask.value.id, body)
      toast.success('任务已更新')
    } else {
      await api.post('/tasks', body)
      toast.success('任务已创建')
    }
    showTaskForm.value = false
    await loadTasks()
  } catch (e: any) { setError(e) }
  finally { saving.value = false }
}

async function toggleTask(task: Task) {
  try {
    const newEnabled = !task.enabled
    await api.put('/tasks/' + task.id + '/enabled', { enabled: newEnabled })
    task.enabled = newEnabled
  } catch (e: any) { setError(e) }
}

async function deleteTask(id: string) {
  if (!await useConfirm(t('taskCard.deleteConfirm'))) return
  try { await api.delete('/tasks/' + id); await loadTasks(); toast.success('已删除') } catch (e: any) { setError(e) }
}

async function triggerTask(id: string) {
  try { await api.post('/tasks/' + id + '/trigger'); await loadTasks(); toast.success('已生成待办') } catch (e: any) { setError(e) }
}

async function viewLogs(taskId: string) {
  logTaskId.value = taskId
  const t = tasks.value.find(t => t.id === taskId)
  logTaskName.value = t?.name || taskId.slice(0, 8)
  showLogs.value = true
  showSystemLogs.value = false
  logPage.value = 1
  logTotal.value = 0
  logSearch.value = ''
  await loadLogs()
}

async function loadLogs() {
  logsLoading.value = true
  try {
    const type = showSystemLogs.value ? '' : 'user'
    const params = new URLSearchParams({ type, limit: String(LOG_PAGE_SIZE), offset: String((logPage.value - 1) * LOG_PAGE_SIZE) })
    const result = await api.get<any[]>('/tasks/' + logTaskId.value + '/logs?' + params)
    if (logSearch.value) {
      const filtered = result.filter((l: any) =>
        (l.message || '').includes(logSearch.value) || (l.status || '').includes(logSearch.value)
      )
      logs.value = filtered
      logTotal.value = filtered.length
    } else {
      logs.value = result
      logTotal.value = result.length < LOG_PAGE_SIZE ? (logPage.value - 1) * LOG_PAGE_SIZE + result.length : (logPage.value + 1) * LOG_PAGE_SIZE
    }
  } catch { logs.value = []; logTotal.value = 0 }
  finally { logsLoading.value = false }
}

async function toggleLogType() {
  showSystemLogs.value = !showSystemLogs.value
  logPage.value = 1
  await loadLogs()
}

function goLogPage(page: number) {
  logPage.value = page
  loadLogs()
}

function onLogSearch() {
  logPage.value = 1
  loadLogs()
}

const LOG_LABELS: Record<string, string> = {
  done: '完成', skipped: '跳过', manual: '手动生成',
  created: '创建', completed: '完成', follow_up: '创建跟进',
}
const LOG_CLASSES: Record<string, string> = {
  done: 'text-green-500', skipped: 'text-gray-400', manual: 'text-blue-500',
  created: 'text-green-500', completed: 'text-green-500', follow_up: 'text-purple-500',
}

function getLocName(id: string) { return locations.value.find(l => l.id === id)?.name || id }
function getLocColor(id: string) { return locations.value.find(l => l.id === id)?.color || '#888' }
function getGroupName(id: string) { return groups.value.find(g => g.id === id)?.name || id }

function toggleDay(d: number) {
  const idx = taskDays.value.indexOf(d)
  if (idx >= 0) taskDays.value.splice(idx, 1)
  else taskDays.value.push(d)
}

function scheduleSummary(task: Task): string {
  return scheduleLabel(task.schedule_type, task.schedule_data || {}, td)
}
</script>

<template>
  <div>

    <LoadingSpinner :text="t('app.loading')" v-if="loading" />

    <template v-else>

    <!-- Content -->
    <div>
      <div class="flex items-center gap-2 mb-3">
        <button class="btn-primary text-sm" data-testid="task-create-btn" @click="openCreate()">+ {{ t('taskKind.create') }}</button>
        <button class="btn-secondary text-sm" data-testid="task-template-btn" @click="showTemplatePicker = true">📋 从模板创建</button>
        <span class="flex-1" />
        <button
          class="text-xs px-2 py-1 rounded border transition-colors"
          :class="{
            'border-gray-200 dark:border-gray-600 text-gray-400': archivedFilter === 'unset',
            'bg-primary/10 border-primary text-primary': archivedFilter === 'true',
            'bg-danger/10 border-danger text-danger': archivedFilter === 'false',
          }"
          @click="archivedFilter = cycleFilter(archivedFilter)"
          :title="archivedFilter === 'unset' ? t('taskCard.filterArchivedUnset') : archivedFilter === 'true' ? t('taskCard.filterArchivedOnly') : t('taskCard.filterArchivedHide')"
        >📦 {{ archivedFilter === 'unset' ? t('taskCard.showArchived') : archivedFilter === 'true' ? (t('taskCard.showArchived') + ' ✓') : (t('taskCard.showArchived') + ' ✗') }}</button>
        <button
          class="text-xs px-2 py-1 rounded border transition-colors"
          :class="{
            'border-gray-200 dark:border-gray-600 text-gray-400': disabledFilter === 'unset',
            'bg-primary/10 border-primary text-primary': disabledFilter === 'true',
            'bg-danger/10 border-danger text-danger': disabledFilter === 'false',
          }"
          @click="disabledFilter = cycleFilter(disabledFilter)"
          :title="disabledFilter === 'unset' ? t('taskCard.filterDisabledUnset') : disabledFilter === 'true' ? t('taskCard.filterDisabledOnly') : t('taskCard.filterDisabledHide')"
        >🚫 {{ disabledFilter === 'unset' ? t('taskCard.showDisabled') : disabledFilter === 'true' ? (t('taskCard.showDisabled') + ' ✓') : (t('taskCard.showDisabled') + ' ✗') }}</button>
      </div>

      <div v-if="filteredTasks.length === 0" class="text-center text-gray-400 py-8">{{ t('taskCard.noTasks') }}</div>
      <div class="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-2 items-start">
        <TaskCard
          v-for="task in filteredTasks" :key="task.id"
          :task="task"
          :loc-name="getLocName"
          :loc-color="getLocColor"
          :group-name="getGroupName"
          :summary="scheduleSummary"
          @edit="openEdit"
          @logs="viewLogs"
          @trigger="triggerTask"
          @toggle="toggleTask"
          @delete="deleteTask"
        />
      </div>
    </div>

    <!-- Log Modal -->
    <Teleport to="body">
      <div v-if="showLogs" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60" v-esc="() => showLogs = false">
        <div class="bg-white dark:bg-gray-800 rounded-xl shadow-2xl w-[90vw] max-w-2xl max-h-[75vh] flex flex-col">
          <div class="flex items-center justify-between px-4 py-3 border-b dark:border-gray-700 flex-shrink-0">
            <h3 class="font-bold dark:text-gray-200 truncate mr-2">📋 {{ logTaskName }}</h3>
            <div class="flex items-center gap-2 flex-shrink-0">
              <input
                v-model="logSearch"
                class="text-xs px-2 py-1 rounded border border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-gray-700 w-28"
                placeholder="搜索..."
                @input="onLogSearch"
              />
              <label class="text-xs text-gray-400 flex items-center gap-1 cursor-pointer whitespace-nowrap">
                <input type="checkbox" :checked="showSystemLogs" @change="toggleLogType" class="accent-primary" />
                系统
              </label>
              <button class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 text-lg" @click="showLogs = false">✕</button>
            </div>
          </div>
          <div class="flex-1 overflow-auto p-4" style="min-height: 300px; max-height: 340px">
            <div v-if="logsLoading" class="flex items-center justify-center py-8">
              <LoadingSpinner :text="t('app.loadingLogs')" />
            </div>
            <div v-else-if="logs.length === 0" class="text-center text-gray-400 py-4 text-sm">暂无操作记录</div>
            <template v-else>
            <!-- Header -->
            <div class="flex items-center gap-2 py-1.5 text-xs text-gray-400 border-b dark:border-gray-700">
              <span class="w-32 flex-shrink-0">时间</span>
              <span class="w-20 flex-shrink-0">任务</span>
              <span class="w-16 flex-shrink-0">状态</span>
              <span class="flex-1 min-w-0">内容</span>
            </div>
            <div v-for="log in logs" :key="log.id" class="flex items-start gap-2 py-1.5 text-sm border-b dark:border-gray-700 last:border-0">
              <span class="text-xs text-gray-400 w-32 flex-shrink-0">{{ new Date(log.created_at).toLocaleString() }}</span>
              <span class="text-xs text-primary bg-primary/10 px-1 rounded flex-shrink-0 w-20 truncate" :title="log.task_name">{{ log.task_name || '' }}</span>
              <span class="font-medium w-16 flex-shrink-0" :class="LOG_CLASSES[log.status] || 'text-gray-500'">{{ LOG_LABELS[log.status] || log.status }}</span>
              <span class="text-gray-400 truncate flex-1 min-w-0">{{ log.message || '' }}</span>
              <span v-if="log.log_type === 'system'" class="text-xs text-gray-400 flex-shrink-0">系统</span>
            </div>
            </template>
          </div>
          <!-- Pagination -->
          <div v-if="logTotalPages > 1" class="flex items-center justify-center gap-1 px-4 py-2 border-t dark:border-gray-700 flex-shrink-0">
            <button class="w-7 h-7 rounded text-xs hover:bg-gray-100 dark:hover:bg-gray-700 disabled:opacity-30" :disabled="logPage <= 1" @click="goLogPage(logPage - 1)">◀</button>
            <template v-for="p in logTotalPages" :key="p">
              <button v-if="p <= 3 || p > logTotalPages - 3 || Math.abs(p - logPage) <= 1"
                class="w-7 h-7 rounded text-xs"
                :class="p === logPage ? 'bg-primary text-white' : 'hover:bg-gray-100 dark:hover:bg-gray-700'"
                @click="goLogPage(p)"
              >{{ p }}</button>
              <span v-else-if="p === 4 || p === logTotalPages - 3" class="text-xs text-gray-400">…</span>
            </template>
            <button class="w-7 h-7 rounded text-xs hover:bg-gray-100 dark:hover:bg-gray-700 disabled:opacity-30" :disabled="logPage >= logTotalPages" @click="goLogPage(logPage + 1)">▶</button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Create/Edit Task Modal — main panel owns common fields + save/cancel -->
    <Teleport to="body">
      <div v-if="showTaskForm" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60" v-esc="() => showTaskForm = false">
        <div class="bg-white dark:bg-gray-800 rounded-xl shadow-2xl w-[90vw] max-w-2xl max-h-[85vh] flex flex-col">
          <div class="flex items-center justify-between px-4 py-3 border-b dark:border-gray-700">
            <h3 class="font-bold dark:text-gray-200">{{ editingTask ? t('taskCard.edit') : t('taskKind.create') }}</h3>
            <button class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 text-lg" @click="showTaskForm = false">✕</button>
          </div>
          <div class="flex-1 overflow-auto p-4 space-y-3">
            <ErrorDisplay :error="error" @close="clearError" />
            <!-- Kind selector (disabled when editing) -->
            <div>
              <label class="text-xs text-gray-400 block mb-1">任务类型</label>
              <select v-model="taskKind" data-testid="task-kind" class="input" :disabled="!!editingTask">
                <option v-for="k in allKinds" :key="k.kind" :value="k.kind">{{ t(k.labelKey) }}</option>
              </select>
            </div>
            <div>
              <label class="text-xs text-gray-400 block mb-1">任务名称</label>
              <input v-model="taskName" data-testid="task-name" class="input" placeholder="输入任务名称" />
            </div>
            <div>
              <label class="text-xs text-gray-400 block mb-1">调度方式</label>
              <select v-model="taskSchedule" class="input">
                <option v-for="s in SCHEDULE_TYPES" :key="s.value" :value="s.value">{{ t(s.labelKey) }}</option>
              </select>
            </div>
            <div>
              <label class="text-xs text-gray-400 block mb-1">触发时间</label>
              <input v-model="taskTime" type="time" class="input" />
            </div>
            <div v-if="taskSchedule === 'once'">
              <label class="text-xs text-gray-400 block mb-1">执行日期</label>
              <input v-model="taskDate" type="date" class="input" />
              <p class="text-xs text-gray-400 mt-1">选择过去的日期会立即生成待办</p>
            </div>
            <div v-if="taskSchedule !== 'daily' && taskSchedule !== 'once'">
              <label class="text-xs text-gray-400 block mb-1">
                {{ taskSchedule === 'weekly' ? '选择星期' : taskSchedule === 'monthly' ? '选择日期' : taskSchedule === 'yearly' ? '选择月份' : '间隔天数' }}
              </label>
              <div class="flex flex-wrap gap-1">
                <template v-if="taskSchedule === 'weekly'">
                  <button v-for="(name, i) in WEEKDAYS" :key="i"
                    class="text-xs px-2 py-1 rounded border transition-colors"
                    :class="taskDays.includes(i+1) ? 'bg-primary text-white border-primary' : 'border-gray-200 dark:border-gray-600 dark:text-gray-400'"
                    @click="toggleDay(i+1)">{{ name }}</button>
                </template>
                <template v-else-if="taskSchedule === 'monthly'">
                  <button v-for="d in MONTH_DAYS" :key="d"
                    class="text-xs w-7 h-7 rounded border transition-colors flex items-center justify-center"
                    :class="taskDays.includes(d) ? 'bg-primary text-white border-primary' : 'border-gray-200 dark:border-gray-600 dark:text-gray-400'"
                    @click="toggleDay(d)">{{ d }}</button>
                </template>
                <template v-else-if="taskSchedule === 'yearly'">
                  <div class="flex flex-wrap gap-1 mb-2">
                    <button v-for="m in 12" :key="m"
                      class="text-xs px-2 py-1 rounded border transition-colors"
                      :class="taskDays.includes(m) ? 'bg-primary text-white border-primary' : 'border-gray-200 dark:border-gray-600 dark:text-gray-400'"
                      @click="toggleDay(m)">{{ m }}月</button>
                  </div>
                  <label class="text-xs text-gray-400 block mb-1">选择日期</label>
                  <div class="flex flex-wrap gap-1">
                    <button v-for="d in 31" :key="d"
                      class="text-xs w-7 h-7 rounded border transition-colors flex items-center justify-center"
                      :class="taskYearDay === d ? 'bg-primary text-white border-primary' : 'border-gray-200 dark:border-gray-600 dark:text-gray-400'"
                      @click="taskYearDay = d">{{ d }}</button>
                  </div>
                </template>
                <template v-else>
                  <input type="number" v-model.number="taskDays[0]" class="input w-20" placeholder="天数" min="1" />
                </template>
              </div>
            </div>
            <!-- Duration -->
            <div>
              <label class="text-xs text-gray-400 block mb-1">{{ t('taskForm.duration') }}（可选）</label>
              <div class="flex flex-wrap gap-1.5 mb-1.5">
                <button v-for="opt in DURATION_OPTIONS" :key="opt.value"
                  type="button"
                  data-testid="task-duration-preset"
                  class="text-xs px-2.5 py-1 rounded border transition-colors"
                  :class="taskDuration === opt.value ? 'bg-primary text-white border-primary' : 'border-gray-200 dark:border-gray-600 dark:text-gray-400 hover:border-primary'"
                  @click="taskDuration = taskDuration === opt.value ? '' : opt.value">
                  {{ opt.label }}
                </button>
              </div>
              <div class="flex gap-1.5">
                <input v-model="taskDuration" data-testid="task-duration" :placeholder="t('taskForm.durationPlaceholder')" class="input flex-1" />
                <button v-if="taskDuration" type="button" class="text-xs text-gray-400 hover:text-danger flex-shrink-0" @click="taskDuration = ''">{{ t('taskForm.durationClear') }}</button>
              </div>
              <p class="text-xs text-gray-400 mt-1">{{ t('taskForm.durationHint') }}</p>
            </div>
            <div>
              <label class="text-xs text-gray-400 block mb-1">关联地点（可选）</label>
              <div class="flex gap-2 items-center">
                <select v-model="taskLocationID" class="input flex-1">
                  <option value="">不关联</option>
                  <option v-for="loc in locations" :key="loc.id" :value="loc.id">{{ loc.name }}</option>
                </select>
                <button v-if="taskLocationID" class="text-xs text-gray-400 hover:text-danger flex-shrink-0" @click="taskLocationID = ''">清除</button>
              </div>
            </div>
            <div>
              <label class="text-xs text-gray-400 block mb-1">分配给小组（可选）</label>
              <div class="flex gap-2 items-center">
                <select v-model="taskGroupID" class="input flex-1">
                  <option value="">全部成员</option>
                  <option v-for="g in groups" :key="g.id" :value="g.id">{{ g.name }}</option>
                </select>
                <button v-if="taskGroupID" class="text-xs text-gray-400 hover:text-danger flex-shrink-0" @click="taskGroupID = ''">清除</button>
              </div>
            </div>
            <!-- Kind-specific extra fields (only) — sub-component provides just the extra UI -->
            <component
              :is="getFormComponent(taskKind)"
              v-if="getFormComponent(taskKind)"
              :key="taskKind"
              v-model="checkItems"
              :groups="groups"
              :locations="locations"
            />
          </div>
          <div class="flex gap-2 px-4 py-3 border-t dark:border-gray-700">
            <button class="btn-primary flex-1" data-testid="task-submit" :disabled="saving" @click="saveTask">{{ saving ? t('app.saving') : editingTask ? '保存' : '创建' }}</button>
            <button class="btn-secondary" data-testid="task-cancel" @click="showTaskForm = false">取消</button>
          </div>
        </div>
      </div>
    </Teleport>
    </template>

    <!-- Template Picker Dialog -->
    <TemplatePickerDialog
      v-if="showTemplatePicker"
      @close="showTemplatePicker = false"
      @apply="applyTemplate"
    />
  </div>
</template>
