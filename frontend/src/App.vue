<script setup lang="ts">
import { RouterView } from 'vue-router'
import AppHeader from '@/components/AppHeader.vue'
import ToastContainer from '@/components/ToastContainer.vue'
</script>

<template>
  <div id="na-app" class="h-screen overflow-hidden flex flex-col">
    <AppHeader />
    <RouterView class="flex-1 overflow-hidden" />
    <ToastContainer />
  </div>
</template>
