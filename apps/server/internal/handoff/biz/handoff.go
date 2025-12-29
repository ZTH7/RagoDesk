package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// Handoff domain model (placeholder)
type Handoff struct {
	ID string
}

// HandoffRepo is a repository interface (placeholder)
type HandoffRepo interface {
	Ping(context.Context) error
}

// HandoffUsecase handles handoff business logic (placeholder)
type HandoffUsecase struct {
	repo HandoffRepo
	log  *log.Helper
}

// NewHandoffUsecase creates a new HandoffUsecase
func NewHandoffUsecase(repo HandoffRepo, logger log.Logger) *HandoffUsecase {
	return &HandoffUsecase{repo: repo, log: log.NewHelper(logger)}
}

// ProviderSet is handoff biz providers.
var ProviderSet = wire.NewSet(NewHandoffUsecase)
