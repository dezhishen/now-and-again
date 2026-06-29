<script setup lang="ts">
import { computed, ref } from 'vue'
import type { ApiRequestError } from '@/types'
import { translateFieldError } from '@/composables/useErrorHandler'
import { useI18n } from '@/i18n'

const { t } = useI18n()

const props = defineProps<{
  error: ApiRequestError | null
}>()

const emit = defineEmits<{
  close: []
}>()

const expanded = ref(false)

const isServerError = computed(() => props.error?.code === 'INTERNAL_ERROR')

function toggle() {
  expanded.value = !expanded.value
}
</script>

<template>
  <!-- 400-level: amber warning. 500-level: red danger. -->
  <div
    v-if="error"
    class="rounded-lg p-3 text-sm border"
    :class="isServerError
      ? 'bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800'
      : 'bg-amber-50 dark:bg-amber-900/20 border-amber-200 dark:border-amber-800'"
  >
    <div class="flex items-start justify-between gap-2">
      <div class="flex-1 min-w-0">
        <!-- Summary -->
        <p
          class="font-medium"
          :class="isServerError
            ? 'text-red-600 dark:text-red-400'
            : 'text-amber-600 dark:text-amber-400'"
        >
          {{ isServerError ? '⚠️ ' : '⚠️ ' }}{{ error.summary || error.message }}
        </p>
        <!-- Detail count + toggle -->
        <p
          v-if="error.details?.length"
          class="text-xs mt-1"
          :class="isServerError
            ? 'text-red-400 dark:text-red-500'
            : 'text-amber-400 dark:text-amber-500'"
        >
          {{ error.details.length }} 个字段存在问题
          <button
            class="underline hover:opacity-80 ml-1"
            :class="isServerError
              ? 'hover:text-red-600 dark:hover:text-red-300'
              : 'hover:text-amber-600 dark:hover:text-amber-300'"
            @click="toggle"
          >
            {{ expanded ? '收起' : '展开' }}
          </button>
        </p>
        <!-- 500 extra hint -->
        <p v-if="isServerError" class="text-xs text-red-400 dark:text-red-500 mt-1">
          请稍后重试，如持续出现请联系管理员
        </p>
      </div>
      <button
        class="flex-shrink-0"
        :class="isServerError
          ? 'text-red-400 hover:text-red-600 dark:hover:text-red-300'
          : 'text-amber-400 hover:text-amber-600 dark:hover:text-amber-300'"
        @click="emit('close')"
      >✕</button>
    </div>

    <!-- Detail list — v-if to avoid rendering when collapsed -->
    <ul
      v-if="expanded && error.details?.length"
      class="mt-2 space-y-1 pl-4 border-l-2"
      :class="isServerError
        ? 'border-red-200 dark:border-red-800'
        : 'border-amber-200 dark:border-amber-800'"
    >
      <li
        v-for="(f, i) in error.details"
        :key="`${f.field}-${i}`"
        class="text-xs"
        :class="isServerError
          ? 'text-red-500 dark:text-red-400'
          : 'text-amber-600 dark:text-amber-400'"
      >
        {{ translateFieldError(f, t) }}
      </li>
    </ul>
  </div>
</template>
