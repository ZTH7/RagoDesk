package service

import (
	biz "github.com/ZTH7/RAGDesk/apps/server/internal/rag/biz"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// RAGService handles rag service layer (placeholder)
type RAGService struct {
	uc  *biz.RAGUsecase
	log *log.Helper
}

// NewRAGService creates a new RAGService
func NewRAGService(uc *biz.RAGUsecase, logger log.Logger) *RAGService {
	return &RAGService{uc: uc, log: log.NewHelper(logger)}
}

// ProviderSet is rag service providers.
var ProviderSet = wire.NewSet(NewRAGService)
