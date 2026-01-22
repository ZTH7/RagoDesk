package service

import (
	"context"
	"strings"
	"time"

	ragv1 "github.com/ZTH7/RAGDesk/apps/server/api/rag/v1"
	apimgmtbiz "github.com/ZTH7/RAGDesk/apps/server/internal/apimgmt/biz"
	convbiz "github.com/ZTH7/RAGDesk/apps/server/internal/conversation/biz"
	biz "github.com/ZTH7/RAGDesk/apps/server/internal/rag/biz"
	"github.com/ZTH7/RAGDesk/apps/server/internal/tenant"
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
	log  *log.Helper
}

// NewRAGService creates a new RAGService.
func NewRAGService(uc *biz.RAGUsecase, conv *convbiz.ConversationUsecase, api *apimgmtbiz.APIMgmtUsecase, logger log.Logger) *RAGService {
	return &RAGService{uc: uc, conv: conv, api: api, log: log.NewHelper(logger)}
}

// SendMessage handles RAG message requests.
func (s *RAGService) SendMessage(ctx context.Context, req *ragv1.SendMessageRequest) (*ragv1.SendMessageResponse, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	start := time.Now()
	operation := operationFromContext(ctx)
	ctx, key, err := s.requireAPIKey(ctx, apimgmtbiz.ScopeRAG)
	if err != nil {
		s.recordUsage(ctx, key, operation, err, start)
		return nil, err
	}
	var callErr error
	defer func() {
		s.recordUsage(ctx, key, operation, callErr, start)
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
	if s.conv != nil && strings.TrimSpace(req.SessionId) != "" {
		if callErr = s.conv.RecordRAGExchange(
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

func (s *RAGService) requireAPIKey(ctx context.Context, scope string) (context.Context, apimgmtbiz.APIKey, error) {
	if s.api == nil {
		return ctx, apimgmtbiz.APIKey{}, errors.InternalServer("API_KEY_RESOLVER_MISSING", "api key resolver missing")
	}
	tr, ok := transport.FromServerContext(ctx)
	if !ok {
		return ctx, apimgmtbiz.APIKey{}, errors.Unauthorized("API_KEY_MISSING", "api key missing")
	}
	rawKey := strings.TrimSpace(tr.RequestHeader().Get(apimgmtbiz.DefaultAPIKeyHeader))
	key, err := s.api.AuthorizeAPIKeyWithScope(ctx, rawKey, scope)
	if key.TenantID != "" {
		ctx = tenant.WithTenantID(ctx, key.TenantID)
	}
	if err != nil {
		return ctx, key, err
	}
	return ctx, key, nil
}

func (s *RAGService) recordUsage(ctx context.Context, key apimgmtbiz.APIKey, operation string, err error, start time.Time) {
	if s == nil || s.api == nil {
		return
	}
	status := apimgmtbiz.StatusCodeFromError(err)
	s.api.RecordUsage(ctx, key, operation, status, time.Since(start))
}

func operationFromContext(ctx context.Context) string {
	if tr, ok := transport.FromServerContext(ctx); ok {
		return tr.Operation()
	}
	return ""
}
