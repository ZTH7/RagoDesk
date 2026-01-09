package biz

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
)

// EmbeddingProvider embeds text into vector representations.
type EmbeddingProvider interface {
	Embed(ctx context.Context, inputs []string) ([][]float32, error)
	Model() string
	Dim() int
}

func newEmbeddingProvider(opts ingestionOptions) EmbeddingProvider {
	switch strings.ToLower(strings.TrimSpace(opts.embeddingProvider)) {
	case "", "fake":
		return fakeEmbeddingProvider{
			model: opts.embeddingModel,
			dim:   opts.embeddingDim,
		}
	case "openai", "http":
		endpoint := strings.TrimSpace(opts.embeddingEndpoint)
		if endpoint == "" {
			return fakeEmbeddingProvider{
				model: opts.embeddingModel,
				dim:   opts.embeddingDim,
			}
		}
		return &openAIEmbeddingProvider{
			endpoint: strings.TrimRight(endpoint, "/"),
			apiKey:   opts.embeddingAPIKey,
			model:    opts.embeddingModel,
			dim:      opts.embeddingDim,
			client: &http.Client{
				Timeout: time.Duration(opts.embeddingTimeoutMs) * time.Millisecond,
			},
		}
	default:
		// fallback to fake provider for now
		return fakeEmbeddingProvider{
			model: opts.embeddingModel,
			dim:   opts.embeddingDim,
		}
	}
}

type fakeEmbeddingProvider struct {
	model string
	dim   int
}

func (p fakeEmbeddingProvider) Embed(ctx context.Context, inputs []string) ([][]float32, error) {
	out := make([][]float32, 0, len(inputs))
	for _, text := range inputs {
		out = append(out, deterministicEmbedding(text, p.dim))
	}
	return out, nil
}

func (p fakeEmbeddingProvider) Model() string {
	if p.model == "" {
		return defaultEmbeddingModel
	}
	return p.model
}

func (p fakeEmbeddingProvider) Dim() int {
	if p.dim <= 0 {
		return defaultEmbeddingDim
	}
	return p.dim
}

type openAIEmbeddingProvider struct {
	endpoint string
	apiKey   string
	model    string
	dim      int
	client   *http.Client
}

type openAIEmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type openAIEmbeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}

func (p *openAIEmbeddingProvider) Embed(ctx context.Context, inputs []string) ([][]float32, error) {
	if p == nil || p.endpoint == "" {
		return nil, errors.InternalServer("EMBEDDING_ENDPOINT_MISSING", "embedding endpoint missing")
	}
	if p.client == nil {
		p.client = &http.Client{Timeout: 15 * time.Second}
	}
	reqBody := openAIEmbeddingRequest{
		Model: p.Model(),
		Input: inputs,
	}
	raw, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	url := p.endpoint
	if !strings.HasSuffix(url, "/embeddings") {
		url = url + "/embeddings"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.InternalServer("EMBEDDING_REQUEST_FAILED", "embedding request failed")
	}
	var parsed openAIEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	out := make([][]float32, 0, len(parsed.Data))
	for _, item := range parsed.Data {
		vec := make([]float32, 0, len(item.Embedding))
		for _, v := range item.Embedding {
			vec = append(vec, float32(v))
		}
		out = append(out, vec)
	}
	if len(out) > 0 && p.dim <= 0 {
		p.dim = len(out[0])
	}
	return out, nil
}

func (p *openAIEmbeddingProvider) Model() string {
	if p.model == "" {
		return defaultEmbeddingModel
	}
	return p.model
}

func (p *openAIEmbeddingProvider) Dim() int {
	if p.dim <= 0 {
		return defaultEmbeddingDim
	}
	return p.dim
}

func (uc *KnowledgeUsecase) embedChunks(ctx context.Context, chunks []DocChunk) ([]EmbeddedChunk, error) {
	if len(chunks) == 0 {
		return nil, nil
	}
	if uc.embedder == nil {
		uc.embedder = fakeEmbeddingProvider{model: defaultEmbeddingModel, dim: defaultEmbeddingDim}
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

func deterministicEmbedding(s string, dim int) []float32 {
	if dim <= 0 {
		return nil
	}
	sum := sha256.Sum256([]byte(s))
	seed := binary.LittleEndian.Uint64(sum[:8])
	rng := rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
	out := make([]float32, dim)
	for i := 0; i < dim; i++ {
		// [-1, 1)
		out[i] = float32(rng.Float64()*2.0 - 1.0)
	}
	return out
}
