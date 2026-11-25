import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import type { LoginResult, MeResult, Role, UserInfo } from '@/api/types'
import { STORAGE_KEYS } from '@/utils/constants'

function readUser(): UserInfo | null {
  const raw = localStorage.getItem(STORAGE_KEYS.user)
  if (!raw) return null
  try {
    return JSON.parse(raw) as UserInfo
  } catch {
    return null
  }
}

const MANAGE_ROLES: Role[] = ['super', 'tenant_admin', 'operator']

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>(localStorage.getItem(STORAGE_KEYS.token) || '')
  const role = ref<Role>((localStorage.getItem(STORAGE_KEYS.role) as Role | null) || 'user')
  const tenantId = ref<string | null>(localStorage.getItem(STORAGE_KEYS.tenantId))
  const user = ref<UserInfo | null>(readUser())

  const isLogged = computed(() => !!token.value)
  const isManager = computed(() => MANAGE_ROLES.includes(role.value))
  const canRefund = computed(
    () => role.value === 'super' || role.value === 'tenant_admin',
  )

  function setAuth(payload: LoginResult): void {
    token.value = payload.token
    role.value = payload.role
    tenantId.value = payload.tenantId
    user.value = payload.user
    localStorage.setItem(STORAGE_KEYS.token, payload.token)
    localStorage.setItem(STORAGE_KEYS.role, payload.role)
    if (payload.tenantId) {
      localStorage.setItem(STORAGE_KEYS.tenantId, payload.tenantId)
    } else {
      localStorage.removeItem(STORAGE_KEYS.tenantId)
    }
    localStorage.setItem(STORAGE_KEYS.user, JSON.stringify(payload.user))
  }

  function setMe(payload: MeResult): void {
    role.value = payload.role
    tenantId.value = payload.tenantId
    user.value = payload.user
    localStorage.setItem(STORAGE_KEYS.role, payload.role)
    if (payload.tenantId) {
      localStorage.setItem(STORAGE_KEYS.tenantId, payload.tenantId)
    } else {
      localStorage.removeItem(STORAGE_KEYS.tenantId)
    }
    localStorage.setItem(STORAGE_KEYS.user, JSON.stringify(payload.user))
  }

  function clearAuth(): void {
    token.value = ''
    role.value = 'user'
    tenantId.value = null
    user.value = null
    localStorage.removeItem(STORAGE_KEYS.token)
    localStorage.removeItem(STORAGE_KEYS.role)
    localStorage.removeItem(STORAGE_KEYS.tenantId)
    localStorage.removeItem(STORAGE_KEYS.user)
  }

  return {
    token,
    role,
    tenantId,
    user,
    isLogged,
    isManager,
    canRefund,
    setAuth,
    setMe,
    clearAuth,
  }
})
