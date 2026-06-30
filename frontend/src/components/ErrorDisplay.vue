
<script setup lang="ts">
import { computed, ref, watch, onMounted, onUnmounted } from 'vue'
import type { ApiRequestError } from '@/types'
import { ERROR_HANDLERS, translateFieldError, getDisplayMode, getSeverity } from '@/composables/useErrorHandler'
import type { DisplayMode, Severity } from '@/composables/useErrorHandler'
import { useI18n } from '@/i18n'

const { t } = useI18n()

const props = withDefaults(defineProps<{
  error: ApiRequestError | null
  /** Override the registered display mode. Omit to use registry default. */
  mode?: DisplayMode
}>(), {})

const emit = defineEmits<{
  close: []
}>()

const expanded = ref(false)
const showDialog = ref(false)
const showToast = ref(false)
let toastTimer: ReturnType<typeof setTimeout> | null = null

// ── Resolved display mode ────────────────────────────────────────

const resolvedMode = computed<DisplayMode>(() => {
  if (props.mode) return props.mode
  if (!props.error) return 'inline'
  return getDisplayMode(props.error.code)
})

const isDialog = computed(() => resolvedMode.value === 'dialog')
const isInline = computed(() => resolvedMode.value === 'inline')

const severity = computed<Severity>(() => {
  if (!props.error) return 'warning'
  return getSeverity(props.error.code)
})

// ── Severity → Tailwind classes ───────────────────────────────────

const toastClasses = computed(() => {
  switch (severity.value) {
    case 'info':    return 'bg-blue-500 text-white'
    case 'warning': return 'bg-amber-500 text-white'
    case 'error':   return 'bg-red-500 text-white'
    case 'success': return 'bg-green-500 text-white'
  }
})

const inlineClasses = computed(() => {
  switch (severity.value) {
    case 'info':    return 'bg-blue-50 dark:bg-blue-900/20 border-blue-200 dark:border-blue-800'
    case 'warning': return 'bg-amber-50 dark:bg-amber-900/20 border-amber-200 dark:border-amber-800'
    case 'error':   return 'bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800'
    case 'success': return 'bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800'
  }
})

const textMutedClasses = computed(() => {
  switch (severity.value) {
    case 'info':    return 'text-blue-400 dark:text-blue-500'
    case 'warning': return 'text-amber-400 dark:text-amber-500'
    case 'error':   return 'text-red-400 dark:text-red-500'
    case 'success': return 'text-green-400 dark:text-green-500'
  }
})

const textStrongClasses = computed(() => {
  switch (severity.value) {
    case 'info':    return 'text-blue-600 dark:text-blue-400'
    case 'warning': return 'text-amber-600 dark:text-amber-400'
    case 'error':   return 'text-red-600 dark:text-red-400'
    case 'success': return 'text-green-600 dark:text-green-400'
  }
})

/** i18n summary from ERROR_HANDLERS registry. */
const i18nSummary = computed(() => {
  if (!props.error) return ''
  const handler = ERROR_HANDLERS[props.error.code]
  return handler ? handler(props.error, t) : (props.error.summary || props.error.message)
})

function toggle() {
  expanded.value = !expanded.value
}

function dismiss() {
  if (isDialog.value) {
    showDialog.value = false
  }
  emit('close')
}

// ── Auto-clear stale error on mount (dialog reopened) ─────────────

onMounted(() => {
  if (props.error) emit('close')
})

onUnmounted(() => {
  if (toastTimer) clearTimeout(toastTimer)
})

// ── Toast mode: centered, auto-dismiss ────────────────────────────

watch(() => props.error, (err) => {
  if (err && resolvedMode.value === 'toast') {
    if (toastTimer) clearTimeout(toastTimer)
    showToast.value = true
    toastTimer = setTimeout(() => {
      showToast.value = false
      emit('close')
    }, 3000)
  }
}, { immediate: true })

// ── Dialog mode: show modal when error is set ─────────────────────

watch(() => props.error, (err) => {
  if (err && isDialog.value) {
    showDialog.value = true
  }
})
</script>

<template>
  <!-- ── Toast mode: centered overlay, auto-dismiss 3s ── -->
  <Teleport v-if="resolvedMode === 'toast' && error && showToast" to="body">
    <Transition name="toast-fade">
      <div
        class="fixed top-6 left-1/2 -translate-x-1/2 z-50 px-5 py-3 rounded-lg shadow-lg text-sm max-w-md w-[90vw] text-center"
        :class="toastClasses"
      >
        {{ i18nSummary }}
      </div>
    </Transition>
  </Teleport>

  <!-- ── Dialog mode: modal overlay ── -->
  <div
    v-else-if="isDialog && error && showDialog"
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
    @click.self="dismiss"
  >
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-md mx-4 p-6">
      <div class="flex items-start justify-between mb-3">
        <h4 class="text-base font-semibold" :class="textStrongClasses">
          {{ severity === 'error' ? '服务器错误' : '操作提示' }}
        </h4>
        <button class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300" @click="dismiss">✕</button>
      </div>
      <p class="text-sm text-gray-700 dark:text-gray-300 mb-2">{{ i18nSummary }}</p>
      <p v-if="severity === 'error'" class="text-xs text-red-400 dark:text-red-500 mb-3">
        请稍后重试，如持续出现请联系管理员
      </p>

      <!-- Field details in dialog -->
      <ul v-if="error.details?.length" class="space-y-1 pl-4 border-l-2 border-gray-200 dark:border-gray-700 mb-3">
        <li
          v-for="(f, i) in error.details" :key="`${f.field}-${i}`"
          class="text-xs text-gray-500 dark:text-gray-400"
        >
          {{ translateFieldError(f, t) }}
        </li>
      </ul>

      <div class="flex justify-end">
        <button
          class="px-4 py-2 text-sm rounded-md bg-green-500 hover:bg-green-600 text-white font-medium transition-colors"
          @click="dismiss"
        >确定</button>
      </div>
    </div>
  </div>

  <!-- ── Inline mode ── -->
  <div
    v-else-if="isInline && error"
    class="rounded-lg p-3 text-sm border"
    :class="inlineClasses"
  >
    <div class="flex items-start justify-between gap-2">
      <div class="flex-1 min-w-0">
        <p class="font-medium" :class="textStrongClasses">
          {{ i18nSummary }}
        </p>
        <p
          v-if="error.details?.length"
          class="text-xs mt-1"
          :class="textMutedClasses"
        >
          {{ error.details.length }} 个字段存在问题
          <button class="underline hover:opacity-80 ml-1" @click="toggle">
            {{ expanded ? '收起' : '展开' }}
          </button>
        </p>
        <p v-if="severity === 'error'" class="text-xs text-red-400 dark:text-red-500 mt-1">
          请稍后重试，如持续出现请联系管理员
        </p>
      </div>
      <button class="flex-shrink-0" :class="textMutedClasses" @click="emit('close')">✕</button>
    </div>

    <ul
      v-if="expanded && error.details?.length"
      class="mt-2 space-y-1 pl-4 border-l-2"
      :class="inlineClasses"
    >
      <li
        v-for="(f, i) in error.details"
        :key="`${f.field}-${i}`"
        class="text-xs"
        :class="textStrongClasses"
      >
        {{ translateFieldError(f, t) }}
      </li>
    </ul>
  </div>
</template>

<style>
.toast-fade-enter-active {
  transition: all 0.3s ease-out;
}
.toast-fade-leave-active {
  transition: all 0.25s ease-in;
}
.toast-fade-enter-from {
  opacity: 0;
  transform: translate(-50%, -12px);
}
.toast-fade-leave-to {
  opacity: 0;
  transform: translate(-50%, -8px);
}
</style>
