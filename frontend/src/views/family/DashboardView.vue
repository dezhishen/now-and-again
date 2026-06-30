<script setup lang="ts">
import {computed, inject, onMounted, ref, type Ref, watch} from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from '@/i18n'
import { api } from '@/api/client'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import { useToast } from '@/composables/useToast'
import { useLoading } from '@/composables/useLoading'
import { useErrorHandler } from '@/composables/useErrorHandler'
import ErrorDisplay from '@/components/ErrorDisplay.vue'
import { getTodoActions, getTodoInfo, getTodoBadgeKey } from '@/composables/useTaskKinds'
import { initTaskKinds } from '@/components/tasks/init'
import type { Family, Todo } from '@/types'

initTaskKinds()

const { t, td, locale } = useI18n()
const auth = useAuthStore()
const localeCode = computed(() => locale.value)
const toast = useToast()
const familyId = () => auth.activeFamilyId || ''

// Reload on tab activation — loading spinner shown automatically
const refreshKey = inject<Ref<string>>('refreshKey', ref(''))

const activeTab = ref<'todos' | 'overview'>('todos')
const { loading, withLoading } = useLoading()
const { error, setError, clearError } = useErrorHandler()
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

async function completeTodo(todo: Todo, _status: string) {
  remarkTodo.value = todo
  remarkText.value = ''
  showRemark.value = true
}

// ─── Remark modal ──────────────────────────────────────────────

const showRemark = ref(false)
const remarkTodo = ref<Todo | null>(null)
const remarkText = ref('')
const remarkSubmitting = ref(false)

async function submitRemark() {
  const todo = remarkTodo.value
  if (!todo || remarkSubmitting.value) return
  remarkSubmitting.value = true
  try {
    await api.put('/todos/' + todo.id, { todo: { status: 'done', remark: remarkText.value } })
    await withLoading(loadTodos)
    toast.success(t('dashboard.completed'))
    showRemark.value = false
  } catch (e: any) { setError(e) }
  finally { remarkSubmitting.value = false }
}

const processingTodos = ref<Set<string>>(new Set())

async function completeTodoDirect(todo: Todo, status: string) {
  if (processingTodos.value.has(todo.id)) return
  processingTodos.value = new Set([...processingTodos.value, todo.id])
  try {
    await api.put('/todos/' + todo.id, { todo: { status: status as 'done' | 'skipped' } })
    await withLoading(loadTodos)
    toast.success(status === 'done' ? t('dashboard.completed') : t('dashboard.skipped'))
  } catch (e: any) { setError(e) }
  finally {
    const next = new Set(processingTodos.value)
    next.delete(todo.id)
    processingTodos.value = next
  }
}

async function loadTodos() {
  todos.value = await api.get<Todo[]>('/todos?status=pending')
}

async function loadAll() {
  await withLoading(async () => {
    await Promise.all([
      loadTodos(),
      loadLocations(),
      (async () => { try { family.value = await api.get<Family>('/families/' + familyId()) } catch { /* */ } })(),
      (async () => { try { memberCount.value = (await api.get<any[]>('/members')).length } catch { /* */ } })(),
      (async () => { try { groupCount.value = (await api.get<any[]>('/groups')).length } catch { /* */ } })(),
    ])
  })
}

function fmtRange(start: string, end: string): string {
  const s = new Date(start)
  const e = new Date(end)
  const locale = (localeCode.value === 'en' ? 'en-US' : 'zh-CN')
  const opts: Intl.DateTimeFormatOptions = { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' }
  return s.toLocaleDateString(locale, opts) + ' → ' + e.toLocaleDateString(locale, opts)
}

function getLocName(id: string) { return locations.value.find(l => l.id === id)?.name || '' }

async function loadLocations() {
  locations.value = await api.get<any[]>('/locations')
}

// Reload on tab activation — loading spinner shown automatically
watch(refreshKey, (newVal) => {
  if (newVal === 'dashboard') withLoading(async () => { await loadTodos(); await loadLocations() })
})

onMounted(() => { loadAll() })
</script>

<template>
  <div>

    <LoadingSpinner :text="t('app.loading')" v-if="loading" />

    <template v-else>

    <!-- Tabs -->
    <div class="flex gap-1 mb-4 border-b dark:border-gray-700">
      <button class="px-4 py-2 text-sm font-medium border-b-2 transition-colors"
        :class="activeTab === 'todos' ? 'border-primary text-primary' : 'border-transparent text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'"
        @click="activeTab = 'todos'"
      >{{ t('dashboard.todos') }} ({{ todos.length }})</button>
      <button class="px-4 py-2 text-sm font-medium border-b-2 transition-colors"
        :class="activeTab === 'overview' ? 'border-primary text-primary' : 'border-transparent text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'"
        @click="activeTab = 'overview'"
      >{{ t('dashboard.overview') }}</button>
    </div>

    <!-- Todos Tab -->
    <div v-if="activeTab === 'todos'">
      <div v-if="todos.length === 0" class="text-center text-gray-400 py-8">{{ t('dashboard.noTodos') }}</div>
      <div v-else class="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-3 items-stretch">
        <div v-for="todo in displayTodos" :key="todo.id"
          class="card flex flex-col gap-1.5 hover:shadow-md transition-shadow h-full"
        >
          <!-- Header: name + kind badge -->
          <div class="flex items-start justify-between gap-2">
            <p class="font-medium dark:text-gray-200 text-sm leading-snug line-clamp-2">{{ todo.task?.name || todo.task_name || todo.task_id }}</p>
            <span v-if="getTodoBadgeKey(todo.task?.kind || todo.task_kind || '')"
              class="text-[10px] px-1 py-0.5 rounded bg-purple-100 dark:bg-purple-900/30 text-purple-600 dark:text-purple-300 font-medium"
            >{{ td(getTodoBadgeKey(todo.task?.kind || todo.task_kind || '')) }}</span>
          </div>

          <!-- Meta info -->
          <div class="space-y-0.5">
            <p class="text-xs text-gray-400 flex items-center gap-1">
              <span>🕐</span>
              <span>{{ fmtRange(todo.due_start, todo.due_date) }}</span>
            </p>
            <!-- System-generated context (carry-over from previous cycle) -->
            <p v-if="todo.display_summary" class="text-xs text-blue-400 flex items-center gap-1">
              <span>↩️</span>
              <span>{{ todo.display_summary }}</span>
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
            <p v-if="todo.status !== 'pending' && todo.remark" class="text-xs text-gray-400 dark:text-gray-500 italic line-clamp-2">
              💬 {{ todo.remark }}
            </p>
          </div>

          <!-- Actions -->
          <div class="flex gap-1.5 pt-1.5 border-t border-gray-100 dark:border-gray-700 mt-auto">
            <component
              :is="getTodoActions(todo.task?.kind || '')"
              :todo="todo"
              @done="completeTodoDirect($event, 'done')"
              @remark="completeTodo($event, 'done')"
              @skip="completeTodo($event, 'skipped')"
              @completed="loadTodos"
            />
          </div>
        </div>
      </div>
      <div v-if="hasMore" class="text-center mt-4">
        <button class="text-sm text-primary hover:underline font-medium" @click="showAll = !showAll">
          {{ showAll ? t('dashboard.collapse') : t('dashboard.showAll').replace('{count}', String(todos.length)) }}
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
            <h3 class="font-bold dark:text-gray-200">📝 {{ t('dashboard.remarkTitle') }}</h3>
            <button class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 text-lg" @click="showRemark = false">✕</button>
          </div>
          <div class="p-4 space-y-3">
            <ErrorDisplay :error="error" @close="clearError" />
            <p class="text-sm text-gray-500 dark:text-gray-400">{{ remarkTodo?.task?.name }}</p>
            <textarea
              v-model="remarkText"
              class="input w-full h-20 resize-none text-sm"
              :placeholder="t('dashboard.remarkPlaceholder')"
              @keydown.ctrl.enter="submitRemark"
            ></textarea>
            <div class="flex gap-2">
              <button class="btn-primary text-sm flex-1" :disabled="remarkSubmitting" @click="submitRemark">{{ remarkSubmitting ? '...' : t('dashboard.remarkConfirm') }}</button>
              <button class="btn-secondary" @click="showRemark = false">{{ t('dashboard.cancel') }}</button>
            </div>
            <p class="text-[10px] text-gray-400">{{ t('dashboard.ctrlEnter') }}</p>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
