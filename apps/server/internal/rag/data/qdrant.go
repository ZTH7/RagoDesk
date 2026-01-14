package data

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	kerrors "github.com/go-kratos/kratos/v2/errors"
)

type qdrantSearchClient struct {
	endpoint   string
	apiKey     string
	httpClient *http.Client
}

type qdrantSearchRequest struct {
	Vector         []float32     `json:"vector"`
	Limit          int           `json:"limit"`
	WithPayload    bool          `json:"with_payload"`
	Filter         *qdrantFilter `json:"filter,omitempty"`
	ScoreThreshold float32       `json:"score_threshold,omitempty"`
}

type qdrantFilter struct {
	Must []qdrantCondition `json:"must,omitempty"`
}

type qdrantCondition struct {
	Key   string          `json:"key"`
	Match *qdrantMatchAny `json:"match,omitempty"`
}

type qdrantMatchAny struct {
	Value any `json:"value"`
}

type qdrantSearchResponse struct {
	Result []qdrantPoint `json:"result"`
}

type qdrantPoint struct {
	ID      any            `json:"id"`
	Score   float32        `json:"score"`
	Payload map[string]any `json:"payload"`
}

type qdrantSearchResult struct {
	ChunkID           string
	DocumentID        string
	DocumentVersionID string
	KBID              string
	Score             float32
}

func newQdrantSearchClient(endpoint string, apiKey string, timeoutMs int) *qdrantSearchClient {
	endpoint = strings.TrimSpace(endpoint)
	endpoint = strings.TrimRight(endpoint, "/")
	if timeoutMs <= 0 {
		timeoutMs = 10000
	}
	return &qdrantSearchClient{
		endpoint: endpoint,
		apiKey:   strings.TrimSpace(apiKey),
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutMs) * time.Millisecond,
		},
	}
}

func (c *qdrantSearchClient) Search(ctx context.Context, collection string, vector []float32, topK int, tenantID string, kbID string, threshold float32) ([]qdrantSearchResult, error) {
	if c == nil || c.endpoint == "" {
		return nil, kerrors.InternalServer("QDRANT_ENDPOINT_MISSING", "qdrant endpoint missing")
	}
	collection = strings.TrimSpace(collection)
	if collection == "" {
		return nil, kerrors.InternalServer("QDRANT_COLLECTION_MISSING", "qdrant collection missing")
	}
	if len(vector) == 0 {
		return nil, nil
	}
	if topK <= 0 {
		topK = 5
	}
	filter := &qdrantFilter{
		Must: []qdrantCondition{
			{Key: "tenant_id", Match: &qdrantMatchAny{Value: tenantID}},
			{Key: "kb_id", Match: &qdrantMatchAny{Value: kbID}},
		},
	}
	reqBody := qdrantSearchRequest{
		Vector:         vector,
		Limit:          topK,
		WithPayload:    true,
		Filter:         filter,
		ScoreThreshold: threshold,
	}
	raw, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	searchURL := c.endpoint + "/collections/" + url.PathEscape(collection) + "/points/search"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, searchURL, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	c.applyAuth(req)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body := readBodyLimit(resp.Body, 16<<10)
		return nil, kerrors.InternalServer("QDRANT_SEARCH_FAILED", fmt.Sprintf("qdrant search failed: %s", body))
	}
	var parsed qdrantSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	out := make([]qdrantSearchResult, 0, len(parsed.Result))
	for _, item := range parsed.Result {
		payload := item.Payload
		out = append(out, qdrantSearchResult{
			ChunkID:           payloadString(payload, "chunk_id", item.ID),
			DocumentID:        payloadString(payload, "document_id", ""),
			DocumentVersionID: payloadString(payload, "document_version_id", ""),
			KBID:              payloadString(payload, "kb_id", ""),
			Score:             item.Score,
		})
	}
	return out, nil
}

func (c *qdrantSearchClient) applyAuth(req *http.Request) {
	if c.apiKey == "" {
		return
	}
	req.Header.Set("api-key", c.apiKey)
}

func payloadString(payload map[string]any, key string, fallback any) string {
	if payload != nil {
		if raw, ok := payload[key]; ok {
			switch v := raw.(type) {
			case string:
				return v
			case fmt.Stringer:
				return v.String()
			case float64:
				if v == float64(int64(v)) {
					return fmt.Sprintf("%d", int64(v))
				}
				return fmt.Sprintf("%.4f", v)
			}
		}
	}
	if fallback == nil {
		return ""
	}
	switch v := fallback.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func readBodyLimit(r io.Reader, n int64) string {
	if n <= 0 {
		n = 8 << 10
	}
	b, _ := io.ReadAll(io.LimitReader(r, n))
	return string(b)
}
