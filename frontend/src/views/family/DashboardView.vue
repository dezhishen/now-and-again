<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { api } from '@/api/client'

const { t } = useI18n()
const route = useRoute()
const familyId = route.params.familyId as string

const memberCount = ref(0)
const groupCount = ref(0)

onMounted(async () => {
  try {
    const members = await api.get<any[]>('/families/' + familyId + '/members')
    memberCount.value = members.length
  } catch { /* */ }
  try {
    const groups = await api.get<any[]>('/families/' + familyId + '/groups')
    groupCount.value = groups.length
  } catch { /* */ }
})
</script>

<template>
  <div>
    <h2 class="text-xl md:text-2xl font-bold mb-4 dark:text-gray-200">{{ t('dashboard.heading') }}</h2>
    <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-6">
      <div class="card"><p class="text-gray-400 text-sm">{{ t('dashboard.members') }}</p><p class="text-2xl font-bold dark:text-gray-200">{{ memberCount }}</p></div>
      <div class="card"><p class="text-gray-400 text-sm">{{ t('dashboard.groups') }}</p><p class="text-2xl font-bold dark:text-gray-200">{{ groupCount }}</p></div>
    </div>
  </div>
</template>
