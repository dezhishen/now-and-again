<script setup lang="ts">
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
</script>

<template>
  <div class="flex min-h-screen">
    <aside class="w-[200px] bg-white border-r border-gray-200 p-4">
      <h3 class="text-lg font-semibold mb-3">{{ t('nav.family') }}</h3>
      <nav class="flex flex-col gap-1">
        <router-link :to="`/family/${$route.params.familyId}`" class="px-3 py-2 rounded-lg text-gray-700 hover:bg-primary hover:text-white transition-colors">
          {{ t('nav.dashboard') }}
        </router-link>
        <router-link :to="`/family/${$route.params.familyId}/tasks`" class="px-3 py-2 rounded-lg text-gray-700 hover:bg-primary hover:text-white transition-colors">
          {{ t('nav.tasks') }}
        </router-link>
        <router-link :to="`/family/${$route.params.familyId}/chains`" class="px-3 py-2 rounded-lg text-gray-700 hover:bg-primary hover:text-white transition-colors">
          {{ t('nav.chains') }}
        </router-link>
        <router-link :to="`/family/${$route.params.familyId}/inspections`" class="px-3 py-2 rounded-lg text-gray-700 hover:bg-primary hover:text-white transition-colors">
          {{ t('nav.inspections') }}
        </router-link>
        <router-link :to="`/family/${$route.params.familyId}/members`" class="px-3 py-2 rounded-lg text-gray-700 hover:bg-primary hover:text-white transition-colors">
          {{ t('nav.members') }}
        </router-link>
        <router-link :to="`/family/${$route.params.familyId}/settings`" class="px-3 py-2 rounded-lg text-gray-700 hover:bg-primary hover:text-white transition-colors">
          {{ t('nav.settings') }}
        </router-link>
      </nav>
    </aside>
    <main class="flex-1 p-6">
      <router-view />
    </main>
  </div>
</template>
