package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
)

type openAIProvider struct {
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

func init() {
	RegisterProvider("openai", newOpenAIProvider)
	RegisterProvider("http", newOpenAIProvider)
}

func newOpenAIProvider(cfg Config) Provider {
	endpoint := strings.TrimSpace(cfg.Endpoint)
	if endpoint == "" {
		return newFakeProvider(cfg)
	}
	timeout := time.Duration(cfg.TimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	return &openAIProvider{
		endpoint: strings.TrimRight(endpoint, "/"),
		apiKey:   strings.TrimSpace(cfg.APIKey),
		model:    cfg.Model,
		dim:      cfg.Dim,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (p *openAIProvider) Embed(ctx context.Context, inputs []string) ([][]float32, error) {
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

func (p *openAIProvider) Model() string {
	return p.model
}

func (p *openAIProvider) Dim() int {
	return p.dim
}
