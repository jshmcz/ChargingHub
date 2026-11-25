import { get, post, put } from './request'
import type { Role, SmtpConfig, Tenant, UserInfo } from './types'

export interface CreateUserPayload {
  phone: string
  email: string
  password: string
  name: string
  role?: Role
  tenantId?: string
}

export interface UpdateUserPayload {
  status: string
  email: string
  name: string
}

export const adminApi = {
  users(keyword?: string): Promise<UserInfo[]> {
    return get<UserInfo[]>('/users', { keyword })
  },

  createUser(payload: CreateUserPayload): Promise<UserInfo> {
    return post<UserInfo>('/users', payload)
  },

  updateUser(id: string, payload: UpdateUserPayload): Promise<UserInfo> {
    return put<UserInfo>(`/users/${id}`, payload)
  },

  tenants(): Promise<Tenant[]> {
    return get<Tenant[]>('/tenants')
  },

  createTenant(payload: { name: string; status?: string }): Promise<Tenant> {
    return post<Tenant>('/tenants', payload)
  },

  updateTenant(id: string, payload: { name: string; status: string }): Promise<Tenant> {
    return put<Tenant>(`/tenants/${id}`, payload)
  },

  smtp(): Promise<SmtpConfig> {
    return get<SmtpConfig>('/smtp')
  },

  updateSmtp(payload: SmtpConfig): Promise<SmtpConfig> {
    return put<SmtpConfig>('/smtp', payload)
  },
}
