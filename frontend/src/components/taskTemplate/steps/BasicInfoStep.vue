<script setup lang="ts">
import { useI18n } from '@/i18n'
import { getTaskKind } from '@/composables/useTaskKinds'

const { t } = useI18n()

const props = defineProps<{
  editing?: any
}>()

const templateCode = defineModel<string>('templateCode', { required: true })
const name = defineModel<string>('name', { required: true })
const description = defineModel<string>('description')
const kind = defineModel<string>('kind', { required: true })
const icon = defineModel<string>('icon')
const sortOrder = defineModel<number>('sortOrder')
const enabled = defineModel<boolean>('enabled')

const codeManuallyEdited = defineModel<boolean>('codeManuallyEdited')

const kindOptions = [
  { value: 'simple', label: '简单任务' },
  { value: 'inspection', label: '巡检任务' },
]

function slugify(text: string): string {
  const hasLatin = /[a-zA-Z]/.test(text)
  if (!hasLatin) return ''
  return text.toLowerCase().replace(/[^a-z0-9]+/g, '_').replace(/^_|_$/g, '').replace(/_+/g, '_').substring(0, 32)
}

function onNameChange() {
  if (props.editing) return
  if (codeManuallyEdited.value) return
  const slug = slugify(name.value)
  templateCode.value = 'tt_' + (slug || Date.now().toString(36))
}

function regenerateCode() {
  codeManuallyEdited.value = false
  templateCode.value = 'tt_' + (slugify(name.value) || Date.now().toString(36))
}

function onCodeInput() {
  codeManuallyEdited.value = true
}
</script>

<template>
  <div class="space-y-4">
    <div class="grid grid-cols-2 gap-3">
      <!-- Template Code -->
      <div>
        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Code *</label>
        <div class="flex gap-1">
          <input v-model="templateCode" required
            class="flex-1 rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm text-gray-900 dark:text-gray-100"
            :disabled="!!editing"
            @input="onCodeInput" />
          <button v-if="!editing" type="button"
            class="px-2 rounded-md border border-gray-300 dark:border-gray-600 text-gray-500 hover:text-gray-700 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 text-xs"
            title="根据名称重新生成"
            @click="regenerateCode">↻</button>
        </div>
        <p v-if="!editing && !codeManuallyEdited" class="text-xs text-gray-400 mt-0.5">根据名称自动生成，也可手动修改</p>
      </div>

      <!-- Name -->
      <div>
        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">名称 *</label>
        <input v-model="name" required
          class="w-full rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm text-gray-900 dark:text-gray-100"
          @input="onNameChange" />
      </div>
    </div>

    <!-- Description -->
    <div>
      <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">描述</label>
      <input v-model="description"
        class="w-full rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm text-gray-900 dark:text-gray-100" />
    </div>

    <!-- Kind / Icon / Sort -->
    <div class="grid grid-cols-3 gap-3">
      <div>
        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('taskTemplate.kind') }}</label>
        <select v-model="kind"
          class="w-full rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm text-gray-900 dark:text-gray-100">
          <option v-for="k in kindOptions" :key="k.value" :value="k.value">{{ k.label }}</option>
        </select>
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">图标</label>
        <input v-model="icon" placeholder="📋"
          class="w-full rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm text-gray-900 dark:text-gray-100" />
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">排序</label>
        <input v-model.number="sortOrder" type="number"
          class="w-full rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm text-gray-900 dark:text-gray-100" />
      </div>
    </div>

    <label class="flex items-center gap-2 cursor-pointer">
      <input v-model="enabled" type="checkbox" class="rounded border-gray-300 text-green-500 focus:ring-green-500" />
      <span class="text-sm text-gray-700 dark:text-gray-300">启用</span>
    </label>
  </div>
</template>
