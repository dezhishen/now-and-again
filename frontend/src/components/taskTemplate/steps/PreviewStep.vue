<script setup lang="ts">
import { computed } from 'vue'
import { getTaskKind } from '@/composables/useTaskKinds'
import * as yaml from 'js-yaml'

const props = defineProps<{
  templateCode: string
  name: string
  description: string
  kind: string
  icon: string
  sortOrder: number
  enabled: boolean
  parameters: any[]
  taskDefaults: any
  extraSchema: any
}>()

const kindPlugin = computed(() => getTaskKind(props.kind))
const kindLabel = computed(() => kindPlugin.value?.wizardStepLabel || props.kind)

const paramSummary = computed(() => {
  if (!props.parameters || props.parameters.length === 0) return '无'
  return props.parameters.map(p => `${p.label || '(未命名)'} (${p.type}${p.required ? ', 必填' : ''})`).join('、')
})

const taskDefaultsYaml = computed(() => {
  if (!props.taskDefaults || Object.keys(props.taskDefaults).length === 0) return '{}'
  return yaml.dump(props.taskDefaults, { lineWidth: -1 })
})

const extraSummary = computed(() => {
  if (!props.extraSchema || Object.keys(props.extraSchema).length === 0) return '无'
  const plugin = kindPlugin.value
  // If plugin has a display helper, delegate; otherwise show raw YAML
  if (plugin?.buildDisplaySummary) {
    return plugin.buildDisplaySummary({ task: {}, extra: props.extraSchema }) || '无'
  }
  return yaml.dump(props.extraSchema, { lineWidth: -1 })
})
</script>

<template>
  <div class="space-y-4">
    <h5 class="text-sm font-semibold text-gray-700 dark:text-gray-200">确认模板信息</h5>

    <div class="grid grid-cols-2 gap-3 text-sm">
      <div>
        <span class="text-gray-400">Code:</span>
        <code class="ml-1 text-primary">{{ templateCode }}</code>
      </div>
      <div>
        <span class="text-gray-400">名称:</span>
        <span class="ml-1 dark:text-gray-200">{{ name }}</span>
      </div>
      <div>
        <span class="text-gray-400">类型:</span>
        <span class="ml-1 dark:text-gray-200">{{ kindLabel }}</span>
      </div>
      <div>
        <span class="text-gray-400">图标:</span>
        <span class="ml-1">{{ icon || '—' }}</span>
      </div>
      <div>
        <span class="text-gray-400">排序:</span>
        <span class="ml-1 dark:text-gray-200">{{ sortOrder }}</span>
      </div>
      <div>
        <span class="text-gray-400">状态:</span>
        <span class="ml-1" :class="enabled ? 'text-green-600' : 'text-gray-400'">{{ enabled ? '启用' : '禁用' }}</span>
      </div>
    </div>

    <div v-if="description" class="text-sm">
      <span class="text-gray-400">描述:</span>
      <span class="ml-1 dark:text-gray-200">{{ description }}</span>
    </div>

    <div>
      <span class="text-sm text-gray-400">参数:</span>
      <span class="ml-1 text-sm dark:text-gray-200">{{ paramSummary }}</span>
    </div>

    <div>
      <span class="text-sm text-gray-400 block mb-1">task_defaults:</span>
      <pre class="bg-gray-50 dark:bg-gray-900 rounded-md p-3 text-xs font-mono text-gray-600 dark:text-gray-400 overflow-x-auto">{{ taskDefaultsYaml }}</pre>
    </div>

    <div>
      <span class="text-sm text-gray-400 block mb-1">extra_schema:</span>
      <pre class="bg-gray-50 dark:bg-gray-900 rounded-md p-3 text-xs font-mono text-gray-600 dark:text-gray-400 overflow-x-auto">{{ extraSummary }}</pre>
    </div>
  </div>
</template>
