package data

import (
	"context"

	biz "github.com/ZTH7/RAGDesk/apps/server/internal/knowledge/biz"
	"github.com/ZTH7/RAGDesk/apps/server/internal/tenant"
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
	if _, err := tenant.RequireTenantID(ctx); err != nil {
		return err
	}
	return nil
}

func (r *knowledgeRepo) ListBotKB(ctx context.Context, botID string) ([]biz.BotKB, error) {
	if _, err := tenant.RequireTenantID(ctx); err != nil {
		return nil, err
	}
	return []biz.BotKB{}, nil
}

func (r *knowledgeRepo) EnsureIngestion(ctx context.Context, docVersionID string) (bool, error) {
	if _, err := tenant.RequireTenantID(ctx); err != nil {
		return false, err
	}
	return false, nil
}

// ProviderSet is knowledge data providers.
var ProviderSet = wire.NewSet(NewKnowledgeRepo)
