package service

import (
	analyticsservice "github.com/ZTH7/RAGDesk/apps/server/internal/analytics/service"
	apimgmtservice "github.com/ZTH7/RAGDesk/apps/server/internal/apimgmt/service"
	conversationservice "github.com/ZTH7/RAGDesk/apps/server/internal/conversation/service"
	handoffservice "github.com/ZTH7/RAGDesk/apps/server/internal/handoff/service"
	iamservice "github.com/ZTH7/RAGDesk/apps/server/internal/iam/service"
	knowledgeservice "github.com/ZTH7/RAGDesk/apps/server/internal/knowledge/service"
	platformservice "github.com/ZTH7/RAGDesk/apps/server/internal/platform/service"
	ragservice "github.com/ZTH7/RAGDesk/apps/server/internal/rag/service"

	"github.com/google/wire"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(
	NewGreeterService,
	analyticsservice.ProviderSet,
	apimgmtservice.ProviderSet,
	conversationservice.ProviderSet,
	handoffservice.ProviderSet,
	iamservice.ProviderSet,
	knowledgeservice.ProviderSet,
	platformservice.ProviderSet,
	ragservice.ProviderSet,
)
