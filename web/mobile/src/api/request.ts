import { showToast } from 'vant'
import router from '@/router'
import { useAuthStore } from '@/stores/auth'
import { API_BASE } from '@/utils/constants'

/** 后端统一响应包装：{ code, message, data } */
export interface ApiEnvelope<T> {
  code: number
  message: string
  data: T
}

export class ApiError extends Error {
  readonly code: number

  constructor(code: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.code = code
  }
}

export type QueryValue = string | number | boolean | null | undefined

function buildUrl(path: string): string {
  if (/^https?:\/\//i.test(path)) return path
  return `${API_BASE}${path.startsWith('/') ? path : `/${path}`}`
}

function withQuery(path: string, params?: Record<string, QueryValue>): string {
  if (!params) return path
  const search = new URLSearchParams()
  Object.entries(params).forEach(([key, value]) => {
    if (value === undefined || value === null || value === '') return
    search.append(key, String(value))
  })
  const query = search.toString()
  return query ? `${path}${path.includes('?') ? '&' : '?'}${query}` : path
}

function handleUnauthorized(): void {
  const auth = useAuthStore()
  auth.clearAuth()
  const current = router.currentRoute.value
  if (current.path !== '/login') {
    void router.push({
      path: '/login',
      query: current.fullPath === '/home' ? {} : { redirect: current.fullPath },
    })
  }
}

/** 原生 fetch 封装：自动携带 Bearer token、统一解包、错误提示、401 跳登录 */
export async function request<T = unknown>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const auth = useAuthStore()
  const headers = new Headers(options.headers)
  if (options.body !== undefined && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }
  if (auth.token) {
    headers.set('Authorization', `Bearer ${auth.token}`)
  }

  let response: Response
  try {
    response = await fetch(buildUrl(path), { ...options, headers })
  } catch {
    showToast('网络连接失败，请稍后重试')
    throw new ApiError(-1, '网络连接失败，请稍后重试')
  }

  if (response.status === 401) {
    handleUnauthorized()
    showToast('登录已过期，请重新登录')
    throw new ApiError(401, '登录已过期，请重新登录')
  }

  let envelope: ApiEnvelope<T>
  try {
    envelope = (await response.json()) as ApiEnvelope<T>
  } catch {
    showToast('服务返回数据异常')
    throw new ApiError(-2, '服务返回数据异常')
  }

  if (envelope.code === 401) {
    handleUnauthorized()
  }
  if (envelope.code !== 0) {
    showToast(envelope.message || '请求失败')
    throw new ApiError(envelope.code, envelope.message || '请求失败')
  }

  return envelope.data
}

export function get<T = unknown>(path: string, params?: Record<string, QueryValue>): Promise<T> {
  return request<T>(withQuery(path, params), { method: 'GET' })
}

export function post<T = unknown>(path: string, body?: unknown): Promise<T> {
  return request<T>(path, {
    method: 'POST',
    ...(body === undefined ? {} : { body: JSON.stringify(body) }),
  })
}

export function put<T = unknown>(path: string, body?: unknown): Promise<T> {
  return request<T>(path, {
    method: 'PUT',
    ...(body === undefined ? {} : { body: JSON.stringify(body) }),
  })
}
