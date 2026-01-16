package data

import (
	"context"
	"database/sql"

	biz "github.com/ZTH7/RAGDesk/apps/server/internal/apimgmt/biz"
	internaldata "github.com/ZTH7/RAGDesk/apps/server/internal/data"
	"github.com/ZTH7/RAGDesk/apps/server/internal/tenant"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

type apimgmtRepo struct {
	log *log.Helper
	db  *sql.DB
}

// NewAPIMgmtRepo creates a new apimgmt repo (placeholder)
func NewAPIMgmtRepo(data *internaldata.Data, logger log.Logger) biz.APIMgmtRepo {
	return &apimgmtRepo{log: log.NewHelper(logger), db: data.DB}
}

func (r *apimgmtRepo) Ping(ctx context.Context) error {
	if _, err := tenant.RequireTenantID(ctx); err != nil {
		return err
	}
	return nil
}

func (r *apimgmtRepo) RotateKey(ctx context.Context, keyID string) (biz.APIKey, error) {
	if _, err := tenant.RequireTenantID(ctx); err != nil {
		return biz.APIKey{}, err
	}
	return biz.APIKey{}, nil
}

func (r *apimgmtRepo) ValidateScope(ctx context.Context, keyID string, scope string) error {
	if _, err := tenant.RequireTenantID(ctx); err != nil {
		return err
	}
	return nil
}

func (r *apimgmtRepo) GetAPIKeyByHash(ctx context.Context, keyHash string) (biz.APIKey, error) {
	var key biz.APIKey
	if r.db == nil {
		return biz.APIKey{}, sql.ErrConnDone
	}
	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, tenant_id, bot_id, status FROM api_key WHERE key_hash = ? LIMIT 1`,
		keyHash,
	).Scan(&key.ID, &key.TenantID, &key.BotID, &key.Status)
	if err != nil {
		return biz.APIKey{}, err
	}
	return key, nil
}

// ProviderSet is apimgmt data providers.
var ProviderSet = wire.NewSet(NewAPIMgmtRepo)
