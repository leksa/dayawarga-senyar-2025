import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/LoginView.vue'),
    meta: { requiresAuth: false, layout: 'blank' },
  },
  {
    path: '/callback',
    name: 'callback',
    component: () => import('@/views/CallbackView.vue'),
    meta: { requiresAuth: false, layout: 'blank' },
  },
  {
    path: '/',
    name: 'dashboard',
    component: () => import('@/views/DashboardView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/organizations',
    name: 'organizations',
    component: () => import('@/views/OrganizationsView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/organizations/:id',
    name: 'organization-detail',
    component: () => import('@/views/OrganizationDetailView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/groups',
    name: 'groups',
    component: () => import('@/views/GroupsView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/groups/:id',
    name: 'group-detail',
    component: () => import('@/views/GroupDetailView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/relawan',
    name: 'relawan',
    component: () => import('@/views/RelawanView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/relawan/:id',
    name: 'relawan-detail',
    component: () => import('@/views/RelawanDetailView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/project-requests',
    name: 'project-requests',
    component: () => import('@/views/ProjectRequestsView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/users',
    name: 'users',
    component: () => import('@/views/UsersView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: () => import('@/views/NotFoundView.vue'),
    meta: { requiresAuth: false, layout: 'blank' },
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

// Navigation guard for authentication
router.beforeEach((to, _from, next) => {
  // Check for OIDC user in localStorage (oidc-client-ts default storage)
  // Normalize authority by removing trailing slash for consistent key matching
  const authority = (import.meta.env.VITE_OIDC_AUTHORITY || '').replace(/\/$/, '')
  const clientId = import.meta.env.VITE_OIDC_CLIENT_ID
  const oidcStorageKey = `oidc.user:${authority}:${clientId}`
  let storedUserStr = localStorage.getItem(oidcStorageKey)

  // Also try with trailing slash if not found
  if (!storedUserStr) {
    const altKey = `oidc.user:${authority}/:${clientId}`
    storedUserStr = localStorage.getItem(altKey)
  }

  // Check if user exists and token is not expired
  let isAuthenticated = false
  if (storedUserStr) {
    try {
      const storedUser = JSON.parse(storedUserStr)
      const now = Math.floor(Date.now() / 1000)
      isAuthenticated = storedUser.expires_at > now
    } catch {
      isAuthenticated = false
    }
  }

  if (to.meta.requiresAuth && !isAuthenticated) {
    next({ name: 'login' })
  } else if (to.name === 'login' && isAuthenticated) {
    next({ name: 'dashboard' })
  } else {
    next()
  }
})

export default router
