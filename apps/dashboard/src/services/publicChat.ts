import { ApiError } from './client'

const API_BASE = (import.meta.env.VITE_API_BASE_URL as string) || ''

type RawError = {
  code?: number
  message?: string
}

function parseError(status: number, payload: RawError | null) {
  const message = payload?.message || `Request failed (${status})`
  return new ApiError(message, { status, code: payload?.code })
}

async function requestWithChatID<T>(chatID: string, path: string, init?: RequestInit): Promise<T> {
  const headers: Record<string, string> = {
    'X-Chat-Key': chatID,
    ...(init?.headers as Record<string, string>),
  }
  if (!(init?.body instanceof FormData)) {
    headers['Content-Type'] = headers['Content-Type'] || 'application/json'
  }
  const response = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers,
  })

  let payload: any = null
  try {
    payload = await response.json()
  } catch {
    // ignore
  }

  if (!response.ok) {
    throw parseError(response.status, payload)
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

type APITimestamp = {
  seconds?: number
  nanos?: number
}

export type PublicMessage = {
  id: string
  role: string
  content: string
  created_at?: APITimestamp
}

export type CreatePublicSessionResponse = {
  session: {
    id: string
  }
}

export type GetPublicSessionResponse = {
  session: {
    id: string
    status: string
  }
  messages: PublicMessage[]
}

export type SendPublicMessageResponse = {
  reply: string
  confidence?: number
  references?: Array<{
    document_id: string
    snippet?: string
  }>
}

export const publicChatApi = {
  createSession(chatID: string, userExternalID: string) {
    return requestWithChatID<CreatePublicSessionResponse>(chatID, '/api/v1/session', {
      method: 'POST',
      body: JSON.stringify({
        user_external_id: userExternalID,
        metadata: {
          source: 'public_chat',
        },
      }),
    })
  },
  getSession(chatID: string, sessionID: string) {
    return requestWithChatID<GetPublicSessionResponse>(
      chatID,
      `/api/v1/session/${encodeURIComponent(sessionID)}?include_messages=true&limit=200`,
      { method: 'GET' },
    )
  },
  sendMessage(chatID: string, sessionID: string, message: string) {
    return requestWithChatID<SendPublicMessageResponse>(chatID, '/api/v1/message', {
      method: 'POST',
      body: JSON.stringify({
        session_id: sessionID,
        message,
      }),
    })
  },
}
