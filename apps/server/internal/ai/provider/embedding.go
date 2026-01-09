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
