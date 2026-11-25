import { get, post } from './request'
import type { Order } from './types'

export interface GunByToken {
  gunId: number
  pileId: number
  gunNo: number
  gunCode: string
  status: string
  deviceNo: string
  stationId: number
  stationName: string
  stationAddress: string
  balanceCents: number
}

export const chargeApi = {
  /** 扫码后按枪令牌定位到具体枪 */
  gunByToken(token: string): Promise<GunByToken> {
    return get<GunByToken>('/gun/by-token', { token })
  },

  /** 立即启动，服务端校验余额 >= 场站 startThresholdCents */
  start(gunId: string): Promise<Order> {
    return post<Order>('/charge/start', { gunId })
  },

  /** 停止 charging 中的订单 */
  stop(gunId: string): Promise<Order> {
    return post<Order>('/charge/stop', { gunId })
  },
}
