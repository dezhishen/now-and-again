<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { api } from '@/api/client'
import type { Todo } from '@/types'
import InspectionInspect from './InspectionInspect.vue'

const { t } = useI18n()

const props = defineProps<{ todo: Todo }>()
const emit = defineEmits<{ completed: [] }>()

const showModal = ref(false)
const selections = ref<Record<string, { item_id: string; item_name: string; branch_name: string }>>({})
const loading = ref(false)
const submiting = ref(false)
const fullTask = ref<any>(null)

async function openInspect() {
  showModal.value = true
  selections.value = {}
  loading.value = true
  try {
    const res = await api.get<{ todo: any; extra: any }>('/todos/' + props.todo.id + '?with_extra=true')
    fullTask.value = { task: res.todo?.task || props.todo.task, extra: res.extra }
  } catch { /* */ }
  finally { loading.value = false }
}

async function submit() {
  const sels = Object.entries(selections.value).map(([branchId, v]) => ({
    item_id: v.item_id,
    branch_id: branchId,
    item_name: v.item_name,
    branch_name: v.branch_name,
  }))
  if (sels.length === 0) return
  const displaySummary = sels.map(s => s.item_name + ':' + s.branch_name).join(', ')
  submiting.value = true
  try {
    await api.put('/todos/' + props.todo.id, {
      todo: { status: 'done' },
      extra: { selections: sels, display_summary: displaySummary },
    })
    showModal.value = false
    emit('completed')
  } catch { /* */ }
  finally { submiting.value = false }
}
</script>

<template>
  <button class="flex-1 text-xs py-1.5 rounded-lg bg-purple-50 dark:bg-purple-900/30 text-purple-700 dark:text-purple-300 hover:bg-purple-100 dark:hover:bg-purple-900/50 transition-colors font-medium" @click="openInspect">🔍 {{ t('todo.inspect') }}</button>

  <Teleport to="body">
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60" @mousedown.self="showModal = false">
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-2xl w-[90vw] max-w-lg max-h-[80vh] flex flex-col">
        <div class="flex items-center justify-between px-4 py-3 border-b dark:border-gray-700">
          <h3 class="font-bold dark:text-gray-200">🔍 {{ todo.task?.name }}</h3>
          <button class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 text-lg" @click="showModal = false">✕</button>
        </div>
        <div v-if="loading" class="flex-1 flex items-center justify-center py-8">
          <span class="animate-spin text-2xl">⏳</span>
        </div>
        <InspectionInspect
          v-else-if="fullTask"
          v-model:task="fullTask"
          v-model:selections="selections"
        />
        <div class="flex gap-2 px-4 py-3 border-t dark:border-gray-700">
          <button class="btn-primary text-sm flex-1" :disabled="submiting" @click="submit">
            {{ submiting ? t('todo.submitting') : t('todo.submitInspect') }}
          </button>
          <button class="btn-secondary" @click="showModal = false">{{ t('todo.cancel') }}</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>
