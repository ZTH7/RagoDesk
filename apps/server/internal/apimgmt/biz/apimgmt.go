package biz

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/ZTH7/RagoDesk/apps/server/internal/ai/provider"
	"github.com/ZTH7/RagoDesk/apps/server/internal/conf"
	"github.com/ZTH7/RagoDesk/apps/server/internal/kit/paging"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"github.com/google/wire"
)

// Permission codes (tenant scope).
const (
	PermissionAPIKeyRead   = "tenant.api_key.read"
	PermissionAPIKeyWrite  = "tenant.api_key.write"
	PermissionAPIKeyDelete = "tenant.api_key.delete"
	PermissionAPIKeyRotate = "tenant.api_key.rotate"
	PermissionAPIUsageRead = "tenant.api_usage.read"
)

// APIKeyStatus represents api key lifecycle state.
type APIKeyStatus string

const (
	APIKeyStatusActive  APIKeyStatus = "active"
	APIKeyStatusRevoked APIKeyStatus = "revoked"
)

const (
	ScopeAll          = "*"
	ScopeRAG          = "rag"
	ScopeConversation = "conversation"
)

var DefaultAPIKeyScopes = []string{ScopeRAG, ScopeConversation}

var DefaultAPIVersions = []string{"v1"}

const (
	defaultRotationGraceMinutes = 60
	defaultUsageListWindow      = 7 * 24 * time.Hour
	defaultUsageSummaryWindow   = 30 * 24 * time.Hour
)

// APIKey represents an API key with scope and rotation metadata.
type APIKey struct {
	ID                string
	TenantID          string
	BotID             string
	Name              string
	KeyHash           string
	PublicChatID      string
	PublicChatEnabled bool
	APIVersions       []string
	Scopes            []string
	Status            APIKeyStatus
	RotatedFrom       string
	QuotaDaily        int32
	QPSLimit          int32
	CreatedAt         time.Time
	LastUsedAt        time.Time
}

// UsageLog records API usage for analytics.
type UsageLog struct {
	ID               string
	APIKeyID         string
	TenantID         string
	BotID            string
	Path             string
	APIVersion       string
	Model            string
	StatusCode       int32
	LatencyMs        int32
	PromptTokens     int32
	CompletionTokens int32
	TotalTokens      int32
	ClientIP         string
	UserAgent        string
	CreatedAt        time.Time
}

// UsageSummary aggregates API usage metrics.
type UsageSummary struct {
	Total            int64
	ErrorCount       int64
	AvgLatencyMs     float64
	PromptTokens     int64
	CompletionTokens int64
	TotalTokens      int64
}

// UsageExportResult represents exported usage payload.
type UsageExportResult struct {
	Content     string
	ContentType string
	Filename    string
	ObjectURI   string
	DownloadURL string
}

// UsageFilter describes usage query filters.
type UsageFilter struct {
	APIKeyID   string
	BotID      string
	APIVersion string
	Model      string
	Start      time.Time
	End        time.Time
	Limit      int
	Offset     int
}

// UsageEvent is an analytics hook payload.
type UsageEvent struct {
	TenantID         string
	BotID            string
	APIKeyID         string
	Path             string
	APIVersion       string
	Model            string
	StatusCode       int32
	LatencyMs        int32
	PromptTokens     int32
	CompletionTokens int32
	TotalTokens      int32
	ClientIP         string
	UserAgent        string
	CreatedAt        time.Time
}

// APIMgmtRepo is a repository interface (placeholder)
type APIMgmtRepo interface {
	CreateAPIKey(ctx context.Context, key APIKey) (APIKey, error)
	GetAPIKey(ctx context.Context, keyID string) (APIKey, error)
	GetAPIKeyByPublicChatID(ctx context.Context, chatID string) (APIKey, error)
	ListAPIKeys(ctx context.Context, botID string, limit int, offset int) ([]APIKey, error)
	UpdateAPIKey(ctx context.Context, key APIKey) (APIKey, error)
	RegeneratePublicChatID(ctx context.Context, keyID string, chatID string) (APIKey, error)
	DeleteAPIKey(ctx context.Context, keyID string) error
	RotateKey(ctx context.Context, keyID string, newHash string, graceUntil time.Time) (APIKey, error)
	UpdateLastUsedAt(ctx context.Context, keyID string, lastUsedAt time.Time) error
	CreateUsageLog(ctx context.Context, log UsageLog) error
	GetAPIKeyByHash(ctx context.Context, keyHash string) (APIKey, error)
	ListUsageLogs(ctx context.Context, filter UsageFilter) ([]UsageLog, error)
	GetUsageSummary(ctx context.Context, filter UsageFilter) (UsageSummary, error)
}

// RateLimiter enforces QPS/quota limits.
type RateLimiter interface {
	Check(ctx context.Context, key APIKey) error
}

// UsageExporter writes usage exports to object storage.
type UsageExporter interface {
	ExportUsageCSV(ctx context.Context, tenantID string, filename string, reader io.Reader, contentType string) (objectURI string, downloadURL string, err error)
}

// UsageSink forwards usage events to analytics (phase 6 placeholder).
type UsageSink interface {
	CaptureAPIUsage(ctx context.Context, event UsageEvent) error
}

type noopUsageSink struct{}

func (n noopUsageSink) CaptureAPIUsage(ctx context.Context, event UsageEvent) error {
	return nil
}

// NewUsageSink returns a no-op usage sink.
func NewUsageSink() UsageSink {
	return noopUsageSink{}
}

// APIMgmtUsecase handles apimgmt business logic (placeholder)
type APIMgmtUsecase struct {
	repo          APIMgmtRepo
	exporter      UsageExporter
	limiter       RateLimiter
	sink          UsageSink
	log           *log.Helper
	rotationGrace time.Duration
}

// NewAPIMgmtUsecase creates a new APIMgmtUsecase
func NewAPIMgmtUsecase(repo APIMgmtRepo, exporter UsageExporter, limiter RateLimiter, sink UsageSink, cfg *conf.Data, logger log.Logger) *APIMgmtUsecase {
	rotationGrace := time.Duration(defaultRotationGraceMinutes) * time.Minute
	if cfg != nil && cfg.Apimgmt != nil && cfg.Apimgmt.RotationGraceMinutes > 0 {
		rotationGrace = time.Duration(cfg.Apimgmt.RotationGraceMinutes) * time.Minute
	}
	return &APIMgmtUsecase{
		repo:          repo,
		exporter:      exporter,
		limiter:       limiter,
		sink:          sink,
		log:           log.NewHelper(logger),
		rotationGrace: rotationGrace,
	}
}

const DefaultAPIKeyHeader = "X-API-Key"
const DefaultPublicChatHeader = "X-Chat-Key"

func (uc *APIMgmtUsecase) CreateAPIKey(ctx context.Context, name string, botID string, scopes []string, apiVersions []string, quotaDaily int32, qpsLimit int32, publicChatEnabled *bool) (APIKey, string, error) {
	name = strings.TrimSpace(name)
	botID = strings.TrimSpace(botID)
	if name == "" {
		return APIKey{}, "", errors.BadRequest("API_KEY_NAME_MISSING", "api key name missing")
	}
	if botID == "" {
		return APIKey{}, "", errors.BadRequest("BOT_ID_MISSING", "bot id missing")
	}
	scopes = normalizeScopes(scopes)
	if len(scopes) == 0 {
		scopes = append([]string(nil), DefaultAPIKeyScopes...)
	}
	apiVersions = normalizeVersions(apiVersions)
	if len(apiVersions) == 0 {
		apiVersions = append([]string(nil), DefaultAPIVersions...)
	}
	if quotaDaily < 0 {
		quotaDaily = 0
	}
	if qpsLimit < 0 {
		qpsLimit = 0
	}
	chatEnabled := true
	if publicChatEnabled != nil {
		chatEnabled = *publicChatEnabled
	}
	for i := 0; i < 3; i++ {
		rawKey, keyHash := generateAPIKey()
		now := time.Now()
		key := APIKey{
			ID:                uuid.NewString(),
			BotID:             botID,
			Name:              name,
			KeyHash:           keyHash,
			PublicChatID:      generatePublicChatID(),
			PublicChatEnabled: chatEnabled,
			APIVersions:       apiVersions,
			Scopes:            scopes,
			Status:            APIKeyStatusActive,
			QuotaDaily:        quotaDaily,
			QPSLimit:          qpsLimit,
			CreatedAt:         now,
		}
		created, err := uc.repo.CreateAPIKey(ctx, key)
		if err == nil {
			return created, rawKey, nil
		}
		if !isDuplicateKeyError(err) {
			return APIKey{}, "", err
		}
	}
	return APIKey{}, "", errors.InternalServer("API_KEY_GENERATE_FAILED", "generate api key failed")
}

func (uc *APIMgmtUsecase) ListAPIKeys(ctx context.Context, botID string, limit int, offset int) ([]APIKey, error) {
	limit, offset = paging.Normalize(limit, offset)
	return uc.repo.ListAPIKeys(ctx, strings.TrimSpace(botID), limit, offset)
}

func (uc *APIMgmtUsecase) GetAPIKey(ctx context.Context, keyID string) (APIKey, error) {
	keyID = strings.TrimSpace(keyID)
	if keyID == "" {
		return APIKey{}, errors.BadRequest("API_KEY_ID_MISSING", "api key id missing")
	}
	return uc.repo.GetAPIKey(ctx, keyID)
}

func (uc *APIMgmtUsecase) UpdateAPIKey(ctx context.Context, keyID string, name string, status string, scopes []string, apiVersions []string, quotaDaily *int32, qpsLimit *int32, publicChatEnabled *bool) (APIKey, error) {
	keyID = strings.TrimSpace(keyID)
	if keyID == "" {
		return APIKey{}, errors.BadRequest("API_KEY_ID_MISSING", "api key id missing")
	}
	if strings.TrimSpace(name) == "" && strings.TrimSpace(status) == "" && len(scopes) == 0 && len(apiVersions) == 0 && quotaDaily == nil && qpsLimit == nil && publicChatEnabled == nil {
		return APIKey{}, errors.BadRequest("API_KEY_UPDATE_EMPTY", "api key update empty")
	}
	current, err := uc.repo.GetAPIKey(ctx, keyID)
	if err != nil {
		return APIKey{}, err
	}
	if strings.TrimSpace(name) != "" {
		current.Name = strings.TrimSpace(name)
	}
	if strings.TrimSpace(status) != "" {
		current.Status = normalizeStatus(status)
	}
	if len(scopes) > 0 {
		current.Scopes = normalizeScopes(scopes)
	}
	if len(apiVersions) > 0 {
		current.APIVersions = normalizeVersions(apiVersions)
	}
	if quotaDaily != nil {
		if *quotaDaily < 0 {
			current.QuotaDaily = 0
		} else {
			current.QuotaDaily = *quotaDaily
		}
	}
	if qpsLimit != nil {
		if *qpsLimit < 0 {
			current.QPSLimit = 0
		} else {
			current.QPSLimit = *qpsLimit
		}
	}
	if publicChatEnabled != nil {
		current.PublicChatEnabled = *publicChatEnabled
	}
	return uc.repo.UpdateAPIKey(ctx, current)
}

func (uc *APIMgmtUsecase) DeleteAPIKey(ctx context.Context, keyID string) error {
	keyID = strings.TrimSpace(keyID)
	if keyID == "" {
		return errors.BadRequest("API_KEY_ID_MISSING", "api key id missing")
	}
	return uc.repo.DeleteAPIKey(ctx, keyID)
}

func (uc *APIMgmtUsecase) RotateAPIKey(ctx context.Context, keyID string) (APIKey, string, error) {
	keyID = strings.TrimSpace(keyID)
	if keyID == "" {
		return APIKey{}, "", errors.BadRequest("API_KEY_ID_MISSING", "api key id missing")
	}
	rawKey, keyHash := generateAPIKey()
	graceUntil := time.Now().Add(uc.rotationGrace)
	updated, err := uc.repo.RotateKey(ctx, keyID, keyHash, graceUntil)
	if err != nil {
		return APIKey{}, "", err
	}
	return updated, rawKey, nil
}

func (uc *APIMgmtUsecase) RegeneratePublicChatID(ctx context.Context, keyID string) (APIKey, error) {
	keyID = strings.TrimSpace(keyID)
	if keyID == "" {
		return APIKey{}, errors.BadRequest("API_KEY_ID_MISSING", "api key id missing")
	}
	for i := 0; i < 3; i++ {
		updated, err := uc.repo.RegeneratePublicChatID(ctx, keyID, generatePublicChatID())
		if err == nil {
			return updated, nil
		}
		if !isDuplicateKeyError(err) {
			return APIKey{}, err
		}
	}
	return APIKey{}, errors.InternalServer("PUBLIC_CHAT_ID_GENERATE_FAILED", "generate public chat id failed")
}

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

func (uc *APIMgmtUsecase) AuthorizeAPIKey(ctx context.Context, rawKey string) (APIKey, error) {
	key, err := uc.ResolveAPIKey(ctx, rawKey)
	if err != nil {
		return APIKey{}, err
	}
	if uc.limiter != nil {
		if err := uc.limiter.Check(ctx, key); err != nil {
			return key, err
		}
	}
	_ = uc.repo.UpdateLastUsedAt(ctx, key.ID, time.Now())
	return key, nil
}

func (uc *APIMgmtUsecase) AuthorizeAPIKeyWithScope(ctx context.Context, rawKey string, requiredScope string, requiredVersion string) (APIKey, error) {
	key, err := uc.AuthorizeAPIKey(ctx, rawKey)
	if err != nil {
		return key, err
	}
	if requiredScope != "" && !scopeAllowed(key.Scopes, requiredScope) {
		return key, errors.Forbidden("API_SCOPE_FORBIDDEN", "api key scope forbidden")
	}
	if requiredVersion != "" && !versionAllowed(key.APIVersions, requiredVersion) {
		return key, errors.Forbidden("API_VERSION_FORBIDDEN", "api version forbidden")
	}
	return key, nil
}

func (uc *APIMgmtUsecase) AuthorizePublicChatIDWithScope(ctx context.Context, chatID string, requiredScope string, requiredVersion string) (APIKey, error) {
	chatID = strings.TrimSpace(chatID)
	if chatID == "" {
		return APIKey{}, errors.Unauthorized("CHAT_KEY_MISSING", "chat key missing")
	}
	key, err := uc.repo.GetAPIKeyByPublicChatID(ctx, chatID)
	if err != nil {
		return APIKey{}, errors.Unauthorized("CHAT_KEY_INVALID", "chat key invalid")
	}
	if key.Status != APIKeyStatusActive || key.TenantID == "" || key.BotID == "" {
		return APIKey{}, errors.Unauthorized("CHAT_KEY_INVALID", "chat key invalid")
	}
	if !key.PublicChatEnabled {
		return APIKey{}, errors.Forbidden("CHAT_KEY_DISABLED", "public chat disabled")
	}
	if uc.limiter != nil {
		if err := uc.limiter.Check(ctx, key); err != nil {
			return key, err
		}
	}
	_ = uc.repo.UpdateLastUsedAt(ctx, key.ID, time.Now())
	if requiredScope != "" && !scopeAllowed(key.Scopes, requiredScope) {
		return key, errors.Forbidden("API_SCOPE_FORBIDDEN", "api key scope forbidden")
	}
	if requiredVersion != "" && !versionAllowed(key.APIVersions, requiredVersion) {
		return key, errors.Forbidden("API_VERSION_FORBIDDEN", "api version forbidden")
	}
	return key, nil
}

func (uc *APIMgmtUsecase) RecordUsage(ctx context.Context, key APIKey, operation string, apiVersion string, model string, usage provider.LLMUsage, statusCode int32, latency time.Duration, clientIP string, userAgent string) {
	if uc == nil || uc.repo == nil || key.ID == "" {
		return
	}
	if statusCode <= 0 {
		statusCode = 200
	}
	log := UsageLog{
		APIKeyID:         key.ID,
		TenantID:         key.TenantID,
		BotID:            key.BotID,
		Path:             strings.TrimSpace(operation),
		APIVersion:       strings.TrimSpace(apiVersion),
		Model:            strings.TrimSpace(model),
		StatusCode:       statusCode,
		LatencyMs:        int32(latency.Milliseconds()),
		PromptTokens:     int32(usage.PromptTokens),
		CompletionTokens: int32(usage.CompletionTokens),
		TotalTokens:      int32(usage.TotalTokens),
		ClientIP:         strings.TrimSpace(clientIP),
		UserAgent:        strings.TrimSpace(userAgent),
		CreatedAt:        time.Now(),
	}
	if err := uc.repo.CreateUsageLog(ctx, log); err != nil && uc.log != nil {
		uc.log.Warnf("record usage failed: %v", err)
	}
	if uc.sink != nil {
		_ = uc.sink.CaptureAPIUsage(ctx, UsageEvent{
			TenantID:         key.TenantID,
			BotID:            key.BotID,
			APIKeyID:         key.ID,
			Path:             log.Path,
			APIVersion:       log.APIVersion,
			Model:            log.Model,
			StatusCode:       log.StatusCode,
			LatencyMs:        log.LatencyMs,
			PromptTokens:     log.PromptTokens,
			CompletionTokens: log.CompletionTokens,
			TotalTokens:      log.TotalTokens,
			ClientIP:         log.ClientIP,
			UserAgent:        log.UserAgent,
			CreatedAt:        log.CreatedAt,
		})
	}
}

func (uc *APIMgmtUsecase) ListUsageLogs(ctx context.Context, filter UsageFilter) ([]UsageLog, error) {
	filter = normalizeUsageFilter(filter, 200, defaultUsageListWindow)
	return uc.repo.ListUsageLogs(ctx, filter)
}

func (uc *APIMgmtUsecase) GetUsageSummary(ctx context.Context, filter UsageFilter) (UsageSummary, error) {
	filter = normalizeUsageFilter(filter, 200, defaultUsageSummaryWindow)
	return uc.repo.GetUsageSummary(ctx, filter)
}

func (uc *APIMgmtUsecase) ExportUsageLogs(ctx context.Context, tenantID string, filter UsageFilter) (UsageExportResult, error) {
	filter = normalizeUsageFilter(filter, 10000, defaultUsageSummaryWindow)
	items, err := uc.repo.ListUsageLogs(ctx, filter)
	if err != nil {
		return UsageExportResult{}, err
	}
	filename := fmt.Sprintf("api_usage_%s.csv", time.Now().UTC().Format("20060102_150405"))
	result := UsageExportResult{
		ContentType: "text/csv",
		Filename:    filename,
	}
	if uc.exporter == nil {
		var buf bytes.Buffer
		if err := writeUsageCSV(&buf, items); err != nil {
			return UsageExportResult{}, err
		}
		result.Content = buf.String()
		return result, nil
	}
	reader, writer := io.Pipe()
	writerErr := make(chan error, 1)
	go func() {
		err := writeUsageCSV(writer, items)
		_ = writer.CloseWithError(err)
		writerErr <- err
	}()
	objectURI, downloadURL, err := uc.exporter.ExportUsageCSV(ctx, tenantID, filename, reader, result.ContentType)
	if err != nil {
		return UsageExportResult{}, err
	}
	if err := <-writerErr; err != nil {
		return UsageExportResult{}, err
	}
	result.ObjectURI = objectURI
	result.DownloadURL = downloadURL
	return result, nil
}

func StatusCodeFromError(err error) int32 {
	if err == nil {
		return 200
	}
	if kratosErr := errors.FromError(err); kratosErr != nil {
		if kratosErr.Code > 0 {
			return kratosErr.Code
		}
	}
	return 500
}

func hashAPIKey(rawKey string) string {
	sum := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(sum[:])
}

func generateAPIKey() (string, string) {
	payload := make([]byte, 32)
	if _, err := rand.Read(payload); err != nil {
		return "", hashAPIKey(uuid.NewString())
	}
	rawKey := base64.RawURLEncoding.EncodeToString(payload)
	return rawKey, hashAPIKey(rawKey)
}

func generatePublicChatID() string {
	payload := make([]byte, 24)
	if _, err := rand.Read(payload); err != nil {
		return "chat_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	}
	return "chat_" + base64.RawURLEncoding.EncodeToString(payload)
}

func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	kratosErr := errors.FromError(err)
	return kratosErr != nil && kratosErr.Reason == "DUPLICATE_KEY"
}

func normalizeStatus(status string) APIKeyStatus {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case string(APIKeyStatusActive):
		return APIKeyStatusActive
	case "disabled":
		return APIKeyStatusRevoked
	default:
		return APIKeyStatusRevoked
	}
}

func normalizeScopes(scopes []string) []string {
	if len(scopes) == 0 {
		return nil
	}
	seen := make(map[string]struct{})
	out := make([]string, 0, len(scopes))
	for _, item := range scopes {
		value := strings.ToLower(strings.TrimSpace(item))
		if value == "" {
			continue
		}
		if value == ScopeAll {
			return []string{ScopeAll}
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func scopeAllowed(scopes []string, required string) bool {
	if required == "" {
		return true
	}
	if len(scopes) == 0 {
		return true
	}
	required = strings.ToLower(strings.TrimSpace(required))
	for _, item := range scopes {
		value := strings.ToLower(strings.TrimSpace(item))
		if value == ScopeAll || value == required {
			return true
		}
	}
	return false
}

func normalizeVersions(versions []string) []string {
	if len(versions) == 0 {
		return nil
	}
	seen := make(map[string]struct{})
	out := make([]string, 0, len(versions))
	for _, item := range versions {
		value := strings.ToLower(strings.TrimSpace(item))
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func versionAllowed(versions []string, required string) bool {
	if required == "" {
		return true
	}
	if len(versions) == 0 {
		return true
	}
	required = strings.ToLower(strings.TrimSpace(required))
	for _, item := range versions {
		if strings.ToLower(strings.TrimSpace(item)) == required {
			return true
		}
	}
	return false
}

func normalizeUsageFilter(filter UsageFilter, maxLimit int, defaultWindow time.Duration) UsageFilter {
	if filter.Limit <= 0 || filter.Limit > maxLimit {
		filter.Limit = maxLimit
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	if defaultWindow > 0 && filter.Start.IsZero() && filter.End.IsZero() {
		end := time.Now()
		filter.End = end
		filter.Start = end.Add(-defaultWindow)
	}
	if !filter.Start.IsZero() && !filter.End.IsZero() && filter.End.Before(filter.Start) {
		filter.Start, filter.End = filter.End, filter.Start
	}
	return filter
}

func writeUsageCSV(writer io.Writer, items []UsageLog) error {
	csvWriter := csv.NewWriter(writer)
	if err := csvWriter.Write([]string{
		"id",
		"api_key_id",
		"tenant_id",
		"bot_id",
		"api_version",
		"model",
		"path",
		"status_code",
		"latency_ms",
		"prompt_tokens",
		"completion_tokens",
		"total_tokens",
		"client_ip",
		"user_agent",
		"created_at",
	}); err != nil {
		csvWriter.Flush()
		return err
	}
	for _, item := range items {
		record := []string{
			item.ID,
			item.APIKeyID,
			item.TenantID,
			item.BotID,
			item.APIVersion,
			item.Model,
			item.Path,
			strconv.Itoa(int(item.StatusCode)),
			strconv.Itoa(int(item.LatencyMs)),
			strconv.Itoa(int(item.PromptTokens)),
			strconv.Itoa(int(item.CompletionTokens)),
			strconv.Itoa(int(item.TotalTokens)),
			item.ClientIP,
			item.UserAgent,
			item.CreatedAt.Format(time.RFC3339),
		}
		if err := csvWriter.Write(record); err != nil {
			csvWriter.Flush()
			return err
		}
	}
	csvWriter.Flush()
	return csvWriter.Error()
}

// ProviderSet is apimgmt biz providers.
var ProviderSet = wire.NewSet(NewAPIMgmtUsecase, NewUsageSink)
