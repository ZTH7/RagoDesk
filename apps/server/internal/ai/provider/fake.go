package provider

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"math/rand/v2"
)

type fakeProvider struct {
	model string
	dim   int
}

func init() {
	RegisterProvider("fake", newFakeProvider)
}

func newFakeProvider(cfg Config) Provider {
	return fakeProvider{
		model: cfg.Model,
		dim:   cfg.Dim,
	}
}

func (p fakeProvider) Embed(ctx context.Context, inputs []string) ([][]float32, error) {
	out := make([][]float32, 0, len(inputs))
	for _, text := range inputs {
		out = append(out, deterministicEmbedding(text, p.dim))
	}
	return out, nil
}

func (p fakeProvider) Model() string {
	return p.model
}

func (p fakeProvider) Dim() int {
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
