/** 生成 UUID v4，优先使用浏览器原生实现。 */
export function genUuid(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (char) => {
    const random = (Math.random() * 16) | 0
    const value = char === 'x' ? random : (random & 0x3) | 0x8
    return value.toString(16)
  })
}

const DEVICE_KEY = 'charginghub_device_id'

/** 设备 ID 持久化：不存在则生成，用于登录会话互踢策略。 */
export function getOrCreateDeviceId(): string {
  let deviceId = localStorage.getItem(DEVICE_KEY)
  if (!deviceId) {
    deviceId = genUuid()
    localStorage.setItem(DEVICE_KEY, deviceId)
  }
  return deviceId
}
