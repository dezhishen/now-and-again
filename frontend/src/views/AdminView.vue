<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api/client'
import type { User } from '@/types'

const users = ref<User[]>([])
const activeTab = ref<'users' | 'storage'>('users')
const settings = ref<Record<string, string>>({})
const saved = ref(false)
const error = ref('')

onMounted(async () => {
  try { users.value = await api.get<User[]>('/admin/users') } catch { /* */ }
  await loadSettings()
})

async function loadSettings() {
  try {
    const list = await api.get<{ Key: string; Value: string }[]>('/admin/settings')
    const map: Record<string, string> = {}
    for (const s of list) { map[s.Key] = s.Value }
    settings.value = map
  } catch { /* */ }
}

async function saveSettings() {
  error.value = ''
  saved.value = false
  try {
    await api.put('/admin/settings', settings.value)
    saved.value = true
    setTimeout(() => { saved.value = false }, 2000)
  } catch (e: any) { error.value = e.message }
}

const STORAGE_OPTIONS = [
  { value: 'local', label: '本地存储 (Local)' },
  { value: 's3', label: 'AWS S3（预留）', disabled: true },
  { value: 'oss', label: '阿里云 OSS（预留）', disabled: true },
  { value: 'minio', label: 'MinIO（预留）', disabled: true },
]
</script>

<template>
  <div>
    <h2 class="text-xl md:text-2xl font-bold mb-4 dark:text-gray-200">管理面板</h2>

    <!-- Tabs -->
    <div class="flex gap-1 mb-4 border-b dark:border-gray-700">
      <button class="px-4 py-2 text-sm font-medium border-b-2 transition-colors"
        :class="activeTab === 'users' ? 'border-primary text-primary' : 'border-transparent text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'"
        @click="activeTab = 'users'"
      >用户管理</button>
      <button class="px-4 py-2 text-sm font-medium border-b-2 transition-colors"
        :class="activeTab === 'storage' ? 'border-primary text-primary' : 'border-transparent text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'"
        @click="activeTab = 'storage'"
      >存储配置</button>
    </div>

    <!-- Users Tab -->
    <div v-if="activeTab === 'users'" class="card overflow-x-auto">
      <table class="w-full text-sm min-w-[500px]">
        <thead>
          <tr class="border-b dark:border-gray-700 text-left text-gray-500 dark:text-gray-400">
            <th class="py-2 px-3">显示名称</th><th class="py-2 px-3">邮箱</th><th class="py-2 px-3">角色</th><th class="py-2 px-3 hidden sm:table-cell">注册时间</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="u in users" :key="u.id" class="border-b dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700">
            <td class="py-2 px-3 font-medium dark:text-gray-200">{{ u.display_name }}</td>
            <td class="py-2 px-3 dark:text-gray-300">{{ u.email }}</td>
            <td class="py-2 px-3"><span v-if="u.roles.includes('admin')" class="text-primary font-medium">管理员</span><span v-else class="text-gray-400">成员</span></td>
            <td class="py-2 px-3 text-gray-400 hidden sm:table-cell">{{ u.created_at?.split('T')[0] }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Storage Tab -->
    <div v-if="activeTab === 'storage'">
      <p v-if="error" class="text-danger text-sm mb-3">{{ error }}</p>

      <div class="card mb-4 max-w-lg">
        <h3 class="font-medium mb-3 dark:text-gray-200">图片存储配置</h3>
        <p class="text-xs text-gray-400 mb-4">选择图片文件的存储后端。切换存储类型后，新上传的图片将使用新的存储后端，已有图片不受影响。</p>

        <label class="block text-sm text-gray-500 dark:text-gray-400 mb-1">存储类型</label>
        <select v-model="settings['storage.type']" class="input mb-4">
          <option v-for="opt in STORAGE_OPTIONS" :key="opt.value" :value="opt.value" :disabled="opt.disabled">{{ opt.label }}</option>
        </select>

        <button class="btn-primary text-sm" @click="saveSettings">
          {{ saved ? '已保存' : '保存配置' }}
        </button>
      </div>

      <div class="card max-w-lg">
        <h3 class="font-medium mb-2 dark:text-gray-200">当前状态</h3>
        <div class="text-sm text-gray-500 dark:text-gray-400 space-y-1">
          <p>存储类型：<code class="bg-gray-100 dark:bg-gray-700 px-1.5 py-0.5 rounded text-xs">{{ settings['storage.type'] || 'local' }}</code></p>
          <p v-if="settings['storage.type'] === 'local' || !settings['storage.type']" class="text-xs text-green-600 mt-2">✓ 本地存储已激活，图片保存在服务器的 uploads 目录下。</p>
        </div>
      </div>
    </div>
  </div>
</template>
