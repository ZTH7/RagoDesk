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
  public_chat_id?: string
  public_chat_enabled?: boolean
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
  references?: ReferenceItem[]
  created_at: string
}

export type ReferenceItem = {
  document_id: string
  document_version_id: string
  chunk_id: string
  score: number
  rank: number
  snippet?: string
}

export type UsageSummary = {
  total: number
  error_count: number
  avg_latency_ms: number
  prompt_tokens: number
  completion_tokens: number
  total_tokens: number
}

export type UsageExportResult = {
  content?: string
  content_type?: string
  filename?: string
  download_url?: string
  object_uri?: string
}

export type BotItem = {
  id: string
  name: string
  description?: string
  status: string
  created_at: string
  updated_at?: string
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
  email?: string
  phone?: string
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
  public_chat_enabled?: boolean
  scopes: string[]
  api_versions: string[]
  quota_daily: number
  qps_limit: number
}

export type UpdateApiKeyInput = {
  name?: string
  status?: string
  public_chat_enabled?: boolean
  scopes?: string[]
  api_versions?: string[]
  quota_daily?: number
  qps_limit?: number
}

export type CreateUserInput = {
  name: string
  email?: string
  phone?: string
  status: string
  password?: string
  send_invite?: boolean
  invite_base_url?: string
}

export type CreateRoleInput = {
  name: string
}

type RawApiKeyItem = Partial<ApiKeyItem> & {
  botId?: string
  publicChatId?: string
  publicChatEnabled?: boolean
  apiVersions?: string[]
  quotaDaily?: number
  qpsLimit?: number
  createdAt?: string
  lastUsedAt?: string
}

type RawApiKeyEnvelope = {
  api_key?: RawApiKeyItem
  apiKey?: RawApiKeyItem
  raw_key?: string
  rawKey?: string
}

export const consoleApi = {
  normalizeApiKey(input: RawApiKeyItem | null | undefined): ApiKeyItem {
    if (!input) {
      return {
        id: '',
        bot_id: '',
        name: '',
        status: '',
        public_chat_id: '',
        public_chat_enabled: false,
        scopes: [],
        api_versions: [],
        quota_daily: 0,
        qps_limit: 0,
        created_at: '',
        last_used_at: '',
      }
    }
    const chatID = input.public_chat_id ?? input.publicChatId ?? ''
    const enabledRaw = input.public_chat_enabled ?? input.publicChatEnabled
    return {
      id: input.id ?? '',
      bot_id: input.bot_id ?? input.botId ?? '',
      name: input.name ?? '',
      status: input.status ?? '',
      public_chat_id: chatID,
      public_chat_enabled: enabledRaw ?? Boolean(chatID),
      scopes: input.scopes ?? [],
      api_versions: input.api_versions ?? input.apiVersions ?? [],
      quota_daily: input.quota_daily ?? input.quotaDaily ?? 0,
      qps_limit: input.qps_limit ?? input.qpsLimit ?? 0,
      created_at: input.created_at ?? input.createdAt ?? '',
      last_used_at: input.last_used_at ?? input.lastUsedAt ?? '',
    }
  },
  normalizeDocument(input: any): DocumentItem {
    if (!input) {
      return {
        id: '',
        kb_id: '',
        title: '',
        source_type: '',
        status: '',
        current_version: 0,
        updated_at: '',
      }
    }
    return {
      id: input.id ?? '',
      kb_id: input.kb_id ?? input.kbId ?? '',
      title: input.title ?? '',
      source_type: input.source_type ?? input.sourceType ?? '',
      status: input.status ?? '',
      current_version: input.current_version ?? input.currentVersion ?? 0,
      updated_at: input.updated_at ?? input.updatedAt ?? '',
      created_at: input.created_at ?? input.createdAt ?? '',
    }
  },
  normalizeDocumentVersion(input: any): DocumentVersion {
    if (!input) {
      return { id: '', version: 0, status: '', created_at: '' }
    }
    return {
      id: input.id ?? '',
      version: input.version ?? 0,
      status: input.status ?? '',
      created_at: input.created_at ?? input.createdAt ?? '',
    }
  },
  normalizeKnowledgeBaseResponse(input: any): { knowledge_base: KnowledgeBase } {
    if (!input) return { knowledge_base: { id: '', name: '', description: '', created_at: '' } }
    return {
      knowledge_base: input.knowledge_base ?? input.knowledgeBase ?? input.knowledge ?? input.kb ?? input,
    }
  },
  listKnowledgeBases() {
    return request<{ items: KnowledgeBase[] }>('/console/v1/knowledge_bases')
  },
  createKnowledgeBase(payload: CreateKnowledgeBaseInput) {
    return request<any>('/console/v1/knowledge_bases', {
      method: 'POST',
      body: JSON.stringify(payload),
    }).then((res) => consoleApi.normalizeKnowledgeBaseResponse(res))
  },
  getKnowledgeBase(id: string) {
    return request<any>(`/console/v1/knowledge_bases/${id}`).then((res) =>
      consoleApi.normalizeKnowledgeBaseResponse(res),
    )
  },
  updateKnowledgeBase(id: string, payload: CreateKnowledgeBaseInput) {
    return request<any>(`/console/v1/knowledge_bases/${id}`, {
      method: 'PATCH',
      body: JSON.stringify({ id, ...payload }),
    }).then((res) => consoleApi.normalizeKnowledgeBaseResponse(res))
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
    return request<{ items: DocumentItem[] }>(`/console/v1/documents${suffix}`).then((res) => ({
      items: (res.items ?? []).map(consoleApi.normalizeDocument),
    }))
  },
  uploadDocument(payload: UploadDocumentInput) {
    return request<{ document: DocumentItem; version: DocumentVersion }>(
      '/console/v1/documents/upload',
      {
        method: 'POST',
        body: JSON.stringify(payload),
      },
    ).then((res) => ({
      document: consoleApi.normalizeDocument(res.document),
      version: consoleApi.normalizeDocumentVersion(res.version),
    }))
  },
  uploadDocumentFile(payload: FormData) {
    return request<{ items: { document: DocumentItem; version: DocumentVersion }[] }>(
      '/console/v1/documents/upload_file',
      {
        method: 'POST',
        body: payload,
      },
    ).then((res) => ({
      items: (res.items ?? []).map((item) => ({
        document: consoleApi.normalizeDocument(item.document),
        version: consoleApi.normalizeDocumentVersion(item.version),
      })),
    }))
  },
  getDocument(id: string) {
    return request<{ document: DocumentItem; versions: DocumentVersion[] }>(
      `/console/v1/documents/${id}`,
    ).then((res) => ({
      document: consoleApi.normalizeDocument(res.document),
      versions: (res.versions ?? []).map(consoleApi.normalizeDocumentVersion),
    }))
  },
  updateDocument(id: string, payload: { kb_id?: string }) {
    const body: Record<string, unknown> = { id }
    if (payload.kb_id !== undefined) body.kb_id = payload.kb_id
    return request<{ document: DocumentItem }>(`/console/v1/documents/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(body),
    })
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
  getBot(id: string) {
    return request<{ bot: BotItem }>(`/console/v1/bots/${id}`)
  },
  createBot(payload: { name: string; description?: string; status?: string }) {
    return request<{ bot: BotItem }>('/console/v1/bots', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  updateBot(id: string, payload: { name?: string; description?: string; status?: string }) {
    return request<{ bot: BotItem }>(`/console/v1/bots/${id}`, {
      method: 'PATCH',
      body: JSON.stringify({ id, ...payload }),
    })
  },
  deleteBot(id: string) {
    return request<void>(`/console/v1/bots/${id}`, {
      method: 'DELETE',
    })
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
    return request<{ items: RawApiKeyItem[] }>(`/console/v1/api_keys${suffix}`).then((res) => ({
      items: (res.items ?? []).map(consoleApi.normalizeApiKey),
    }))
  },
  getApiKey(id: string) {
    return request<RawApiKeyEnvelope>(`/console/v1/api_keys/${id}`).then((res) => ({
      api_key: consoleApi.normalizeApiKey(res.api_key ?? res.apiKey),
    }))
  },
  createApiKey(payload: CreateApiKeyInput) {
    return request<RawApiKeyEnvelope>('/console/v1/api_keys', {
      method: 'POST',
      body: JSON.stringify(payload),
    }).then((res) => ({
      api_key: consoleApi.normalizeApiKey(res.api_key ?? res.apiKey),
      raw_key: res.raw_key ?? res.rawKey ?? '',
    }))
  },
  updateApiKey(id: string, payload: UpdateApiKeyInput) {
    const body: Record<string, unknown> = { id }
    if (payload.name !== undefined) body.name = payload.name
    if (payload.status !== undefined) body.status = payload.status
    if (payload.public_chat_enabled !== undefined) body.public_chat_enabled = payload.public_chat_enabled
    if (payload.scopes !== undefined) body.scopes = payload.scopes
    if (payload.api_versions !== undefined) body.api_versions = payload.api_versions
    if (payload.quota_daily !== undefined) body.quota_daily = payload.quota_daily
    if (payload.qps_limit !== undefined) body.qps_limit = payload.qps_limit

    return request<RawApiKeyEnvelope>(`/console/v1/api_keys/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(body),
    }).then((res) => ({
      api_key: consoleApi.normalizeApiKey(res.api_key ?? res.apiKey),
    }))
  },
  deleteApiKey(id: string) {
    return request<void>(`/console/v1/api_keys/${id}`, {
      method: 'DELETE',
    })
  },
  rotateApiKey(id: string) {
    return request<RawApiKeyEnvelope>(`/console/v1/api_keys/${id}/rotate`, {
      method: 'POST',
      body: JSON.stringify({ id }),
    }).then((res) => ({
      api_key: consoleApi.normalizeApiKey(res.api_key ?? res.apiKey),
      raw_key: res.raw_key ?? res.rawKey ?? '',
    }))
  },
  regenerateApiKeyPublicChatID(id: string) {
    return request<RawApiKeyEnvelope>(`/console/v1/api_keys/${id}/public_chat/regenerate`, {
      method: 'POST',
      body: JSON.stringify({ id }),
    }).then((res) => ({
      api_key: consoleApi.normalizeApiKey(res.api_key ?? res.apiKey),
    }))
  },
  getUsageSummary(params?: {
    bot_id?: string
    api_key_id?: string
    api_version?: string
    model?: string
    start_time?: string
    end_time?: string
  }) {
    const query = new URLSearchParams()
    if (params?.bot_id) query.set('bot_id', params.bot_id)
    if (params?.api_key_id) query.set('api_key_id', params.api_key_id)
    if (params?.api_version) query.set('api_version', params.api_version)
    if (params?.model) query.set('model', params.model)
    if (params?.start_time) query.set('start_time', params.start_time)
    if (params?.end_time) query.set('end_time', params.end_time)
    const suffix = query.toString() ? `?${query.toString()}` : ''
    return request<{ summary: UsageSummary }>(`/console/v1/api_usage/summary${suffix}`)
  },
  listUsageLogs(
    params?: ListParams & {
      bot_id?: string
      api_key_id?: string
      api_version?: string
      model?: string
      start_time?: string
      end_time?: string
    },
  ) {
    const query = new URLSearchParams()
    if (params?.bot_id) query.set('bot_id', params.bot_id)
    if (params?.api_key_id) query.set('api_key_id', params.api_key_id)
    if (params?.api_version) query.set('api_version', params.api_version)
    if (params?.model) query.set('model', params.model)
    if (params?.start_time) query.set('start_time', params.start_time)
    if (params?.end_time) query.set('end_time', params.end_time)
    if (params?.limit) query.set('limit', String(params.limit))
    if (params?.offset) query.set('offset', String(params.offset))
    const suffix = query.toString() ? `?${query.toString()}` : ''
    return request<{ items: UsageLogItem[] }>(`/console/v1/api_usage${suffix}`)
  },
  exportUsageLogs(payload: {
    api_key_id?: string
    bot_id?: string
    api_version?: string
    model?: string
    start_time?: string
    end_time?: string
    format?: string
    limit?: number
    offset?: number
  }) {
    return request<UsageExportResult>('/console/v1/api_usage/export', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  listSessions(params?: ListParams) {
    const query = new URLSearchParams()
    if (params?.limit) query.set('limit', String(params.limit))
    if (params?.offset) query.set('offset', String(params.offset))
    const suffix = query.toString() ? `?${query.toString()}` : ''
    return request<{ sessions: SessionItem[] }>(`/console/v1/sessions${suffix}`).then((res) => ({
      items: res.sessions ?? [],
    }))
  },
  listMessages(sessionId: string) {
    return request<{ messages: MessageItem[] }>(`/console/v1/sessions/${sessionId}/messages`).then(
      (res) => ({
        items: res.messages ?? [],
      }),
    )
  },
  createUser(tenantId: string, payload: CreateUserInput) {
    return request<{ user: UserItem; invite_link?: string }>(`/console/v1/tenants/${tenantId}/users`, {
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
