<script setup lang="ts">
import { useI18n } from '@/i18n'
import type { TemplateParameter } from '@/types'

const { t } = useI18n()

const parameters = defineModel<TemplateParameter[]>('parameters', { required: true })

const typeOptions = [
  { value: 'string', label: '文本' },
  { value: 'int', label: '整数' },
  { value: 'float', label: '小数' },
  { value: 'bool', label: '布尔' },
  { value: 'time', label: '时间' },
  { value: 'select', label: '下拉选择' },
  { value: 'location', label: '地点' },
  { value: 'schedule', label: '调度' },
  { value: 'group', label: '小组' },
]

function slugify(text: string): string {
  const hasLatin = /[a-zA-Z]/.test(text)
  if (!hasLatin) return ''
  return text.toLowerCase().replace(/[^a-z0-9]+/g, '_').replace(/^_|_$/g, '').replace(/_+/g, '_').substring(0, 32)
}

function addParam() {
  parameters.value.push({
    key: '',
    label: '',
    type: 'string',
    description: '',
    required: false,
    placeholder: '',
  })
}

function removeParam(index: number) {
  parameters.value.splice(index, 1)
}

function onParamLabelInput(idx: number) {
  const p = parameters.value[idx]
  if (!p.key && p.label) {
    p.key = slugify(p.label) || 'param_' + idx
  }
}

// Ensure we always have an array
if (!parameters.value) parameters.value = []
</script>

<template>
  <div>
    <div class="flex justify-between items-center mb-2">
      <label class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('taskTemplate.parameters') }}</label>
      <button class="px-3 py-1 text-xs rounded-md bg-green-500 hover:bg-green-600 text-white" @click="addParam">+ 添加</button>
    </div>
    <div v-if="parameters.length === 0" class="text-xs text-gray-400">{{ t('taskTemplate.noParameters') }}</div>
    <div v-for="(p, i) in parameters" :key="i" class="border dark:border-gray-700 rounded-md p-3 mb-2">
      <div class="grid grid-cols-4 gap-2 mb-2">
        <input v-model="p.key" placeholder="key" class="col-span-1 rounded border dark:border-gray-600 bg-white dark:bg-gray-700 px-2 py-1 text-xs" />
        <input v-model="p.label" placeholder="标签" class="col-span-1 rounded border dark:border-gray-600 bg-white dark:bg-gray-700 px-2 py-1 text-xs" @input="onParamLabelInput(i)" />
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
      <!-- Options for select type -->
      <div v-if="p.type === 'select'" class="mt-2">
        <p class="text-xs text-gray-400 mb-1">选项（每行一个，格式：label: value）</p>
        <textarea
          :value="(p.options || []).map((o: any) => o.label + ': ' + o.value).join('\n')"
          @input="(e: any) => { const lines = (e.target as HTMLTextAreaElement).value.split('\n').filter(Boolean); p.options = lines.map((l: string) => { const [label, ...rest] = l.split(':'); return { label: label.trim(), value: (rest.join(':').trim() || label.trim()) } }) }"
          rows="3"
          class="w-full rounded border dark:border-gray-600 bg-white dark:bg-gray-700 px-2 py-1 text-xs"
          placeholder="正常: normal&#10;异常: abnormal"></textarea>
      </div>
    </div>
  </div>
</template>
