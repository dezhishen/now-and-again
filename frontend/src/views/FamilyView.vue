<script setup lang="ts">
import { ref, onMounted, markRaw } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { api } from '@/api/client'

import DashboardView from './family/DashboardView.vue'
import GroupListView from './family/GroupListView.vue'
import MemberListView from './family/MemberListView.vue'
import FloorPlanView from './family/FloorPlanView.vue'
import TaskView from './family/TaskView.vue'
import IcsView from './family/IcsView.vue'
import SettingsView from './family/SettingsView.vue'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const showMenu = ref(false)
const familyName = ref('')
const isFamilyAdmin = ref(false)

interface Tab {
  id: string
  label: string
  icon: string
  component: any
}

const NAV_ITEMS: { id: string; icon: string; labelKey: string; component: any; adminOnly?: boolean }[] = [
  { id: 'dashboard', icon: '📊', labelKey: 'nav.dashboard', component: markRaw(DashboardView) },
  { id: 'groups', icon: '👥', labelKey: 'nav.groups', component: markRaw(GroupListView) },
  { id: 'members', icon: '👤', labelKey: 'nav.members', component: markRaw(MemberListView) },
  { id: 'floor-plan', icon: '🏠', labelKey: 'nav.floorPlan', component: markRaw(FloorPlanView) },
  { id: 'tasks', icon: '✅', labelKey: 'nav.tasks', component: markRaw(TaskView) },
  { id: 'ics', icon: '📅', labelKey: 'nav.ics', component: markRaw(IcsView) },
  { id: 'settings', icon: '⚙️', labelKey: 'nav.settings', component: markRaw(SettingsView), adminOnly: true },
]

const tabs = ref<Tab[]>([])
const activeTabId = ref('')

function findNav(id: string) {
  return NAV_ITEMS.find(n => n.id === id)
}

function openTab(id: string) {
  const nav = findNav(id)
  if (!nav) return

  if (id === 'calendar') {
    window.open(`/calendar/${route.params.familyId}`, '_blank')
    return
  }

  // Activate existing tab or create new one
  const existing = tabs.value.find(t => t.id === id)
  if (existing) {
    activeTabId.value = id
  } else {
    tabs.value.push({ id: nav.id, label: t(nav.labelKey), icon: nav.icon, component: nav.component })
    activeTabId.value = id
  }

  // Update URL without navigation
  router.replace({ name: `family-${id}` })
}

function closeTab(id: string) {
  const idx = tabs.value.findIndex(t => t.id === id)
  if (idx === -1) return
  tabs.value.splice(idx, 1)
  if (activeTabId.value === id) {
    activeTabId.value = tabs.value[Math.min(idx, tabs.value.length - 1)]?.id || ''
  }
  if (!activeTabId.value) {
    // Default to dashboard
    openTab('dashboard')
  }
}

onMounted(async () => {
  try {
    const f = await api.get<{ name: string }>('/families/' + route.params.familyId)
    familyName.value = f.name
  } catch { /* */ }
  try {
    const members = await api.get<{ user_id: string; role: string }[]>('/families/' + route.params.familyId + '/members')
    const me = members.find(m => m.user_id === auth.user?.id)
    isFamilyAdmin.value = me?.role === 'owner' || me?.role === 'admin'
  } catch { /* */ }

  // Open tab based on current route
  const routeName = (route.name as string) || ''
  const tabId = routeName.replace('family-', '') || 'dashboard'
  openTab(tabId)
})

async function leaveFamily() {
  if (!confirm('确定要离开这个家庭吗？')) return
  try {
    await api.post('/families/' + route.params.familyId + '/leave')
    window.location.href = '/'
  } catch (e: any) { alert(e.message) }
}
</script>

<template>
  <div class="flex flex-col md:flex-row min-h-screen">
    <!-- Mobile hamburger -->
    <button class="md:hidden fixed top-2 left-3 z-40 w-8 h-8 rounded-lg flex items-center justify-center bg-gray-200 dark:bg-gray-700 shadow text-sm" @click="showMenu = !showMenu">
      {{ showMenu ? '✕' : '☰' }}
    </button>

    <!-- Sidebar -->
    <aside
      class="w-[200px] bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 p-4 flex-shrink-0 transition-transform"
      :class="showMenu ? 'fixed inset-y-0 left-0 z-30 translate-x-0' : 'max-md:fixed max-md:inset-y-0 max-md:left-0 max-md:z-30 max-md:-translate-x-full'"
      @click="showMenu = false"
    >
      <nav class="flex flex-col gap-1">
        <button
          v-for="nav in NAV_ITEMS.filter(n => !n.adminOnly || isFamilyAdmin)"
          :key="nav.id"
          class="px-3 py-2 rounded-lg text-left text-gray-700 dark:text-gray-300 hover:bg-primary hover:text-white transition-colors text-sm"
          :class="activeTabId === nav.id ? 'bg-primary/10 text-primary font-medium' : ''"
          @click="openTab(nav.id)"
        >{{ nav.icon }} {{ t(nav.labelKey) }}</button>
        <button class="px-3 py-2 rounded-lg text-left text-gray-700 dark:text-gray-300 hover:bg-primary hover:text-white transition-colors text-sm" @click="openTab('calendar')">🖥️ {{ t('nav.calendar') }}</button>
        <hr class="my-2 border-gray-200 dark:border-gray-700" />
        <button class="px-3 py-2 rounded-lg text-left text-gray-400 hover:text-danger hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors text-sm" @click="leaveFamily">🚪 离开家庭</button>
      </nav>
    </aside>

    <!-- Overlay -->
    <div v-if="showMenu" class="md:hidden fixed inset-0 bg-black/30 z-20" @click="showMenu = false" />

    <!-- Content -->
    <main class="flex-1 flex flex-col pt-14 md:pt-0 min-w-0">
      <!-- Tab bar -->
      <div class="flex items-center border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50 px-2 overflow-x-auto flex-shrink-0">
        <div class="flex items-center text-sm font-medium text-gray-600 dark:text-gray-300 px-3 py-2 flex-shrink-0 mr-2">
          {{ familyName || t('nav.family') }}
        </div>
        <button
          v-for="tab in tabs" :key="tab.id"
          class="group flex items-center gap-1 px-3 py-2 text-sm border-b-2 transition-colors flex-shrink-0 max-w-[160px]"
          :class="activeTabId === tab.id ? 'border-primary text-primary' : 'border-transparent text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'"
          @click="activeTabId = tab.id"
        >
          <span class="truncate">{{ tab.icon }} {{ tab.label }}</span>
          <span class="ml-1 w-4 h-4 rounded-full flex items-center justify-center text-[10px] opacity-0 group-hover:opacity-100 hover:bg-gray-200 dark:hover:bg-gray-600 transition-opacity flex-shrink-0" @click.stop="closeTab(tab.id)">✕</span>
        </button>
      </div>

      <!-- Tab content -->
      <div class="flex-1 p-4 md:p-6 overflow-auto">
        <component
          v-for="tab in tabs" :key="tab.id"
          :is="tab.component"
          v-show="activeTabId === tab.id"
        />
      </div>
    </main>
  </div>
</template>
