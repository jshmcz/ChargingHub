import { getDeviceId } from '@/utils/device'
import { get, post } from './request'
import type { LoginResult, MeResult } from './types'

export interface LoginCredentials {
  username: string
  password: string
}

export const authApi = {
  /** POST /api/login，deviceId 本地持久化，不存在时自动生成 UUID */
  login(credentials: LoginCredentials): Promise<LoginResult> {
    return post<LoginResult>('/login', {
      ...credentials,
      deviceId: getDeviceId(),
    })
  },

  me(): Promise<MeResult> {
    return get<MeResult>('/me')
  },

  logout(): Promise<unknown> {
    return post('/logout')
  },
}
