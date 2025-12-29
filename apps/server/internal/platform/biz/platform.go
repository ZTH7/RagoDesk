package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// Platform domain model (placeholder)
type Platform struct {
	ID string
}

// PlatformRepo is a repository interface (placeholder)
type PlatformRepo interface {
	Ping(context.Context) error
}

// PlatformUsecase handles platform business logic (placeholder)
type PlatformUsecase struct {
	repo PlatformRepo
	log  *log.Helper
}

// NewPlatformUsecase creates a new PlatformUsecase
func NewPlatformUsecase(repo PlatformRepo, logger log.Logger) *PlatformUsecase {
	return &PlatformUsecase{repo: repo, log: log.NewHelper(logger)}
}

// ProviderSet is platform biz providers.
var ProviderSet = wire.NewSet(NewPlatformUsecase)
