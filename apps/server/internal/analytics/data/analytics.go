package data

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	biz "github.com/ZTH7/RAGDesk/apps/server/internal/analytics/biz"
	"github.com/ZTH7/RAGDesk/apps/server/internal/conf"
	internaldata "github.com/ZTH7/RAGDesk/apps/server/internal/data"
	"github.com/ZTH7/RAGDesk/apps/server/internal/tenant"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
)

const metricsTTL = 7 * 24 * time.Hour

type analyticsRepo struct {
	log   *log.Helper
	db    *sql.DB
	redis *redis.Client
}

// NewAnalyticsRepo creates a new analytics repo.
func NewAnalyticsRepo(data *internaldata.Data, cfg *conf.Data, logger log.Logger) biz.AnalyticsRepo {
	repo := &analyticsRepo{log: log.NewHelper(logger)}
	if data != nil {
		repo.db = data.DB
	}
	if cfg != nil && cfg.Redis != nil && cfg.Redis.Addr != "" {
		repo.redis = redis.NewClient(&redis.Options{
			Addr:         cfg.Redis.Addr,
			ReadTimeout:  cfg.Redis.ReadTimeout.AsDuration(),
			WriteTimeout: cfg.Redis.WriteTimeout.AsDuration(),
		})
	}
	return repo
}

func (r *analyticsRepo) CreateEvent(ctx context.Context, event biz.AnalyticsEvent) error {
	if r == nil || r.db == nil {
		return sql.ErrConnDone
	}
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return err
	}
	if event.TenantID == "" {
		event.TenantID = tenantID
	}
	if event.TenantID != tenantID {
		return errors.Forbidden("TENANT_MISMATCH", "tenant mismatch")
	}
	if event.ID == "" {
		event.ID = uuid.NewString()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}
	_, err = r.db.ExecContext(
		ctx,
		`INSERT INTO analytics_event
			(id, tenant_id, bot_id, event_type, session_id, message_id, query, hit, confidence, latency_ms, status_code, rating, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		event.ID,
		event.TenantID,
		strings.TrimSpace(event.BotID),
		strings.TrimSpace(event.EventType),
		nullString(event.SessionID),
		nullString(event.MessageID),
		nullString(event.Query),
		boolToInt(event.Hit),
		event.Confidence,
		event.LatencyMs,
		event.StatusCode,
		nullInt32(event.Rating),
		event.CreatedAt,
	)
	if err != nil {
		return err
	}
	if r.redis != nil && event.EventType == biz.EventRAGQuery {
		r.recordRealtime(ctx, event)
	}
	return nil
}

func (r *analyticsRepo) GetOverview(ctx context.Context, filter biz.AnalyticsFilter) (biz.OverviewStats, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return biz.OverviewStats{}, err
	}
	query := `SELECT COUNT(*),
		SUM(CASE WHEN hit = 1 THEN 1 ELSE 0 END),
		AVG(latency_ms),
		SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END)
		FROM analytics_event
		WHERE tenant_id = ? AND event_type = ?`
	args := []any{tenantID, biz.EventRAGQuery}
	query, args = applyEventFilters(query, args, filter)
	var summary biz.OverviewStats
	var avgLatency sql.NullFloat64
	var hitCount sql.NullInt64
	var errorCount sql.NullInt64
	var total int64
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&total, &hitCount, &avgLatency, &errorCount); err != nil {
		return biz.OverviewStats{}, err
	}
	summary.Total = total
	if hitCount.Valid {
		summary.HitCount = hitCount.Int64
	}
	if avgLatency.Valid {
		summary.AvgLatencyMs = avgLatency.Float64
	}
	if errorCount.Valid {
		summary.ErrorCount = errorCount.Int64
	}
	if summary.Total > 0 {
		summary.HitRate = float64(summary.HitCount) / float64(summary.Total)
		summary.ErrorRate = float64(summary.ErrorCount) / float64(summary.Total)
	}
	if summary.Total > 0 {
		offset := int64(math.Ceil(float64(summary.Total)*0.95)) - 1
		if offset < 0 {
			offset = 0
		}
		p95Query := `SELECT latency_ms FROM analytics_event
			WHERE tenant_id = ? AND event_type = ?`
		p95Args := []any{tenantID, biz.EventRAGQuery}
		p95Query, p95Args = applyEventFilters(p95Query, p95Args, filter)
		p95Query += " ORDER BY latency_ms LIMIT 1 OFFSET ?"
		p95Args = append(p95Args, offset)
		var p95 sql.NullInt64
		if err := r.db.QueryRowContext(ctx, p95Query, p95Args...).Scan(&p95); err == nil && p95.Valid {
			summary.P95LatencyMs = float64(p95.Int64)
		}
	}
	return summary, nil
}

func (r *analyticsRepo) ListTopQuestions(ctx context.Context, filter biz.AnalyticsFilter) ([]biz.QuestionStat, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return nil, err
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	query := `SELECT query, COUNT(*) AS cnt, SUM(CASE WHEN hit = 1 THEN 1 ELSE 0 END), MAX(created_at)
		FROM analytics_event
		WHERE tenant_id = ? AND event_type = ? AND query IS NOT NULL AND query <> ''`
	args := []any{tenantID, biz.EventRAGQuery}
	query, args = applyEventFilters(query, args, filter)
	query += " GROUP BY query ORDER BY cnt DESC LIMIT ?"
	args = append(args, limit)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]biz.QuestionStat, 0)
	for rows.Next() {
		var item biz.QuestionStat
		var hitCount sql.NullInt64
		if err := rows.Scan(&item.Query, &item.Count, &hitCount, &item.LastSeenAt); err != nil {
			return nil, err
		}
		if hitCount.Valid && item.Count > 0 {
			item.HitRate = float64(hitCount.Int64) / float64(item.Count)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *analyticsRepo) ListKBGaps(ctx context.Context, filter biz.AnalyticsFilter) ([]biz.GapStat, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return nil, err
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	query := `SELECT query, COUNT(*) AS cnt, AVG(confidence), MAX(created_at)
		FROM analytics_event
		WHERE tenant_id = ? AND event_type = ? AND hit = 0 AND query IS NOT NULL AND query <> ''`
	args := []any{tenantID, biz.EventRAGQuery}
	query, args = applyEventFilters(query, args, filter)
	query += " GROUP BY query ORDER BY cnt DESC LIMIT ?"
	args = append(args, limit)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]biz.GapStat, 0)
	for rows.Next() {
		var item biz.GapStat
		var avgConfidence sql.NullFloat64
		if err := rows.Scan(&item.Query, &item.MissCount, &avgConfidence, &item.LastSeenAt); err != nil {
			return nil, err
		}
		if avgConfidence.Valid {
			item.AvgConfidence = avgConfidence.Float64
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *analyticsRepo) RefreshDaily(ctx context.Context, filter biz.AnalyticsFilter) error {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return err
	}
	query := `SELECT DATE(created_at), latency_ms, hit
		FROM analytics_event
		WHERE tenant_id = ? AND event_type = ?`
	args := []any{tenantID, biz.EventRAGQuery}
	query, args = applyEventFilters(query, args, filter)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	type agg struct {
		date       time.Time
		latencies  []int
		total      int64
		hitCount   int64
		latencySum int64
	}
	buckets := make(map[string]*agg)
	for rows.Next() {
		var day time.Time
		var latency int
		var hit int
		if err := rows.Scan(&day, &latency, &hit); err != nil {
			return err
		}
		key := day.Format("2006-01-02")
		bucket := buckets[key]
		if bucket == nil {
			bucket = &agg{date: day}
			buckets[key] = bucket
		}
		bucket.total++
		bucket.latencySum += int64(latency)
		if hit > 0 {
			bucket.hitCount++
		}
		bucket.latencies = append(bucket.latencies, latency)
	}
	for _, bucket := range buckets {
		avgLatency := 0.0
		if bucket.total > 0 {
			avgLatency = float64(bucket.latencySum) / float64(bucket.total)
		}
		p95 := computePercentile(bucket.latencies, 0.95)
		hitRate := 0.0
		if bucket.total > 0 {
			hitRate = float64(bucket.hitCount) / float64(bucket.total)
		}
		if err := r.upsertDaily(ctx, tenantID, filter.BotID, bucket.date, bucket.total, bucket.hitCount, hitRate, avgLatency, p95); err != nil {
			return err
		}
	}
	return rows.Err()
}

func (r *analyticsRepo) ListDaily(ctx context.Context, filter biz.AnalyticsFilter) ([]biz.DailyStat, error) {
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return nil, err
	}
	query := `SELECT date, total_queries, hit_queries, hit_rate, avg_latency_ms, p95_latency_ms
		FROM analytics_daily
		WHERE tenant_id = ?`
	args := []any{tenantID}
	if strings.TrimSpace(filter.BotID) != "" {
		query += " AND bot_id = ?"
		args = append(args, strings.TrimSpace(filter.BotID))
	}
	if !filter.Start.IsZero() {
		query += " AND date >= ?"
		args = append(args, filter.Start.Format("2006-01-02"))
	}
	if !filter.End.IsZero() {
		query += " AND date <= ?"
		args = append(args, filter.End.Format("2006-01-02"))
	}
	query += " ORDER BY date ASC"
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]biz.DailyStat, 0)
	for rows.Next() {
		var item biz.DailyStat
		if err := rows.Scan(&item.Date, &item.Total, &item.HitCount, &item.HitRate, &item.AvgLatencyMs, &item.P95LatencyMs); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *analyticsRepo) upsertDaily(ctx context.Context, tenantID string, botID string, date time.Time, total int64, hits int64, hitRate float64, avgLatency float64, p95 float64) error {
	if r == nil || r.db == nil {
		return sql.ErrConnDone
	}
	if strings.TrimSpace(botID) == "" {
		botID = "all"
	}
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO analytics_daily
			(id, tenant_id, bot_id, date, total_queries, hit_queries, hit_rate, avg_latency_ms, p95_latency_ms, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE total_queries = VALUES(total_queries),
			hit_queries = VALUES(hit_queries),
			hit_rate = VALUES(hit_rate),
			avg_latency_ms = VALUES(avg_latency_ms),
			p95_latency_ms = VALUES(p95_latency_ms),
			updated_at = NOW()`,
		uuid.NewString(),
		tenantID,
		botID,
		date.Format("2006-01-02"),
		total,
		hits,
		hitRate,
		avgLatency,
		p95,
	)
	return err
}

func applyEventFilters(query string, args []any, filter biz.AnalyticsFilter) (string, []any) {
	if strings.TrimSpace(filter.BotID) != "" {
		query += " AND bot_id = ?"
		args = append(args, strings.TrimSpace(filter.BotID))
	}
	if !filter.Start.IsZero() {
		query += " AND created_at >= ?"
		args = append(args, filter.Start)
	}
	if !filter.End.IsZero() {
		query += " AND created_at <= ?"
		args = append(args, filter.End)
	}
	return query, args
}

func computePercentile(values []int, percentile float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sort.Ints(values)
	if percentile <= 0 {
		return float64(values[0])
	}
	if percentile >= 1 {
		return float64(values[len(values)-1])
	}
	index := int(math.Ceil(float64(len(values))*percentile)) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(values) {
		index = len(values) - 1
	}
	return float64(values[index])
}

func (r *analyticsRepo) recordRealtime(ctx context.Context, event biz.AnalyticsEvent) {
	if r.redis == nil {
		return
	}
	day := event.CreatedAt.UTC().Format("20060102")
	prefix := fmt.Sprintf("analytics:%s:%s:%s", event.TenantID, event.BotID, day)
	pipe := r.redis.Pipeline()
	pipe.Incr(ctx, prefix+":total")
	if event.Hit {
		pipe.Incr(ctx, prefix+":hit")
	}
	pipe.IncrBy(ctx, prefix+":latency_sum", int64(event.LatencyMs))
	pipe.Incr(ctx, prefix+":latency_count")
	pipe.Expire(ctx, prefix+":total", metricsTTL)
	pipe.Expire(ctx, prefix+":hit", metricsTTL)
	pipe.Expire(ctx, prefix+":latency_sum", metricsTTL)
	pipe.Expire(ctx, prefix+":latency_count", metricsTTL)
	_, _ = pipe.Exec(ctx)
}

func nullString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return strings.TrimSpace(value)
}

func nullInt32(value int32) any {
	if value == 0 {
		return nil
	}
	return value
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

// ProviderSet is analytics data providers.
var ProviderSet = wire.NewSet(NewAnalyticsRepo)
