package service

import (
	"context"
	"time"

	v1 "github.com/ZTH7/RAGDesk/apps/server/api/analytics/v1"
	biz "github.com/ZTH7/RAGDesk/apps/server/internal/analytics/biz"
	iambiz "github.com/ZTH7/RAGDesk/apps/server/internal/iam/biz"
	"github.com/ZTH7/RAGDesk/apps/server/internal/tenant"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AnalyticsService handles analytics service layer.
type AnalyticsService struct {
	v1.UnimplementedConsoleAnalyticsServer

	uc  *biz.AnalyticsUsecase
	iam *iambiz.IAMUsecase
	log *log.Helper
}

// NewAnalyticsService creates a new AnalyticsService.
func NewAnalyticsService(uc *biz.AnalyticsUsecase, iam *iambiz.IAMUsecase, logger log.Logger) *AnalyticsService {
	return &AnalyticsService{uc: uc, iam: iam, log: log.NewHelper(logger)}
}

func (s *AnalyticsService) GetOverview(ctx context.Context, req *v1.GetOverviewRequest) (*v1.GetOverviewResponse, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	if err := requireTenantContext(ctx); err != nil {
		return nil, err
	}
	if err := s.iam.RequirePermission(ctx, "tenant.analytics.read"); err != nil {
		return nil, err
	}
	overview, err := s.uc.GetOverview(ctx, biz.AnalyticsFilter{
		BotID: req.GetBotId(),
		Start: fromTimestamp(req.GetStartTime()),
		End:   fromTimestamp(req.GetEndTime()),
	})
	if err != nil {
		return nil, err
	}
	return &v1.GetOverviewResponse{Overview: &v1.Overview{
		TotalQueries: overview.Total,
		HitQueries:   overview.HitCount,
		HitRate:      overview.HitRate,
		AvgLatencyMs: overview.AvgLatencyMs,
		P95LatencyMs: overview.P95LatencyMs,
		ErrorCount:   overview.ErrorCount,
		ErrorRate:    overview.ErrorRate,
	}}, nil
}

func (s *AnalyticsService) GetLatency(ctx context.Context, req *v1.GetLatencyRequest) (*v1.GetLatencyResponse, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	if err := requireTenantContext(ctx); err != nil {
		return nil, err
	}
	if err := s.iam.RequirePermission(ctx, "tenant.analytics.read"); err != nil {
		return nil, err
	}
	points, err := s.uc.ListLatencySeries(ctx, biz.AnalyticsFilter{
		BotID: req.GetBotId(),
		Start: fromTimestamp(req.GetStartTime()),
		End:   fromTimestamp(req.GetEndTime()),
	})
	if err != nil {
		return nil, err
	}
	resp := &v1.GetLatencyResponse{Points: make([]*v1.LatencyPoint, 0, len(points))}
	for _, point := range points {
		resp.Points = append(resp.Points, &v1.LatencyPoint{
			Date:         toTimestamp(point.Date),
			AvgLatencyMs: point.AvgLatencyMs,
			P95LatencyMs: point.P95LatencyMs,
			TotalQueries: point.Total,
			HitQueries:   point.HitCount,
		})
	}
	return resp, nil
}

func (s *AnalyticsService) GetTopQuestions(ctx context.Context, req *v1.GetTopQuestionsRequest) (*v1.GetTopQuestionsResponse, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	if err := requireTenantContext(ctx); err != nil {
		return nil, err
	}
	if err := s.iam.RequirePermission(ctx, "tenant.analytics.read"); err != nil {
		return nil, err
	}
	items, err := s.uc.ListTopQuestions(ctx, biz.AnalyticsFilter{
		BotID: req.GetBotId(),
		Start: fromTimestamp(req.GetStartTime()),
		End:   fromTimestamp(req.GetEndTime()),
		Limit: int(req.GetLimit()),
	})
	if err != nil {
		return nil, err
	}
	resp := &v1.GetTopQuestionsResponse{Items: make([]*v1.QuestionStat, 0, len(items))}
	for _, item := range items {
		resp.Items = append(resp.Items, &v1.QuestionStat{
			Query:      item.Query,
			Count:      item.Count,
			HitRate:    item.HitRate,
			LastSeenAt: toTimestamp(item.LastSeenAt),
		})
	}
	return resp, nil
}

func (s *AnalyticsService) GetKBGaps(ctx context.Context, req *v1.GetKBGapsRequest) (*v1.GetKBGapsResponse, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	if err := requireTenantContext(ctx); err != nil {
		return nil, err
	}
	if err := s.iam.RequirePermission(ctx, "tenant.analytics.read"); err != nil {
		return nil, err
	}
	items, err := s.uc.ListKBGaps(ctx, biz.AnalyticsFilter{
		BotID: req.GetBotId(),
		Start: fromTimestamp(req.GetStartTime()),
		End:   fromTimestamp(req.GetEndTime()),
		Limit: int(req.GetLimit()),
	})
	if err != nil {
		return nil, err
	}
	resp := &v1.GetKBGapsResponse{Items: make([]*v1.GapStat, 0, len(items))}
	for _, item := range items {
		resp.Items = append(resp.Items, &v1.GapStat{
			Query:         item.Query,
			MissCount:     item.MissCount,
			AvgConfidence: item.AvgConfidence,
			LastSeenAt:    toTimestamp(item.LastSeenAt),
		})
	}
	return resp, nil
}

func requireTenantContext(ctx context.Context) error {
	if _, err := tenant.RequireTenantID(ctx); err != nil {
		return errors.Forbidden("TENANT_MISSING", "tenant missing")
	}
	return nil
}

func toTimestamp(value time.Time) *timestamppb.Timestamp {
	if value.IsZero() {
		return nil
	}
	return timestamppb.New(value)
}

func fromTimestamp(value *timestamppb.Timestamp) time.Time {
	if value == nil {
		return time.Time{}
	}
	return value.AsTime()
}

// ProviderSet is analytics service providers.
var ProviderSet = wire.NewSet(NewAnalyticsService)
