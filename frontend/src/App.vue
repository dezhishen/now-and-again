<script setup lang="ts">
import { RouterView } from 'vue-router'
import AppHeader from '@/components/AppHeader.vue'
</script>

<template>
  <div id="na-app" class="min-h-screen">
    <AppHeader />
    <RouterView />
  </div>
</template>
