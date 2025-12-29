package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// IAM domain model (placeholder)
type IAM struct {
	ID string
}

// IAMRepo is a repository interface (placeholder)
type IAMRepo interface {
	Ping(context.Context) error
}

// IAMUsecase handles iam business logic (placeholder)
type IAMUsecase struct {
	repo IAMRepo
	log  *log.Helper
}

// NewIAMUsecase creates a new IAMUsecase
func NewIAMUsecase(repo IAMRepo, logger log.Logger) *IAMUsecase {
	return &IAMUsecase{repo: repo, log: log.NewHelper(logger)}
}

// ProviderSet is iam biz providers.
var ProviderSet = wire.NewSet(NewIAMUsecase)
