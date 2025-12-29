package data

import (
	"context"

	biz "github.com/ZTH7/RAGDesk/apps/server/internal/conversation/biz"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

type conversationRepo struct {
	log *log.Helper
}

// NewConversationRepo creates a new conversation repo (placeholder)
func NewConversationRepo(logger log.Logger) biz.ConversationRepo {
	return &conversationRepo{log: log.NewHelper(logger)}
}

func (r *conversationRepo) Ping(ctx context.Context) error {
	return nil
}

// ProviderSet is conversation data providers.
var ProviderSet = wire.NewSet(NewConversationRepo)
