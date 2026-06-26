<script setup lang="ts">
import { ref, computed, watch, onMounted, inject, type Ref } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { api } from '@/api/client'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import { useToast } from '@/composables/useToast'
import type { Family, Todo } from '@/types'

const { t } = useI18n()
const toast = useToast()
const route = useRoute()
const familyId = route.params.familyId as string

// Reload data when this tab becomes active (switching tabs)
const refreshKey = inject<Ref<number>>('refreshKey', ref(0))
watch(refreshKey, () => { loadTodos(); loadLocations() })

const activeTab = ref<'todos' | 'overview' | 'stats'>('todos')
const loading = ref(true)
const family = ref<Family | null>(null)
const memberCount = ref(0)
const groupCount = ref(0)
const todos = ref<Todo[]>([])
const locations = ref<{ id: string; name: string }[]>([])
const showAll = ref(false)
const copied = ref(false)
const PAGE_SIZE = 5

const displayTodos = computed(() => showAll.value ? todos.value : todos.value.slice(0, PAGE_SIZE))
const hasMore = computed(() => todos.value.length > PAGE_SIZE)

async function copyInviteCode() {
  if (!family.value?.invite_code) return
  try {
    await navigator.clipboard.writeText(family.value.invite_code)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch { /* */ }
}

async function completeTodo(todo: Todo, status: string) {
  // Show remark prompt
  remarkTodo.value = todo
  remarkAction.value = status
  remarkText.value = ''
  showRemark.value = true
}

// ─── Remark modal ──────────────────────────────────────────────

const showRemark = ref(false)
const remarkTodo = ref<Todo | null>(null)
const remarkAction = ref('')
const remarkText = ref('')

async function submitRemark() {
  const todo = remarkTodo.value
  if (!todo) return
  try {
    await api.put('/todos/' + todo.id, { status: remarkAction.value, remark: remarkText.value })
    await loadTodos()
    toast.success(remarkAction.value === 'done' ? '已完成' : '已跳过')
    showRemark.value = false
  } catch (e: any) { toast.error(e.message) }
}

async function completeBranch(todo: Todo, branchName: string) {
  try {
    await api.put('/todos/' + todo.id, { status: 'done', branch_name: branchName })
    await loadTodos()
    toast.success('已完成: ' + branchName)
  } catch (e: any) { toast.error(e.message) }
}

function skipRemarkAndComplete() {
  showRemark.value = false
  completeTodoDirect(remarkTodo.value!, remarkAction.value)
}

async function completeTodoDirect(todo: Todo, status: string) {
  try {
    await api.put('/todos/' + todo.id, { status })
    await loadTodos()
    toast.success(status === 'done' ? '已完成' : '已跳过')
  } catch (e: any) { toast.error(e.message) }
}

// ─── Inspection Modal ──────────────────────────────────────────

const showInspect = ref(false)
const inspectTodo = ref<Todo | null>(null)
const inspectSelections = ref<Record<string, string>>({})

function openInspect(todo: Todo) {
  inspectTodo.value = todo
  inspectSelections.value = {}
  showInspect.value = true
}

async function submitInspection() {
  const todo = inspectTodo.value
  if (!todo) return
  const selections = Object.entries(inspectSelections.value).map(([item, branch]) => ({
    item, branch
  }))
  if (selections.length === 0) {
    toast.warning('请至少选择一个检查项')
    return
  }
  try {
    await api.post('/tasks/' + todo.task_id + '/inspection', { todo_id: todo.id, selections })
    await loadTodos()
    toast.success('巡检已提交')
    showInspect.value = false
  } catch (e: any) { toast.error(e.message) }
}

async function loadTodos() {
  try {
    todos.value = await api.get<Todo[]>('/families/' + familyId + '/todos?status=pending')
  } catch { todos.value = [] }
}

function fmtRange(start: string, end: string): string {
  const s = new Date(start)
  const e = new Date(end)
  const opts: Intl.DateTimeFormatOptions = { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' }
  return s.toLocaleDateString('zh-CN', opts) + ' → ' + e.toLocaleDateString('zh-CN', opts)
}

function getLocName(id: string) { return locations.value.find(l => l.id === id)?.name || '' }

async function loadLocations() {
  try {
    const plans = await api.get<any[]>('/families/' + familyId + '/floor-plans')
    const all: any[] = []
    for (const p of plans) {
      try { const locs = await api.get<any[]>('/floor-plans/' + p.id + '/locations'); all.push(...locs) } catch { /* */ }
    }
    locations.value = all
  } catch { locations.value = [] }
}

// ─── Statistics ──────────────────────────────────────────────────

interface StatsResponse {
  period: string
  start_date: string
  end_date: string
  summary: { total_completed: number; total_skipped: number; total_manual: number; total_tasks: number; completion_rate: number }
  daily: { date: string; completed: number; skipped: number; manual: number }[]
  by_task: { task_id: string; task_name: string; completed: number; skipped: number; manual: number }[]
}

const statsPeriod = ref<'week' | 'month' | 'year'>('week')
const statsDate = ref('')  // reference date for the period, empty = now
const statsLoading = ref(false)
const stats = ref<StatsResponse | null>(null)

const PERIOD_LABELS: Record<string, string> = { week: '周', month: '月', year: '年' }

function shiftPeriod(delta: number) {
  const d = statsDate.value ? new Date(statsDate.value) : new Date()
  switch (statsPeriod.value) {
    case 'week': d.setDate(d.getDate() + delta * 7); break
    case 'month': d.setMonth(d.getMonth() + delta); break
    case 'year': d.setFullYear(d.getFullYear() + delta); break
  }
  statsDate.value = d.toISOString().slice(0, 10)
}

function resetStatsDate() {
  statsDate.value = ''
}

async function loadStats() {
  statsLoading.value = true
  try {
    const params = new URLSearchParams({ period: statsPeriod.value })
    if (statsDate.value) params.set('date', statsDate.value)
    stats.value = await api.get<StatsResponse>(`/families/${familyId}/statistics?${params}`)
  } catch {
    stats.value = null
  } finally {
    statsLoading.value = false
  }
}

// Compute bar chart max for scaling
const maxDailyCount = computed(() => {
  if (!stats.value) return 1
  let max = 0
  for (const d of stats.value.daily) {
    const sum = d.completed + d.skipped + d.manual
    if (sum > max) max = sum
  }
  return max || 1
})

const statsRangeLabel = computed(() => {
  if (!stats.value) return ''
  const fmt = (d: string) => d.slice(5) // MM-DD
  return `${fmt(stats.value.start_date)} ~ ${fmt(stats.value.end_date)}`
})

watch([statsPeriod, statsDate], () => { if (activeTab.value === 'stats') loadStats() })
watch(activeTab, (tab) => { if (tab === 'stats' && !stats.value) loadStats() })

onMounted(async () => {
  loading.value = true
  await Promise.all([
    loadTodos(),
    loadLocations(),
    (async () => {
      try { family.value = await api.get<Family>('/families/' + familyId) } catch { /* */ }
    })(),
    (async () => {
      try { memberCount.value = (await api.get<any[]>('/families/' + familyId + '/members')).length } catch { /* */ }
    })(),
    (async () => {
      try { groupCount.value = (await api.get<any[]>('/families/' + familyId + '/groups')).length } catch { /* */ }
    })(),
  ])
  loading.value = false
})
</script>

<template>
  <div>
    <h2 class="text-xl md:text-2xl font-bold mb-4 dark:text-gray-200">{{ t('dashboard.heading') }}</h2>

    <LoadingSpinner v-if="loading" />

    <template v-else>

    <!-- Tabs -->
    <div class="flex gap-1 mb-4 border-b dark:border-gray-700">
      <button class="px-4 py-2 text-sm font-medium border-b-2 transition-colors"
        :class="activeTab === 'todos' ? 'border-primary text-primary' : 'border-transparent text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'"
        @click="activeTab = 'todos'"
      >待办 ({{ todos.length }})</button>
      <button class="px-4 py-2 text-sm font-medium border-b-2 transition-colors"
        :class="activeTab === 'overview' ? 'border-primary text-primary' : 'border-transparent text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'"
        @click="activeTab = 'overview'"
      >概览</button>
      <button class="px-4 py-2 text-sm font-medium border-b-2 transition-colors"
        :class="activeTab === 'stats' ? 'border-primary text-primary' : 'border-transparent text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'"
        @click="activeTab = 'stats'"
      >📊 统计</button>
    </div>

    <!-- Todos Tab -->
    <div v-if="activeTab === 'todos'">
      <div v-if="todos.length === 0" class="text-center text-gray-400 py-8">暂无待办事项 🎉</div>
      <div v-else class="grid grid-cols-1 sm:grid-cols-2 gap-3 max-w-2xl items-stretch">
        <div v-for="todo in displayTodos" :key="todo.id"
          class="card flex flex-col gap-1.5 hover:shadow-md transition-shadow h-full"
        >
          <!-- Header: name + kind badge -->
          <div class="flex items-start justify-between gap-2">
            <p class="font-medium dark:text-gray-200 text-sm leading-snug line-clamp-2">{{ todo.task?.name || todo.task_id }}</p>
            <span v-if="todo.task?.kind === 'inspection'"
              class="text-[10px] px-1.5 py-0.5 rounded-full bg-purple-100 dark:bg-purple-900/40 text-purple-600 dark:text-purple-400 flex-shrink-0"
            >巡检</span>
          </div>

          <!-- Meta info -->
          <div class="space-y-0.5">
            <p class="text-xs text-gray-400 flex items-center gap-1">
              <span>🕐</span>
              <span>{{ fmtRange(todo.due_start, todo.due_date) }}</span>
            </p>
            <p v-if="todo.task?.kind === 'inspection' && todo.task?.check_items?.length" class="text-xs text-purple-400 flex items-center gap-1">
              <span>📋</span>
              <span>{{ todo.task!.check_items!.length }} 个检查项</span>
            </p>
            <p v-if="todo.location_id && getLocName(todo.location_id)" class="text-xs text-primary flex items-center gap-1">
              <span>📍</span>
              <span>{{ getLocName(todo.location_id) }}</span>
            </p>
          </div>

          <!-- Actions -->
          <div class="flex gap-1.5 pt-1.5 border-t border-gray-100 dark:border-gray-700 mt-auto">
            <!-- Inspection -->
            <template v-if="todo.task?.kind === 'inspection'">
              <button class="flex-1 text-xs py-1.5 rounded-lg bg-purple-50 dark:bg-purple-900/30 text-purple-700 dark:text-purple-300 hover:bg-purple-100 dark:hover:bg-purple-900/50 transition-colors font-medium" @click="openInspect(todo)">
                🔍 巡检
              </button>
            </template>
            <!-- Regular -->
            <template v-else>
              <button class="flex-1 text-xs py-1.5 rounded-lg bg-green-50 dark:bg-green-900/30 text-green-700 dark:text-green-300 hover:bg-green-100 dark:hover:bg-green-900/50 transition-colors font-medium" @click="completeTodo(todo, 'done')">
                ✅ 完成
              </button>
              <button v-if="todo.task?.schedule_type !== 'once'" class="flex-1 text-xs py-1.5 rounded-lg bg-gray-50 dark:bg-gray-700/50 text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors font-medium" @click="completeTodo(todo, 'skipped')">
                ⏭️ 跳过
              </button>
            </template>
          </div>
        </div>
      </div>
      <div v-if="hasMore" class="text-center mt-4">
        <button class="text-sm text-primary hover:underline font-medium" @click="showAll = !showAll">
          {{ showAll ? '收起' : `显示全部 (${todos.length})` }}
        </button>
      </div>
    </div>

    <!-- Overview Tab -->
    <div v-if="activeTab === 'overview'">
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-6">
        <div class="card"><p class="text-gray-400 text-sm">{{ t('dashboard.members') }}</p><p class="text-2xl font-bold dark:text-gray-200">{{ memberCount }}</p></div>
        <div class="card"><p class="text-gray-400 text-sm">{{ t('dashboard.groups') }}</p><p class="text-2xl font-bold dark:text-gray-200">{{ groupCount }}</p></div>
      </div>

      <div v-if="family?.invite_code" class="card">
        <p class="text-gray-400 text-sm mb-1">{{ t('dashboard.inviteCode') }}</p>
        <p class="text-gray-500 text-xs mb-3">{{ t('dashboard.inviteCodeHint') }}</p>
        <div class="flex items-center gap-3">
          <code class="flex-1 bg-gray-100 dark:bg-gray-700 px-4 py-2 rounded-lg text-lg font-mono tracking-wider text-center select-all dark:text-gray-200">{{ family.invite_code }}</code>
          <button class="px-4 py-2 rounded-lg bg-primary text-white text-sm font-medium hover:opacity-90 transition-opacity whitespace-nowrap" @click="copyInviteCode">{{ copied ? t('dashboard.copied') : t('dashboard.copy') }}</button>
        </div>
      </div>
    </div>

    <!-- Stats Tab -->
    <div v-if="activeTab === 'stats'">
      <!-- Period selector -->
      <div class="flex items-center gap-2 mb-6 flex-wrap">
        <div class="flex gap-0.5 bg-gray-100 dark:bg-gray-800 rounded-lg p-0.5">
          <button v-for="p in (['week','month','year'] as const)" :key="p"
            class="px-4 py-1.5 text-sm rounded-md transition-colors font-medium"
            :class="statsPeriod === p ? 'bg-white dark:bg-gray-700 text-primary shadow-sm' : 'text-gray-500 hover:text-gray-700 dark:hover:text-gray-300'"
            @click="statsPeriod = p; resetStatsDate()"
          >{{ PERIOD_LABELS[p] }}</button>
        </div>
        <button class="w-7 h-7 rounded flex items-center justify-center text-sm hover:bg-gray-100 dark:hover:bg-gray-700" @click="shiftPeriod(-1)">◀</button>
        <span class="text-sm text-gray-600 dark:text-gray-400 min-w-[140px] text-center font-medium">{{ statsRangeLabel }}</span>
        <button class="w-7 h-7 rounded flex items-center justify-center text-sm hover:bg-gray-100 dark:hover:bg-gray-700" @click="shiftPeriod(1)">▶</button>
        <button v-if="statsDate" class="text-xs text-primary hover:underline ml-1" @click="resetStatsDate()">今天</button>
      </div>

      <LoadingSpinner v-if="statsLoading" />

      <template v-else-if="stats">
        <!-- Summary row -->
        <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 mb-6">
          <div class="card text-center py-3">
            <p class="text-3xl font-bold text-green-500">{{ stats.summary.total_completed }}</p>
            <p class="text-xs text-gray-400 mt-1">已完成</p>
          </div>
          <div class="card text-center py-3">
            <p class="text-3xl font-bold text-orange-400">{{ stats.summary.total_skipped }}</p>
            <p class="text-xs text-gray-400 mt-1">已跳过</p>
          </div>
          <div class="card text-center py-3">
            <p class="text-3xl font-bold text-blue-400">{{ stats.summary.total_manual }}</p>
            <p class="text-xs text-gray-400 mt-1">手动触发</p>
          </div>
          <div class="card text-center py-3">
            <p class="text-3xl font-bold" :class="stats.summary.completion_rate >= 0.8 ? 'text-green-500' : stats.summary.completion_rate >= 0.5 ? 'text-amber-500' : 'text-red-400'">{{ (stats.summary.completion_rate * 100).toFixed(0) }}%</p>
            <p class="text-xs text-gray-400 mt-1">完成率</p>
          </div>
        </div>

        <!-- Per-task with progress bars -->
        <div class="card mb-6">
          <h3 class="text-sm font-medium text-gray-500 dark:text-gray-400 mb-4">任务完成情况</h3>
          <div v-if="stats.by_task.length === 0" class="text-center text-gray-400 py-4 text-sm">暂无数据</div>
          <div v-else class="space-y-3">
            <div v-for="bt in stats.by_task" :key="bt.task_id">
              <div class="flex items-center justify-between mb-1">
                <span class="text-sm dark:text-gray-200 truncate max-w-[60%]">{{ bt.task_name || bt.task_id.slice(0, 8) }}</span>
                <span class="text-xs text-gray-400">
                  ✅ {{ bt.completed || 0 }} ⏭️ {{ bt.skipped || 0 }}
                  <span class="ml-2 font-medium" :class="bt.completed + bt.skipped > 0 ? (bt.completed / (bt.completed + bt.skipped) >= 0.8 ? 'text-green-500' : bt.completed / (bt.completed + bt.skipped) >= 0.5 ? 'text-amber-500' : 'text-red-400') : 'text-gray-300'">
                    {{ bt.completed + bt.skipped > 0 ? (bt.completed / (bt.completed + bt.skipped) * 100).toFixed(0) + '%' : '-' }}
                  </span>
                </span>
              </div>
              <div class="w-full bg-gray-100 dark:bg-gray-700 rounded-full h-2 overflow-hidden">
                <div class="h-full rounded-full transition-all"
                  :class="bt.completed + bt.skipped > 0 ? (bt.completed / (bt.completed + bt.skipped) >= 0.8 ? 'bg-green-400' : bt.completed / (bt.completed + bt.skipped) >= 0.5 ? 'bg-amber-400' : 'bg-red-400') : 'bg-gray-300'"
                  :style="{ width: (bt.completed + bt.skipped > 0 ? bt.completed / (bt.completed + bt.skipped) * 100 : 0) + '%' }"
                ></div>
              </div>
            </div>
          </div>
        </div>

        <!-- Daily chart -->
        <div class="card">
          <h3 class="text-sm font-medium text-gray-500 dark:text-gray-400 mb-3">每日趋势</h3>
          <div v-if="stats.daily.length === 0" class="text-center text-gray-400 py-4 text-sm">暂无数据</div>
          <div v-else class="flex items-end gap-0.5 h-24">
            <div v-for="d in stats.daily" :key="d.date" class="flex-1 flex flex-col items-center min-w-0" :title="`${d.date.slice(5)} 完成${d.completed} 跳过${d.skipped}`">
              <div class="w-full flex flex-col justify-end" style="height: 80px">
                <div v-if="d.completed + d.skipped > 0"
                  class="w-full rounded-t-sm transition-all"
                  :class="d.completed / Math.max(d.completed + d.skipped, 1) >= 0.5 ? 'bg-green-400 dark:bg-green-600' : 'bg-orange-300 dark:bg-orange-700'"
                  :style="{ height: ((d.completed + d.skipped) / maxDailyCount * 100) + '%' }"
                ></div>
              </div>
              <span class="text-[9px] text-gray-400 mt-1">{{ d.date.slice(5) }}</span>
            </div>
          </div>
        </div>
      </template>

      <div v-else class="text-center text-gray-400 py-8 text-sm">加载统计失败</div>
    </div>
    </template>

    <!-- Inspection Modal -->
    <Teleport to="body">
      <div v-if="showInspect" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60" @click.self="showInspect = false">
        <div class="bg-white dark:bg-gray-800 rounded-xl shadow-2xl w-[90vw] max-w-lg max-h-[80vh] flex flex-col">
          <div class="flex items-center justify-between px-4 py-3 border-b dark:border-gray-700">
            <h3 class="font-bold dark:text-gray-200">🔍 {{ inspectTodo?.task?.name }}</h3>
            <button class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 text-lg" @click="showInspect = false">✕</button>
          </div>
          <div class="flex-1 overflow-auto p-4 space-y-4">
            <div v-for="item in (inspectTodo?.task?.check_items as any[] || [])" :key="item.name" class="space-y-1">
              <p class="text-sm font-medium text-gray-600 dark:text-gray-300">{{ item.name }}</p>
              <div class="flex flex-wrap gap-1">
                <button v-for="b in item.branches" :key="b.name"
                  class="text-xs px-2 py-1 rounded border transition-colors"
                  :class="inspectSelections[item.name] === b.name
                    ? (b.create_todo ? 'bg-red-500 text-white border-red-500' : 'bg-green-500 text-white border-green-500')
                    : 'border-gray-200 dark:border-gray-600 text-gray-500 dark:text-gray-400 hover:border-primary'"
                  @click="inspectSelections[item.name] = b.name"
                >{{ b.name }}</button>
              </div>
            </div>
          </div>
          <div class="flex gap-2 px-4 py-3 border-t dark:border-gray-700">
            <button class="btn-primary text-sm flex-1" @click="submitInspection">提交巡检</button>
            <button class="text-sm px-4 py-2 rounded text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-700" @click="showInspect = false">取消</button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Remark Modal -->
    <Teleport to="body">
      <div v-if="showRemark" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60" @click.self="showRemark = false">
        <div class="bg-white dark:bg-gray-800 rounded-xl shadow-2xl w-[90vw] max-w-md">
          <div class="flex items-center justify-between px-4 py-3 border-b dark:border-gray-700">
            <h3 class="font-bold dark:text-gray-200">{{ remarkAction === 'done' ? '✅ 完成' : '⏭️ 跳过' }} — 备注</h3>
            <button class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 text-lg" @click="showRemark = false">✕</button>
          </div>
          <div class="p-4 space-y-3">
            <p class="text-sm text-gray-500 dark:text-gray-400">{{ remarkTodo?.task?.name }}</p>
            <textarea
              v-model="remarkText"
              class="input w-full h-20 resize-none text-sm"
              placeholder="添加备注（可选）..."
              @keydown.ctrl.enter="submitRemark"
            ></textarea>
            <div class="flex gap-2">
              <button class="btn-primary text-sm flex-1" @click="submitRemark">确认</button>
              <button class="text-sm px-4 py-2 rounded text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-700" @click="skipRemarkAndComplete">跳过备注</button>
            </div>
            <p class="text-[10px] text-gray-400">Ctrl+Enter 快速提交</p>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
