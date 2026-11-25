/** 接口基础路径，默认 /api，可通过 VITE_API_BASE 覆盖 */
export const API_BASE = (import.meta.env.VITE_API_BASE || '/api').replace(/\/$/, '')

/** localStorage key */
export const STORAGE_KEYS = {
  token: 'ch_token',
  deviceId: 'ch_device_id',
  role: 'ch_role',
  tenantId: 'ch_tenant_id',
  user: 'ch_user',
} as const
