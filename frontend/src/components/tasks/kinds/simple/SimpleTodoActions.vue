<script setup lang="ts">
import type { Todo } from '@/types'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

defineProps<{ todo: Todo }>()
defineEmits<{ done: [todo: Todo]; skip: [todo: Todo]; remark: [todo: Todo] }>()
</script>

<template>
  <button class="flex-1 text-xs py-1.5 rounded-lg bg-green-50 dark:bg-green-900/30 text-green-700 dark:text-green-300 hover:bg-green-100 dark:hover:bg-green-900/50 transition-colors font-medium" @click="$emit('done', todo)">✅ {{ t('todo.quickDone') }}</button>
  <button class="flex-1 text-xs py-1.5 rounded-lg bg-blue-50 dark:bg-blue-900/30 text-blue-600 dark:text-blue-300 hover:bg-blue-100 dark:hover:bg-blue-900/50 transition-colors font-medium" @click="$emit('remark', todo)">📝 {{ t('todo.remark') }}</button>
  <button v-if="todo.task?.schedule_type !== 'once'" class="flex-1 text-xs py-1.5 rounded-lg bg-gray-50 dark:bg-gray-700/50 text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors font-medium" @click="$emit('skip', todo)">⏭️ {{ t('todo.skip') }}</button>
</template>
