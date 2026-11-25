import type { RouteRecordRaw } from 'vue-router'
import type { Role } from '@/api/types'

declare module 'vue-router' {
  interface RouteMeta {
    title?: string
    icon?: string
    roles?: Role[]
  }
}

/** 业务页面路由，同时作为侧边栏菜单数据源。 */
export const appRoutes: RouteRecordRaw[] = [
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: () => import('@/views/Dashboard.vue'),
    meta: { title: '概览', icon: 'DataLine', roles: ['super', 'tenant_admin', 'operator', 'user'] },
  },
  {
    path: '/tenants',
    name: 'Tenants',
    component: () => import('@/views/Tenants.vue'),
    meta: { title: '租户管理', icon: 'OfficeBuilding', roles: ['super'] },
  },
  {
    path: '/users',
    name: 'Users',
    component: () => import('@/views/Users.vue'),
    meta: { title: '用户管理', icon: 'User', roles: ['super', 'tenant_admin', 'operator'] },
  },
  {
    path: '/stations',
    name: 'Stations',
    component: () => import('@/views/Stations.vue'),
    meta: { title: '场站管理', icon: 'Location', roles: ['super', 'tenant_admin', 'operator', 'user'] },
  },
  {
    path: '/piles',
    name: 'Piles',
    component: () => import('@/views/Piles.vue'),
    meta: { title: '桩与枪管理', icon: 'Connection', roles: ['super', 'tenant_admin', 'operator'] },
  },
  {
    path: '/wallets',
    name: 'Wallets',
    component: () => import('@/views/Wallets.vue'),
    meta: { title: '钱包管理', icon: 'Wallet', roles: ['super', 'tenant_admin', 'operator', 'user'] },
  },
  {
    path: '/ledgers',
    name: 'Ledgers',
    component: () => import('@/views/Ledgers.vue'),
    meta: { title: '资金流水', icon: 'Tickets', roles: ['super', 'tenant_admin', 'operator', 'user'] },
  },
  {
    path: '/reservations',
    name: 'Reservations',
    component: () => import('@/views/Reservations.vue'),
    meta: { title: '预约管理', icon: 'Calendar', roles: ['super', 'tenant_admin', 'operator', 'user'] },
  },
  {
    path: '/orders',
    name: 'Orders',
    component: () => import('@/views/Orders.vue'),
    meta: { title: '订单管理', icon: 'List', roles: ['super', 'tenant_admin', 'operator', 'user'] },
  },
  {
    path: '/smtp',
    name: 'SmtpConfig',
    component: () => import('@/views/SmtpConfig.vue'),
    meta: { title: 'SMTP 配置', icon: 'Message', roles: ['super', 'tenant_admin'] },
  },
  {
    path: '/notifications',
    name: 'Notifications',
    component: () => import('@/views/Notifications.vue'),
    meta: { title: '通知记录', icon: 'Bell', roles: ['super', 'tenant_admin', 'operator', 'user'] },
  },
]
