<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { api } from '@/api/client'
import type { FamilyGroup, FamilyGroupMember } from '@/types'

const { t } = useI18n()
const route = useRoute()
const auth = useAuthStore()
const familyId = route.params.familyId as string

const groups = ref<FamilyGroup[]>([])
const showCreate = ref(false)
const newName = ref('')
const newDesc = ref('')
const error = ref('')
const loading = ref<Record<string, boolean>>({})

// Expanded group: { groupId: { members: [...], requests: [...] } }
interface GroupDetail {
  members: FamilyGroupMember[]
  requests: FamilyGroupMember[]
}
const details = ref<Record<string, GroupDetail>>({})
const activeTab = ref<Record<string, 'members' | 'requests'>>({})

onMounted(async () => { await loadGroups() })

async function loadGroups() {
  try { groups.value = await api.get<FamilyGroup[]>('/families/' + familyId + '/groups') } catch { groups.value = [] }
}

async function createGroup() {
  error.value = ''
  try {
    await api.post('/families/' + familyId + '/groups', { name: newName.value, description: newDesc.value })
    newName.value = ''; newDesc.value = ''; showCreate.value = false
    await loadGroups()
  } catch (e: any) { error.value = e.message }
}

async function toggleGroup(groupId: string) {
  if (details.value[groupId]) {
    delete details.value[groupId]
    delete activeTab.value[groupId]
    return
  }
  activeTab.value[groupId] = 'members'
  await loadDetail(groupId)
}

async function loadDetail(groupId: string) {
  try {
    const [members, requests] = await Promise.all([
      api.get<FamilyGroupMember[]>('/groups/' + groupId + '/members'),
      api.get<FamilyGroupMember[]>('/groups/' + groupId + '/join-requests').catch(() => [] as FamilyGroupMember[]),
    ])
    details.value[groupId] = { members, requests }
  } catch {
    details.value[groupId] = { members: [], requests: [] }
  }
}

async function joinGroup(groupId: string) {
  loading.value[groupId] = true
  error.value = ''
  try {
    await api.post('/groups/' + groupId + '/join')
    await loadDetail(groupId)
  } catch (e: any) { error.value = e.message }
  finally { loading.value[groupId] = false }
}

async function leaveGroup(groupId: string) {
  loading.value[groupId] = true
  error.value = ''
  try {
    await api.post('/groups/' + groupId + '/leave')
    await loadDetail(groupId)
  } catch (e: any) { error.value = e.message }
  finally { loading.value[groupId] = false }
}

function isGroupOwner(detail: GroupDetail): boolean {
  return detail.members.some(m => m.user_id === auth.user?.id && m.role === 'owner')
}

async function reviewRequest(groupId: string, userId: string, action: 'active' | 'rejected') {
  loading.value[userId] = true
  error.value = ''
  try {
    await api.put('/groups/' + groupId + '/join-requests', { user_id: userId, action })
    await loadDetail(groupId)
  } catch (e: any) { error.value = e.message }
  finally { loading.value[userId] = false }
}

async function removeMember(groupId: string, userId: string) {
  if (!confirm(t('groups.removeConfirm'))) return
  loading.value[userId] = true
  error.value = ''
  try {
    await api.delete('/groups/' + groupId + '/members/' + userId)
    await loadDetail(groupId)
  } catch (e: any) { error.value = e.message }
  finally { loading.value[userId] = false }
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-xl md:text-2xl font-bold dark:text-gray-200">{{ t('groups.heading') }}</h2>
      <button class="btn-primary text-sm" @click="showCreate = !showCreate">
        {{ showCreate ? t('groups.cancel') : '+ ' + t('groups.create') }}
      </button>
    </div>

    <p v-if="error" class="text-danger text-sm mb-3">{{ error }}</p>

    <!-- Create form -->
    <div v-if="showCreate" class="card mb-4 flex flex-col gap-2">
      <input v-model="newName" class="input" :placeholder="t('groups.name')" @keyup.enter="createGroup" />
      <input v-model="newDesc" class="input" :placeholder="t('groups.desc')" />
      <button class="btn-primary self-end" @click="createGroup">{{ t('groups.createBtn') }}</button>
    </div>

    <div v-if="groups.length === 0" class="text-center text-gray-400 py-8">{{ t('groups.empty') }}</div>

    <!-- Group cards -->
    <div v-for="g in groups" :key="g.id" class="card mb-3">
      <!-- Header -->
      <div class="flex items-center justify-between cursor-pointer select-none" @click="toggleGroup(g.id)">
        <div>
          <span class="font-medium dark:text-gray-200">{{ g.name }}</span>
          <span v-if="g.description" class="text-xs text-gray-400 ml-2">{{ g.description }}</span>
        </div>
        <span class="text-xs text-gray-400">{{ details[g.id] ? '▲' : '▼' }}</span>
      </div>

      <!-- Expanded detail -->
      <div v-if="details[g.id]" class="mt-3 border-t dark:border-gray-700 pt-3">
        <!-- Action buttons -->
        <div class="flex gap-2 mb-3">
          <button class="text-xs px-2 py-1 rounded bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-300 hover:opacity-80" :disabled="loading[g.id]" @click="joinGroup(g.id)">{{ t('groups.join') }}</button>
          <button class="text-xs px-2 py-1 rounded bg-red-100 dark:bg-red-900 text-red-700 dark:text-red-300 hover:opacity-80" :disabled="loading[g.id]" @click="leaveGroup(g.id)">{{ t('groups.leave') }}</button>
        </div>

        <!-- Tabs -->
        <div class="flex gap-1 mb-3 border-b dark:border-gray-700">
          <button
            class="px-3 py-1.5 text-xs font-medium border-b-2 transition-colors"
            :class="(activeTab[g.id] || 'members') === 'members' ? 'border-primary text-primary' : 'border-transparent text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'"
            @click="activeTab[g.id] = 'members'"
          >{{ t('groups.role_member') }} ({{ details[g.id].members.length }})</button>
          <button
            v-if="isGroupOwner(details[g.id])"
            class="px-3 py-1.5 text-xs font-medium border-b-2 transition-colors relative"
            :class="activeTab[g.id] === 'requests' ? 'border-primary text-primary' : 'border-transparent text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'"
            @click="activeTab[g.id] = 'requests'"
          >
            {{ t('groups.pendingTab') }}
            <span v-if="details[g.id].requests.length" class="ml-1 px-1 py-0.5 rounded-full bg-danger text-white text-[10px]">{{ details[g.id].requests.length }}</span>
          </button>
        </div>

        <!-- Members tab -->
        <div v-if="(activeTab[g.id] || 'members') === 'members'">
          <div v-if="details[g.id].members.length === 0" class="text-sm text-gray-400 py-4 text-center">{{ t('groups.noMembers') }}</div>
          <div v-for="m in details[g.id].members" :key="m.id" class="flex items-center justify-between py-2">
            <div class="flex items-center gap-2 min-w-0">
              <div class="w-7 h-7 rounded-full bg-primary/20 flex items-center justify-center text-primary font-bold text-xs flex-shrink-0">
                {{ (m.user?.display_name || m.user_id)[0]?.toUpperCase() }}
              </div>
              <span class="text-sm dark:text-gray-300 truncate">{{ m.user?.display_name || m.user_id }}</span>
              <span class="text-xs px-1.5 py-0.5 rounded-full"
                :class="m.role === 'owner' ? 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300' : 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400'"
              >{{ t(m.role === 'owner' ? 'groups.role_owner' : 'groups.role_member') }}</span>
            </div>
            <button
              v-if="isGroupOwner(details[g.id]) && m.role !== 'owner'"
              :disabled="loading[m.user_id]"
              class="text-xs px-2 py-0.5 rounded text-danger hover:bg-red-50 dark:hover:bg-red-900/30 transition-colors"
              @click="removeMember(g.id, m.user_id)"
            >{{ t('groups.remove') }}</button>
          </div>
        </div>

        <!-- Requests tab -->
        <div v-if="activeTab[g.id] === 'requests'">
          <div v-if="details[g.id].requests.length === 0" class="text-sm text-gray-400 py-4 text-center">{{ t('groups.noPending') }}</div>
          <div v-for="r in details[g.id].requests" :key="r.id" class="flex items-center justify-between py-2">
            <div class="flex items-center gap-2 min-w-0">
              <div class="w-7 h-7 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center text-gray-500 font-bold text-xs flex-shrink-0">
                {{ (r.user?.display_name || r.user_id)[0]?.toUpperCase() }}
              </div>
              <span class="text-sm dark:text-gray-300 truncate">{{ r.user?.display_name || r.user_id }}</span>
            </div>
            <div class="flex gap-1.5 flex-shrink-0">
              <button
                :disabled="loading[r.user_id]"
                class="text-xs px-2 py-1 rounded bg-green-100 dark:bg-green-900 text-green-700 dark:text-green-300 hover:opacity-80"
                @click="reviewRequest(g.id, r.user_id, 'active')"
              >{{ t('groups.approve') }}</button>
              <button
                :disabled="loading[r.user_id]"
                class="text-xs px-2 py-1 rounded bg-red-100 dark:bg-red-900 text-red-700 dark:text-red-300 hover:opacity-80"
                @click="reviewRequest(g.id, r.user_id, 'rejected')"
              >{{ t('groups.reject') }}</button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
