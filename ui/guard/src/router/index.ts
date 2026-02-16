import type { RouteRecordRaw } from 'vue-router'
import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import ConsentView from '../views/ConsentView.vue'

const router = createRouter({
  // 使用 Vite 的 base 配置作为路由 base
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      redirect: to => ({ path: '/login', query: to.query })
    },
    {
      path: '/login',
      name: 'login',
      component: LoginView
    },
    {
      path: '/consent',
      name: 'consent',
      component: ConsentView
    }
  ]
})

export default router
