# Development Plan — RAGDesk (Module-First Roadmap)

> Goal: ship a working MVP by first closing one end‑to‑end loop, then expanding each module.
> This roadmap matches the current modular monolith layout under `apps/server/internal/<module>`.

---

## MVP Target
- Multi‑tenant base (IAM)
- Knowledge ingestion pipeline
- RAG chat with citations + confidence
- Basic analytics
- External API + Console API

---

## Phase 0: Baseline Skeleton (1–2 days)
**Purpose**: clean template leftovers, align docs ↔ repo, and make the server boot without demo services.
- Remove Greeter template code and registrations
- Align `README.md` with actual repo structure
- Expand `conf.proto` / `config.yaml` to include: vector DB, RabbitMQ, object storage
- Add global middleware placeholders (auth, logging, tracing, error)
- Establish standard error codes (match `docs/API.md`)

**Deliverable**: service builds and runs with empty module endpoints.

---

## Phase 1: IAM Foundation (multi‑tenant + RBAC)
**Purpose**: all other modules depend on tenant context.
- Tables: `tenant`, `user`, `role`, `permission`, `user_role`, `role_permission`, `platform_admin`, `platform_role`, `platform_admin_role`, `platform_role_permission`
- Enforce `tenant_id` filter in data access
- Platform API: create tenant, manage platform roles/admins
- Console API: invite users, assign roles
- Auth: JWT for platform/console, context injection of `tenant_id`

**Deliverable**: tenant creation and member management flow.

---

## Phase 2: Knowledge & Ingestion
**Purpose**: content ingestion pipeline and storage.
- Tables: `knowledge_base`, `document`, `document_version`, `doc_chunk`, `embedding`
- Table: `bot_kb` (bot-to-knowledge-base association)
- Object storage adapter (S3/MinIO)
- Upload strategy: clients upload to OSS and pass `raw_uri` only (no raw text in MySQL)
- Async ingestion job (RabbitMQ/Redis) + worker skeleton
- Pipeline: upload → parse/clean → split → embed → upsert to vector DB
- Chunking/索引数据契约（MVP 版，详见 `docs/RAG.md`）
- Chunking策略：结构优先（按 block）+ 句子边界切分 + token 目标长度 + overlap（默认 max 800 / 10-15%）
- Metadata extraction (basic: title/section/page/source) + index build (vector + metadata)
- Status tracking + idempotency (by `document_version`)
- Console API: KB CRUD / document upload / reindex / rollback

**Deliverable**: documents upload and index successfully.

---

## Phase 3: RAG Engine
**Purpose**: answer generation with references and confidence.
- Vector store adapter (Qdrant only)
- Retriever (vector search, TopK configurable)
- Prompt templates + system instruction control + safety policy (basic, via system prompt)
- Eino pipeline orchestration + tracing hooks（详见 `docs/RAG.md`）
- Confidence scorer with configurable threshold + refusal strategy
- RAG API: message → retrieve → prompt → LLM → refs

**Deliverable**: `POST /api/v1/message` returns reply + refs + confidence.

---

## Phase 4: Conversation
**Purpose**: manage conversation lifecycle and message history.
- Tables: `chat_session`, `chat_message`, `session_event`
- State machine: `bot → closed` + close reason
- Message persistence + refs + confidence
- Feedback capture: `message_feedback`
- Low‑confidence strategy: conservative answer or refusal
- Console API: session/message listing (tenant scope)

**Deliverable**: chat sessions with message history and audit trail.

---

## Phase 5: API Management
**Purpose**: external usage control.
- Tables: `api_key`, `api_usage_log`
- API Key lifecycle: create/disable/rotate
- API Key scope & tenant binding
- QPS + quota limits (Redis) + per‑tenant throttling
- Audit logs + export
- Billing model + usage aggregation (per request/token/model)
- Usage report endpoints

**Deliverable**: key lifecycle + rate limits enforced.

---

## Phase 6: Analytics
**Purpose**: visibility and optimization.
- Event schema and collector (session/message/retrieval/feedback)
- Real‑time metrics (Redis/Prometheus)
- Daily aggregation: `analytics_daily`
- Dashboard queries: hit rate, latency, top questions, KB gaps

**Deliverable**: hit rate and latency metrics.

---

## Phase 7: Console Web
**Purpose**: operational UI.
- Pages: tenants, bots, KB/documents, API keys, analytics
- Role‑aware navigation (platform vs tenant)
- Form + list + charts
- Basic audit views

**Deliverable**: Console UI working end‑to‑end.

---

## Phase 8: Stability & Operations
**Purpose**: make it production‑ready.
- Observability: logs/metrics/traces
- Retry + idempotency for ingestion and API calls
- DB migrations (replace auto‑schema in code)
- Backups + data retention policy
- Performance tests & tuning
- Security hardening (rate limits, audit, secret rotation)

**Deliverable**: stable release candidate.

---

## Phase 9: RAG Optimization & Quality (Post-MVP)
**Purpose**: improve quality, cost, and reliability with measurable feedback loops.
- 上线后的 CE（Continuous Evaluation）：采样、标注/弱监督、漂移检测、灰度对比
- 离线回归评测：retrieval-first（Recall@K/MRR/过滤正确性）→ 再扩展生成侧指标
- 检索/索引数据契约增强：增量更新、可见性策略、重建索引工具化
- 引用来源 + 置信度增强：校准、引用一致性校验、拒答策略调优
- 吞吐与延迟预算：分段 budget、并发上限、降级开关（跳过 rerank/降低 topK）
- 并发与限流：embedding 批量、LLM 超时/重试策略、租户级限流
- 队列 & 重试 & 幂等：dead-letter、退避、去重键、任务可观测与人工重放
- 缓存策略：embedding/retrieval/response cache + 版本化失效
- hybrid + rerank：融合引擎选择、alpha 权重、默认重排模型、可配置化与评测驱动调参
- Prompt registry & A/B：版本化、灰度、回滚、效果对比
- 多语言/多模型路由：按租户/机器人配置，配合评测与成本控制
