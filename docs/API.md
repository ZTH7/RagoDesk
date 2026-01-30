# API Specification — RagoDesk

> 对外 API 仅提供机器人能力与会话接口；管理后台 API 仅供企业/平台内部使用。
>
> 采用 **模块化单体**，对外通过 Kratos HTTP 提供 API；gRPC 作为未来拆分微服务的预留能力。

---

## 1. 通用约定

### 1.1 认证
- Header: `X-API-Key: <key>`（API Key 绑定 bot，用于定位租户与机器人配置）
- 可选签名：`X-Timestamp`, `X-Nonce`, `X-Signature` (HMAC-SHA256)
- 服务端校验时间窗口与 nonce 防重放
- 管理后台使用 `Authorization: Bearer <JWT>`（通过登录接口获取）

### 1.2 统一响应结构
```json
{
  "code": 0,
  "message": "ok",
  "data": {},
  "request_id": "req_abc123"
}
```

### 1.3 通用错误码
- `40001` 参数错误
- `40101` API Key 无效
- `40102` 签名校验失败
- `40301` 无权限
- `40401` 资源不存在
- `42901` 超出配额或限流
- `50001` 内部错误

### 1.4 速率限制
- Header: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`

---

## 2. 对外核心 API（租户调用）

### 2.1 创建会话
`POST /api/v1/session`

**Request**
```json
{
  "user_external_id": "u_998",
  "metadata": {"source": "web"}
}
```

**Response**
```json
{
  "code": 0,
  "data": {
    "session_id": "sess_abc",
    "status": "bot"
  }
}
```

---

### 2.2 发送消息（机器人回复）
`POST /api/v1/message`

**Headers**
- `X-API-Key`: required (绑定 bot，用于定位租户与模型/知识库)

**Request**
```json
{
  "session_id": "sess_abc",
  "message": "如何申请退款？"
}
```

**Response**
```json
{
  "code": 0,
  "data": {
    "reply": "您可以在订单页面点击申请退款...",
    "confidence": 0.78,
    "references": [
      {"doc_id": "doc_12", "chunk_id": "ck_99", "score": 0.82}
    ]
  }
}
```

---

### 2.3 获取会话状态
`GET /api/v1/session/{id}`

**Response**
```json
{
  "code": 0,
  "data": {
    "session_id": "sess_abc",
    "status": "bot",
    "metadata": {"source": "web"}
  }
}
```

---

### 2.4 结束会话
`POST /api/v1/session/{id}/close`

**Request**
```json
{
  "session_id": "sess_abc",
  "close_reason": "user_end"
}
```

---

### 2.5 用户反馈（用于优化）
`POST /api/v1/feedback`

**Request**
```json
{
  "session_id": "sess_abc",
  "message_id": "msg_321",
  "rating": 1,
  "comment": "回答不准确",
  "correction": "正确答案应为..."
}
```

---

## 3. 平台 API（Platform）

### 3.0 平台登录
- `POST /platform/v1/login`

**Request**
```json
{
  "account": "admin@ragodesk.ai",
  "password": "******"
}
```

**Response**
```json
{
  "code": 0,
  "data": {
    "token": "jwt_xxx",
    "expires_at": "2026-02-22T00:00:00Z",
    "profile": {
      "subject_id": "admin_123",
      "account": "admin@ragodesk.ai",
      "name": "Platform Admin",
      "roles": ["platform_admin"]
    }
  }
}
```

### 3.1 租户管理
- `POST /platform/v1/tenants`
  - body: `name`, `plan`, `status`, `type`（`personal|enterprise`，默认 `enterprise`）
- `GET /platform/v1/tenants`
- `GET /platform/v1/tenants/{id}`

### 3.2 权限目录
- `POST /platform/v1/permissions`（`scope=platform|tenant`）
- `GET /platform/v1/permissions`

### 3.3 平台管理员
- `POST /platform/v1/admins`
- `GET /platform/v1/admins`
- `POST /platform/v1/admins/{id}/roles`

**Create**
`POST /platform/v1/admins`
```json
{
  "name": "Platform Admin",
  "email": "admin@company.com",
  "status": "active",
  "password": "******"
}
```

**Invite Link**
```json
{
  "name": "Platform Admin",
  "email": "admin@company.com",
  "status": "active",
  "send_invite": true,
  "invite_base_url": "http://localhost:5173"
}
```

### 3.4 平台角色
- `POST /platform/v1/roles`
- `GET /platform/v1/roles`
- `POST /platform/v1/roles/{id}/permissions`
- `GET /platform/v1/roles/{id}/permissions`

---

## 4. 管理后台 API（Console）

### 4.0 注册 / 登录
- `POST /console/v1/register`
- `POST /console/v1/login`

**Register**
```json
{
  "tenant_name": "Acme Inc",
  "tenant_type": "enterprise",
  "admin_name": "Alice",
  "email": "alice@acme.com",
  "password": "******"
}
```

**Login**
```json
{
  "account": "alice@acme.com",
  "password": "******",
  "tenant_id": "tenant_123"
}
```
> 若同一账号存在多租户，需指定 `tenant_id`。

**Response**
```json
{
  "code": 0,
  "data": {
    "token": "jwt_xxx",
    "expires_at": "2026-02-22T00:00:00Z",
    "profile": {
      "subject_id": "user_123",
      "tenant_id": "tenant_123",
      "account": "alice@acme.com",
      "name": "Alice",
      "roles": ["tenant_admin"]
    }
  }
}
```

### 4.1 成员与角色
- `POST /console/v1/tenants/{id}/users`（邀请成员）
- `GET /console/v1/tenants/{id}/users`
- `POST /console/v1/roles`
- `GET /console/v1/roles`
- `POST /console/v1/users/{id}/roles`（分配角色）

### 4.2 权限分配
- `GET /console/v1/permissions`（仅租户可见权限）
- `POST /console/v1/roles/{id}/permissions`（分配权限）
- `GET /console/v1/roles/{id}/permissions`

### 4.3 机器人管理
- `POST /console/v1/bots`
- `GET /console/v1/bots`
- `GET /console/v1/bots/{id}`
- `PATCH /console/v1/bots/{id}`
- `DELETE /console/v1/bots/{id}`
- `GET /console/v1/bots/{id}/knowledge_bases`
- `POST /console/v1/bots/{id}/knowledge_bases`（绑定）
- `DELETE /console/v1/bots/{id}/knowledge_bases/{kb_id}`（解绑）
绑定请求字段：`kb_id`, `weight`（可选）

### 4.4 知识库管理
- `POST /console/v1/knowledge_bases`
- `GET /console/v1/knowledge_bases`
- `GET /console/v1/knowledge_bases/{id}`
- `PATCH /console/v1/knowledge_bases/{id}`
- `DELETE /console/v1/knowledge_bases/{id}`
- `POST /console/v1/documents/upload`
- `GET /console/v1/documents`（可选 `kb_id` 过滤）
- `GET /console/v1/documents/{id}`
- `DELETE /console/v1/documents/{id}`
- `POST /console/v1/documents/{id}/reindex`
- `POST /console/v1/documents/{id}/rollback`
上传请求字段：`kb_id`, `title`, `source_type`, `raw_uri`（OSS URI 或预签 URL）

### 4.5 API Key 管理
- `POST /console/v1/api_keys`
- `GET /console/v1/api_keys`
- `PATCH /console/v1/api_keys/{id}`
- `DELETE /console/v1/api_keys/{id}`
- `POST /console/v1/api_keys/{id}/rotate`
- `GET /console/v1/api_usage`
- `GET /console/v1/api_usage/summary`
- `POST /console/v1/api_usage/export`
> 支持 scope 配置与 Key 轮换（旧 Key 进入过渡期）。`scopes` 默认：`["rag","conversation"]`，可用 `*` 表示全量权限。

**Create**
`POST /console/v1/api_keys`
```json
{
  "bot_id": "bot_123",
  "name": "prod-key",
  "scopes": ["rag", "conversation"],
  "api_versions": ["v1"],
  "quota_daily": 20000,
  "qps_limit": 50
}
```
返回：`api_key` + `raw_key`（仅创建/轮换时返回一次）。

**List**
`GET /console/v1/api_keys?bot_id=bot_123`

**Update**
`PATCH /console/v1/api_keys/{id}`
```json
{
  "name": "prod-key-v2",
  "status": "active",
  "scopes": ["rag"],
  "api_versions": ["v1"],
  "quota_daily": 10000,
  "qps_limit": 20
}
```
> 未传字段保持不变，`quota_daily/qps_limit=0` 表示不限制。

**Rotate**
`POST /console/v1/api_keys/{id}/rotate`
返回：`api_key` + `raw_key`
> 轮换后旧 Key 进入“过渡期”，默认 60 分钟（可通过配置调整）。

**Usage Logs**
`GET /console/v1/api_usage?api_key_id=...&bot_id=...&api_version=v1&model=...&start_time=...&end_time=...`
返回字段包含 `path/api_version/model/status_code/latency_ms/token_usage/created_at`，并附带 `client_ip/user_agent` 便于审计。
> 未指定时间范围时默认查询最近 7 天。

**Usage Summary**
`GET /console/v1/api_usage/summary?api_key_id=...&bot_id=...&api_version=v1&model=...&start_time=...&end_time=...`
> 未指定时间范围时默认汇总最近 30 天。

**Usage Export**
`POST /console/v1/api_usage/export`
```json
{
  "api_key_id": "key_123",
  "bot_id": "bot_123",
  "start_time": "2026-02-01T00:00:00Z",
  "end_time": "2026-02-20T23:59:59Z",
  "format": "csv"
}
```
返回 `download_url/object_uri`（若配置对象存储）；未配置时返回 `content`（CSV 文本）。
> 未指定时间范围时默认导出最近 30 天。

### 4.6 统计看板
- `GET /console/v1/analytics/overview`
- `GET /console/v1/analytics/latency`
- `GET /console/v1/analytics/top_questions`
- `GET /console/v1/analytics/kb_gaps`

**Overview**
`GET /console/v1/analytics/overview?bot_id=...&start_time=...&end_time=...`
返回：总请求数、命中率、平均/95 分位延迟、错误率。
> 未指定时间范围时默认统计最近 7 天。

**Latency**
`GET /console/v1/analytics/latency?bot_id=...&start_time=...&end_time=...`
返回：按天聚合的平均/95 分位延迟与命中数。
> 未指定时间范围时默认统计最近 7 天。

**Top Questions**
`GET /console/v1/analytics/top_questions?bot_id=...&start_time=...&end_time=...&limit=20`
返回：热门问题（query + count + hit_rate）。
> 未指定时间范围时默认统计最近 7 天。

**KB Gaps**
`GET /console/v1/analytics/kb_gaps?bot_id=...&start_time=...&end_time=...&limit=20`
返回：疑似知识缺口（低命中 query + 计数）。
> 未指定时间范围时默认统计最近 7 天。

### 4.7 会话管理
- `GET /console/v1/sessions`（返回租户下所有会话，API Key 绑定 bot 无需显式传递 bot_id）
- `GET /console/v1/sessions/{id}/messages`

---

## 4. 安全与审计
- 请求必须记录：租户 ID、API Key、调用 IP、耗时
- 重要操作需审计日志（创建/删除 API Key）

