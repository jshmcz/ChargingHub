/** 金额在接口中是分，展示时转为元。 */
export function formatCents(cents?: number): string {
  return `¥${((cents ?? 0) / 100).toFixed(2)}`
}

export function centsToYuan(cents?: number): number {
  return Number(((cents ?? 0) / 100).toFixed(2))
}

export function yuanToCents(yuan: number | string): number {
  const value = typeof yuan === 'string' ? Number(yuan) : yuan
  return Math.round((Number.isFinite(value) ? value : 0) * 100)
}

export function formatDate(value?: string): string {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString('zh-CN', { hour12: false })
}

const ROLE_TEXT: Record<string, string> = {
  super: '平台超管',
  tenant_admin: '租户管理员',
  operator: '租户操作员',
  user: '充电用户',
}

export function roleText(role?: string): string {
  return ROLE_TEXT[role || ''] || role || '未知'
}

export type RoleTagType = 'primary' | 'success' | 'info' | 'warning' | 'danger'

export function roleTagType(role?: string): RoleTagType {
  if (role === 'super') return 'danger'
  if (role === 'tenant_admin') return 'success'
  if (role === 'operator') return 'primary'
  return 'info'
}
