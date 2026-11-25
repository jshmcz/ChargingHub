import { get, post } from './request'
import type { Order } from './types'

export const orderApi = {
  mine(): Promise<Order[]> {
    return get<Order[]>('/my/orders')
  },

  list(userId?: string): Promise<Order[]> {
    return get<Order[]>('/orders', { userId })
  },

  /** 手动结算：pending_settle -> settled */
  settle(id: string): Promise<Order> {
    return post<Order>(`/orders/${id}/settle`)
  },
}
