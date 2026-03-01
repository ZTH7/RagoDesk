package service

import (
	"context"
	"strings"
	"time"

	ragv1 "github.com/ZTH7/RagoDesk/apps/server/api/rag/v1"
	"github.com/ZTH7/RagoDesk/apps/server/internal/ai/provider"
	analyticsbiz "github.com/ZTH7/RagoDesk/apps/server/internal/analytics/biz"
	apimgmtbiz "github.com/ZTH7/RagoDesk/apps/server/internal/apimgmt/biz"
	convbiz "github.com/ZTH7/RagoDesk/apps/server/internal/conversation/biz"
	"github.com/ZTH7/RagoDesk/apps/server/internal/kit/tenant"
	biz "github.com/ZTH7/RagoDesk/apps/server/internal/rag/biz"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/google/wire"
)

// RAGService handles rag service layer.
type RAGService struct {
	ragv1.UnimplementedRAGServer

	uc   *biz.RAGUsecase
	conv *convbiz.ConversationUsecase
	api  *apimgmtbiz.APIMgmtUsecase
	ana  *analyticsbiz.AnalyticsUsecase
	log  *log.Helper
}

// NewRAGService creates a new RAGService.
func NewRAGService(uc *biz.RAGUsecase, conv *convbiz.ConversationUsecase, api *apimgmtbiz.APIMgmtUsecase, ana *analyticsbiz.AnalyticsUsecase, logger log.Logger) *RAGService {
	return &RAGService{uc: uc, conv: conv, api: api, ana: ana, log: log.NewHelper(logger)}
}

// SendMessage handles RAG message requests.
func (s *RAGService) SendMessage(ctx context.Context, req *ragv1.SendMessageRequest) (*ragv1.SendMessageResponse, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	start := time.Now()
	operation := operationFromContext(ctx)
	apiVersion := apiVersionFromOperation(operation)
	clientIP, userAgent := clientInfoFromContext(ctx)
	ctx, key, err := s.requireAPIKey(ctx, apimgmtbiz.ScopeRAG, apiVersion)
	if err != nil {
		s.recordUsage(ctx, key, operation, apiVersion, "", provider.LLMUsage{}, err, start, clientIP, userAgent)
		return nil, err
	}
	var callErr error
	var respModel string
	var respUsage provider.LLMUsage
	var respConfidence float32
	var respRefused bool
	var respRefs biz.References
	defer func() {
		model := ""
		var usage provider.LLMUsage
		if callErr == nil && respModel != "" {
			model = respModel
			usage = respUsage
		}
		s.recordUsage(ctx, key, operation, apiVersion, model, usage, callErr, start, clientIP, userAgent)
		s.recordAnalytics(ctx, key, req, respConfidence, respRefused, respRefs, callErr, start)
	}()
	resp, callErr := s.uc.SendMessage(ctx, biz.MessageRequest{
		SessionID: req.SessionId,
		BotID:     key.BotID,
		Message:   req.Message,
		TopK:      req.TopK,
		Threshold: req.Threshold,
	})
	if callErr != nil {
		return nil, callErr
	}
	respModel = resp.Model
	respUsage = resp.Usage
	respConfidence = resp.Confidence
	respRefused = resp.Refused
	respRefs = resp.References
	if s.conv != nil && strings.TrimSpace(req.SessionId) != "" {
		var userMsgID string
		if userMsgID, callErr = s.conv.RecordRAGExchange(
			ctx,
			req.SessionId,
			key.BotID,
			req.Message,
			resp.Reply,
			resp.Confidence,
			resp.Refused,
			convbiz.EncodeReferences(toConversationReferences(resp.References)),
		); callErr != nil {
			return nil, callErr
		}
		s.recordMessageAnalytics(ctx, key, req.GetSessionId(), userMsgID, req.GetMessage())
	}
	return &ragv1.SendMessageResponse{
		Reply:      resp.Reply,
		Confidence: resp.Confidence,
		References: toAPIReferences(resp.References),
	}, nil
}

func toAPIReferences(refs biz.References) []*ragv1.Reference {
	if len(refs) == 0 {
		return nil
	}
	out := make([]*ragv1.Reference, 0, len(refs))
	for _, item := range refs {
		out = append(out, &ragv1.Reference{
			DocumentId:        item.DocumentID,
			DocumentVersionId: item.DocumentVersionID,
			ChunkId:           item.ChunkID,
			Score:             item.Score,
			Rank:              item.Rank,
			Snippet:           item.Snippet,
		})
	}
	return out
}

// ProviderSet is rag service providers.
var ProviderSet = wire.NewSet(NewRAGService)

func toConversationReferences(refs biz.References) []convbiz.Reference {
	if len(refs) == 0 {
		return nil
	}
	out := make([]convbiz.Reference, 0, len(refs))
	for _, item := range refs {
		out = append(out, convbiz.Reference{
			DocumentID:        item.DocumentID,
			DocumentVersionID: item.DocumentVersionID,
			ChunkID:           item.ChunkID,
			Score:             item.Score,
			Rank:              item.Rank,
			Snippet:           item.Snippet,
		})
	}
	return out
}

func (s *RAGService) requireAPIKey(ctx context.Context, scope string, apiVersion string) (context.Context, apimgmtbiz.APIKey, error) {
	if s.api == nil {
		return ctx, apimgmtbiz.APIKey{}, errors.InternalServer("API_KEY_RESOLVER_MISSING", "api key resolver missing")
	}
	tr, ok := transport.FromServerContext(ctx)
	if !ok {
		return ctx, apimgmtbiz.APIKey{}, errors.Unauthorized("API_KEY_MISSING", "api key missing")
	}
	header := tr.RequestHeader()
	rawKey := strings.TrimSpace(header.Get(apimgmtbiz.DefaultAPIKeyHeader))
	chatKey := strings.TrimSpace(header.Get(apimgmtbiz.DefaultPublicChatHeader))
	var (
		key apimgmtbiz.APIKey
		err error
	)
	switch {
	case rawKey != "":
		key, err = s.api.AuthorizeAPIKeyWithScope(ctx, rawKey, scope, apiVersion)
	case chatKey != "":
		key, err = s.api.AuthorizePublicChatIDWithScope(ctx, chatKey, scope, apiVersion)
	default:
		err = errors.Unauthorized("API_KEY_MISSING", "api key missing")
	}
	if key.TenantID != "" {
		ctx = tenant.WithTenantID(ctx, key.TenantID)
	}
	if err != nil {
		return ctx, key, err
	}
	return ctx, key, nil
}

func (s *RAGService) recordUsage(ctx context.Context, key apimgmtbiz.APIKey, operation string, apiVersion string, model string, usage provider.LLMUsage, err error, start time.Time, clientIP string, userAgent string) {
	if s == nil || s.api == nil {
		return
	}
	status := apimgmtbiz.StatusCodeFromError(err)
	s.api.RecordUsage(ctx, key, operation, apiVersion, model, usage, status, time.Since(start), clientIP, userAgent)
}

func (s *RAGService) recordAnalytics(ctx context.Context, key apimgmtbiz.APIKey, req *ragv1.SendMessageRequest, confidence float32, refused bool, refs biz.References, err error, start time.Time) {
	if s == nil || s.ana == nil || req == nil {
		return
	}
	hit := len(refs) > 0 && !refused
	status := apimgmtbiz.StatusCodeFromError(err)
	s.ana.RecordRAGEvent(ctx, analyticsbiz.AnalyticsEvent{
		TenantID:   key.TenantID,
		BotID:      key.BotID,
		SessionID:  strings.TrimSpace(req.GetSessionId()),
		Query:      req.GetMessage(),
		Hit:        hit,
		Confidence: float64(confidence),
		LatencyMs:  int32(time.Since(start).Milliseconds()),
		StatusCode: status,
		CreatedAt:  time.Now(),
	})
	if err == nil {
		s.ana.RecordRetrievalEvent(ctx, analyticsbiz.AnalyticsEvent{
			TenantID:   key.TenantID,
			BotID:      key.BotID,
			SessionID:  strings.TrimSpace(req.GetSessionId()),
			Query:      req.GetMessage(),
			Hit:        hit,
			Confidence: float64(confidence),
			LatencyMs:  int32(time.Since(start).Milliseconds()),
			StatusCode: status,
			CreatedAt:  time.Now(),
		})
	}
}

func (s *RAGService) recordMessageAnalytics(ctx context.Context, key apimgmtbiz.APIKey, sessionID string, userMsgID string, userMessage string) {
	if s == nil || s.ana == nil {
		return
	}
	if strings.TrimSpace(userMsgID) != "" {
		s.ana.RecordMessageEvent(ctx, analyticsbiz.AnalyticsEvent{
			TenantID:  key.TenantID,
			BotID:     key.BotID,
			MessageID: userMsgID,
			SessionID: strings.TrimSpace(sessionID),
			Query:     userMessage,
			CreatedAt: time.Now(),
		})
	}
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
