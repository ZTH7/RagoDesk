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
	RegisterLLMProvider("openai", newOpenAIChatProvider)
	RegisterLLMProvider("http", newOpenAIChatProvider)
}

func newOpenAIProvider(cfg Config) Provider {
	endpoint := strings.TrimSpace(cfg.Endpoint)
	if endpoint == "" {
		return newTemplateProvider(cfg)
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

type openAIChatProvider struct {
	endpoint string
	apiKey   string
	model    string
	client   *http.Client
}

type openAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatRequest struct {
	Model       string              `json:"model"`
	Messages    []openAIChatMessage `json:"messages"`
	Temperature float32             `json:"temperature,omitempty"`
	MaxTokens   int                 `json:"max_tokens,omitempty"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message openAIChatMessage `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func newOpenAIChatProvider(cfg LLMConfig) LLMProvider {
	endpoint := strings.TrimSpace(cfg.Endpoint)
	if endpoint == "" {
		return newTemplateLLMProvider(cfg)
	}
	timeout := time.Duration(cfg.TimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	return &openAIChatProvider{
		endpoint: strings.TrimRight(endpoint, "/"),
		apiKey:   strings.TrimSpace(cfg.APIKey),
		model:    strings.TrimSpace(cfg.Model),
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (p *openAIChatProvider) Generate(ctx context.Context, req LLMRequest) (LLMResponse, error) {
	if p == nil || p.endpoint == "" {
		return LLMResponse{}, errors.InternalServer("LLM_ENDPOINT_MISSING", "llm endpoint missing")
	}
	if p.client == nil {
		p.client = &http.Client{Timeout: 20 * time.Second}
	}
	messages := make([]openAIChatMessage, 0, 2)
	if system := strings.TrimSpace(req.System); system != "" {
		messages = append(messages, openAIChatMessage{Role: "system", Content: system})
	}
	messages = append(messages, openAIChatMessage{Role: "user", Content: req.Prompt})

	payload := openAIChatRequest{
		Model:       p.Model(),
		Messages:    messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return LLMResponse{}, err
	}
	url := p.endpoint
	if !strings.HasSuffix(url, "/chat/completions") {
		url = url + "/chat/completions"
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return LLMResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	}
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return LLMResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return LLMResponse{}, errors.InternalServer("LLM_REQUEST_FAILED", "llm request failed")
	}
	var parsed openAIChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return LLMResponse{}, err
	}
	if len(parsed.Choices) == 0 {
		return LLMResponse{}, errors.InternalServer("LLM_EMPTY_RESPONSE", "llm response empty")
	}
	text := strings.TrimSpace(parsed.Choices[0].Message.Content)
	return LLMResponse{
		Text: text,
		Usage: LLMUsage{
			PromptTokens:     parsed.Usage.PromptTokens,
			CompletionTokens: parsed.Usage.CompletionTokens,
			TotalTokens:      parsed.Usage.TotalTokens,
		},
	}, nil
}

func (p *openAIChatProvider) Model() string {
	if p.model == "" {
		return "gpt-4o-mini"
	}
	return p.model
}
