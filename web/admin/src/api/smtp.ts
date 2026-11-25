import { http } from './request'
import type { SmtpConfig } from './types'

export const smtpApi = {
  get: () => http.get<SmtpConfig>('/smtp'),
  update: (body: SmtpConfig) => http.put<SmtpConfig>('/smtp', body),
}
