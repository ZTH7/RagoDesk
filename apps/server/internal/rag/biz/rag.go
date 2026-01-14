package biz

import (
	"context"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/ZTH7/RAGDesk/apps/server/internal/ai/provider"
	"github.com/ZTH7/RAGDesk/apps/server/internal/conf"
	"github.com/cloudwego/eino/compose"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
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
}

// BotKnowledgeBase describes bot knowledge base binding.
type BotKnowledgeBase struct {
	KBID     string
	Priority int32
	Weight   float64
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

// RAGRepo is a repository interface.
type RAGRepo interface {
	ResolveBotKnowledgeBases(ctx context.Context, botID string) ([]BotKnowledgeBase, error)
	Search(ctx context.Context, req VectorSearchRequest) ([]VectorSearchResult, error)
	LoadChunks(ctx context.Context, chunkIDs []string) (map[string]ChunkMeta, error)
}

// RAGUsecase handles rag business logic.
type RAGUsecase struct {
	repo     RAGRepo
	log      *log.Helper
	pipeline compose.Runnable[MessageRequest, MessageResponse]

	embedder provider.Provider
	llm      provider.LLMProvider
	opts     ragOptions
}

// NewRAGUsecase creates a new RAGUsecase.
func NewRAGUsecase(repo RAGRepo, cfg *conf.Data, logger log.Logger) *RAGUsecase {
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
		repo:     repo,
		log:      log.NewHelper(logger),
		embedder: embedder,
		llm:      llm,
		opts:     opts,
	}
	pipeline, err := uc.buildPipeline()
	if err != nil {
		uc.log.Warnf("rag pipeline init failed: %v", err)
	} else {
		uc.pipeline = pipeline
	}
	return uc
}

// SendMessage handles a RAG request and returns the response.
func (uc *RAGUsecase) SendMessage(ctx context.Context, req MessageRequest) (MessageResponse, error) {
	if uc == nil || uc.repo == nil {
		return MessageResponse{}, errors.InternalServer("RAG_REPO_MISSING", "rag repo missing")
	}
	ctx, cancel := withTimeout(ctx, uc.opts.ragTimeoutMs)
	defer cancel()

	if uc.pipeline != nil {
		return uc.pipeline.Invoke(ctx, req)
	}
	return uc.runPipeline(ctx, req)
}

func (uc *RAGUsecase) RequireAPIKey() bool {
	return uc != nil && uc.opts.apiKeyRequired
}

func (uc *RAGUsecase) APIKeyHeader() string {
	if uc == nil || strings.TrimSpace(uc.opts.apiKeyHeader) == "" {
		return defaultAPIKeyHeader
	}
	return uc.opts.apiKeyHeader
}

func (uc *RAGUsecase) buildPipeline() (compose.Runnable[MessageRequest, MessageResponse], error) {
	graph := compose.NewGraph[MessageRequest, MessageResponse]()
	if err := graph.AddLambdaNode("init", compose.InvokableLambda(func(ctx context.Context, req MessageRequest) (*ragContext, error) {
		return uc.initContext(ctx, req)
	})); err != nil {
		return nil, err
	}
	if err := graph.AddLambdaNode("resolve", compose.InvokableLambda(func(ctx context.Context, rc *ragContext) (*ragContext, error) {
		return uc.resolveContext(ctx, rc)
	})); err != nil {
		return nil, err
	}
	if err := graph.AddLambdaNode("embed", compose.InvokableLambda(func(ctx context.Context, rc *ragContext) (*ragContext, error) {
		return uc.embedContext(ctx, rc)
	})); err != nil {
		return nil, err
	}
	if err := graph.AddLambdaNode("retrieve", compose.InvokableLambda(func(ctx context.Context, rc *ragContext) (*ragContext, error) {
		return uc.retrieveContext(ctx, rc)
	})); err != nil {
		return nil, err
	}
	if err := graph.AddLambdaNode("chunks", compose.InvokableLambda(func(ctx context.Context, rc *ragContext) (*ragContext, error) {
		return uc.loadChunksContext(ctx, rc)
	})); err != nil {
		return nil, err
	}
	if err := graph.AddLambdaNode("rerank", compose.InvokableLambda(func(ctx context.Context, rc *ragContext) (*ragContext, error) {
		return uc.rerankContext(ctx, rc)
	})); err != nil {
		return nil, err
	}
	if err := graph.AddLambdaNode("assess", compose.InvokableLambda(func(ctx context.Context, rc *ragContext) (*ragContext, error) {
		return uc.assessContext(ctx, rc)
	})); err != nil {
		return nil, err
	}
	if err := graph.AddLambdaNode("prompt", compose.InvokableLambda(func(ctx context.Context, rc *ragContext) (*ragContext, error) {
		return uc.promptContext(ctx, rc)
	})); err != nil {
		return nil, err
	}
	if err := graph.AddLambdaNode("llm", compose.InvokableLambda(func(ctx context.Context, rc *ragContext) (*ragContext, error) {
		return uc.llmContext(ctx, rc)
	})); err != nil {
		return nil, err
	}
	if err := graph.AddLambdaNode("output", compose.InvokableLambda(func(ctx context.Context, rc *ragContext) (MessageResponse, error) {
		return uc.buildResponse(ctx, rc)
	})); err != nil {
		return nil, err
	}
	if err := graph.AddEdge(compose.START, "init"); err != nil {
		return nil, err
	}
	if err := graph.AddEdge("init", "resolve"); err != nil {
		return nil, err
	}
	if err := graph.AddEdge("resolve", "embed"); err != nil {
		return nil, err
	}
	if err := graph.AddEdge("embed", "retrieve"); err != nil {
		return nil, err
	}
	if err := graph.AddEdge("retrieve", "chunks"); err != nil {
		return nil, err
	}
	if err := graph.AddEdge("chunks", "rerank"); err != nil {
		return nil, err
	}
	if err := graph.AddEdge("rerank", "assess"); err != nil {
		return nil, err
	}
	if err := graph.AddEdge("assess", "prompt"); err != nil {
		return nil, err
	}
	if err := graph.AddEdge("prompt", "llm"); err != nil {
		return nil, err
	}
	if err := graph.AddEdge("llm", "output"); err != nil {
		return nil, err
	}
	if err := graph.AddEdge("output", compose.END); err != nil {
		return nil, err
	}
	return graph.Compile(context.Background(), compose.WithGraphName("rag_pipeline"))
}

func (uc *RAGUsecase) runPipeline(ctx context.Context, req MessageRequest) (MessageResponse, error) {
	rc, err := uc.initContext(ctx, req)
	if err != nil {
		return MessageResponse{}, err
	}
	steps := []func(context.Context, *ragContext) (*ragContext, error){
		uc.resolveContext,
		uc.embedContext,
		uc.retrieveContext,
		uc.loadChunksContext,
		uc.rerankContext,
		uc.assessContext,
		uc.promptContext,
		uc.llmContext,
	}
	for _, step := range steps {
		rc, err = step(ctx, rc)
		if err != nil {
			return MessageResponse{}, err
		}
	}
	return uc.buildResponse(ctx, rc)
}

type ragContext struct {
	req          MessageRequest
	topK         int
	threshold    float32
	kbs          []BotKnowledgeBase
	queryVector  []float32
	normalized   string
	ranked       []scoredChunk
	chunks       map[string]ChunkMeta
	prompt       string
	reply        string
	confidence   float32
	shouldRefuse bool
}

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
	return &ragContext{req: req, topK: topK, threshold: threshold, normalized: normalized}, nil
}

func (uc *RAGUsecase) resolveContext(ctx context.Context, rc *ragContext) (*ragContext, error) {
	if rc == nil || rc.shouldRefuse {
		return rc, nil
	}
	ctx, span := uc.startSpan(ctx, "rag.resolve", attribute.String("rag.bot_id", rc.req.BotID))
	defer span.End()
	start := time.Now()
	kbs, err := uc.repo.ResolveBotKnowledgeBases(ctx, rc.req.BotID)
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
	vecs, err := uc.embedder.Embed(embedCtx, []string{rc.normalized})
	uc.logStep("embed", start, err)
	if err != nil {
		uc.recordSpanError(span, err)
		return rc, err
	}
	if len(vecs) == 0 {
		err := errors.InternalServer("EMBEDDING_EMPTY", "embedding empty")
		uc.recordSpanError(span, err)
		return rc, errors.InternalServer("EMBEDDING_EMPTY", "embedding empty")
	}
	rc.queryVector = vecs[0]
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
	scored := make([]scoredChunk, 0)
	for _, kb := range rc.kbs {
		kbID := strings.TrimSpace(kb.KBID)
		if kbID == "" {
			continue
		}
		results, err := uc.repo.Search(retrieveCtx, VectorSearchRequest{
			Vector:         rc.queryVector,
			KBID:           kbID,
			TopK:           rc.topK,
			ScoreThreshold: 0,
		})
		if err != nil {
			uc.logStep("retrieve", start, err)
			uc.recordSpanError(span, err)
			return rc, err
		}
		weight := kb.Weight
		if weight <= 0 {
			weight = 1
		}
		for _, item := range results {
			vecScore := item.Score * float32(weight)
			scored = append(scored, scoredChunk{
				result:      item,
				vectorScore: vecScore,
				score:       vecScore,
			})
		}
	}
	uc.logStep("retrieve", start, nil)
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
	chunkIDs := make([]string, 0, len(rc.ranked))
	for _, item := range rc.ranked {
		chunkIDs = append(chunkIDs, item.result.ChunkID)
	}
	start := time.Now()
	chunks, err := uc.repo.LoadChunks(ctx, chunkIDs)
	uc.logStep("chunks", start, err)
	if err != nil {
		uc.recordSpanError(span, err)
		return rc, err
	}
	rc.chunks = chunks
	return rc, nil
}

func (uc *RAGUsecase) rerankContext(ctx context.Context, rc *ragContext) (*ragContext, error) {
	if rc == nil || rc.shouldRefuse || len(rc.ranked) == 0 {
		return rc, nil
	}
	ctx, span := uc.startSpan(ctx, "rag.rerank", attribute.Float64("rag.rerank_weight", float64(uc.opts.rerankWeight)))
	defer span.End()
	start := time.Now()
	for i := range rc.ranked {
		chunk := rc.ranked[i]
		meta := rc.chunks[chunk.result.ChunkID]
		textScore := overlapScore(rc.normalized, meta.Content)
		sectionScore := overlapScore(rc.normalized, meta.Section)
		if sectionScore > 0 {
			textScore = maxFloat32(textScore, sectionScore*1.2)
		}
		chunk.textScore = textScore
		chunk.score = combineScores(chunk.vectorScore, textScore, uc.opts.rerankWeight)
		rc.ranked[i] = chunk
	}
	sort.SliceStable(rc.ranked, func(i, j int) bool {
		return rc.ranked[i].score > rc.ranked[j].score
	})
	uc.logStep("rerank", start, nil)
	return rc, nil
}

func (uc *RAGUsecase) assessContext(ctx context.Context, rc *ragContext) (*ragContext, error) {
	if rc == nil {
		return rc, nil
	}
	ctx, span := uc.startSpan(ctx, "rag.assess")
	defer span.End()
	rc.confidence = computeConfidence(rc.ranked, rc.topK)
	span.SetAttributes(attribute.Float64("rag.confidence", float64(rc.confidence)))
	if len(rc.ranked) == 0 || rc.confidence < rc.threshold {
		rc.shouldRefuse = true
		rc.reply = uc.opts.refusalMessage
	}
	return rc, nil
}

func (uc *RAGUsecase) promptContext(ctx context.Context, rc *ragContext) (*ragContext, error) {
	if rc == nil || rc.shouldRefuse {
		return rc, nil
	}
	_, span := uc.startSpan(ctx, "rag.prompt")
	defer span.End()
	rc.prompt = buildPrompt(rc.req.Message, rc.ranked, rc.chunks)
	return rc, nil
}

func (uc *RAGUsecase) llmContext(ctx context.Context, rc *ragContext) (*ragContext, error) {
	if rc == nil || rc.shouldRefuse {
		return rc, nil
	}
	ctx, span := uc.startSpan(ctx, "rag.llm", attribute.String("rag.llm_model", uc.llm.Model()))
	defer span.End()
	llmCtx, cancel := withTimeout(ctx, uc.opts.llmTimeoutMs)
	defer cancel()
	start := time.Now()
	resp, err := uc.llm.Generate(llmCtx, provider.LLMRequest{
		System:      uc.opts.systemPrompt,
		Prompt:      rc.prompt,
		Temperature: uc.opts.llmTemperature,
		MaxTokens:   uc.opts.llmMaxTokens,
	})
	uc.logStep("llm", start, err)
	if err != nil {
		uc.recordSpanError(span, err)
		return rc, err
	}
	rc.reply = strings.TrimSpace(resp.Text)
	return rc, nil
}

func (uc *RAGUsecase) buildResponse(_ context.Context, rc *ragContext) (MessageResponse, error) {
	if rc == nil {
		return MessageResponse{}, errors.InternalServer("RAG_CONTEXT_MISSING", "rag context missing")
	}
	reply := strings.TrimSpace(rc.reply)
	if reply == "" {
		reply = uc.opts.refusalMessage
	}
	return MessageResponse{
		Reply:      reply,
		Confidence: rc.confidence,
		References: buildReferences(rc.ranked, rc.chunks),
	}, nil
}

type scoredChunk struct {
	result      VectorSearchResult
	vectorScore float32
	textScore   float32
	score       float32
}

func rankAndFilter(items []scoredChunk, topK int) []scoredChunk {
	if len(items) == 0 {
		return nil
	}
	byChunk := make(map[string]scoredChunk, len(items))
	for _, item := range items {
		if item.result.ChunkID == "" {
			continue
		}
		prev, ok := byChunk[item.result.ChunkID]
		if !ok || item.score > prev.score {
			byChunk[item.result.ChunkID] = item
		}
	}
	merged := make([]scoredChunk, 0, len(byChunk))
	for _, item := range byChunk {
		merged = append(merged, item)
	}
	sort.SliceStable(merged, func(i, j int) bool {
		return merged[i].score > merged[j].score
	})
	if topK > 0 && len(merged) > topK {
		merged = merged[:topK]
	}
	return merged
}

func computeConfidence(ranked []scoredChunk, topK int) float32 {
	if len(ranked) == 0 {
		return 0
	}
	limit := 3
	if len(ranked) < limit {
		limit = len(ranked)
	}
	var sum float32
	for i := 0; i < limit; i++ {
		sum += ranked[i].score
	}
	avg := sum / float32(limit)
	coverage := float32(1)
	if topK > 0 {
		coverage = float32(len(ranked)) / float32(topK)
		if coverage > 1 {
			coverage = 1
		}
	}
	conf := 0.8*avg + 0.2*coverage
	if conf < 0 {
		return 0
	}
	if conf > 1 {
		return 1
	}
	return conf
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

func combineScores(vectorScore float32, textScore float32, weight float32) float32 {
	if weight <= 0 {
		return vectorScore
	}
	if weight >= 1 {
		return textScore
	}
	return vectorScore*(1-weight) + textScore*weight
}

func overlapScore(question string, content string) float32 {
	qTokens := tokenSet(question, 64)
	if len(qTokens) == 0 {
		return 0
	}
	cTokens := tokenSet(content, 256)
	if len(cTokens) == 0 {
		return 0
	}
	matched := 0
	for token := range qTokens {
		if _, ok := cTokens[token]; ok {
			matched++
		}
	}
	return float32(matched) / float32(len(qTokens))
}

func tokenSet(text string, limit int) map[string]struct{} {
	if limit <= 0 {
		limit = 64
	}
	fields := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsNumber(r))
	})
	out := make(map[string]struct{})
	count := 0
	for _, token := range fields {
		if len(token) < 2 {
			continue
		}
		out[token] = struct{}{}
		count++
		if count >= limit {
			break
		}
	}
	return out
}

func normalizeQuery(text string) string {
	if strings.TrimSpace(text) == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(text))
	space := false
	for _, r := range strings.ToLower(text) {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			b.WriteRune(r)
			space = false
			continue
		}
		if unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r) {
			if !space {
				b.WriteRune(' ')
				space = true
			}
			continue
		}
	}
	return strings.TrimSpace(b.String())
}

func maxFloat32(a float32, b float32) float32 {
	if a >= b {
		return a
	}
	return b
}

func withTimeout(ctx context.Context, timeoutMs int) (context.Context, context.CancelFunc) {
	if timeoutMs <= 0 {
		return ctx, func() {}
	}
	d := time.Duration(timeoutMs) * time.Millisecond
	if deadline, ok := ctx.Deadline(); ok {
		if time.Until(deadline) <= d {
			return ctx, func() {}
		}
	}
	return context.WithTimeout(ctx, d)
}

func (uc *RAGUsecase) logStep(step string, start time.Time, err error) {
	if uc == nil || uc.log == nil {
		return
	}
	dur := time.Since(start).Milliseconds()
	if err != nil {
		uc.log.Warnf("rag step=%s dur_ms=%d err=%v", step, dur, err)
		return
	}
	uc.log.Infof("rag step=%s dur_ms=%d", step, dur)
}

func (uc *RAGUsecase) startSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	tracer := otel.Tracer("ragdesk/rag")
	ctx, span := tracer.Start(ctx, name)
	if len(attrs) > 0 {
		span.SetAttributes(attrs...)
	}
	return ctx, span
}

func (uc *RAGUsecase) recordSpanError(span trace.Span, err error) {
	if span == nil || err == nil {
		return
	}
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// ProviderSet is rag biz providers.
var ProviderSet = wire.NewSet(NewRAGUsecase)
