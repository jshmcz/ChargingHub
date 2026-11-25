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
