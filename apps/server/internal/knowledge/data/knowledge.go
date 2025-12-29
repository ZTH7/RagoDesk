package data

import (
	"context"

	biz "github.com/ZTH7/RAGDesk/apps/server/internal/knowledge/biz"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

type knowledgeRepo struct {
	log *log.Helper
}

// NewKnowledgeRepo creates a new knowledge repo (placeholder)
func NewKnowledgeRepo(logger log.Logger) biz.KnowledgeRepo {
	return &knowledgeRepo{log: log.NewHelper(logger)}
}

func (r *knowledgeRepo) Ping(ctx context.Context) error {
	return nil
}

// ProviderSet is knowledge data providers.
var ProviderSet = wire.NewSet(NewKnowledgeRepo)
