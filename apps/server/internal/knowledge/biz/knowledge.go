package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// Knowledge domain model (placeholder)
type Knowledge struct {
	ID string
}

// BotKB represents the bot to knowledge base relationship.
type BotKB struct {
	BotID    string
	KBID     string
	Priority int32
	Weight   int32
}

// KnowledgeRepo is a repository interface (placeholder)
type KnowledgeRepo interface {
	Ping(context.Context) error
	ListBotKB(ctx context.Context, botID string) ([]BotKB, error)
	EnsureIngestion(ctx context.Context, docVersionID string) (bool, error)
}

// KnowledgeUsecase handles knowledge business logic (placeholder)
type KnowledgeUsecase struct {
	repo KnowledgeRepo
	log  *log.Helper
}

// NewKnowledgeUsecase creates a new KnowledgeUsecase
func NewKnowledgeUsecase(repo KnowledgeRepo, logger log.Logger) *KnowledgeUsecase {
	return &KnowledgeUsecase{repo: repo, log: log.NewHelper(logger)}
}

// ProviderSet is knowledge biz providers.
var ProviderSet = wire.NewSet(NewKnowledgeUsecase)
