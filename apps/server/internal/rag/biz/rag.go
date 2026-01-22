package biz

import (
	"context"
	"strings"

	"github.com/ZTH7/RAGDesk/apps/server/internal/ai/provider"
	"github.com/ZTH7/RAGDesk/apps/server/internal/conf"
	"github.com/cloudwego/eino/compose"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// MessageRequest represents an online RAG query.
type MessageRequest struct {
	SessionID string
	BotID     string
	Message   string
	TopK      int32
	Threshold float32
}

// MessageResponse represents a RAG answer.
type MessageResponse struct {
	Reply      string
	Confidence float32
	References References
	Refused    bool
	Model      string
	Usage      provider.LLMUsage
}

// BotKnowledgeBase describes bot knowledge base binding.
type BotKnowledgeBase struct {
	KBID   string
	Weight float64
}

// VectorSearchRequest describes a vector search input.
type VectorSearchRequest struct {
	Vector         []float32
	KBID           string
	TopK           int
	ScoreThreshold float32
}

// VectorSearchResult describes a vector search output.
type VectorSearchResult struct {
	ChunkID           string
	DocumentID        string
	DocumentVersionID string
	KBID              string
	Score             float32
}

// ChunkMeta contains chunk content and metadata.
type ChunkMeta struct {
	ChunkID           string
	DocumentID        string
	DocumentVersionID string
	KBID              string
	Content           string
	Section           string
	PageNo            int32
	SourceURI         string
}

// BotKBResolver resolves bot knowledge base bindings.
type BotKBResolver interface {
	ResolveBotKnowledgeBases(ctx context.Context, botID string) ([]BotKnowledgeBase, error)
}

// VectorSearcher handles vector search.
type VectorSearcher interface {
	Search(ctx context.Context, req VectorSearchRequest) ([]VectorSearchResult, error)
}

// ChunkLoader loads chunk metadata.
type ChunkLoader interface {
	LoadChunks(ctx context.Context, chunkIDs []string) (map[string]ChunkMeta, error)
}

// RAGUsecase handles rag business logic.
type RAGUsecase struct {
	kbRepo     BotKBResolver
	vectorRepo VectorSearcher
	chunkRepo  ChunkLoader
	log        *log.Helper
	pipeline   compose.Runnable[MessageRequest, MessageResponse]

	embedder provider.Provider
	llm      provider.LLMProvider
	opts     ragOptions
}

// NewRAGUsecase creates a new RAGUsecase.
func NewRAGUsecase(kbRepo BotKBResolver, vectorRepo VectorSearcher, chunkRepo ChunkLoader, cfg *conf.Data, logger log.Logger) (*RAGUsecase, error) {
	opts := loadRAGOptions(cfg)
	embedder := provider.NewProvider(opts.embeddingConfig)
	llm := provider.NewLLMProvider(provider.LLMConfig{
		Provider:  opts.llmProvider,
		Endpoint:  opts.llmEndpoint,
		APIKey:    opts.llmAPIKey,
		Model:     opts.llmModel,
		TimeoutMs: opts.llmTimeoutMs,
	})
	uc := &RAGUsecase{
		kbRepo:     kbRepo,
		vectorRepo: vectorRepo,
		chunkRepo:  chunkRepo,
		log:        log.NewHelper(logger),
		embedder:   embedder,
		llm:        llm,
		opts:       opts,
	}
	pipeline, err := uc.buildPipeline()
	if err != nil {
		return nil, err
	}
	uc.pipeline = pipeline
	return uc, nil
}

// SendMessage handles a RAG request and returns the response.
func (uc *RAGUsecase) SendMessage(ctx context.Context, req MessageRequest) (MessageResponse, error) {
	if uc == nil || uc.kbRepo == nil || uc.vectorRepo == nil || uc.chunkRepo == nil {
		return MessageResponse{}, errors.InternalServer("RAG_DEPENDENCY_MISSING", "rag dependency missing")
	}
	ctx, cancel := withTimeout(ctx, uc.opts.ragTimeoutMs)
	defer cancel()

	return uc.pipeline.Invoke(ctx, req)
}

func buildReferences(ranked []scoredChunk, chunks map[string]ChunkMeta) References {
	if len(ranked) == 0 {
		return nil
	}
	refs := make(References, 0, len(ranked))
	for idx, item := range ranked {
		meta := ChunkMeta{}
		if chunks != nil {
			meta = chunks[item.result.ChunkID]
		}
		refs = append(refs, Reference{
			DocumentID:        pickString(meta.DocumentID, item.result.DocumentID),
			DocumentVersionID: pickString(meta.DocumentVersionID, item.result.DocumentVersionID),
			ChunkID:           item.result.ChunkID,
			Score:             item.score,
			Rank:              int32(idx + 1),
			Snippet:           truncateText(meta.Content, 200),
		})
	}
	return refs
}

func pickString(primary string, fallback string) string {
	if strings.TrimSpace(primary) != "" {
		return primary
	}
	return fallback
}

// ProviderSet is rag biz providers.
var ProviderSet = wire.NewSet(NewRAGUsecase)
