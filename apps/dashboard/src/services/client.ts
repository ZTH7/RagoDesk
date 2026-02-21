export class ApiError extends Error {
  code?: number
  requestId?: string
  status?: number

  constructor(message: string, options?: { code?: number; requestId?: string; status?: number }) {
    super(message)
    this.name = 'ApiError'
    this.code = options?.code
    this.requestId = options?.requestId
    this.status = options?.status
  }
}

import { getToken } from '../auth/storage'

const API_BASE = (import.meta.env.VITE_API_BASE_URL as string) || ''

export async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const token = getToken()
  const rawHeaders =
    init?.headers instanceof Headers ? Object.fromEntries(init.headers.entries()) : init?.headers || {}
  const isForm = typeof FormData !== 'undefined' && init?.body instanceof FormData
  const headers: Record<string, string> = {
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...(rawHeaders as Record<string, string>),
  }
  if (!isForm && !('Content-Type' in headers)) {
    headers['Content-Type'] = 'application/json'
  }
  const response = await fetch(`${API_BASE}${path}`, {
    credentials: 'include',
    headers,
    ...init,
  })

  let payload: any = null
  try {
    payload = await response.json()
  } catch (_) {
    // ignore
  }

  if (!response.ok) {
    const message = payload?.message || response.statusText || 'Request failed'
    throw new ApiError(message, { status: response.status, code: payload?.code, requestId: payload?.request_id })
  }

  if (payload && typeof payload === 'object' && 'code' in payload) {
    if (payload.code !== 0) {
      throw new ApiError(payload.message || 'Request failed', {
        code: payload.code,
        requestId: payload.request_id,
      })
    }
    return payload.data as T
  }

  return payload as T
}
