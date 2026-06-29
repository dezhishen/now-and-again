<script lang="ts">
// ── Types & helpers exported for consumers ────────────────────────

export interface SubTaskModel {
  name: string
  kind: string
  schedule_type: string
  schedule_data: {
    time?: string
    date?: string
    days?: number[]
  }
  group_id?: string
  location_id?: string
}

/** The full data shape this component edits via v-model. */
export interface SubTaskData {
  task: SubTaskModel
  extra?: any
}

const WEEKDAYS = ['一', '二', '三', '四', '五', '六', '日']

export function scheduleSummary(task: SubTaskModel): string {
  const d = task.schedule_data || {}
  switch (task.schedule_type) {
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
    default: return task.schedule_type || ''
  }
}

/** One-line description: "任务名 (每天09:00)" */
export function subTaskOneLiner(task: SubTaskModel): string {
  const name = task.name || '未命名'
  return name + ' (' + scheduleSummary(task) + ')'
}
</script>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useI18n } from '@/i18n'
import type { I18nKey } from '@/i18n'
import type { FamilyGroup } from '@/types'
import { getTaskKinds, getFormComponent, getDefaultCheckItems, serializeExtra, parseExtra } from '@/composables/useTaskKinds'

const { t } = useI18n()

const model = defineModel<SubTaskData>({ required: true })

const props = defineProps<{
  groups: FamilyGroup[]
  locations: { id: string; name: string }[]
  startExpanded?: boolean
}>()

const SCHEDULE_TYPES: { value: string; labelKey: I18nKey }[] = [
  { value: 'daily', labelKey: 'schedule.daily' },
  { value: 'once', labelKey: 'schedule.once' },
  { value: 'weekly', labelKey: 'schedule.weekly' },
  { value: 'monthly', labelKey: 'schedule.monthly' },
  { value: 'interval', labelKey: 'schedule.interval' },
]

const _WEEKDAYS = ['一', '二', '三', '四', '五', '六', '日']
const MONTH_DAYS = Array.from({ length: 31 }, (_, i) => i + 1)

const allKinds = computed(() => getTaskKinds())
const showModal = ref(props.startExpanded ?? false)

// ── Working draft (snapshot taken on modal open, written back on confirm) ──

const draft = reactive<SubTaskData>({
  task: { name: '', kind: 'simple', schedule_type: 'daily', schedule_data: { time: '09:00' } },
})
const draftExtra = ref<any[]>([])

// Take snapshot from model into draft
function openModal() {
  // Deep-clone model value into draft
  const snap = cloneDeep(model.value)
  draft.task = snap.task
  // Ensure schedule_data
  if (!draft.task.schedule_data || typeof draft.task.schedule_data !== 'object') {
    draft.task.schedule_data = {}
  }
  // Load draft extra
  const kind = draft.task.kind || 'simple'
  if (snap.extra !== undefined) {
    draftExtra.value = cloneDeep(parseExtra(kind, snap.extra))
  } else {
    draftExtra.value = getDefaultCheckItems(kind) ? cloneDeep(getDefaultCheckItems(kind))! : []
  }
  showModal.value = true
}

// Reset draft extra when kind changes inside modal
watch(() => draft.task.kind, (kind) => {
  draftExtra.value = getDefaultCheckItems(kind) ? cloneDeep(getDefaultCheckItems(kind))! : []
})

// Commit draft back to model
function confirm() {
  model.value.task = cloneDeep(draft.task)
  model.value.extra = serializeExtra(draft.task.kind || 'simple', draftExtra.value)
  showModal.value = false
}

// Close without saving
function cancel() {
  showModal.value = false
}

// ── Helpers ─────────────────────────────────────────────────────

function cloneDeep<T>(obj: T): T {
  if (obj === undefined || obj === null) return obj
  return JSON.parse(JSON.stringify(obj))
}

const kindLabel = computed(() => {
  const k = allKinds.value.find(k => k.kind === model.value.task?.kind)
  return k ? t(k.labelKey) : model.value.task?.kind || ''
})

function ensureScheduleData(task: SubTaskModel) {
  if (!task.schedule_data || typeof task.schedule_data !== 'object') {
    task.schedule_data = {}
  }
}

function toggleDay(task: SubTaskModel, d: number) {
  ensureScheduleData(task)
  if (!Array.isArray(task.schedule_data.days)) {
    task.schedule_data.days = []
  }
  const idx = task.schedule_data.days.indexOf(d)
  if (idx >= 0) task.schedule_data.days.splice(idx, 1)
  else task.schedule_data.days.push(d)
}

const task = computed(() => model.value.task)

defineExpose({ scheduleSummary, subTaskOneLiner })
</script>

<template>
  <div class="space-y-1.5">
    <!-- Compact row: name + kind badge + config button (reads from model) -->
    <div class="flex items-center gap-2">
      <input
        v-model="task.name"
        class="input text-xs flex-1 min-w-[100px]"
        :placeholder="t('taskForm.taskName')"
      />
      <span class="text-[10px] px-1.5 py-0.5 rounded bg-blue-100 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400 flex-shrink-0">{{ kindLabel }}</span>
      <button
        class="text-xs text-blue-500 hover:text-blue-600 flex-shrink-0 whitespace-nowrap"
        @click="openModal"
      >
        配置 ▸
      </button>
    </div>

    <!-- Summary pill (reads from model) -->
    <div v-if="task.name" class="text-xs text-blue-400 bg-blue-50 dark:bg-blue-900/20 rounded px-2 py-0.5">
      📋 {{ subTaskOneLiner(task) }}
    </div>

    <!-- Modal editor (works on draft snapshot) -->
    <Teleport to="body">
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60" @mousedown.self="cancel">
        <div class="bg-white dark:bg-gray-800 rounded-xl shadow-2xl w-[90vw] max-w-lg max-h-[85vh] flex flex-col">
          <div class="flex items-center justify-between px-4 py-3 border-b dark:border-gray-700">
            <h3 class="font-bold dark:text-gray-200">📋 编辑子任务</h3>
            <button class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 text-lg" @click="cancel">✕</button>
          </div>
          <div class="flex-1 overflow-auto p-4 space-y-3">
            <!-- Kind selector -->
            <div class="flex items-center gap-2">
              <label class="text-xs text-gray-400 w-12 flex-shrink-0">类型</label>
              <select v-model="draft.task.kind" class="input text-xs flex-1">
                <option v-for="k in allKinds" :key="k.kind" :value="k.kind">{{ t(k.labelKey) }}</option>
              </select>
            </div>

            <!-- Schedule type -->
            <div class="flex items-center gap-2">
              <label class="text-xs text-gray-400 w-12 flex-shrink-0">{{ t('taskForm.schedule') }}</label>
              <select v-model="draft.task.schedule_type" class="input text-xs flex-1">
                <option v-for="s in SCHEDULE_TYPES" :key="s.value" :value="s.value">{{ t(s.labelKey) }}</option>
              </select>
            </div>

            <!-- Time -->
            <div class="flex items-center gap-2">
              <label class="text-xs text-gray-400 w-12 flex-shrink-0">{{ t('taskForm.time') }}</label>
              <input v-model="draft.task.schedule_data.time" type="time" class="input text-xs flex-1" />
            </div>

            <!-- Once: date -->
            <div v-if="draft.task.schedule_type === 'once'" class="flex items-center gap-2">
              <label class="text-xs text-gray-400 w-12 flex-shrink-0">{{ t('taskForm.date') }}</label>
              <input v-model="draft.task.schedule_data.date" type="date" class="input text-xs flex-1" />
            </div>

            <!-- Weekly: day picker -->
            <div v-if="draft.task.schedule_type === 'weekly'" class="flex items-start gap-2">
              <label class="text-xs text-gray-400 w-12 flex-shrink-0 pt-0.5">{{ t('taskForm.days') }}</label>
              <div class="flex flex-wrap gap-1">
                <button v-for="(name, di) in _WEEKDAYS" :key="di"
                  class="text-xs w-7 h-7 rounded border transition-colors flex items-center justify-center"
                  :class="(draft.task.schedule_data.days || []).includes(di + 1) ? 'bg-primary text-white border-primary' : 'border-gray-200 dark:border-gray-600 dark:text-gray-400'"
                  @click="toggleDay(draft.task, di + 1)">{{ name }}</button>
              </div>
            </div>

            <!-- Monthly: day picker -->
            <div v-if="draft.task.schedule_type === 'monthly'" class="flex items-start gap-2">
              <label class="text-xs text-gray-400 w-12 flex-shrink-0 pt-0.5">{{ t('taskForm.days') }}</label>
              <div class="flex flex-wrap gap-0.5">
                <button v-for="d in MONTH_DAYS" :key="d"
                  class="text-xs w-6 h-6 rounded border transition-colors flex items-center justify-center"
                  :class="(draft.task.schedule_data.days || []).includes(d) ? 'bg-primary text-white border-primary' : 'border-gray-200 dark:border-gray-600 dark:text-gray-400'"
                  @click="toggleDay(draft.task, d)">{{ d }}</button>
              </div>
            </div>

            <!-- Interval -->
            <div v-if="draft.task.schedule_type === 'interval'" class="flex items-center gap-2">
              <label class="text-xs text-gray-400 w-12 flex-shrink-0">{{ t('taskForm.intervalDays') }}</label>
              <input type="number" :model-value="draft.task.schedule_data.days?.[0] || 1"
                @input="ensureScheduleData(draft.task); draft.task.schedule_data.days = [parseInt(($event.target as HTMLInputElement).value) || 1]"
                class="input text-xs w-20" min="1" />
            </div>

            <!-- Group -->
            <div class="flex items-center gap-2">
              <label class="text-xs text-gray-400 w-12 flex-shrink-0">{{ t('taskForm.group') }}</label>
              <select v-model="draft.task.group_id" class="input text-xs flex-1">
                <option value="">{{ t('taskForm.allMembers') }}</option>
                <option v-for="g in props.groups" :key="g.id" :value="g.id">{{ g.name }}</option>
              </select>
            </div>

            <!-- Location -->
            <div class="flex items-center gap-2">
              <label class="text-xs text-gray-400 w-12 flex-shrink-0">{{ t('taskForm.location') }}</label>
              <select v-model="draft.task.location_id" class="input text-xs flex-1">
                <option value="">{{ t('taskForm.noLocation') }}</option>
                <option v-for="loc in props.locations" :key="loc.id" :value="loc.id">{{ loc.name }}</option>
              </select>
            </div>

            <!-- Kind-specific extra fields (works on draftExtra) -->
            <component
              :is="getFormComponent(draft.task.kind)"
              v-if="getFormComponent(draft.task.kind)"
              v-model="draftExtra"
              :groups="props.groups"
              :locations="props.locations"
            />
          </div>
          <div class="flex gap-2 px-4 py-3 border-t dark:border-gray-700">
            <button class="btn-primary flex-1 text-sm" @click="confirm">确定</button>
            <button class="btn-secondary text-sm" @click="cancel">取消</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
