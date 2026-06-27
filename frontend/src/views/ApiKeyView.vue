<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { api } from '@/api/client'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import type { ApiKey } from '@/types'

const { t } = useI18n()

const keys = ref<ApiKey[]>([])
const loading = ref(true)
const showCreate = ref(false)
const newName = ref('')
const newScopes = ref('')
const newExpires = ref('')
const error = ref('')
const creating = ref(false)
const createdKey = ref<string | null>(null)

const SCOPE_OPTIONS = [
  { groupKey: 'apiKey.scope.family', items: [
    { value: 'family:read', labelKey: 'apiKey.scope.familyRead' },
    { value: 'family:write', labelKey: 'apiKey.scope.familyWrite' },
    { value: 'family:admin', labelKey: 'apiKey.scope.familyAdmin' },
  ]},
  { groupKey: 'apiKey.scope.floorPlan', items: [
    { value: 'floorplan:read', labelKey: 'apiKey.scope.floorPlanRead' },
    { value: 'floorplan:write', labelKey: 'apiKey.scope.floorPlanWrite' },
  ]},
  { groupKey: 'apiKey.scope.task', items: [
    { value: 'task:read', labelKey: 'apiKey.scope.taskRead' },
    { value: 'task:write', labelKey: 'apiKey.scope.taskWrite' },
  ]},
  { groupKey: 'apiKey.scope.ics', items: [
    { value: 'ics:read', labelKey: 'apiKey.scope.icsRead' },
  ]},
  { groupKey: 'apiKey.scope.user', items: [
    { value: 'user:read', labelKey: 'apiKey.scope.userRead' },
  ]},
  { groupKey: 'apiKey.scope.admin', items: [
    { value: 'admin:read', labelKey: 'apiKey.scope.adminRead' },
    { value: 'admin:write', labelKey: 'apiKey.scope.adminWrite' },
  ]},
]

const SCOPE_LABELS: Record<string, string> = {}
for (const g of SCOPE_OPTIONS) {
  for (const s of g.items) {
    SCOPE_LABELS[s.value] = s.labelKey
  }
}

onMounted(async () => { loading.value = true; await load(); loading.value = false })

async function load() {
  try { keys.value = await api.get<ApiKey[]>('/users/me/api-keys') } catch { keys.value = [] }
}

function toggleScope(scope: string) {
  const current = newScopes.value ? newScopes.value.split(',').map(s => s.trim()).filter(Boolean) : []
  // Group shortcuts expand to individual scopes
  if (scope === 'read') {
    const readSet = ['family:read','floorplan:read','task:read','ics:read','user:read']
    if (readSet.every(s => current.includes(s))) {
      newScopes.value = current.filter(s => !readSet.includes(s)).join(', ')
    } else {
      const merged = [...new Set([...current, ...readSet])]
      newScopes.value = merged.join(', ')
    }
    return
  }
  if (scope === 'write') {
    const writeSet = ['family:read','family:write','floorplan:read','floorplan:write','task:read','task:write','ics:read','user:read']
    if (writeSet.every(s => current.includes(s))) {
      newScopes.value = current.filter(s => !writeSet.includes(s)).join(', ')
    } else {
      const merged = [...new Set([...current, ...writeSet])]
      newScopes.value = merged.join(', ')
    }
    return
  }
  if (scope === 'admin') {
    const adminSet = SCOPE_OPTIONS.flatMap(g => g.items.map(i => i.value))
    if (adminSet.every(s => current.includes(s))) {
      newScopes.value = ''
    } else {
      newScopes.value = adminSet.join(', ')
    }
    return
  }
  const idx = current.indexOf(scope)
  if (idx >= 0) current.splice(idx, 1)
  else current.push(scope)
  newScopes.value = current.join(', ')
}

async function create() {
  if (creating.value) return
  creating.value = true
  error.value = ''
  try {
    const scopes = newScopes.value ? newScopes.value.split(',').map(s => s.trim()).filter(Boolean) : undefined
    const body: any = { name: newName.value }
    if (scopes && scopes.length > 0) body.scopes = scopes
    if (newExpires.value) body.expires_at = new Date(newExpires.value).toISOString()

    const res = await api.post<{ api_key: ApiKey; message: string }>('/users/me/api-keys', body)
    createdKey.value = res.api_key.raw_key || null
    newName.value = ''; newScopes.value = ''; newExpires.value = ''
    await load()
  } catch (e: any) { error.value = e.message }
  finally { creating.value = false }
}

async function revoke(id: string) {
  try {
    await api.delete('/users/me/api-keys/' + id)
    await load()
  } catch (e: any) { error.value = e.message }
}

function copyKey() {
  if (createdKey.value) { navigator.clipboard.writeText(createdKey.value); createdKey.value = null }
}
</script>

<template>
  <div class="max-w-2xl mx-auto py-4 md:py-8 px-4">
    <h2 class="text-xl md:text-2xl font-bold mb-4 dark:text-gray-200">API Keys</h2>
    <p class="text-sm text-gray-400 mb-4">API Key 用于 CLI 或第三方工具访问系统。每个 Key 可设置权限范围和过期时间。</p>

    <p v-if="error" class="text-danger text-sm mb-3">{{ error }}</p>

    <LoadingSpinner v-if="loading" />
    <template v-else>

    <!-- Show just-created key -->
    <div v-if="createdKey" class="card mb-4 border border-yellow-300 dark:border-yellow-700 bg-yellow-50 dark:bg-yellow-900/20">
      <p class="text-sm font-medium text-yellow-800 dark:text-yellow-300 mb-2">⚠️ {{ t('apiKey.created') }}</p>
      <code class="block bg-white dark:bg-gray-800 px-3 py-2 rounded text-sm break-all select-all mb-2">{{ createdKey }}</code>
      <button class="btn-primary text-xs" @click="copyKey">{{ t('apiKey.copyAndClose') }}</button>
    </div>

    <button class="btn-primary mb-4" @click="showCreate = !showCreate">{{ showCreate ? t('apiKey.cancel') : '+ ' + t('apiKey.createBtn') }}</button>

    <!-- Create form -->
    <div v-if="showCreate" class="card mb-4 space-y-3">
      <input v-model="newName" class="input" :placeholder="t('apiKey.namePlaceholder')" />
      <div>
        <p class="text-xs text-gray-400 mb-2">{{ t('apiKey.scopeLabel') }}</p>
        <!-- Quick shortcuts -->
        <div class="flex flex-wrap gap-1.5 mb-3">
          <button v-for="grp in ['read','write','admin']" :key="grp"
            class="text-xs px-2.5 py-1 rounded border font-medium transition-colors"
            :class="newScopes.includes(grp === 'admin' ? 'admin:write' : grp === 'write' ? 'family:write' : 'family:read') ? 'bg-primary text-white border-primary' : 'border-gray-200 dark:border-gray-600 text-gray-500 dark:text-gray-400 hover:border-primary'"
            @click="toggleScope(grp)"
          >{{ grp === 'read' ? '📖 ' + t('apiKey.readonly') : grp === 'write' ? '✏️ ' + t('apiKey.readwrite') : '🔧 ' + t('apiKey.admin') }}</button>
        </div>
        <!-- Grouped scopes -->
        <div v-for="g in SCOPE_OPTIONS" :key="g.groupKey" class="mb-2">
          <p class="text-xs text-gray-500 mb-1">{{ t(g.groupKey) }}</p>
          <div class="flex flex-wrap gap-1">
            <button v-for="s in g.items" :key="s.value"
              class="text-xs px-2 py-1 rounded border transition-colors"
              :class="newScopes.includes(s.value) ? 'bg-primary text-white border-primary' : 'border-gray-200 dark:border-gray-600 text-gray-500 dark:text-gray-400 hover:border-primary'"
              @click="toggleScope(s.value)"
            >{{ t(s.labelKey) }}</button>
          </div>
        </div>
      </div>
      <div>
        <label class="text-xs text-gray-400 block mb-1">{{ t('apiKey.expiresLabel') }}</label>
        <input v-model="newExpires" type="datetime-local" class="input" />
      </div>
      <button class="btn-primary" :disabled="!newName || creating" @click="create">{{ creating ? '...' : t('apiKey.create') }}</button>
    </div>

    <!-- Key list -->
    <div v-for="k in keys" :key="k.id" class="card mb-2">
      <div class="flex items-center justify-between">
        <div>
          <span class="font-medium text-sm dark:text-gray-200">{{ k.name }}</span>
          <code class="text-xs text-gray-400 ml-2">{{ k.key_prefix }}...</code>
        </div>
        <button class="text-xs text-danger hover:underline" @click="revoke(k.id)">{{ t('apiKey.revoke') }}</button>
      </div>
      <div class="flex gap-3 mt-1.5 text-xs text-gray-400">
        <span v-if="k.scopes?.length">
          <span v-for="(s, i) in k.scopes" :key="s">
            <span v-if="i > 0">, </span>{{ SCOPE_LABELS[s] ? t(SCOPE_LABELS[s]) : s }}
          </span>
        </span>
        <span v-else class="text-primary font-medium">{{ t('apiKey.allScopes') }}</span>
        <span v-if="k.expires_at">{{ t('apiKey.expires') }}{{ new Date(k.expires_at).toLocaleString() }}</span>
        <span v-else>{{ t('apiKey.never') }}</span>
        <span v-if="k.last_used_at">{{ t('apiKey.lastUsed') }}{{ new Date(k.last_used_at).toLocaleString() }}</span>
      </div>
    </div>

    <div v-if="keys.length === 0 && !showCreate" class="text-center text-gray-400 py-8">{{ t('apiKey.empty') }}</div>
    </template>
  </div>
</template>
