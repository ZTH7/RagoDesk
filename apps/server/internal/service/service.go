package service

import (
	iamservice "github.com/ZTH7/RAGDesk/apps/server/internal/iam/service"
	knowledgeservice "github.com/ZTH7/RAGDesk/apps/server/internal/knowledge/service"
	ragservice "github.com/ZTH7/RAGDesk/apps/server/internal/rag/service"

	"github.com/google/wire"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(
	iamservice.ProviderSet,
	knowledgeservice.ProviderSet,
	ragservice.ProviderSet,
)
