# API Specification — RAGDesk

> 对外 API 仅提供机器人能力与会话接口；管理后台 API 仅供企业/平台内部使用。
>
> 采用 **模块化单体**，对外通过 Kratos HTTP 提供 API；gRPC 作为未来拆分微服务的预留能力。

---

## 1. 通用约定

### 1.1 认证
- Header: `X-API-Key: <key>`
- 可选签名：`X-Timestamp`, `X-Nonce`, `X-Signature` (HMAC-SHA256)
- 服务端校验时间窗口与 nonce 防重放

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
  "bot_id": "bot_123",
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
    "status": "agent"
  }
}
```

---

### 2.4 用户反馈（用于优化）
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

## 3. 管理后台 API（企业内部）

### 3.1 机器人管理
- `POST /admin/v1/bots`
- `GET /admin/v1/bots`
- `PATCH /admin/v1/bots/{id}`
- `DELETE /admin/v1/bots/{id}`
- `GET /admin/v1/bots/{id}/knowledge_bases`
- `POST /admin/v1/bots/{id}/knowledge_bases`（绑定）
- `DELETE /admin/v1/bots/{id}/knowledge_bases/{kb_id}`（解绑）

### 3.2 知识库管理
- `POST /admin/v1/knowledge_bases`
- `GET /admin/v1/knowledge_bases`
- `POST /admin/v1/documents/upload`
- `GET /admin/v1/documents/{id}`
- `POST /admin/v1/documents/{id}/reindex`

### 3.3 API Key 管理
- `POST /admin/v1/api_keys`
- `GET /admin/v1/api_keys`
- `PATCH /admin/v1/api_keys/{id}`
- `DELETE /admin/v1/api_keys/{id}`
> 支持 scope 配置与 Key 轮换（可保留历史 Key 过渡期）

### 3.4 统计看板
- `GET /admin/v1/analytics/overview`
- `GET /admin/v1/analytics/latency`

---

## 4. 安全与审计
- 请求必须记录：租户 ID、API Key、调用 IP、耗时
- 重要操作需审计日志（创建/删除 API Key）
