package service

import (
	"context"
	"strings"

	ragv1 "github.com/ZTH7/RAGDesk/apps/server/api/rag/v1"
	biz "github.com/ZTH7/RAGDesk/apps/server/internal/rag/biz"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/google/wire"
)

// RAGService handles rag service layer.
type RAGService struct {
	ragv1.UnimplementedRAGServer

	uc  *biz.RAGUsecase
	log *log.Helper
}

// NewRAGService creates a new RAGService.
func NewRAGService(uc *biz.RAGUsecase, logger log.Logger) *RAGService {
	return &RAGService{uc: uc, log: log.NewHelper(logger)}
}

// SendMessage handles RAG message requests.
func (s *RAGService) SendMessage(ctx context.Context, req *ragv1.SendMessageRequest) (*ragv1.SendMessageResponse, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	if s.uc != nil && s.uc.RequireAPIKey() {
		tr, ok := transport.FromServerContext(ctx)
		if !ok {
			return nil, errors.Unauthorized("API_KEY_MISSING", "api key missing")
		}
		header := s.uc.APIKeyHeader()
		apiKey := strings.TrimSpace(tr.RequestHeader().Get(header))
		if apiKey == "" {
			return nil, errors.Unauthorized("API_KEY_MISSING", "api key missing")
		}
	}
	resp, err := s.uc.SendMessage(ctx, biz.MessageRequest{
		SessionID: req.SessionId,
		BotID:     req.BotId,
		Message:   req.Message,
		TopK:      req.TopK,
		Threshold: req.Threshold,
	})
	if err != nil {
		return nil, err
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
