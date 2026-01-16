package biz

import (
	conversationbiz "github.com/ZTH7/RAGDesk/apps/server/internal/conversation/biz"
	iambiz "github.com/ZTH7/RAGDesk/apps/server/internal/iam/biz"
	knowledgebiz "github.com/ZTH7/RAGDesk/apps/server/internal/knowledge/biz"
	ragbiz "github.com/ZTH7/RAGDesk/apps/server/internal/rag/biz"

	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	conversationbiz.ProviderSet,
	iambiz.ProviderSet,
	knowledgebiz.ProviderSet,
	ragbiz.ProviderSet,
)
