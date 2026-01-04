# RAG & Ingestion Design — RAGDesk

> 目标：把 `Knowledge & Ingestion` 与 `RAG Engine` 的边界、数据契约、向量库模型、以及 Eino 编排落到可实现的细节，避免“只有概念没有落地”。
>
> 本文也会区分：哪些是 **基础能力（Phase 2/3）**，哪些属于 **优化能力（单独的优化 Phase）**，用于排期与取舍。

---

## 1. 优先级划分（用于排期）

**基础能力（Phase 2/3）**
- 清晰模块边界：Ingestion 负责建索引与更新；RAG 负责在线检索与生成
- 最小可用的数据契约：chunk schema、向量库 payload schema、引用 refs schema
- Ingestion：上传 → 解析/清洗 → 切分 → 文档 embedding → upsert 向量库（按 `document_version_id` 幂等）
- RAG：query embedding → 向量检索 → Prompt → LLM → 返回 `answer + refs + confidence`（置信度低拒答/保守答复）
- 必要的超时与取消：embedding / 向量检索 / LLM 调用必须有 timeout（避免请求堆积）

**优化能力（单独优化 Phase）**
- 上线后的 CE（Continuous Evaluation）：在线采样、漂移检测、自动/半自动标注与回归
- 更完整的数据契约：增量更新策略、重建索引策略、删除/回滚的强约束与工具化
- 缓存策略：embedding cache / retrieval cache / response cache + 版本化失效
- 吞吐与延迟预算：分段 budget、并发上限、排队策略与降级开关
- 队列/重试/幂等的工程化：dead-letter、退避、去重键、任务可观测与人工重放
- hybrid + rerank：融合策略、alpha 权重、默认重排模型与可配置化
- Prompt registry 与 A/B：版本化、灰度、回滚、效果对比

---

## 2. 模块边界（Who owns what）

- `Knowledge & Ingestion`（离线/异步）：上传 → 解析/清洗 → 切分 → **文档 embedding** → upsert 向量库 → 索引/元数据更新 → 状态追踪与重试。
- `RAG Engine`（在线/同步）：输入问题 → **query embedding** → 检索/混合检索 → 可选重排 → Prompt 拼接 → **LLM 调用** → 返回 `answer + refs + confidence`。
- 交互关系：RAG Engine 只“读”向量索引与元数据；Ingestion 负责“写”和“更新/重建”索引。

---

## 3. Embedding 属于哪里？

- 文档 embedding：属于 `Knowledge & Ingestion`（构建索引的产物）。
- 查询 embedding：属于 `RAG Engine`（在线检索的输入）。
- 一致性约束：同一 KB/collection 内的向量必须来自同一套 embedding 模型与维度（`embedding_model + dim`）；切换模型需要新索引或全量重建。
- 批量与并发（优化项）：离线文档 embedding 优先做批量与并发控制；在线 query embedding 保持低延迟优先（批量通常收益不大）。

---

## 3.1 当前实现（Phase 2 已落地）

- 入口：`/admin/v1/documents/upload`（写 `document` + `document_version`，触发 ingestion）
- 执行方式：RabbitMQ 入队 + `apps/server/cmd/ingester` 消费（API 进程只负责入队）
- 解析/清洗：
- `text/markdown/html` 走清洗（HTML strip + 规范化空白）
- `url` 走 HTTP GET 拉取（HTML 自动 strip）
- `docx` 以 base64 传入，解析 `word/document.xml`
- `pdf`/`doc` 目前视为纯文本（待完善）
- Chunking：token-based 估算切分 + overlap（可通过环境变量配置）
- Embedding：默认 fake provider；支持 OpenAI 兼容 HTTP `/embeddings`
- 向量写入：Qdrant `upsert`，payload 包含 `tenant_id/kb_id/document_id/document_version_id/document_title/source_type/chunk_id/...`
- 删除：`DELETE /admin/v1/documents/{id}` 会清理 MySQL 元数据 + Qdrant points（按 `tenant_id` + `document_id` filter）

**当前可配置（env）**
- `RAGDESK_CHUNK_SIZE_TOKENS`
- `RAGDESK_CHUNK_OVERLAP_TOKENS`
- `RAGDESK_EMBEDDING_PROVIDER`（`fake`/`openai`）
- `RAGDESK_EMBEDDING_ENDPOINT`
- `RAGDESK_EMBEDDING_API_KEY`
- `RAGDESK_EMBEDDING_MODEL`
- `RAGDESK_EMBEDDING_DIM`
- `RAGDESK_EMBEDDING_TIMEOUT_MS`
- `RAGDESK_INGESTION_MAX_RETRIES`
- `RAGDESK_INGESTION_BACKOFF_MS`

---

## 4. 检索/索引的数据契约（chunk schema / metadata / 增量与重建）

- 目标：让“切分 → 向量化 → 检索 → 引用”可追溯、可删除、可重建；避免后期补字段导致返工。
- 最小可用契约（基础）：
- `chunk schema`（MySQL）：`tenant_id`, `kb_id`, `document_id`, `document_version_id`, `chunk_id`, `chunk_index`, `content`, `token_count`, `content_hash`, `language`, `created_at`
- `vector payload`（VectorDB）：`tenant_id`, `kb_id`, `document_id`, `document_version_id`, `document_title`, `source_type`, `chunk_id`, `chunk_index`, `token_count`, `content_hash`, `language`, `created_at`
- `refs schema`（用于引用来源）：`document_id`, `document_version_id`, `chunk_id`, `score`, `rank`, `snippet(optional)`
- Chunking 默认（基础）：固定窗口（token-based）+ overlap。
- Chunking 可配置项（优化）：`chunk_size_tokens`, `chunk_overlap_tokens`, `split_strategy`（fixed/semantic）, `min_chunk_tokens`, `max_chunk_tokens`，并允许按 KB/文档覆盖。
- 增量更新（优化）：document 更新生成新 `document_version_id`，索引采取 append-only；上线后再逐步完善“版本可见性”（例如只检索 latest version）。
- 重建索引（优化）：以下变化触发全量 rebuild 或新索引：embedding 模型/维度、chunking 策略、payload schema、hybrid/rerank 关键参数。

---

## 5. 向量库数据模型（以 Qdrant 为例）

- MVP 推荐：单 collection（例如 `ragdesk_chunks`）+ payload 强制过滤 `tenant_id` + `kb_id IN (...)`。
- 备选（更强隔离）：`collection per tenant` 或 `collection per kb`，优点是天然隔离，缺点是 collection 数量增多、生命周期管理更复杂。
- payload 字段（必须）：`tenant_id`, `kb_id`, `document_id`, `document_version_id`, `chunk_id`, `chunk_index`, `token_count`, `content_hash`, `language`, `created_at`。
- payload 字段（可选）：`tags`, `title`, `section`, `source_type`, `page_no`。
- 删除策略（基础）：按 `document_version_id` filter delete；回滚/重建索引走同一逻辑。
- 更新策略：document 更新必然生成新 `document_version_id`；避免“原地更新”导致引用不可追溯。

---

## 6. 检索策略：vector / hybrid / rerank（以及默认取舍）

- 基础（Phase 3）：先做“向量检索 + TopK”，保证引用与拒答策略可落地。
- 优化：hybrid 检索（dense + sparse/BM25）提升覆盖；在融合后对候选做 rerank 提升相关性。
- hybrid 的实现路径（优化）：
- 方案 A：向量库/检索引擎自带 hybrid 能力（实现成本低、但绑定能力边界）
- 方案 B：自建 sparse 检索（例如 BM25）+ dense 检索两路召回，自行做融合（RRF/加权融合）
- alpha 权重（优化）：作为可配置项（按 bot/kb），默认从 `0.5` 起步并用离线评测调参。
- rerank 默认（优化）：先用轻量交叉编码器 rerank（更稳定的延迟）；LLM rerank 作为可选高成本策略。
- 多 KB 并发（优化）：对多个 KB 并发 retrieve，之后做 merge/dedup，再进入 rerank/LLM。

---

## 7. 引用来源 + 置信度（基础能力，但可持续优化）

- 引用来源（基础）：返回 TopN chunk 的来源信息（`document_id/document_version_id/chunk_id`）+ 分数/排序，必要时带 snippet。
- 引用绑定（基础）：引用必须可回溯到“某个 document_version 的某个 chunk”，避免文档更新后引用漂移。
- 置信度（基础）：以检索分数/覆盖度/一致性为主要信号，给出 `confidence` 并设置阈值触发“拒答/保守答复”。
- 置信度优化：加入 rerank 分数、答案与引用一致性校验、以及基于真实反馈的校准（calibration）。

---

## 8. 吞吐、延迟预算、并发与限流（基础 + 优化）

- 为什么要拆分 budget：RAG 是多段链路（embedding/retrieve/rerank/llm），任何一段失控都会拖垮整体延迟与成本。
- 基础：为每段设置 timeout，并支持 request cancel（context deadline）。
- 基础：为关键资源设置并发上限（embedding、向量检索、LLM），避免雪崩。
- 优化：定义分段 latency budget（例如 embedding/retrieve/rerank/llm 各自的 p95 目标），并把指标落到可观测性。
- 优化：限流策略（按 tenant/api_key/bot），并在高压下做降级（跳过 rerank、降低 topK、返回保守答复）。
- LLM 是否异步：在线问答通常仍是同步返回（可做 streaming）；异步更适合“离线总结/质检/批处理”。

---

## 9. 缓存策略（优化 Phase）

- query embedding cache：key = `embedding_model + normalized_query`，TTL 短；降低重复问答成本。
- retrieval cache：key = `tenant_id + bot_id + kb_set + params + query_embedding_hash`，TTL 极短；命中可显著降延迟。
- response cache：只对“无个性化/无敏感上下文”的问答启用，并用 prompt/model 版本做 cache key。
- 失效机制：通过 `kb_index_version` 或 `document_version` 的变更触发失效，避免索引更新后返回旧结果。

---

## 10. 队列 & 重试 & 幂等（优化 Phase）

- ingestion 语义：队列通常是 at-least-once，需要“消费端幂等”。
- 去重键：按 `tenant_id + document_version_id + step` 做幂等；同一步骤重复执行应得到一致结果。
- embedding 失败：记录失败原因与重试次数；指数退避；超过阈值进入 dead-letter，支持人工重放。
- 向量写入失败：可重试的 upsert；失败时不推进 document_version 状态机。

---

## 11. Eino 在哪里用？怎么用？

- Eino 的价值：把 RAG 链路拆成可观测的节点（embedding/retrieve/rerank/prompt/llm），并在链路里统一做 tracing、耗时与成本统计。
- RAG Engine pipeline（建议节点）：`DetectLanguage` → `EmbedQuery` → `Retrieve(topK, per kb)` → `Merge & Dedup` → `Rerank(optional)` → `BuildPrompt` → `CallLLM` → `PostProcess` → `PersistMessage`。
- 并发点：多 KB 检索可并发；merge 后进入 rerank/LLM 串行。
- 观测点：每个节点记录 `latency_ms`、命中数量、TopK 分布、以及 LLM token usage（如果可获取）。
- Ingestion 是否用 Eino：可选。MVP 里更常见做法是 worker pipeline（队列 + step），等流程稳定再考虑统一到 Eino。

---

## 12. 离线评测 vs 上线后的 CE（Continuous Evaluation）

- 离线评测（优化 Phase）：用于“改动前/后”的可比对回归，优先从检索侧开始（Recall@K/MRR/过滤正确性）。
- CE（优化 Phase）：上线后基于真实流量做持续评估，包括采样、标注/弱监督、漂移检测、以及发布灰度对比。
- MVP 为什么仍建议做最小离线评测：MVP 阶段调参密集，缺少回归基线会导致“看起来改好了但其实变差”，上线/演示都不稳定。
- 数据集形式建议：`jsonl`（问题 + 期望命中的 doc/chunk 集合），从真实问题与知识库抽样构建小规模即可。
- 报告输出：`report.json` + 简单 diff（当前 vs 基线），可选接入 CI 做 smoke eval。
