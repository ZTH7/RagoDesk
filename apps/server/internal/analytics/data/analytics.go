package data

import (
	"context"

	biz "github.com/ZTH7/RAGDesk/apps/server/internal/analytics/biz"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

type analyticsRepo struct {
	log *log.Helper
}

// NewAnalyticsRepo creates a new analytics repo (placeholder)
func NewAnalyticsRepo(logger log.Logger) biz.AnalyticsRepo {
	return &analyticsRepo{log: log.NewHelper(logger)}
}

func (r *analyticsRepo) Ping(ctx context.Context) error {
	return nil
}

// ProviderSet is analytics data providers.
var ProviderSet = wire.NewSet(NewAnalyticsRepo)
