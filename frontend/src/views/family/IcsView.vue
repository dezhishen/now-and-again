<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { api } from '@/api/client'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import { useToast } from '@/composables/useToast'
import type { FamilyGroup, ApiKey } from '@/types'

interface IcsFeed {
  id: string; family_id: string; name: string; description?: string
  filter_days: number; filter_group_id?: string; filter_type: string
  auth_type: string; app_username?: string; api_key_prefix?: string
  ics_url: string; enabled: boolean; created_at: string
}

const route = useRoute()
const familyId = route.params.familyId as string
const feeds = ref<IcsFeed[]>([])
const groups = ref<FamilyGroup[]>([])
const apiKeys = ref<ApiKey[]>([])
const loading = ref(true)
const toast = useToast()

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

onMounted(async () => {
  loading.value = true
  await Promise.all([loadFeeds(), loadGroups(), loadApiKeys()])
  loading.value = false
})

async function loadFeeds() {
  try { feeds.value = await api.get<IcsFeed[]>('/families/' + familyId + '/ics-feeds') } catch { feeds.value = [] }
}
async function loadGroups() {
  try { groups.value = await api.get<FamilyGroup[]>('/families/' + familyId + '/groups') } catch { groups.value = [] }
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
      toast.success('订阅已更新')
    } else {
      await api.post('/families/' + familyId + '/ics-feeds', body)
      toast.success('ICS 订阅已创建')
    }
    showForm.value = false
    await loadFeeds()
  } catch (e: any) { toast.error(e.message) }
}

async function deleteFeed(id: string) {
  if (!confirm('确定删除此 ICS 订阅？')) return
  try { await api.delete('/ics-feeds/' + id); await loadFeeds(); toast.success('已删除') } catch (e: any) { toast.error(e.message) }
}

function copyLink(url: string) {
  const full = baseUrl + url
  navigator.clipboard.writeText(full).then(() => {
    toast.success('链接已复制！')
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
    <h2 class="text-xl md:text-2xl font-bold mb-4 dark:text-gray-200">📅 日历订阅 (ICS)</h2>

    <LoadingSpinner v-if="loading" />
    <template v-else>

    <!-- Guide -->
    <div v-if="feeds.length === 0 && !showForm" class="card mb-6 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800">
      <h3 class="font-bold text-blue-700 dark:text-blue-300 mb-2">� 日历订阅</h3>
      <ol class="text-sm text-blue-600 dark:text-blue-400 space-y-1 list-decimal pl-4">
        <li>创建一个订阅，配置待办范围和认证方式</li>
        <li>复制生成的链接</li>
        <li>粘贴到 Apple 日历、Google 日历、Outlook 等任意日历应用中</li>
      </ol>
    </div>

    <button class="btn-primary text-sm mb-4" @click="openCreate">+ 创建订阅</button>

    <!-- Create/Edit Form -->
    <div v-if="showForm" class="card mb-6 space-y-3">
      <h3 class="font-bold dark:text-gray-200">{{ editingFeed ? '编辑订阅' : '创建 ICS 订阅' }}</h3>

      <div>
        <label class="text-xs text-gray-400 block mb-1">订阅名称</label>
        <input v-model="feedName" class="input" placeholder="如：家庭待办日历" />
      </div>
      <div>
        <label class="text-xs text-gray-400 block mb-1">描述（可选）</label>
        <input v-model="feedDesc" class="input" placeholder="简要描述此订阅" />
      </div>
      <div class="grid grid-cols-2 gap-3">
        <div>
          <label class="text-xs text-gray-400 block mb-1">显示未来 N 天</label>
          <input v-model.number="feedDays" type="number" class="input" min="1" max="365" />
        </div>
        <div>
          <label class="text-xs text-gray-400 block mb-1">筛选小组（可选）</label>
          <select v-model="feedGroupID" class="input">
            <option value="">全部</option>
            <option v-for="g in groups" :key="g.id" :value="g.id">{{ g.name }}</option>
          </select>
        </div>
      </div>

      <!-- Auth Type -->
      <div>
        <label class="text-xs text-gray-400 block mb-2">认证方式</label>
        <div class="flex gap-3">
          <label class="flex items-center gap-1 text-sm cursor-pointer">
            <input type="radio" v-model="feedAuthType" value="api_key" class="accent-primary" />
            <span class="dark:text-gray-300">API Key</span>
          </label>
          <label class="flex items-center gap-1 text-sm cursor-pointer">
            <input type="radio" v-model="feedAuthType" value="basic" class="accent-primary" />
            <span class="dark:text-gray-300">Basic Auth（用户名+密码）</span>
          </label>
        </div>
      </div>

      <!-- API Key selection -->
      <div v-if="feedAuthType === 'api_key'">
        <label class="text-xs text-gray-400 block mb-1">选择 API Key</label>
        <select v-model="feedApiKeyID" class="input">
          <option value="">-- 选择 --</option>
          <option v-for="k in apiKeys" :key="k.id" :value="k.id">{{ k.name }} ({{ k.key_prefix }})</option>
        </select>
        <p class="text-xs text-gray-400 mt-1">
          尚无？<router-link to="/api-keys" class="text-primary underline">新建 API Key</router-link>
        </p>
        <p class="text-xs text-gray-500 mt-2">
          💡 创建后，订阅 URL 格式为：<br/>
          <code class="text-primary">{{ baseUrl }}/api/ics/xxx.ics?key=你的API_KEY</code>
        </p>
      </div>

      <!-- Basic Auth -->
      <div v-if="feedAuthType === 'basic'" class="space-y-2">
        <p class="text-xs text-gray-500">用户名使用您的登录账号，只需设置密码</p>
        <div>
          <label class="text-xs text-gray-400 block mb-1">密码</label>
          <input v-model="feedAppPass" type="password" class="input" placeholder="设置 ICS 订阅密码" />
        </div>
      </div>

      <div class="flex gap-2">
        <button class="btn-primary text-sm" @click="saveFeed">{{ editingFeed ? '保存' : '创建' }}</button>
        <button class="text-sm px-3 py-1 rounded text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-700" @click="showForm = false">取消</button>
      </div>
    </div>

    <!-- Feed List -->
    <div v-if="feeds.length > 0">
      <div v-for="feed in feeds" :key="feed.id" class="card mb-3">
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2">
              <span class="font-medium dark:text-gray-200">{{ feed.name }}</span>
              <span v-if="!feed.enabled" class="text-xs text-gray-400">(已禁用)</span>
            </div>
            <p v-if="feed.description" class="text-xs text-gray-400 mt-0.5">{{ feed.description }}</p>
            <div class="flex flex-wrap items-center gap-2 mt-1.5">
              <span class="text-xs px-1.5 py-0.5 rounded bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400">
                {{ feed.filter_days }} 天
              </span>
              <span class="text-xs text-gray-400">
                {{ getAuthLabel(feed) }}
              </span>
            </div>
            <!-- URL -->
            <div class="mt-2 space-y-1">
              <div class="flex items-center gap-2 bg-gray-50 dark:bg-gray-800 rounded px-2 py-1.5">
                <code class="text-xs text-primary break-all flex-1">{{ getIcsUrl(feed) }}</code>
                <button class="text-xs px-2 py-0.5 rounded bg-primary text-white hover:opacity-80 flex-shrink-0" @click="copyLink(getIcsUrl(feed))">复制</button>
              </div>
              <p class="text-xs text-gray-500">{{ getUsageHint(feed) }}</p>
            </div>
          </div>
          <div class="flex gap-1 flex-shrink-0">
            <button class="text-xs px-2 py-1 rounded hover:bg-gray-100 dark:hover:bg-gray-700 dark:text-gray-400" @click="openEdit(feed)">编辑</button>
            <button class="text-xs px-2 py-1 rounded text-danger hover:bg-red-50 dark:hover:bg-red-900/30" @click="deleteFeed(feed.id)">删除</button>
          </div>
        </div>
      </div>
    </div>
    </template>
  </div>
</template>
