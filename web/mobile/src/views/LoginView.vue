<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Button, Cell, CellGroup, Field, NavBar, showToast } from 'vant'
import { authApi } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'
import { getDeviceId } from '@/utils/device'
import { roleText } from '@/utils/format'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const username = ref('')
const password = ref('')
const loading = ref(false)
const deviceId = getDeviceId()

async function handleLogin(): Promise<void> {
  if (!username.value.trim()) {
    showToast('请输入账号')
    return
  }
  if (!password.value) {
    showToast('请输入密码')
    return
  }

  loading.value = true
  try {
    const result = await authApi.login({
      username: username.value.trim(),
      password: password.value,
    })
    auth.setAuth(result)
    showToast({ type: 'success', message: `欢迎回来，${roleText(result.role)}` })
    const redirect = route.query.redirect
    const target =
      typeof redirect === 'string' && redirect.startsWith('/') ? redirect : '/home'
    await router.replace(target)
  } catch {
    // request.ts 已统一提示错误
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="page--plain login-page">
    <NavBar title="登录" />

    <div class="login-brand">
      <div class="login-logo">CH</div>
      <div class="login-title">ChargingHub</div>
      <div class="muted">多租户智能充电管理平台</div>
    </div>

    <div class="card card--pad">
      <Field
        v-model="username"
        label="账号"
        placeholder="请输入用户名"
        clearable
      />
      <Field
        v-model="password"
        label="密码"
        type="password"
        placeholder="请输入密码"
      />
      <div class="login-actions">
        <Button type="primary" block round :loading="loading" @click="handleLogin">
          登 录
        </Button>
      </div>
    </div>

    <CellGroup inset class="login-tip">
      <Cell title="设备标识" :value="deviceId" label="设备 ID 保存在本机，换设备登录会挤下线旧会话" />
    </CellGroup>
  </div>
</template>

<style scoped>
.login-brand {
  padding: 48px 0 28px;
  text-align: center;
}

.login-logo {
  width: 64px;
  height: 64px;
  margin: 0 auto 12px;
  border-radius: 18px;
  background: var(--ch-primary);
  color: #fff;
  font-size: 22px;
  font-weight: 700;
  line-height: 64px;
  letter-spacing: 1px;
}

.login-title {
  font-size: 20px;
  font-weight: 600;
  margin-bottom: 4px;
}

.login-actions {
  margin-top: 20px;
}

.login-tip {
  margin-top: 12px;
}
</style>
