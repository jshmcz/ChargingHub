import { createRouter, createWebHashHistory, type RouteRecordRaw } from 'vue-router'
import type { Role } from '@/api/types'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  { path: '/', redirect: '/home' },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/LoginView.vue'),
    meta: { title: '登录', public: true },
  },
  {
    path: '/home',
    name: 'Home',
    component: () => import('@/views/HomeView.vue'),
    meta: { title: '附近场站', showTabbar: true },
  },
  {
    path: '/station/:id',
    name: 'StationDetail',
    component: () => import('@/views/StationDetailView.vue'),
    meta: { title: '场站详情' },
  },
  {
    path: '/wallet',
    name: 'MyWallet',
    component: () => import('@/views/MyWalletView.vue'),
    meta: { title: '我的钱包', showTabbar: true },
  },
  {
    path: '/orders',
    name: 'Orders',
    component: () => import('@/views/OrdersView.vue'),
    meta: { title: '充电订单', showTabbar: true },
  },
  {
    path: '/profile',
    name: 'Profile',
    component: () => import('@/views/ProfileView.vue'),
    meta: { title: '我的', showTabbar: true },
  },
  {
    path: '/reservations',
    name: 'Reservations',
    component: () => import('@/views/ReservationsView.vue'),
    meta: { title: '我的预约' },
  },
  {
    path: '/charge',
    name: 'Charge',
    component: () => import('@/views/ChargeView.vue'),
    meta: { title: '充电控制' },
  },
  {
    path: '/notifications',
    name: 'Notifications',
    component: () => import('@/views/NotificationsView.vue'),
    meta: { title: '通知记录' },
  },
  {
    path: '/admin',
    name: 'AdminHome',
    component: () => import('@/views/admin/AdminHomeView.vue'),
    meta: { title: '管理后台', roles: ['super', 'tenant_admin', 'operator'] },
  },
  {
    path: '/admin/users',
    name: 'AdminUsers',
    component: () => import('@/views/admin/AdminUsersView.vue'),
    meta: { title: '用户管理', roles: ['super', 'tenant_admin', 'operator'] },
  },
  {
    path: '/admin/stations',
    name: 'AdminStations',
    component: () => import('@/views/admin/AdminStationsView.vue'),
    meta: { title: '场站管理', roles: ['super', 'tenant_admin', 'operator'] },
  },
  {
    path: '/admin/wallet',
    name: 'AdminWallet',
    component: () => import('@/views/admin/AdminWalletView.vue'),
    meta: { title: '充值 / 退款', roles: ['super', 'tenant_admin', 'operator'] },
  },
  { path: '/:pathMatch(.*)*', redirect: '/home' },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
  scrollBehavior: () => ({ top: 0 }),
})

router.beforeEach((to) => {
  const title = to.meta.title as string | undefined
  document.title = title ? `${title} - ChargingHub` : 'ChargingHub'

  const auth = useAuthStore()
  if (!to.meta.public && !auth.isLogged) {
    return {
      path: '/login',
      query: to.fullPath === '/home' ? {} : { redirect: to.fullPath },
    }
  }

  const roles = to.meta.roles as Role[] | undefined
  if (roles && !roles.includes(auth.role)) {
    return { path: '/home' }
  }

  return true
})

export default router
