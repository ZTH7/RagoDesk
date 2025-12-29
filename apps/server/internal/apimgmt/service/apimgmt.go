package service

import (
	biz "github.com/ZTH7/RAGDesk/apps/server/internal/apimgmt/biz"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// APIMgmtService handles apimgmt service layer (placeholder)
type APIMgmtService struct {
	uc  *biz.APIMgmtUsecase
	log *log.Helper
}

// NewAPIMgmtService creates a new APIMgmtService
func NewAPIMgmtService(uc *biz.APIMgmtUsecase, logger log.Logger) *APIMgmtService {
	return &APIMgmtService{uc: uc, log: log.NewHelper(logger)}
}

// ProviderSet is apimgmt service providers.
var ProviderSet = wire.NewSet(NewAPIMgmtService)
