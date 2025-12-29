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
- **RAG Engine**
  - Query Embedding → 向量检索 → 重排 → Prompt → LLM
  - 引用来源 + 置信度
- **Conversation & Handoff**
  - 会话状态机（机器人 → 人工 → 关闭）
  - 工单/转人工逻辑
- **Analytics**
  - 指标聚合、日报/周报、看板
- **API Management**
  - API Key、配额、QPS、审计
- **Observability**
  - 日志 / Metrics / Trace / 审计

### 1.2 技术基础设施
- **MySQL**：业务核心数据
- **Redis**：会话、限流、缓存、分布式锁
- **向量库**：Qdrant / pgvector
- **消息队列**：异步文档处理 / 统计聚合
- **对象存储**：原始文档 / 处理结果

---

## 2. 关键流程

### 2.1 文档处理流程（Ingestion）
```mermaid
flowchart LR
A[上传文档] --> B[入库 document]
B --> C[异步任务 queue]
C --> D[清洗/解析]
D --> E[切分 chunk]
E --> F[Embedding]
F --> G[向量库入库]
G --> H[状态更新]
```

### 2.2 对话流程（RAG + Handoff）
```mermaid
sequenceDiagram
participant Client
participant API
participant RAG
participant VectorDB
participant LLM
participant Handoff

Client->>API: POST /message
API->>RAG: query embedding
RAG->>VectorDB: topK retrieve
VectorDB-->>RAG: chunks
RAG->>LLM: prompt + chunks
LLM-->>RAG: answer
RAG->>API: answer + confidence + refs
API-->>Client: reply

alt confidence < threshold
  API->>Handoff: create handoff request
  Handoff-->>Client: status=handoff
end
```

### 2.3 统计分析流程
- 事件写入：会话/消息/召回/转人工
- 实时指标：Redis + Prometheus
- 离线聚合：定时作业写入 `analytics_daily`

---

## 3. 系统边界与演进

### 3.1 模块化单体划分
建议按 **领域包/模块** 划分，确保边界清晰：
- `iam`（租户与权限）
- `knowledge`（文档与知识库）
- `rag`（召回与生成）
- `conversation`（会话与工单）
- `handoff`（人工客服接入）
- `analytics`（统计）
- `apimgmt`（API Key 与限流）
- `platform`（系统管理）

### 3.2 未来微服务演进方向
可按吞吐和职责拆分：
- **Doc Processing Service**（异步处理）
- **RAG Service**（高并发推理）
- **Analytics Service**（批处理/流处理）
- **Handoff Service**（客服系统）

---

## 4. 安全与隔离设计
- 所有表强制 `tenant_id`
- Query 层统一 `tenant filter`
- API Key + HMAC 签名（可选）
- 限流 / 配额 / 审计日志

---

## 5. 可观测性
- **日志**：请求、响应、错误、审计
- **指标**：QPS、P95、转人工率、命中率
- **Trace**：对话链路、RAG 召回链路

---

## 6. 部署与可用性
- 单体部署（Docker + k8s）
- Redis / MySQL / VectorDB 独立部署
- 消息队列保障异步任务稳定

---

## 7. SLA 与性能目标
- P95 响应时间 < 2s
- 召回 TopK < 200ms
- 文档处理成功率 > 98%
