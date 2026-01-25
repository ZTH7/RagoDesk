package service

import (
	analyticsservice "github.com/ZTH7/RagoDesk/apps/server/internal/analytics/service"
	apimgmtservice "github.com/ZTH7/RagoDesk/apps/server/internal/apimgmt/service"
	authnservice "github.com/ZTH7/RagoDesk/apps/server/internal/authn/service"
	botservice "github.com/ZTH7/RagoDesk/apps/server/internal/bot/service"
	conversationservice "github.com/ZTH7/RagoDesk/apps/server/internal/conversation/service"
	iamservice "github.com/ZTH7/RagoDesk/apps/server/internal/iam/service"
	knowledgeservice "github.com/ZTH7/RagoDesk/apps/server/internal/knowledge/service"
	ragservice "github.com/ZTH7/RagoDesk/apps/server/internal/rag/service"

	"github.com/google/wire"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(
	analyticsservice.ProviderSet,
	apimgmtservice.ProviderSet,
	authnservice.ProviderSet,
	botservice.ProviderSet,
	conversationservice.ProviderSet,
	iamservice.ProviderSet,
	knowledgeservice.ProviderSet,
	ragservice.ProviderSet,
)
