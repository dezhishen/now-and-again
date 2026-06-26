<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { api } from '@/api/client'
import type { FamilyGroup, FamilyGroupMember } from '@/types'

const { t } = useI18n()
const route = useRoute()
const familyId = route.params.familyId as string

const groups = ref<FamilyGroup[]>([])
const showCreate = ref(false)
const newName = ref('')
const newDesc = ref('')
const error = ref('')
const expandedId = ref<string | null>(null)
const members = ref<Record<string, FamilyGroupMember[]>>({})

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

async function toggleMembers(groupId: string) {
  if (expandedId.value === groupId) { expandedId.value = null; return }
  expandedId.value = groupId
  try {
    members.value[groupId] = await api.get<FamilyGroupMember[]>('/groups/' + groupId + '/members')
  } catch { members.value[groupId] = [] }
}

async function joinGroup(groupId: string) {
  try { await api.post('/groups/' + groupId + '/join'); await toggleMembers(groupId) } catch (e: any) { error.value = e.message }
}

async function leaveGroup(groupId: string) {
  try { await api.post('/groups/' + groupId + '/leave'); await toggleMembers(groupId) } catch (e: any) { error.value = e.message }
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-xl md:text-2xl font-bold dark:text-gray-200">{{ t('nav.groups') }}</h2>
      <button class="btn-primary text-sm" @click="showCreate = !showCreate">{{ showCreate ? '取消' : '+ 新建小组' }}</button>
    </div>

    <div v-if="showCreate" class="card mb-4 flex flex-col gap-2">
      <input v-model="newName" class="input" placeholder="小组名称" @keyup.enter="createGroup" />
      <input v-model="newDesc" class="input" placeholder="描述（可选）" />
      <button class="btn-primary self-end" @click="createGroup">创建</button>
      <p v-if="error" class="text-danger text-sm">{{ error }}</p>
    </div>

    <div v-if="groups.length === 0" class="text-center text-gray-400 py-8">暂无小组</div>

    <div v-for="g in groups" :key="g.id" class="card mb-2">
      <div class="flex items-center justify-between cursor-pointer" @click="toggleMembers(g.id)">
        <div>
          <span class="font-medium dark:text-gray-200">{{ g.name }}</span>
          <span class="text-xs text-gray-400 ml-2">{{ g.description }}</span>
        </div>
        <span class="text-xs">{{ expandedId === g.id ? '▲' : '▼' }}</span>
      </div>

      <div v-if="expandedId === g.id" class="mt-3 border-t dark:border-gray-700 pt-3">
        <div class="flex gap-2 mb-2">
          <button class="text-xs px-2 py-1 rounded bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-300" @click="joinGroup(g.id)">加入</button>
          <button class="text-xs px-2 py-1 rounded bg-red-100 dark:bg-red-900 text-red-700 dark:text-red-300" @click="leaveGroup(g.id)">离开</button>
        </div>
        <div v-if="members[g.id]?.length">
          <div v-for="m in members[g.id]" :key="m.id" class="flex items-center justify-between py-1 text-sm">
            <span class="dark:text-gray-300">{{ m.user?.display_name || m.user_id }}</span>
            <span class="text-xs text-gray-400">{{ m.role === 'owner' ? '组长' : '成员' }} · {{ m.status === 'active' ? '已加入' : m.status === 'pending' ? '待审核' : '已拒绝' }}</span>
          </div>
        </div>
        <div v-else class="text-sm text-gray-400">暂无成员</div>
      </div>
    </div>
  </div>
</template>
