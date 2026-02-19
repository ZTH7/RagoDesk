package service

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	v1 "github.com/ZTH7/RagoDesk/apps/server/api/conversation/v1"
	"github.com/ZTH7/RagoDesk/apps/server/internal/ai/provider"
	analyticsbiz "github.com/ZTH7/RagoDesk/apps/server/internal/analytics/biz"
	apimgmtbiz "github.com/ZTH7/RagoDesk/apps/server/internal/apimgmt/biz"
	biz "github.com/ZTH7/RagoDesk/apps/server/internal/conversation/biz"
	"github.com/ZTH7/RagoDesk/apps/server/internal/kit/tenant"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/google/wire"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ConversationService handles conversation service layer.
type ConversationService struct {
	v1.UnimplementedConversationServer
	v1.UnimplementedConsoleConversationServer

	uc  *biz.ConversationUsecase
	api *apimgmtbiz.APIMgmtUsecase
	ana *analyticsbiz.AnalyticsUsecase
}

// NewConversationService creates a new ConversationService
func NewConversationService(uc *biz.ConversationUsecase, api *apimgmtbiz.APIMgmtUsecase, ana *analyticsbiz.AnalyticsUsecase) *ConversationService {
	return &ConversationService{uc: uc, api: api, ana: ana}
}

func (s *ConversationService) CreateSession(ctx context.Context, req *v1.CreateSessionRequest) (*v1.CreateSessionResponse, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	start := time.Now()
	operation := operationFromContext(ctx)
	apiVersion := apiVersionFromOperation(operation)
	clientIP, userAgent := clientInfoFromContext(ctx)
	ctx, key, err := s.requireAPIKey(ctx, apimgmtbiz.ScopeConversation, apiVersion)
	if err != nil {
		s.recordUsage(ctx, key, operation, apiVersion, provider.LLMUsage{}, err, start, clientIP, userAgent)
		return nil, err
	}
	var callErr error
	defer func() {
		s.recordUsage(ctx, key, operation, apiVersion, provider.LLMUsage{}, callErr, start, clientIP, userAgent)
	}()
	meta := structToMap(req.Metadata)
	session, callErr := s.uc.CreateSession(ctx, key.BotID, req.UserExternalId, meta)
	if callErr != nil {
		return nil, callErr
	}
	if s.ana != nil {
		s.ana.RecordSessionEvent(ctx, analyticsbiz.AnalyticsEvent{
			TenantID:  key.TenantID,
			BotID:     key.BotID,
			SessionID: session.ID,
			CreatedAt: time.Now(),
		}, analyticsbiz.EventSessionOpen)
	}
	return &v1.CreateSessionResponse{Session: toAPISession(session)}, nil
}

func (s *ConversationService) GetSession(ctx context.Context, req *v1.GetSessionRequest) (*v1.GetSessionResponse, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	start := time.Now()
	operation := operationFromContext(ctx)
	apiVersion := apiVersionFromOperation(operation)
	clientIP, userAgent := clientInfoFromContext(ctx)
	ctx, key, err := s.requireAPIKey(ctx, apimgmtbiz.ScopeConversation, apiVersion)
	if err != nil {
		s.recordUsage(ctx, key, operation, apiVersion, provider.LLMUsage{}, err, start, clientIP, userAgent)
		return nil, err
	}
	var callErr error
	defer func() {
		s.recordUsage(ctx, key, operation, apiVersion, provider.LLMUsage{}, callErr, start, clientIP, userAgent)
	}()
	session, messages, callErr := s.uc.GetSession(ctx, req.SessionId, req.IncludeMessages, int(req.Limit), int(req.Offset))
	if callErr != nil {
		return nil, callErr
	}
	if key.BotID != "" && session.BotID != "" && key.BotID != session.BotID {
		callErr = errors.Forbidden("SESSION_FORBIDDEN", "session bot mismatch")
		return nil, callErr
	}
	resp := &v1.GetSessionResponse{
		Session:  toAPISession(session),
		Messages: toAPIMessages(messages),
	}
	return resp, nil
}

func (s *ConversationService) CloseSession(ctx context.Context, req *v1.CloseSessionRequest) (*emptypb.Empty, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	start := time.Now()
	operation := operationFromContext(ctx)
	apiVersion := apiVersionFromOperation(operation)
	clientIP, userAgent := clientInfoFromContext(ctx)
	ctx, key, err := s.requireAPIKey(ctx, apimgmtbiz.ScopeConversation, apiVersion)
	if err != nil {
		s.recordUsage(ctx, key, operation, apiVersion, provider.LLMUsage{}, err, start, clientIP, userAgent)
		return nil, err
	}
	var callErr error
	defer func() {
		s.recordUsage(ctx, key, operation, apiVersion, provider.LLMUsage{}, callErr, start, clientIP, userAgent)
	}()
	session, _, callErr := s.uc.GetSession(ctx, req.SessionId, false, 0, 0)
	if callErr != nil {
		return nil, callErr
	}
	if key.BotID != "" && session.BotID != "" && key.BotID != session.BotID {
		callErr = errors.Forbidden("SESSION_FORBIDDEN", "session bot mismatch")
		return nil, callErr
	}
	if callErr = s.uc.CloseSession(ctx, req.SessionId, req.CloseReason); callErr != nil {
		return nil, callErr
	}
	if s.ana != nil {
		s.ana.RecordSessionEvent(ctx, analyticsbiz.AnalyticsEvent{
			TenantID:  key.TenantID,
			BotID:     key.BotID,
			SessionID: strings.TrimSpace(req.GetSessionId()),
			CreatedAt: time.Now(),
		}, analyticsbiz.EventSessionClose)
	}
	return &emptypb.Empty{}, nil
}

func (s *ConversationService) CreateFeedback(ctx context.Context, req *v1.CreateFeedbackRequest) (*emptypb.Empty, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	start := time.Now()
	operation := operationFromContext(ctx)
	apiVersion := apiVersionFromOperation(operation)
	clientIP, userAgent := clientInfoFromContext(ctx)
	ctx, key, err := s.requireAPIKey(ctx, apimgmtbiz.ScopeConversation, apiVersion)
	if err != nil {
		s.recordUsage(ctx, key, operation, apiVersion, provider.LLMUsage{}, err, start, clientIP, userAgent)
		return nil, err
	}
	var callErr error
	defer func() {
		s.recordUsage(ctx, key, operation, apiVersion, provider.LLMUsage{}, callErr, start, clientIP, userAgent)
	}()
	session, _, callErr := s.uc.GetSession(ctx, req.SessionId, false, 0, 0)
	if callErr != nil {
		return nil, callErr
	}
	if key.BotID != "" && session.BotID != "" && key.BotID != session.BotID {
		callErr = errors.Forbidden("SESSION_FORBIDDEN", "session bot mismatch")
		return nil, callErr
	}
	if callErr = s.uc.CreateFeedback(ctx, req.SessionId, req.MessageId, req.Rating, req.Comment, req.Correction); callErr != nil {
		return nil, callErr
	}
	if s.ana != nil {
		s.ana.RecordFeedback(ctx, analyticsbiz.AnalyticsEvent{
			TenantID:  key.TenantID,
			BotID:     key.BotID,
			SessionID: strings.TrimSpace(req.GetSessionId()),
			MessageID: strings.TrimSpace(req.GetMessageId()),
			Rating:    req.GetRating(),
			CreatedAt: time.Now(),
		})
	}
	return &emptypb.Empty{}, nil
}

func (s *ConversationService) ListSessions(ctx context.Context, req *v1.ListSessionsRequest) (*v1.ListSessionsResponse, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	sessions, err := s.uc.ListSessions(ctx, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, err
	}
	out := make([]*v1.Session, 0, len(sessions))
	for _, item := range sessions {
		out = append(out, toAPISession(item))
	}
	return &v1.ListSessionsResponse{Sessions: out}, nil
}

func (s *ConversationService) ListMessages(ctx context.Context, req *v1.ListMessagesRequest) (*v1.ListMessagesResponse, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	messages, err := s.uc.ListMessages(ctx, req.SessionId, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, err
	}
	return &v1.ListMessagesResponse{Messages: toAPIMessages(messages)}, nil
}

// ProviderSet is conversation service providers.
var ProviderSet = wire.NewSet(NewConversationService)

func toAPISession(session biz.Session) *v1.Session {
	if session.ID == "" {
		return nil
	}
	return &v1.Session{
		Id:             session.ID,
		TenantId:       session.TenantID,
		BotId:          session.BotID,
		Status:         session.Status,
		CloseReason:    session.CloseReason,
		UserExternalId: session.UserExternal,
		Metadata:       parseMetadata(session.Metadata),
		CreatedAt:      timestamppb.New(session.CreatedAt),
		UpdatedAt:      timestamppb.New(session.UpdatedAt),
		ClosedAt:       timeOrNil(session.ClosedAt),
	}
}

func toAPIMessages(messages []biz.Message) []*v1.Message {
	if len(messages) == 0 {
		return nil
	}
	out := make([]*v1.Message, 0, len(messages))
	for _, item := range messages {
		out = append(out, &v1.Message{
			Id:         item.ID,
			SessionId:  item.SessionID,
			Role:       item.Role,
			Content:    item.Content,
			Confidence: item.Confidence,
			References: toAPIReferences(item.References),
			CreatedAt:  timestamppb.New(item.CreatedAt),
		})
	}
	return out
}

func toAPIReferences(raw string) []*v1.Reference {
	refs := biz.DecodeReferences(raw)
	if len(refs) == 0 {
		return nil
	}
	out := make([]*v1.Reference, 0, len(refs))
	for _, ref := range refs {
		out = append(out, &v1.Reference{
			DocumentId:        ref.DocumentID,
			DocumentVersionId: ref.DocumentVersionID,
			ChunkId:           ref.ChunkID,
			Score:             ref.Score,
			Rank:              ref.Rank,
			Snippet:           ref.Snippet,
		})
	}
	return out
}

func parseMetadata(raw string) *structpb.Struct {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var data map[string]any
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil
	}
	metadata, err := structpb.NewStruct(data)
	if err != nil {
		return nil
	}
	return metadata
}

func structToMap(value *structpb.Struct) map[string]any {
	if value == nil {
		return nil
	}
	return value.AsMap()
}

func timeOrNil(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

func (s *ConversationService) requireAPIKey(ctx context.Context, scope string, apiVersion string) (context.Context, apimgmtbiz.APIKey, error) {
	if s.api == nil {
		return ctx, apimgmtbiz.APIKey{}, errors.InternalServer("API_KEY_RESOLVER_MISSING", "api key resolver missing")
	}
	tr, ok := transport.FromServerContext(ctx)
	if !ok {
		return ctx, apimgmtbiz.APIKey{}, errors.Unauthorized("API_KEY_MISSING", "api key missing")
	}
	rawKey := strings.TrimSpace(tr.RequestHeader().Get(apimgmtbiz.DefaultAPIKeyHeader))
	key, err := s.api.AuthorizeAPIKeyWithScope(ctx, rawKey, scope, apiVersion)
	if key.TenantID != "" {
		ctx = tenant.WithTenantID(ctx, key.TenantID)
	}
	if err != nil {
		return ctx, key, err
	}
	return ctx, key, nil
}

func (s *ConversationService) recordUsage(ctx context.Context, key apimgmtbiz.APIKey, operation string, apiVersion string, usage provider.LLMUsage, err error, start time.Time, clientIP string, userAgent string) {
	if s == nil || s.api == nil {
		return
	}
	status := apimgmtbiz.StatusCodeFromError(err)
	s.api.RecordUsage(ctx, key, operation, apiVersion, "", usage, status, time.Since(start), clientIP, userAgent)
}

func operationFromContext(ctx context.Context) string {
	if tr, ok := transport.FromServerContext(ctx); ok {
		return tr.Operation()
	}
	return ""
}

func clientInfoFromContext(ctx context.Context) (string, string) {
	tr, ok := transport.FromServerContext(ctx)
	if !ok {
		return "", ""
	}
	header := tr.RequestHeader()
	forwarded := strings.TrimSpace(header.Get("X-Forwarded-For"))
	if forwarded != "" {
		if parts := strings.Split(forwarded, ","); len(parts) > 0 {
			forwarded = strings.TrimSpace(parts[0])
		}
	}
	if forwarded == "" {
		forwarded = strings.TrimSpace(header.Get("X-Real-IP"))
	}
	return forwarded, strings.TrimSpace(header.Get("User-Agent"))
}

func apiVersionFromOperation(operation string) string {
	if operation == "" {
		return "v1"
	}
	if idx := strings.Index(operation, "/api/v"); idx >= 0 {
		segment := operation[idx+6:]
		digits := ""
		for _, r := range segment {
			if r < '0' || r > '9' {
				break
			}
			digits += string(r)
		}
		if digits != "" {
			return "v" + digits
		}
	}
	parts := strings.Split(operation, ".")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if len(part) > 1 && part[0] == 'v' {
			return strings.ToLower(part)
		}
	}
	return "v1"
}
