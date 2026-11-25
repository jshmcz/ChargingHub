<template>
  <el-container class="app-layout">
    <el-aside :class="['app-aside', { 'is-collapsed': isCollapsed }]">
      <div class="app-logo">{{ isCollapsed ? 'CH' : 'ChargingHub' }}</div>
      <el-menu
        class="app-menu"
        :default-active="route.path"
        :collapse="isCollapsed"
        :router="true"
        background-color="#001529"
        text-color="#bfcbd9"
        active-text-color="#409eff"
      >
        <el-menu-item v-for="item in menus" :key="item.path" :index="item.path">
          <el-icon><component :is="item.meta?.icon" /></el-icon>
          <template #title>{{ item.meta?.title }}</template>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="app-header">
        <div class="app-header-left">
          <el-button text @click="isCollapsed = !isCollapsed">
            <el-icon><component :is="isCollapsed ? 'Expand' : 'Fold'" /></el-icon>
          </el-button>
          <span>{{ route.meta.title || 'ChargingHub' }}</span>
        </div>
        <div class="app-header-right">
          <el-tag size="small" effect="dark">{{ roleLabel }}</el-tag>
          <span class="text-muted">{{ auth.user?.name || auth.user?.phone || '' }}</span>
          <el-button size="small" @click="handleLogout">退出</el-button>
        </div>
      </el-header>
      <el-main class="app-main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import { appRoutes } from '@/router/routes'
import type { Role } from '@/api/types'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const isCollapsed = ref(false)

function updateCollapse() {
  isCollapsed.value = window.innerWidth < 900
}

onMounted(() => {
  updateCollapse()
  window.addEventListener('resize', updateCollapse)
})

onUnmounted(() => {
  window.removeEventListener('resize', updateCollapse)
})

const roleLabel = computed(() => {
  const labels: Record<Role | '', string> = {
    super: '平台超管',
    tenant_admin: '租户管理员',
    operator: '操作员',
    user: '充电用户',
    '': '',
  }
  return labels[auth.role]
})

const menus = computed(() =>
  appRoutes.filter((item) => {
    const roles = item.meta?.roles
    return !roles || roles.includes(auth.role as Role)
  }),
)

async function handleLogout() {
  try {
    await ElMessageBox.confirm('确认退出登录？', '提示', { type: 'warning' })
  } catch {
    return
  }
  await auth.logout()
  router.push('/login')
}
</script>
