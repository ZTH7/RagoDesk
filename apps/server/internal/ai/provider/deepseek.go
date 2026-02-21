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

type deepSeekChatProvider struct {
	endpoint string
	apiKey   string
	model    string
	client   *http.Client
}

func init() {
	RegisterProvider("deepseek", newOpenAIProvider)
	RegisterLLMProvider("deepseek", newDeepSeekChatProvider)
}

func newDeepSeekChatProvider(cfg LLMConfig) LLMProvider {
	endpoint := strings.TrimSpace(cfg.Endpoint)
	if endpoint == "" {
		return newTemplateLLMProvider(cfg)
	}
	timeout := time.Duration(cfg.TimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	return &deepSeekChatProvider{
		endpoint: strings.TrimRight(endpoint, "/"),
		apiKey:   strings.TrimSpace(cfg.APIKey),
		model:    strings.TrimSpace(cfg.Model),
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (p *deepSeekChatProvider) Generate(ctx context.Context, req LLMRequest) (LLMResponse, error) {
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

func (p *deepSeekChatProvider) Model() string {
	if p.model == "" {
		return "deepseek-chat"
	}
	return p.model
}
