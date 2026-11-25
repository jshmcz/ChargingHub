import { http } from './request'

export interface PlatformConfig {
  baseUrl: string
}

export const platformApi = {
  get: () => http.get<PlatformConfig>('/platform/config'),
  update: (body: PlatformConfig) => http.put<PlatformConfig>('/platform/config', body),
}
