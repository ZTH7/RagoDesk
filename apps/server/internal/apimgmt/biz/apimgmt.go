package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// APIMgmt domain model (placeholder)
type APIMgmt struct {
	ID string
}

// APIMgmtRepo is a repository interface (placeholder)
type APIMgmtRepo interface {
	Ping(context.Context) error
}

// APIMgmtUsecase handles apimgmt business logic (placeholder)
type APIMgmtUsecase struct {
	repo APIMgmtRepo
	log  *log.Helper
}

// NewAPIMgmtUsecase creates a new APIMgmtUsecase
func NewAPIMgmtUsecase(repo APIMgmtRepo, logger log.Logger) *APIMgmtUsecase {
	return &APIMgmtUsecase{repo: repo, log: log.NewHelper(logger)}
}

// ProviderSet is apimgmt biz providers.
var ProviderSet = wire.NewSet(NewAPIMgmtUsecase)
