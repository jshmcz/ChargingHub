/** 严格对应 docs/API.md，金额单位一律为分，时间为 ISO 8601 字符串 */

export type Role = 'super' | 'tenant_admin' | 'operator' | 'user'

export interface UserInfo {
  id: string
  phone: string
  email: string
  name: string
  status?: string
  role?: Role
  tenantId?: string | null
  balanceCents?: number
}

export interface LoginResult {
  token: string
  role: Role
  tenantId: string | null
  user: UserInfo
}

export type MeResult = Omit<LoginResult, 'token'>

export type GunStatus =
  | 'available'
  | 'idle'
  | 'charging'
  | 'reserved'
  | 'offline'
  | 'fault'
  | string

export interface Gun {
  id: string
  gunNo?: number | string
  no?: number | string
  name?: string
  gunCode?: string
  qrToken?: string
  status: GunStatus
  stationId?: string
  pileId?: string
  deviceNo?: string
}

export interface Pile {
  id: string
  stationId?: string
  deviceNo?: string
  gunCount?: number
  status?: string
  guns?: Gun[]
}

export interface Station {
  id: string
  name: string
  address: string
  areaCode?: string
  areaName?: string
  defaultStartTime: string
  startThresholdCents: number
  stopThresholdCents: number
  warnThresholdCents: number
  settlementTimeoutHours: number
  status?: string
  piles?: Pile[]
  guns?: Gun[]
  availableGuns?: number
  totalGuns?: number
}

export type StationPayload = Omit<Station, 'id' | 'status' | 'piles' | 'guns' | 'availableGuns' | 'totalGuns'>

export interface PilePayload {
  stationId: string
  gunCount: number
}

export interface Wallet {
  id?: string
  userId?: string
  tenantId?: string
  tenantName?: string
  balanceCents: number
  status?: string
  frozen?: boolean
}

export type LedgerType =
  | 'recharge'
  | 'refund'
  | 'consume'
  | 'settle'
  | string

export interface Ledger {
  id: string
  userId?: string
  type?: LedgerType
  amountCents: number
  balanceCents?: number
  method?: string
  voucherNo?: string
  remark?: string
  createdAt?: string
  createTime?: string
}

export type ReservationStatus =
  | 'pending'
  | 'active'
  | 'completed'
  | 'cancelled'
  | 'expired'
  | string

export interface Reservation {
  id: string
  stationId: string
  gunId: string
  userId?: string
  startAt: string
  status: ReservationStatus
  stationName?: string
  stationAddress?: string
  gunNo?: number | string
  createdAt?: string
  expireAt?: string
}

export type OrderStatus =
  | 'charging'
  | 'pending_settle'
  | 'settled'
  | 'stopped'
  | 'cancelled'
  | string

export interface Order {
  id: string
  orderNo?: string
  stationId?: string
  gunId?: string
  userId?: string
  status: OrderStatus
  amountCents?: number
  energyKwh?: number
  startTime?: string
  endTime?: string
  createdAt?: string
  settledAt?: string
}

export interface Notification {
  id: string
  title?: string
  content?: string
  type?: string
  channel?: string
  status?: string
  sentAt?: string
  createdAt?: string
  userId?: string
}

export interface Tenant {
  id: string
  name: string
  status?: string
  createdAt?: string
}

export interface SmtpConfig {
  host: string
  port: number
  username: string
  password: string
  from: string
  encryption: 'none' | 'ssl' | 'starttls'
  enabled: boolean
}
