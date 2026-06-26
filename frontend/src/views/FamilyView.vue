<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const showMenu = ref(false)
</script>

<template>
  <div class="flex flex-col md:flex-row min-h-screen">
    <!-- Mobile hamburger -->
    <button class="md:hidden fixed top-3 left-3 z-40 w-9 h-9 rounded bg-gray-200 dark:bg-gray-700 flex items-center justify-center shadow" @click="showMenu = !showMenu">
      {{ showMenu ? '✕' : '☰' }}
    </button>

    <!-- Sidebar -->
    <aside
      class="w-[200px] bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 p-4 flex-shrink-0 transition-transform"
      :class="showMenu ? 'fixed inset-y-0 left-0 z-30 translate-x-0' : 'max-md:fixed max-md:inset-y-0 max-md:left-0 max-md:z-30 max-md:-translate-x-full'"
      @click="showMenu = false"
    >
      <h3 class="text-lg font-semibold mb-3 dark:text-gray-200">{{ t('nav.family') }}</h3>
      <nav class="flex flex-col gap-1">
        <router-link :to="`/family/${$route.params.familyId}`" class="px-3 py-2 rounded-lg text-gray-700 dark:text-gray-300 hover:bg-primary hover:text-white transition-colors">{{ t('nav.dashboard') }}</router-link>
        <router-link :to="`/family/${$route.params.familyId}/groups`" class="px-3 py-2 rounded-lg text-gray-700 dark:text-gray-300 hover:bg-primary hover:text-white transition-colors">{{ t('nav.groups') }}</router-link>
        <router-link :to="`/family/${$route.params.familyId}/members`" class="px-3 py-2 rounded-lg text-gray-700 dark:text-gray-300 hover:bg-primary hover:text-white transition-colors">{{ t('nav.members') }}</router-link>
        <router-link :to="`/family/${$route.params.familyId}/settings`" class="px-3 py-2 rounded-lg text-gray-700 dark:text-gray-300 hover:bg-primary hover:text-white transition-colors">{{ t('nav.settings') }}</router-link>
      </nav>
    </aside>

    <!-- Overlay -->
    <div v-if="showMenu" class="md:hidden fixed inset-0 bg-black/30 z-20" @click="showMenu = false" />

    <!-- Content -->
    <main class="flex-1 p-4 md:p-6 pt-14 md:pt-6">
      <router-view />
    </main>
  </div>
</template>
