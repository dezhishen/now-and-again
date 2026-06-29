<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { api } from '@/api/client'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import type { Family } from '@/types'

const { t } = useI18n()
const router = useRouter()
const auth = useAuthStore()
const families = ref<Family[]>([])
const loading = ref(true)
const showCreate = ref(false)
const familyName = ref('')
const inviteCode = ref('')
const error = ref('')

function favId(): string | null {
  return auth.user?.default_family_id || null
}

async function toggleFav(familyId: string) {
  const newId = favId() === familyId ? null : familyId
  try {
    await api.put('/users/me', { default_family_id: newId })
    if (auth.user) auth.user.default_family_id = newId || undefined
  } catch { /* */ }
}

const sortedFamilies = computed(() => {
  return [...families.value].sort((a, b) => {
    if (a.id === favId()) return -1
    if (b.id === favId()) return 1
    return 0
  })
})

onMounted(async () => {
  loading.value = true
  try { families.value = await api.get<Family[]>('/users/me/families') } catch { /* */ }
  loading.value = false
})

const hasCreatedFamily = computed(() =>
  families.value.some(f => f.created_by === auth.user?.id)
)

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
  <div class="max-w-3xl mx-auto py-4 md:py-8 px-4">
    <div class="mb-8">
      <div class="flex items-center justify-between mb-3">
        <h2 class="text-lg md:text-xl font-bold dark:text-gray-200">{{ t('home.heading') }}</h2>
        <button v-if="!hasCreatedFamily" class="btn-primary" @click="showCreate = !showCreate">{{ showCreate ? t('home.cancel') : '+ ' + t('home.createFamily') }}</button>
      </div>

      <LoadingSpinner v-if="loading" />
      <div v-if="!loading">
      <div v-if="showCreate" class="card mb-4 flex flex-col sm:flex-row gap-2">
        <input v-model="familyName" class="input flex-1" :placeholder="t('home.familyName')" @keyup.enter="createFamily" />
        <button class="btn-primary" @click="createFamily">{{ t('home.create') }}</button>
      </div>

      <div class="card mb-4 flex flex-col sm:flex-row gap-2">
        <input v-model="inviteCode" class="input flex-1" :placeholder="t('home.joinByCode')" @keyup.enter="joinFamily" />
        <button class="btn-primary" @click="joinFamily">{{ t('home.join') }}</button>
      </div>

      <p v-if="error" class="text-danger text-sm mb-2">{{ error }}</p>

      <div v-if="families.length === 0" class="text-center text-gray-400 dark:text-gray-500 py-8">{{ t('home.noFamily') }}</div>

      <!-- Family cards grid -->
      <div class="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-3">
        <div
          v-for="f in sortedFamilies"
          :key="f.id"
          class="card cursor-pointer hover:shadow-lg transition-shadow dark:hover:bg-gray-700 relative overflow-hidden group"
          @click="auth.switchFamily(f.id); router.push('/family')"
        >
          <span v-if="f.created_by === auth.user?.id" class="absolute top-2 left-2 z-10 px-2 py-0.5 rounded text-xs bg-primary/90 text-white font-medium">Owner</span>

          <!-- Favorite star -->
          <button
            class="absolute top-2 right-2 z-10 w-7 h-7 flex items-center justify-center rounded-full bg-white/80 dark:bg-gray-900/80 hover:bg-yellow-100 transition-colors"
            @click.stop="toggleFav(f.id)"
            :title="favId() === f.id ? t('home.unfavorite') : t('home.favorite')"
          >
            <span v-if="favId() === f.id" class="text-yellow-500 text-lg">★</span>
            <span v-else class="text-gray-300 dark:text-gray-600 text-lg group-hover:text-yellow-400 transition-colors">☆</span>
          </button>

          <!-- Thumbnail -->
          <div v-if="f.thumbnail_url" class="mb-3 -mx-4 -mt-4 overflow-hidden rounded-t-lg aspect-video bg-gray-200 dark:bg-gray-700">
            <img :src="f.thumbnail_url" class="w-full h-full object-cover" />
          </div>
          <div v-else class="mb-3 -mx-4 -mt-4 aspect-video bg-gradient-to-br from-primary/10 to-primary/5 dark:from-primary/20 dark:to-gray-800 flex items-center justify-center rounded-t-lg">
            <span class="text-4xl opacity-30">{{ f.name[0] }}</span>
          </div>

          <!-- Info -->
          <div class="flex items-center justify-between">
            <span class="font-medium dark:text-gray-200">{{ f.name }}</span>
            <span class="text-xs text-gray-400">{{ t('home.inviteCodePrefix') }}{{ f.invite_code }}</span>
          </div>
        </div>
      </div>
    </div>
    </div>
  </div>
</template>
