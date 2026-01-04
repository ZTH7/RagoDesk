# Development Plan — RAGDesk (Module-First Roadmap)

> Goal: ship a working MVP by first closing one end‑to‑end loop, then expanding each module.
> This roadmap matches the current modular monolith layout under `apps/server/internal/<module>`.

---

## MVP Target
- Multi‑tenant base (IAM)
- Knowledge ingestion pipeline
- RAG chat with citations + confidence
- Basic analytics
- External API + Admin API

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
- Admin API: create tenant, invite users, assign roles
- Auth: JWT for admin, context injection of `tenant_id`

**Deliverable**: tenant creation and member management flow.

---

## Phase 2: Knowledge & Ingestion
**Purpose**: content ingestion pipeline and storage.
- Tables: `knowledge_base`, `document`, `document_version`, `doc_chunk`, `embedding`
- Table: `bot_kb` (bot-to-knowledge-base association)
- Object storage adapter (S3/MinIO)
- Async ingestion job (RabbitMQ/Redis)
- Pipeline: clean → split → embed → upsert to vector DB
- Metadata extraction + index build (vector + metadata)
- Status tracking with retries
- Batch embedding + concurrency controls

**Deliverable**: documents upload and index successfully.

---

## Phase 3: RAG Engine
**Purpose**: answer generation with references and confidence.
- Vector store adapter (Qdrant/pgvector)
- Retriever + optional reranker (hybrid search)
- Prompt templates + system instruction control
- Eino pipeline orchestration
- RAG evaluation metrics (Recall@K, MRR, nDCG, Faithfulness)
- Confidence scorer with configurable threshold
- Feedback loop: user rating/correction → review → reindex
- Cost/latency: cache, concurrency, batch embedding
- Multi-model/multi-language routing and configuration

**Deliverable**: `POST /api/v1/message` returns reply + refs + confidence.

---

## Phase 4: Conversation
**Purpose**: manage conversation lifecycle and message history.
- Tables: `chat_session`, `chat_message`, `session_event`
- State machine: `bot → closed`
- Low‑confidence strategy: conservative answer or refusal

**Deliverable**: chat sessions with message history and audit trail.

---

## Phase 5: API Management
**Purpose**: external usage control.
- Tables: `api_key`, `api_usage_log`
- QPS + quota limits
- Audit logs + export
- Billing model + usage aggregation (per request/token/model)

**Deliverable**: key lifecycle + rate limits enforced.

---

## Phase 6: Analytics
**Purpose**: visibility and optimization.
- Event schema and collector
- Real‑time metrics (Redis/Prometheus)
- Daily aggregation: `analytics_daily`

**Deliverable**: hit rate and latency metrics.

---

## Phase 7: Admin Web
**Purpose**: operational UI.
- Pages: tenants, bots, KB/documents, API keys, analytics
- Form + list + charts

**Deliverable**: Admin UI working end‑to‑end.

---

## Phase 8: Stability & Operations
**Purpose**: make it production‑ready.
- Observability: logs/metrics/traces
- Retry + idempotency for ingestion and API calls
- Performance tests & tuning

**Deliverable**: stable release candidate.
