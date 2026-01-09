package data

import "context"

// VectorStore abstracts the vector database operations.
type VectorStore interface {
	EnsureCollection(ctx context.Context, collection string, dim int) error
	UpsertPoints(ctx context.Context, collection string, points []VectorPoint) error
	DeletePoints(ctx context.Context, collection string, filter VectorFilter) error
}

// VectorPoint represents a vector entry.
type VectorPoint struct {
	ID      string         `json:"id"`
	Vector  []float32      `json:"vector"`
	Payload map[string]any `json:"payload,omitempty"`
}

// VectorFilter describes a boolean filter.
type VectorFilter struct {
	Must []VectorCondition `json:"must,omitempty"`
}

// VectorCondition matches on payload value equality.
type VectorCondition struct {
	Key   string            `json:"key"`
	Match *VectorMatchValue `json:"match,omitempty"`
}

// VectorMatchValue wraps an exact match value.
type VectorMatchValue struct {
	Value any `json:"value"`
}

func VectorMatchCondition(key string, value any) VectorCondition {
	return VectorCondition{
		Key:   key,
		Match: &VectorMatchValue{Value: value},
	}
}
