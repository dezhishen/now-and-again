<script setup lang="ts">
import { ref, computed, onMounted, inject, watch, type Ref } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { api } from '@/api/client'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import { useConfirm } from '@/composables/useConfirm'
import type { FamilyMember, FamilyRole } from '@/types'

const { t } = useI18n()
const route = useRoute()
const auth = useAuthStore()
const familyId = route.params.familyId as string

// Reload data when this tab becomes active
const refreshKey = inject<Ref<string>>('refreshKey', ref(''))
watch(refreshKey, (newVal) => { if (newVal === 'members') { loadMembers(); loadRequests() } })

const members = ref<FamilyMember[]>([])
const pageLoading = ref(true)
const requests = ref<FamilyMember[]>([])
const activeTab = ref<'members' | 'requests'>('members')
const error = ref('')
const loading = ref<Record<string, boolean>>({})

const myRole = computed(() => {
  const me = members.value.find(m => m.user_id === auth.user?.id)
  return (me?.role || 'member') as FamilyRole
})
const canManage = computed(() => myRole.value === 'owner' || myRole.value === 'admin')

onMounted(async () => {
  pageLoading.value = true
  await Promise.all([loadMembers(), loadRequests()])
  pageLoading.value = false
})

async function loadMembers() {
  try {
    members.value = await api.get<FamilyMember[]>('/families/' + familyId + '/members')
  } catch { members.value = [] }
}

async function loadRequests() {
  try {
    requests.value = await api.get<FamilyMember[]>('/families/' + familyId + '/join-requests')
  } catch { requests.value = [] }
}

async function changeRole(userId: string, role: FamilyRole) {
  loading.value[userId] = true
  error.value = ''
  try {
    await api.put('/families/' + familyId + '/members/' + userId + '/role', { role })
    await loadMembers()
  } catch (e: any) { error.value = e.message }
  finally { loading.value[userId] = false }
}

async function removeMember(userId: string) {
  if (!await useConfirm(t('members.removeConfirm'))) return
  loading.value[userId] = true
  error.value = ''
  try {
    await api.delete('/families/' + familyId + '/members/' + userId)
    await loadMembers()
  } catch (e: any) { error.value = e.message }
  finally { loading.value[userId] = false }
}

async function reviewRequest(userId: string, action: 'active' | 'rejected') {
  loading.value[userId] = true
  error.value = ''
  try {
    await api.put('/families/' + familyId + '/join-requests', { user_id: userId, action })
    await Promise.all([loadMembers(), loadRequests()])
    if (activeTab.value === 'requests' && requests.value.length === 0) {
      activeTab.value = 'members'
    }
  } catch (e: any) { error.value = e.message }
  finally { loading.value[userId] = false }
}

const ROLE_LABELS: Record<string, string> = {
  owner: 'members.role_owner',
  admin: 'members.role_admin',
  member: 'members.role_member',
}
</script>

<template>
  <div>

    <p v-if="error" class="text-danger text-sm mb-3">{{ error }}</p>

    <LoadingSpinner v-if="pageLoading" />
    <template v-else>

    <!-- Tabs -->
    <div class="flex gap-1 mb-4 border-b dark:border-gray-700">
      <button
        class="px-4 py-2 text-sm font-medium border-b-2 transition-colors"
        :class="activeTab === 'members' ? 'border-primary text-primary' : 'border-transparent text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'"
        @click="activeTab = 'members'"
      >
        {{ t('members.tabMembers') }}
      </button>
      <button
        v-if="canManage"
        class="px-4 py-2 text-sm font-medium border-b-2 transition-colors relative"
        :class="activeTab === 'requests' ? 'border-primary text-primary' : 'border-transparent text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'"
        @click="activeTab = 'requests'"
      >
        {{ t('members.tabRequests') }}
        <span v-if="requests.length" class="ml-1.5 px-1.5 py-0.5 rounded-full bg-danger text-white text-xs">{{ requests.length }}</span>
      </button>
    </div>

    <!-- Members Tab -->
    <div v-if="activeTab === 'members'">
      <div v-if="members.length === 0" class="text-center text-gray-400 py-8">{{ t('members.empty') }}</div>
      <div v-for="m in members" :key="m.id" class="card mb-2 flex items-center justify-between gap-3">
        <div class="flex items-center gap-3 min-w-0">
          <div class="w-9 h-9 rounded-full bg-primary/20 flex items-center justify-center text-primary font-bold text-sm flex-shrink-0">
            {{ (m.user?.display_name || m.user_id)[0]?.toUpperCase() }}
          </div>
          <div class="min-w-0">
            <p class="font-medium dark:text-gray-200 truncate">{{ m.user?.display_name || m.user_id }}</p>
            <p class="text-xs text-gray-400 truncate">{{ m.user?.email }}</p>
          </div>
        </div>
        <div class="flex items-center gap-2 flex-shrink-0">
          <!-- Role badge / selector -->
          <template v-if="canManage && m.role !== 'owner'">
            <select
              :value="m.role"
              :disabled="loading[m.user_id]"
              class="text-xs border border-gray-200 dark:border-gray-600 rounded px-2 py-1 bg-white dark:bg-gray-700 dark:text-gray-200"
              @change="changeRole(m.user_id, ($event.target as HTMLSelectElement).value as FamilyRole)"
            >
              <option value="admin">{{ t('members.role_admin') }}</option>
              <option value="member">{{ t('members.role_member') }}</option>
            </select>
          </template>
          <span v-else class="text-xs px-2 py-1 rounded-full" :class="{
            'bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300': m.role === 'owner',
            'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300': m.role === 'admin',
            'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400': m.role === 'member',
          }">{{ t(ROLE_LABELS[m.role]) }}</span>

          <!-- Remove button (not for owner) -->
          <button
            v-if="canManage && m.role !== 'owner' && m.user_id !== auth.user?.id"
            :disabled="loading[m.user_id]"
            class="text-xs px-2 py-1 rounded text-danger hover:bg-red-50 dark:hover:bg-red-900/30 transition-colors"
            @click="removeMember(m.user_id)"
          >{{ t('members.remove') }}</button>
        </div>
      </div>
    </div>

    <!-- Requests Tab -->
    <div v-if="activeTab === 'requests'">
      <div v-if="requests.length === 0" class="text-center text-gray-400 py-8">{{ t('members.emptyRequests') }}</div>
      <div v-for="r in requests" :key="r.id" class="card mb-2 flex items-center justify-between gap-3">
        <div class="flex items-center gap-3 min-w-0">
          <div class="w-9 h-9 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center text-gray-500 font-bold text-sm flex-shrink-0">
            {{ (r.user?.display_name || r.user_id)[0]?.toUpperCase() }}
          </div>
          <div class="min-w-0">
            <p class="font-medium dark:text-gray-200 truncate">{{ r.user?.display_name || r.user_id }}</p>
            <p class="text-xs text-gray-400 truncate">{{ r.user?.email }}</p>
          </div>
        </div>
        <div class="flex items-center gap-2 flex-shrink-0">
          <button
            :disabled="loading[r.user_id]"
            class="text-xs px-3 py-1.5 rounded-lg bg-green-100 dark:bg-green-900 text-green-700 dark:text-green-300 hover:bg-green-200 dark:hover:bg-green-800 transition-colors"
            @click="reviewRequest(r.user_id, 'active')"
          >{{ t('members.approve') }}</button>
          <button
            :disabled="loading[r.user_id]"
            class="text-xs px-3 py-1.5 rounded-lg bg-red-100 dark:bg-red-900 text-red-700 dark:text-red-300 hover:bg-red-200 dark:hover:bg-red-800 transition-colors"
            @click="reviewRequest(r.user_id, 'rejected')"
          >{{ t('members.reject') }}</button>
        </div>
      </div>
    </div>
    </template>
  </div>
</template>
