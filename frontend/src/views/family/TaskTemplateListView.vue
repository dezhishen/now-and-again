<script setup lang="ts">
import { ref, inject, watch, onMounted } from 'vue'
import type { Ref } from 'vue'
import { useI18n } from '@/i18n'
import { useToast } from '@/composables/useToast'
import { useLoading } from '@/composables/useLoading'
import { useConfirm } from '@/composables/useConfirm'
import { useErrorHandler } from '@/composables/useErrorHandler'
import ErrorDisplay from '@/components/ErrorDisplay.vue'
import { useRouter } from 'vue-router'
import { listTemplates, deleteFamilyTemplate, listFamilySubscriptions, createFamilySubscription, updateFamilySubscription, deleteFamilySubscription } from '@/api/task-templates'
import type { TaskTemplate, TaskTemplateSubscription } from '@/types'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import TemplateRenderDialog from '@/components/taskTemplate/TemplateRenderDialog.vue'
import TemplateWizard from '@/components/taskTemplate/TemplateWizard.vue'

const { t } = useI18n()
const toast = useToast()
const router = useRouter()
const { loading, withLoading } = useLoading()
const { error, setError, clearError } = useErrorHandler()

const templates = ref<TaskTemplate[]>([])
const selectedTemplate = ref<TaskTemplate | null>(null)
const showRenderDialog = ref(false)
const showFormDialog = ref(false)
const editingTemplate = ref<TaskTemplate | null>(null)

// ── Subscription state ────────────────────────────────────────────
const subscriptions = ref<TaskTemplateSubscription[]>([])
const showSubForm = ref(false)
const editingSub = ref<TaskTemplateSubscription | null>(null)
const subForm = ref({
  provider_code: 'http',
  url: '',
  name: '',
  auto_refresh: false,
  refresh_interval_hours: 24,
})

const refreshKey = inject<Ref<string>>('refreshKey', ref(''))
const isFamilyAdmin = inject<Ref<boolean>>('isFamilyAdmin', ref(false))
const refreshId = 'templates'

// ── Tab state ─────────────────────────────────────────────────────
const activeTab = ref<'builtin' | 'subscriptions'>('builtin')

async function loadData() {
  await withLoading(async () => {
    const [tpls, subs] = await Promise.all([
      listTemplates(),
      listFamilySubscriptions(),
    ])
    templates.value = tpls
    subscriptions.value = subs
  })
}

// ── Template actions ──────────────────────────────────────────────

function handleUseTemplate(tmpl: TaskTemplate) {
  selectedTemplate.value = tmpl
  showRenderDialog.value = true
}

function handleCreateTemplate() {
  editingTemplate.value = null
  showFormDialog.value = true
}

function handleEditTemplate(tmpl: TaskTemplate, event: Event) {
  event.stopPropagation()
  editingTemplate.value = tmpl
  showFormDialog.value = true
}

async function handleDeleteTemplate(tmpl: TaskTemplate, event: Event) {
  event.stopPropagation()
  if (!await useConfirm(t('taskTemplate.deleteConfirm'))) return
  try {
    await deleteFamilyTemplate(tmpl.template_code)
    toast.success(t('taskTemplate.deleted'))
    await loadData()
  } catch (e: any) {
    setError(e)
  }
}

function onFormSaved() {
  showFormDialog.value = false
  editingTemplate.value = null
  loadData()
}

// ── Subscription CRUD ─────────────────────────────────────────────

function openCreateSub() {
  editingSub.value = null
  subForm.value = { provider_code: 'http', url: '', name: '', auto_refresh: false, refresh_interval_hours: 24 }
  showSubForm.value = true
}

function openEditSub(sub: TaskTemplateSubscription) {
  editingSub.value = sub
  subForm.value = {
    provider_code: sub.provider_code,
    url: sub.url,
    name: sub.name,
    auto_refresh: sub.auto_refresh,
    refresh_interval_hours: sub.refresh_interval_hours,
  }
  showSubForm.value = true
}

async function handleSaveSub() {
  try {
    if (editingSub.value) {
      await updateFamilySubscription(editingSub.value.id, {
        url: subForm.value.url,
        name: subForm.value.name,
        auto_refresh: subForm.value.auto_refresh,
        refresh_interval_hours: subForm.value.refresh_interval_hours,
      })
      toast.success(t('taskTemplate.subscription.updated'))
    } else {
      await createFamilySubscription(subForm.value)
      toast.success(t('taskTemplate.subscription.created'))
    }
    showSubForm.value = false
    await loadData()
  } catch (e: any) {
    setError(e)
  }
}

async function handleDeleteSub(sub: TaskTemplateSubscription) {
  if (!await useConfirm(t('taskTemplate.subscription.deleteConfirm'))) return
  try {
    await deleteFamilySubscription(sub.id)
    toast.success(t('taskTemplate.subscription.deleted'))
    await loadData()
  } catch (e: any) {
    setError(e)
  }
}

function handleCreateTask(taskDefaults: any, extraSchema: any) {
  router.push({
    path: '/family/tasks',
    query: { fromTemplate: selectedTemplate.value?.template_code },
    state: { taskDefaults, extraSchema, kind: selectedTemplate.value?.kind },
  })
  showRenderDialog.value = false
}

const kindLabel: Record<string, string> = {
  simple: '简单任务',
  inspection: '巡检任务',
}

onMounted(() => {
  withLoading(async () => { await loadData() })
})

watch(refreshKey, (newVal) => {
  if (newVal === refreshId) withLoading(async () => { await loadData() })
})
</script>

<template>
  <div>
    <ErrorDisplay :error="error" @close="clearError" />
  <LoadingSpinner :text="t('app.loading')" v-if="loading && templates.length === 0" />
  <template v-else>
    <!-- Tabs -->
    <div class="tabs">
      <button class="tab" :class="{ active: activeTab === 'builtin' }"
        @click="activeTab = 'builtin'"
      >{{ t('taskTemplate.builtinTab') }}</button>
      <button class="tab" :class="{ active: activeTab === 'subscriptions' }"
        @click="activeTab = 'subscriptions'"
      >{{ t('taskTemplate.subscription.heading') }}</button>
    </div>

    <!-- Tab: Built-in Templates -->
    <div v-if="activeTab === 'builtin'">
      <!-- Toolbar -->
      <div class="flex items-center gap-2 mb-3">
        <button
          v-if="isFamilyAdmin"
          class="btn-primary text-sm"
          @click="handleCreateTemplate"
        >+ {{ t('taskTemplate.createFamily') }}</button>
        <span class="flex-1" />
      </div>

      <!-- Template Cards -->
      <div>
        <div v-if="templates.length === 0" class="text-center text-gray-400 dark:text-gray-500 py-8">
          {{ t('taskTemplate.empty') }}
        </div>

        <div v-else class="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-2 items-start">
          <div
            v-for="tmpl in templates" :key="tmpl.id"
            class="relative bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4
                   hover:shadow-md transition-shadow cursor-pointer group"
            @click="handleUseTemplate(tmpl)"
          >
            <div v-if="isFamilyAdmin && tmpl.family_id" class="absolute top-2 right-2 flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
              <button
                class="px-2 py-0.5 text-xs rounded bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-600 dark:text-gray-300"
                @click="handleEditTemplate(tmpl, $event)"
              >✎</button>
              <button
                class="px-2 py-0.5 text-xs rounded bg-red-50 dark:bg-red-900 hover:bg-red-100 dark:hover:bg-red-800 text-red-600 dark:text-red-400"
                @click="handleDeleteTemplate(tmpl, $event)"
              >✕</button>
            </div>

            <div class="flex items-start justify-between mb-2">
              <div class="flex items-center gap-2">
                <span class="text-xl">{{ tmpl.icon || '📋' }}</span>
                <h3 class="font-medium text-gray-900 dark:text-gray-100">{{ tmpl.name }}</h3>
              </div>
            </div>
            <p v-if="tmpl.description" class="text-sm text-gray-500 dark:text-gray-400 mb-2 line-clamp-2">
              {{ tmpl.description }}
            </p>
            <div class="flex items-center justify-between text-xs text-gray-400 dark:text-gray-500">
              <span>{{ kindLabel[tmpl.kind] || tmpl.kind }}</span>
              <span v-if="tmpl.parameters?.length">
                {{ tmpl.parameters.length }} {{ t('taskTemplate.parameters') }}
              </span>
            </div>
          </div>
        </div>
      </div>

    </div>

    <!-- Tab: Subscriptions -->
    <div v-if="activeTab === 'subscriptions'">
      <div v-if="isFamilyAdmin" class="flex items-center gap-2 mb-3">
        <button
          class="btn-primary text-sm"
          @click="openCreateSub"
        >+ {{ t('taskTemplate.subscription.create') }}</button>
        <span class="flex-1" />
      </div>

      <div v-if="!isFamilyAdmin" class="text-center text-gray-400 dark:text-gray-500 py-8">
        {{ t('taskTemplate.subscription.adminOnly') }}
      </div>

      <div v-else-if="subscriptions.length === 0" class="text-center text-gray-400 dark:text-gray-500 py-8">
        {{ t('taskTemplate.subscription.empty') }}
      </div>

      <div v-else class="space-y-2">
        <div
          v-for="sub in subscriptions" :key="sub.id"
          class="flex items-center justify-between bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 px-4 py-3"
        >
          <div class="flex-1 min-w-0">
            <div class="text-sm font-medium text-gray-700 dark:text-gray-200 truncate">{{ sub.name }}</div>
            <div class="text-xs text-gray-400 dark:text-gray-500 truncate">{{ sub.url }}</div>
            <div class="flex items-center gap-2 mt-1">
              <span v-if="sub.auto_refresh" class="text-xs text-green-600 dark:text-green-400">
                ⟳ {{ sub.refresh_interval_hours }}h
              </span>
              <span v-if="!sub.enabled" class="text-xs text-red-500">Disabled</span>
            </div>
          </div>
          <div class="flex items-center gap-1 ml-2">
            <button
              class="px-2 py-1 text-xs rounded hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-500 dark:text-gray-400 transition-colors"
              @click="openEditSub(sub)"
            >✎</button>
            <button
              class="px-2 py-1 text-xs rounded hover:bg-red-50 dark:hover:bg-red-900 text-red-500 transition-colors"
              @click="handleDeleteSub(sub)"
            >✕</button>
          </div>
        </div>
      </div>
    </div>
    </template>

    <!-- Subscription Form Modal -->
    <div v-if="showSubForm" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-md mx-4 p-6">
        <h4 class="text-base font-semibold text-gray-900 dark:text-gray-100 mb-4">
          {{ editingSub ? t('taskTemplate.subscription.edit') : t('taskTemplate.subscription.create') }}
        </h4>
        <div class="space-y-3">
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('taskTemplate.subscription.name') }}</label>
            <input v-model="subForm.name" :placeholder="t('taskTemplate.subscription.namePlaceholder')"
              class="input" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('taskTemplate.subscription.url') }}</label>
            <input v-model="subForm.url" :placeholder="t('taskTemplate.subscription.urlPlaceholder')"
              class="input" />
          </div>
          <label class="flex items-center gap-2 cursor-pointer">
            <input v-model="subForm.auto_refresh" type="checkbox" class="rounded border-gray-300 text-green-500 focus:ring-green-500" />
            <span class="text-sm text-gray-700 dark:text-gray-300">{{ t('taskTemplate.subscription.autoRefresh') }}</span>
          </label>
          <div v-if="subForm.auto_refresh">
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('taskTemplate.subscription.intervalHours') }}</label>
            <input v-model.number="subForm.refresh_interval_hours" type="number" min="1"
              class="input" />
          </div>
        </div>
        <div class="flex justify-end gap-2 mt-4">
          <button class="btn-secondary text-sm" @click="showSubForm = false">{{ t('confirm.cancel') }}</button>
          <button class="btn-primary text-sm" @click="handleSaveSub">{{ t('confirm.ok') }}</button>
        </div>
      </div>
    </div>

    <!-- Render Dialog -->
    <TemplateRenderDialog
      v-if="showRenderDialog && selectedTemplate"
      :template="selectedTemplate"
      @close="showRenderDialog = false"
      @create="handleCreateTask"
    />

    <!-- Create/Edit Form Dialog -->
    <TemplateWizard
      v-if="showFormDialog"
      :editing="editingTemplate"
      @close="showFormDialog = false; editingTemplate = null"
      @saved="onFormSaved"
    />
  </div>
</template>
