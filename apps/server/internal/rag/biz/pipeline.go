package biz

import (
	"context"

	"github.com/ZTH7/RAGDesk/apps/server/internal/ai/provider"
	"github.com/cloudwego/eino/compose"
)

type ragContext struct {
	req          MessageRequest
	topK         int
	threshold    float32
	kbs          []BotKnowledgeBase
	queryVector  []float32
	queryVectors [][]float32
	queryWeights []float32
	normalized   string
	queries      []string
	ranked       []scoredChunk
	chunks       map[string]ChunkMeta
	prompt       string
	reply        string
	confidence   float32
	shouldRefuse bool
	llmUsage     provider.LLMUsage
	llmModel     string
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
