<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RouterView } from 'vue-router'
import AppHeader from '@/components/AppHeader.vue'
import ToastContainer from '@/components/ToastContainer.vue'
import ConfirmModal from '@/components/ConfirmModal.vue'
import { registerConfirmModal } from '@/composables/useConfirm'
import { appLoadingText } from '@/composables/useAppLoading'

const confirmRef = ref<InstanceType<typeof ConfirmModal> | null>(null)
onMounted(() => {
  if (confirmRef.value) registerConfirmModal(confirmRef.value)
})
</script>

<template>
  <div id="na-app" class="h-screen overflow-hidden flex flex-col">
    <AppHeader />
    <RouterView class="flex-1 overflow-hidden" />
    <ToastContainer />
    <ConfirmModal ref="confirmRef" />

    <!-- App-level loading overlay (shown during auth checks) -->
    <Transition name="fade">
      <div
        v-if="appLoadingText"
        class="fixed inset-0 z-[100] flex items-center justify-center bg-white dark:bg-gray-900"
      >
        <div class="flex flex-col items-center gap-3">
          <div class="w-8 h-8 border-2 border-primary border-t-transparent rounded-full animate-spin"></div>
          <span class="text-sm text-gray-500 dark:text-gray-400">{{ appLoadingText }}</span>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style>
.fade-enter-active, .fade-leave-active {
  transition: opacity 0.25s ease;
}
.fade-enter-from, .fade-leave-to {
  opacity: 0;
}
</style>
