<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useDark } from '@/composables/useDark'

const { t, locale } = useI18n()
const router = useRouter()
const auth = useAuthStore()
const { isDark, toggle: toggleDark } = useDark()

const showUserMenu = ref(false)
const showLangMenu = ref(false)

function switchLang(lang: string) {
  locale.value = lang
  localStorage.setItem('na_lang', lang)
  showLangMenu.value = false
}

function logout() {
  showUserMenu.value = false
  auth.logout()
  router.push('/login')
}

const LANGS = [
  { code: 'zh-CN', label: '中文', flag: '🇨🇳' },
  { code: 'en', label: 'English', flag: '🇺🇸' },
]

// Close menus on outside click
function onWindowClick() { showUserMenu.value = false; showLangMenu.value = false }
window.addEventListener('click', onWindowClick)
</script>

<template>
  <header class="sticky top-0 z-40 bg-white/90 dark:bg-gray-800/90 backdrop-blur border-b dark:border-gray-700">
    <div class="px-4 h-12 flex items-center justify-between">
      <!-- Left: Logo -->
      <button class="flex items-center gap-2 font-bold text-primary hover:opacity-80 transition-opacity" @click="router.push('/')">
        <span class="text-lg">🏠</span>
        <span class="hidden sm:inline">{{ t('app.title') }}</span>
      </button>

      <!-- Right: Actions -->
      <div class="flex items-center gap-1">
        <!-- Language -->
        <div class="relative" @click.stop>
          <button class="w-8 h-8 rounded-lg flex items-center justify-center text-sm hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors" @click="showLangMenu = !showLangMenu" title="切换语言">
            {{ locale === 'zh-CN' ? '🇨🇳' : '🇺🇸' }}
          </button>
          <div v-if="showLangMenu" class="absolute right-0 top-full mt-1 bg-white dark:bg-gray-800 rounded-lg shadow-lg border dark:border-gray-700 py-1 min-w-[120px]">
            <button v-for="l in LANGS" :key="l.code" class="flex items-center gap-2 px-3 py-2 text-sm w-full text-left hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
              :class="locale === l.code ? 'text-primary font-medium' : 'dark:text-gray-300'"
              @click="switchLang(l.code)"
            >{{ l.flag }} {{ l.label }}</button>
          </div>
        </div>

        <!-- Dark mode -->
        <button class="w-8 h-8 rounded-lg flex items-center justify-center text-sm hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors" @click="toggleDark" :title="isDark ? '切换亮色' : '切换暗色'">
          {{ isDark ? '☀️' : '🌙' }}
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
                <span v-if="auth.isAdmin" class="text-primary">管理员</span>
                <span v-else>成员</span>
              </p>
            </div>
            <button class="flex items-center gap-2 px-3 py-2 text-sm w-full text-left hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors dark:text-gray-300" @click="router.push('/api-keys'); showUserMenu = false">🔑 API Keys</button>
            <button v-if="auth.isAdmin" class="flex items-center gap-2 px-3 py-2 text-sm w-full text-left hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors dark:text-gray-300" @click="router.push('/admin'); showUserMenu = false">⚙️ 管理面板</button>
            <button class="flex items-center gap-2 px-3 py-2 text-sm w-full text-left text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors" @click="logout">🚪 退出登录</button>
          </div>
        </div>
      </div>
    </div>
  </header>
</template>
