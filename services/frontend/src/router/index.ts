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
      path: '/feeds/:code',
      name: 'feed-detail',
      component: () => import('@/views/FeedDetailView.vue'),
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
    {
      path: '/share-preview',
      name: 'share-preview',
      component: () => import('@/views/SharePreview.vue'),
    },
  ],
})

export default router
