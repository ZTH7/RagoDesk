package service

import (
	biz "github.com/ZTH7/RAGDesk/apps/server/internal/analytics/biz"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// AnalyticsService handles analytics service layer (placeholder)
type AnalyticsService struct {
	uc  *biz.AnalyticsUsecase
	log *log.Helper
}

// NewAnalyticsService creates a new AnalyticsService
func NewAnalyticsService(uc *biz.AnalyticsUsecase, logger log.Logger) *AnalyticsService {
	return &AnalyticsService{uc: uc, log: log.NewHelper(logger)}
}

// ProviderSet is analytics service providers.
var ProviderSet = wire.NewSet(NewAnalyticsService)
