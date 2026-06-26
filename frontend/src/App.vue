<script setup lang="ts">
import { RouterView } from 'vue-router'
import AppHeader from '@/components/AppHeader.vue'
import ToastContainer from '@/components/ToastContainer.vue'
</script>

<template>
  <div id="na-app" class="min-h-screen">
    <AppHeader />
    <RouterView />
    <ToastContainer />
  </div>
</template>
