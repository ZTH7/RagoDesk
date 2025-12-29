package service

import (
	biz "github.com/ZTH7/RAGDesk/apps/server/internal/conversation/biz"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ConversationService handles conversation service layer (placeholder)
type ConversationService struct {
	uc  *biz.ConversationUsecase
	log *log.Helper
}

// NewConversationService creates a new ConversationService
func NewConversationService(uc *biz.ConversationUsecase, logger log.Logger) *ConversationService {
	return &ConversationService{uc: uc, log: log.NewHelper(logger)}
}

// ProviderSet is conversation service providers.
var ProviderSet = wire.NewSet(NewConversationService)
