package biz

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/sync/errgroup"
)

func (uc *RAGUsecase) initContext(_ context.Context, req MessageRequest) (*ragContext, error) {
	req.BotID = strings.TrimSpace(req.BotID)
	req.Message = strings.TrimSpace(req.Message)
	if req.BotID == "" {
		return nil, errors.BadRequest("BOT_ID_MISSING", "bot_id missing")
	}
	if req.Message == "" {
		return nil, errors.BadRequest("MESSAGE_MISSING", "message missing")
	}
	topK := int(req.TopK)
	if topK <= 0 {
		topK = uc.opts.topK
	}
	threshold := req.Threshold
	if threshold <= 0 {
		threshold = uc.opts.scoreThreshold
	}
	normalized := normalizeQuery(req.Message)
	if normalized == "" {
		normalized = strings.TrimSpace(req.Message)
	}
	queries := dedupeQueries([]string{normalized})
	weights := alignQueryWeights(queries, []float32{1})
	return &ragContext{
		req:          req,
		topK:         topK,
		threshold:    threshold,
		normalized:   normalized,
		queries:      queries,
		queryWeights: weights,
	}, nil
}

func (uc *RAGUsecase) resolveContext(ctx context.Context, rc *ragContext) (*ragContext, error) {
	if rc == nil || rc.shouldRefuse {
		return rc, nil
	}
	ctx, span := uc.startSpan(ctx, "rag.resolve", attribute.String("rag.bot_id", rc.req.BotID))
	defer span.End()
	start := time.Now()
	kbs, err := uc.kbRepo.ResolveBotKnowledgeBases(ctx, rc.req.BotID)
	uc.logStep("resolve", start, err)
	if err != nil {
		uc.recordSpanError(span, err)
		return rc, err
	}
	rc.kbs = kbs
	span.SetAttributes(attribute.Int("rag.kb_count", len(kbs)))
	if len(kbs) == 0 {
		rc.shouldRefuse = true
	}
	return rc, nil
}

func (uc *RAGUsecase) embedContext(ctx context.Context, rc *ragContext) (*ragContext, error) {
	if rc == nil || rc.shouldRefuse {
		return rc, nil
	}
	ctx, span := uc.startSpan(ctx, "rag.embed", attribute.String("rag.model", uc.embedder.Model()))
	defer span.End()
	embedCtx, cancel := withTimeout(ctx, uc.opts.embeddingConfig.TimeoutMs)
	defer cancel()
	start := time.Now()
	vecs, err := uc.embedder.Embed(embedCtx, rc.queries)
	uc.logStep("embed", start, err)
	if err != nil {
		uc.recordSpanError(span, err)
		return rc, err
	}
	if len(vecs) == 0 {
		err := errors.InternalServer("EMBEDDING_EMPTY", "embedding empty")
		uc.recordSpanError(span, err)
		return rc, err
	}
	if len(vecs) != len(rc.queries) {
		err := errors.InternalServer("EMBEDDING_COUNT_MISMATCH", "embedding count mismatch")
		uc.recordSpanError(span, err)
		return rc, err
	}
	rc.queryVector = vecs[0]
	rc.queryVectors = vecs
	span.SetAttributes(attribute.Int("rag.embedding_dim", len(rc.queryVector)))
	return rc, nil
}

func (uc *RAGUsecase) retrieveContext(ctx context.Context, rc *ragContext) (*ragContext, error) {
	if rc == nil || rc.shouldRefuse {
		return rc, nil
	}
	ctx, span := uc.startSpan(ctx, "rag.retrieve", attribute.Int("rag.top_k", rc.topK))
	defer span.End()
	retrieveCtx, cancel := withTimeout(ctx, uc.opts.retrieveTimeoutMs)
	defer cancel()
	start := time.Now()
	minScore := deriveRetrieveThreshold(rc.threshold)
	span.SetAttributes(attribute.Float64("rag.retrieve_min_score", float64(minScore)))
	scored := make([]scoredChunk, 0)
	var mu sync.Mutex
	var errCount int
	var firstErr error
	group, groupCtx := errgroup.WithContext(retrieveCtx)
	limit := uc.opts.retrieveConcurrency
	if limit <= 0 {
		limit = 1
	}
	group.SetLimit(limit)
	for qIdx, vec := range rc.queryVectors {
		vec := vec
		qWeight := float32(1)
		if qIdx < len(rc.queryWeights) {
			qWeight = rc.queryWeights[qIdx]
		}
		for _, kb := range rc.kbs {
			kb := kb
			kbID := strings.TrimSpace(kb.KBID)
			if kbID == "" {
				continue
			}
			group.Go(func() error {
				results, err := uc.vectorRepo.Search(groupCtx, VectorSearchRequest{
					Vector:         vec,
					KBID:           kbID,
					TopK:           rc.topK,
					ScoreThreshold: minScore,
				})
				if err != nil {
					mu.Lock()
					errCount++
					if firstErr == nil {
						firstErr = err
					}
					mu.Unlock()
					return nil
				}
				weight := kb.Weight
				if weight <= 0 {
					weight = 1
				}
				local := make([]scoredChunk, 0, len(results))
				for _, item := range results {
					if minScore > 0 && item.Score < minScore {
						continue
					}
					vecScore := item.Score * float32(weight) * qWeight
					local = append(local, scoredChunk{
						result:      item,
						vectorScore: vecScore,
						score:       vecScore,
					})
				}
				if len(local) == 0 {
					return nil
				}
				mu.Lock()
				scored = append(scored, local...)
				mu.Unlock()
				return nil
			})
		}
	}
	if err := group.Wait(); err != nil {
		uc.logStep("retrieve", start, err)
		uc.recordSpanError(span, err)
		return rc, err
	}
	if errCount > 0 {
		span.SetAttributes(attribute.Int("rag.retrieve_error_count", errCount))
		if firstErr != nil {
			uc.logStep("retrieve", start, firstErr)
		} else {
			uc.logStep("retrieve", start, nil)
		}
	} else {
		uc.logStep("retrieve", start, nil)
	}
	if len(scored) == 0 && errCount > 0 && firstErr != nil {
		uc.recordSpanError(span, firstErr)
		return rc, firstErr
	}
	rc.ranked = rankAndFilter(scored, rc.topK)
	span.SetAttributes(attribute.Int("rag.candidate_count", len(rc.ranked)))
	return rc, nil
}

func (uc *RAGUsecase) loadChunksContext(ctx context.Context, rc *ragContext) (*ragContext, error) {
	if rc == nil || rc.shouldRefuse {
		return rc, nil
	}
	if len(rc.ranked) == 0 {
		return rc, nil
	}
	ctx, span := uc.startSpan(ctx, "rag.load_chunks", attribute.Int("rag.chunk_count", len(rc.ranked)))
	defer span.End()
	seen := make(map[string]struct{}, len(rc.ranked))
	chunkIDs := make([]string, 0, len(rc.ranked))
	for _, item := range rc.ranked {
		if item.result.ChunkID == "" {
			continue
		}
		if _, ok := seen[item.result.ChunkID]; ok {
			continue
		}
		seen[item.result.ChunkID] = struct{}{}
		chunkIDs = append(chunkIDs, item.result.ChunkID)
	}
	start := time.Now()
	chunks, err := uc.chunkRepo.LoadChunks(ctx, chunkIDs)
	uc.logStep("chunks", start, err)
	if err != nil {
		uc.recordSpanError(span, err)
		return rc, err
	}
	rc.chunks = chunks
	return rc, nil
}
