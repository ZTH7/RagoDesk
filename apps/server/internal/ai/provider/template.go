package provider

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand/v2"
	"strings"
)

type templateProvider struct {
	model string
	dim   int
}

func newTemplateProvider(cfg Config) Provider {
	return templateProvider{
		model: cfg.Model,
		dim:   cfg.Dim,
	}
}

func (p templateProvider) Embed(ctx context.Context, inputs []string) ([][]float32, error) {
	out := make([][]float32, 0, len(inputs))
	for _, text := range inputs {
		out = append(out, deterministicEmbedding(text, p.dim))
	}
	return out, nil
}

func (p templateProvider) Model() string {
	return p.model
}

func (p templateProvider) Dim() int {
	return p.dim
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

type templateLLMProvider struct {
	model string
}

func init() {
	RegisterProvider("template", newTemplateProvider)
	RegisterLLMProvider("template", newTemplateLLMProvider)
}

func newTemplateLLMProvider(cfg LLMConfig) LLMProvider {
	return templateLLMProvider{
		model: strings.TrimSpace(cfg.Model),
	}
}

func (p templateLLMProvider) Generate(ctx context.Context, req LLMRequest) (LLMResponse, error) {
	text := strings.TrimSpace(req.Prompt)
	if text == "" {
		text = "No input provided."
	}
	return LLMResponse{
		Text:  fmt.Sprintf("RAG response (template): %s", truncateText(text, 180)),
		Usage: LLMUsage{},
	}, nil
}

func (p templateLLMProvider) Model() string {
	if p.model == "" {
		return "template-llm-v1"
	}
	return p.model
}

func truncateText(text string, limit int) string {
	if limit <= 0 || len(text) <= limit {
		return text
	}
	return text[:limit] + "..."
}
