<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { useRoute } from 'vue-router'
import { api } from '@/api/client'

const route = useRoute()
const familyId = computed(() => route.params.familyId as string)
const isFullscreen = computed(() => route.name === 'calendar-full')
const apiKey = computed(() => (route.query.key as string) || '')
const refreshParam = computed(() => {
  const v = parseInt(route.query.refresh as string)
  return isNaN(v) || v < 0 ? null : v
})

// ── State ────────────────────────────────────────────────────────
interface CalendarEvent {
  task_id: string
  name: string
  kind: string
  time: string
  schedule_type: string
  group_name?: string
}

interface CalendarDay {
  date: string
  weekday: number  // 0=Sun..6=Sat
  isCurrentMonth: boolean
  events: CalendarEvent[]
}

const loading = ref(true)
const days = ref<CalendarDay[]>([])
const currentYear = ref(new Date().getFullYear())
const currentMonth = ref(new Date().getMonth() + 1) // 1-12
const groupID = ref('')
const groups = ref<{ id: string; name: string }[]>([])

// ── Computed ─────────────────────────────────────────────────────
const WEEKDAYS = ['日', '一', '二', '三', '四', '五', '六']
const monthLabel = computed(() => `${currentYear.value}年${currentMonth.value}月`)

const isToday = (dateStr: string) => {
  const today = new Date()
  const y = today.getFullYear()
  const m = String(today.getMonth() + 1).padStart(2, '0')
  const d = String(today.getDate()).padStart(2, '0')
  return dateStr === `${y}-${m}-${d}`
}

// ── Event style helpers ──────────────────────────────────────────
const kindColor = (kind: string) => {
  const map: Record<string, string> = {
    simple: 'bg-blue-500',
    inspection: 'bg-amber-500',
    chain: 'bg-purple-500',
  }
  return map[kind] || 'bg-gray-500'
}

// ── Data loading ─────────────────────────────────────────────────
const BASE_URL = '/api'

async function fetchApi<T>(path: string): Promise<T> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  if (apiKey.value) {
    headers['X-API-Key'] = apiKey.value
  } else {
    const token = api.getAccessToken()
    if (token) headers['Authorization'] = `Bearer ${token}`
  }

  const res = await fetch(BASE_URL + path, { headers, credentials: 'include' })
  const json = await res.json()
  if (!json.success) throw new Error(json.error || 'request failed')
  return json.data
}

async function loadCalendar() {
  loading.value = true
  try {
    const params = new URLSearchParams({
      year: String(currentYear.value),
      month: String(currentMonth.value),
    })
    if (groupID.value) params.set('group_id', groupID.value)

    days.value = await fetchApi<CalendarDay[]>(
      `/families/${familyId.value}/calendar?${params}`
    )
  } catch {
    days.value = []
  } finally {
    loading.value = false
  }
}

async function loadGroups() {
  try {
    groups.value = await fetchApi<{ id: string; name: string }[]>(
      `/families/${familyId.value}/groups`
    )
  } catch { /* ignore */ }
}

// ── Navigation ───────────────────────────────────────────────────
function prevMonth() {
  if (currentMonth.value === 1) {
    currentYear.value--
    currentMonth.value = 12
  } else {
    currentMonth.value--
  }
}

function nextMonth() {
  if (currentMonth.value === 12) {
    currentYear.value++
    currentMonth.value = 1
  } else {
    currentMonth.value++
  }
}

function goToday() {
  const now = new Date()
  currentYear.value = now.getFullYear()
  currentMonth.value = now.getMonth() + 1
}

// ── Fullscreen ───────────────────────────────────────────────────
const isBrowserFullscreen = ref(false)

function toggleFullscreen() {
  if (!document.fullscreenElement) {
    document.documentElement.requestFullscreen()
    isBrowserFullscreen.value = true
  } else {
    document.exitFullscreen()
    isBrowserFullscreen.value = false
  }
}

// Listen for ESC key to exit fullscreen
function onFullscreenChange() {
  isBrowserFullscreen.value = !!document.fullscreenElement
}

// ── Auto-refresh ────────────────────────────────────────────────
const REFRESH_OPTIONS = [
  { label: '30秒', value: 30_000 },
  { label: '1分钟', value: 60_000 },
  { label: '2分钟', value: 120_000 },
  { label: '5分钟', value: 300_000 },
  { label: '关闭', value: 0 },
]
const refreshInterval = ref(refreshParam.value != null ? refreshParam.value * 1000 : 60_000) // default 1 min, or from URL ?refresh=N (seconds)

let pollTimer: ReturnType<typeof setInterval> | null = null

function startPolling() {
  stopPolling()
  if (refreshInterval.value > 0 && isFullscreen.value) {
    pollTimer = setInterval(loadCalendar, refreshInterval.value)
  }
}

function stopPolling() {
  if (pollTimer) { clearInterval(pollTimer); pollTimer = null }
}

// ── Watch & Init ─────────────────────────────────────────────────
watch([currentYear, currentMonth, groupID], loadCalendar)
watch([isFullscreen, refreshInterval], () => startPolling(), { immediate: true })

onMounted(() => {
  loadGroups(); loadCalendar()
  document.addEventListener('fullscreenchange', onFullscreenChange)
})

onBeforeUnmount(() => {
  stopPolling()
  document.removeEventListener('fullscreenchange', onFullscreenChange)
})
</script>

<template>
  <div :class="isFullscreen ? 'h-screen flex flex-col p-4 bg-white dark:bg-gray-900' : 'space-y-4'">
    <!-- Header -->
    <div class="flex items-center justify-between flex-wrap gap-2" :class="isFullscreen ? 'flex-shrink-0' : ''">
      <div class="flex items-center gap-2">
        <button
          class="w-8 h-8 rounded-lg flex items-center justify-center hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
          @click="prevMonth"
        >◀</button>
        <h2 class="text-xl font-bold min-w-[120px] text-center">{{ monthLabel }}</h2>
        <button
          class="w-8 h-8 rounded-lg flex items-center justify-center hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
          @click="nextMonth"
        >▶</button>
        <button
          class="ml-2 px-3 py-1 text-sm rounded-lg border border-gray-300 dark:border-gray-600 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
          @click="goToday"
        >今天</button>
      </div>

      <div class="flex items-center gap-2">
        <select
          v-model="groupID"
          class="px-2 py-1 text-xs rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800"
        >
          <option value="">所有小组</option>
          <option v-for="g in groups" :key="g.id" :value="g.id">{{ g.name }}</option>
        </select>

        <select v-if="isFullscreen" v-model="refreshInterval" class="px-2 py-1 text-xs rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800">
          <option v-for="opt in REFRESH_OPTIONS" :key="opt.value" :value="opt.value">🔄 {{ opt.label }}</option>
        </select>

        <button v-if="isFullscreen"
          class="px-2 py-1 text-xs rounded-lg border border-gray-300 dark:border-gray-600 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
          @click="toggleFullscreen"
          :title="isBrowserFullscreen ? '退出全屏' : '全屏'"
        >{{ isBrowserFullscreen ? '↙️' : '↗️' }}</button>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center py-20">
      <span class="animate-spin text-2xl">⏳</span>
    </div>

    <!-- Calendar Grid -->
    <div v-else class="border border-gray-200 dark:border-gray-700 rounded-xl overflow-hidden" :class="isFullscreen ? 'flex-1 flex flex-col' : ''">
      <!-- Weekday headers -->
      <div class="grid grid-cols-7 bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700" :class="isFullscreen ? 'flex-shrink-0' : ''">
        <div
          v-for="wd in WEEKDAYS" :key="wd"
          class="py-2 text-center text-xs font-semibold text-gray-500 dark:text-gray-400"
        >{{ wd }}</div>
      </div>

      <!-- Day cells -->
      <div class="grid grid-cols-7" :class="isFullscreen ? 'flex-1' : ''">
        <div
          v-for="day in days" :key="day.date"
          class="min-h-[80px] border-b border-r border-gray-100 dark:border-gray-800 p-1.5 transition-colors"
          :class="[
            day.isCurrentMonth ? 'bg-white dark:bg-gray-900' : 'bg-gray-50/50 dark:bg-gray-800/50',
            isToday(day.date) ? 'ring-2 ring-primary ring-inset' : '',
            isFullscreen ? 'min-h-0' : '',
          ]"
        >
          <!-- Day number -->
          <div
            class="text-xs mb-1 w-6 h-6 flex items-center justify-center rounded-full"
            :class="[
              isToday(day.date) ? 'bg-primary text-white font-bold' : 'text-gray-500 dark:text-gray-400',
              !day.isCurrentMonth ? 'opacity-40' : '',
            ]"
          >{{ new Date(day.date).getDate() }}</div>

          <!-- Events -->
          <div class="space-y-0.5">
            <div
              v-for="evt in day.events.slice(0, 3)" :key="evt.task_id"
              class="text-xs px-1 py-0.5 rounded truncate cursor-default"
              :class="kindColor(evt.kind)"
              :title="`${evt.name} (${evt.time})${evt.group_name ? ' - ' + evt.group_name : ''}`"
            >
              <span class="text-white/90">{{ evt.time?.slice(0, 5) }}</span>
              <span class="text-white ml-1">{{ evt.name }}</span>
            </div>
            <div
              v-if="day.events.length > 3"
              class="text-xs text-gray-400 pl-1"
            >+{{ day.events.length - 3 }} 更多</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Legend -->
    <div class="flex items-center gap-4 text-xs text-gray-500 dark:text-gray-400">
      <span class="flex items-center gap-1">
        <span class="w-3 h-3 rounded bg-blue-500"></span> 简单任务
      </span>
      <span class="flex items-center gap-1">
        <span class="w-3 h-3 rounded bg-amber-500"></span> 巡检任务
      </span>
    </div>
  </div>
</template>
