package biz

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"github.com/ZTH7/RAGDesk/apps/server/internal/ai/provider"
	"github.com/ZTH7/RAGDesk/apps/server/internal/conf"
	"github.com/ZTH7/RAGDesk/apps/server/internal/paging"
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

const (
	ScopeAll          = "*"
	ScopeRAG          = "rag"
	ScopeConversation = "conversation"
)

var DefaultAPIKeyScopes = []string{ScopeRAG, ScopeConversation}

var DefaultAPIVersions = []string{"v1"}

const (
	defaultRotationGraceMinutes = 60
)

// APIKey represents an API key with scope and rotation metadata.
type APIKey struct {
	ID          string
	TenantID    string
	BotID       string
	Name        string
	KeyHash     string
	APIVersions []string
	Scopes      []string
	Status      APIKeyStatus
	RotatedFrom string
	QuotaDaily  int32
	QPSLimit    int32
	CreatedAt   time.Time
	LastUsedAt  time.Time
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
	Ping(context.Context) error
	CreateAPIKey(ctx context.Context, key APIKey) (APIKey, error)
	GetAPIKey(ctx context.Context, keyID string) (APIKey, error)
	ListAPIKeys(ctx context.Context, botID string, limit int, offset int) ([]APIKey, error)
	UpdateAPIKey(ctx context.Context, key APIKey) (APIKey, error)
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
	limiter       RateLimiter
	sink          UsageSink
	log           *log.Helper
	rotationGrace time.Duration
}

// NewAPIMgmtUsecase creates a new APIMgmtUsecase
func NewAPIMgmtUsecase(repo APIMgmtRepo, limiter RateLimiter, sink UsageSink, cfg *conf.Data, logger log.Logger) *APIMgmtUsecase {
	rotationGrace := time.Duration(defaultRotationGraceMinutes) * time.Minute
	if cfg != nil && cfg.Apimgmt != nil && cfg.Apimgmt.RotationGraceMinutes > 0 {
		rotationGrace = time.Duration(cfg.Apimgmt.RotationGraceMinutes) * time.Minute
	}
	return &APIMgmtUsecase{
		repo:          repo,
		limiter:       limiter,
		sink:          sink,
		log:           log.NewHelper(logger),
		rotationGrace: rotationGrace,
	}
}

const DefaultAPIKeyHeader = "X-API-Key"

func (uc *APIMgmtUsecase) CreateAPIKey(ctx context.Context, name string, botID string, scopes []string, apiVersions []string, quotaDaily int32, qpsLimit int32) (APIKey, string, error) {
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
	rawKey, keyHash := generateAPIKey()
	now := time.Now()
	key := APIKey{
		ID:          uuid.NewString(),
		BotID:       botID,
		Name:        name,
		KeyHash:     keyHash,
		APIVersions: apiVersions,
		Scopes:      scopes,
		Status:      APIKeyStatusActive,
		QuotaDaily:  quotaDaily,
		QPSLimit:    qpsLimit,
		CreatedAt:   now,
	}
	created, err := uc.repo.CreateAPIKey(ctx, key)
	if err != nil {
		return APIKey{}, "", err
	}
	return created, rawKey, nil
}

func (uc *APIMgmtUsecase) ListAPIKeys(ctx context.Context, botID string, limit int, offset int) ([]APIKey, error) {
	limit, offset = paging.Normalize(limit, offset)
	return uc.repo.ListAPIKeys(ctx, strings.TrimSpace(botID), limit, offset)
}

func (uc *APIMgmtUsecase) UpdateAPIKey(ctx context.Context, keyID string, name string, status string, scopes []string, apiVersions []string, quotaDaily *int32, qpsLimit *int32) (APIKey, error) {
	keyID = strings.TrimSpace(keyID)
	if keyID == "" {
		return APIKey{}, errors.BadRequest("API_KEY_ID_MISSING", "api key id missing")
	}
	if strings.TrimSpace(name) == "" && strings.TrimSpace(status) == "" && len(scopes) == 0 && len(apiVersions) == 0 && quotaDaily == nil && qpsLimit == nil {
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
	filter = normalizeUsageFilter(filter, 200)
	return uc.repo.ListUsageLogs(ctx, filter)
}

func (uc *APIMgmtUsecase) GetUsageSummary(ctx context.Context, filter UsageFilter) (UsageSummary, error) {
	filter = normalizeUsageFilter(filter, 200)
	return uc.repo.GetUsageSummary(ctx, filter)
}

func (uc *APIMgmtUsecase) ExportUsageLogs(ctx context.Context, filter UsageFilter) (string, error) {
	filter = normalizeUsageFilter(filter, 10000)
	items, err := uc.repo.ListUsageLogs(ctx, filter)
	if err != nil {
		return "", err
	}
	builder := &strings.Builder{}
	builder.WriteString("id,api_key_id,tenant_id,bot_id,api_version,model,path,status_code,latency_ms,prompt_tokens,completion_tokens,total_tokens,client_ip,user_agent,created_at\n")
	for _, item := range items {
		builder.WriteString(csvEscape(item.ID))
		builder.WriteString(",")
		builder.WriteString(csvEscape(item.APIKeyID))
		builder.WriteString(",")
		builder.WriteString(csvEscape(item.TenantID))
		builder.WriteString(",")
		builder.WriteString(csvEscape(item.BotID))
		builder.WriteString(",")
		builder.WriteString(csvEscape(item.APIVersion))
		builder.WriteString(",")
		builder.WriteString(csvEscape(item.Model))
		builder.WriteString(",")
		builder.WriteString(csvEscape(item.Path))
		builder.WriteString(",")
		builder.WriteString(intToString(int(item.StatusCode)))
		builder.WriteString(",")
		builder.WriteString(intToString(int(item.LatencyMs)))
		builder.WriteString(",")
		builder.WriteString(intToString(int(item.PromptTokens)))
		builder.WriteString(",")
		builder.WriteString(intToString(int(item.CompletionTokens)))
		builder.WriteString(",")
		builder.WriteString(intToString(int(item.TotalTokens)))
		builder.WriteString(",")
		builder.WriteString(csvEscape(item.ClientIP))
		builder.WriteString(",")
		builder.WriteString(csvEscape(item.UserAgent))
		builder.WriteString(",")
		builder.WriteString(csvEscape(item.CreatedAt.Format(time.RFC3339)))
		builder.WriteString("\n")
	}
	return builder.String(), nil
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

func normalizeUsageFilter(filter UsageFilter, maxLimit int) UsageFilter {
	if filter.Limit <= 0 || filter.Limit > maxLimit {
		filter.Limit = maxLimit
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	return filter
}

func csvEscape(value string) string {
	value = strings.ReplaceAll(value, "\"", "\"\"")
	if strings.ContainsAny(value, ",\"\n") {
		return "\"" + value + "\""
	}
	return value
}

func intToString(value int) string {
	return strconv.Itoa(value)
}

// ProviderSet is apimgmt biz providers.
var ProviderSet = wire.NewSet(NewAPIMgmtUsecase, NewUsageSink)
