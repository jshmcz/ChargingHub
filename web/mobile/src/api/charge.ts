import { post } from './request'
import type { Order } from './types'

export const chargeApi = {
  /** 立即启动，服务端校验余额 >= 场站 startThresholdCents */
  start(gunId: string): Promise<Order> {
    return post<Order>('/charge/start', { gunId })
  },

  /** 停止 charging 中的订单 */
  stop(gunId: string): Promise<Order> {
    return post<Order>('/charge/stop', { gunId })
  },
}
