package biz

import (
	"context"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"github.com/google/wire"
)

const (
	EventRAGQuery       = "rag_query"
	EventFeedback       = "feedback"
	EventRetrieval      = "retrieval"
	EventSessionOpen    = "session_open"
	EventSessionClose   = "session_close"
	EventMessageCreated = "message_created"
)

const (
	PermissionAnalyticsRead = "tenant.analytics.read"
)

const (
	defaultAnalyticsWindow = 7 * 24 * time.Hour
	maxQueryLength         = 1024
	maxListLimit           = 100
)

// AnalyticsEvent describes a captured analytics event.
type AnalyticsEvent struct {
	ID         string
	TenantID   string
	BotID      string
	EventType  string
	SessionID  string
	MessageID  string
	Query      string
	Hit        bool
	Confidence float64
	LatencyMs  int32
	StatusCode int32
	Rating     int32
	CreatedAt  time.Time
}

// AnalyticsFilter defines analytics query filters.
type AnalyticsFilter struct {
	BotID string
	Start time.Time
	End   time.Time
	Limit int
}

// OverviewStats aggregates usage metrics.
type OverviewStats struct {
	Total        int64
	HitCount     int64
	HitRate      float64
	AvgLatencyMs float64
	P95LatencyMs float64
	ErrorCount   int64
	ErrorRate    float64
}

// LatencyPoint describes daily latency stats.
type LatencyPoint struct {
	Date         time.Time
	Total        int64
	HitCount     int64
	AvgLatencyMs float64
	P95LatencyMs float64
}

// QuestionStat represents aggregated question stats.
type QuestionStat struct {
	Query      string
	Count      int64
	HitRate    float64
	LastSeenAt time.Time
}

// GapStat represents knowledge gap stats.
type GapStat struct {
	Query         string
	MissCount     int64
	AvgConfidence float64
	LastSeenAt    time.Time
}

// DailyStat represents daily aggregates stored in DB.
type DailyStat struct {
	Date         time.Time
	Total        int64
	HitCount     int64
	HitRate      float64
	AvgLatencyMs float64
	P95LatencyMs float64
}

// AnalyticsRepo is a repository interface.
type AnalyticsRepo interface {
	CreateEvent(ctx context.Context, event AnalyticsEvent) error
	GetOverview(ctx context.Context, filter AnalyticsFilter) (OverviewStats, error)
	ListTopQuestions(ctx context.Context, filter AnalyticsFilter) ([]QuestionStat, error)
	ListKBGaps(ctx context.Context, filter AnalyticsFilter) ([]GapStat, error)
	RefreshDaily(ctx context.Context, filter AnalyticsFilter) error
	ListDaily(ctx context.Context, filter AnalyticsFilter) ([]DailyStat, error)
}

// AnalyticsUsecase handles analytics business logic.
type AnalyticsUsecase struct {
	repo AnalyticsRepo
	log  *log.Helper
}

// NewAnalyticsUsecase creates a new AnalyticsUsecase.
func NewAnalyticsUsecase(repo AnalyticsRepo, logger log.Logger) *AnalyticsUsecase {
	return &AnalyticsUsecase{repo: repo, log: log.NewHelper(logger)}
}

func (uc *AnalyticsUsecase) RecordRAGEvent(ctx context.Context, event AnalyticsEvent) {
	uc.recordEvent(ctx, event, EventRAGQuery)
}

func (uc *AnalyticsUsecase) RecordFeedback(ctx context.Context, event AnalyticsEvent) {
	uc.recordEvent(ctx, event, EventFeedback)
}

func (uc *AnalyticsUsecase) RecordRetrievalEvent(ctx context.Context, event AnalyticsEvent) {
	uc.recordEvent(ctx, event, EventRetrieval)
}

func (uc *AnalyticsUsecase) RecordSessionEvent(ctx context.Context, event AnalyticsEvent, eventType string) {
	if eventType != EventSessionOpen && eventType != EventSessionClose {
		eventType = EventSessionOpen
	}
	uc.recordEvent(ctx, event, eventType)
}

func (uc *AnalyticsUsecase) RecordMessageEvent(ctx context.Context, event AnalyticsEvent) {
	uc.recordEvent(ctx, event, EventMessageCreated)
}

func (uc *AnalyticsUsecase) recordEvent(ctx context.Context, event AnalyticsEvent, eventType string) {
	if uc == nil || uc.repo == nil {
		return
	}
	event.EventType = eventType
	if event.Query != "" {
		event.Query = normalizeQuery(event.Query)
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}
	if event.ID == "" {
		event.ID = uuid.NewString()
	}
	if err := uc.repo.CreateEvent(ctx, event); err != nil && uc.log != nil {
		uc.log.Warnf("record analytics event failed: %v", err)
	}
}

func (uc *AnalyticsUsecase) GetOverview(ctx context.Context, filter AnalyticsFilter) (OverviewStats, error) {
	filter = normalizeFilter(filter, defaultAnalyticsWindow)
	return uc.repo.GetOverview(ctx, filter)
}

func (uc *AnalyticsUsecase) ListLatencySeries(ctx context.Context, filter AnalyticsFilter) ([]DailyStat, error) {
	filter = normalizeFilter(filter, defaultAnalyticsWindow)
	if err := uc.repo.RefreshDaily(ctx, filter); err != nil && uc.log != nil {
		uc.log.Warnf("refresh analytics daily failed: %v", err)
	}
	return uc.repo.ListDaily(ctx, filter)
}

func (uc *AnalyticsUsecase) ListTopQuestions(ctx context.Context, filter AnalyticsFilter) ([]QuestionStat, error) {
	filter = normalizeFilter(filter, defaultAnalyticsWindow)
	return uc.repo.ListTopQuestions(ctx, filter)
}

func (uc *AnalyticsUsecase) ListKBGaps(ctx context.Context, filter AnalyticsFilter) ([]GapStat, error) {
	filter = normalizeFilter(filter, defaultAnalyticsWindow)
	return uc.repo.ListKBGaps(ctx, filter)
}

func normalizeFilter(filter AnalyticsFilter, window time.Duration) AnalyticsFilter {
	if filter.Limit <= 0 || filter.Limit > maxListLimit {
		filter.Limit = maxListLimit
	}
	if window > 0 && filter.Start.IsZero() && filter.End.IsZero() {
		end := time.Now()
		filter.End = end
		filter.Start = end.Add(-window)
	}
	if !filter.Start.IsZero() && !filter.End.IsZero() && filter.End.Before(filter.Start) {
		filter.Start, filter.End = filter.End, filter.Start
	}
	return filter
}

func normalizeQuery(query string) string {
	query = strings.TrimSpace(query)
	if query == "" {
		return ""
	}
	if len(query) > maxQueryLength {
		return query[:maxQueryLength]
	}
	return query
}

// ProviderSet is analytics biz providers.
var ProviderSet = wire.NewSet(NewAnalyticsUsecase)
