import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/setup', name: 'setup', component: () => import('@/views/SetupView.vue') },
    { path: '/login', name: 'login', component: () => import('@/views/LoginView.vue') },
    { path: '/register', name: 'register', component: () => import('@/views/RegisterView.vue') },
    { path: '/', name: 'home', component: () => import('@/views/HomeView.vue'), meta: { requiresAuth: true } },
    { path: '/admin', name: 'admin', component: () => import('@/views/AdminView.vue'), meta: { requiresAuth: true, requiresAdmin: true } },
    { path: '/api-keys', name: 'api-keys', component: () => import('@/views/ApiKeyView.vue'), meta: { requiresAuth: true } },
    {
      path: '/family/:familyId', name: 'family',
      component: () => import('@/views/FamilyView.vue'), meta: { requiresAuth: true },
      children: [
        { path: '', name: 'family-dashboard', component: () => import('@/views/family/DashboardView.vue') },
        { path: 'groups', name: 'family-groups', component: () => import('@/views/family/GroupListView.vue') },
        { path: 'members', name: 'family-members', component: () => import('@/views/family/MemberListView.vue') },
        { path: 'floor-plan', name: 'family-floor-plan', component: () => import('@/views/family/FloorPlanView.vue') },
        { path: 'tasks', name: 'family-tasks', component: () => import('@/views/family/TaskView.vue') },
        { path: 'ics', name: 'family-ics', component: () => import('@/views/family/IcsView.vue') },
        { path: 'settings', name: 'family-settings', component: () => import('@/views/family/SettingsView.vue') },
      ],
    },
    { path: '/:pathMatch(.*)*', name: 'not-found', component: () => import('@/views/NotFoundView.vue') },
  ],
})

router.beforeEach(async (to, _from, next) => {
  const auth = useAuthStore()

  if (auth.initialized === null) await auth.checkInit()
  if (auth.needsSetup && to.name !== 'setup') return next('/setup')
  if (!auth.needsSetup && to.name === 'setup') return next('/login')

  if (!auth.sessionChecked) await auth.initSession()

  if (to.meta.requiresAuth && !auth.isLoggedIn) return next('/login')
  if (to.meta.requiresAdmin && !auth.isAdmin) return next('/')
  if (auth.isLoggedIn && (to.name === 'login' || to.name === 'register')) return next('/')

  next()
})

export default router
