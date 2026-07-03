<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from '@/i18n'
import { useLoading } from '@/composables/useLoading'
import { useErrorHandler } from '@/composables/useErrorHandler'
import ErrorDisplay from '@/components/ErrorDisplay.vue'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import { api } from '@/api/client'
import { listTemplates, renderTemplate } from '@/api/task-templates'
import type { TaskTemplate, TemplateParameter, Location } from '@/types'

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
const scheduleSub = ref<Record<string, any>>({})  // sub-fields for schedule params
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
  { key: '_schedule', label: '调度策略', type: 'schedule', required: true, default: 'daily' },
  { key: '_location', label: '执行地点', type: 'location', required: false },
  { key: '_group',    label: '指派小组', type: 'group',    required: false },
]

/** Merge standard params into template params (auto-append if not already defined) */
function mergedParams(tmpl: TaskTemplate | null): TemplateParameter[] {
  if (!tmpl) return STANDARD_PARAMS
  const existing = new Set((tmpl.parameters || []).map(p => p.key))
  const extras = STANDARD_PARAMS.filter(p => !existing.has(p.key))
  return [...(tmpl.parameters || []), ...extras]
}

function selectTemplate(tmpl: TaskTemplate) {
  selectedTemplate.value = tmpl
  const allParams = mergedParams(tmpl)
  // Initialize params with defaults
  const p: Record<string, any> = {}
  const sub: Record<string, any> = {}
  allParams.forEach(param => {
    if (param.default !== undefined) p[param.key] = param.default
    else if (param.type === 'bool') p[param.key] = false
    else if (param.type === 'int' || param.type === 'float') p[param.key] = 0
    else if (param.type === 'location') p[param.key] = param.default || ''
    else if (param.type === 'group') p[param.key] = param.default || ''
    else if (param.type === 'schedule') {
      p[param.key] = param.default || 'daily'
      sub[param.key] = { time: '09:00', day: 1, hours: 24 }
    }
    else p[param.key] = ''
  })
  params.value = p
  scheduleSub.value = sub
  rendered.value = null
  step.value = 'params'
}

function backToSelect() {
  step.value = 'select'
  selectedTemplate.value = null
  rendered.value = null
}

// Lazy-init schedule sub-fields
const scheduleKeys = computed(() =>
  mergedParams(selectedTemplate.value).filter(p => p.type === 'schedule').map(p => p.key)
)
watch(() => scheduleKeys.value.map(k => params.value[k]), () => {
  scheduleKeys.value.forEach(key => {
    if (!scheduleSub.value[key]) {
      scheduleSub.value[key] = { time: '09:00', day: 1, hours: 24 }
    }
  })
})

// ── Step: fill params & render ────────────────────────────────────

const hasParameters = computed(() => mergedParams(selectedTemplate.value).length > 0)

async function handleRender() {
  if (!selectedTemplate.value) return
  rendering.value = true
  try {
    // Merge schedule sub-fields into params (e.g. params["time"] from scheduleSub["my_schedule"].time)
    const mergedParams: Record<string, any> = { ...params.value }
    for (const [key, sub] of Object.entries(scheduleSub.value)) {
      if (sub.time) mergedParams[key + '_time'] = sub.time
      if (sub.day !== undefined) mergedParams[key + '_day'] = sub.day
      if (sub.hours !== undefined) mergedParams[key + '_hours'] = sub.hours
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
  emit('apply', selectedTemplate.value, rendered.value.task_defaults, rendered.value.extra_schema)
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
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="emit('close')">
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-lg mx-4 max-h-[85vh] flex flex-col">
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
          <div v-else class="space-y-2">
            <div
              v-for="tmpl in templates" :key="tmpl.id"
              class="flex items-center gap-3 p-3 rounded-lg border border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700 cursor-pointer transition-colors"
              @click="selectTemplate(tmpl)"
            >
              <span class="text-xl">{{ tmpl.icon || '📋' }}</span>
              <div class="flex-1 min-w-0">
                <div class="font-medium text-sm text-gray-900 dark:text-gray-100">{{ tmpl.name }}</div>
              <div class="text-xs text-gray-400 truncate">{{ tmpl.description || td('taskKind.' + tmpl.kind) || tmpl.kind }}</div>
            </div>
              <span class="text-xs text-gray-400">{{ td('taskKind.' + tmpl.kind) || tmpl.kind }}</span>
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
              <template v-else-if="p.type === 'schedule'">
                <select v-model="params[p.key]"
                  class="w-full rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm">
                  <option value="once">一次性</option>
                  <option value="daily">每天</option>
                  <option value="weekly">每周</option>
                  <option value="monthly">每月</option>
                  <option value="interval">间隔</option>
                </select>
                <div v-if="params[p.key] && params[p.key] !== 'once'" class="flex gap-2 mt-1">
                  <input v-if="params[p.key] === 'daily'" v-model="scheduleSub[p.key].time" type="time"
                    class="flex-1 rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm" />
                  <template v-if="params[p.key] === 'weekly'">
                    <select v-model="scheduleSub[p.key].day"
                      class="flex-1 rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm">
                      <option :value="1">周一</option><option :value="2">周二</option><option :value="3">周三</option>
                      <option :value="4">周四</option><option :value="5">周五</option><option :value="6">周六</option><option :value="0">周日</option>
                    </select>
                    <input v-model="scheduleSub[p.key].time" type="time"
                      class="flex-1 rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm" />
                  </template>
                  <template v-if="params[p.key] === 'monthly'">
                    <input v-model.number="scheduleSub[p.key].day" type="number" min="1" max="31" placeholder="几号"
                      class="w-20 rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm" />
                    <input v-model="scheduleSub[p.key].time" type="time"
                      class="flex-1 rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm" />
                  </template>
                  <div v-if="params[p.key] === 'interval'" class="flex items-center gap-1 flex-1">
                    <input v-model.number="scheduleSub[p.key].hours" type="number" min="1" placeholder="间隔"
                      class="w-20 rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm" />
                    <span class="text-xs text-gray-400">小时</span>
                  </div>
                </div>
              </template>
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
