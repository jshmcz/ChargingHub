import { http } from './request'
import type { Notification } from './types'

export const notificationsApi = {
  list: (userId?: number) => http.get<Notification[]>('/notifications', { userId }),
}
