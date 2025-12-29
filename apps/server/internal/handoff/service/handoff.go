package service

import (
	biz "github.com/ZTH7/RAGDesk/apps/server/internal/handoff/biz"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// HandoffService handles handoff service layer (placeholder)
type HandoffService struct {
	uc  *biz.HandoffUsecase
	log *log.Helper
}

// NewHandoffService creates a new HandoffService
func NewHandoffService(uc *biz.HandoffUsecase, logger log.Logger) *HandoffService {
	return &HandoffService{uc: uc, log: log.NewHelper(logger)}
}

// ProviderSet is handoff service providers.
var ProviderSet = wire.NewSet(NewHandoffService)
