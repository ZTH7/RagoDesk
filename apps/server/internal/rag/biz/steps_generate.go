package biz

import (
	"context"
	"strings"
	"time"

	"github.com/ZTH7/RAGDesk/apps/server/internal/ai/provider"
	"github.com/go-kratos/kratos/v2/errors"
	"go.opentelemetry.io/otel/attribute"
)

func (uc *RAGUsecase) promptContext(ctx context.Context, rc *ragContext) (*ragContext, error) {
	if rc == nil || rc.shouldRefuse {
		return rc, nil
	}
	_, span := uc.startSpan(ctx, "rag.prompt")
	defer span.End()
	rc.prompt = buildPrompt(rc.req.Message, rc.ranked, rc.chunks)
	return rc, nil
}

func (uc *RAGUsecase) llmContext(ctx context.Context, rc *ragContext) (*ragContext, error) {
	if rc == nil || rc.shouldRefuse {
		return rc, nil
	}
	ctx, span := uc.startSpan(ctx, "rag.llm", attribute.String("rag.llm_model", uc.llm.Model()))
	defer span.End()
	llmCtx, cancel := withTimeout(ctx, uc.opts.llmTimeoutMs)
	defer cancel()
	start := time.Now()
	resp, err := uc.llm.Generate(llmCtx, provider.LLMRequest{
		System:      uc.opts.systemPrompt,
		Prompt:      rc.prompt,
		Temperature: uc.opts.llmTemperature,
		MaxTokens:   uc.opts.llmMaxTokens,
	})
	uc.logStep("llm", start, err)
	if err != nil {
		uc.recordSpanError(span, err)
		return rc, err
	}
	rc.reply = strings.TrimSpace(resp.Text)
	return rc, nil
}

func (uc *RAGUsecase) buildResponse(_ context.Context, rc *ragContext) (MessageResponse, error) {
	if rc == nil {
		return MessageResponse{}, errors.InternalServer("RAG_CONTEXT_MISSING", "rag context missing")
	}
	reply := strings.TrimSpace(rc.reply)
	if reply == "" {
		reply = uc.opts.refusalMessage
	}
	return MessageResponse{
		Reply:      reply,
		Confidence: rc.confidence,
		References: buildReferences(rc.ranked, rc.chunks),
	}, nil
}
