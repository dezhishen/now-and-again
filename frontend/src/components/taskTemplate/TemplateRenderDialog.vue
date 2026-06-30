<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from '@/i18n'
import { useErrorHandler } from '@/composables/useErrorHandler'
import ErrorDisplay from '@/components/ErrorDisplay.vue'
import { renderTemplate } from '@/api/task-templates'
import type { TaskTemplate, TemplateParameter } from '@/types'

const { t } = useI18n()
const { error, setError, clearError } = useErrorHandler()

const props = defineProps<{
  template: TaskTemplate
}>()

const emit = defineEmits<{
  close: []
  create: [taskDefaults: any, extraSchema: any]
}>()

const params = ref<Record<string, any>>({})
const rendered = ref<any>(null)
const rendering = ref(false)

// Initialize default values
props.template.parameters?.forEach(p => {
  if (p.default !== undefined) {
    params.value[p.key] = p.default
  } else if (p.type === 'bool') {
    params.value[p.key] = false
  } else if (p.type === 'int' || p.type === 'float') {
    params.value[p.key] = 0
  } else {
    params.value[p.key] = ''
  }
})

const hasParameters = computed(() => (props.template.parameters?.length || 0) > 0)

async function handleRender() {
  rendering.value = true
  try {
    const result = await renderTemplate(props.template.template_code, params.value)
    rendered.value = result
  } catch (e: any) {
    setError(e)
  } finally {
    rendering.value = false
  }
}

function handleCreate() {
  if (!rendered.value) return
  emit('create', rendered.value.task_defaults, rendered.value.extra_schema)
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
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-lg mx-4 p-6">
      <ErrorDisplay :error="error" @close="clearError" />
      <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
        {{ template.name }}
      </h3>

      <!-- Parameters -->
      <div v-if="hasParameters" class="mb-4 space-y-3">
        <h4 class="text-sm font-medium text-gray-600 dark:text-gray-400">
          {{ t('taskTemplate.parameters') }}
        </h4>
        <div v-for="p in template.parameters" :key="p.key" class="flex flex-col gap-1">
          <label class="text-sm font-medium text-gray-700 dark:text-gray-300">
            {{ p.label }}
            <span v-if="p.required" class="text-red-500">*</span>
          </label>
          <p v-if="p.description" class="text-xs text-gray-400">{{ p.description }}</p>

          <!-- Select -->
          <select
            v-if="p.type === 'select'"
            v-model="params[p.key]"
            class="w-full rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm text-gray-900 dark:text-gray-100"
          >
            <option v-for="opt in p.options" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
          </select>

          <!-- Checkbox -->
          <label v-else-if="p.type === 'bool'" class="flex items-center gap-2 cursor-pointer">
            <input
              v-model="params[p.key]"
              type="checkbox"
              class="rounded border-gray-300 dark:border-gray-600 text-green-500 focus:ring-green-500"
            />
            <span class="text-sm text-gray-700 dark:text-gray-300">{{ p.label }}</span>
          </label>

          <!-- Text / Number -->
          <input
            v-else
            v-model="params[p.key]"
            :type="inputType(p)"
            :placeholder="p.placeholder || ''"
            :required="p.required"
            class="w-full rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm text-gray-900 dark:text-gray-100"
          />
        </div>
      </div>

      <p v-else class="text-sm text-gray-400 dark:text-gray-500 mb-4">
        {{ t('taskTemplate.noParameters') }}
      </p>

      <!-- Actions -->
      <div class="flex items-center justify-between gap-3">
        <button
          v-if="!rendered"
          class="px-4 py-2 rounded-md bg-green-500 hover:bg-green-600 text-white text-sm font-medium disabled:opacity-50 transition-colors"
          :disabled="rendering"
          @click="handleRender"
        >
          {{ rendering ? '...' : t('taskTemplate.preview') }}
        </button>

        <template v-else>
          <button
            class="px-4 py-2 rounded-md border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 text-sm hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
            @click="rendered = null"
          >
            ← {{ t('taskTemplate.parameters') }}
          </button>
          <button
            class="px-4 py-2 rounded-md bg-green-500 hover:bg-green-600 text-white text-sm font-medium transition-colors"
            @click="handleCreate"
          >
            {{ t('taskTemplate.createTask') }}
          </button>
        </template>
      </div>

      <!-- Rendered Preview -->
      <div v-if="rendered" class="mt-4 p-3 bg-gray-50 dark:bg-gray-900 rounded-md border border-gray-200 dark:border-gray-700">
        <h4 class="text-sm font-medium text-gray-600 dark:text-gray-400 mb-2">{{ t('taskTemplate.preview') }}</h4>
        <pre class="text-xs text-gray-700 dark:text-gray-300 overflow-x-auto">{{ JSON.stringify(rendered.task_defaults, null, 2) }}</pre>
      </div>
    </div>
  </div>
</template>
