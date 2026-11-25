<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { Button, Cell, CellGroup, NavBar, Tag, showConfirmDialog } from 'vant'
import { authApi } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'
import { roleText } from '@/utils/format'

const router = useRouter()
const auth = useAuthStore()
const loggingOut = ref(false)

async function logout(): Promise<void> {
  try {
    await showConfirmDialog({
      title: '退出登录',
      message: '确认退出当前账号？',
      confirmButtonText: '退出',
    })
  } catch {
    return
  }

  loggingOut.value = true
  try {
    await authApi.logout()
  } catch {
    // 即使服务端登出失败，也清理本地会话
  } finally {
    auth.clearAuth()
    loggingOut.value = false
    await router.replace('/login')
  }
}
</script>

<template>
  <div class="page">
    <NavBar title="我的" />

    <div class="profile-card">
      <div class="avatar">{{ (auth.user?.name || 'U').slice(0, 1) }}</div>
      <div class="flex-1">
        <div class="profile-name">{{ auth.user?.name || '未登录用户' }}</div>
        <div class="muted">{{ auth.user?.phone || '--' }}</div>
      </div>
      <Tag type="primary" size="medium">{{ roleText(auth.role) }}</Tag>
    </div>

    <CellGroup inset>
      <Cell title="用户 ID" :value="auth.user?.id || '-'" />
      <Cell title="邮箱" :value="auth.user?.email || '-'" />
      <Cell title="租户 ID" :value="auth.tenantId || '-'" />
    </CellGroup>

    <div class="section-title">功能</div>
    <CellGroup inset>
      <Cell title="预约记录" icon="calendar-o" is-link to="/reservations" />
      <Cell title="通知记录" icon="bell" is-link to="/notifications" />
      <Cell v-if="auth.isManager" title="管理入口" icon="setting-o" is-link to="/admin" />
    </CellGroup>

    <div class="logout-wrap">
      <Button block round type="danger" plain :loading="loggingOut" @click="logout">
        退出登录
      </Button>
    </div>
  </div>
</template>

<style scoped>
.profile-card {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 12px;
  padding: 20px 16px;
  background: var(--ch-card);
  border-radius: var(--ch-radius);
}

.avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: var(--ch-primary);
  color: #fff;
  font-size: 20px;
  text-align: center;
  line-height: 48px;
}

.profile-name {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 2px;
}

.logout-wrap {
  padding: 24px 12px;
}
</style>
