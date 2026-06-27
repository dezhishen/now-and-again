<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RouterView } from 'vue-router'
import AppHeader from '@/components/AppHeader.vue'
import ToastContainer from '@/components/ToastContainer.vue'
import ConfirmModal from '@/components/ConfirmModal.vue'
import { registerConfirmModal } from '@/composables/useConfirm'

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
  </div>
</template>
