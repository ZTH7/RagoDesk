package data

import (
	"context"

	biz "github.com/ZTH7/RAGDesk/apps/server/internal/iam/biz"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

type iamRepo struct {
	log *log.Helper
}

// NewIAMRepo creates a new iam repo (placeholder)
func NewIAMRepo(logger log.Logger) biz.IAMRepo {
	return &iamRepo{log: log.NewHelper(logger)}
}

func (r *iamRepo) Ping(ctx context.Context) error {
	return nil
}

// ProviderSet is iam data providers.
var ProviderSet = wire.NewSet(NewIAMRepo)
