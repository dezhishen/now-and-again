<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { api } from '@/api/client'
import type { Family } from '@/types'

const { t } = useI18n()
const router = useRouter()
const auth = useAuthStore()
const families = ref<Family[]>([])
const showCreate = ref(false)
const familyName = ref('')
const inviteCode = ref('')
const error = ref('')

onMounted(async () => {
  try { families.value = await api.get<Family[]>('/users/me/families') } catch { /* */ }
})

async function createFamily() {
  error.value = ''
  try {
    const f = await api.post<Family>('/families', { name: familyName.value })
    families.value.push(f)
    familyName.value = ''
    showCreate.value = false
  } catch (e: any) { error.value = e.message }
}

async function joinFamily() {
  error.value = ''
  try {
    await api.post('/families/join', { invite_code: inviteCode.value })
    inviteCode.value = ''
    families.value = await api.get<Family[]>('/users/me/families')
  } catch (e: any) { error.value = e.message }
}
</script>

<template>
  <div class="max-w-3xl mx-auto py-8 px-4">
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-3xl font-bold text-primary">{{ t('app.title') }}</h1>
      <div class="flex gap-2">
        <button v-if="auth.user?.is_admin" class="btn-primary" @click="router.push('/admin')">管理面板</button>
        <button class="btn-primary" @click="auth.logout(); router.push('/login')">退出登录</button>
      </div>
    </div>

    <div class="mb-8">
      <div class="flex items-center justify-between mb-3">
        <h2 class="text-xl font-bold">我的家庭</h2>
        <button class="btn-primary text-sm" @click="showCreate = !showCreate">{{ showCreate ? '取消' : '+ 创建家庭' }}</button>
      </div>

      <div v-if="showCreate" class="card mb-4 flex gap-2">
        <input v-model="familyName" class="input flex-1" placeholder="家庭名称" @keyup.enter="createFamily" />
        <button class="btn-primary" @click="createFamily">创建</button>
      </div>

      <div class="card mb-3 flex gap-2">
        <input v-model="inviteCode" class="input flex-1" placeholder="输入邀请码加入家庭" @keyup.enter="joinFamily" />
        <button class="btn-primary" @click="joinFamily">加入</button>
      </div>

      <p v-if="error" class="text-danger text-sm mb-2">{{ error }}</p>

      <div v-if="families.length === 0" class="text-center text-muted py-8">还没有家庭，创建一个或通过邀请码加入</div>

      <div v-for="f in families" :key="f.id" class="card mb-2 cursor-pointer hover:shadow-lg transition-shadow" @click="router.push(`/family/${f.id}`)">
        <div class="flex items-center justify-between">
          <span class="font-medium">{{ f.name }}</span>
          <span class="text-xs text-muted">邀请码: {{ f.invite_code }}</span>
        </div>
      </div>
    </div>
  </div>
</template>
