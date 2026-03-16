import LoginView from '@/views/LoginView.vue'
import RoomManagementView from '@/views/RoomManagementView.vue'
import RoomDetailView from '@/views/RoomDetailView.vue'
import RoomListView from '@/views/RoomListView.vue'
import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/login', component: LoginView },
    { path: '/', component: RoomListView },
    {
      path: '/rooms/manage',
      component: RoomManagementView,
      meta: { requiresAuth: true }, // 需要登录才能访问
    },
    { path: '/rooms/:id', component: RoomDetailView, props: true }, // 第四个界面
  ],
})

router.beforeEach((to) => {
  if (!to.meta.requiresAuth) {
    return true
  }

  const token = localStorage.getItem('accesstoken')
  if (token) {
    return true
  }

  return {
    path: '/login',
    query: { redirect: to.fullPath, auth: 'missing' },
  }
})

export default router
