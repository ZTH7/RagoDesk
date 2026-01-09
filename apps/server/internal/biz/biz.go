package biz

import (
	iambiz "github.com/ZTH7/RAGDesk/apps/server/internal/iam/biz"
	knowledgebiz "github.com/ZTH7/RAGDesk/apps/server/internal/knowledge/biz"

	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	iambiz.ProviderSet,
	knowledgebiz.ProviderSet,
)
