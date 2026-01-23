package data

import (
	"context"

	biz "github.com/ZTH7/RagoDesk/apps/server/internal/platform/biz"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

type platformRepo struct {
	log *log.Helper
}

// NewPlatformRepo creates a new platform repo (placeholder)
func NewPlatformRepo(logger log.Logger) biz.PlatformRepo {
	return &platformRepo{log: log.NewHelper(logger)}
}

func (r *platformRepo) Ping(ctx context.Context) error {
	return nil
}

// ProviderSet is platform data providers.
var ProviderSet = wire.NewSet(NewPlatformRepo)
