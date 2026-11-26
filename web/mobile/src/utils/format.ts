/** 金额从分转换为元字符串 */
export function formatCents(cents?: number | null): string {
  return (Number(cents ?? 0) / 100).toFixed(2)
}

/** 带人民币符号的金额 */
export function yuan(cents?: number | null): string {
  return `¥${formatCents(cents)}`
}

/** 带正负号的流水金额 */
export function signedYuan(cents?: number | null): string {
  const value = Number(cents ?? 0)
  return `${value >= 0 ? '+' : '-'}¥${formatCents(Math.abs(value))}`
}

/** 按 Asia/Shanghai 展示时间 */
export function formatTime(iso?: string | null): string {
  if (!iso) return '-'
  const hasOffset = /(?:Z|[+-]\d{2}:?\d{2})$/i.test(iso)
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return iso
  const target = hasOffset ? new Date(date.getTime() + 8 * 60 * 60 * 1000) : date
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${target.getUTCFullYear()}-${pad(target.getUTCMonth() + 1)}-${pad(target.getUTCDate())} ${pad(target.getUTCHours())}:${pad(target.getUTCMinutes())}`
}

export function formatDate(iso?: string | null): string {
  const text = formatTime(iso)
  return text === '-' ? text : text.slice(0, 10)
}

export type VantTagType = 'default' | 'primary' | 'success' | 'warning' | 'danger'

const GUN_STATUS: Record<string, { text: string; type: VantTagType }> = {
  available: { text: '空闲', type: 'success' },
  idle: { text: '空闲', type: 'success' },
  free: { text: '空闲', type: 'success' },
  charging: { text: '充电中', type: 'primary' },
  busy: { text: '充电中', type: 'primary' },
  reserved: { text: '已预约', type: 'warning' },
  offline: { text: '离线', type: 'default' },
  fault: { text: '故障', type: 'danger' },
}

export function gunStatusText(status?: string | null): string {
  return GUN_STATUS[status || '']?.text ?? status ?? '未知'
}

export function gunStatusType(status?: string | null): VantTagType {
  return GUN_STATUS[status || '']?.type ?? 'default'
}

const ORDER_STATUS: Record<string, { text: string; type: VantTagType }> = {
  charging: { text: '充电中', type: 'primary' },
  pending_settle: { text: '待结算', type: 'warning' },
  settled: { text: '已结算', type: 'success' },
  stopped: { text: '已停止', type: 'default' },
  cancelled: { text: '已取消', type: 'default' },
}

export function orderStatusText(status?: string | null): string {
  return ORDER_STATUS[status || '']?.text ?? status ?? '未知'
}

export function orderStatusType(status?: string | null): VantTagType {
  return ORDER_STATUS[status || '']?.type ?? 'default'
}

const RESERVATION_STATUS: Record<string, { text: string; type: VantTagType }> = {
  pending: { text: '待启动', type: 'warning' },
  active: { text: '已生效', type: 'primary' },
  completed: { text: '已完成', type: 'success' },
  cancelled: { text: '已取消', type: 'default' },
  expired: { text: '已过期', type: 'default' },
}

export function reservationStatusText(status?: string | null): string {
  return RESERVATION_STATUS[status || '']?.text ?? status ?? '未知'
}

export function reservationStatusType(status?: string | null): VantTagType {
  return RESERVATION_STATUS[status || '']?.type ?? 'default'
}

const ROLE_TEXT: Record<string, string> = {
  super: '平台超管',
  tenant_admin: '租户管理员',
  operator: '租户操作员',
  user: '充电用户',
}

export function roleText(role?: string | null): string {
  return ROLE_TEXT[role || ''] ?? role ?? '未知'
}

export function roleTagType(role?: string | null): VantTagType {
  if (role === 'super') return 'danger'
  if (role === 'tenant_admin') return 'success'
  if (role === 'operator') return 'primary'
  return 'default'
}

const LEDGER_TEXT: Record<string, string> = {
  recharge: '充值',
  refund: '退款',
  consume: '充电消费',
  settle: '结算',
}

export function ledgerText(type?: string | null): string {
  return LEDGER_TEXT[type || ''] ?? type ?? '流水'
}
