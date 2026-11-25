import { http } from './request'
import type { Order } from './types'

export const ordersApi = {
  list: (userId?: number) => http.get<Order[]>('/orders', { userId }),
  settle: (id: number) => http.post<Order>(`/orders/${id}/settle`),
  my: () => http.get<Order[]>('/my/orders'),
}
