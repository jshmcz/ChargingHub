import { http } from './request'
import type { Role, Status, User } from './types'

export interface CreateUserPayload {
  phone: string
  email: string
  password: string
  name: string
  /** 仅超管创建租户管理员时需要 */
  role?: Role
  tenantId?: number | string
}

export interface UpdateUserPayload {
  status: Status
  email: string
  name: string
}

export const usersApi = {
  list: (keyword?: string) => http.get<User[]>('/users', { keyword }),
  create: (body: CreateUserPayload) => http.post<User>('/users', body),
  update: (id: number, body: UpdateUserPayload) => http.put<User>(`/users/${id}`, body),
}
