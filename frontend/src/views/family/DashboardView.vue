<script setup lang="ts">
import { ref, computed, watch, onMounted, inject, type Ref } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { api } from '@/api/client'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import { useToast } from '@/composables/useToast'
import { getTodoActions, getTodoInfo, getTodoBadge } from '@/composables/useTaskKinds'
import { initTaskKinds } from '@/components/tasks/init'
import type { Family, Todo } from '@/types'

initTaskKinds()

const { t } = useI18n()
const toast = useToast()
const route = useRoute()
const familyId = route.params.familyId as string

// Reload data when this tab becomes active (switching tabs)
const refreshKey = inject<Ref<number>>('refreshKey', ref(0))
watch(refreshKey, () => { loadTodos(); loadLocations() })

const activeTab = ref<'todos' | 'overview'>('todos')
const loading = ref(true)
const family = ref<Family | null>(null)
const memberCount = ref(0)
const groupCount = ref(0)
const todos = ref<Todo[]>([])
const locations = ref<{ id: string; name: string }[]>([])
const showAll = ref(false)
const copied = ref(false)
const PAGE_SIZE = 5

const displayTodos = computed(() => showAll.value ? todos.value : todos.value.slice(0, PAGE_SIZE))
const hasMore = computed(() => todos.value.length > PAGE_SIZE)

async function copyInviteCode() {
  if (!family.value?.invite_code) return
  try {
    await navigator.clipboard.writeText(family.value.invite_code)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch { /* */ }
}

async function completeTodo(todo: Todo, status: string) {
  // Show remark prompt
  remarkTodo.value = todo
  remarkAction.value = status
  remarkText.value = ''
  showRemark.value = true
}

// ─── Remark modal ──────────────────────────────────────────────

const showRemark = ref(false)
const remarkTodo = ref<Todo | null>(null)
const remarkAction = ref('')
const remarkText = ref('')

async function submitRemark() {
  const todo = remarkTodo.value
  if (!todo) return
  try {
    await api.put('/todos/' + todo.id, { status: remarkAction.value, remark: remarkText.value })
    await loadTodos()
    toast.success(remarkAction.value === 'done' ? '已完成' : '已跳过')
    showRemark.value = false
  } catch (e: any) { toast.error(e.message) }
}

function skipRemarkAndComplete() {
  showRemark.value = false
  completeTodoDirect(remarkTodo.value!, remarkAction.value)
}

async function completeTodoDirect(todo: Todo, status: string) {
  try {
    await api.put('/todos/' + todo.id, { status })
    await loadTodos()
    toast.success(status === 'done' ? '已完成' : '已跳过')
  } catch (e: any) { toast.error(e.message) }
}

async function loadTodos() {
  try {
    todos.value = await api.get<Todo[]>('/families/' + familyId + '/todos?status=pending')
  } catch { todos.value = [] }
}

function fmtRange(start: string, end: string): string {
  const s = new Date(start)
  const e = new Date(end)
  const opts: Intl.DateTimeFormatOptions = { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' }
  return s.toLocaleDateString('zh-CN', opts) + ' → ' + e.toLocaleDateString('zh-CN', opts)
}

function getLocName(id: string) { return locations.value.find(l => l.id === id)?.name || '' }

async function loadLocations() {
  try {
    locations.value = await api.get<any[]>('/families/' + familyId + '/locations')
  } catch { locations.value = [] }
}

onMounted(async () => {
  loading.value = true
  await Promise.all([
    loadTodos(),
    loadLocations(),
    (async () => {
      try { family.value = await api.get<Family>('/families/' + familyId) } catch { /* */ }
    })(),
    (async () => {
      try { memberCount.value = (await api.get<any[]>('/families/' + familyId + '/members')).length } catch { /* */ }
    })(),
    (async () => {
      try { groupCount.value = (await api.get<any[]>('/families/' + familyId + '/groups')).length } catch { /* */ }
    })(),
  ])
  loading.value = false
})
</script>

<template>
  <div>

    <LoadingSpinner v-if="loading" />

    <template v-else>

    <!-- Tabs -->
    <div class="flex gap-1 mb-4 border-b dark:border-gray-700">
      <button class="px-4 py-2 text-sm font-medium border-b-2 transition-colors"
        :class="activeTab === 'todos' ? 'border-primary text-primary' : 'border-transparent text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'"
        @click="activeTab = 'todos'"
      >待办 ({{ todos.length }})</button>
      <button class="px-4 py-2 text-sm font-medium border-b-2 transition-colors"
        :class="activeTab === 'overview' ? 'border-primary text-primary' : 'border-transparent text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'"
        @click="activeTab = 'overview'"
      >{{ t('dashboard.overview') }}</button>
    </div>

    <!-- Todos Tab -->
    <div v-if="activeTab === 'todos'">
      <div v-if="todos.length === 0" class="text-center text-gray-400 py-8">暂无待办事项 🎉</div>
      <div v-else class="grid grid-cols-1 sm:grid-cols-2 gap-3 max-w-2xl items-stretch">
        <div v-for="todo in displayTodos" :key="todo.id"
          class="card flex flex-col gap-1.5 hover:shadow-md transition-shadow h-full"
        >
          <!-- Header: name + kind badge -->
          <div class="flex items-start justify-between gap-2">
            <p class="font-medium dark:text-gray-200 text-sm leading-snug line-clamp-2">{{ todo.task?.name || todo.task_id }}</p>
            <span v-if="getTodoBadge(todo.task?.kind || '')"
              class="text-[10px] px-1.5 py-0.5 rounded-full bg-purple-100 dark:bg-purple-900/40 text-purple-600 dark:text-purple-400 flex-shrink-0"
            >{{ getTodoBadge(todo.task?.kind || '') }}</span>
          </div>

          <!-- Meta info -->
          <div class="space-y-0.5">
            <p class="text-xs text-gray-400 flex items-center gap-1">
              <span>🕐</span>
              <span>{{ fmtRange(todo.due_start, todo.due_date) }}</span>
            </p>
            <component
              :is="getTodoInfo(todo.task?.kind || '')"
              v-if="getTodoInfo(todo.task?.kind || '')"
              :todo="todo"
            />
            <p v-if="todo.location_id && getLocName(todo.location_id)" class="text-xs text-primary flex items-center gap-1">
              <span>📍</span>
              <span>{{ getLocName(todo.location_id) }}</span>
            </p>
          </div>

          <!-- Actions -->
          <div class="flex gap-1.5 pt-1.5 border-t border-gray-100 dark:border-gray-700 mt-auto">
            <component
              :is="getTodoActions(todo.task?.kind || '')"
              :todo="todo"
              @done="completeTodo($event, 'done')"
              @skip="completeTodo($event, 'skipped')"
              @completed="loadTodos"
            />
          </div>
        </div>
      </div>
      <div v-if="hasMore" class="text-center mt-4">
        <button class="text-sm text-primary hover:underline font-medium" @click="showAll = !showAll">
          {{ showAll ? '收起' : `显示全部 (${todos.length})` }}
        </button>
      </div>
    </div>

    <!-- Overview Tab -->
    <div v-if="activeTab === 'overview'">
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-6">
        <div class="card"><p class="text-gray-400 text-sm">{{ t('dashboard.members') }}</p><p class="text-2xl font-bold dark:text-gray-200">{{ memberCount }}</p></div>
        <div class="card"><p class="text-gray-400 text-sm">{{ t('dashboard.groups') }}</p><p class="text-2xl font-bold dark:text-gray-200">{{ groupCount }}</p></div>
      </div>

      <div v-if="family?.invite_code" class="card">
        <p class="text-gray-400 text-sm mb-1">{{ t('dashboard.inviteCode') }}</p>
        <p class="text-gray-500 text-xs mb-3">{{ t('dashboard.inviteCodeHint') }}</p>
        <div class="flex items-center gap-3">
          <code class="flex-1 bg-gray-100 dark:bg-gray-700 px-4 py-2 rounded-lg text-lg font-mono tracking-wider text-center select-all dark:text-gray-200">{{ family.invite_code }}</code>
          <button class="px-4 py-2 rounded-lg bg-primary text-white text-sm font-medium hover:opacity-90 transition-opacity whitespace-nowrap" @click="copyInviteCode">{{ copied ? t('dashboard.copied') : t('dashboard.copy') }}</button>
        </div>
      </div>
    </div>

    </template>

    <!-- Remark Modal -->
    <Teleport to="body">
      <div v-if="showRemark" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60" @mousedown.self="showRemark = false">
        <div class="bg-white dark:bg-gray-800 rounded-xl shadow-2xl w-[90vw] max-w-md">
          <div class="flex items-center justify-between px-4 py-3 border-b dark:border-gray-700">
            <h3 class="font-bold dark:text-gray-200">{{ remarkAction === 'done' ? '✅ 完成' : '⏭️ 跳过' }} — 备注</h3>
            <button class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 text-lg" @click="showRemark = false">✕</button>
          </div>
          <div class="p-4 space-y-3">
            <p class="text-sm text-gray-500 dark:text-gray-400">{{ remarkTodo?.task?.name }}</p>
            <textarea
              v-model="remarkText"
              class="input w-full h-20 resize-none text-sm"
              placeholder="添加备注（可选）..."
              @keydown.ctrl.enter="submitRemark"
            ></textarea>
            <div class="flex gap-2">
              <button class="btn-primary text-sm flex-1" @click="submitRemark">确认</button>
              <button class="text-sm px-4 py-2 rounded text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-700" @click="skipRemarkAndComplete">跳过备注</button>
            </div>
            <p class="text-[10px] text-gray-400">Ctrl+Enter 快速提交</p>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
