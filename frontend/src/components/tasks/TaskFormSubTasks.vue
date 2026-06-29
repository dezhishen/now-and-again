<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from '@/i18n'
import type { I18nKey } from '@/i18n'
import type { FamilyGroup } from '@/types'

// ── Subtask definition (inline, not persisted until parent task is saved) ──
export interface SubTaskDef {
  _key: string
  name: string
  schedule_type: string
  schedule_data: {
    time?: string
    date?: string
    days?: number[]
  }
  group_id?: string
  location_id?: string
  enabled: boolean
}

const { t } = useI18n()

const subTasks = defineModel<SubTaskDef[]>({ required: true })

const props = defineProps<{
  groups: FamilyGroup[]
  locations: { id: string; name: string }[]
}>()

// Expose buildDisplaySummary for the parent to use
defineExpose({ buildDisplaySummary })

let _keyCounter = 0
function nextKey(): string {
  return 'st_' + (++_keyCounter) + '_' + Date.now()
}

const SCHEDULE_TYPES: { value: string; labelKey: I18nKey }[] = [
  { value: 'daily', labelKey: 'schedule.daily' },
  { value: 'once', labelKey: 'schedule.once' },
  { value: 'weekly', labelKey: 'schedule.weekly' },
  { value: 'monthly', labelKey: 'schedule.monthly' },
  { value: 'interval', labelKey: 'schedule.interval' },
]

const WEEKDAYS = ['一', '二', '三', '四', '五', '六', '日']
const MONTH_DAYS = Array.from({ length: 31 }, (_, i) => i + 1)

// ── Expanded state: which subtask row is expanded to show full schedule editor ──
const expandedKey = ref<string | null>(null)

function toggleExpand(key: string) {
  expandedKey.value = expandedKey.value === key ? null : key
}

function addSubTask() {
  subTasks.value.push({
    _key: nextKey(),
    name: '',
    schedule_type: 'daily',
    schedule_data: { time: '09:00' },
    enabled: true,
  })
  // Auto-expand the newly added subtask
  const newest = subTasks.value[subTasks.value.length - 1]
  expandedKey.value = newest._key
}

function removeSubTask(index: number) {
  subTasks.value.splice(index, 1)
}

function ensureScheduleData(st: SubTaskDef) {
  if (!st.schedule_data || typeof st.schedule_data !== 'object') {
    st.schedule_data = {}
  }
}

function toggleDay(st: SubTaskDef, d: number) {
  ensureScheduleData(st)
  if (!Array.isArray(st.schedule_data.days)) {
    st.schedule_data.days = []
  }
  const idx = st.schedule_data.days.indexOf(d)
  if (idx >= 0) st.schedule_data.days.splice(idx, 1)
  else st.schedule_data.days.push(d)
}

// ── Summary generation ──────────────────────────────────────────

function scheduleLabel(st: SubTaskDef): string {
  const d = st.schedule_data || {}
  switch (st.schedule_type) {
    case 'once': return `${d.date || '?'} ${d.time || ''}`.trim()
    case 'daily': return `每天${d.time || ''}`
    case 'weekly': {
      const days = (d.days || []).map((n: number) => WEEKDAYS[n - 1] || n).join(',')
      return `每周${days}${d.time || ''}`
    }
    case 'monthly': {
      const days = (d.days || []).join(',')
      return `每月${days}日${d.time || ''}`
    }
    case 'interval': return `每${d.days?.[0] || '?'}天${d.time || ''}`
    default: return st.schedule_type
  }
}

export function buildDisplaySummary(subTasks: SubTaskDef[]): string {
  const names = subTasks
    .filter(st => st.name.trim())
    .map(st => st.name + '(' + scheduleLabel(st) + ')')
  return names.length > 0 ? '子任务: ' + names.join(', ') : ''
}
</script>

<template>
  <div class="space-y-3 border-l-2 border-blue-400 pl-3">
    <div class="flex items-center justify-between">
      <p class="text-xs text-blue-600 dark:text-blue-400 font-medium">📋 {{ t('taskForm.subTasks') }}</p>
      <button class="text-xs text-primary hover:underline" @click="addSubTask">+ {{ t('taskForm.addSubTask') }}</button>
    </div>

    <div v-if="subTasks.length === 0" class="text-xs text-gray-400 py-2">
      {{ t('taskForm.noSubTasks') }}
    </div>

    <div class="max-h-80 overflow-y-auto space-y-2">
      <div
        v-for="(st, i) in subTasks"
        :key="st._key"
        class="border border-gray-200 dark:border-gray-700 rounded-lg p-2 space-y-2"
      >
        <!-- Compact row: name + schedule summary + actions -->
        <div class="flex items-center gap-2">
          <span class="text-xs text-gray-400 flex-shrink-0">#{{ i + 1 }}</span>
          <input
            v-model="st.name"
            class="input text-sm flex-1 min-w-0"
            :placeholder="t('taskForm.subTaskName')"
          />
          <span
            class="text-xs text-gray-400 flex-shrink-0 cursor-pointer hover:text-primary truncate max-w-[120px]"
            @click="toggleExpand(st._key)"
          >{{ scheduleLabel(st) }} ▾</span>
          <button
            class="text-xs text-danger hover:underline flex-shrink-0"
            @click="removeSubTask(i)"
          >{{ t('taskForm.delete') }}</button>
        </div>

        <!-- Expanded schedule editor -->
        <div v-if="expandedKey === st._key" class="space-y-2 pl-4 border-l-2 border-blue-200 dark:border-blue-800">
          <!-- Schedule type -->
          <div class="flex items-center gap-2">
            <label class="text-xs text-gray-400 w-14 flex-shrink-0">{{ t('taskForm.schedule') }}</label>
            <select v-model="st.schedule_type" class="input text-xs flex-1">
              <option v-for="s in SCHEDULE_TYPES" :key="s.value" :value="s.value">{{ t(s.labelKey) }}</option>
            </select>
          </div>

          <!-- Time -->
          <div class="flex items-center gap-2">
            <label class="text-xs text-gray-400 w-14 flex-shrink-0">{{ t('taskForm.time') }}</label>
            <input
              v-model="st.schedule_data.time"
              type="time"
              class="input text-xs flex-1"
            />
          </div>

          <!-- Once: date -->
          <div v-if="st.schedule_type === 'once'" class="flex items-center gap-2">
            <label class="text-xs text-gray-400 w-14 flex-shrink-0">{{ t('taskForm.date') }}</label>
            <input
              v-model="st.schedule_data.date"
              type="date"
              class="input text-xs flex-1"
            />
          </div>

          <!-- Weekly: day picker -->
          <div v-if="st.schedule_type === 'weekly'" class="flex items-start gap-2">
            <label class="text-xs text-gray-400 w-14 flex-shrink-0 pt-0.5">{{ t('taskForm.days') }}</label>
            <div class="flex flex-wrap gap-1">
              <button
                v-for="(name, di) in WEEKDAYS"
                :key="di"
                class="text-xs w-7 h-7 rounded border transition-colors flex items-center justify-center"
                :class="(st.schedule_data.days || []).includes(di + 1)
                  ? 'bg-primary text-white border-primary'
                  : 'border-gray-200 dark:border-gray-600 dark:text-gray-400'"
                @click="toggleDay(st, di + 1)"
              >{{ name }}</button>
            </div>
          </div>

          <!-- Monthly: day picker (1-31) -->
          <div v-if="st.schedule_type === 'monthly'" class="flex items-start gap-2">
            <label class="text-xs text-gray-400 w-14 flex-shrink-0 pt-0.5">{{ t('taskForm.days') }}</label>
            <div class="flex flex-wrap gap-0.5">
              <button
                v-for="d in MONTH_DAYS"
                :key="d"
                class="text-xs w-6 h-6 rounded border transition-colors flex items-center justify-center"
                :class="(st.schedule_data.days || []).includes(d)
                  ? 'bg-primary text-white border-primary'
                  : 'border-gray-200 dark:border-gray-600 dark:text-gray-400'"
                @click="toggleDay(st, d)"
              >{{ d }}</button>
            </div>
          </div>

          <!-- Interval: number input -->
          <div v-if="st.schedule_type === 'interval'" class="flex items-center gap-2">
            <label class="text-xs text-gray-400 w-14 flex-shrink-0">{{ t('taskForm.intervalDays') }}</label>
            <input
              type="number"
              :model-value="st.schedule_data.days?.[0] || 1"
              @input="(e: any) => { ensureScheduleData(st); st.schedule_data.days = [parseInt(e.target.value) || 1] }"
              class="input text-xs w-20"
              min="1"
            />
          </div>

          <!-- Group (optional) -->
          <div class="flex items-center gap-2">
            <label class="text-xs text-gray-400 w-14 flex-shrink-0">{{ t('taskForm.group') }}</label>
            <select v-model="st.group_id" class="input text-xs flex-1">
              <option value="">{{ t('taskForm.allMembers') }}</option>
              <option v-for="g in props.groups" :key="g.id" :value="g.id">{{ g.name }}</option>
            </select>
          </div>

          <!-- Location (optional) -->
          <div class="flex items-center gap-2">
            <label class="text-xs text-gray-400 w-14 flex-shrink-0">{{ t('taskForm.location') }}</label>
            <select v-model="st.location_id" class="input text-xs flex-1">
              <option value="">{{ t('taskForm.noLocation') }}</option>
              <option v-for="loc in props.locations" :key="loc.id" :value="loc.id">{{ loc.name }}</option>
            </select>
          </div>

          <!-- Enabled toggle -->
          <div class="flex items-center gap-2">
            <label class="text-xs text-gray-400 w-14 flex-shrink-0">{{ t('taskForm.enabled') }}</label>
            <label class="flex items-center gap-1 text-xs cursor-pointer">
              <input type="checkbox" v-model="st.enabled" class="accent-primary" />
              <span class="text-gray-400">{{ st.enabled ? t('taskForm.enabledYes') : t('taskForm.enabledNo') }}</span>
            </label>
          </div>
        </div>
      </div>
    </div>

    <!-- Generated summary preview (read-only) -->
    <div v-if="subTasks.length > 0" class="text-xs text-blue-400 bg-blue-50 dark:bg-blue-900/20 rounded px-2 py-1.5">
      📋 {{ buildDisplaySummary(subTasks) }}
    </div>
  </div>
</template>
