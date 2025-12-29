package biz

import (
	iambiz "github.com/ZTH7/RAGDesk/apps/server/internal/iam/biz"
	knowledgebiz "github.com/ZTH7/RAGDesk/apps/server/internal/knowledge/biz"
	ragbiz "github.com/ZTH7/RAGDesk/apps/server/internal/rag/biz"
	conversationbiz "github.com/ZTH7/RAGDesk/apps/server/internal/conversation/biz"
	handoffbiz "github.com/ZTH7/RAGDesk/apps/server/internal/handoff/biz"
	analyticsbiz "github.com/ZTH7/RAGDesk/apps/server/internal/analytics/biz"
	apimgmtbiz "github.com/ZTH7/RAGDesk/apps/server/internal/apimgmt/biz"
	platformbiz "github.com/ZTH7/RAGDesk/apps/server/internal/platform/biz"

	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewGreeterUsecase,
	iambiz.ProviderSet,
	knowledgebiz.ProviderSet,
	ragbiz.ProviderSet,
	conversationbiz.ProviderSet,
	handoffbiz.ProviderSet,
	analyticsbiz.ProviderSet,
	apimgmtbiz.ProviderSet,
	platformbiz.ProviderSet,
)
