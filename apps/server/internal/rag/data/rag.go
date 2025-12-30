package data

import (
	"context"

	biz "github.com/ZTH7/RAGDesk/apps/server/internal/rag/biz"
	"github.com/ZTH7/RAGDesk/apps/server/internal/tenant"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

type ragRepo struct {
	log *log.Helper
}

// NewRAGRepo creates a new rag repo (placeholder)
func NewRAGRepo(logger log.Logger) biz.RAGRepo {
	return &ragRepo{log: log.NewHelper(logger)}
}

func (r *ragRepo) Ping(ctx context.Context) error {
	if _, err := tenant.RequireTenantID(ctx); err != nil {
		return err
	}
	return nil
}

func (r *ragRepo) Search(ctx context.Context, req biz.VectorSearchRequest, mode biz.IsolationMode) ([]biz.VectorSearchResult, error) {
	if _, err := tenant.RequireTenantID(ctx); err != nil {
		return nil, err
	}
	return []biz.VectorSearchResult{}, nil
}

// ProviderSet is rag data providers.
var ProviderSet = wire.NewSet(NewRAGRepo)
