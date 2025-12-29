package data

import (
	"context"

	biz "github.com/ZTH7/RAGDesk/apps/server/internal/rag/biz"
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
	return nil
}

// ProviderSet is rag data providers.
var ProviderSet = wire.NewSet(NewRAGRepo)
