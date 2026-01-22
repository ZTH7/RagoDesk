package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	biz "github.com/ZTH7/RAGDesk/apps/server/internal/apimgmt/biz"
	"github.com/ZTH7/RAGDesk/apps/server/internal/conf"
	internaldata "github.com/ZTH7/RAGDesk/apps/server/internal/data"
	"github.com/ZTH7/RAGDesk/apps/server/internal/tenant"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
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

func (r *apimgmtRepo) GetAPIKeyByHash(ctx context.Context, keyHash string) (biz.APIKey, error) {
	var key biz.APIKey
	if r.db == nil {
		return biz.APIKey{}, sql.ErrConnDone
	}
	var scopesRaw sql.NullString
	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, tenant_id, bot_id, status, scopes, quota_daily, qps_limit
		FROM api_key WHERE key_hash = ? LIMIT 1`,
		keyHash,
	).Scan(&key.ID, &key.TenantID, &key.BotID, &key.Status, &scopesRaw, &key.QuotaDaily, &key.QPSLimit)
	if err != nil {
		return biz.APIKey{}, err
	}
	key.Scopes = decodeScopes(scopesRaw)
	return key, nil
}

func (r *apimgmtRepo) CreateAPIKey(ctx context.Context, key biz.APIKey) (biz.APIKey, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return biz.APIKey{}, err
	}
	key.TenantID = tenantID
	if key.CreatedAt.IsZero() {
		key.CreatedAt = time.Now()
	}
	_, err = r.db.ExecContext(
		ctx,
		`INSERT INTO api_key
			(id, tenant_id, bot_id, name, key_hash, scopes, status, quota_daily, qps_limit, created_at, last_used_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		key.ID,
		key.TenantID,
		key.BotID,
		key.Name,
		key.KeyHash,
		encodeScopes(key.Scopes),
		key.Status,
		key.QuotaDaily,
		key.QPSLimit,
		key.CreatedAt,
		nullTime(key.LastUsedAt),
	)
	if err != nil {
		return biz.APIKey{}, err
	}
	return key, nil
}

func (r *apimgmtRepo) GetAPIKey(ctx context.Context, keyID string) (biz.APIKey, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return biz.APIKey{}, err
	}
	var key biz.APIKey
	var scopesRaw sql.NullString
	var lastUsedAt sql.NullTime
	err = r.db.QueryRowContext(
		ctx,
		`SELECT id, tenant_id, bot_id, name, key_hash, scopes, status, quota_daily, qps_limit, created_at, last_used_at
		FROM api_key WHERE tenant_id = ? AND id = ?`,
		tenantID,
		keyID,
	).Scan(
		&key.ID,
		&key.TenantID,
		&key.BotID,
		&key.Name,
		&key.KeyHash,
		&scopesRaw,
		&key.Status,
		&key.QuotaDaily,
		&key.QPSLimit,
		&key.CreatedAt,
		&lastUsedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return biz.APIKey{}, errors.NotFound("API_KEY_NOT_FOUND", "api key not found")
		}
		return biz.APIKey{}, err
	}
	key.Scopes = decodeScopes(scopesRaw)
	if lastUsedAt.Valid {
		key.LastUsedAt = lastUsedAt.Time
	}
	return key, nil
}

func (r *apimgmtRepo) ListAPIKeys(ctx context.Context, botID string, limit int, offset int) ([]biz.APIKey, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return nil, err
	}
	query := `SELECT id, tenant_id, bot_id, name, key_hash, scopes, status, quota_daily, qps_limit, created_at, last_used_at
		FROM api_key WHERE tenant_id = ?`
	args := []any{tenantID}
	if strings.TrimSpace(botID) != "" {
		query += " AND bot_id = ?"
		args = append(args, strings.TrimSpace(botID))
	}
	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]biz.APIKey, 0)
	for rows.Next() {
		var key biz.APIKey
		var scopesRaw sql.NullString
		var lastUsedAt sql.NullTime
		if err := rows.Scan(
			&key.ID,
			&key.TenantID,
			&key.BotID,
			&key.Name,
			&key.KeyHash,
			&scopesRaw,
			&key.Status,
			&key.QuotaDaily,
			&key.QPSLimit,
			&key.CreatedAt,
			&lastUsedAt,
		); err != nil {
			return nil, err
		}
		key.Scopes = decodeScopes(scopesRaw)
		if lastUsedAt.Valid {
			key.LastUsedAt = lastUsedAt.Time
		}
		items = append(items, key)
	}
	return items, rows.Err()
}

func (r *apimgmtRepo) UpdateAPIKey(ctx context.Context, key biz.APIKey) (biz.APIKey, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return biz.APIKey{}, err
	}
	_, err = r.db.ExecContext(
		ctx,
		`UPDATE api_key SET name = ?, status = ?, scopes = ?, quota_daily = ?, qps_limit = ?
		WHERE tenant_id = ? AND id = ?`,
		key.Name,
		key.Status,
		encodeScopes(key.Scopes),
		key.QuotaDaily,
		key.QPSLimit,
		tenantID,
		key.ID,
	)
	if err != nil {
		return biz.APIKey{}, err
	}
	return r.GetAPIKey(ctx, key.ID)
}

func (r *apimgmtRepo) DeleteAPIKey(ctx context.Context, keyID string) error {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return err
	}
	res, err := r.db.ExecContext(
		ctx,
		`DELETE FROM api_key WHERE tenant_id = ? AND id = ?`,
		tenantID,
		keyID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err == nil && rows == 0 {
		return errors.NotFound("API_KEY_NOT_FOUND", "api key not found")
	}
	return err
}

func (r *apimgmtRepo) RotateKey(ctx context.Context, keyID string, newHash string) (biz.APIKey, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return biz.APIKey{}, err
	}
	_, err = r.db.ExecContext(
		ctx,
		`UPDATE api_key SET key_hash = ?, status = ?
		WHERE tenant_id = ? AND id = ?`,
		newHash,
		biz.APIKeyStatusActive,
		tenantID,
		keyID,
	)
	if err != nil {
		return biz.APIKey{}, err
	}
	return r.GetAPIKey(ctx, keyID)
}

func (r *apimgmtRepo) UpdateLastUsedAt(ctx context.Context, keyID string, lastUsedAt time.Time) error {
	if keyID == "" {
		return nil
	}
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE api_key SET last_used_at = ? WHERE id = ?`,
		lastUsedAt,
		keyID,
	)
	return err
}

func (r *apimgmtRepo) CreateUsageLog(ctx context.Context, log biz.UsageLog) error {
	if log.APIKeyID == "" {
		return nil
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO api_usage_log (id, api_key_id, path, status_code, latency_ms, created_at)
		VALUES (UUID(), ?, ?, ?, ?, ?)`,
		log.APIKeyID,
		log.Path,
		log.StatusCode,
		log.LatencyMs,
		log.CreatedAt,
	)
	return err
}

func (r *apimgmtRepo) ListUsageLogs(ctx context.Context, filter biz.UsageFilter) ([]biz.UsageLog, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return nil, err
	}
	query := `SELECT l.id, l.api_key_id, k.bot_id, l.path, l.status_code, l.latency_ms, l.created_at
		FROM api_usage_log l
		JOIN api_key k ON l.api_key_id = k.id
		WHERE k.tenant_id = ?`
	args := []any{tenantID}
	if strings.TrimSpace(filter.APIKeyID) != "" {
		query += " AND l.api_key_id = ?"
		args = append(args, strings.TrimSpace(filter.APIKeyID))
	}
	if strings.TrimSpace(filter.BotID) != "" {
		query += " AND k.bot_id = ?"
		args = append(args, strings.TrimSpace(filter.BotID))
	}
	if !filter.Start.IsZero() {
		query += " AND l.created_at >= ?"
		args = append(args, filter.Start)
	}
	if !filter.End.IsZero() {
		query += " AND l.created_at <= ?"
		args = append(args, filter.End)
	}
	query += " ORDER BY l.created_at DESC LIMIT ? OFFSET ?"
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]biz.UsageLog, 0)
	for rows.Next() {
		var item biz.UsageLog
		if err := rows.Scan(
			&item.ID,
			&item.APIKeyID,
			&item.BotID,
			&item.Path,
			&item.StatusCode,
			&item.LatencyMs,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *apimgmtRepo) GetUsageSummary(ctx context.Context, filter biz.UsageFilter) (biz.UsageSummary, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return biz.UsageSummary{}, err
	}
	query := `SELECT COUNT(*),
		SUM(CASE WHEN l.status_code >= 400 THEN 1 ELSE 0 END),
		AVG(l.latency_ms)
		FROM api_usage_log l
		JOIN api_key k ON l.api_key_id = k.id
		WHERE k.tenant_id = ?`
	args := []any{tenantID}
	if strings.TrimSpace(filter.APIKeyID) != "" {
		query += " AND l.api_key_id = ?"
		args = append(args, strings.TrimSpace(filter.APIKeyID))
	}
	if strings.TrimSpace(filter.BotID) != "" {
		query += " AND k.bot_id = ?"
		args = append(args, strings.TrimSpace(filter.BotID))
	}
	if !filter.Start.IsZero() {
		query += " AND l.created_at >= ?"
		args = append(args, filter.Start)
	}
	if !filter.End.IsZero() {
		query += " AND l.created_at <= ?"
		args = append(args, filter.End)
	}
	var summary biz.UsageSummary
	var avgLatency sql.NullFloat64
	err = r.db.QueryRowContext(ctx, query, args...).Scan(&summary.Total, &summary.ErrorCount, &avgLatency)
	if err != nil {
		return biz.UsageSummary{}, err
	}
	if avgLatency.Valid {
		summary.AvgLatencyMs = avgLatency.Float64
	}
	return summary, nil
}

// ProviderSet is apimgmt data providers.
var ProviderSet = wire.NewSet(NewAPIMgmtRepo, NewRateLimiter)

type rateLimiter struct {
	client *redis.Client
	log    *log.Helper
}

func NewRateLimiter(cfg *conf.Data, logger log.Logger) biz.RateLimiter {
	if cfg == nil || cfg.Redis == nil || cfg.Redis.Addr == "" {
		return nil
	}
	options := &redis.Options{
		Addr:         cfg.Redis.Addr,
		ReadTimeout:  cfg.Redis.ReadTimeout.AsDuration(),
		WriteTimeout: cfg.Redis.WriteTimeout.AsDuration(),
	}
	client := redis.NewClient(options)
	if err := client.Ping(context.Background()).Err(); err != nil {
		log.NewHelper(logger).Warnf("redis ping failed for rate limiter: %v", err)
		return nil
	}
	return &rateLimiter{client: client, log: log.NewHelper(logger)}
}

func (l *rateLimiter) Check(ctx context.Context, key biz.APIKey) error {
	if l == nil || l.client == nil {
		return nil
	}
	now := time.Now()
	if key.QPSLimit > 0 {
		window := now.Format("20060102150405")
		if err := l.checkLimit(ctx, "ragdesk:qps:key:"+key.ID+":"+window, int64(key.QPSLimit), 2*time.Second, "API_QPS_LIMIT", "api key qps limit exceeded"); err != nil {
			return err
		}
		if key.TenantID != "" {
			if err := l.checkLimit(ctx, "ragdesk:qps:tenant:"+key.TenantID+":"+window, int64(key.QPSLimit), 2*time.Second, "API_TENANT_QPS_LIMIT", "tenant qps limit exceeded"); err != nil {
				return err
			}
		}
	}
	if key.QuotaDaily > 0 {
		day := now.Format("20060102")
		if err := l.checkLimit(ctx, "ragdesk:quota:key:"+key.ID+":"+day, int64(key.QuotaDaily), 48*time.Hour, "API_QUOTA_LIMIT", "api key quota exceeded"); err != nil {
			return err
		}
		if key.TenantID != "" {
			if err := l.checkLimit(ctx, "ragdesk:quota:tenant:"+key.TenantID+":"+day, int64(key.QuotaDaily), 48*time.Hour, "API_TENANT_QUOTA_LIMIT", "tenant quota exceeded"); err != nil {
				return err
			}
		}
	}
	return nil
}

func (l *rateLimiter) checkLimit(ctx context.Context, key string, limit int64, ttl time.Duration, code string, message string) error {
	if limit <= 0 {
		return nil
	}
	count, err := l.client.Incr(ctx, key).Result()
	if err != nil {
		if l.log != nil {
			l.log.Warnf("limiter incr failed: %v", err)
		}
		return nil
	}
	if count == 1 {
		_ = l.client.Expire(ctx, key, ttl).Err()
	}
	if count > limit {
		return errors.TooManyRequests(code, message)
	}
	return nil
}

func nullTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}

func encodeScopes(scopes []string) string {
	if len(scopes) == 0 {
		return ""
	}
	raw, err := json.Marshal(scopes)
	if err != nil {
		return strings.Join(scopes, ",")
	}
	return string(raw)
}

func decodeScopes(raw sql.NullString) []string {
	if !raw.Valid {
		return nil
	}
	value := strings.TrimSpace(raw.String)
	if value == "" {
		return nil
	}
	var scopes []string
	if err := json.Unmarshal([]byte(value), &scopes); err == nil && len(scopes) > 0 {
		return scopes
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, item := range parts {
		item = strings.TrimSpace(item)
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}
