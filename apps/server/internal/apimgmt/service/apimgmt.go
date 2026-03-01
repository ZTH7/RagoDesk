package service

import (
	"context"
	"strings"
	"time"

	v1 "github.com/ZTH7/RagoDesk/apps/server/api/apimgmt/v1"
	biz "github.com/ZTH7/RagoDesk/apps/server/internal/apimgmt/biz"
	iambiz "github.com/ZTH7/RagoDesk/apps/server/internal/iam/biz"
	"github.com/ZTH7/RagoDesk/apps/server/internal/kit/tenant"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// APIMgmtService handles apimgmt service layer.
type APIMgmtService struct {
	v1.UnimplementedConsoleAPIMgmtServer

	uc  *biz.APIMgmtUsecase
	iam *iambiz.IAMUsecase
	log *log.Helper
}

// NewAPIMgmtService creates a new APIMgmtService
func NewAPIMgmtService(uc *biz.APIMgmtUsecase, iam *iambiz.IAMUsecase, logger log.Logger) *APIMgmtService {
	return &APIMgmtService{uc: uc, iam: iam, log: log.NewHelper(logger)}
}

func (s *APIMgmtService) CreateAPIKey(ctx context.Context, req *v1.CreateAPIKeyRequest) (*v1.CreateAPIKeyResponse, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	if err := requireTenantContext(ctx); err != nil {
		return nil, err
	}
	if err := s.iam.RequirePermission(ctx, biz.PermissionAPIKeyWrite); err != nil {
		return nil, err
	}
	var publicChatEnabled *bool
	if req.PublicChatEnabled != nil {
		value := req.PublicChatEnabled.Value
		publicChatEnabled = &value
	}
	created, rawKey, err := s.uc.CreateAPIKey(ctx, req.GetName(), req.GetBotId(), req.GetScopes(), req.GetApiVersions(), req.GetQuotaDaily(), req.GetQpsLimit(), publicChatEnabled)
	if err != nil {
		return nil, err
	}
	return &v1.CreateAPIKeyResponse{
		ApiKey: toAPIKey(created),
		RawKey: rawKey,
	}, nil
}

func (s *APIMgmtService) ListAPIKeys(ctx context.Context, req *v1.ListAPIKeysRequest) (*v1.ListAPIKeysResponse, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	if err := requireTenantContext(ctx); err != nil {
		return nil, err
	}
	if err := s.iam.RequirePermission(ctx, biz.PermissionAPIKeyRead); err != nil {
		return nil, err
	}
	items, err := s.uc.ListAPIKeys(ctx, req.GetBotId(), int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, err
	}
	resp := &v1.ListAPIKeysResponse{Items: make([]*v1.APIKey, 0, len(items))}
	for _, item := range items {
		resp.Items = append(resp.Items, toAPIKey(item))
	}
	return resp, nil
}

func (s *APIMgmtService) GetAPIKey(ctx context.Context, req *v1.GetAPIKeyRequest) (*v1.GetAPIKeyResponse, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	if err := requireTenantContext(ctx); err != nil {
		return nil, err
	}
	if err := s.iam.RequirePermission(ctx, biz.PermissionAPIKeyRead); err != nil {
		return nil, err
	}
	item, err := s.uc.GetAPIKey(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &v1.GetAPIKeyResponse{ApiKey: toAPIKey(item)}, nil
}

func (s *APIMgmtService) UpdateAPIKey(ctx context.Context, req *v1.UpdateAPIKeyRequest) (*v1.UpdateAPIKeyResponse, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	if err := requireTenantContext(ctx); err != nil {
		return nil, err
	}
	if err := s.iam.RequirePermission(ctx, biz.PermissionAPIKeyWrite); err != nil {
		return nil, err
	}
	var quotaDaily *int32
	if req.QuotaDaily != nil {
		value := req.QuotaDaily.Value
		quotaDaily = &value
	}
	var qpsLimit *int32
	if req.QpsLimit != nil {
		value := req.QpsLimit.Value
		qpsLimit = &value
	}
	var publicChatEnabled *bool
	if req.PublicChatEnabled != nil {
		value := req.PublicChatEnabled.Value
		publicChatEnabled = &value
	}
	updated, err := s.uc.UpdateAPIKey(ctx, req.GetId(), req.GetName(), req.GetStatus(), req.GetScopes(), req.GetApiVersions(), quotaDaily, qpsLimit, publicChatEnabled)
	if err != nil {
		return nil, err
	}
	return &v1.UpdateAPIKeyResponse{ApiKey: toAPIKey(updated)}, nil
}

func (s *APIMgmtService) DeleteAPIKey(ctx context.Context, req *v1.DeleteAPIKeyRequest) (*emptypb.Empty, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	if err := requireTenantContext(ctx); err != nil {
		return nil, err
	}
	if err := s.iam.RequirePermission(ctx, biz.PermissionAPIKeyDelete); err != nil {
		return nil, err
	}
	if err := s.uc.DeleteAPIKey(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *APIMgmtService) RotateAPIKey(ctx context.Context, req *v1.RotateAPIKeyRequest) (*v1.RotateAPIKeyResponse, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	if err := requireTenantContext(ctx); err != nil {
		return nil, err
	}
	if err := s.iam.RequirePermission(ctx, biz.PermissionAPIKeyRotate); err != nil {
		return nil, err
	}
	updated, rawKey, err := s.uc.RotateAPIKey(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &v1.RotateAPIKeyResponse{
		ApiKey: toAPIKey(updated),
		RawKey: rawKey,
	}, nil
}

func (s *APIMgmtService) RegeneratePublicChatID(ctx context.Context, req *v1.RegeneratePublicChatIDRequest) (*v1.RegeneratePublicChatIDResponse, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	if err := requireTenantContext(ctx); err != nil {
		return nil, err
	}
	if err := s.iam.RequirePermission(ctx, biz.PermissionAPIKeyRotate); err != nil {
		return nil, err
	}
	updated, err := s.uc.RegeneratePublicChatID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &v1.RegeneratePublicChatIDResponse{ApiKey: toAPIKey(updated)}, nil
}

func (s *APIMgmtService) ListUsageLogs(ctx context.Context, req *v1.ListUsageLogsRequest) (*v1.ListUsageLogsResponse, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	if err := requireTenantContext(ctx); err != nil {
		return nil, err
	}
	if err := s.iam.RequirePermission(ctx, biz.PermissionAPIUsageRead); err != nil {
		return nil, err
	}
	filter := biz.UsageFilter{
		APIKeyID:   req.GetApiKeyId(),
		BotID:      req.GetBotId(),
		APIVersion: req.GetApiVersion(),
		Model:      req.GetModel(),
		Start:      fromTimestamp(req.GetStartTime()),
		End:        fromTimestamp(req.GetEndTime()),
		Limit:      int(req.GetLimit()),
		Offset:     int(req.GetOffset()),
	}
	items, err := s.uc.ListUsageLogs(ctx, filter)
	if err != nil {
		return nil, err
	}
	resp := &v1.ListUsageLogsResponse{Items: make([]*v1.UsageLog, 0, len(items))}
	for _, item := range items {
		resp.Items = append(resp.Items, toUsageLog(item))
	}
	return resp, nil
}

func (s *APIMgmtService) GetUsageSummary(ctx context.Context, req *v1.GetUsageSummaryRequest) (*v1.GetUsageSummaryResponse, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	if err := requireTenantContext(ctx); err != nil {
		return nil, err
	}
	if err := s.iam.RequirePermission(ctx, biz.PermissionAPIUsageRead); err != nil {
		return nil, err
	}
	summary, err := s.uc.GetUsageSummary(ctx, biz.UsageFilter{
		APIKeyID:   req.GetApiKeyId(),
		BotID:      req.GetBotId(),
		APIVersion: req.GetApiVersion(),
		Model:      req.GetModel(),
		Start:      fromTimestamp(req.GetStartTime()),
		End:        fromTimestamp(req.GetEndTime()),
	})
	if err != nil {
		return nil, err
	}
	return &v1.GetUsageSummaryResponse{Summary: &v1.UsageSummary{
		Total:            summary.Total,
		ErrorCount:       summary.ErrorCount,
		AvgLatencyMs:     summary.AvgLatencyMs,
		PromptTokens:     summary.PromptTokens,
		CompletionTokens: summary.CompletionTokens,
		TotalTokens:      summary.TotalTokens,
	}}, nil
}

func (s *APIMgmtService) ExportUsageLogs(ctx context.Context, req *v1.ExportUsageLogsRequest) (*v1.ExportUsageLogsResponse, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	tenantID, err := tenant.RequireTenantID(ctx)
	if err != nil {
		return nil, errors.Forbidden("TENANT_MISSING", "tenant missing")
	}
	if err := s.iam.RequirePermission(ctx, biz.PermissionAPIUsageRead); err != nil {
		return nil, err
	}
	format := strings.ToLower(strings.TrimSpace(req.GetFormat()))
	if format == "" {
		format = "csv"
	}
	if format != "csv" {
		return nil, errors.BadRequest("EXPORT_FORMAT_INVALID", "unsupported export format")
	}
	result, err := s.uc.ExportUsageLogs(ctx, tenantID, biz.UsageFilter{
		APIKeyID:   req.GetApiKeyId(),
		BotID:      req.GetBotId(),
		APIVersion: req.GetApiVersion(),
		Model:      req.GetModel(),
		Start:      fromTimestamp(req.GetStartTime()),
		End:        fromTimestamp(req.GetEndTime()),
		Limit:      int(req.GetLimit()),
		Offset:     int(req.GetOffset()),
	})
	if err != nil {
		return nil, err
	}
	return &v1.ExportUsageLogsResponse{
		Content:     result.Content,
		ContentType: result.ContentType,
		Filename:    result.Filename,
		DownloadUrl: result.DownloadURL,
		ObjectUri:   result.ObjectURI,
	}, nil
}

// ProviderSet is apimgmt service providers.
var ProviderSet = wire.NewSet(NewAPIMgmtService)

func requireTenantContext(ctx context.Context) error {
	if _, err := tenant.RequireTenantID(ctx); err != nil {
		return errors.Forbidden("TENANT_MISSING", "tenant missing")
	}
	return nil
}

func toAPIKey(key biz.APIKey) *v1.APIKey {
	if key.ID == "" {
		return nil
	}
	return &v1.APIKey{
		Id:                key.ID,
		TenantId:          key.TenantID,
		BotId:             key.BotID,
		Name:              key.Name,
		Status:            string(key.Status),
		Scopes:            key.Scopes,
		ApiVersions:       key.APIVersions,
		QuotaDaily:        key.QuotaDaily,
		QpsLimit:          key.QPSLimit,
		CreatedAt:         toTimestamp(key.CreatedAt),
		LastUsedAt:        toTimestamp(key.LastUsedAt),
		PublicChatId:      key.PublicChatID,
		PublicChatEnabled: key.PublicChatEnabled,
	}
}

func toUsageLog(log biz.UsageLog) *v1.UsageLog {
	if log.ID == "" {
		return nil
	}
	return &v1.UsageLog{
		Id:               log.ID,
		ApiKeyId:         log.APIKeyID,
		BotId:            log.BotID,
		Path:             log.Path,
		ApiVersion:       log.APIVersion,
		Model:            log.Model,
		StatusCode:       log.StatusCode,
		LatencyMs:        log.LatencyMs,
		PromptTokens:     log.PromptTokens,
		CompletionTokens: log.CompletionTokens,
		TotalTokens:      log.TotalTokens,
		CreatedAt:        toTimestamp(log.CreatedAt),
		ClientIp:         log.ClientIP,
		UserAgent:        log.UserAgent,
	}
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
