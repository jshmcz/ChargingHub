import { get, post } from './request'
import type { Reservation } from './types'

export const reservationApi = {
  mine(): Promise<Reservation[]> {
    return get<Reservation[]>('/my/reservations')
  },

  list(userId?: string): Promise<Reservation[]> {
    return get<Reservation[]>('/reservations', { userId })
  },

  /** 启动时间由服务端按场站 defaultStartTime 计算，当天过点顺延次日，用户不可修改 */
  create(stationId: string, gunId: string): Promise<Reservation> {
    return post<Reservation>('/reservations', { stationId, gunId })
  },

  cancel(id: string): Promise<unknown> {
    return post(`/reservations/${id}/cancel`)
  },
}
