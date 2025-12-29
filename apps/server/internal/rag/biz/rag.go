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

// RAGRepo is a repository interface (placeholder)
type RAGRepo interface {
	Ping(context.Context) error
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
