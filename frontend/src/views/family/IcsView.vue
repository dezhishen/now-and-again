<script setup lang="ts">
import {computed, inject, onMounted, ref, type Ref, watch} from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from '@/i18n'
import { useConfirm } from '@/composables/useConfirm'
import { api } from '@/api/client'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import { useToast } from '@/composables/useToast'
import { useLoading } from '@/composables/useLoading'
import { useErrorHandler } from '@/composables/useErrorHandler'
import ErrorDisplay from '@/components/ErrorDisplay.vue'
import type { FamilyGroup, ApiKey } from '@/types'

const { t } = useI18n()
  const auth = useAuthStore()

interface IcsFeed {
  id: string; family_id: string; name: string; description?: string
  filter_days: number; filter_group_id?: string; filter_type: string
  auth_type: string; app_username?: string; api_key_prefix?: string
  ics_url: string; enabled: boolean; created_at: string
}

const familyId = () => auth.activeFamilyId || ''

// Reload data when this tab becomes active
const refreshKey = inject<Ref<string>>('refreshKey', ref(''))
watch(refreshKey, (newVal) => { if (newVal === 'ics') withLoading(async () => { await loadFeeds(); await loadGroups(); await loadApiKeys() }) })

const feeds = ref<IcsFeed[]>([])
const groups = ref<FamilyGroup[]>([])
const apiKeys = ref<ApiKey[]>([])
const { loading, withLoading } = useLoading()
const toast = useToast()
const { error, setError, clearError } = useErrorHandler()

// Form state

// Form state
const showForm = ref(false)
const editingFeed = ref<IcsFeed | null>(null)
const feedName = ref('')
const feedDesc = ref('')
const feedDays = ref(7)
const feedGroupID = ref('')
const feedAuthType = ref<'api_key' | 'basic'>('api_key')
const feedApiKeyID = ref('')
const feedAppPass = ref('')

const baseUrl = window.location.origin

// ─── Embed calendar ─────────────────────────────────────────────
const showEmbed = ref(false)
const embedApiKey = ref('')
const embedGroupID = ref('')
const embedRefresh = ref(30)  // default 30 seconds
const embedCopied = ref(false)

const embedUrl = computed(() => {
  let url = `${baseUrl}/calendar/${familyId}?key=${embedApiKey.value || 'YOUR_API_KEY'}`
  if (embedGroupID.value) url += `&group_id=${embedGroupID.value}`
  if (embedRefresh.value > 0) url += `&refresh=${embedRefresh.value}`
  return url
})

const embedCode = computed(() => {
  return `<embed src="${embedUrl.value}" width="100%" height="600" type="text/html" style="border:1px solid #ddd;border-radius:8px" />`
})

function copyEmbed() {
  navigator.clipboard.writeText(embedCode.value).then(() => {
    embedCopied.value = true
    setTimeout(() => { embedCopied.value = false }, 2000)
  })
}

onMounted(() => {
  withLoading(async () => {
    await loadFeeds()
    await loadGroups()
    await loadApiKeys()
  })
})

async function loadFeeds() {
  try { feeds.value = await api.get<IcsFeed[]>('/ics-feeds') } catch { feeds.value = [] }
}
async function loadGroups() {
  try { groups.value = await api.get<FamilyGroup[]>('/groups') } catch { groups.value = [] }
}
async function loadApiKeys() {
  try { apiKeys.value = await api.get<ApiKey[]>('/users/me/api-keys') } catch { apiKeys.value = [] }
}

function openCreate() {
  editingFeed.value = null
  resetForm()
  showForm.value = true
}
function openEdit(feed: IcsFeed) {
  editingFeed.value = feed
  feedName.value = feed.name
  feedDesc.value = feed.description || ''
  feedDays.value = feed.filter_days
  feedGroupID.value = feed.filter_group_id || ''
  feedAuthType.value = feed.auth_type as 'api_key' | 'basic'
  feedApiKeyID.value = ''
  feedAppPass.value = ''
  showForm.value = true
}
function resetForm() {
  feedName.value = ''; feedDesc.value = ''; feedDays.value = 7
  feedGroupID.value = ''; feedAuthType.value = 'api_key'
  feedApiKeyID.value = ''; feedAppPass.value = ''
}

async function saveFeed() {
  const body: any = {
    name: feedName.value,
    description: feedDesc.value,
    filter_days: feedDays.value,
    filter_group_id: feedGroupID.value,
    auth_type: feedAuthType.value,
  }
  if (feedAuthType.value === 'api_key') {
    body.api_key_id = feedApiKeyID.value
  } else {
    body.app_password = feedAppPass.value
  }

  try {
    if (editingFeed.value) {
      await api.put('/ics-feeds/' + editingFeed.value.id, body)
      toast.success(t('ics.updated'))
    } else {
      await api.post('/ics-feeds', body)
      toast.success(t('ics.created'))
    }
    showForm.value = false
    await loadFeeds()
  } catch (e: any) { setError(e) }
}

async function deleteFeed(id: string) {
  if (!await useConfirm(t('ics.deleteConfirm'))) return
  try { await api.delete('/ics-feeds/' + id); await loadFeeds(); toast.success(t('ics.deleted')) } catch (e: any) { setError(e) }
}

function copyLink(url: string) {
  const full = baseUrl + url
  navigator.clipboard.writeText(full).then(() => {
    toast.success(t('ics.copyLinkDone'))
  })
}

function getAuthLabel(feed: IcsFeed): string {
  if (feed.auth_type === 'basic') return `Basic Auth (${feed.app_username})`
  if (feed.api_key_prefix) return `API Key: ${feed.api_key_prefix}`
  return 'API Key'
}

function getIcsUrl(feed: IcsFeed): string {
  const base = baseUrl + feed.ics_url
  if (feed.auth_type === 'api_key') {
    return base + '?key=你的API_KEY'
  }
  return base
}

function getUsageHint(feed: IcsFeed): string {
  if (feed.auth_type === 'api_key') {
    return '将 ?key= 替换为你的 API Key 完整值'
  }
  if (feed.auth_type === 'basic') {
    return `用户名: ${feed.app_username}，密码: 你设置的密码`
  }
  return ''
}
</script>

<template>
  <div>

    <LoadingSpinner :text="t('app.loading')" v-if="loading" />
    <template v-else>

    <!-- Guide -->
    <div v-if="feeds.length === 0 && !showForm" class="card mb-6 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800">
      <h3 class="font-bold text-blue-700 dark:text-blue-300 mb-2">📅 {{ t('ics.guide') }}</h3>
      <ol class="text-sm text-blue-600 dark:text-blue-400 space-y-1 list-decimal pl-4">
        <li>{{ t('ics.guideStep1') }}</li>
        <li>{{ t('ics.guideStep2') }}</li>
        <li>{{ t('ics.guideStep3') }}</li>
      </ol>
    </div>

    <div class="flex gap-2 mb-4">
      <button class="btn-primary" @click="openCreate">+ {{ t('ics.createFeed') }}</button>
      <button class="btn-secondary" @click="showEmbed = true">🖥️ {{ t('ics.embedBtn') }}</button>
    </div>

    <!-- Create/Edit Form -->
    <div v-if="showForm" class="card mb-6 space-y-3">
      <h3 class="font-bold dark:text-gray-200">{{ editingFeed ? t('ics.editFeed') : t('ics.createFeedTitle') }}</h3>

      <ErrorDisplay :error="error" @close="clearError" />

      <div>
        <label class="text-xs text-gray-400 block mb-1">{{ t('ics.feedName') }}</label>
        <input v-model="feedName" class="input" :placeholder="t('ics.feedNamePlaceholder')" />
      </div>
      <div>
        <label class="text-xs text-gray-400 block mb-1">{{ t('ics.description') }}</label>
        <input v-model="feedDesc" class="input" :placeholder="t('ics.descPlaceholder')" />
      </div>
      <div class="grid grid-cols-2 gap-3">
        <div>
          <label class="text-xs text-gray-400 block mb-1">{{ t('ics.filterDays') }}</label>
          <input v-model.number="feedDays" type="number" class="input" min="1" max="365" />
        </div>
        <div>
          <label class="text-xs text-gray-400 block mb-1">{{ t('ics.filterGroup') }}</label>
          <select v-model="feedGroupID" class="input">
            <option value="">{{ t('ics.all') }}</option>
            <option v-for="g in groups" :key="g.id" :value="g.id">{{ g.name }}</option>
          </select>
        </div>
      </div>

      <!-- Auth Type -->
      <div>
        <label class="text-xs text-gray-400 block mb-2">{{ t('ics.authType') }}</label>
        <div class="flex gap-3">
          <label class="flex items-center gap-1 text-sm cursor-pointer">
            <input type="radio" v-model="feedAuthType" value="api_key" class="accent-primary" />
            <span class="dark:text-gray-300">API Key</span>
          </label>
          <label class="flex items-center gap-1 text-sm cursor-pointer">
            <input type="radio" v-model="feedAuthType" value="basic" class="accent-primary" />
            <span class="dark:text-gray-300">{{ t('ics.basicAuth') }}</span>
          </label>
        </div>
      </div>

      <!-- API Key selection -->
      <div v-if="feedAuthType === 'api_key'">
        <label class="text-xs text-gray-400 block mb-1">{{ t('ics.selectKey') }}</label>
        <select v-model="feedApiKeyID" class="input">
          <option value="">{{ t('ics.select') }}</option>
          <option v-for="k in apiKeys" :key="k.id" :value="k.id">{{ k.name }} ({{ k.key_prefix }})</option>
        </select>
        <p class="text-xs text-gray-400 mt-1">
          {{ t('ics.noKeyHint') }}<router-link to="/api-keys" class="text-primary underline">{{ t('ics.createKey') }}</router-link>
        </p>
        <p class="text-xs text-gray-500 mt-2">
          💡 {{ t('ics.keyUrlHint') }}<br/>
          <code class="text-primary">{{ baseUrl }}/api/ics/xxx.ics?key=你的API_KEY</code>
        </p>
      </div>

      <!-- Basic Auth -->
      <div v-if="feedAuthType === 'basic'" class="space-y-2">
        <p class="text-xs text-gray-500">{{ t('ics.basicHint') }}</p>
        <div>
          <label class="text-xs text-gray-400 block mb-1">{{ t('ics.password') }}</label>
          <input v-model="feedAppPass" type="password" class="input" :placeholder="t('ics.passwordPlaceholder')" />
        </div>
      </div>

      <div class="flex gap-2">
        <button class="btn-primary" @click="saveFeed">{{ editingFeed ? t('ics.save') : t('ics.create') }}</button>
        <button class="btn-secondary" @click="showForm = false">{{ t('ics.cancel') }}</button>
      </div>
    </div>

    <!-- Feed List -->
    <div v-if="feeds.length > 0">
      <div v-for="feed in feeds" :key="feed.id" class="card mb-3">
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2">
              <span class="font-medium dark:text-gray-200">{{ feed.name }}</span>
              <span v-if="!feed.enabled" class="text-xs text-gray-400">{{ t('ics.disabled') }}</span>
            </div>
            <p v-if="feed.description" class="text-xs text-gray-400 mt-0.5">{{ feed.description }}</p>
            <div class="flex flex-wrap items-center gap-2 mt-1.5">
              <span class="text-xs px-1.5 py-0.5 rounded bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400">
                {{ feed.filter_days }} {{ t('ics.days') }}
              </span>
              <span class="text-xs text-gray-400">
                {{ getAuthLabel(feed) }}
              </span>
            </div>
            <!-- URL -->
            <div class="mt-2 space-y-1">
              <div class="flex items-center gap-2 bg-gray-50 dark:bg-gray-800 rounded px-2 py-1.5">
                <code class="text-xs text-primary break-all flex-1">{{ getIcsUrl(feed) }}</code>
                <button class="text-xs px-2 py-0.5 rounded bg-primary text-white hover:opacity-80 flex-shrink-0" @click="copyLink(getIcsUrl(feed))">{{ t('ics.copy') }}</button>
              </div>
              <p class="text-xs text-gray-500">{{ getUsageHint(feed) }}</p>
            </div>
          </div>
          <div class="flex gap-1 flex-shrink-0">
            <button class="btn-ghost text-xs" @click="openEdit(feed)">{{ t('ics.edit') }}</button>
            <button class="btn-ghost text-xs text-danger hover:bg-red-50 dark:hover:bg-red-900/30" @click="deleteFeed(feed.id)">{{ t('ics.delete') }}</button>
          </div>
        </div>
      </div>
    </div>
    </template>

    <!-- Embed Modal -->
    <div v-if="showEmbed" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40" @mousedown.self="showEmbed = false">
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl w-full max-w-lg mx-4 p-6 max-h-[90vh] overflow-y-auto">
        <div class="flex items-center justify-between mb-4">
          <h3 class="font-bold text-lg dark:text-gray-200">🖥️ {{ t('ics.embedTitle') }}</h3>
          <button class="btn-icon" @click="showEmbed = false">✕</button>
        </div>
        <p class="text-xs text-gray-400 mb-4">{{ t('ics.embedDesc') }}</p>

        <div class="space-y-3">
          <div>
            <label class="text-xs text-gray-400 block mb-1">{{ t('ics.apiKeyLabel') }}</label>
            <input v-model="embedApiKey" class="input" :placeholder="t('ics.apiKeyPlaceholder')" />
          </div>
          <div>
            <label class="text-xs text-gray-400 block mb-1">筛选小组（可选）</label>
            <select v-model="embedGroupID" class="input text-sm">
              <option value="">全部</option>
              <option v-for="g in groups" :key="g.id" :value="g.id">{{ g.name }}</option>
            </select>
          </div>
          <div>
            <label class="text-xs text-gray-400 block mb-1">{{ t('ics.autoRefresh') }}</label>
            <select v-model="embedRefresh" class="input">
              <option :value="0">{{ t('ics.off') }}</option>
              <option :value="30">{{ t('ics.seconds30') }}</option>
              <option :value="60">{{ t('ics.minute1') }}</option>
              <option :value="120">{{ t('ics.minute2') }}</option>
              <option :value="300">{{ t('ics.minute5') }}</option>
            </select>
          </div>

          <div>
            <label class="text-xs text-gray-400 block mb-1">{{ t('ics.embedCode') }}</label>
            <div class="flex gap-2">
              <code class="flex-1 bg-gray-100 dark:bg-gray-700 px-3 py-2 rounded-lg text-xs font-mono break-all select-all dark:text-gray-200">{{ embedCode }}</code>
              <button class="px-4 py-2 rounded-lg bg-primary text-white text-sm font-medium hover:opacity-90 transition-opacity whitespace-nowrap" @click="copyEmbed">
                {{ embedCopied ? t('ics.copied') : t('ics.copy') }}
              </button>
            </div>
          </div>

          <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-3">
            <p class="text-xs text-gray-400 mb-1">{{ t('ics.previewUrl') }}</p>
            <a :href="embedUrl" target="_blank" class="text-xs text-primary hover:underline break-all">{{ embedUrl }}</a>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
