import { http } from './request'
import type { Status, Tenant } from './types'

export interface TenantInput {
  name: string
  status: Status
}

export const tenantsApi = {
  list: () => http.get<Tenant[]>('/tenants'),
  create: (body: TenantInput) => http.post<Tenant>('/tenants', body),
  update: (id: number, body: TenantInput) => http.put<Tenant>(`/tenants/${id}`, body),
}
