package biz

import (
	"context"

	"github.com/ZTH7/RagoDesk/apps/server/internal/ai/provider"
	"github.com/go-kratos/kratos/v2/errors"
)

func newEmbeddingProvider(opts ingestionOptions) provider.Provider {
	cfg := provider.Config{
		Provider:  opts.embeddingProvider,
		Endpoint:  opts.embeddingEndpoint,
		APIKey:    opts.embeddingAPIKey,
		Model:     opts.embeddingModel,
		Dim:       opts.embeddingDim,
		TimeoutMs: opts.embeddingTimeoutMs,
		Proxy:     opts.proxy,
	}
	return provider.NewProvider(cfg)
}

func (uc *KnowledgeUsecase) embedChunks(ctx context.Context, chunks []DocChunk) ([]EmbeddedChunk, error) {
	if len(chunks) == 0 {
		return nil, nil
	}
	if uc.embedder == nil {
		uc.embedder = provider.NewProvider(provider.Config{
			Provider: defaultEmbeddingProvider,
			Model:    defaultEmbeddingModel,
			Dim:      defaultEmbeddingDim,
		})
	}
	batchSize := uc.embeddingBatchSize
	if batchSize <= 0 {
		batchSize = 64
	}
	out := make([]EmbeddedChunk, 0, len(chunks))
	for start := 0; start < len(chunks); start += batchSize {
		end := start + batchSize
		if end > len(chunks) {
			end = len(chunks)
		}
		texts := make([]string, 0, end-start)
		for _, ch := range chunks[start:end] {
			texts = append(texts, ch.Content)
		}
		vectors, err := uc.embedder.Embed(ctx, texts)
		if err != nil {
			return nil, err
		}
		if len(vectors) != len(texts) {
			return nil, errors.InternalServer("EMBEDDING_COUNT_MISMATCH", "embedding count mismatch")
		}
		for i, ch := range chunks[start:end] {
			out = append(out, EmbeddedChunk{
				Chunk:  ch,
				Vector: vectors[i],
			})
		}
	}
	return out, nil
}
