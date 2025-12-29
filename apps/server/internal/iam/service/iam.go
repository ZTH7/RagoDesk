package service

import (
	biz "github.com/ZTH7/RAGDesk/apps/server/internal/iam/biz"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// IAMService handles iam service layer (placeholder)
type IAMService struct {
	uc  *biz.IAMUsecase
	log *log.Helper
}

// NewIAMService creates a new IAMService
func NewIAMService(uc *biz.IAMUsecase, logger log.Logger) *IAMService {
	return &IAMService{uc: uc, log: log.NewHelper(logger)}
}

// ProviderSet is iam service providers.
var ProviderSet = wire.NewSet(NewIAMService)
