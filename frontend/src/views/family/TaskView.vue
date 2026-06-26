<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { api } from '@/api/client'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import TaskCard from '@/components/tasks/TaskCard.vue'
import TaskFormCheckItems from '@/components/tasks/TaskFormCheckItems.vue'
import { useToast } from '@/composables/useToast'
import { getTaskCard, getCreateLabel, getDefaultCheckItems, getTaskKinds } from '@/composables/useTaskKinds'
import { initTaskKinds } from '@/components/tasks/init'
import type { TaskTemplate, FamilyGroup, CheckItem } from '@/types'

// Initialize plugin task kinds
initTaskKinds()
const TASK_KINDS = getTaskKinds()

const { t } = useI18n()
const toast = useToast()
const route = useRoute()
const familyId = route.params.familyId as string

const tasks = ref<TaskTemplate[]>([])
const groups = ref<FamilyGroup[]>([])
const locations = ref<{ id: string; name: string; color: string; floor_plan_id: string }[]>([])
const activeTab = ref('all')
const loading = ref(true)

// Log modal
const showLogs = ref(false)
const logTaskId = ref('')
const logTaskName = ref('')
const showSystemLogs = ref(false)
const logsLoading = ref(false)
const logPage = ref(1)
const logTotal = ref(0)
const logSearch = ref('')
const LOG_PAGE_SIZE = 20
const logs = ref<{ id: string; task_id: string; status: string; message?: string; log_type: string; operator_id?: string; created_at: string }[]>([])

const logTotalPages = computed(() => Math.max(1, Math.ceil(logTotal.value / LOG_PAGE_SIZE)))

// Task form
const showTaskForm = ref(false)
const taskName = ref('')
const taskSchedule = ref('daily')
const taskTime = ref('09:00')
const taskDate = ref('')
const taskDays = ref<number[]>([])
const taskGroupID = ref('')
const taskLocationID = ref('')
const taskKind = ref<'simple' | 'inspection'>('simple')
const checkItems = ref<CheckItem[]>([])
const editingTask = ref<TaskTemplate | null>(null)

const SCHEDULE_TYPES = [
  { value: 'once', label: '一次性' },
  { value: 'daily', label: '每天' },
  { value: 'weekly', label: '每周' },
  { value: 'monthly', label: '每月' },
  { value: 'interval', label: '间隔天数' },
]

const WEEKDAYS = ['一', '二', '三', '四', '五', '六', '日']
const MONTH_DAYS = Array.from({ length: 31 }, (_, i) => i + 1)

onMounted(async () => {
  loading.value = true
  await Promise.all([loadTasks(), loadGroups(), loadLocations()])
  loading.value = false
})

// Active tasks: exclude disabled one-shot tasks (already completed).
const activeTasks = computed(() => tasks.value.filter(t => t.enabled || t.schedule_type !== 'once'))
const displayTasks = computed(() => activeTab.value === 'all' ? activeTasks.value : activeTasks.value.filter(t => t.kind === activeTab.value))

function openCreateInspection() {
  openCreate()
  taskKind.value = 'inspection'
  checkItems.value = getDefaultCheckItems('inspection') || []
}

async function loadLocations() {
  try {
    // Get all floor plans, then their locations
    const plans = await api.get<any[]>('/families/' + familyId + '/floor-plans')
    const allLocs: any[] = []
    for (const p of plans) {
      try {
        const locs = await api.get<any[]>('/floor-plans/' + p.id + '/locations')
        allLocs.push(...locs.map((l: any) => ({ ...l, floor_plan_id: p.id })))
      } catch { /* */ }
    }
    locations.value = allLocs
  } catch { locations.value = [] }
}

async function loadGroups() {
  try { groups.value = await api.get<FamilyGroup[]>('/families/' + familyId + '/groups') } catch { groups.value = [] }
}

async function loadTasks() {
  try { tasks.value = await api.get<TaskTemplate[]>('/families/' + familyId + '/tasks') } catch { tasks.value = [] }
}

function buildScheduleData(): any {
  switch (taskSchedule.value) {
    case 'once': return { date: taskDate.value, time: taskTime.value }
    case 'daily': return { time: taskTime.value }
    case 'weekly': return { days: taskDays.value, time: taskTime.value }
    case 'monthly': return { days: taskDays.value, time: taskTime.value }
    case 'interval': return { days: taskDays.value[0] || 1, time: taskTime.value }
  }
}

function openCreate() {
  editingTask.value = null
  taskName.value = ''; taskSchedule.value = 'daily'; taskTime.value = '09:00'
  taskDate.value = ''; taskDays.value = []; taskGroupID.value = ''; taskLocationID.value = ''
  taskKind.value = 'simple'; checkItems.value = []
  showTaskForm.value = true
}

function openEdit(task: TaskTemplate) {
  editingTask.value = task
  taskName.value = task.name
  taskSchedule.value = task.schedule_type
  taskGroupID.value = task.group_id || ''
  taskLocationID.value = task.location_id || ''
  taskKind.value = (task.kind as 'simple' | 'inspection') || 'simple'
  checkItems.value = Array.isArray(task.check_items) ? [...task.check_items] : []
  showTaskForm.value = true
  const data = task.schedule_data || {}
  taskTime.value = data.time || '09:00'
  taskDate.value = data.date || ''
  taskDays.value = data.days || []
}

async function saveTask() {
  const data = buildScheduleData()
  const body: any = {
    name: taskName.value,
    schedule_type: taskSchedule.value,
    schedule_data: data,
  }
  if (taskGroupID.value) body.group_id = taskGroupID.value
  if (taskLocationID.value) body.location_id = taskLocationID.value
  body.kind = taskKind.value
  if (taskKind.value === 'inspection') body.check_items = checkItems.value

  try {
    if (editingTask.value) {
      await api.put('/tasks/' + editingTask.value.id, body)
      toast.success('任务已更新')
    } else {
      await api.post('/families/' + familyId + '/tasks', body)
      toast.success('任务已创建')
    }
    showTaskForm.value = false
    await loadTasks()
  } catch (e: any) { toast.error(e.message) }
}

async function toggleTask(task: TaskTemplate) {
  try {
    await api.put('/tasks/' + task.id, { enabled: !task.enabled })
    task.enabled = !task.enabled
  } catch (e: any) { toast.error(e.message) }
}

async function deleteTask(id: string) {
  if (!confirm('确定删除此任务？')) return
  try { await api.delete('/tasks/' + id); await loadTasks(); toast.success('已删除') } catch (e: any) { toast.error(e.message) }
}

async function triggerTask(id: string) {
  try { await api.post('/tasks/' + id + '/trigger'); await loadTasks(); toast.success('已生成待办') } catch (e: any) { toast.error(e.message) }
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

function scheduleSummary(task: TaskTemplate): string {
  const d = task.schedule_data || {}
  switch (task.schedule_type) {
    case 'once': return `一次性 ${d.date || ''} ${d.time || ''}`
    case 'daily': return `每天 ${d.time || '09:00'}`
    case 'weekly': return `每周 ${(d.days || []).map((n: number) => WEEKDAYS[n-1] || n).join(',')} ${d.time}`
    case 'monthly': return `每月 ${(d.days || []).join(',')}日 ${d.time}`
    case 'interval': return `每 ${d.days || 1} 天 ${d.time}`
    default: return task.schedule_type
  }
}
</script>

<template>
  <div>
    <h2 class="text-xl md:text-2xl font-bold mb-4 dark:text-gray-200">📋 任务管理</h2>

    <LoadingSpinner v-if="loading" />

    <template v-else>

    <!-- Tabs -->
    <div class="flex gap-1 mb-4 border-b dark:border-gray-700">
      <button class="px-4 py-2 text-sm font-medium border-b-2 transition-colors"
        :class="activeTab === 'all' ? 'border-primary text-primary' : 'border-transparent text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'"
        @click="activeTab = 'all'"
      >📋 全部 ({{ activeTasks.length }})</button>
      <button v-for="k in TASK_KINDS" :key="k.kind"
        class="px-4 py-2 text-sm font-medium border-b-2 transition-colors"
        :class="activeTab === k.kind ? 'border-primary text-primary' : 'border-transparent text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'"
        @click="activeTab = k.kind"
      >{{ k.label }} ({{ activeTasks.filter(t => t.kind === k.kind).length }})</button>
    </div>

    <!-- Content -->
    <div>
      <button class="btn-primary text-sm mb-3" @click="activeTab === 'all' ? openCreate() : openCreateInspection()">
        + {{ activeTab === 'all' ? '创建任务' : getCreateLabel(activeTab) }}
      </button>

      <div v-if="displayTasks.length === 0" class="text-center text-gray-400 py-8">暂无任务</div>
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-2 items-start">
        <TaskCard
          v-for="task in displayTasks" :key="task.id"
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
        >
          <template #body>
            <component :is="getTaskCard(task.kind)" :task="task" :loc-name="getLocName" :loc-color="getLocColor" :group-name="getGroupName" :summary="scheduleSummary" />
          </template>
        </TaskCard>
      </div>
    </div>

    <!-- Log Modal -->
    <Teleport to="body">
      <div v-if="showLogs" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60" @click.self="showLogs = false">
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
          <div class="flex-1 overflow-auto p-4" style="min-height: 320px; max-height: 400px">
            <div v-if="logsLoading" class="flex items-center justify-center py-8">
              <span class="animate-spin text-xl">⏳</span>
            </div>
            <div v-else-if="logs.length === 0" class="text-center text-gray-400 py-4 text-sm">暂无操作记录</div>
            <template v-else>
            <div v-for="log in logs" :key="log.id" class="flex items-start gap-2 py-1.5 text-sm border-b dark:border-gray-700 last:border-0">
              <span class="text-xs text-gray-400 w-32 flex-shrink-0">{{ new Date(log.created_at).toLocaleString() }}</span>
              <span class="font-medium w-20 flex-shrink-0" :class="LOG_CLASSES[log.status] || 'text-gray-500'">{{ LOG_LABELS[log.status] || log.status }}</span>
              <span v-if="log.message" class="text-gray-400 truncate flex-1">{{ log.message }}</span>
              <span v-if="log.log_type === 'system'" class="text-xs text-gray-400">系统</span>
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

    <!-- Create/Edit Task Modal -->
    <Teleport to="body">
      <div v-if="showTaskForm" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60" @click.self="showTaskForm = false">
        <div class="bg-white dark:bg-gray-800 rounded-xl shadow-2xl w-[90vw] max-w-xl max-h-[85vh] flex flex-col">
          <div class="flex items-center justify-between px-4 py-3 border-b dark:border-gray-700">
            <h3 class="font-bold dark:text-gray-200">{{ editingTask ? '编辑' : taskKind === 'inspection' ? '创建巡检' : '创建任务' }}</h3>
            <button class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 text-lg" @click="showTaskForm = false">✕</button>
          </div>
          <div class="flex-1 overflow-auto p-4 space-y-3">
            <input v-model="taskName" class="input" :placeholder="taskKind === 'inspection' ? '巡检名称，如：厨房安全检查' : '任务名称，如：每日倒垃圾'" />
            <div>
              <label class="text-xs text-gray-400 block mb-1">调度方式</label>
              <select v-model="taskSchedule" class="input">
                <option v-for="s in SCHEDULE_TYPES" :key="s.value" :value="s.value">{{ s.label }}</option>
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
                {{ taskSchedule === 'weekly' ? '选择星期' : taskSchedule === 'monthly' ? '选择日期' : '间隔天数' }}
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
                <template v-else>
                  <input type="number" v-model.number="taskDays[0]" class="input w-20" placeholder="天数" min="1" />
                </template>
              </div>
            </div>
            <div>
              <label class="text-xs text-gray-400 block mb-1">关联地点（可选）</label>
              <select v-model="taskLocationID" class="input">
                <option value="">不关联</option>
                <option v-for="loc in locations" :key="loc.id" :value="loc.id">{{ loc.name }}</option>
              </select>
              <p class="text-xs text-gray-400 mt-1">可在 <router-link :to="`/family/${familyId}/floor-plan`" class="text-primary underline">户型图</router-link> 中管理地点</p>
            </div>
            <div>
              <label class="text-xs text-gray-400 block mb-1">分配给小组（可选）</label>
              <select v-model="taskGroupID" class="input">
                <option value="">全部成员</option>
                <option v-for="g in groups" :key="g.id" :value="g.id">{{ g.name }}</option>
              </select>
            </div>
            <!-- Branch editor -->
            <TaskFormCheckItems v-if="taskKind === 'inspection'" v-model="checkItems" :groups="groups" />
          </div>
          <div class="flex gap-2 px-4 py-3 border-t dark:border-gray-700">
            <button class="btn-primary text-sm flex-1" @click="saveTask">{{ editingTask ? '保存' : '创建' }}</button>
            <button class="text-sm px-4 py-2 rounded text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-700" @click="showTaskForm = false">取消</button>
          </div>
        </div>
      </div>
    </Teleport>
    </template>
  </div>
</template>
