package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// Conversation domain model (placeholder)
type Conversation struct {
	ID string
}

// ConversationRepo is a repository interface (placeholder)
type ConversationRepo interface {
	Ping(context.Context) error
}

// ConversationUsecase handles conversation business logic (placeholder)
type ConversationUsecase struct {
	repo ConversationRepo
	log  *log.Helper
}

// NewConversationUsecase creates a new ConversationUsecase
func NewConversationUsecase(repo ConversationRepo, logger log.Logger) *ConversationUsecase {
	return &ConversationUsecase{repo: repo, log: log.NewHelper(logger)}
}

// ProviderSet is conversation biz providers.
var ProviderSet = wire.NewSet(NewConversationUsecase)
