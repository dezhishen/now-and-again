import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { api } from '@/api/client'
import { appLoadingText } from '@/composables/useAppLoading'
import i18n from '@/i18n'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', name: 'login', component: () => import('@/views/LoginView.vue') },
    { path: '/register', name: 'register', component: () => import('@/views/RegisterView.vue') },
    { path: '/', name: 'home', component: () => import('@/views/HomeView.vue'), meta: { requiresAuth: true } },
    { path: '/admin', name: 'admin', component: () => import('@/views/AdminView.vue'), meta: { requiresAuth: true, requiresAdmin: true } },
    { path: '/api-keys', name: 'api-keys', component: () => import('@/views/ApiKeyView.vue'), meta: { requiresAuth: true } },
    { path: '/profile', name: 'profile', component: () => import('@/views/ProfileView.vue'), meta: { requiresAuth: true } },
    { path: '/families', name: 'family-manage', component: () => import('@/views/FamilyManageView.vue'), meta: { requiresAuth: true } },
    {
      path: '/family', name: 'family',
      component: () => import('@/views/FamilyView.vue'), meta: { requiresAuth: true, requiresFamily: true },
      children: [
        { path: '', name: 'family-dashboard', component: () => import('@/views/family/DashboardView.vue') },
        { path: 'groups', name: 'family-groups', component: () => import('@/views/family/GroupListView.vue') },
        { path: 'members', name: 'family-members', component: () => import('@/views/family/MemberListView.vue') },
        { path: 'floor-plan', name: 'family-floor-plan', component: () => import('@/views/family/FloorPlanView.vue') },
        { path: 'tasks', name: 'family-tasks', component: () => import('@/views/family/TaskView.vue') },
        { path: 'ics', name: 'family-ics', component: () => import('@/views/family/IcsView.vue') },
        { path: 'calendar', name: 'family-calendar', component: () => import('@/views/family/CalendarView.vue') },
        { path: 'settings', name: 'family-settings', component: () => import('@/views/family/SettingsView.vue') },
        { path: 'templates', name: 'family-templates', component: () => import('@/views/family/TaskTemplateListView.vue') },
      ],
    },
    { path: '/:pathMatch(.*)*', name: 'not-found', component: () => import('@/views/NotFoundView.vue') },
    { path: '/calendar', name: 'calendar-full', component: () => import('@/views/family/CalendarView.vue'), meta: { requiresAuth: true, requiresFamily: true, fullscreen: true } },
  ],
})

router.beforeEach(async (to, _from, next) => {
  const auth = useAuthStore()
  const t = i18n.global.t

  // ── Silent token restore (only for protected routes) ─────
  if (to.meta.requiresAuth && !api.hasValidToken()) {
    appLoadingText.value = t('app.checkingSession')
    await auth.initSession()
  }
  // Lazy-load user profile (token valid but user lost on refresh).
  // fetchUser clears the token on 401 (stale token after db-reset).
  if (auth.isLoggedIn && !auth.user) {
    appLoadingText.value = t('app.fetchingUser')
    await auth.fetchUser()
  }

  // Lazy-load families list (needed for header display)
  if (auth.isLoggedIn && auth.families.length === 0) {
    appLoadingText.value = t('app.loadingFamilies')
    await auth.loadFamilies()
  }

  appLoadingText.value = ''

  // ── Auth guard ────────────────────────────────────────────
  if (to.meta.requiresAuth && !auth.isLoggedIn) {
    if (to.name === 'calendar-full' && to.query.key) return next()
    return next({ path: '/login', query: { redirect: to.fullPath } })
  }
  if (to.meta.requiresAdmin && !auth.isAdmin) return next('/')

  // ── Family guard ──────────────────────────────────────────
  if (to.meta.requiresFamily && !auth.activeFamilyId) {
    return next('/families')
  }

  // ── Home redirect: skip family selection if a family is already active ──
  if (to.name === 'home' && auth.isLoggedIn) {
    // Ensure user profile is loaded (contains default_family_id)
    if (!auth.user) await auth.fetchUser()

    if (auth.activeFamilyId) {
      // Local storage has a family → go straight in
      return next('/family')
    }
    if (auth.user?.default_family_id) {
      // User has a default family → activate and enter
      auth.switchFamily(auth.user.default_family_id)
      return next('/family')
    }
  }

  // Already logged in — don't show login/register
  if (auth.isLoggedIn && (to.name === 'login' || to.name === 'register')) {
    return next('/')
  }

  next()
})

export default router
