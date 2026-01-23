package service

import (
	biz "github.com/ZTH7/RagoDesk/apps/server/internal/platform/biz"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// PlatformService handles platform service layer (placeholder)
type PlatformService struct {
	uc  *biz.PlatformUsecase
	log *log.Helper
}

// NewPlatformService creates a new PlatformService
func NewPlatformService(uc *biz.PlatformUsecase, logger log.Logger) *PlatformService {
	return &PlatformService{uc: uc, log: log.NewHelper(logger)}
}

// ProviderSet is platform service providers.
var ProviderSet = wire.NewSet(NewPlatformService)
