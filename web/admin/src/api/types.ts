/** docs/API.md v1 中的实体类型。金额字段单位均为分。 */

export type Role = 'super' | 'tenant_admin' | 'operator' | 'user'
export type Encryption = 'none' | 'ssl' | 'starttls'

/** 状态码取值文档未明确列出，使用 string 以兼容服务端定义。 */
export type Status = string

export type OrderStatus = 'charging' | 'pending_settle' | 'settled' | string
export type ReservationStatus = 'pending' | 'active' | 'cancelled' | 'completed' | string
export type GunStatus = 'idle' | 'reserved' | 'charging' | 'offline' | string
export type PileStatus = 'online' | 'offline' | 'disabled' | string
export type LedgerType = 'recharge' | 'refund' | 'charge' | 'settle' | string

export interface LoginUser {
  id: number
  phone: string
  email: string
  name: string
}

export interface LoginResult {
  token: string
  role: Role
  tenantId: number | null
  user: LoginUser
}

export interface MeResult {
  role: Role
  tenantId: number | null
  user: LoginUser
}

export interface Tenant {
  id: number
  name: string
  status: Status
  createdAt?: string
}

export interface User {
  id: number
  phone: string
  email: string
  name: string
  role?: Role
  tenantId?: number | null
  status?: Status
  balanceCents?: number
  createdAt?: string
}

export interface Station {
  id: number
  name: string
  address: string
  defaultStartTime: string
  startThresholdCents: number
  stopThresholdCents: number
  warnThresholdCents: number
  settlementTimeoutHours: number
  createdAt?: string
}

export interface Gun {
  id: number
  pileId: number
  stationId: number
  gunNo: string
  status: GunStatus
}

export interface Pile {
  id: number
  stationId: number
  deviceNo: string
  secret: string
  status: PileStatus
  guns?: Gun[]
  createdAt?: string
}

export interface Wallet {
  id: number
  userId: number
  tenantId?: number
  balanceCents: number
  frozen?: boolean
}

export interface Ledger {
  id: number
  userId: number
  amountCents: number
  type: LedgerType
  method: string
  voucherNo: string
  remark: string
  createdAt?: string
}

export interface Reservation {
  id: number
  userId: number
  stationId: number
  gunId: number
  stationName?: string
  startTime: string
  status: ReservationStatus
  createdAt?: string
}

export interface Order {
  id: number
  userId: number
  stationId: number
  gunId: number
  status: OrderStatus
  amountCents: number
  startTime?: string
  endTime?: string
  createdAt?: string
}

export interface SmtpConfig {
  host: string
  port: number
  username: string
  password: string
  from: string
  encryption: Encryption
  enabled: boolean
}

export interface Notification {
  id: number
  userId: number
  type: string
  title: string
  content: string
  status: string
  createdAt?: string
}

/** 统一响应包装 */
export interface ApiResponse<T> {
  code: number
  message: string
  data: T
}
