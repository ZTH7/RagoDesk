# RagoDesk Admin Web (GUI) 设计稿

> 基于当前 **PRD.md / API.md / 代码实现** 的前端信息架构与页面设计。前端采用 **React + Vite + Ant Design**。
> 设计分为两大分区：**Platform 管理区** 与 **Tenant Console 区**。

---

## 0. 全局设计原则
- **权限驱动导航**：根据 RBAC 权限动态显示菜单与操作按钮。
- **租户隔离**：Console 区所有数据必须在 tenant scope 内。
- **异步任务可观测**：文档上传与 ingestion 状态必须可视化。
- **审计可追溯**：API Key、会话、反馈等都支持检索和导出。

---

## 1. 全局布局（共用）
- **顶部栏**：Logo / 当前用户 / 退出 / 语言切换 / 时间区
- **侧边栏**：按“平台 / 控制台”分区展示菜单
- **内容区**：列表、表单、详情、弹窗
- **全局通知**：成功/失败 toast + API 错误提示

---

# PART A — Platform 管理区
> 平台管理员使用（Platform Admin），主要管理租户与平台权限。

## A1. 登录页（Platform）
- 字段：账号（email/phone）、密码
- 行为：登录成功后换取 JWT（后端暂无登录接口，UI 预留）
- 失败提示：账号无效 / 密码错误

## A2. 平台侧边栏
- **租户管理**
- **平台管理员**
- **平台角色**
- **权限目录**
- （可选）平台配置 / 审计日志

## A3. 租户管理
**页面：租户列表 / 新建 / 详情**
- **列表字段**：id / name / type / plan / status / created_at
- **操作**：新建、查看、编辑 plan/status
- **API**：
  - `GET /platform/v1/tenants`
  - `POST /platform/v1/tenants`
  - `GET /platform/v1/tenants/{id}`

## A4. 平台管理员管理
**页面：平台管理员列表 / 新建管理员**
- 字段：id / name / email / status / created_at
- 操作：创建管理员、分配角色
- **API**：
  - `GET /platform/v1/admins`
  - `POST /platform/v1/admins`
  - `POST /platform/v1/admins/{id}/roles`

## A5. 平台角色管理
**页面：角色列表 / 新建角色 / 授权权限**
- 字段：id / name
- 操作：新增角色、分配权限
- **API**：
  - `GET /platform/v1/roles`
  - `POST /platform/v1/roles`
  - `POST /platform/v1/roles/{id}/permissions`
  - `GET /platform/v1/roles/{id}/permissions`

## A6. 权限目录（平台）
**页面：权限列表**
- 支持过滤 `scope=platform|tenant`
- **API**：
  - `GET /platform/v1/permissions`
  - `POST /platform/v1/permissions`

---

# PART B — Tenant Console 区
> 企业租户管理后台（Tenant Admin / Supervisor / Developer）。

## B1. 注册 / 登录
- **注册**：若支持“自助入驻”则提供注册入口（当前 API 未实现，UI 预留）
- **登录**：租户管理员登录换取 JWT（后端暂无登录接口，UI 预留）

## B2. Console 侧边栏（按权限展示）
- **仪表盘**（命中率、延迟、QPS）
- **成员与角色**
- **机器人（Bots）**
- **知识库**
- **文档管理**
- **API Keys**
- **调用日志 / 导出**
- **会话管理**
- **设置 / 个人中心**

---

## B3. 成员与角色
### 1) 成员管理
- 列表字段：id / name / email / status / created_at
- 操作：邀请/禁用成员
- **API**：
  - `POST /console/v1/tenants/{id}/users`
  - `GET /console/v1/tenants/{id}/users`

### 2) 角色管理
- 列表字段：id / name
- 操作：创建角色、分配权限、为用户分配角色
- **API**：
  - `POST /console/v1/roles`
  - `GET /console/v1/roles`
  - `POST /console/v1/roles/{id}/permissions`
  - `GET /console/v1/roles/{id}/permissions`
  - `POST /console/v1/users/{id}/roles`

### 3) 权限目录（租户）
- 列表过滤 `scope=tenant`
- **API**：`GET /console/v1/permissions`

---

## B4. 机器人管理（Bots）
> API 已在文档中定义，但后端服务尚未落地，UI 先预留。
- 列表字段：id / name / status / created_at
- 操作：创建、编辑、删除
- 绑定知识库：关联 KB、设置权重
- **API（规划）**：
  - `POST /console/v1/bots`
  - `GET /console/v1/bots`
  - `PATCH /console/v1/bots/{id}`
  - `DELETE /console/v1/bots/{id}`
  - `POST /console/v1/bots/{id}/knowledge_bases`
  - `GET /console/v1/bots/{id}/knowledge_bases`
  - `DELETE /console/v1/bots/{id}/knowledge_bases/{kb_id}`

---

## B5. 知识库
### 1) 知识库列表
- 字段：id / name / description / created_at
- 操作：新增 / 编辑 / 删除
- **API**：
  - `POST /console/v1/knowledge_bases`
  - `GET /console/v1/knowledge_bases`
  - `GET /console/v1/knowledge_bases/{id}`
  - `PATCH /console/v1/knowledge_bases/{id}`
  - `DELETE /console/v1/knowledge_bases/{id}`

### 2) 文档管理
- 列表字段：title / source_type / status / current_version / updated_at
- 操作：上传 / 删除 / Reindex / Rollback
- 上传字段：`kb_id, title, source_type, raw_uri`
- **API**：
  - `POST /console/v1/documents/upload`
  - `GET /console/v1/documents`
  - `GET /console/v1/documents/{id}`
  - `DELETE /console/v1/documents/{id}`
  - `POST /console/v1/documents/{id}/reindex`
  - `POST /console/v1/documents/{id}/rollback`

### 3) Ingestion 状态可视化
- 状态：`uploaded/processing/ready/failed`
- 显示 error_message（失败原因）
- 提示“异步任务处理中”

---

## B6. API Key 管理
### 1) API Keys 列表
- 字段：name / bot / status / scopes / quota_daily / qps_limit / created_at / last_used_at
- 操作：创建 / 编辑 / 删除 / 轮换
- 创建/轮换返回 raw_key（仅显示一次）

### 2) Usage Logs / Summary / Export
- 过滤条件：api_key_id / bot_id / api_version / model / 时间范围
- 展示字段：path / status_code / latency_ms / token_usage / created_at / client_ip / user_agent
- 导出：CSV 下载或 OSS 对象链接

**API**：
- `POST /console/v1/api_keys`
- `GET /console/v1/api_keys`
- `PATCH /console/v1/api_keys/{id}`
- `DELETE /console/v1/api_keys/{id}`
- `POST /console/v1/api_keys/{id}/rotate`
- `GET /console/v1/api_usage`
- `GET /console/v1/api_usage/summary`
- `POST /console/v1/api_usage/export`

---

## B7. 会话管理
### 1) 会话列表
- 字段：session_id / bot_id / status / close_reason / user_external_id / created_at
- **API**：`GET /console/v1/sessions`

### 2) 会话详情 / 消息记录
- 展示消息列表（role/content/confidence/references/created_at）
- **API**：`GET /console/v1/sessions/{id}/messages`

---

## B8. 外部 API 调用（开发者视图）
- 展示 API Key 使用方式（X-API-Key）
- 简易 API 调试（发送 `/api/v1/message`）
- 引导接入流程（创建 bot → 绑定 KB → 创建 API key）

---

# 2. 权限与可见性（UI 规则）
- 每个菜单项绑定权限码，例如：
  - `tenant.knowledge_base.read` → 知识库列表
  - `tenant.document.upload` → 上传按钮
  - `tenant.api_key.write` → 创建 Key
  - `tenant.chat_session.read` → 会话管理
- 无权限时：菜单隐藏 + 按钮禁用

---

# 3. 交互细节与校验
- **表单校验**：必填项、长度限制、枚举值验证（status/plan/scope）
- **高风险操作二次确认**：删除 KB / 删除文档 / 删除 API Key
- **分页默认**：limit=50，支持自定义
- **状态提示**：loading / success / failed（尤其是 ingestion）

---

# 4. 待实现/前端预留
- 平台/租户的登录注册接口
- Bot CRUD 实际后端接口（目前仅文档定义）
- 统计看板 API（Phase 6）
- 审计日志高级检索

---

> 以上设计保持与当前 API 和代码一致，并为后续模块预留位置。