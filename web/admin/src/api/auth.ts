import { http } from './request'
import type { LoginResult, MeResult } from './types'

export interface LoginPayload {
  username: string
  password: string
  deviceId: string
  tenantId?: number | string
}

export const authApi = {
  login: (payload: LoginPayload) => http.post<LoginResult>('/login', payload),
  me: () => http.get<MeResult>('/me'),
  logout: () => http.post<null>('/logout'),
}
