import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('@/views/HomeView.vue'),
    },
    {
      path: '/feeds',
      name: 'feeds',
      component: () => import('@/views/FeedsView.vue'),
    },
    {
      path: '/tentang',
      name: 'tentang',
      component: () => import('@/views/TentangView.vue'),
    },
    {
      path: '/pakai-dayawarga',
      name: 'pakai-dayawarga',
      component: () => import('@/views/PakaiDayawargaView.vue'),
    },
    {
      path: '/belakang-layar',
      name: 'belakang-layar',
      component: () => import('@/views/BelakangLayarView.vue'),
    },
  ],
})

export default router
