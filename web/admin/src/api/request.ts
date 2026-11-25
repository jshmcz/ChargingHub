import type { ApiResponse } from './types'

/** 与 auth store 共用的 token 存储键 */
export const TOKEN_KEY = 'charginghub_token'

export type QueryValue = string | number | boolean | undefined | null

/** 业务错误：code 为服务端业务码，HTTP 401 时为 401，网络异常时为 -1。 */
export class ApiError extends Error {
  code: number

  constructor(code: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.code = code
  }
}

function getBaseUrl(): string {
  return (import.meta.env.VITE_API_BASE as string | undefined) || '/api'
}

function buildQuery(params?: Record<string, QueryValue>): string {
  if (!params) return ''
  const search = new URLSearchParams()
  Object.entries(params).forEach(([key, value]) => {
    if (value === undefined || value === null || value === '') return
    search.append(key, String(value))
  })
  const query = search.toString()
  return query ? `?${query}` : ''
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const url = `${getBaseUrl()}${path}`
  const headers = new Headers(init.headers)
  const token = localStorage.getItem(TOKEN_KEY)
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }
  if (init.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  let response: Response
  try {
    response = await fetch(url, { ...init, headers })
  } catch {
    throw new ApiError(-1, '网络连接失败，请稍后重试')
  }

  if (response.status === 401) {
    localStorage.removeItem(TOKEN_KEY)
    // 登录接口自身的 401 表示账号密码错误，交由页面展示错误，不做跳转。
    if (!path.includes('/login')) {
      window.location.href = '/login'
    }
    throw new ApiError(401, '登录已过期，请重新登录')
  }

  let payload: ApiResponse<T>
  try {
    payload = (await response.json()) as ApiResponse<T>
  } catch {
    throw new ApiError(response.status, '服务响应格式异常')
  }

  if (!payload || payload.code !== 0) {
    throw new ApiError(payload?.code ?? response.status, payload?.message || '请求失败')
  }
  return payload.data
}

function jsonInit(method: string, body?: unknown): RequestInit {
  const init: RequestInit = { method }
  if (body !== undefined) {
    init.body = JSON.stringify(body)
  }
  return init
}

export const http = {
  get<T>(path: string, params?: Record<string, QueryValue>): Promise<T> {
    return request<T>(`${path}${buildQuery(params)}`, { method: 'GET' })
  },
  post<T>(path: string, body?: unknown): Promise<T> {
    return request<T>(path, jsonInit('POST', body))
  },
  put<T>(path: string, body?: unknown): Promise<T> {
    return request<T>(path, jsonInit('PUT', body))
  },
  delete<T>(path: string): Promise<T> {
    return request<T>(path, { method: 'DELETE' })
  },
}
