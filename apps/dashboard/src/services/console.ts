import { request } from './client'
import type { ListParams } from './types'

export type KnowledgeBase = {
  id: string
  name: string
  description: string
  document_count?: number
  created_at: string
  updated_at?: string
}

export type DocumentItem = {
  id: string
  kb_id?: string
  title: string
  source_type: string
  status: string
  current_version: number
  updated_at: string
  created_at?: string
}

export type DocumentVersion = {
  id: string
  version: number
  status: string
  created_at: string
}

export type ApiKeyItem = {
  id: string
  bot_id: string
  name: string
  status: string
  scopes: string[]
  api_versions: string[]
  quota_daily: number
  qps_limit: number
  created_at: string
  last_used_at: string
}

export type UsageLogItem = {
  id: string
  path: string
  status_code: number
  latency_ms: number
  prompt_tokens: number
  completion_tokens: number
  total_tokens: number
  api_version?: string
  model?: string
  client_ip?: string
  user_agent?: string
  created_at: string
}

export type SessionItem = {
  id: string
  bot_id: string
  status: string
  close_reason: string
  user_external_id: string
  created_at: string
}

export type MessageItem = {
  id: string
  role: string
  content: string
  confidence?: number
  references_json?: string
  created_at: string
}

export type UsageSummary = {
  total: number
  error_count: number
  avg_latency_ms: number
  prompt_tokens: number
  completion_tokens: number
  total_tokens: number
}

export type BotItem = {
  id: string
  name: string
  status: string
  created_at: string
}

export type BotKnowledgeBase = {
  id: string
  bot_id: string
  kb_id: string
  weight?: number
  created_at: string
}

export type UserItem = {
  id: string
  name: string
  email: string
  status: string
  created_at: string
}

export type RoleItem = {
  id: string
  name: string
  created_at: string
}

export type PermissionItem = {
  code: string
  scope: string
  description?: string
}

export type CreateKnowledgeBaseInput = {
  name: string
  description: string
}

export type UploadDocumentInput = {
  kb_id: string
  title: string
  source_type: string
  raw_uri: string
}

export type CreateApiKeyInput = {
  bot_id: string
  name: string
  scopes: string[]
  api_versions: string[]
  quota_daily: number
  qps_limit: number
}

export type UpdateApiKeyInput = {
  name?: string
  status?: string
  scopes?: string[]
  api_versions?: string[]
  quota_daily?: number
  qps_limit?: number
}

export type CreateUserInput = {
  name: string
  email: string
  phone?: string
  status: string
}

export type CreateRoleInput = {
  name: string
}

export const consoleApi = {
  listKnowledgeBases() {
    return request<{ items: KnowledgeBase[] }>('/console/v1/knowledge_bases')
  },
  createKnowledgeBase(payload: CreateKnowledgeBaseInput) {
    return request<{ knowledge_base: KnowledgeBase }>('/console/v1/knowledge_bases', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  getKnowledgeBase(id: string) {
    return request<{ knowledge_base: KnowledgeBase }>(`/console/v1/knowledge_bases/${id}`)
  },
  updateKnowledgeBase(id: string, payload: CreateKnowledgeBaseInput) {
    return request<{ knowledge_base: KnowledgeBase }>(`/console/v1/knowledge_bases/${id}`, {
      method: 'PATCH',
      body: JSON.stringify({ id, ...payload }),
    })
  },
  deleteKnowledgeBase(id: string) {
    return request<void>(`/console/v1/knowledge_bases/${id}`, {
      method: 'DELETE',
    })
  },
  listDocuments(params?: ListParams & { kb_id?: string }) {
    const query = new URLSearchParams()
    if (params?.kb_id) query.set('kb_id', params.kb_id)
    if (params?.limit) query.set('limit', String(params.limit))
    if (params?.offset) query.set('offset', String(params.offset))
    const suffix = query.toString() ? `?${query.toString()}` : ''
    return request<{ items: DocumentItem[] }>(`/console/v1/documents${suffix}`)
  },
  uploadDocument(payload: UploadDocumentInput) {
    return request<{ document: DocumentItem; version: DocumentVersion }>(
      '/console/v1/documents/upload',
      {
        method: 'POST',
        body: JSON.stringify(payload),
      },
    )
  },
  getDocument(id: string) {
    return request<{ document: DocumentItem; versions: DocumentVersion[] }>(
      `/console/v1/documents/${id}`,
    )
  },
  deleteDocument(id: string) {
    return request<void>(`/console/v1/documents/${id}`, {
      method: 'DELETE',
    })
  },
  reindexDocument(id: string) {
    return request<void>(`/console/v1/documents/${id}/reindex`, {
      method: 'POST',
      body: JSON.stringify({ id }),
    })
  },
  rollbackDocument(id: string, version: number) {
    return request<void>(`/console/v1/documents/${id}/rollback`, {
      method: 'POST',
      body: JSON.stringify({ id, version }),
    })
  },
  listBots() {
    return request<{ items: BotItem[] }>('/console/v1/bots')
  },
  listBotKnowledgeBases(botId: string) {
    return request<{ items: BotKnowledgeBase[] }>(`/console/v1/bots/${botId}/knowledge_bases`)
  },
  bindBotKnowledgeBase(botId: string, kbId: string, weight?: number) {
    return request<{ bot_kb: BotKnowledgeBase }>(`/console/v1/bots/${botId}/knowledge_bases`, {
      method: 'POST',
      body: JSON.stringify({ bot_id: botId, kb_id: kbId, weight }),
    })
  },
  unbindBotKnowledgeBase(botId: string, kbId: string) {
    return request<void>(`/console/v1/bots/${botId}/knowledge_bases/${kbId}`, {
      method: 'DELETE',
    })
  },
  listApiKeys(params?: ListParams & { bot_id?: string }) {
    const query = new URLSearchParams()
    if (params?.bot_id) query.set('bot_id', params.bot_id)
    if (params?.limit) query.set('limit', String(params.limit))
    if (params?.offset) query.set('offset', String(params.offset))
    const suffix = query.toString() ? `?${query.toString()}` : ''
    return request<{ items: ApiKeyItem[] }>(`/console/v1/api_keys${suffix}`)
  },
  createApiKey(payload: CreateApiKeyInput) {
    return request<{ api_key: ApiKeyItem; raw_key: string }>('/console/v1/api_keys', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  updateApiKey(id: string, payload: UpdateApiKeyInput) {
    const body: Record<string, unknown> = { id }
    if (payload.name !== undefined) body.name = payload.name
    if (payload.status !== undefined) body.status = payload.status
    if (payload.scopes !== undefined) body.scopes = payload.scopes
    if (payload.api_versions !== undefined) body.api_versions = payload.api_versions
    if (payload.quota_daily !== undefined) body.quota_daily = payload.quota_daily
    if (payload.qps_limit !== undefined) body.qps_limit = payload.qps_limit

    return request<{ api_key: ApiKeyItem }>(`/console/v1/api_keys/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(body),
    })
  },
  deleteApiKey(id: string) {
    return request<void>(`/console/v1/api_keys/${id}`, {
      method: 'DELETE',
    })
  },
  rotateApiKey(id: string) {
    return request<{ api_key: ApiKeyItem; raw_key: string }>(`/console/v1/api_keys/${id}/rotate`, {
      method: 'POST',
      body: JSON.stringify({ id }),
    })
  },
  getUsageSummary(params?: { bot_id?: string; api_key_id?: string; api_version?: string; model?: string }) {
    const query = new URLSearchParams()
    if (params?.bot_id) query.set('bot_id', params.bot_id)
    if (params?.api_key_id) query.set('api_key_id', params.api_key_id)
    if (params?.api_version) query.set('api_version', params.api_version)
    if (params?.model) query.set('model', params.model)
    const suffix = query.toString() ? `?${query.toString()}` : ''
    return request<{ summary: UsageSummary }>(`/console/v1/api_usage/summary${suffix}`)
  },
  listUsageLogs(params?: ListParams & { bot_id?: string; api_key_id?: string; api_version?: string; model?: string }) {
    const query = new URLSearchParams()
    if (params?.bot_id) query.set('bot_id', params.bot_id)
    if (params?.api_key_id) query.set('api_key_id', params.api_key_id)
    if (params?.api_version) query.set('api_version', params.api_version)
    if (params?.model) query.set('model', params.model)
    if (params?.limit) query.set('limit', String(params.limit))
    if (params?.offset) query.set('offset', String(params.offset))
    const suffix = query.toString() ? `?${query.toString()}` : ''
    return request<{ items: UsageLogItem[] }>(`/console/v1/api_usage${suffix}`)
  },
  listSessions(params?: ListParams & { bot_id?: string }) {
    const query = new URLSearchParams()
    if (params?.bot_id) query.set('bot_id', params.bot_id)
    if (params?.limit) query.set('limit', String(params.limit))
    if (params?.offset) query.set('offset', String(params.offset))
    const suffix = query.toString() ? `?${query.toString()}` : ''
    return request<{ items: SessionItem[] }>(`/console/v1/sessions${suffix}`)
  },
  listMessages(sessionId: string) {
    return request<{ items: MessageItem[] }>(`/console/v1/sessions/${sessionId}/messages`)
  },
  createUser(tenantId: string, payload: CreateUserInput) {
    return request<{ user: UserItem }>(`/console/v1/tenants/${tenantId}/users`, {
      method: 'POST',
      body: JSON.stringify({ tenant_id: tenantId, ...payload }),
    })
  },
  listUsers(tenantId: string, params?: ListParams) {
    const query = new URLSearchParams()
    if (params?.limit) query.set('limit', String(params.limit))
    if (params?.offset) query.set('offset', String(params.offset))
    const suffix = query.toString() ? `?${query.toString()}` : ''
    return request<{ items: UserItem[] }>(`/console/v1/tenants/${tenantId}/users${suffix}`)
  },
  createRole(payload: CreateRoleInput) {
    return request<{ role: RoleItem }>('/console/v1/roles', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  listRoles(params?: ListParams) {
    const query = new URLSearchParams()
    if (params?.limit) query.set('limit', String(params.limit))
    if (params?.offset) query.set('offset', String(params.offset))
    const suffix = query.toString() ? `?${query.toString()}` : ''
    return request<{ items: RoleItem[] }>(`/console/v1/roles${suffix}`)
  },
  assignRole(userId: string, roleId: string) {
    return request<void>(`/console/v1/users/${userId}/roles`, {
      method: 'POST',
      body: JSON.stringify({ user_id: userId, role_id: roleId }),
    })
  },
  listPermissions() {
    return request<{ items: PermissionItem[] }>('/console/v1/permissions')
  },
  assignRolePermissions(roleId: string, permissionCodes: string[]) {
    return request<void>(`/console/v1/roles/${roleId}/permissions`, {
      method: 'POST',
      body: JSON.stringify({ role_id: roleId, permission_codes: permissionCodes }),
    })
  },
  listRolePermissions(roleId: string) {
    return request<{ items: PermissionItem[] }>(`/console/v1/roles/${roleId}/permissions`)
  },
}
