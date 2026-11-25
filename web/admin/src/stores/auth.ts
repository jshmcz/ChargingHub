import { defineStore } from 'pinia'
import { authApi } from '@/api/auth'
import type { LoginPayload } from '@/api/auth'
import { TOKEN_KEY } from '@/api/request'
import type { LoginResult, LoginUser, MeResult, Role } from '@/api/types'
import { getOrCreateDeviceId } from '@/utils/uuid'

const PROFILE_KEY = 'charginghub_profile'

interface Profile {
  role: Role | ''
  tenantId: number | null
  user: LoginUser | null
}

function loadProfile(): Profile {
  const empty: Profile = { role: '', tenantId: null, user: null }
  try {
    const raw = localStorage.getItem(PROFILE_KEY)
    if (!raw) return empty
    const parsed = JSON.parse(raw) as Partial<Profile>
    return {
      role: parsed.role || '',
      tenantId: parsed.tenantId ?? null,
      user: parsed.user ?? null,
    }
  } catch {
    return empty
  }
}

const savedProfile = loadProfile()

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem(TOKEN_KEY) || '',
    role: savedProfile.role as Role | '',
    tenantId: savedProfile.tenantId as number | null,
    user: savedProfile.user as LoginUser | null,
    deviceId: getOrCreateDeviceId(),
  }),
  getters: {
    isLoggedIn: (state) => !!state.token,
    isSuper: (state) => state.role === 'super',
    isTenantAdmin: (state) => state.role === 'tenant_admin',
    isOperator: (state) => state.role === 'operator',
    isEndUser: (state) => state.role === 'user',
    isStaff: (state) => ['super', 'tenant_admin', 'operator'].includes(state.role),
    canManageTenants: (state) => state.role === 'super',
    /** operator 无退款权限；super 拥有一切权限。 */
    canRefund: (state) => state.role === 'super' || state.role === 'tenant_admin',
    /** operator 无 SMTP 配置权限。 */
    canConfigSmtp: (state) => state.role === 'super' || state.role === 'tenant_admin',
  },
  actions: {
    setSession(result: LoginResult) {
      this.token = result.token
      this.role = result.role
      this.tenantId = result.tenantId
      this.user = result.user
      localStorage.setItem(TOKEN_KEY, result.token)
      this.saveProfile()
    },
    setProfile(profile: MeResult) {
      this.role = profile.role
      this.tenantId = profile.tenantId
      this.user = profile.user
      this.saveProfile()
    },
    saveProfile() {
      localStorage.setItem(
        PROFILE_KEY,
        JSON.stringify({ role: this.role, tenantId: this.tenantId, user: this.user }),
      )
    },
    async login(payload: { username: string; password: string; tenantId?: number | string }) {
      const body: LoginPayload = {
        username: payload.username,
        password: payload.password,
        deviceId: this.deviceId,
      }
      if (payload.tenantId !== undefined && payload.tenantId !== '') {
        body.tenantId = payload.tenantId
      }
      const result = await authApi.login(body)
      this.setSession(result)
      return result
    },
    async fetchMe() {
      const profile = await authApi.me()
      this.setProfile(profile)
    },
    async logout() {
      try {
        await authApi.logout()
      } catch {
        // 退出时忽略网络/鉴权异常，本地会话始终清理。
      }
      this.clearSession()
    },
    clearSession() {
      this.token = ''
      this.role = ''
      this.tenantId = null
      this.user = null
      localStorage.removeItem(TOKEN_KEY)
      localStorage.removeItem(PROFILE_KEY)
    },
  },
})
