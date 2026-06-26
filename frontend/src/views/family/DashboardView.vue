<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { api } from '@/api/client'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import type { Family, Todo } from '@/types'

const { t } = useI18n()
const route = useRoute()
const familyId = route.params.familyId as string

const activeTab = ref<'todos' | 'overview'>('todos')
const loading = ref(true)
const family = ref<Family | null>(null)
const memberCount = ref(0)
const groupCount = ref(0)
const todos = ref<Todo[]>([])
const showAll = ref(false)
const copied = ref(false)
const PAGE_SIZE = 5

const displayTodos = computed(() => showAll.value ? todos.value : todos.value.slice(0, PAGE_SIZE))
const hasMore = computed(() => todos.value.length > PAGE_SIZE)
const error = ref('')

async function copyInviteCode() {
  if (!family.value?.invite_code) return
  try {
    await navigator.clipboard.writeText(family.value.invite_code)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch { /* */ }
}

async function completeTodo(todo: Todo, status: string) {
  try {
    await api.put('/todos/' + todo.id, { status })
    await loadTodos()
  } catch (e: any) { error.value = e.message }
}

async function completeInspection(todo: Todo, result: string) {
  try {
    await api.put('/todos/' + todo.id, { status: 'done', inspection_result: result })
    await loadTodos()
  } catch (e: any) { error.value = e.message }
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

onMounted(async () => {
  loading.value = true
  await Promise.all([
    loadTodos(),
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
    <h2 class="text-xl md:text-2xl font-bold mb-4 dark:text-gray-200">{{ t('dashboard.heading') }}</h2>
    <p v-if="error" class="text-danger text-sm mb-3">{{ error }}</p>

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
      >概览</button>
    </div>

    <!-- Todos Tab -->
    <div v-if="activeTab === 'todos'">
      <div v-if="todos.length === 0" class="text-center text-gray-400 py-8">暂无待办事项 🎉</div>
      <div v-for="todo in displayTodos" :key="todo.id" class="card mb-2 flex items-center justify-between gap-3">
        <div class="min-w-0 flex-1">
          <div class="flex items-center gap-2">
            <p class="font-medium dark:text-gray-200">{{ todo.task?.name || todo.task_id }}</p>
            <span v-if="todo.todo_type === 'inspection'" class="text-xs px-1 rounded bg-warning/20 text-amber-600 dark:text-amber-400">巡检</span>
          </div>
          <p class="text-xs text-gray-400">
            🕐 {{ fmtRange(todo.due_start, todo.due_date) }}
            <span v-if="todo.location_id" class="ml-2 text-primary">📍 {{ todo.location_id }}</span>
          </p>
        </div>
        <!-- Inspection branch buttons -->
        <div v-if="todo.todo_type === 'inspection'" class="flex gap-1 flex-shrink-0 flex-wrap max-w-[120px] justify-end">
          <template v-if="todo.task?.inspection_config?.length">
            <button v-for="b in todo.task.inspection_config" :key="b.name"
              class="text-xs px-2 py-1 rounded hover:opacity-80"
              :class="b.create_todo ? 'bg-red-100 dark:bg-red-900 text-red-600 dark:text-red-300' : 'bg-green-100 dark:bg-green-900 text-green-700 dark:text-green-300'"
              @click="completeInspection(todo, b.name)"
            >{{ b.name }}</button>
          </template>
          <template v-else>
            <button class="text-xs px-2 py-1 rounded bg-green-100 dark:bg-green-900 text-green-700 dark:text-green-300 hover:opacity-80" @click="completeInspection(todo, 'normal')">正常</button>
            <button class="text-xs px-2 py-1 rounded bg-red-100 dark:bg-red-900 text-red-600 dark:text-red-300 hover:opacity-80" @click="completeInspection(todo, 'abnormal')">异常</button>
          </template>
        </div>
        <!-- Regular task buttons -->
        <div v-else class="flex gap-1 flex-shrink-0">
          <button class="text-xs px-2 py-1 rounded bg-green-100 dark:bg-green-900 text-green-700 dark:text-green-300 hover:opacity-80" @click="completeTodo(todo, 'done')">完成</button>
          <button v-if="todo.task?.schedule_type !== 'once'" class="text-xs px-2 py-1 rounded bg-gray-100 dark:bg-gray-700 text-gray-500 hover:opacity-80" @click="completeTodo(todo, 'skipped')">跳过</button>
        </div>
      </div>
      <div v-if="hasMore" class="text-center mt-3">
        <button class="text-xs text-primary hover:underline" @click="showAll = !showAll">
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
  </div>
</template>
