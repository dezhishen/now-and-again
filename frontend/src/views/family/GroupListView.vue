<script setup lang="ts">
import { ref, onMounted, inject, watch, type Ref } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { api } from '@/api/client'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import { useConfirm } from '@/composables/useConfirm'
import type { FamilyGroup, FamilyGroupMember } from '@/types'

const { t } = useI18n()
const route = useRoute()
const auth = useAuthStore()
const familyId = route.params.familyId as string

// Reload data when this tab becomes active
const refreshKey = inject<Ref<string>>('refreshKey', ref(''))
watch(refreshKey, (newVal) => { if (newVal === 'groups') loadGroups() })

const groups = ref<FamilyGroup[]>([])
const pageLoading = ref(true)
const showCreate = ref(false)
const newName = ref('')
const newDesc = ref('')
const error = ref('')
const loading = ref<Record<string, boolean>>({})

interface GroupDetail {
  members: FamilyGroupMember[]
  requests: FamilyGroupMember[]
}
const memberCache = ref<Record<string, FamilyGroupMember[]>>({})

// Member management modal
const manageGroupId = ref('')
const manageTab = ref<'members' | 'requests'>('members')
const manageDetail = ref<GroupDetail | null>(null)
const manageLoading = ref(false)
const isFamilyAdmin = ref(false)

onMounted(async () => { pageLoading.value = true; await loadGroups(); pageLoading.value = false })

async function loadGroups() {
  try { groups.value = await api.get<FamilyGroup[]>('/families/' + familyId + '/groups') } catch { groups.value = [] }
  for (const g of groups.value) {
    try {
      memberCache.value[g.id] = await api.get<FamilyGroupMember[]>('/groups/' + g.id + '/members')
    } catch { memberCache.value[g.id] = [] }
  }
  // Check if current user is family owner/admin
  try {
    const members = await api.get<{ user_id: string; role: string }[]>('/families/' + familyId + '/members')
    const me = members.find(m => m.user_id === auth.user?.id)
    isFamilyAdmin.value = me?.role === 'owner' || me?.role === 'admin'
  } catch { /* */ }
}

async function createGroup() {
  error.value = ''
  try {
    await api.post('/families/' + familyId + '/groups', { name: newName.value, description: newDesc.value })
    newName.value = ''; newDesc.value = ''; showCreate.value = false
    await loadGroups()
  } catch (e: any) { error.value = e.message }
}

async function openManage(groupId: string) {
  manageGroupId.value = groupId
  manageTab.value = 'members'
  manageLoading.value = true
  try {
    const [members, requests] = await Promise.all([
      api.get<FamilyGroupMember[]>('/groups/' + groupId + '/members'),
      api.get<FamilyGroupMember[]>('/groups/' + groupId + '/join-requests').catch(() => [] as FamilyGroupMember[]),
    ])
    manageDetail.value = { members, requests }
  } catch {
    manageDetail.value = { members: [], requests: [] }
  }
  finally { manageLoading.value = false }
}

function isGroupOwner(): boolean {
  if (isFamilyAdmin.value) return true
  return manageDetail.value?.members.some(m => m.user_id === auth.user?.id && m.role === 'owner') ?? false
}

function isMember(groupId: string): boolean {
  return memberCache.value[groupId]?.some(m => m.user_id === auth.user?.id) ?? false
}

function isGroupOwnerOf(groupId: string): boolean {
  return memberCache.value[groupId]?.some(m => m.user_id === auth.user?.id && m.role === 'owner') ?? false
}

function isGroupAdmin(groupId: string): boolean {
  if (isFamilyAdmin.value) return true
  return memberCache.value[groupId]?.some(m => m.user_id === auth.user?.id && m.role === 'owner') ?? false
}

async function joinGroup(groupId: string) {
  loading.value[groupId] = true
  try {
    await api.post('/groups/' + groupId + '/join')
    // Refresh member cache so buttons update
    memberCache.value[groupId] = await api.get<FamilyGroupMember[]>('/groups/' + groupId + '/members')
  } catch (e: any) { error.value = e.message }
  finally { loading.value[groupId] = false }
}

async function leaveGroup(groupId: string) {
  loading.value[groupId] = true
  try {
    await api.post('/groups/' + groupId + '/leave')
    memberCache.value[groupId] = await api.get<FamilyGroupMember[]>('/groups/' + groupId + '/members')
  } catch (e: any) { error.value = e.message }
  finally { loading.value[groupId] = false }
}

async function reviewRequest(userId: string, action: 'active' | 'rejected') {
  loading.value[userId] = true
  try {
    await api.put('/groups/' + manageGroupId.value + '/join-requests', { user_id: userId, action })
    await openManage(manageGroupId.value)
  } catch (e: any) { error.value = e.message }
  finally { loading.value[userId] = false }
}

async function removeMember(userId: string) {
  if (!await useConfirm(t('groups.removeConfirm'))) return
  loading.value[userId] = true
  try {
    await api.delete('/groups/' + manageGroupId.value + '/members/' + userId)
    await openManage(manageGroupId.value)
  } catch (e: any) { error.value = e.message }
  finally { loading.value[userId] = false }
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <button class="btn-primary text-sm" @click="showCreate = !showCreate">
        {{ showCreate ? t('groups.cancel') : '+ ' + t('groups.create') }}
      </button>
    </div>

    <p v-if="error" class="text-danger text-sm mb-3">{{ error }}</p>

    <LoadingSpinner v-if="pageLoading" />
    <template v-else>

    <!-- Create form -->
    <div v-if="showCreate" class="card mb-4 flex flex-col gap-2">
      <input v-model="newName" class="input" :placeholder="t('groups.name')" @keyup.enter="createGroup" />
      <input v-model="newDesc" class="input" :placeholder="t('groups.desc')" />
      <button class="btn-primary self-end" @click="createGroup">{{ t('groups.createBtn') }}</button>
    </div>

    <div v-if="groups.length === 0" class="text-center text-gray-400 py-8">{{ t('groups.empty') }}</div>

    <!-- Group cards grid -->
    <div v-else class="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-3 items-start">
      <div v-for="g in groups" :key="g.id"
        class="card flex flex-col gap-1.5 hover:shadow-md transition-shadow h-full"
      >
        <!-- Header -->
        <div class="min-w-0">
          <p class="font-medium dark:text-gray-200 text-sm leading-snug truncate">{{ g.name }}</p>
          <p v-if="g.description" class="text-xs text-gray-400 truncate mt-0.5">{{ g.description }}</p>
        </div>

        <!-- Member avatars -->
        <div class="flex items-center -space-x-2">
          <template v-if="memberCache[g.id]?.length">
            <div v-for="(m, i) in memberCache[g.id].slice(0, 3)" :key="m.id"
              class="w-7 h-7 rounded-full bg-primary/20 flex items-center justify-center text-primary font-bold text-[10px] ring-2 ring-white dark:ring-gray-800 flex-shrink-0"
              :style="{ zIndex: 3 - i }"
            >
              {{ (m.user?.display_name || m.user_id)[0]?.toUpperCase() }}
            </div>
            <div v-if="memberCache[g.id].length > 3"
              class="w-7 h-7 rounded-full bg-gray-200 dark:bg-gray-600 flex items-center justify-center text-gray-400 dark:text-gray-300 font-bold text-[10px] ring-2 ring-white dark:ring-gray-800 flex-shrink-0"
            >+{{ memberCache[g.id].length - 3 }}</div>
          </template>
          <span v-else class="text-xs text-gray-400">-</span>
        </div>

        <!-- Actions -->
        <div class="flex gap-1.5 pt-1.5 border-t border-gray-100 dark:border-gray-700 mt-auto">
          <button v-if="!isMember(g.id)" class="flex-1 text-xs py-1.5 rounded-lg bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-300 hover:bg-blue-100 dark:hover:bg-blue-900/40 transition-colors font-medium" :disabled="loading[g.id]" @click="joinGroup(g.id)">{{ t('groups.join') }}</button>
          <button v-if="isMember(g.id) && !isGroupOwnerOf(g.id)" class="flex-1 text-xs py-1.5 rounded-lg bg-gray-50 dark:bg-gray-700/50 text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors font-medium" :disabled="loading[g.id]" @click="leaveGroup(g.id)">{{ t('groups.leave') }}</button>
          <button v-if="isGroupAdmin(g.id)" class="flex-1 text-xs py-1.5 rounded-lg bg-gray-50 dark:bg-gray-700/50 text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors font-medium" :disabled="loading[g.id]" @click="openManage(g.id)">{{ t('groups.manage') }}</button>
        </div>
      </div>
    </div>
    </template>

    <!-- Member Management Modal -->
    <Teleport to="body">
      <div v-if="manageDetail" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60" @mousedown.self="manageDetail = null">
        <div class="bg-white dark:bg-gray-800 rounded-xl shadow-2xl w-[90vw] max-w-lg max-h-[80vh] flex flex-col">
          <div class="flex items-center justify-between px-4 py-3 border-b dark:border-gray-700">
            <h3 class="font-bold dark:text-gray-200">成员管理</h3>
            <button class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 text-lg" @click="manageDetail = null">✕</button>
          </div>

          <LoadingSpinner v-if="manageLoading" />

          <template v-else>
            <!-- Tabs -->
            <div class="flex gap-1 px-4 pt-3 border-b dark:border-gray-700">
              <button
                class="px-3 py-1.5 text-xs font-medium border-b-2 transition-colors"
                :class="manageTab === 'members' ? 'border-primary text-primary' : 'border-transparent text-gray-400'"
                @click="manageTab = 'members'"
              >成员 ({{ manageDetail.members.length }})</button>
              <button
                v-if="isGroupOwner()"
                class="px-3 py-1.5 text-xs font-medium border-b-2 transition-colors relative"
                :class="manageTab === 'requests' ? 'border-primary text-primary' : 'border-transparent text-gray-400'"
                @click="manageTab = 'requests'"
              >
                待审核
                <span v-if="manageDetail.requests.length" class="ml-1 px-1 py-0.5 rounded-full bg-danger text-white text-[10px]">{{ manageDetail.requests.length }}</span>
              </button>
            </div>

            <!-- Members list -->
            <div class="flex-1 overflow-auto p-4">
              <div v-if="manageTab === 'members'">
                <div v-if="manageDetail.members.length === 0" class="text-sm text-gray-400 py-4 text-center">暂无成员</div>
                <div v-for="m in manageDetail.members" :key="m.id" class="flex items-center justify-between py-2">
                  <div class="flex items-center gap-2 min-w-0">
                    <div class="w-7 h-7 rounded-full bg-primary/20 flex items-center justify-center text-primary font-bold text-xs flex-shrink-0">
                      {{ (m.user?.display_name || m.user_id)[0]?.toUpperCase() }}
                    </div>
                    <span class="text-sm dark:text-gray-300 truncate">{{ m.user?.display_name || m.user_id }}</span>
                    <span class="text-xs px-1.5 py-0.5 rounded-full"
                      :class="m.role === 'owner' ? 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300' : 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400'"
                    >{{ m.role === 'owner' ? '组长' : '成员' }}</span>
                  </div>
                  <button
                    v-if="isGroupOwner() && m.role !== 'owner'"
                    :disabled="loading[m.user_id]"
                    class="text-xs px-2 py-0.5 rounded text-danger hover:bg-red-50 dark:hover:bg-red-900/30 transition-colors"
                    @click="removeMember(m.user_id)"
                  >移除</button>
                </div>
              </div>

              <!-- Requests list -->
              <div v-if="manageTab === 'requests'">
                <div v-if="manageDetail.requests.length === 0" class="text-sm text-gray-400 py-4 text-center">没有待审核的申请</div>
                <div v-for="r in manageDetail.requests" :key="r.id" class="flex items-center justify-between py-2">
                  <div class="flex items-center gap-2 min-w-0">
                    <div class="w-7 h-7 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center text-gray-500 font-bold text-xs flex-shrink-0">
                      {{ (r.user?.display_name || r.user_id)[0]?.toUpperCase() }}
                    </div>
                    <span class="text-sm dark:text-gray-300 truncate">{{ r.user?.display_name || r.user_id }}</span>
                  </div>
                  <div class="flex gap-1.5 flex-shrink-0">
                    <button :disabled="loading[r.user_id]" class="text-xs px-2 py-1 rounded bg-green-100 dark:bg-green-900 text-green-700 dark:text-green-300 hover:opacity-80" @click="reviewRequest(r.user_id, 'active')">通过</button>
                    <button :disabled="loading[r.user_id]" class="text-xs px-2 py-1 rounded bg-red-100 dark:bg-red-900 text-red-700 dark:text-red-300 hover:opacity-80" @click="reviewRequest(r.user_id, 'rejected')">拒绝</button>
                  </div>
                </div>
              </div>
            </div>
          </template>
        </div>
      </div>
    </Teleport>
  </div>
</template>
