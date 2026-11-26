import { http } from './request'
import type { SmtpConfig } from './types'

function tenantQuery(tenantId?: number): string {
  return tenantId && tenantId > 0 ? `?tenantId=${tenantId}` : ''
}

export const smtpApi = {
  get: (tenantId?: number) => http.get<SmtpConfig>(`/smtp${tenantQuery(tenantId)}`),
  update: (body: SmtpConfig, tenantId?: number) =>
    http.put<SmtpConfig>(`/smtp${tenantQuery(tenantId)}`, body),
}
