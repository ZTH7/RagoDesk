# 技术栈说明（Tech Stack）— RAGDesk

> 目标：清晰展示“前后端 + RAG + 基础设施 + 可运维性”的完整技术体系，适合作为面试项目技术说明。

---

## 1. 前端技术栈

### 1.1 管理后台（Admin Web）
- **框架**：React 18
- **构建工具**：Vite
- **UI 组件库**：Ant Design
- **状态管理**：Zustand / Redux Toolkit（可选）
- **路由**：React Router
- **请求层**：Axios + 拦截器（JWT 注入）
- **图表**：ECharts / AntV（统计看板）

**用途**：企业后台管理（知识库、机器人、API Key、统计分析）

---

## 2. 后端技术栈（模块化单体）

### 2.1 服务框架
- **Kratos**：主干工程框架（配置、日志、指标、代码规范）
- **HTTP（Kratos Transport）**：对外 API
- **gRPC（可选）**：为后续微服务拆分预留

### 2.2 鉴权与安全
- **JWT**：用户登录鉴权
- **API Key**：企业对外接口访问鉴权
- **RBAC**：角色-权限控制
- **审计日志**：关键操作记录

### 2.3 并发与性能
- **Goroutine + Worker Pool**：高并发请求与任务处理
- **异步任务队列**：文档处理、统计聚合、向量化
- **限流/熔断**：网关 + 服务内中间件

---

## 3. RAG 与 AI 技术栈

### 3.1 RAG 引擎
- **Eino**：RAG Pipeline 构建
- **Embedding**：向量化生成
- **Retrieval**：TopK 召回
- **Rerank**：可选重排策略

### 3.2 向量数据库
- **Qdrant**（主选）
- 备选：Milvus / pgvector

---

## 4. 数据与缓存

- **MySQL**：核心业务数据（租户、会话、API 管理）
- **Redis**：会话状态、缓存、限流、分布式锁
- **对象存储**：文档与清洗结果（S3/MinIO）

---

## 5. 基础设施与部署

- **Docker**：本地开发与部署
- **Kubernetes (K8s)**：生产环境部署与弹性扩展
- **CI/CD**：GitHub Actions / GitLab CI
- **配置管理**：Kratos config / 环境变量

---

## 6. 可观测性

- **日志**：结构化日志（Zap / Kratos logger）
- **监控指标**：Prometheus + Grafana
- **链路追踪**：OpenTelemetry
- **审计日志**：API Key / 管理操作追踪

---

## 7. 典型技术亮点（面试可强调）

- **模块化单体**：Kratos + 清晰模块边界，便于演进微服务
- **多租户隔离**：tenant_id + RBAC 权限
- **RAG Pipeline**：文档清洗 → 向量化 → 召回 → Prompt
- **高并发**：Goroutine + 限流 + 任务队列
- **生产化能力**：K8s / Docker / 可观测性

---

## 8. 版本建议（可写入 README）
- Go 1.22+
- Kratos v2
- React 18
- Vite 5
- Ant Design 5
- Qdrant 1.x
