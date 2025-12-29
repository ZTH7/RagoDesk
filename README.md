# RAGDesk
A Go-based multi-tenant support platform powered by RAG.

## Monorepo Structure
```
RAGDesk/
├── apps/
│   ├── server/        # Go backend (Kratos, modular monolith)
│   └── admin-web/     # React + Vite + Ant Design admin console
├── docs/              # PRD / Architecture / Data Model / API / Tech Stack
├── deploy/            # Docker / K8s manifests
├── tools/             # local dev tools/scripts (future)
├── go.work
└── go.work.sum
```

## Backend Module Layout (apps/server)
```
cmd/            # entrypoint (main)
internal/
  iam/
  knowledge/
  rag/
  conversation/
  handoff/
  analytics/
  apimgmt/
  platform/
pkg/            # reusable libs/middleware
api/            # proto/openapi
configs/        # config templates
migrations/     # DB migrations
scripts/        # admin scripts
test/           # integration tests
```

## Quick Start

### Infrastructure (Docker)
```bash
cd deploy
# start mysql/redis/qdrant
docker compose up -d
```

### Backend (Kratos)
```bash
cd apps/server
kratos run
```

### Admin Web (React + Vite)
```bash
cd apps/admin-web
npm run dev
```

## Documentation
- [PRD](docs/PRD.md)
- [Architecture](docs/Architecture.md)
- [Modules](docs/Modules.md)
- [Data Model](docs/DataModel.md)
- [API Spec](docs/API.md)
- [Tech Stack](docs/TechStack.md)

## Notes
- Go module path: `github.com/ZTH7/RAGDesk/apps/server`
- Frontend uses npm + Vite + Ant Design
