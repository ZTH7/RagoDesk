package biz

import (
	"context"
	"strings"
	"time"
)

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

func alignQueryWeights(queries []string, weights []float32) []float32 {
	out := make([]float32, len(queries))
	for i := range out {
		out[i] = 1
		if i < len(weights) && weights[i] > 0 {
			out[i] = weights[i]
		}
	}
	return out
}

func dedupeQueries(queries []string) []string {
	seen := make(map[string]struct{}, len(queries))
	out := make([]string, 0, len(queries))
	for _, q := range queries {
		q = strings.TrimSpace(q)
		if q == "" {
			continue
		}
		key := strings.ToLower(q)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, q)
	}
	if len(out) == 0 && len(queries) > 0 {
		return []string{queries[0]}
	}
	return out
}
