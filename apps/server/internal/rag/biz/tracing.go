package biz

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func (uc *RAGUsecase) logStep(step string, start time.Time, err error) {
	if uc == nil || uc.log == nil {
		return
	}
	dur := time.Since(start).Milliseconds()
	if err != nil {
		uc.log.Warnf("rag step=%s dur_ms=%d err=%v", step, dur, err)
		return
	}
	uc.log.Infof("rag step=%s dur_ms=%d", step, dur)
}

func (uc *RAGUsecase) startSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	tracer := otel.Tracer("ragdesk/rag")
	ctx, span := tracer.Start(ctx, name)
	if len(attrs) > 0 {
		span.SetAttributes(attrs...)
	}
	return ctx, span
}

func (uc *RAGUsecase) recordSpanError(span trace.Span, err error) {
	if span == nil || err == nil {
		return
	}
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}
