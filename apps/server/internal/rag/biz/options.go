package biz

import (
	"os"
	"strconv"
	"strings"

	"github.com/ZTH7/RAGDesk/apps/server/internal/ai/provider"
	"github.com/ZTH7/RAGDesk/apps/server/internal/conf"
)

const (
	defaultRagTopK           = 5
	defaultRagScoreThreshold = 0.2
	defaultRagTimeoutMs      = 20000
	defaultRetrieveTimeoutMs = 8000
	defaultRetrieveConcurrency = 8
	defaultLLMTimeoutMs      = 15000
	defaultLLMProvider       = "fake"
	defaultLLMModel          = "fake-llm-v1"
	defaultLLMTemperature    = 0.2
	defaultLLMMaxTokens      = 512
	defaultRerankWeight      = 0.3
	defaultAPIKeyHeader      = "X-API-Key"
	defaultEmbeddingModel    = "fake-embedding-v1"
	defaultEmbeddingDim      = 384
	defaultEmbeddingProvider = "fake"
	defaultSystemPrompt      = "You are a helpful assistant. Answer using the provided context. If the context is insufficient, say you don't know."
	defaultRefusalMessage    = "I don't have enough information to answer that based on the provided knowledge."
)

type ragOptions struct {
	topK              int
	scoreThreshold    float32
	ragTimeoutMs      int
	retrieveTimeoutMs int
	retrieveConcurrency int
	llmTimeoutMs      int
	rerankWeight      float32
	llmProvider       string
	llmEndpoint       string
	llmAPIKey         string
	llmModel          string
	llmTemperature    float32
	llmMaxTokens      int
	systemPrompt      string
	refusalMessage    string
	apiKeyRequired    bool
	apiKeyHeader      string
	embeddingConfig   provider.Config
}

func loadRAGOptions(cfg *conf.Data) ragOptions {
	opts := ragOptions{
		topK:              defaultRagTopK,
		scoreThreshold:    float32(defaultRagScoreThreshold),
		ragTimeoutMs:      defaultRagTimeoutMs,
		retrieveTimeoutMs: defaultRetrieveTimeoutMs,
		retrieveConcurrency: defaultRetrieveConcurrency,
		llmTimeoutMs:      defaultLLMTimeoutMs,
		rerankWeight:      float32(defaultRerankWeight),
		llmProvider:       defaultLLMProvider,
		llmEndpoint:       "",
		llmAPIKey:         "",
		llmModel:          defaultLLMModel,
		llmTemperature:    float32(defaultLLMTemperature),
		llmMaxTokens:      defaultLLMMaxTokens,
		systemPrompt:      defaultSystemPrompt,
		refusalMessage:    defaultRefusalMessage,
		apiKeyRequired:    false,
		apiKeyHeader:      defaultAPIKeyHeader,
		embeddingConfig: provider.Config{
			Provider:  defaultEmbeddingProvider,
			Endpoint:  "",
			APIKey:    "",
			Model:     defaultEmbeddingModel,
			Dim:       defaultEmbeddingDim,
			TimeoutMs: defaultLLMTimeoutMs,
		},
	}
	dimSet := false

	if cfg != nil {
		if rag := cfg.Rag; rag != nil {
			if rag.TimeoutMs > 0 {
				opts.ragTimeoutMs = int(rag.TimeoutMs)
			}
			if retrieval := rag.Retrieval; retrieval != nil {
				if retrieval.TopK > 0 {
					opts.topK = int(retrieval.TopK)
				}
				if retrieval.Threshold > 0 {
					opts.scoreThreshold = retrieval.Threshold
				}
				if retrieval.TimeoutMs > 0 {
					opts.retrieveTimeoutMs = int(retrieval.TimeoutMs)
				}
				if retrieval.MaxConcurrency > 0 {
					opts.retrieveConcurrency = int(retrieval.MaxConcurrency)
				}
				if retrieval.RerankWeight > 0 {
					opts.rerankWeight = retrieval.RerankWeight
				}
			}
			if llm := rag.Llm; llm != nil {
				if strings.TrimSpace(llm.Provider) != "" {
					opts.llmProvider = llm.Provider
				}
				if strings.TrimSpace(llm.Endpoint) != "" {
					opts.llmEndpoint = llm.Endpoint
				}
				if strings.TrimSpace(llm.ApiKey) != "" {
					opts.llmAPIKey = llm.ApiKey
				}
				if strings.TrimSpace(llm.Model) != "" {
					opts.llmModel = llm.Model
				}
				if llm.TimeoutMs > 0 {
					opts.llmTimeoutMs = int(llm.TimeoutMs)
				}
				if llm.Temperature > 0 {
					opts.llmTemperature = llm.Temperature
				}
				if llm.MaxTokens > 0 {
					opts.llmMaxTokens = int(llm.MaxTokens)
				}
				if strings.TrimSpace(llm.SystemPrompt) != "" {
					opts.systemPrompt = llm.SystemPrompt
				}
				if strings.TrimSpace(llm.RefusalMessage) != "" {
					opts.refusalMessage = llm.RefusalMessage
				}
			}
			opts.apiKeyRequired = rag.ApiKeyRequired
			if strings.TrimSpace(rag.ApiKeyHeader) != "" {
				opts.apiKeyHeader = rag.ApiKeyHeader
			}
		}
		if knowledge := cfg.Knowledge; knowledge != nil {
			if embedding := knowledge.Embedding; embedding != nil {
				if strings.TrimSpace(embedding.Provider) != "" {
					opts.embeddingConfig.Provider = embedding.Provider
				}
				if strings.TrimSpace(embedding.Endpoint) != "" {
					opts.embeddingConfig.Endpoint = embedding.Endpoint
				}
				if strings.TrimSpace(embedding.ApiKey) != "" {
					opts.embeddingConfig.APIKey = embedding.ApiKey
				}
				if strings.TrimSpace(embedding.Model) != "" {
					opts.embeddingConfig.Model = embedding.Model
				}
				if embedding.Dim > 0 {
					opts.embeddingConfig.Dim = int(embedding.Dim)
					dimSet = true
				}
				if embedding.TimeoutMs > 0 {
					opts.embeddingConfig.TimeoutMs = int(embedding.TimeoutMs)
				}
			}
		}
	}

	opts.topK = envInt("RAGDESK_RAG_TOP_K", opts.topK)
	opts.scoreThreshold = envFloat32("RAGDESK_RAG_SCORE_THRESHOLD", opts.scoreThreshold)
	opts.ragTimeoutMs = envInt("RAGDESK_RAG_TIMEOUT_MS", opts.ragTimeoutMs)
	opts.retrieveTimeoutMs = envInt("RAGDESK_RETRIEVE_TIMEOUT_MS", opts.retrieveTimeoutMs)
	opts.retrieveConcurrency = envInt("RAGDESK_RETRIEVE_MAX_CONCURRENCY", opts.retrieveConcurrency)
	opts.llmProvider = envString("RAGDESK_LLM_PROVIDER", opts.llmProvider)
	opts.llmEndpoint = envString("RAGDESK_LLM_ENDPOINT", opts.llmEndpoint)
	opts.llmAPIKey = envString("RAGDESK_LLM_API_KEY", opts.llmAPIKey)
	opts.llmModel = envString("RAGDESK_LLM_MODEL", opts.llmModel)
	opts.llmTimeoutMs = envInt("RAGDESK_LLM_TIMEOUT_MS", opts.llmTimeoutMs)
	opts.llmTemperature = envFloat32("RAGDESK_LLM_TEMPERATURE", opts.llmTemperature)
	opts.llmMaxTokens = envInt("RAGDESK_LLM_MAX_TOKENS", opts.llmMaxTokens)
	opts.systemPrompt = envString("RAGDESK_RAG_SYSTEM_PROMPT", opts.systemPrompt)
	opts.refusalMessage = envString("RAGDESK_RAG_REFUSAL_MESSAGE", opts.refusalMessage)
	opts.rerankWeight = envFloat32("RAGDESK_RAG_RERANK_WEIGHT", opts.rerankWeight)
	opts.apiKeyRequired = envBool("RAGDESK_API_KEY_REQUIRED", opts.apiKeyRequired)
	opts.apiKeyHeader = envString("RAGDESK_API_KEY_HEADER", opts.apiKeyHeader)

	opts.embeddingConfig.Provider = envString("RAGDESK_EMBEDDING_PROVIDER", opts.embeddingConfig.Provider)
	opts.embeddingConfig.Endpoint = envString("RAGDESK_EMBEDDING_ENDPOINT", opts.embeddingConfig.Endpoint)
	opts.embeddingConfig.APIKey = envString("RAGDESK_EMBEDDING_API_KEY", opts.embeddingConfig.APIKey)
	opts.embeddingConfig.Model = envString("RAGDESK_EMBEDDING_MODEL", opts.embeddingConfig.Model)
	opts.embeddingConfig.TimeoutMs = envInt("RAGDESK_EMBEDDING_TIMEOUT_MS", opts.embeddingConfig.TimeoutMs)
	if raw := strings.TrimSpace(os.Getenv("RAGDESK_EMBEDDING_DIM")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			opts.embeddingConfig.Dim = parsed
			dimSet = true
		}
	}
	if opts.topK <= 0 {
		opts.topK = defaultRagTopK
	}
	if opts.llmTimeoutMs <= 0 {
		opts.llmTimeoutMs = defaultLLMTimeoutMs
	}
	if opts.ragTimeoutMs <= 0 {
		opts.ragTimeoutMs = defaultRagTimeoutMs
	}
	if opts.retrieveTimeoutMs <= 0 {
		opts.retrieveTimeoutMs = defaultRetrieveTimeoutMs
	}
	if opts.retrieveConcurrency <= 0 {
		opts.retrieveConcurrency = defaultRetrieveConcurrency
	}
	if opts.retrieveConcurrency > 64 {
		opts.retrieveConcurrency = 64
	}
	if opts.llmMaxTokens <= 0 {
		opts.llmMaxTokens = defaultLLMMaxTokens
	}
	if opts.rerankWeight < 0 {
		opts.rerankWeight = 0
	}
	if opts.rerankWeight > 1 {
		opts.rerankWeight = 1
	}
	if opts.embeddingConfig.Dim < 0 {
		opts.embeddingConfig.Dim = defaultEmbeddingDim
	}
	if opts.embeddingConfig.TimeoutMs <= 0 {
		opts.embeddingConfig.TimeoutMs = defaultLLMTimeoutMs
	}
	if !dimSet && (strings.EqualFold(opts.embeddingConfig.Provider, "openai") || strings.EqualFold(opts.embeddingConfig.Provider, "http")) {
		opts.embeddingConfig.Dim = 0
	}
	return opts
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

func envFloat32(key string, fallback float32) float32 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseFloat(value, 32)
	if err != nil {
		return fallback
	}
	return float32(parsed)
}

func envBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	switch strings.ToLower(value) {
	case "1", "true", "yes", "y":
		return true
	case "0", "false", "no", "n":
		return false
	default:
		return fallback
	}
}
