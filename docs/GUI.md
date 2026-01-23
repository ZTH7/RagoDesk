# RagoDesk Dashboard (GUI) 设计稿

> 基于 **PRD.md / API.md / 代码实现** 的前端信息架构与页面设计。前端采用 **React + Vite + Ant Design**，并在 Ant Design 基础上做主题定制与交互规范增强。
> 设计分为两大分区：**Platform 管理区** 与 **Tenant Console 区**。

---

## 0. 设计目标与约束
- 面向企业级运维与开发者用户，强调可观测性、可追踪性、可审计性。
- 权限驱动导航，所有菜单与动作都必须经过 RBAC 控制。
- Console 区严格 tenant scope 隔离，Platform 区跨租户管理。
- 异步任务全链路可视化，文档 ingestion 的状态和错误必须可见。
- 以“标准工作流”为默认路径，RAG 关键步骤（例如 rerank）不可配置关闭。

---

## 1. 视觉系统（前端主题）
- 字体：`IBM Plex Sans`（正文/表单），`IBM Plex Mono`（代码/Key/ID）。
- 主色：`#1B4B66`，强调色：`#2BB3B1`，告警：`#D97706`，错误：`#DC2626`。
- 背景：`#F6F7FB`，分隔线：`#E6E8EF`，卡片：`#FFFFFF`。
- 圆角：6px；阴影：浅层（桌面端卡片与弹窗）。
- 布局：1440px 桌面宽，内容区最大宽度 1200px；移动端单列布局。
- 交互：表格与筛选区域固定高度，滚动区域独立，避免整页滚动抖动。

---

## 2. 全局布局与组件
- 顶部栏：Logo、当前用户、平台/控制台切换、语言切换、时间区。
- 侧边栏：按“Platform / Console”分组，权限驱动展示。
- 主内容区：列表、详情、表单、弹窗、任务状态中心。
- 全局通知：操作成功/失败 toast，API 错误带错误码与 trace_id。
- 全局搜索：仅在 Console 区启用，搜索对象为 Bot、KB、文档。
- 统一表格组件：支持列显隐、固定列、分页、服务端排序与筛选。
- 统一状态组件：Empty State、Loading Skeleton、Error State。

---

# PART A — Platform 管理区

## A1. 登录页（Platform）
- 登录字段：账号（email/phone）、密码。
- 登录成功后换取 JWT（后端暂无接口，UI 预留）。
- 错误态：账号不存在、密码错误、账号禁用。

## A2. 平台导航
- 租户管理
- 平台管理员
- 平台角色
- 权限目录
- 平台审计（预留）

## A3. 租户管理
**页面：租户列表 / 新建 / 详情**
- 列表列：id、name、type、plan、status、created_at。
- 筛选：type、plan、status、创建时间范围。
- 操作：新建、编辑 plan/status、进入详情。
- 详情页：Overview、绑定 bots 数量、API 使用统计（预留）。
- API：
  - `GET /platform/v1/tenants`
  - `POST /platform/v1/tenants`
  - `GET /platform/v1/tenants/{id}`

## A4. 平台管理员管理
**页面：管理员列表 / 新建管理员**
- 列表列：id、name、email、status、created_at。
- 操作：创建管理员、分配角色、禁用。
- API：
  - `GET /platform/v1/admins`
  - `POST /platform/v1/admins`
  - `POST /platform/v1/admins/{id}/roles`

## A5. 平台角色管理
**页面：角色列表 / 详情 / 授权权限**
- 列表列：id、name、created_at。
- 详情页：权限清单、授权记录。
- API：
  - `GET /platform/v1/roles`
  - `POST /platform/v1/roles`
  - `POST /platform/v1/roles/{id}/permissions`
  - `GET /platform/v1/roles/{id}/permissions`

## A6. 权限目录（平台）
**页面：权限列表**
- 列表列：code、name、scope、created_at。
- 筛选：scope=platform|tenant。
- API：
  - `GET /platform/v1/permissions`
  - `POST /platform/v1/permissions`

---

# PART B — Tenant Console 区

## B1. 注册 / 登录
- 注册：自助入驻入口（后端未实现，UI 预留）。
- 登录：租户管理员登录换取 JWT（后端未实现，UI 预留）。

## B2. Console 侧边栏（按权限展示）
- 仪表盘
- 统计分析（Analytics）
- 成员与角色
- 机器人（Bots）
- 知识库
- 文档管理
- API Keys
- 调用日志
- 会话管理
- 设置 / 个人中心

---

## B3. 仪表盘与分析
**目标：快速感知系统健康与效果**
- 核心指标卡：QPS、P95 延迟、成功率、检索命中率、日调用量。
- 趋势图：调用量趋势、平均延迟趋势、RAG 命中率趋势。
- Top 列表：Top Bots、Top API Keys、最近失败 ingestion。
- 时间范围：24h / 7d / 30d / 自定义。
- 路由：`/console/analytics/overview`（默认入口）。
- API：
  - `GET /console/v1/analytics/overview`
  - `GET /console/v1/analytics/latency`

### Top Questions
- 展示热门问题（query / count / hit_rate / last_seen_at）。
- 路由：`/console/analytics/top-questions`。
- API：`GET /console/v1/analytics/top_questions`

### KB Gaps
- 展示疑似知识缺口（query / miss_count / avg_confidence / last_seen_at）。
- 路由：`/console/analytics/kb-gaps`。
- API：`GET /console/v1/analytics/kb_gaps`

---

## B4. 成员与角色
### 成员管理
- 列表列：id、name、email、status、created_at。
- 操作：邀请、禁用、重置角色。
- API：
  - `POST /console/v1/tenants/{id}/users`
  - `GET /console/v1/tenants/{id}/users`

### 角色管理
- 列表列：id、name、created_at。
- 操作：创建角色、分配权限、为用户分配角色。
- API：
  - `POST /console/v1/roles`
  - `GET /console/v1/roles`
  - `POST /console/v1/roles/{id}/permissions`
  - `GET /console/v1/roles/{id}/permissions`
  - `POST /console/v1/users/{id}/roles`

### 权限目录（租户）
- 列表列：code、name、scope。
- API：`GET /console/v1/permissions`

---

## B5. 机器人（Bots）
**页面：列表 / 新建 / 详情**
- 列表列：id、name、status、created_at、kb_count、api_key_count。
- 操作：创建、编辑、删除。
- 详情页分区：
  - Overview：名称、状态、创建时间。
  - 绑定知识库：查看/新增/移除 KB。
  - API Keys：关联的 Key 列表（只读或跳转）。
  - RAG Pipeline：显示默认流水线与关键参数（只读）。
- API（规划）：
  - `POST /console/v1/bots`
  - `GET /console/v1/bots`
  - `PATCH /console/v1/bots/{id}`
  - `DELETE /console/v1/bots/{id}`
  - `POST /console/v1/bots/{id}/knowledge_bases`
  - `GET /console/v1/bots/{id}/knowledge_bases`
  - `DELETE /console/v1/bots/{id}/knowledge_bases/{kb_id}`

---

## B6. 知识库
### 知识库列表
- 列表列：id、name、description、document_count、created_at。
- 操作：新增、编辑、删除。
- 路由：`/console/knowledge-bases`，`/console/knowledge-bases/:id`。
- API：
  - `POST /console/v1/knowledge_bases`
  - `GET /console/v1/knowledge_bases`
  - `GET /console/v1/knowledge_bases/{id}`
  - `PATCH /console/v1/knowledge_bases/{id}`
  - `DELETE /console/v1/knowledge_bases/{id}`

### 文档管理
- 列表列：title、source_type、status、current_version、updated_at。
- 筛选：status、source_type、时间范围。
- 上传表单字段：kb_id、title、source_type、raw_uri。
- 操作：上传、删除、Reindex、Rollback。
- 路由：`/console/documents`，`/console/documents/:id`。
- 详情页：
  - 版本历史：version / status / created_at（来自 `GetDocumentResponse.versions`）。
  - Ingestion 时间线：uploaded → parsing → chunking → embedding → indexed。
  - 失败原因：error_message + 重试提示。
- API：
  - `POST /console/v1/documents/upload`
  - `GET /console/v1/documents`
  - `GET /console/v1/documents/{id}`
  - `DELETE /console/v1/documents/{id}`
  - `POST /console/v1/documents/{id}/reindex`
  - `POST /console/v1/documents/{id}/rollback`

---

## B7. API Key 管理
### API Keys 列表
- 列表列：name、bot、status、scopes、api_versions、quota_daily、qps_limit、created_at、last_used_at。
- 操作：创建、编辑、删除、轮换。
- 创建/轮换返回 raw_key（仅显示一次）。
- 详情页：展示 scopes / api_versions / 配额配置与最近使用情况。
- 路由：`/console/api-keys`，`/console/api-keys/:id`。

### Usage Logs / Summary / Export
- 过滤条件：api_key_id、bot_id、api_version、model、时间范围。
- 列表列：path、status_code、latency_ms、token_usage、created_at、client_ip、user_agent。
- 导出：CSV 下载或 OSS 对象链接。
- 路由：`/console/api-usage`，`/console/api-usage/summary`。
- API：
  - `POST /console/v1/api_keys`
  - `GET /console/v1/api_keys`
  - `PATCH /console/v1/api_keys/{id}`
  - `DELETE /console/v1/api_keys/{id}`
  - `POST /console/v1/api_keys/{id}/rotate`
  - `GET /console/v1/api_usage`
  - `GET /console/v1/api_usage/summary`
  - `POST /console/v1/api_usage/export`

---

## B8. 会话管理
### 会话列表
- 列表列：session_id、bot_id、status、close_reason、user_external_id、created_at。
- 筛选：bot_id、status、时间范围。
- 路由：`/console/sessions`。
- API：`GET /console/v1/sessions`

### 会话详情 / 消息记录
- 时间线视图：role/content/confidence/references/created_at。
- 支持引用来源折叠展示，便于审计。
- 路由：`/console/sessions/:id`。
- API：`GET /console/v1/sessions/{id}/messages`

---

## B9. 外部 API 调用（开发者视图）
- 使用说明：X-API-Key 方式调用。
- Quick Start：创建 bot → 绑定 KB → 创建 API key → 调用接口。
- API 调试：发送 `/api/v1/message`，显示请求与响应。

---

# 3. 交互与校验规范
- 表单校验：必填项、长度限制、枚举值验证（status/plan/scope）。
- 高风险操作二次确认：删除 KB、删除文档、删除 API Key。
- 分页默认：limit=50，支持自定义。
- 任务状态：异步任务需有“处理中”提示与失败原因。
- 访问控制：无权限时菜单隐藏、按钮禁用。

---

# 4. 关键流程
- Onboarding：创建 KB → 上传文档 → 创建 Bot → 绑定 KB → 创建 API Key → 开始调用。
- 文档更新：上传新版本 → 旧文档标记失效 → Qdrant 中仅保留最新向量。
- API Key 轮换：生成新 Key → 仅显示一次 → 旧 Key 进入过渡期 → 失效。

---

# 5. 待实现/前端预留
- 平台/租户登录注册接口。
- Bot CRUD 实际后端接口落地。
- 平台审计日志高级检索。

---

# 6. 路由与页面细化（建议实现清单）

## 6.1 Platform 区路由
- `/platform/login`：平台登录页（预留）。
- `/platform/tenants`：租户列表与筛选。
- `/platform/tenants/:id`：租户详情（基础信息 + 关联资源概览）。
- `/platform/admins`：平台管理员列表与创建。
- `/platform/admins/:id`：管理员详情（基础信息 + 角色列表）。
- `/platform/roles`：平台角色列表与创建。
- `/platform/roles/:id`：角色详情（权限树 / 已授权权限）。
- `/platform/permissions`：权限目录（scope 过滤）。

## 6.2 Console 区路由
- `/console/login`：租户登录页（预留）。
- `/console/analytics/overview`：统计总览。
- `/console/analytics/latency`：延迟趋势。
- `/console/analytics/top-questions`：热门问题。
- `/console/analytics/kb-gaps`：知识缺口。
- `/console/users`：成员管理。
- `/console/roles`：角色管理。
- `/console/permissions`：租户权限目录。
- `/console/bots`：机器人列表（规划）。
- `/console/bots/:id`：机器人详情（规划）。
- `/console/knowledge-bases`：知识库列表。
- `/console/knowledge-bases/:id`：知识库详情（基础信息 + 关联文档概览）。
- `/console/documents`：文档列表。
- `/console/documents/:id`：文档详情（版本与 ingestion 状态）。
- `/console/api-keys`：API Key 列表。
- `/console/api-keys/:id`：API Key 详情。
- `/console/api-usage`：API 调用日志。
- `/console/api-usage/summary`：API 用量汇总。
- `/console/sessions`：会话列表。
- `/console/sessions/:id`：会话详情与消息列表。
- `/console/devtools/api`：外部 API 调试工具（使用 X-API-Key）。
- `/console/profile`：个人中心（基本资料 / 退出登录）。

---

# 7. 页面结构与组件拆分（建议）

## 7.1 通用页面框架
- `PageHeader`：标题、描述、主操作按钮。
- `FilterBar`：时间范围、状态、关键词过滤。
- `DataTable`：列配置、服务端分页、行内操作。
- `DetailPanel`：详情卡片与信息分组。
- `ActionDrawer/Modal`：创建、编辑、授权等表单弹窗。

## 7.2 关键页面细化
### Platform 租户详情
- Tab：Overview / Roles / Usage（预留）。
- Overview：租户基本信息、状态、套餐、创建时间。
- Usage（预留）：聚合使用量与限额。

### Console 知识库详情
- Overview：基本信息、创建时间、文档数量。
- Documents：关联文档列表（可直接跳转）。

### Console 文档详情
- Info：标题、source_type、current_version、状态。
- Versions：版本历史（version / status / created_at）。
- Ingestion Timeline：阶段与错误原因。

### Console API Key 详情
- Info：名称、bot、status、scopes、api_versions。
- Limits：quota_daily、qps_limit。
- Usage Snapshot：最近调用次数、错误率（来自 summary）。

### Console 会话详情
- Message Timeline：user/assistant 消息流。
- References：引用来源折叠展示。
- Feedback：展示用户反馈（若有）。

---

# 8. 权限与路由绑定（补充）
- `/console/analytics/*` → `tenant.analytics.read`
- `/console/users` → `tenant.user.read`
- `/console/roles` → `tenant.role.read`
- `/console/permissions` → `tenant.permission.read`
- `/console/knowledge-bases*` → `tenant.knowledge_base.read`
- `/console/documents*` → `tenant.document.read`
- `/console/api-keys*` → `tenant.api_key.read`
- `/console/api-usage*` → `tenant.api_usage.read`
- `/console/sessions*` → `tenant.chat_session.read`

> 以上设计覆盖当前功能与已规划模块，并保持与后端架构一致。

