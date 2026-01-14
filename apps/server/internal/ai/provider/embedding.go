package provider

import (
	"context"
	"strings"
)

// Provider embeds text into vector representations.
type Provider interface {
	Embed(ctx context.Context, inputs []string) ([][]float32, error)
	Model() string
	Dim() int
}

// Config configures an embedding provider.
type Config struct {
	Provider  string
	Endpoint  string
	APIKey    string
	Model     string
	Dim       int
	TimeoutMs int
}

// Factory builds a provider from config.
type Factory func(cfg Config) Provider

var registry = map[string]Factory{}

// RegisterProvider registers an embedding provider factory.
func RegisterProvider(name string, factory Factory) {
	key := strings.ToLower(strings.TrimSpace(name))
	if key == "" || factory == nil {
		return
	}
	registry[key] = factory
}

// NewProvider returns a provider based on config. Falls back to fake.
func NewProvider(cfg Config) Provider {
	name := strings.ToLower(strings.TrimSpace(cfg.Provider))
	if name == "" {
		name = "fake"
	}
	if factory, ok := registry[name]; ok {
		return factory(cfg)
	}
	return newFakeProvider(cfg)
}

// LLMProvider generates text responses from prompts.
type LLMProvider interface {
	Generate(ctx context.Context, req LLMRequest) (LLMResponse, error)
	Model() string
}

// LLMRequest describes a generation input.
type LLMRequest struct {
	System      string
	Prompt      string
	Temperature float32
	MaxTokens   int
}

// LLMResponse describes a generation output.
type LLMResponse struct {
	Text  string
	Usage LLMUsage
}

// LLMUsage captures token usage, if available.
type LLMUsage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// LLMConfig configures an LLM provider.
type LLMConfig struct {
	Provider  string
	Endpoint  string
	APIKey    string
	Model     string
	TimeoutMs int
}

// LLMFactory builds an LLM provider from config.
type LLMFactory func(cfg LLMConfig) LLMProvider

var llmRegistry = map[string]LLMFactory{}

// RegisterLLMProvider registers an LLM provider factory.
func RegisterLLMProvider(name string, factory LLMFactory) {
	key := strings.ToLower(strings.TrimSpace(name))
	if key == "" || factory == nil {
		return
	}
	llmRegistry[key] = factory
}

// NewLLMProvider returns an LLM provider based on config. Falls back to fake.
func NewLLMProvider(cfg LLMConfig) LLMProvider {
	name := strings.ToLower(strings.TrimSpace(cfg.Provider))
	if name == "" {
		name = "fake"
	}
	if factory, ok := llmRegistry[name]; ok {
		return factory(cfg)
	}
	return newFakeLLMProvider(cfg)
}
