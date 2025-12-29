package data

import (
	"context"

	biz "github.com/ZTH7/RAGDesk/apps/server/internal/apimgmt/biz"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

type apimgmtRepo struct {
	log *log.Helper
}

// NewAPIMgmtRepo creates a new apimgmt repo (placeholder)
func NewAPIMgmtRepo(logger log.Logger) biz.APIMgmtRepo {
	return &apimgmtRepo{log: log.NewHelper(logger)}
}

func (r *apimgmtRepo) Ping(ctx context.Context) error {
	return nil
}

// ProviderSet is apimgmt data providers.
var ProviderSet = wire.NewSet(NewAPIMgmtRepo)
