package biz

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ZTH7/RAGDesk/apps/server/internal/ai/provider"
	"go.opentelemetry.io/otel/attribute"
)

func (uc *RAGUsecase) rerankContext(ctx context.Context, rc *ragContext) (*ragContext, error) {
	if rc == nil || rc.shouldRefuse || len(rc.ranked) == 0 {
		return rc, nil
	}
	ctx, span := uc.startSpan(ctx, "rag.rerank", attribute.Float64("rag.rerank_weight", float64(uc.opts.rerankWeight)))
	defer span.End()
	start := time.Now()
	for i := range rc.ranked {
		chunk := rc.ranked[i]
		meta := rc.chunks[chunk.result.ChunkID]
		textScore := overlapScore(rc.normalized, meta.Content)
		sectionScore := overlapScore(rc.normalized, meta.Section)
		if sectionScore > 0 {
			textScore = maxFloat32(textScore, sectionScore*1.2)
		}
		chunk.textScore = textScore
		chunk.score = combineScores(chunk.vectorScore, textScore, uc.opts.rerankWeight)
		rc.ranked[i] = chunk
	}
	sort.SliceStable(rc.ranked, func(i, j int) bool {
		return rc.ranked[i].score > rc.ranked[j].score
	})
	uc.logStep("rerank", start, nil)
	return rc, nil
}

func (uc *RAGUsecase) assessContext(ctx context.Context, rc *ragContext) (*ragContext, error) {
	if rc == nil {
		return rc, nil
	}
	ctx, span := uc.startSpan(ctx, "rag.assess")
	defer span.End()
	conf := computeConfidence(rc.ranked, rc.topK)
	span.SetAttributes(attribute.Float64("rag.confidence", float64(conf)))
	if len(rc.ranked) > 1 && conf < rc.threshold {
		if err := uc.crossEncoderRerank(ctx, rc); err == nil {
			conf = computeConfidence(rc.ranked, rc.topK)
			span.SetAttributes(attribute.Float64("rag.confidence_after", float64(conf)))
		}
	}
	rc.confidence = conf
	if len(rc.ranked) == 0 || conf < rc.threshold {
		rc.shouldRefuse = true
		rc.reply = uc.opts.refusalMessage
	}
	return rc, nil
}

const (
	crossEncoderTopN = 8
	crossTimeoutMin  = 1200
	crossTimeoutMax  = 2500
)

func (uc *RAGUsecase) crossEncoderRerank(ctx context.Context, rc *ragContext) error {
	if uc == nil || rc == nil || len(rc.ranked) < 2 {
		return nil
	}
	if uc.llm == nil {
		return nil
	}
	model := strings.ToLower(strings.TrimSpace(uc.llm.Model()))
	if strings.Contains(model, "fake") {
		return nil
	}
	ctx, span := uc.startSpan(ctx, "rag.rerank_cross", attribute.String("rag.llm_model", uc.llm.Model()))
	defer span.End()

	n := crossEncoderTopN
	if len(rc.ranked) < n {
		n = len(rc.ranked)
	}
	prompt := buildCrossEncoderPrompt(rc.req.Message, rc.ranked[:n], rc.chunks)
	if prompt == "" {
		return nil
	}
	timeout := uc.opts.llmTimeoutMs / 5
	if timeout < crossTimeoutMin {
		timeout = crossTimeoutMin
	}
	if timeout > crossTimeoutMax {
		timeout = crossTimeoutMax
	}
	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining > 0 {
			remainingMs := int(remaining.Milliseconds())
			if remainingMs < timeout {
				timeout = remainingMs
			}
		}
	}
	if timeout <= 0 {
		return nil
	}
	llmCtx, cancel := withTimeout(ctx, timeout)
	defer cancel()
	start := time.Now()
	resp, err := uc.llm.Generate(llmCtx, provider.LLMRequest{
		System:      "You are a ranking model that orders passages by relevance to a question.",
		Prompt:      prompt,
		Temperature: 0,
		MaxTokens:   256,
	})
	uc.logStep("rerank_cross", start, err)
	if err != nil {
		uc.recordSpanError(span, err)
		return err
	}
	order := parseRerankOrder(resp.Text)
	if len(order) == 0 {
		return nil
	}
	applyRerankOrder(order, rc)
	return nil
}

func buildCrossEncoderPrompt(question string, ranked []scoredChunk, chunks map[string]ChunkMeta) string {
	var b strings.Builder
	b.WriteString("Rank the following passages by relevance to the question. ")
	b.WriteString("Return a JSON array of chunk_id in descending order.\n\n")
	b.WriteString("Question: ")
	b.WriteString(strings.TrimSpace(question))
	b.WriteString("\n\nPassages:\n")
	for idx, item := range ranked {
		meta := chunks[item.result.ChunkID]
		b.WriteString("[")
		b.WriteString(strconv.Itoa(idx + 1))
		b.WriteString("] chunk_id=")
		b.WriteString(item.result.ChunkID)
		if meta.Section != "" {
			b.WriteString(" section=")
			b.WriteString(meta.Section)
		}
		b.WriteString("\n")
		b.WriteString(truncateText(meta.Content, 400))
		b.WriteString("\n\n")
	}
	return b.String()
}

func parseRerankOrder(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	var parsed []string
	if tryParseJSONArray(text, &parsed) {
		return sanitizeStringList(parsed)
	}
	if start := strings.Index(text, "["); start >= 0 {
		if end := strings.LastIndex(text, "]"); end > start {
			if tryParseJSONArray(text[start:end+1], &parsed) {
				return sanitizeStringList(parsed)
			}
		}
	}
	parts := strings.FieldsFunc(text, func(r rune) bool {
		switch r {
		case '\n', '\r', ',', ';', ' ':
			return true
		default:
			return false
		}
	})
	return sanitizeStringList(parts)
}

func applyRerankOrder(order []string, rc *ragContext) {
	if len(order) == 0 || rc == nil {
		return
	}
	byID := make(map[string]scoredChunk, len(rc.ranked))
	for _, item := range rc.ranked {
		byID[item.result.ChunkID] = item
	}
	out := make([]scoredChunk, 0, len(rc.ranked))
	seen := make(map[string]struct{}, len(order))
	for _, id := range order {
		id = strings.TrimSpace(id)
		item, ok := byID[id]
		if !ok {
			continue
		}
		out = append(out, item)
		seen[id] = struct{}{}
	}
	for _, item := range rc.ranked {
		if _, ok := seen[item.result.ChunkID]; ok {
			continue
		}
		out = append(out, item)
	}
	rc.ranked = out
}

func tryParseJSONArray(text string, out *[]string) bool {
	if err := json.Unmarshal([]byte(text), out); err != nil {
		return false
	}
	return true
}

func sanitizeStringList(input []string) []string {
	out := make([]string, 0, len(input))
	for _, item := range input {
		q := strings.TrimSpace(item)
		if q == "" {
			continue
		}
		out = append(out, q)
	}
	return out
}
