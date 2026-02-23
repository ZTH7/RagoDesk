package biz

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ZTH7/RagoDesk/apps/server/internal/conf"
)

const (
	defaultChunkSizeTokens    = 800
	defaultChunkOverlapTokens = 100
	defaultEmbeddingModel     = "text-embedding-3-small"
	defaultEmbeddingDim       = 0
	defaultEmbeddingProvider  = "openai"
	defaultOutboundProxy      = "http://127.0.0.1:10808"
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
	asyncEnabled       bool
	indexConfigHash    string
	proxy              string
}

func loadIngestionOptions(cfg *conf.Data) ingestionOptions {
	opts := ingestionOptions{
		chunkSizeTokens:    defaultChunkSizeTokens,
		chunkOverlapTokens: defaultChunkOverlapTokens,
		embeddingModel:     defaultEmbeddingModel,
		embeddingDim:       defaultEmbeddingDim,
		embeddingProvider:  defaultEmbeddingProvider,
		embeddingEndpoint:  "",
		embeddingAPIKey:    "",
		embeddingTimeoutMs: 15000,
		embeddingBatchSize: 64,
		asyncEnabled:       false,
		proxy:              defaultOutboundProxy,
	}
	dimSet := false
	if cfg != nil && cfg.Knowledge != nil {
		if chunking := cfg.Knowledge.Chunking; chunking != nil {
			if chunking.MaxTokens > 0 {
				opts.chunkSizeTokens = int(chunking.MaxTokens)
			}
			if chunking.OverlapTokens >= 0 {
				opts.chunkOverlapTokens = int(chunking.OverlapTokens)
			}
		}
		if embedding := cfg.Knowledge.Embedding; embedding != nil {
			if strings.TrimSpace(embedding.Provider) != "" {
				opts.embeddingProvider = embedding.Provider
			}
			if strings.TrimSpace(embedding.Endpoint) != "" {
				opts.embeddingEndpoint = embedding.Endpoint
			}
			if strings.TrimSpace(embedding.ApiKey) != "" {
				opts.embeddingAPIKey = embedding.ApiKey
			}
			if strings.TrimSpace(embedding.Model) != "" {
				opts.embeddingModel = embedding.Model
			}
			if embedding.Dim > 0 {
				opts.embeddingDim = int(embedding.Dim)
				dimSet = true
			}
			if embedding.TimeoutMs > 0 {
				opts.embeddingTimeoutMs = int(embedding.TimeoutMs)
			}
			if embedding.BatchSize > 0 {
				opts.embeddingBatchSize = int(embedding.BatchSize)
			}
		}
		if ingestion := cfg.Knowledge.Ingestion; ingestion != nil {
			opts.asyncEnabled = ingestion.AsyncEnabled
		}
	}
	if cfg != nil {
		if rag := cfg.Rag; rag != nil && rag.Llm != nil {
			if strings.TrimSpace(opts.embeddingEndpoint) == "" && strings.TrimSpace(rag.Llm.Endpoint) != "" {
				opts.embeddingEndpoint = rag.Llm.Endpoint
			}
			if strings.TrimSpace(opts.embeddingAPIKey) == "" && strings.TrimSpace(rag.Llm.ApiKey) != "" {
				opts.embeddingAPIKey = rag.Llm.ApiKey
			}
		}
	}
	if cfg != nil {
		if proxy := strings.TrimSpace(cfg.Proxy); proxy != "" {
			opts.proxy = proxy
		}
	}

	opts.chunkSizeTokens = envInt("RAGODESK_CHUNK_SIZE_TOKENS", opts.chunkSizeTokens)
	opts.chunkOverlapTokens = envInt("RAGODESK_CHUNK_OVERLAP_TOKENS", opts.chunkOverlapTokens)
	opts.embeddingModel = envString("RAGODESK_EMBEDDING_MODEL", opts.embeddingModel)
	opts.embeddingProvider = envString("RAGODESK_EMBEDDING_PROVIDER", opts.embeddingProvider)
	opts.embeddingEndpoint = envString("RAGODESK_EMBEDDING_ENDPOINT", opts.embeddingEndpoint)
	opts.embeddingAPIKey = envString("RAGODESK_EMBEDDING_API_KEY", opts.embeddingAPIKey)
	opts.embeddingTimeoutMs = envInt("RAGODESK_EMBEDDING_TIMEOUT_MS", opts.embeddingTimeoutMs)
	opts.embeddingBatchSize = envInt("RAGODESK_EMBEDDING_BATCH_SIZE", opts.embeddingBatchSize)

	if raw := strings.TrimSpace(os.Getenv("RAGODESK_EMBEDDING_DIM")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			opts.embeddingDim = parsed
			dimSet = true
		}
	}
	if !dimSet && (strings.EqualFold(opts.embeddingProvider, "openai") || strings.EqualFold(opts.embeddingProvider, "http")) {
		opts.embeddingDim = 0
	}
	if raw := strings.TrimSpace(os.Getenv("RAGODESK_INGESTION_ASYNC")); raw != "" {
		switch strings.ToLower(raw) {
		case "1", "true", "yes", "y":
			opts.asyncEnabled = true
		default:
			opts.asyncEnabled = false
		}
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
	opts.indexConfigHash = buildIndexConfigHash(opts)
	return opts
}

func buildIndexConfigHash(opts ingestionOptions) string {
	payload := fmt.Sprintf(
		"chunk=%d|overlap=%d|provider=%s|model=%s|dim=%d|endpoint=%s",
		opts.chunkSizeTokens,
		opts.chunkOverlapTokens,
		strings.ToLower(strings.TrimSpace(opts.embeddingProvider)),
		strings.TrimSpace(opts.embeddingModel),
		opts.embeddingDim,
		strings.TrimSpace(opts.embeddingEndpoint),
	)
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
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
