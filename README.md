# RagoDesk
Multi-tenant customer-support platform powered by RAG. Build, operate, and measure AI bots with a console for teams and a public API for end users.

## Product Overview
- Teams can upload knowledge, create bots, issue API keys, and monitor usage/quality.
- Developers integrate via a simple session/message API (API key bound to bot).
- Platform admins manage tenants, roles, and permissions.

## Core Capabilities
- **Console (JWT)**: tenants, users/roles, bots, knowledge bases, documents, API keys.
- **Public API (API key)**: sessions + messages for end-user chat.
- **RAG pipeline**: parse → clean → chunk → embed → index (Qdrant) → retrieve → rerank → LLM answer.
- **Async ingestion**: queue + retry with RabbitMQ (Redis fallback), running inside the main service.
- **Usage & analytics**: API usage logs and basic analytics dashboards.

## Architecture
- **Modular monolith** with Kratos HTTP + gRPC and clear domain modules: `iam`, `knowledge`, `rag`, `conversation`, `apimgmt`, `analytics`, `bot`.
- **Provider abstraction** for LLM/embedding (OpenAI/DeepSeek supported; proxy configurable).
- **Tenant isolation** at DAO layer + RBAC for console operations.

## Data & Storage
- **MySQL**: tenants, bots, knowledge, documents, sessions, usage logs, analytics.
- **MinIO (S3)**: raw document objects (only URI stored in DB).
- **Qdrant**: vector index for chunks.
- **RabbitMQ**: ingestion queue (Redis fallback).
- **Redis**: rate limiting + optional ingestion queue.

## Monorepo Structure
```
RagoDesk/
├── apps/
│   ├── server/        # Go backend (Kratos, modular monolith)
│   └── dashboard/     # React + Vite + Ant Design dashboard
├── docs/              # PRD / Architecture / Data Model / API / Tech Stack
├── deploy/            # Docker / K8s manifests
├── go.work
└── go.work.sum
```

## Backend Module Layout (apps/server)
```
cmd/            # entrypoint (main)
api/            # protobuf APIs (generated via buf)
configs/        # config templates
internal/
  iam/          # authn/authz + RBAC
  knowledge/    # KB, documents, ingestion
  rag/          # query pipeline + generation
  conversation/ # sessions/messages
  apimgmt/      # API keys + usage logs
  analytics/    # dashboards + aggregates
  bot/          # bot management
  middleware/   # auth/cors/logging/tracing
server/         # transport setup (HTTP/GRPC)
buf.yaml        # proto module config (buf)
```

## Quick Start

### 1) Infrastructure (Docker)
```bash
cd deploy
# start mysql/redis/qdrant/rabbitmq/minio
docker compose up -d
```

### 2) Required Initialization
- **MinIO**: create bucket `ragodesk` (or the name in `apps/server/configs/config.yaml` → `data.object_storage.bucket`).
- **MySQL**: no manual setup required; backend auto-creates schema and seeds permissions on startup.

### 3) Backend (Kratos)
```bash
cd apps/server
go run ./cmd/ragodesk -conf ./configs
```

### 4) Dashboard (React + Vite)
```bash
cd apps/dashboard
npm install
npm run dev
```

### Default Ports
- HTTP: `0.0.0.0:3000`
- gRPC: `0.0.0.0:3333`
- Dashboard: `http://localhost:5173`

## Configuration
Primary config file: `apps/server/configs/config.yaml`

Key sections:
- `data.database`: MySQL connection string
- `data.object_storage`: MinIO/S3 endpoint + bucket
- `data.vectordb`: Qdrant endpoint + collection
- `data.provider`: LLM/embedding provider + model
- `data.proxy`: outbound proxy for LLM/embedding
- `data.knowledge.ingestion`: async ingestion + retries

Sensitive keys:
- supported env vars: `OPENAI_API_KEY`, `DEEPSEEK_API_KEY`, `RAGODESK_API_KEY`
- optional local override file: `apps/server/configs/config.local.yaml`
- local template: `apps/server/configs/config.local.yaml.example`
- when `-conf` points to a directory, only `.yaml/.yml` files are loaded, and `config.local.yaml` overrides `config.yaml`

Dashboard config: `apps/dashboard/.env`

## PDF Parsing
PDF parsing uses `github.com/ledongthuc/pdf` with best-effort text extraction.
No external binaries or OCR are required. Image-only/scanned PDFs may yield empty text.

## Documentation
- [PRD](docs/PRD.md)
- [Development Plan](docs/DevelopmentPlan.md)
- [Architecture](docs/Architecture.md)
- [Modules](docs/Modules.md)
- [Data Model](docs/DataModel.md)
- [API Spec](docs/API.md)
- [Tech Stack](docs/TechStack.md)

## Status & Roadmap
Implementation follows the phased plan in `docs/DevelopmentPlan.md`.

## Troubleshooting
- `TENANT_MISSING`: console APIs require a valid JWT in `Authorization: Bearer <token>`.
- `Collection ... doesn't exist`: upload a document to trigger Qdrant collection creation.
- LLM timeouts: check outbound proxy and provider API key settings.

## Notes
- Go module path: `github.com/ZTH7/RagoDesk/apps/server`
- Frontend: React + Vite + Ant Design
