import { STORAGE_KEYS } from './constants'
import { generateUUID } from './uuid'

/** 获取并持久化设备 ID；首次访问时生成 UUID */
export function getDeviceId(): string {
  let deviceId = localStorage.getItem(STORAGE_KEYS.deviceId)
  if (!deviceId) {
    deviceId = generateUUID()
    localStorage.setItem(STORAGE_KEYS.deviceId, deviceId)
  }
  return deviceId
}
