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

type qdrantClient struct {
	endpoint   string
	apiKey     string
	httpClient *http.Client
}

func newQdrantClient(endpoint string, apiKey string) *qdrantClient {
	endpoint = strings.TrimSpace(endpoint)
	endpoint = strings.TrimRight(endpoint, "/")
	return &qdrantClient{
		endpoint: endpoint,
		apiKey:   strings.TrimSpace(apiKey),
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type qdrantPoint struct {
	ID      string         `json:"id"`
	Vector  []float32      `json:"vector"`
	Payload map[string]any `json:"payload,omitempty"`
}

type qdrantUpsertRequest struct {
	Points []qdrantPoint `json:"points"`
}

func (c *qdrantClient) EnsureCollection(ctx context.Context, collection string, dim int) error {
	collection = strings.TrimSpace(collection)
	if collection == "" {
		return kerrors.InternalServer("QDRANT_COLLECTION_MISSING", "qdrant collection missing")
	}
	if dim <= 0 {
		return kerrors.InternalServer("QDRANT_VECTOR_DIM_INVALID", "qdrant vector dim invalid")
	}

	getURL := c.endpoint + "/collections/" + url.PathEscape(collection)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, getURL, nil)
	if err != nil {
		return err
	}
	c.applyAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		// create below
	default:
		body := readBodyLimit(resp.Body, 8<<10)
		return kerrors.InternalServer("QDRANT_COLLECTION_CHECK_FAILED", fmt.Sprintf("qdrant check failed: %s", body))
	}

	createURL := c.endpoint + "/collections/" + url.PathEscape(collection)
	payload := map[string]any{
		"vectors": map[string]any{
			"size":     dim,
			"distance": "Cosine",
		},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req2, err := http.NewRequestWithContext(ctx, http.MethodPut, createURL, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	c.applyAuth(req2)
	req2.Header.Set("Content-Type", "application/json")

	resp2, err := c.httpClient.Do(req2)
	if err != nil {
		return err
	}
	defer resp2.Body.Close()
	if resp2.StatusCode < 200 || resp2.StatusCode >= 300 {
		body := readBodyLimit(resp2.Body, 16<<10)
		return kerrors.InternalServer("QDRANT_COLLECTION_CREATE_FAILED", fmt.Sprintf("qdrant create failed: %s", body))
	}
	return nil
}

func (c *qdrantClient) UpsertPoints(ctx context.Context, collection string, points []qdrantPoint) error {
	collection = strings.TrimSpace(collection)
	if collection == "" {
		return kerrors.InternalServer("QDRANT_COLLECTION_MISSING", "qdrant collection missing")
	}
	if len(points) == 0 {
		return nil
	}
	reqBody := qdrantUpsertRequest{Points: points}
	raw, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}
	upsertURL := c.endpoint + "/collections/" + url.PathEscape(collection) + "/points?wait=true"
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, upsertURL, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	c.applyAuth(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body := readBodyLimit(resp.Body, 16<<10)
		return kerrors.InternalServer("QDRANT_UPSERT_FAILED", fmt.Sprintf("qdrant upsert failed: %s", body))
	}
	return nil
}

func (c *qdrantClient) applyAuth(req *http.Request) {
	if c.apiKey == "" {
		return
	}
	req.Header.Set("api-key", c.apiKey)
}

func readBodyLimit(r io.Reader, n int64) string {
	if n <= 0 {
		n = 8 << 10
	}
	b, _ := io.ReadAll(io.LimitReader(r, n))
	return string(b)
}
