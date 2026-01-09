package biz

import (
	"os"
	"strconv"
	"strings"
)

const (
	defaultChunkSizeTokens    = 400
	defaultChunkOverlapTokens = 50
	defaultEmbeddingModel     = "fake-embedding-v1"
	defaultEmbeddingDim       = 384
	defaultEmbeddingProvider  = "fake"
)

type ingestionOptions struct {
	chunkSizeTokens    int
	chunkOverlapTokens int
	embeddingModel     string
	embeddingDim       int
	embeddingProvider  string
	embeddingEndpoint  string
	embeddingAPIKey    string
	embeddingTimeoutMs int
	embeddingBatchSize int
}

func loadIngestionOptions() ingestionOptions {
	provider := envString("RAGDESK_EMBEDDING_PROVIDER", defaultEmbeddingProvider)
	embeddingDim := defaultEmbeddingDim
	if raw := strings.TrimSpace(os.Getenv("RAGDESK_EMBEDDING_DIM")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			embeddingDim = parsed
		}
	} else if strings.EqualFold(provider, "openai") || strings.EqualFold(provider, "http") {
		embeddingDim = 0
	}
	opts := ingestionOptions{
		chunkSizeTokens:    envInt("RAGDESK_CHUNK_SIZE_TOKENS", defaultChunkSizeTokens),
		chunkOverlapTokens: envInt("RAGDESK_CHUNK_OVERLAP_TOKENS", defaultChunkOverlapTokens),
		embeddingModel:     envString("RAGDESK_EMBEDDING_MODEL", defaultEmbeddingModel),
		embeddingDim:       embeddingDim,
		embeddingProvider:  provider,
		embeddingEndpoint:  envString("RAGDESK_EMBEDDING_ENDPOINT", ""),
		embeddingAPIKey:    envString("RAGDESK_EMBEDDING_API_KEY", ""),
		embeddingTimeoutMs: envInt("RAGDESK_EMBEDDING_TIMEOUT_MS", 15000),
		embeddingBatchSize: envInt("RAGDESK_EMBEDDING_BATCH_SIZE", 64),
	}
	if opts.chunkSizeTokens <= 0 {
		opts.chunkSizeTokens = defaultChunkSizeTokens
	}
	if opts.chunkOverlapTokens < 0 {
		opts.chunkOverlapTokens = defaultChunkOverlapTokens
	}
	if opts.embeddingDim < 0 {
		opts.embeddingDim = defaultEmbeddingDim
	}
	if opts.embeddingBatchSize <= 0 {
		opts.embeddingBatchSize = 64
	}
	return opts
}

func asyncEnabled(queue IngestionQueue) bool {
	if queue == nil {
		return false
	}
	value := strings.TrimSpace(os.Getenv("RAGDESK_INGESTION_ASYNC"))
	switch strings.ToLower(value) {
	case "1", "true", "yes", "y":
		return true
	default:
		return false
	}
}

func envString(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
