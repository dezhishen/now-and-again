<script setup lang="ts">
import { computed } from 'vue'
import * as yaml from 'js-yaml'
import type { TemplateParameter } from '@/types'

const props = defineProps<{
  parameters?: TemplateParameter[]
}>()

const taskDefaults = defineModel<any>('taskDefaults', { required: true })
const extraSchema = defineModel<any>('extraSchema', { required: true })

const taskDefaultsYaml = computed({
  get: () => yaml.dump(taskDefaults.value || {}, { lineWidth: -1 }),
  set: (v) => { try { taskDefaults.value = yaml.load(v) || {} } catch {} },
})

const extraSchemaYaml = computed({
  get: () => yaml.dump(extraSchema.value || {}, { lineWidth: -1 }),
  set: (v) => { try { extraSchema.value = yaml.load(v) || {} } catch {} },
})

const hasParams = computed(() => props.parameters && props.parameters.length > 0 && props.parameters.some(p => p.key))

const paramTags = computed(() => {
  return (props.parameters || []).filter(p => p.key).map(p => ({
    key: p.key,
    label: '\u007b\u007b.' + p.key + '\u007d\u007d',
  }))
})

const exampleSnippet = computed(() => {
  const first = (props.parameters || []).find(p => p.key)
  const key = first?.key || 'param'
  return 'name: "' + '\u007b\u007b.' + key + '\u007d\u007d - \u6bcf\u65e5\u68c0\u67e5"'
})
</script>

<template>
  <div class="space-y-4">
    <div v-if="hasParams" class="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-md p-3">
      <p class="text-xs font-medium text-blue-700 dark:text-blue-300 mb-2">可用参数（在 task_defaults 中使用 Go template 语法引用）：</p>
      <div class="flex flex-wrap gap-1.5">
        <code
          v-for="p in paramTags" :key="p.key"
          class="text-xs bg-blue-100 dark:bg-blue-800 text-blue-700 dark:text-blue-300 px-1.5 py-0.5 rounded"
        >{{ p.label }}</code>
      </div>
      <p class="text-xs text-blue-500 dark:text-blue-400 mt-2">示例：<code class="bg-blue-100 dark:bg-blue-800 px-1 rounded">{{ exampleSnippet }}</code></p>
    </div>
    <div v-else class="bg-gray-50 dark:bg-gray-700/50 border dark:border-gray-600 rounded-md p-3">
      <p class="text-xs text-gray-500 dark:text-gray-400">
        先在参数配置步骤中定义参数，此处即可引用。在 task_defaults 中使用模板语法引用参数。
      </p>
    </div>

    <p class="text-xs text-gray-400 dark:text-gray-500">
      该任务类型暂无可视化配置器，请直接编辑 YAML。
    </p>

    <div>
      <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">task_defaults (YAML)</label>
      <textarea v-model="taskDefaultsYaml" rows="6"
        class="w-full rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-xs text-gray-900 dark:text-gray-100 font-mono"></textarea>
      <p class="text-xs text-gray-400 mt-1">name 支持模板变量、schedule_type (once/daily/weekly/monthly/interval)、schedule_data 等</p>
    </div>
    <div>
      <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">extra_schema (YAML)</label>
      <textarea v-model="extraSchemaYaml" rows="6"
        class="w-full rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-xs text-gray-900 dark:text-gray-100 font-mono"></textarea>
      <p class="text-xs text-gray-400 mt-1">类型特定的额外配置（如巡检项 check_items 等）</p>
    </div>
  </div>
</template>
