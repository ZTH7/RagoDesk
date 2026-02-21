package server

import (
	analyticsv1 "github.com/ZTH7/RagoDesk/apps/server/api/analytics/v1"
	apimgmtv1 "github.com/ZTH7/RagoDesk/apps/server/api/apimgmt/v1"
	authv1 "github.com/ZTH7/RagoDesk/apps/server/api/auth/v1"
	botv1 "github.com/ZTH7/RagoDesk/apps/server/api/bot/v1"
	conversationv1 "github.com/ZTH7/RagoDesk/apps/server/api/conversation/v1"
	iamv1 "github.com/ZTH7/RagoDesk/apps/server/api/iam/v1"
	knowledgev1 "github.com/ZTH7/RagoDesk/apps/server/api/knowledge/v1"
	ragv1 "github.com/ZTH7/RagoDesk/apps/server/api/rag/v1"
	analyticsservice "github.com/ZTH7/RagoDesk/apps/server/internal/analytics/service"
	apimgmtservice "github.com/ZTH7/RagoDesk/apps/server/internal/apimgmt/service"
	authservice "github.com/ZTH7/RagoDesk/apps/server/internal/auth/service"
	botservice "github.com/ZTH7/RagoDesk/apps/server/internal/bot/service"
	"github.com/ZTH7/RagoDesk/apps/server/internal/conf"
	conversationservice "github.com/ZTH7/RagoDesk/apps/server/internal/conversation/service"
	iamservice "github.com/ZTH7/RagoDesk/apps/server/internal/iam/service"
	knowledgeservice "github.com/ZTH7/RagoDesk/apps/server/internal/knowledge/service"
	"github.com/ZTH7/RagoDesk/apps/server/internal/middleware"
	ragservice "github.com/ZTH7/RagoDesk/apps/server/internal/rag/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, logger log.Logger, iamSvc *iamservice.IAMService, knowledgeSvc *knowledgeservice.KnowledgeService, ragSvc *ragservice.RAGService, conversationSvc *conversationservice.ConversationService, apimgmtSvc *apimgmtservice.APIMgmtService, analyticsSvc *analyticsservice.AnalyticsService, botSvc *botservice.BotService, consoleAuthSvc *authservice.ConsoleAuthService, platformAuthSvc *authservice.PlatformAuthService) *http.Server {
	var opts = []http.ServerOption{
		http.Filter(middleware.CORSFilter()),
		http.Middleware(
			recovery.Recovery(),
			middleware.ErrorMiddleware(),
			middleware.TracingMiddleware(),
			middleware.LoggingMiddleware(),
			middleware.AuthMiddleware(c.Auth),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	iamv1.RegisterPlatformIAMHTTPServer(srv, iamSvc)
	iamv1.RegisterConsoleIAMHTTPServer(srv, iamSvc)
	knowledgev1.RegisterConsoleKnowledgeHTTPServer(srv, knowledgeSvc)
	botv1.RegisterConsoleBotHTTPServer(srv, botSvc)
	authv1.RegisterConsoleAuthHTTPServer(srv, consoleAuthSvc)
	authv1.RegisterPlatformAuthHTTPServer(srv, platformAuthSvc)
	apimgmtv1.RegisterConsoleAPIMgmtHTTPServer(srv, apimgmtSvc)
	analyticsv1.RegisterConsoleAnalyticsHTTPServer(srv, analyticsSvc)
	ragv1.RegisterRAGHTTPServer(srv, ragSvc)
	conversationv1.RegisterConversationHTTPServer(srv, conversationSvc)
	conversationv1.RegisterConsoleConversationHTTPServer(srv, conversationSvc)
	srv.Route("/console/v1").POST("/documents/upload_file", knowledgeSvc.UploadDocumentFile)
	return srv
}
