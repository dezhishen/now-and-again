<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from '@/i18n'
import type { I18nKey } from '@/i18n'
import { useAuthStore } from '@/stores/auth'
import { useDark } from '@/composables/useDark'

const { t, locale } = useI18n()
const localeCode = computed(() => locale.value)
const router = useRouter()
const auth = useAuthStore()
const { cycle: cycleTheme, label: themeLabel, title: themeTitle } = useDark()

const showUserMenu = ref(false)
const showLangMenu = ref(false)

const currentFamilyName = computed(() => {
  if (!auth.activeFamilyId) return ''
  const f = auth.families.find(f => f.id === auth.activeFamilyId)
  return f?.name || ''
})

function switchLang(lang: string) {
  locale.value = lang
  localStorage.setItem('na_lang', lang)
  showLangMenu.value = false
}

async function logout() {
  showUserMenu.value = false
  await auth.logout()
  window.location.href = '/login'
}

const LANGS: { code: string; key: I18nKey; flag: string }[] = [
  { code: 'zh-CN', key: 'lang.zhCN', flag: '🇨🇳' },
  { code: 'en', key: 'lang.en', flag: '🇺🇸' },
]

// Close menus on outside click
function onWindowClick() { showUserMenu.value = false; showLangMenu.value = false }
window.addEventListener('click', onWindowClick)
</script>

<template>
  <header class="sticky top-0 z-40 bg-white/90 dark:bg-gray-800/90 backdrop-blur border-b dark:border-gray-700">
    <div class="px-4 h-12 flex items-center justify-between">
      <!-- Left: Logo + Family -->
      <div class="flex items-center gap-3">
        <button class="flex items-center gap-2 font-bold text-primary hover:opacity-80 transition-opacity" @click="router.push('/')">
          <span class="text-lg">🏠</span>
          <span class="hidden sm:inline">{{ t('app.title') }}</span>
        </button>
        <span v-if="currentFamilyName" class="hidden sm:flex items-center gap-1 text-sm text-gray-400">
          <span class="text-gray-300">/</span>
          <span class="text-gray-700 dark:text-gray-300 font-medium truncate max-w-[160px]">{{ currentFamilyName }}</span>
        </span>
      </div>

      <!-- Right: Actions -->
      <div class="flex items-center gap-1">
        <!-- Language -->
        <div class="relative" @click.stop>
          <button class="w-8 h-8 rounded-lg flex items-center justify-center text-sm hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors" @click="showLangMenu = !showLangMenu" :title="t('lang.switch')">
            {{ localeCode === 'zh-CN' ? '🇨🇳' : '🇺🇸' }}
          </button>
          <div v-if="showLangMenu" class="absolute right-0 top-full mt-1 bg-white dark:bg-gray-800 rounded-lg shadow-lg border dark:border-gray-700 py-1 min-w-[120px]">
            <button v-for="l in LANGS" :key="l.code" class="flex items-center gap-2 px-3 py-2 text-sm w-full text-left hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
              :class="localeCode === l.code ? 'text-primary font-medium' : 'dark:text-gray-300'"
              @click="switchLang(l.code)"
            >{{ l.flag }} {{ t(l.key) }}</button>
          </div>
        </div>

        <!-- Dark mode -->
        <button class="w-8 h-8 rounded-lg flex items-center justify-center text-sm hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors" @click="cycleTheme" :title="themeTitle">
          {{ themeLabel }}
        </button>

        <!-- User -->
        <div v-if="auth.isLoggedIn" class="relative" @click.stop>
          <button class="flex items-center gap-1.5 h-8 px-2 rounded-lg text-sm hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors" @click="showUserMenu = !showUserMenu">
            <div class="w-6 h-6 rounded-full bg-primary/20 flex items-center justify-center text-primary font-bold text-xs">
              {{ auth.user?.display_name?.[0]?.toUpperCase() || '?' }}
            </div>
            <span class="hidden sm:inline max-w-[80px] truncate dark:text-gray-200">{{ auth.user?.display_name }}</span>
          </button>
          <div v-if="showUserMenu" class="absolute right-0 top-full mt-1 bg-white dark:bg-gray-800 rounded-lg shadow-lg border dark:border-gray-700 py-1 min-w-[140px]">
            <div class="px-3 py-2 border-b dark:border-gray-700">
              <p class="text-sm font-medium dark:text-gray-200 truncate">{{ auth.user?.display_name }}</p>
              <p class="text-xs text-gray-400 truncate">{{ auth.user?.email }}</p>
              <p class="text-xs text-gray-400 mt-0.5">
                <span v-if="auth.isAdmin" class="text-primary">{{ t('user.admin') }}</span>
                <span v-else>{{ t('user.member') }}</span>
              </p>
            </div>
            <button class="flex items-center gap-2 px-3 py-2 text-sm w-full text-left hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors dark:text-gray-300" @click="router.push('/profile'); showUserMenu = false">👤 {{ t('user.profile') }}</button>
            <button class="flex items-center gap-2 px-3 py-2 text-sm w-full text-left hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors dark:text-gray-300" @click="router.push('/families'); showUserMenu = false">🏠 {{ t('user.familyManage') }}</button>
            <button class="flex items-center gap-2 px-3 py-2 text-sm w-full text-left hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors dark:text-gray-300" @click="router.push('/api-keys'); showUserMenu = false">🔑 API Keys</button>
            <button v-if="auth.isAdmin" class="flex items-center gap-2 px-3 py-2 text-sm w-full text-left hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors dark:text-gray-300" @click="router.push('/admin'); showUserMenu = false">⚙️ {{ t('user.adminPanel') }}</button>
            <button class="flex items-center gap-2 px-3 py-2 text-sm w-full text-left text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors" @click="logout">🚪 {{ t('user.logout') }}</button>
          </div>
        </div>
      </div>
    </div>
  </header>
</template>
