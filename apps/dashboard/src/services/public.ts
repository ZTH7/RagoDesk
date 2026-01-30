import { request } from './client'

export type Session = {
  id: string
  status: string
  bot_id?: string
  tenant_id?: string
  user_external_id?: string
  metadata?: Record<string, unknown>
  created_at?: string
  updated_at?: string
  closed_at?: string
}

export type CreateSessionInput = {
  user_external_id?: string
  metadata?: Record<string, unknown>
}

export type SendMessageInput = {
  session_id: string
  message: string
  top_k?: number
  threshold?: number
}

export type Reference = {
  document_id: string
  document_version_id: string
  chunk_id: string
  score: number
  rank: number
  snippet: string
}

export type SendMessageResult = {
  reply: string
  confidence: number
  references: Reference[]
}

export type SessionMessage = {
  id: string
  role: string
  content: string
  confidence?: number
  references?: Reference[]
  created_at?: string
}

export type GetSessionResult = {
  session: Session
  messages?: SessionMessage[]
}

export const publicApi = {
  createSession(apiKey: string, payload: CreateSessionInput) {
    return request<{ session: Session }>('/api/v1/session', {
      method: 'POST',
      headers: {
        'X-API-Key': apiKey,
      },
      body: JSON.stringify(payload),
    })
  },
  sendMessage(apiKey: string, payload: SendMessageInput) {
    return request<SendMessageResult>('/api/v1/message', {
      method: 'POST',
      headers: {
        'X-API-Key': apiKey,
      },
      body: JSON.stringify(payload),
    })
  },
  getSession(apiKey: string, sessionId: string, options?: { include_messages?: boolean; limit?: number; offset?: number }) {
    const query = new URLSearchParams()
    if (options?.include_messages) query.set('include_messages', 'true')
    if (options?.limit) query.set('limit', String(options.limit))
    if (options?.offset) query.set('offset', String(options.offset))
    const suffix = query.toString() ? `?${query.toString()}` : ''
    return request<GetSessionResult>(`/api/v1/session/${sessionId}${suffix}`, {
      method: 'GET',
      headers: {
        'X-API-Key': apiKey,
      },
    })
  },
  closeSession(apiKey: string, sessionId: string, closeReason?: string) {
    return request<void>(`/api/v1/session/${sessionId}/close`, {
      method: 'POST',
      headers: {
        'X-API-Key': apiKey,
      },
      body: JSON.stringify({ session_id: sessionId, close_reason: closeReason }),
    })
  },
}
