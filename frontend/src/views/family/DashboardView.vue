<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { api } from '@/api/client'
import type { Family } from '@/types'

const { t } = useI18n()
const route = useRoute()
const familyId = route.params.familyId as string

const family = ref<Family | null>(null)
const memberCount = ref(0)
const groupCount = ref(0)
const copied = ref(false)

async function copyInviteCode() {
  if (!family.value?.invite_code) return
  try {
    await navigator.clipboard.writeText(family.value.invite_code)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch { /* */ }
}

onMounted(async () => {
  try {
    family.value = await api.get<Family>('/families/' + familyId)
  } catch { /* */ }
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

    <!-- Stats -->
    <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-6">
      <div class="card"><p class="text-gray-400 text-sm">{{ t('dashboard.members') }}</p><p class="text-2xl font-bold dark:text-gray-200">{{ memberCount }}</p></div>
      <div class="card"><p class="text-gray-400 text-sm">{{ t('dashboard.groups') }}</p><p class="text-2xl font-bold dark:text-gray-200">{{ groupCount }}</p></div>
    </div>

    <!-- Invite Code -->
    <div v-if="family?.invite_code" class="card">
      <p class="text-gray-400 text-sm mb-1">{{ t('dashboard.inviteCode') }}</p>
      <p class="text-gray-500 text-xs mb-3">{{ t('dashboard.inviteCodeHint') }}</p>
      <div class="flex items-center gap-3">
        <code class="flex-1 bg-gray-100 dark:bg-gray-700 px-4 py-2 rounded-lg text-lg font-mono tracking-wider text-center select-all dark:text-gray-200">{{ family.invite_code }}</code>
        <button
          class="px-4 py-2 rounded-lg bg-primary text-white text-sm font-medium hover:opacity-90 transition-opacity whitespace-nowrap"
          @click="copyInviteCode"
        >
          {{ copied ? t('dashboard.copied') : t('dashboard.copy') }}
        </button>
      </div>
    </div>
  </div>
</template>
