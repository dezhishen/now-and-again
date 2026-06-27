<template>
  <Teleport to="body">
    <div v-if="visible" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40" @mousedown.self="cancel">
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl p-6 w-full max-w-sm mx-4">
        <p class="text-sm text-gray-700 dark:text-gray-200 mb-6">{{ message }}</p>
        <div class="flex gap-2 justify-end">
          <button class="btn-secondary" @click="cancel">{{ t('confirm.cancel') }}</button>
          <button class="btn-danger" ref="confirmBtn" @click="ok">{{ t('confirm.ok') }}</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const visible = ref(false)
const message = ref('')
let resolvePromise: ((value: boolean) => void) | null = null

function show(msg: string): Promise<boolean> {
  message.value = msg
  visible.value = true
  return new Promise((resolve) => {
    resolvePromise = resolve
  })
}

function ok() {
  visible.value = false
  resolvePromise?.(true)
  resolvePromise = null
}

function cancel() {
  visible.value = false
  resolvePromise?.(false)
  resolvePromise = null
}

defineExpose({ show })
</script>
