<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from '@/i18n'
import { useToast } from '@/composables/useToast'
import { useLoading } from '@/composables/useLoading'
import { useConfirm } from '@/composables/useConfirm'
import { useErrorHandler } from '@/composables/useErrorHandler'
import ErrorDisplay from '@/components/ErrorDisplay.vue'
import {
  listProviders, refreshSystemProvider,
  listAdminSubscriptions, createAdminSubscription,
  updateAdminSubscription, deleteAdminSubscription,
} from '@/api/task-templates'
import type { TemplateProvider, TaskTemplateSubscription } from '@/types'

const { t } = useI18n()
const toast = useToast()
const { error, setError, clearError } = useErrorHandler()
const { withLoading } = useLoading()

const providers = ref<TemplateProvider[]>([])
const subscriptions = ref<TaskTemplateSubscription[]>([])
const refreshing = ref<string | null>(null)

// ── Subscription form ────────────────────────────────────────────
const showSubForm = ref(false)
const editingSub = ref<TaskTemplateSubscription | null>(null)
const subForm = ref({
  provider_code: 'http',
  url: '',
  name: '',
  auto_refresh: false,
  refresh_interval_hours: 24,
})

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
      await updateAdminSubscription(editingSub.value.id, {
        url: subForm.value.url,
        name: subForm.value.name,
        auto_refresh: subForm.value.auto_refresh,
        refresh_interval_hours: subForm.value.refresh_interval_hours,
      })
      toast.success(t('taskTemplate.subscription.updated'))
    } else {
      await createAdminSubscription(subForm.value)
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
    await deleteAdminSubscription(sub.id)
    toast.success(t('taskTemplate.subscription.deleted'))
    await loadData()
  } catch (e: any) {
    setError(e)
  }
}

async function handleRefresh(providerCode: string) {
  refreshing.value = providerCode
  try {
    await refreshSystemProvider(providerCode)
    toast.success(t('taskTemplate.refreshed'))
    await loadData()
  } catch (e: any) {
    setError(e)
  } finally {
    refreshing.value = null
  }
}

async function loadData() {
  await withLoading(async () => {
    const [provs, subs] = await Promise.all([
      listProviders(),
      listAdminSubscriptions(),
    ])
    providers.value = provs
    subscriptions.value = subs
  })
}

onMounted(() => { loadData() })
</script>

<template>
  <div>
    <ErrorDisplay :error="error" @close="clearError" />
    <!-- ── Providers ──────────────────────────────────────────── -->
    <div class="mb-8">
      <h3 class="text-sm font-medium text-gray-500 dark:text-gray-400 mb-3">
        {{ t('taskTemplate.provider') }}
      </h3>
      <div class="space-y-2">
        <div
          v-for="prov in providers" :key="prov.code"
          class="flex items-center justify-between bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 px-4 py-3"
        >
          <div class="flex items-center gap-3">
            <span class="text-sm font-medium text-gray-700 dark:text-gray-200">{{ prov.name }}</span>
            <span class="text-xs px-2 py-0.5 rounded-full"
              :class="prov.sync_status === 'idle' ? 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300'
                : prov.sync_status === 'syncing' ? 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300'
                : 'bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300'"
            >
              {{ prov.sync_status }}
            </span>
            <span v-if="prov.last_sync_at" class="text-xs text-gray-400 dark:text-gray-500">
              {{ new Date(prov.last_sync_at).toLocaleString() }}
            </span>
          </div>
          <button
            class="px-3 py-1 text-xs rounded-md bg-green-500 hover:bg-green-600 text-white disabled:opacity-50 transition-colors"
            :disabled="refreshing === prov.code"
            @click="handleRefresh(prov.code)"
          >
            {{ refreshing === prov.code ? t('taskTemplate.refreshing') : t('taskTemplate.refresh') }}
          </button>
        </div>
      </div>
    </div>

    <!-- ── Subscriptions ──────────────────────────────────────── -->
    <div>
      <div class="flex justify-between items-center mb-3">
        <h3 class="text-sm font-medium text-gray-500 dark:text-gray-400">
          {{ t('taskTemplate.subscription.heading') }}
        </h3>
        <button
          class="px-3 py-1 text-xs rounded-md bg-green-500 hover:bg-green-600 text-white transition-colors"
          @click="openCreateSub"
        >
          + {{ t('taskTemplate.subscription.create') }}
        </button>
      </div>

      <div v-if="subscriptions.length === 0" class="text-center text-gray-400 dark:text-gray-500 py-4">
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

    <!-- ── Subscription Form Modal ────────────────────────────── -->
    <div v-if="showSubForm" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="showSubForm = false">
      <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-md mx-4 p-6">
        <h4 class="text-base font-semibold text-gray-900 dark:text-gray-100 mb-4">
          {{ editingSub ? t('taskTemplate.subscription.edit') : t('taskTemplate.subscription.create') }}
        </h4>
        <div class="space-y-3">
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('taskTemplate.subscription.name') }}</label>
            <input
              v-model="subForm.name"
              :placeholder="t('taskTemplate.subscription.namePlaceholder')"
              class="w-full rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('taskTemplate.subscription.url') }}</label>
            <input
              v-model="subForm.url"
              :placeholder="t('taskTemplate.subscription.urlPlaceholder')"
              class="w-full rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm"
            />
          </div>
          <label class="flex items-center gap-2 cursor-pointer">
            <input v-model="subForm.auto_refresh" type="checkbox" class="rounded border-gray-300 text-green-500 focus:ring-green-500" />
            <span class="text-sm text-gray-700 dark:text-gray-300">{{ t('taskTemplate.subscription.autoRefresh') }}</span>
          </label>
          <div v-if="subForm.auto_refresh">
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('taskTemplate.subscription.intervalHours') }}</label>
            <input v-model.number="subForm.refresh_interval_hours" type="number" min="1" class="w-full rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm" />
          </div>
        </div>
        <div class="flex justify-end gap-2 mt-4">
          <button class="px-4 py-2 text-sm rounded-md border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors" @click="showSubForm = false">
            {{ t('confirm.cancel') }}
          </button>
          <button class="px-4 py-2 text-sm rounded-md bg-green-500 hover:bg-green-600 text-white font-medium transition-colors" @click="handleSaveSub">
            {{ t('confirm.ok') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
