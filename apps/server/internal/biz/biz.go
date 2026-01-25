package biz

import (
	analyticsbiz "github.com/ZTH7/RagoDesk/apps/server/internal/analytics/biz"
	apimgmtbiz "github.com/ZTH7/RagoDesk/apps/server/internal/apimgmt/biz"
	authnbiz "github.com/ZTH7/RagoDesk/apps/server/internal/authn/biz"
	botbiz "github.com/ZTH7/RagoDesk/apps/server/internal/bot/biz"
	conversationbiz "github.com/ZTH7/RagoDesk/apps/server/internal/conversation/biz"
	iambiz "github.com/ZTH7/RagoDesk/apps/server/internal/iam/biz"
	knowledgebiz "github.com/ZTH7/RagoDesk/apps/server/internal/knowledge/biz"
	ragbiz "github.com/ZTH7/RagoDesk/apps/server/internal/rag/biz"

	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	analyticsbiz.ProviderSet,
	apimgmtbiz.ProviderSet,
	authnbiz.ProviderSet,
	botbiz.ProviderSet,
	conversationbiz.ProviderSet,
	iambiz.ProviderSet,
	knowledgebiz.ProviderSet,
	ragbiz.ProviderSet,
)
