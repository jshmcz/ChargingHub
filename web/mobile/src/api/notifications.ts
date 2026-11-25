import { get } from './request'
import type { Notification } from './types'

export const notificationApi = {
  /** 发送记录；不传 userId 时由服务端按权限返回 */
  list(userId?: string): Promise<Notification[]> {
    return get<Notification[]>('/notifications', { userId })
  },
}
