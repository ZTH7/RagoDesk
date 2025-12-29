package service

import (
	biz "github.com/ZTH7/RAGDesk/apps/server/internal/knowledge/biz"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// KnowledgeService handles knowledge service layer (placeholder)
type KnowledgeService struct {
	uc  *biz.KnowledgeUsecase
	log *log.Helper
}

// NewKnowledgeService creates a new KnowledgeService
func NewKnowledgeService(uc *biz.KnowledgeUsecase, logger log.Logger) *KnowledgeService {
	return &KnowledgeService{uc: uc, log: log.NewHelper(logger)}
}

// ProviderSet is knowledge service providers.
var ProviderSet = wire.NewSet(NewKnowledgeService)
