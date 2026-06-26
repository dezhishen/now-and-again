<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { api } from '@/api/client'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import type { TaskTemplate, FamilyGroup } from '@/types'

const { t } = useI18n()
const route = useRoute()
const familyId = route.params.familyId as string

const tasks = ref<TaskTemplate[]>([])
const groups = ref<FamilyGroup[]>([])
const locations = ref<{ id: string; name: string; color: string; floor_plan_id: string }[]>([])
const activeTab = ref<'tasks' | 'inspections'>('tasks')
const loading = ref(true)
const error = ref('')

// Log modal
const showLogs = ref(false)
const logTaskId = ref('')
const showSystemLogs = ref(false)
const logs = ref<{ id: string; task_id: string; status: string; message?: string; log_type: string; operator_id?: string; created_at: string }[]>([])

// Task form
const showTaskForm = ref(false)
const taskName = ref('')
const taskSchedule = ref('daily')
const taskTime = ref('09:00')
const taskDate = ref('')
const taskDays = ref<number[]>([])
const taskGroupID = ref('')
const taskLocationID = ref('')
const isInspection = ref(false)
interface InspBranch { name: string; create_todo: boolean; todo_name: string; group_id: string }
const inspectionBranches = ref<InspBranch[]>([])
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

const filteredTasks = computed(() => tasks.value.filter(t => !t.is_inspection))
const inspections = computed(() => tasks.value.filter(t => t.is_inspection))

function openCreateInspection() {
  openCreate()
  isInspection.value = true
  inspectionBranches.value = [
    { name: '正常', create_todo: false, todo_name: '', group_id: '' },
    { name: '异常', create_todo: true, todo_name: '修复{name}', group_id: '' },
  ]
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
  isInspection.value = false; inspectionBranches.value = []
  showTaskForm.value = true
}

function openEdit(task: TaskTemplate) {
  editingTask.value = task
  taskName.value = task.name
  taskSchedule.value = task.schedule_type
  taskGroupID.value = task.group_id || ''
  taskLocationID.value = task.location_id || ''
  isInspection.value = task.is_inspection || false
  inspectionBranches.value = Array.isArray(task.inspection_config) ? [...task.inspection_config] : []
  showTaskForm.value = true
  const data = task.schedule_data || {}
  taskTime.value = data.time || '09:00'
  taskDate.value = data.date || ''
  taskDays.value = data.days || []
}

async function saveTask() {
  error.value = ''
  const data = buildScheduleData()
  const body: any = {
    name: taskName.value,
    schedule_type: taskSchedule.value,
    schedule_data: data,
  }
  if (taskGroupID.value) body.group_id = taskGroupID.value
  if (taskLocationID.value) body.location_id = taskLocationID.value
  body.is_inspection = isInspection.value
  if (isInspection.value) body.inspection_config = inspectionBranches.value

  try {
    if (editingTask.value) {
      await api.put('/tasks/' + editingTask.value.id, body)
    } else {
      await api.post('/families/' + familyId + '/tasks', body)
    }
    showTaskForm.value = false
    await loadTasks()
  } catch (e: any) { error.value = e.message }
}

async function toggleTask(task: TaskTemplate) {
  try {
    await api.put('/tasks/' + task.id, { enabled: !task.enabled })
    task.enabled = !task.enabled
  } catch (e: any) { error.value = e.message }
}

async function deleteTask(id: string) {
  if (!confirm('确定删除此任务？')) return
  try { await api.delete('/tasks/' + id); await loadTasks() } catch (e: any) { error.value = e.message }
}

async function triggerTask(id: string) {
  try { await api.post('/tasks/' + id + '/trigger'); await loadTasks() } catch (e: any) { error.value = e.message }
}

async function viewLogs(taskId: string) {
  logTaskId.value = taskId
  showLogs.value = true
  showSystemLogs.value = false
  try {
    logs.value = await api.get<any[]>('/tasks/' + taskId + '/logs?type=user&limit=50')
  } catch { logs.value = [] }
}

async function toggleLogType() {
  showSystemLogs.value = !showSystemLogs.value
  try {
    const type = showSystemLogs.value ? '' : 'user'
    logs.value = await api.get<any[]>('/tasks/' + logTaskId.value + '/logs?type=' + type + '&limit=50')
  } catch { logs.value = [] }
}

const LOG_LABELS: Record<string, string> = {
  done: '完成', skipped: '跳过', manual: '手动生成',
  created: '创建', completed: '完成',
  'inspection:normal': '巡检-正常', 'inspection:abnormal': '巡检-异常',
}
const LOG_CLASSES: Record<string, string> = {
  done: 'text-green-500', skipped: 'text-gray-400', manual: 'text-blue-500',
  created: 'text-green-500', completed: 'text-green-500',
  'inspection:normal': 'text-green-500', 'inspection:abnormal': 'text-danger',
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
    <h2 class="text-xl md:text-2xl font-bold mb-4 dark:text-gray-200">任务管理</h2>
    <p v-if="error" class="text-danger text-sm mb-3">{{ error }}</p>

    <LoadingSpinner v-if="loading" />

    <template v-else>

    <!-- Tabs -->
    <div class="flex gap-1 mb-4 border-b dark:border-gray-700">
      <button class="px-4 py-2 text-sm font-medium border-b-2 transition-colors"
        :class="activeTab === 'tasks' ? 'border-primary text-primary' : 'border-transparent text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'"
        @click="activeTab = 'tasks'"
      >任务模板</button>
      <button class="px-4 py-2 text-sm font-medium border-b-2 transition-colors"
        :class="activeTab === 'inspections' ? 'border-primary text-primary' : 'border-transparent text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'"
        @click="activeTab = 'inspections'"
      >巡检</button>
    </div>

    <!-- Tasks Tab -->
    <div v-if="activeTab === 'tasks'">
      <button class="btn-primary text-sm mb-3" @click="openCreate">+ 创建任务</button>

      <div v-if="tasks.length === 0" class="text-center text-gray-400 py-8">暂无任务模板</div>
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-2">
        <div v-for="task in filteredTasks" :key="task.id" class="card hover:shadow-md transition-shadow">
          <div class="flex items-start justify-between mb-2">
            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-2">
                <span class="font-medium dark:text-gray-200 truncate">{{ task.name }}</span>
                <span class="flex-shrink-0 w-1.5 h-1.5 rounded-full" :class="task.enabled ? 'bg-green-500' : 'bg-gray-300'"></span>
                <span v-if="task.is_inspection" class="text-xs px-1 rounded bg-warning/20 text-amber-600 dark:text-amber-400">巡检</span>
              </div>
              <p class="text-xs text-gray-400 mt-1">{{ scheduleSummary(task) }}</p>
              <!-- Location -->
              <div v-if="task.location_id" class="flex items-center gap-1 mt-1.5">
                <span class="text-xs px-1.5 py-0.5 rounded" :style="{ background: getLocColor(task.location_id) + '20', color: getLocColor(task.location_id) }">
                  📍 {{ getLocName(task.location_id) }}
                </span>
              </div>
              <!-- Group -->
              <div v-if="task.group_id" class="flex items-center gap-1 mt-1">
                <span class="text-xs text-gray-400">👥 {{ getGroupName(task.group_id) }}</span>
              </div>
            </div>
          </div>
          <div class="flex gap-1 border-t dark:border-gray-700 pt-2 mt-2">
            <button class="text-xs px-2 py-0.5 rounded hover:bg-gray-100 dark:hover:bg-gray-700 dark:text-gray-400 flex-1" @click="openEdit(task)">编辑</button>
            <button class="text-xs px-2 py-0.5 rounded hover:bg-gray-100 dark:hover:bg-gray-700 dark:text-gray-400 flex-1" @click="viewLogs(task.id)">日志</button>
            <button class="text-xs px-2 py-0.5 rounded hover:bg-gray-100 dark:hover:bg-gray-700 dark:text-gray-400 flex-1" @click="triggerTask(task.id)">生成</button>
            <button class="text-xs px-2 py-0.5 rounded hover:bg-gray-100 dark:hover:bg-gray-700 dark:text-gray-400 flex-1" @click="toggleTask(task)">{{ task.enabled ? '禁用' : '启用' }}</button>
            <button class="text-xs px-2 py-0.5 rounded text-danger hover:bg-red-50 dark:hover:bg-red-900/30 flex-1" @click="deleteTask(task.id)">删除</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Inspections Tab -->
    <div v-if="activeTab === 'inspections'">
      <button class="btn-primary text-sm mb-3" @click="openCreateInspection">+ 创建巡检</button>

      <div v-if="inspections.length === 0" class="text-center text-gray-400 py-8">暂无巡检模板</div>
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-2">
        <div v-for="task in inspections" :key="task.id" class="card hover:shadow-md transition-shadow">
          <div class="flex items-start justify-between mb-2">
            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-2">
                <span class="font-medium dark:text-gray-200 truncate">{{ task.name }}</span>
                <span class="flex-shrink-0 w-1.5 h-1.5 rounded-full" :class="task.enabled ? 'bg-green-500' : 'bg-gray-300'"></span>
              </div>
              <p class="text-xs text-gray-400 mt-1">{{ scheduleSummary(task) }}</p>
              <div v-if="task.location_id" class="flex items-center gap-1 mt-1.5">
                <span class="text-xs px-1.5 py-0.5 rounded" :style="{ background: getLocColor(task.location_id) + '20', color: getLocColor(task.location_id) }">
                  📍 {{ getLocName(task.location_id) }}
                </span>
              </div>
              <div v-if="task.group_id" class="flex items-center gap-1 mt-1">
                <span class="text-xs text-gray-400">👥 {{ getGroupName(task.group_id) }}</span>
              </div>
              <!-- Branch hints -->
              <div v-if="task.inspection_config?.length" class="text-xs text-gray-400 mt-1 flex flex-wrap gap-1">
                <span v-for="b in task.inspection_config" :key="b.name" class="px-1 rounded" :class="b.create_todo ? 'text-warning' : 'text-green-600'">
                  {{ b.create_todo ? '⚠' : '✓' }}{{ b.name }}
                </span>
              </div>
            </div>
          </div>
          <div class="flex gap-1 border-t dark:border-gray-700 pt-2 mt-2">
            <button class="text-xs px-2 py-0.5 rounded hover:bg-gray-100 dark:hover:bg-gray-700 dark:text-gray-400 flex-1" @click="openEdit(task)">编辑</button>
            <button class="text-xs px-2 py-0.5 rounded hover:bg-gray-100 dark:hover:bg-gray-700 dark:text-gray-400 flex-1" @click="viewLogs(task.id)">日志</button>
            <button class="text-xs px-2 py-0.5 rounded hover:bg-gray-100 dark:hover:bg-gray-700 dark:text-gray-400 flex-1" @click="triggerTask(task.id)">生成</button>
            <button class="text-xs px-2 py-0.5 rounded hover:bg-gray-100 dark:hover:bg-gray-700 dark:text-gray-400 flex-1" @click="toggleTask(task)">{{ task.enabled ? '禁用' : '启用' }}</button>
            <button class="text-xs px-2 py-0.5 rounded text-danger hover:bg-red-50 dark:hover:bg-red-900/30 flex-1" @click="deleteTask(task.id)">删除</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Log Modal -->
    <Teleport to="body">
      <div v-if="showLogs" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60" @click.self="showLogs = false">
        <div class="bg-white dark:bg-gray-800 rounded-xl shadow-2xl w-[90vw] max-w-md max-h-[70vh] flex flex-col">
          <div class="flex items-center justify-between px-4 py-3 border-b dark:border-gray-700">
            <h3 class="font-bold dark:text-gray-200">操作记录</h3>
            <div class="flex items-center gap-2">
              <label class="text-xs text-gray-400 flex items-center gap-1 cursor-pointer">
                <input type="checkbox" :checked="showSystemLogs" @change="toggleLogType" class="accent-primary" />
                系统日志
              </label>
              <button class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 text-lg" @click="showLogs = false">✕</button>
            </div>
          </div>
          <div class="flex-1 overflow-auto p-4">
            <div v-if="logs.length === 0" class="text-center text-gray-400 py-4 text-sm">暂无操作记录</div>
            <div v-for="log in logs" :key="log.id" class="flex items-start gap-2 py-1.5 text-sm border-b dark:border-gray-700 last:border-0">
              <span class="text-xs text-gray-400 w-32 flex-shrink-0">{{ new Date(log.created_at).toLocaleString() }}</span>
              <span class="font-medium w-20 flex-shrink-0" :class="LOG_CLASSES[log.status] || 'text-gray-500'">{{ LOG_LABELS[log.status] || log.status }}</span>
              <span v-if="log.message" class="text-gray-400 truncate flex-1">{{ log.message }}</span>
              <span v-if="log.log_type === 'system'" class="text-xs text-gray-400">系统</span>
            </div>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Create/Edit Task Modal -->
    <Teleport to="body">
      <div v-if="showTaskForm" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60" @click.self="showTaskForm = false">
        <div class="bg-white dark:bg-gray-800 rounded-xl shadow-2xl w-[90vw] max-w-lg max-h-[85vh] flex flex-col">
          <div class="flex items-center justify-between px-4 py-3 border-b dark:border-gray-700">
            <h3 class="font-bold dark:text-gray-200">{{ isInspection ? '创建巡检' : editingTask ? '编辑任务' : '创建任务' }}</h3>
            <button class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 text-lg" @click="showTaskForm = false">✕</button>
          </div>
          <div class="flex-1 overflow-auto p-4 space-y-3">
            <input v-model="taskName" class="input" :placeholder="isInspection ? '巡检名称，如：厨房安全检查' : '任务名称，如：每日倒垃圾'" />
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
            <!-- Inspection branches -->
            <div v-if="isInspection" class="space-y-2 border-l-2 border-warning pl-3">
              <div class="flex items-center justify-between">
                <p class="text-xs text-warning font-medium">📋 巡检分支</p>
                <button class="text-xs text-primary hover:underline" @click="inspectionBranches.push({ name: '', create_todo: false, todo_name: '', group_id: '' })">+ 添加</button>
              </div>
              <div v-for="(b, i) in inspectionBranches" :key="i" class="space-y-1 pb-2 border-b border-gray-100 dark:border-gray-700 last:border-0 last:pb-0">
                <div class="flex gap-2 items-center">
                  <input v-model="b.name" class="input flex-1 text-sm" placeholder="分支名称" />
                  <button class="text-xs text-danger hover:underline flex-shrink-0" @click="inspectionBranches.splice(i, 1)">删除</button>
                </div>
                <label class="flex items-center gap-1 text-xs cursor-pointer">
                  <input type="checkbox" v-model="b.create_todo" class="accent-warning" />
                  <span class="text-gray-500">选择此项时创建待办</span>
                </label>
                <template v-if="b.create_todo">
                  <input v-model="b.todo_name" class="input text-sm" placeholder="待办名称，如 修复{name}" />
                  <select v-model="b.group_id" class="input text-sm">
                    <option value="">分配给小组（可选）</option>
                    <option v-for="g in groups" :key="g.id" :value="g.id">{{ g.name }}</option>
                  </select>
                </template>
              </div>
            </div>
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
