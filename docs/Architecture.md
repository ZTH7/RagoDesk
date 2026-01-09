# Architecture — RAGDesk（模块化单体）

> 设计原则：**模块化单体（Modular Monolith）** + 明确边界 + 可演进微服务。核心优先保证“可落地、可解释、可扩展”。

---

## 1. 高层架构

### 1.1 系统组件（逻辑视图）
- **API / HTTP Layer（Kratos HTTP）**
  - 外部 API + 管理后台入口
  - Auth、限流、审计、统一错误码
- **IAM / Tenant**
  - 多租户、用户、角色、权限
- **Knowledge & Ingestion**
  - 文档上传、清洗、分片、版本
  - 异步任务、向量化
  - 元数据抽取与索引构建（向量索引 + 元数据索引）
  - 维护机器人与知识库的关联（bot_kb）
- **RAG Engine**
  - Query Embedding → 向量检索 → 重排 → Prompt → LLM
  - Eino Pipeline 编排（检索、重排、Prompt、调用链）
  - Prompt 模板与系统指令控制
  - 引用来源 + 置信度
  - 评测指标与反馈闭环
  - 多语言与多模型路由（按租户/机器人/知识库配置）
- **Conversation**
  - 会话状态机（机器人 → 关闭）
  - 会话与消息记录
- **Analytics**
  - 指标聚合、日报/周报、看板
- **API Management**
  - API Key、配额、QPS、审计
  - 计费与用量统计（请求/Token/模型维度）
- **Observability**
  - 日志 / Metrics / Trace / 审计

### 1.2 技术基础设施
- **MySQL**：业务核心数据
- **Redis**：会话、限流、缓存、分布式锁
- **向量库**：Qdrant / pgvector
- **消息队列（RabbitMQ）**：异步文档处理 / 统计聚合
- **对象存储**：原始文档 / 处理结果

---

## 2. 组件交互与边界

> 目标：清晰描述“谁调用谁、数据怎么流动、哪些是同步/异步”。

### 2.1 模块职责与交互
- **API Layer**
  - 对外：`/api/v1/*`
  - 管理后台：`/admin/v1/*`
  - 统一鉴权（API Key/JWT）、限流、审计日志
- **IAM / Tenant**
  - 生产/消费 `tenant_id` 上下文
  - 为所有模块提供租户与权限校验能力
  - DAO 层强制 tenant 过滤（防止越权访问）
- **Knowledge & Ingestion**
  - Admin API 触发上传
  - 写 `document` / `document_version`
  - 异步任务：清洗、切分、向量化、入向量库
  - 索引数据契约：chunk 元信息 + 向量库 payload（按 `document_version` 幂等可重建，详见 `docs/RAG.md`）
  - 元数据抽取与索引构建（向量索引 + 元数据索引）
  - 维护 `bot_kb` 关联关系
  - 幂等与重试：按 `document_version` 保证可重入
- **RAG Engine**
  - 消费 `chat_message` 输入
  - 通过 `bot_kb` 解析可用知识库
  - 检索策略：按 `bot_kb` 权重/优先级组合召回
  - 调用向量库召回 + 混合检索（向量 + 关键词）+ 可选重排（多 KB 并发检索 → 合并/去重）
  - Prompt 模板 + 系统指令控制 + LLM
  - 输出 `answer + refs + confidence`
  - Eino Pipeline 负责链路编排与可观测性埋点（embedding/retrieve/rerank/prompt/llm 分步耗时与统计）
  - 成本与延迟优化：缓存、并发、批量 embedding
  - 多语言与多模型路由（模型选择、Prompt 选择、向量模型一致性校验）
- **Conversation**
  - 维护会话状态机
  - 会话与消息记录
  - 低置信度时采用保守答复或拒答策略
- **Analytics**
  - 接收事件（会话、召回）
  - 实时指标 + 离线聚合写入日报
- **API Management**
  - Key 生成、配额、限流
  - 调用日志审计
- **Observability**
  - Trace 链路贯穿所有模块
  - 关键操作记录审计日志

### 2.2 同步与异步边界
- **同步路径**：用户请求 → RAG → 回复（要求低延迟）
- **异步路径**：文档处理 / 统计聚合（高吞吐、可重试）
- **执行方式**：由独立 ingestion worker（`apps/server/cmd/ingester`）消费 RabbitMQ；API 进程设置 `RAGDESK_INGESTION_ASYNC=1` 仅负责入队
- **重试机制**：使用 retry queue（TTL + DLX）+ DLQ，按指数退避控制重试间隔

### 2.3 RAG 责任边界
- **Knowledge & Ingestion**：负责文档处理、切分、向量化、索引构建与更新；不参与在线生成。
- **RAG Engine**：负责检索、重排、Prompt 编排与生成；不直接修改知识库或索引。
- **Conversation**：负责会话状态与消息持久化；不承载检索/索引逻辑。

---

## 3. 关键流程

### 3.1 文档处理流程（Ingestion）
```mermaid
flowchart LR
A[上传文档] --> B[入库 document]
B --> C[任务执行（可配置：同步 / RabbitMQ 异步）]
C --> D[清洗/解析]
D --> E[切分 chunk]
E --> F[Embedding]
F --> G[向量库入库]
G --> H[状态更新]
```

### 3.2 对话流程（RAG）
```mermaid
sequenceDiagram
participant Client
participant API
participant RAG
participant VectorDB
participant LLM

Client->>API: POST /message
API->>RAG: query embedding
RAG->>VectorDB: topK retrieve
VectorDB-->>RAG: chunks
RAG->>LLM: prompt + chunks
LLM-->>RAG: answer
RAG->>API: answer + confidence + refs
API-->>Client: reply
```

### 3.3 Knowledge & RAG 细节架构（实现视图）

#### 3.3.1 Ingestion 处理链路
```mermaid
flowchart LR
  Client[Admin API] -->|Upload / Reindex| KSvc[Knowledge Service]
  KSvc -->|write metadata| MySQL[(MySQL)]
  KSvc -->|enqueue job| MQ[(RabbitMQ)]
  MQ --> Worker[Ingestion Worker]
  Worker --> Parser[Parse/Clean]
  Parser --> Chunker[Chunking]
  Chunker --> Embed[Embedding Provider]
  Embed --> Qdrant[(Qdrant)]
  Worker -->|write chunks/embedding meta| MySQL
  KSvc -->|raw content| Object[(Object Storage)]
```

#### 3.3.2 在线 RAG 链路
```mermaid
sequenceDiagram
participant Client
participant API
participant RAG
participant VectorDB
participant LLM

Client->>API: POST /api/v1/message
API->>RAG: resolve bot_kb + build prompt
RAG->>VectorDB: query embedding + retrieve
VectorDB-->>RAG: chunks + scores
RAG->>LLM: prompt + refs
LLM-->>RAG: answer
RAG->>API: answer + confidence + refs
API-->>Client: reply
```

#### 3.3.3 数据契约与内部参数（摘要）
- MySQL `doc_chunk`：`tenant_id`, `kb_id`, `document_id`, `document_version_id`, `chunk_id`, `chunk_index`, `content`, `token_count`, `content_hash`, `language`, `created_at`
- Qdrant payload：`tenant_id`, `kb_id`, `document_id`, `document_version_id`, `document_title`, `source_type`, `chunk_id`, `chunk_index`, `token_count`, `content_hash`, `language`, `created_at`
- Chunking（默认）：token-based 切分 + overlap
- 可配置参数：`RAGDESK_CHUNK_SIZE_TOKENS`, `RAGDESK_CHUNK_OVERLAP_TOKENS`, `RAGDESK_EMBEDDING_PROVIDER`, `RAGDESK_EMBEDDING_MODEL`, `RAGDESK_EMBEDDING_DIM`, `RAGDESK_EMBEDDING_ENDPOINT`, `RAGDESK_EMBEDDING_API_KEY`, `RAGDESK_EMBEDDING_TIMEOUT_MS`, `RAGDESK_EMBEDDING_BATCH_SIZE`, `RAGDESK_INGESTION_MAX_RETRIES`, `RAGDESK_INGESTION_BACKOFF_MS`
- 详细契约见 `docs/RAG.md`

### 3.4 统计分析流程
- 事件写入：会话/消息/召回
- 实时指标：Redis + Prometheus
- 离线聚合：定时作业写入 `analytics_daily`

### 3.5 RAG 评测与反馈闭环
- 采集用户反馈（点赞/踩/纠错）并关联 `chat_message`
- 形成评测集与质检任务（抽样/回放/自动评测）
- 指标看板：Recall@K、MRR、nDCG、Faithfulness、Answer Relevancy
- 反馈闭环：人工审核 → 知识修订 → 重新切分/向量化/索引

---

## 4. 系统边界与演进

### 4.1 模块化单体划分
建议按 **领域包/模块** 划分，确保边界清晰：
- `iam`（租户与权限）
- `knowledge`（文档与知识库）
- `rag`（召回与生成）
- `conversation`（会话与消息）
- `analytics`（统计）
- `apimgmt`（API Key 与限流）
- `platform`（系统管理）

### 4.2 未来微服务演进方向
可按吞吐和职责拆分：
- **Doc Processing Service**（异步处理）
- **RAG Service**（高并发推理）
- **Analytics Service**（批处理/流处理）

---

## 5. 安全与隔离设计
- 租户业务表强制 `tenant_id`；平台管理员与权限目录为全局表
- Query 层统一 `tenant filter`
- API Key + HMAC 签名（可选）
- 限流 / 配额 / 审计日志
- 向量库隔离策略：`collection per tenant` 或 `payload filter + tenant_id`

---

## 6. 可观测性
- **日志**：请求、响应、错误、审计
- **指标**：QPS、P95、命中率
- **Trace**：对话链路、RAG 召回链路

---

## 7. 部署与可用性
- 单体部署（Docker + k8s）
- Redis / MySQL / VectorDB 独立部署
- RabbitMQ 保障异步任务稳定

---

## 8. SLA 与性能目标
- P95 响应时间 < 2s
- 召回 TopK < 200ms
- 文档处理成功率 > 98%
