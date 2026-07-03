<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from '@/i18n'
import { useLoading } from '@/composables/useLoading'
import { useErrorHandler } from '@/composables/useErrorHandler'
import ErrorDisplay from '@/components/ErrorDisplay.vue'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import { api } from '@/api/client'
import { listTemplates, renderTemplate } from '@/api/task-templates'
import type { TaskTemplate, TemplateParameter, Location } from '@/types'
import ScheduleInput from './ScheduleInput.vue'

const { td } = useI18n()
const { loading, withLoading } = useLoading()
const { error, setError, clearError } = useErrorHandler()

const emit = defineEmits<{
  close: []
  apply: [template: TaskTemplate, taskDefaults: any, extraSchema: any]
}>()

// ── Step state ────────────────────────────────────────────────────

type Step = 'select' | 'params' | 'preview'

const step = ref<Step>('select')
const templates = ref<TaskTemplate[]>([])
const locations = ref<Location[]>([])
const groups = ref<any[]>([])
const selectedTemplate = ref<TaskTemplate | null>(null)
const params = ref<Record<string, any>>({})
const rendered = ref<any>(null)
const rendering = ref(false)

onMounted(() => {
  withLoading(async () => {
    try { templates.value = await listTemplates() } catch { templates.value = [] }
    try { locations.value = await api.get<Location[]>('/locations') } catch { locations.value = [] }
    try { groups.value = await api.get<any[]>('/groups') } catch { groups.value = [] }
  })
})

// ── Step: select template ─────────────────────────────────────────

// Standard parameters auto-appended to every template
const STANDARD_PARAMS: TemplateParameter[] = [
  { key: '_schedule_type', label: '调度方式', type: 'schedule', required: true },
  { key: '_location', label: '执行地点', type: 'location', required: false },
  { key: '_group',    label: '指派小组', type: 'group',    required: false },
]

// Schedule sub-state (mirrors task form)
const scheduleType = ref('daily')
const scheduleTime = ref('09:00')
const scheduleDate = ref('')
const scheduleDays = ref<number[]>([])
const scheduleYearDay = ref(1)

/** Merge standard params into template params (auto-append if not already defined) */
function mergedParams(tmpl: TaskTemplate | null): TemplateParameter[] {
  if (!tmpl) return STANDARD_PARAMS.filter(p => p.key !== '_schedule_type')
  const existing = new Set((tmpl.parameters || []).map(p => p.key))
  const extras = STANDARD_PARAMS.filter(p => !existing.has(p.key))
  return [...(tmpl.parameters || []), ...extras].filter(p => p.key !== '_schedule_type')
}

function selectTemplate(tmpl: TaskTemplate) {
  selectedTemplate.value = tmpl
  const allParams = mergedParams(tmpl)
  // Initialize params with defaults
  const p: Record<string, any> = {}
  allParams.forEach(param => {
    if (param.default !== undefined) p[param.key] = param.default
    else if (param.type === 'bool') p[param.key] = false
    else if (param.type === 'int' || param.type === 'float') p[param.key] = 0
    else if (param.type === 'location') p[param.key] = param.default || ''
    else if (param.type === 'group') p[param.key] = param.default || ''
    else if (param.type === 'array') p[param.key] = param.default || '[]'
    else p[param.key] = ''
  })
  params.value = p
  rendered.value = null
  step.value = 'params'
}

function backToSelect() {
  step.value = 'select'
  selectedTemplate.value = null
  rendered.value = null
}

// ── Step: fill params & render ────────────────────────────────────

/** Convert JSON array string to newline-separated text */
function arrayToText(val: any): string {
  try {
    const arr = typeof val === 'string' ? JSON.parse(val) : val
    return Array.isArray(arr) ? arr.join('\n') : ''
  } catch {
    return typeof val === 'string' ? val : ''
  }
}

/** Convert newline-separated text to JSON array string */
function textToArray(text: string): string {
  const items = text.split('\n').map(s => s.trim()).filter(Boolean)
  return JSON.stringify(items)
}

const hasParameters = computed(() => mergedParams(selectedTemplate.value).length > 0)

async function handleRender() {
  if (!selectedTemplate.value) return
  rendering.value = true
  try {
    const mergedParams: Record<string, any> = { ...params.value }
    // Merge schedule fields into params for template rendering
    mergedParams['_schedule_type'] = scheduleType.value
    mergedParams['_schedule_time'] = scheduleTime.value
    mergedParams['_schedule_date'] = scheduleDate.value
    mergedParams['_schedule_days'] = scheduleDays.value
    mergedParams['_schedule_year_day'] = scheduleYearDay.value
    // Parse array params from JSON string → actual array for Go template range
    for (const [key, val] of Object.entries(mergedParams)) {
      if (typeof val === 'string' && val.startsWith('[')) {
        try { mergedParams[key] = JSON.parse(val) } catch {}
      }
    }
    const result = await renderTemplate(selectedTemplate.value.template_code, mergedParams)
    rendered.value = result
    step.value = 'preview'
  } catch (e: any) {
    setError(e)
  } finally {
    rendering.value = false
  }
}

// ── Step: apply ───────────────────────────────────────────────────

function handleApply() {
  if (!selectedTemplate.value || !rendered.value) return
  // Merge user-selected schedule into rendered task_defaults
  const td = { ...rendered.value.task_defaults }
  td.schedule_type = scheduleType.value
  const sd: Record<string, any> = { time: scheduleTime.value }
  if (scheduleType.value === 'once' && scheduleDate.value) {
    sd.date = scheduleDate.value
  }
  if (scheduleType.value === 'yearly') {
    sd.day = scheduleYearDay.value
    sd.days = scheduleDays.value
  } else if (scheduleDays.value.length > 0) {
    if (scheduleType.value === 'interval') {
      sd.days = scheduleDays.value[0] || 1
    } else {
      sd.days = scheduleDays.value
    }
  }
  td.schedule_data = sd
  emit('apply', selectedTemplate.value, td, rendered.value.extra_schema)
}

function inputType(p: TemplateParameter): string {
  switch (p.type) {
    case 'int': return 'number'
    case 'float': return 'number'
    case 'bool': return 'checkbox'
    case 'time': return 'time'
    default: return 'text'
  }
}
</script>

<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-4xl mx-4 max-h-[85vh] flex flex-col">
      <!-- Header -->
      <div class="flex items-center justify-between px-4 py-3 border-b dark:border-gray-700 flex-shrink-0">
        <h3 class="font-bold dark:text-gray-200">
          {{ step === 'select' ? '选择模板' : step === 'params' ? '填写参数' : '确认创建' }}
        </h3>
        <button class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 text-lg" @click="emit('close')">✕</button>
      </div>

      <div class="flex-1 overflow-auto p-4">
        <ErrorDisplay :error="error" @close="clearError" />
        <LoadingSpinner :text="td('app.loading')" v-if="loading" />

        <!-- Step 1: Template list -->
        <template v-else-if="step === 'select'">
          <div v-if="templates.length === 0" class="text-center text-gray-400 py-8">
            暂无可用模板
          </div>
          <div v-else class="grid grid-cols-2 gap-2">
            <div
              v-for="tmpl in templates" :key="tmpl.id"
              class="relative flex items-center gap-3 p-3 rounded-lg border border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700 cursor-pointer transition-colors overflow-hidden"
              @click="selectTemplate(tmpl)"
            >
              <!-- Kind badge (top-right corner) -->
              <div class="absolute -top-0.5 -right-0.5 w-14 h-14 overflow-hidden z-10">
                <div class="absolute top-2.5 -right-[18px] w-16 text-white text-[10px] font-medium text-center leading-4 rotate-45 shadow-sm"
                  :class="tmpl.kind === 'inspection' ? 'bg-purple-500' : 'bg-blue-400'"
                >{{ td('taskKind.' + tmpl.kind) || tmpl.kind }}</div>
              </div>
              <span class="text-xl">{{ tmpl.icon || '📋' }}</span>
              <div class="flex-1 min-w-0">
                <div class="font-medium text-sm text-gray-900 dark:text-gray-100">{{ tmpl.name }}</div>
                <div class="text-xs text-gray-400 truncate">{{ tmpl.description || '' }}</div>
              </div>
            </div>
          </div>
        </template>

        <!-- Step 2: Parameters -->
        <template v-else-if="step === 'params' && selectedTemplate">
          <button class="text-xs text-gray-400 hover:text-gray-600 mb-3" @click="backToSelect">← 返回选择</button>

          <div class="mb-3">
            <span class="text-xl mr-2">{{ selectedTemplate.icon || '📋' }}</span>
            <span class="font-medium text-gray-900 dark:text-gray-100">{{ selectedTemplate.name }}</span>
          </div>

          <div v-if="!hasParameters" class="text-sm text-gray-400 mb-4">此模板无需参数</div>
          <div v-else class="space-y-3 mb-4">
            <div v-for="p in mergedParams(selectedTemplate)" :key="p.key" class="flex flex-col gap-1">
              <label class="text-sm font-medium text-gray-700 dark:text-gray-300">
                {{ p.label }}<span v-if="p.required" class="text-red-500">*</span>
              </label>
              <p v-if="p.description" class="text-xs text-gray-400">{{ p.description }}</p>

              <select v-if="p.type === 'select'" v-model="params[p.key]"
                class="w-full rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm">
                <option v-for="opt in p.options" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
              </select>
              <select v-else-if="p.type === 'location'" v-model="params[p.key]"
                class="w-full rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm">
                <option value="">-- 选择地点 --</option>
                <option v-for="loc in locations" :key="loc.id" :value="loc.id">{{ loc.name }}</option>
              </select>
              <select v-else-if="p.type === 'group'" v-model="params[p.key]"
                class="w-full rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm">
                <option value="">-- 选择小组 --</option>
                <option v-for="g in groups" :key="g.id" :value="g.id">{{ g.name }}</option>
              </select>
              <textarea v-else-if="p.type === 'array'" :value="arrayToText(params[p.key])"
                @input="params[p.key] = textToArray(($event.target as HTMLTextAreaElement).value)"
                :placeholder="p.placeholder || '每行一个值'"
                rows="3"
                class="w-full rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm" />
              <label v-else-if="p.type === 'bool'" class="flex items-center gap-2 cursor-pointer">
                <input v-model="params[p.key]" type="checkbox" class="rounded border-gray-300 text-green-500 focus:ring-green-500" />
                <span class="text-sm text-gray-700 dark:text-gray-300">{{ p.label }}</span>
              </label>
              <input v-else-if="p.type === 'time'" v-model="params[p.key]" type="time"
                class="w-full rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm" />
              <input v-else v-model="params[p.key]" :type="inputType(p)" :placeholder="p.placeholder || ''"
                class="w-full rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm" />
            </div>
          </div>

          <!-- Schedule (matches task creation form) -->
          <ScheduleInput
            v-model:schedule-type="scheduleType"
            v-model:schedule-time="scheduleTime"
            v-model:schedule-date="scheduleDate"
            v-model:schedule-days="scheduleDays"
            v-model:schedule-year-day="scheduleYearDay"
            class="mb-4"
          />

          <button
            class="w-full py-2 rounded-md bg-green-500 hover:bg-green-600 text-white text-sm font-medium disabled:opacity-50 transition-colors"
            :disabled="rendering"
            @click="handleRender"
          >{{ rendering ? td('app.rendering') : '预览' }}</button>
        </template>

        <!-- Step 3: Preview & Confirm -->
        <template v-else-if="step === 'preview' && rendered">
          <div class="p-3 bg-gray-50 dark:bg-gray-900 rounded-md border border-gray-200 dark:border-gray-700 mb-4">
            <h4 class="text-sm font-medium text-gray-600 dark:text-gray-400 mb-2">任务预览</h4>
            <div class="space-y-1 text-sm">
              <div v-if="rendered.task_defaults?.name" class="text-gray-900 dark:text-gray-100 font-medium">
                {{ rendered.task_defaults.name }}
              </div>
              <div class="text-gray-500 dark:text-gray-400">
                类型：{{ td('taskKind.' + (selectedTemplate?.kind || '')) || selectedTemplate?.kind }}
              </div>
              <div class="text-gray-500 dark:text-gray-400">
                调度：{{ rendered.task_defaults?.schedule_type || '每天' }}
                <span v-if="rendered.task_defaults?.schedule_data?.time">
                  {{ rendered.task_defaults.schedule_data.time }}
                </span>
              </div>
            </div>
          </div>

          <button
            class="w-full py-2 rounded-md bg-green-500 hover:bg-green-600 text-white text-sm font-medium transition-colors"
            @click="handleApply"
          >填充到任务表单</button>
        </template>
      </div>
    </div>
  </div>
</template>
