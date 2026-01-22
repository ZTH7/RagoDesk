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

func (r *apimgmtRepo) GetAPIKeyByHash(ctx context.Context, keyHash string) (biz.APIKey, error) {
	var key biz.APIKey
	if r.db == nil {
		return biz.APIKey{}, sql.ErrConnDone
	}
	var scopesRaw sql.NullString
	var versionsRaw sql.NullString
	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, tenant_id, bot_id, status, scopes, api_versions, quota_daily, qps_limit
		FROM api_key
		WHERE key_hash = ?
			OR (prev_key_hash = ? AND prev_expires_at IS NOT NULL AND prev_expires_at > ?)
		LIMIT 1`,
		keyHash,
		keyHash,
		time.Now(),
	).Scan(&key.ID, &key.TenantID, &key.BotID, &key.Status, &scopesRaw, &versionsRaw, &key.QuotaDaily, &key.QPSLimit)
	if err != nil {
		return biz.APIKey{}, err
	}
	key.Scopes = decodeStringList(scopesRaw)
	key.APIVersions = decodeStringList(versionsRaw)
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
			(id, tenant_id, bot_id, name, key_hash, scopes, api_versions, status, quota_daily, qps_limit, created_at, last_used_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		key.ID,
		key.TenantID,
		key.BotID,
		key.Name,
		key.KeyHash,
		encodeStringList(key.Scopes),
		encodeStringList(key.APIVersions),
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
	var versionsRaw sql.NullString
	var lastUsedAt sql.NullTime
	err = r.db.QueryRowContext(
		ctx,
		`SELECT id, tenant_id, bot_id, name, key_hash, scopes, api_versions, status, quota_daily, qps_limit, created_at, last_used_at
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
		&versionsRaw,
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
	key.Scopes = decodeStringList(scopesRaw)
	key.APIVersions = decodeStringList(versionsRaw)
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
	query := `SELECT id, tenant_id, bot_id, name, key_hash, scopes, api_versions, status, quota_daily, qps_limit, created_at, last_used_at
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
		var versionsRaw sql.NullString
		var lastUsedAt sql.NullTime
		if err := rows.Scan(
			&key.ID,
			&key.TenantID,
			&key.BotID,
			&key.Name,
			&key.KeyHash,
			&scopesRaw,
			&versionsRaw,
			&key.Status,
			&key.QuotaDaily,
			&key.QPSLimit,
			&key.CreatedAt,
			&lastUsedAt,
		); err != nil {
			return nil, err
		}
		key.Scopes = decodeStringList(scopesRaw)
		key.APIVersions = decodeStringList(versionsRaw)
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
		`UPDATE api_key SET name = ?, status = ?, scopes = ?, api_versions = ?, quota_daily = ?, qps_limit = ?
		WHERE tenant_id = ? AND id = ?`,
		key.Name,
		key.Status,
		encodeStringList(key.Scopes),
		encodeStringList(key.APIVersions),
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

func (r *apimgmtRepo) RotateKey(ctx context.Context, keyID string, newHash string, graceUntil time.Time) (biz.APIKey, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return biz.APIKey{}, err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return biz.APIKey{}, err
	}
	var currentHash string
	err = tx.QueryRowContext(ctx, `SELECT key_hash FROM api_key WHERE tenant_id = ? AND id = ? FOR UPDATE`, tenantID, keyID).Scan(&currentHash)
	if err != nil {
		_ = tx.Rollback()
		return biz.APIKey{}, err
	}
	_, err = tx.ExecContext(
		ctx,
		`UPDATE api_key SET key_hash = ?, prev_key_hash = ?, prev_expires_at = ?, status = ?
		WHERE tenant_id = ? AND id = ?`,
		newHash,
		currentHash,
		nullTime(graceUntil),
		biz.APIKeyStatusActive,
		tenantID,
		keyID,
	)
	if err != nil {
		_ = tx.Rollback()
		return biz.APIKey{}, err
	}
	if err := tx.Commit(); err != nil {
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
		`INSERT INTO api_usage_log (id, tenant_id, bot_id, api_key_id, path, api_version, model, status_code, latency_ms, prompt_tokens, completion_tokens, total_tokens, client_ip, user_agent, created_at)
		VALUES (UUID(), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		log.TenantID,
		log.BotID,
		log.APIKeyID,
		log.Path,
		nullString(log.APIVersion),
		nullString(log.Model),
		log.StatusCode,
		log.LatencyMs,
		log.PromptTokens,
		log.CompletionTokens,
		log.TotalTokens,
		nullString(log.ClientIP),
		nullString(log.UserAgent),
		log.CreatedAt,
	)
	return err
}

func (r *apimgmtRepo) ListUsageLogs(ctx context.Context, filter biz.UsageFilter) ([]biz.UsageLog, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return nil, err
	}
	query := `SELECT id, api_key_id, tenant_id, bot_id, path, api_version, model, status_code, latency_ms, prompt_tokens, completion_tokens, total_tokens, client_ip, user_agent, created_at
		FROM api_usage_log
		WHERE tenant_id = ?`
	args := []any{tenantID}
	if strings.TrimSpace(filter.APIKeyID) != "" {
		query += " AND api_key_id = ?"
		args = append(args, strings.TrimSpace(filter.APIKeyID))
	}
	if strings.TrimSpace(filter.BotID) != "" {
		query += " AND bot_id = ?"
		args = append(args, strings.TrimSpace(filter.BotID))
	}
	if strings.TrimSpace(filter.APIVersion) != "" {
		query += " AND api_version = ?"
		args = append(args, strings.TrimSpace(filter.APIVersion))
	}
	if strings.TrimSpace(filter.Model) != "" {
		query += " AND model = ?"
		args = append(args, strings.TrimSpace(filter.Model))
	}
	if !filter.Start.IsZero() {
		query += " AND created_at >= ?"
		args = append(args, filter.Start)
	}
	if !filter.End.IsZero() {
		query += " AND created_at <= ?"
		args = append(args, filter.End)
	}
	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]biz.UsageLog, 0)
	for rows.Next() {
		var item biz.UsageLog
		var clientIP sql.NullString
		var userAgent sql.NullString
		if err := rows.Scan(
			&item.ID,
			&item.APIKeyID,
			&item.TenantID,
			&item.BotID,
			&item.Path,
			&item.APIVersion,
			&item.Model,
			&item.StatusCode,
			&item.LatencyMs,
			&item.PromptTokens,
			&item.CompletionTokens,
			&item.TotalTokens,
			&clientIP,
			&userAgent,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		if clientIP.Valid {
			item.ClientIP = clientIP.String
		}
		if userAgent.Valid {
			item.UserAgent = userAgent.String
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
		SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END),
		AVG(latency_ms),
		SUM(prompt_tokens),
		SUM(completion_tokens),
		SUM(total_tokens)
		FROM api_usage_log
		WHERE tenant_id = ?`
	args := []any{tenantID}
	if strings.TrimSpace(filter.APIKeyID) != "" {
		query += " AND api_key_id = ?"
		args = append(args, strings.TrimSpace(filter.APIKeyID))
	}
	if strings.TrimSpace(filter.BotID) != "" {
		query += " AND bot_id = ?"
		args = append(args, strings.TrimSpace(filter.BotID))
	}
	if strings.TrimSpace(filter.APIVersion) != "" {
		query += " AND api_version = ?"
		args = append(args, strings.TrimSpace(filter.APIVersion))
	}
	if strings.TrimSpace(filter.Model) != "" {
		query += " AND model = ?"
		args = append(args, strings.TrimSpace(filter.Model))
	}
	if !filter.Start.IsZero() {
		query += " AND created_at >= ?"
		args = append(args, filter.Start)
	}
	if !filter.End.IsZero() {
		query += " AND created_at <= ?"
		args = append(args, filter.End)
	}
	var summary biz.UsageSummary
	var avgLatency sql.NullFloat64
	var promptTokens sql.NullInt64
	var completionTokens sql.NullInt64
	var totalTokens sql.NullInt64
	err = r.db.QueryRowContext(ctx, query, args...).Scan(&summary.Total, &summary.ErrorCount, &avgLatency, &promptTokens, &completionTokens, &totalTokens)
	if err != nil {
		return biz.UsageSummary{}, err
	}
	if avgLatency.Valid {
		summary.AvgLatencyMs = avgLatency.Float64
	}
	if promptTokens.Valid {
		summary.PromptTokens = promptTokens.Int64
	}
	if completionTokens.Valid {
		summary.CompletionTokens = completionTokens.Int64
	}
	if totalTokens.Valid {
		summary.TotalTokens = totalTokens.Int64
	}
	return summary, nil
}

// ProviderSet is apimgmt data providers.
var ProviderSet = wire.NewSet(NewAPIMgmtRepo, NewRateLimiter, NewUsageExporter)

type rateLimiter struct {
	client           *redis.Client
	log              *log.Helper
	tenantQPSLimit   int32
	tenantQuotaDaily int32
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
	tenantQPS := int32(0)
	tenantQuota := int32(0)
	if cfg.Apimgmt != nil {
		if cfg.Apimgmt.TenantQpsLimit > 0 {
			tenantQPS = cfg.Apimgmt.TenantQpsLimit
		}
		if cfg.Apimgmt.TenantQuotaDaily > 0 {
			tenantQuota = cfg.Apimgmt.TenantQuotaDaily
		}
	}
	return &rateLimiter{client: client, log: log.NewHelper(logger), tenantQPSLimit: tenantQPS, tenantQuotaDaily: tenantQuota}
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
	}
	if l.tenantQPSLimit > 0 && key.TenantID != "" {
		window := now.Format("20060102150405")
		if err := l.checkLimit(ctx, "ragdesk:qps:tenant:"+key.TenantID+":"+window, int64(l.tenantQPSLimit), 2*time.Second, "API_TENANT_QPS_LIMIT", "tenant qps limit exceeded"); err != nil {
			return err
		}
	}
	if key.QuotaDaily > 0 {
		day := now.Format("20060102")
		if err := l.checkLimit(ctx, "ragdesk:quota:key:"+key.ID+":"+day, int64(key.QuotaDaily), 48*time.Hour, "API_QUOTA_LIMIT", "api key quota exceeded"); err != nil {
			return err
		}
	}
	if l.tenantQuotaDaily > 0 && key.TenantID != "" {
		day := now.Format("20060102")
		if err := l.checkLimit(ctx, "ragdesk:quota:tenant:"+key.TenantID+":"+day, int64(l.tenantQuotaDaily), 48*time.Hour, "API_TENANT_QUOTA_LIMIT", "tenant quota exceeded"); err != nil {
			return err
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
		return errors.New(429, code, message)
	}
	return nil
}

func nullTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}

func nullString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return strings.TrimSpace(value)
}

func encodeStringList(items []string) string {
	if len(items) == 0 {
		return ""
	}
	raw, err := json.Marshal(items)
	if err != nil {
		return strings.Join(items, ",")
	}
	return string(raw)
}

func decodeStringList(raw sql.NullString) []string {
	if !raw.Valid {
		return nil
	}
	value := strings.TrimSpace(raw.String)
	if value == "" {
		return nil
	}
	var items []string
	if err := json.Unmarshal([]byte(value), &items); err == nil && len(items) > 0 {
		return items
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
