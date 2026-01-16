package biz

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
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
	BotID       string
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
	GetAPIKeyByHash(ctx context.Context, keyHash string) (APIKey, error)
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

const DefaultAPIKeyHeader = "X-API-Key"

func (uc *APIMgmtUsecase) ResolveAPIKey(ctx context.Context, rawKey string) (APIKey, error) {
	rawKey = strings.TrimSpace(rawKey)
	if rawKey == "" {
		return APIKey{}, errors.Unauthorized("API_KEY_MISSING", "api key missing")
	}
	key, err := uc.repo.GetAPIKeyByHash(ctx, hashAPIKey(rawKey))
	if err != nil {
		return APIKey{}, errors.Unauthorized("API_KEY_INVALID", "api key invalid")
	}
	if key.Status != APIKeyStatusActive || key.TenantID == "" || key.BotID == "" {
		return APIKey{}, errors.Unauthorized("API_KEY_INVALID", "api key invalid")
	}
	return key, nil
}

func hashAPIKey(rawKey string) string {
	sum := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(sum[:])
}

// ProviderSet is apimgmt biz providers.
var ProviderSet = wire.NewSet(NewAPIMgmtUsecase)
