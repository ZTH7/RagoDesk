package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// RAG domain model (placeholder)
type RAG struct {
	ID string
}

// IsolationMode defines vector DB isolation strategy.
type IsolationMode string

const (
	IsolationModeCollection IsolationMode = "collection"
	IsolationModePayload    IsolationMode = "payload"
)

// VectorSearchRequest describes a vector search input.
type VectorSearchRequest struct {
	BotID string
	Query string
	TopK  int32
}

// VectorSearchResult describes a vector search output.
type VectorSearchResult struct {
	ChunkID string
	KBID    string
	Score   float32
}

// RAGRepo is a repository interface (placeholder)
type RAGRepo interface {
	Ping(context.Context) error
	Search(ctx context.Context, req VectorSearchRequest, mode IsolationMode) ([]VectorSearchResult, error)
}

// RAGUsecase handles rag business logic (placeholder)
type RAGUsecase struct {
	repo RAGRepo
	log  *log.Helper
}

// NewRAGUsecase creates a new RAGUsecase
func NewRAGUsecase(repo RAGRepo, logger log.Logger) *RAGUsecase {
	return &RAGUsecase{repo: repo, log: log.NewHelper(logger)}
}

// ProviderSet is rag biz providers.
var ProviderSet = wire.NewSet(NewRAGUsecase)
