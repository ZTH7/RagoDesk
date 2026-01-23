# RagoDesk
A Go-based multi-tenant support platform powered by RAG.

## Monorepo Structure
```
RagoDesk/
├── apps/
│   ├── server/        # Go backend (Kratos, modular monolith)
│   └── admin-web/     # React + Vite + Ant Design admin console
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

### Backend (Kratos)
```bash
cd apps/server
kratos run
```

### Ingestion Worker (optional)
```bash
cd apps/server
go run ./cmd/ingester
```

### Admin Web (React + Vite)
```bash
cd apps/admin-web
npm run dev
```

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
