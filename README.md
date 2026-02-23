# RagoDesk
A Go-based multi-tenant support platform powered by RAG.

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
api/            # proto/openapi (empty until API v1 is defined)
configs/        # config templates
internal/
  iam/
  knowledge/
  rag/
  conversation/
  analytics/
  apimgmt/
  platform/
server/         # transport setup (HTTP/GRPC)
buf.yaml        # proto module config (buf)
buf.lock        # locked proto deps (buf)
```

## Quick Start

### Infrastructure (Docker)
```bash
cd deploy
# start mysql/redis/qdrant/rabbitmq/minio
docker compose up -d
```

### Service Initialization (Required)
- **MinIO**: create bucket `ragodesk` (or the name in `apps/server/configs/config.yaml` -> `data.object_storage.bucket`).
- **MySQL**: no manual setup required; backend auto-creates schema and seeds permissions on startup.
- **Qdrant**: no manual setup required; collection is created on first indexing.  
  If you see `Collection ... doesn't exist`, it means no document has been indexed yet—upload a document to trigger auto-create.
- **Redis/RabbitMQ**: no manual setup required.

### Backend (Kratos)
```bash
cd apps/server
go run ./cmd/ragodesk -conf ./configs
```

### Dashboard (React + Vite)
```bash
cd apps/dashboard
npm run dev
```

### Default Ports
- HTTP: `0.0.0.0:8000`
- gRPC: `0.0.0.0:9100`

### Outbound Proxy (LLM/Embedding)
Configure an HTTP proxy for outbound requests (OpenAI/DeepSeek):
```yaml
data:
  proxy: "http://127.0.0.1:10808"
```

### PDF Parsing (pure Go)
PDF parsing uses a pure Go extractor (`github.com/ledongthuc/pdf`) with a best-effort
fallback for raw text. No external binaries or OCR are required. Image-only/scanned
PDFs will not yield text without OCR and may result in empty content.

## Documentation
- [PRD](docs/PRD.md)
- [Development Plan](docs/DevelopmentPlan.md)
- [Architecture](docs/Architecture.md)
- [Modules](docs/Modules.md)
- [Data Model](docs/DataModel.md)
- [API Spec](docs/API.md)
- [Tech Stack](docs/TechStack.md)

## Notes
- Go module path: `github.com/ZTH7/RagoDesk/apps/server`
- Frontend uses npm + Vite + Ant Design
