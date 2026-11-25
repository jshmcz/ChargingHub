import { http } from './request'
import type { Status, Tenant } from './types'

export interface TenantInput {
  name: string
  status: Status
  adminName?: string
  adminPhone?: string
  adminEmail?: string
  adminPassword?: string
}

export interface TenantCreated {
  id: number
  name: string
  adminPhone?: string
  adminUserId?: number
}

export const tenantsApi = {
  list: () => http.get<Tenant[]>('/tenants'),
  create: (body: TenantInput) => http.post<TenantCreated>('/tenants', body),
  update: (id: number, body: TenantInput) => http.put<Tenant>(`/tenants/${id}`, body),
}
