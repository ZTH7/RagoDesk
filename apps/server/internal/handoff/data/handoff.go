package data

import (
	"context"

	biz "github.com/ZTH7/RAGDesk/apps/server/internal/handoff/biz"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

type handoffRepo struct {
	log *log.Helper
}

// NewHandoffRepo creates a new handoff repo (placeholder)
func NewHandoffRepo(logger log.Logger) biz.HandoffRepo {
	return &handoffRepo{log: log.NewHelper(logger)}
}

func (r *handoffRepo) Ping(ctx context.Context) error {
	return nil
}

// ProviderSet is handoff data providers.
var ProviderSet = wire.NewSet(NewHandoffRepo)
