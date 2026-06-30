<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { useI18n } from '@/i18n'
import { useToast } from '@/composables/useToast'
import { useErrorHandler } from '@/composables/useErrorHandler'
import ErrorDisplay from '@/components/ErrorDisplay.vue'
import { createFamilyTemplate, updateFamilyTemplate } from '@/api/task-templates'
import type { TaskTemplate, CreateTaskTemplateRequest, UpdateTaskTemplateRequest } from '@/types'
import * as yaml from 'js-yaml'

const { t } = useI18n()
const toast = useToast()
const { error, setError, clearError } = useErrorHandler()

const props = defineProps<{
  editing?: TaskTemplate | null
}>()

const emit = defineEmits<{
  close: []
  saved: []
}>()

const saving = ref(false)

const form = reactive<CreateTaskTemplateRequest>({
  template_code: '',
  name: '',
  description: '',
  kind: 'simple',
  icon: '',
  sort_order: 0,
  enabled: true,
  parameters: [],
  task_defaults: {},
  extra_schema: {},
})

// Initialize from editing template
watch(() => props.editing, (tmpl) => {
  if (tmpl) {
    form.template_code = tmpl.template_code
    form.name = tmpl.name
    form.description = tmpl.description || ''
    form.kind = tmpl.kind
    form.icon = tmpl.icon || ''
    form.sort_order = tmpl.sort_order
    form.enabled = tmpl.enabled
    form.parameters = tmpl.parameters ? JSON.parse(JSON.stringify(tmpl.parameters)) : []
    form.task_defaults = tmpl.task_defaults ? JSON.parse(JSON.stringify(tmpl.task_defaults)) : {}
    form.extra_schema = tmpl.extra_schema ? JSON.parse(JSON.stringify(tmpl.extra_schema)) : {}
  }
}, { immediate: true })

// Parameter editor
function addParam() {
  form.parameters!.push({
    key: '',
    label: '',
    type: 'string',
    description: '',
    required: false,
    placeholder: '',
  })
}

function removeParam(index: number) {
  form.parameters!.splice(index, 1)
}

// YAML editors
const taskDefaultsYaml = ref('')
const extraSchemaYaml = ref('')
watch(() => form.task_defaults, (v) => { taskDefaultsYaml.value = yaml.dump(v, { lineWidth: -1 }) })
watch(() => form.extra_schema, (v) => { extraSchemaYaml.value = yaml.dump(v, { lineWidth: -1 }) })

function parseYaml(str: string, fallback: any): any {
  try { return yaml.load(str) || fallback } catch { return fallback }
}

async function handleSave() {
  saving.value = true
  try {
    // Parse YAML fields before submit
    const req: any = { ...form }
    req.task_defaults = parseYaml(taskDefaultsYaml.value, form.task_defaults)
    req.extra_schema = parseYaml(extraSchemaYaml.value, form.extra_schema)

    if (props.editing) {
      const upd: UpdateTaskTemplateRequest = {
        name: req.name,
        description: req.description,
        kind: req.kind,
        icon: req.icon,
        sort_order: req.sort_order,
        enabled: req.enabled,
        parameters: req.parameters,
        task_defaults: req.task_defaults,
        extra_schema: req.extra_schema,
      }
      await updateFamilyTemplate(props.editing.template_code, upd)
      toast.success(t('taskTemplate.updated'))
    } else {
      await createFamilyTemplate(req as CreateTaskTemplateRequest)
      toast.success(t('taskTemplate.created'))
    }
    emit('saved')
  } catch (e: any) {
    setError(e)
  } finally {
    saving.value = false
  }
}

const kindOptions = [
  { value: 'simple', label: '简单任务' },
  { value: 'inspection', label: '巡检任务' },
]

const typeOptions = [
  { value: 'string', label: '文本' },
  { value: 'int', label: '整数' },
  { value: 'float', label: '小数' },
  { value: 'bool', label: '布尔' },
  { value: 'select', label: '下拉选择' },
]
</script>

<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="emit('close')">
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-2xl mx-4 max-h-[90vh] overflow-y-auto p-6">
      <ErrorDisplay :error="error" @close="clearError" />
      <h4 class="text-base font-semibold text-gray-900 dark:text-gray-100 mb-4">
        {{ editing ? t('taskTemplate.editFamily') : t('taskTemplate.createFamily') }}
      </h4>

      <div class="space-y-4">
        <!-- Basic fields -->
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Code *</label>
            <input v-model="form.template_code" required
              class="w-full rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm text-gray-900 dark:text-gray-100"
              :disabled="!!editing" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">名称 *</label>
            <input v-model="form.name" required
              class="w-full rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm text-gray-900 dark:text-gray-100" />
          </div>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">描述</label>
          <input v-model="form.description"
            class="w-full rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm text-gray-900 dark:text-gray-100" />
        </div>

        <div class="grid grid-cols-3 gap-3">
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('taskTemplate.kind') }}</label>
            <select v-model="form.kind"
              class="w-full rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm text-gray-900 dark:text-gray-100">
              <option v-for="k in kindOptions" :key="k.value" :value="k.value">{{ k.label }}</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">图标</label>
            <input v-model="form.icon" placeholder="📋"
              class="w-full rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm text-gray-900 dark:text-gray-100" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">排序</label>
            <input v-model.number="form.sort_order" type="number"
              class="w-full rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm text-gray-900 dark:text-gray-100" />
          </div>
        </div>

        <label class="flex items-center gap-2 cursor-pointer">
          <input v-model="form.enabled" type="checkbox" class="rounded border-gray-300 text-green-500 focus:ring-green-500" />
          <span class="text-sm text-gray-700 dark:text-gray-300">启用</span>
        </label>

        <!-- Parameters -->
        <div>
          <div class="flex justify-between items-center mb-2">
            <label class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('taskTemplate.parameters') }}</label>
            <button class="px-3 py-1 text-xs rounded-md bg-green-500 hover:bg-green-600 text-white" @click="addParam">+ 添加</button>
          </div>
          <div v-if="form.parameters!.length === 0" class="text-xs text-gray-400">{{ t('taskTemplate.noParameters') }}</div>
          <div v-for="(p, i) in form.parameters" :key="i" class="border dark:border-gray-700 rounded-md p-3 mb-2">
            <div class="grid grid-cols-4 gap-2 mb-2">
              <input v-model="p.key" placeholder="key" class="col-span-1 rounded border dark:border-gray-600 bg-white dark:bg-gray-700 px-2 py-1 text-xs" />
              <input v-model="p.label" placeholder="标签" class="col-span-1 rounded border dark:border-gray-600 bg-white dark:bg-gray-700 px-2 py-1 text-xs" />
              <select v-model="p.type" class="col-span-1 rounded border dark:border-gray-600 bg-white dark:bg-gray-700 px-2 py-1 text-xs">
                <option v-for="opt in typeOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
              </select>
              <button class="col-span-1 text-xs text-red-500 hover:text-red-700" @click="removeParam(i)">删除</button>
            </div>
            <input v-model="p.description" placeholder="描述" class="w-full rounded border dark:border-gray-600 bg-white dark:bg-gray-700 px-2 py-1 text-xs mb-1" />
            <div class="flex items-center gap-4">
              <label class="flex items-center gap-1 text-xs"><input v-model="p.required" type="checkbox" class="rounded" /> 必填</label>
              <input v-model="p.placeholder" placeholder="占位符" class="flex-1 rounded border dark:border-gray-600 bg-white dark:bg-gray-700 px-2 py-1 text-xs" />
            </div>
          </div>
        </div>

        <!-- YAML editors -->
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">task_defaults (YAML)</label>
          <textarea v-model="taskDefaultsYaml" rows="4"
            class="w-full rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-xs text-gray-900 dark:text-gray-100 font-mono"></textarea>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">extra_schema (YAML)</label>
          <textarea v-model="extraSchemaYaml" rows="4"
            class="w-full rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-xs text-gray-900 dark:text-gray-100 font-mono"></textarea>
        </div>
      </div>

      <div class="flex justify-end gap-2 mt-6">
        <button
          class="px-4 py-2 text-sm rounded-md border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
          @click="emit('close')">{{ t('confirm.cancel') }}</button>
        <button
          class="px-4 py-2 text-sm rounded-md bg-green-500 hover:bg-green-600 text-white font-medium disabled:opacity-50 transition-colors"
          :disabled="saving || !form.template_code || !form.name"
          @click="handleSave">{{ saving ? '...' : t('confirm.ok') }}</button>
      </div>
    </div>
  </div>
</template>
