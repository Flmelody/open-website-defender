import { createRouter, createWebHistory } from 'vue-router'
import { getAppConfig } from '@/utils/config'
import Layout from '@/views/Layout.vue'
import LoginView from '@/views/LoginView.vue'
import DashboardView from '@/views/DashboardView.vue'
import UserView from '@/views/UserView.vue'
import IpWhiteListView from '@/views/IpWhiteListView.vue'
import IpBlackListView from '@/views/IpBlackListView.vue'
import WafRulesView from '@/views/WafRulesView.vue'
import AccessLogView from '@/views/AccessLogView.vue'
import GeoBlockView from '@/views/GeoBlockView.vue'

const config = getAppConfig()
// Clean up double slashes if any
const base = `${config.rootPath}${config.adminPath}`.replace(/\/+/g, '/')

const router = createRouter({
  history: createWebHistory(base),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: LoginView,
      meta: { guest: true }
    },
    {
      path: '/',
      component: Layout,
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          redirect: 'dashboard'
        },
        {
          path: 'dashboard',
          name: 'dashboard',
          component: DashboardView
        },
        {
          path: 'users',
          name: 'users',
          component: UserView
        },
        {
          path: 'ip-white-list',
          name: 'ip-white-list',
          component: IpWhiteListView
        },
        {
          path: 'ip-black-list',
          name: 'ip-black-list',
          component: IpBlackListView
        },
        {
          path: 'waf-rules',
          name: 'waf-rules',
          component: WafRulesView
        },
        {
          path: 'access-logs',
          name: 'access-logs',
          component: AccessLogView
        },
        {
          path: 'geo-block',
          name: 'geo-block',
          component: GeoBlockView
        }
      ]
    }
  ]
})

router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  if (to.meta.requiresAuth && !token) {
    next({ name: 'login' })
  } else if (to.meta.guest && token) {
    next({ name: 'dashboard' })
  } else {
    next()
  }
})

export default router

