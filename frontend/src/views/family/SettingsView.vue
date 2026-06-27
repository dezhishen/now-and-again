<script setup lang="ts">
import { ref, onMounted, inject, watch, type Ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { api } from '@/api/client'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import { useLoading } from '@/composables/useLoading'
import { useConfirm } from '@/composables/useConfirm'
import type { Family } from '@/types'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const familyId = route.params.familyId as string

// Reload data when this tab becomes active
const refreshKey = inject<Ref<string>>('refreshKey', ref(''))
watch(refreshKey, (newVal) => { if (newVal === 'settings') withLoading(loadFamily) })

const family = ref<Family | null>(null)
const { loading, withLoading } = useLoading()
const editName = ref('')
const saving = ref(false)
const saved = ref(false)
const error = ref('')

async function loadFamily() {
  try {
    family.value = await api.get<Family>('/families/' + familyId)
    editName.value = family.value.name
  } catch { /* */ }
}

onMounted(async () => {
  loading.value = true
  await loadFamily()
  loading.value = false
})

async function saveName() {
  saving.value = true
  error.value = ''
  saved.value = false
  try {
    await api.patch('/families/' + familyId, { name: editName.value })
    if (family.value) family.value.name = editName.value
    saved.value = true
    setTimeout(() => { saved.value = false }, 2000)
  } catch (e: any) {
    error.value = e.message
    editName.value = family.value?.name || ''
  } finally { saving.value = false }
}

async function deleteFamily() {
  if (!await useConfirm(t('settingsPage.deleteConfirm'))) return
  try {
    await api.delete('/families/' + familyId)
    router.push('/')
  } catch (e: any) { error.value = e.message }
}

const copied = ref(false)
async function copyInviteCode() {
  if (!family.value?.invite_code) return
  try {
    await navigator.clipboard.writeText(family.value.invite_code)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch { /* */ }
}
</script>

<template>
  <div class="max-w-xl">

    <p v-if="error" class="text-danger text-sm mb-4">{{ error }}</p>
    <LoadingSpinner v-if="loading" />
    <template v-else>
    <!-- Invite Code -->
    <div class="card mb-4">
      <p class="text-gray-400 text-sm mb-1">{{ t('dashboard.inviteCode') }}</p>
      <div class="flex items-center gap-3">
        <code class="flex-1 bg-gray-100 dark:bg-gray-700 px-3 py-2 rounded-lg font-mono text-center select-all dark:text-gray-200">{{ family?.invite_code }}</code>
        <button
          class="px-3 py-2 rounded-lg bg-primary text-white text-sm hover:opacity-90 transition-opacity whitespace-nowrap"
          @click="copyInviteCode"
        >{{ copied ? t('dashboard.copied') : t('dashboard.copy') }}</button>
      </div>
    </div>

    <!-- Family Name -->
    <div class="card mb-4">
      <label class="text-sm text-gray-400 mb-2 block">{{ t('settingsPage.name') }}</label>
      <div class="flex gap-2">
        <input v-model="editName" class="input flex-1" :placeholder="t('settingsPage.namePlaceholder')" @keyup.enter="saveName" />
        <button :disabled="saving" class="btn-primary text-sm whitespace-nowrap" @click="saveName">
          {{ saved ? t('settingsPage.saved') : t('settingsPage.save') }}
        </button>
      </div>
    </div>

    <!-- Danger Zone -->
    <div class="card border border-red-200 dark:border-red-900">
      <p class="text-sm font-medium text-danger mb-1">{{ t('settingsPage.deleteFamily') }}</p>
      <p class="text-xs text-gray-400 mb-3">{{ t('settingsPage.deleteWarning') }}</p>
      <button class="btn-danger text-sm" @click="deleteFamily">{{ t('settingsPage.deleteFamily') }}</button>
    </div>
    </template>
  </div>
</template>
