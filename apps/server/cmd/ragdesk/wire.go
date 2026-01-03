//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	analyticsdata "github.com/ZTH7/RAGDesk/apps/server/internal/analytics/data"
	apimgmtdata "github.com/ZTH7/RAGDesk/apps/server/internal/apimgmt/data"
	"github.com/ZTH7/RAGDesk/apps/server/internal/biz"
	"github.com/ZTH7/RAGDesk/apps/server/internal/conf"
	conversationdata "github.com/ZTH7/RAGDesk/apps/server/internal/conversation/data"
	"github.com/ZTH7/RAGDesk/apps/server/internal/data"
	iamdata "github.com/ZTH7/RAGDesk/apps/server/internal/iam/data"
	knowledgedata "github.com/ZTH7/RAGDesk/apps/server/internal/knowledge/data"
	platformdata "github.com/ZTH7/RAGDesk/apps/server/internal/platform/data"
	ragdata "github.com/ZTH7/RAGDesk/apps/server/internal/rag/data"
	"github.com/ZTH7/RAGDesk/apps/server/internal/server"
	"github.com/ZTH7/RAGDesk/apps/server/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		data.NewData,
		analyticsdata.ProviderSet,
		apimgmtdata.ProviderSet,
		conversationdata.ProviderSet,
		iamdata.ProviderSet,
		knowledgedata.ProviderSet,
		platformdata.ProviderSet,
		ragdata.ProviderSet,
		biz.ProviderSet,
		service.ProviderSet,
		newApp,
	))
}
