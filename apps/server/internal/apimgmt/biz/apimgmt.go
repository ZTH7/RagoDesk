package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// APIMgmt domain model (placeholder)
type APIMgmt struct {
	ID string
}

// APIKeyStatus represents api key lifecycle state.
type APIKeyStatus string

const (
	APIKeyStatusActive  APIKeyStatus = "active"
	APIKeyStatusRevoked APIKeyStatus = "revoked"
)

// APIKey represents an API key with scope and rotation metadata.
type APIKey struct {
	ID          string
	TenantID    string
	Scopes      []string
	Status      APIKeyStatus
	RotatedFrom string
	CreatedAt   time.Time
}

// APIMgmtRepo is a repository interface (placeholder)
type APIMgmtRepo interface {
	Ping(context.Context) error
	RotateKey(ctx context.Context, keyID string) (APIKey, error)
	ValidateScope(ctx context.Context, keyID string, scope string) error
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
