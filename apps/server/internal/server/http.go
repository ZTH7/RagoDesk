package server

import (
	conversationv1 "github.com/ZTH7/RAGDesk/apps/server/api/conversation/v1"
	iamv1 "github.com/ZTH7/RAGDesk/apps/server/api/iam/v1"
	knowledgev1 "github.com/ZTH7/RAGDesk/apps/server/api/knowledge/v1"
	ragv1 "github.com/ZTH7/RAGDesk/apps/server/api/rag/v1"
	"github.com/ZTH7/RAGDesk/apps/server/internal/conf"
	conversationservice "github.com/ZTH7/RAGDesk/apps/server/internal/conversation/service"
	iamservice "github.com/ZTH7/RAGDesk/apps/server/internal/iam/service"
	knowledgeservice "github.com/ZTH7/RAGDesk/apps/server/internal/knowledge/service"
	ragservice "github.com/ZTH7/RAGDesk/apps/server/internal/rag/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, logger log.Logger, iamSvc *iamservice.IAMService, knowledgeSvc *knowledgeservice.KnowledgeService, ragSvc *ragservice.RAGService, conversationSvc *conversationservice.ConversationService) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			ErrorMiddleware(),
			TracingMiddleware(),
			LoggingMiddleware(),
			AuthMiddleware(c.Auth),
			TenantContextMiddleware(),
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
	ragv1.RegisterRAGHTTPServer(srv, ragSvc)
	conversationv1.RegisterConversationHTTPServer(srv, conversationSvc)
	conversationv1.RegisterConsoleConversationHTTPServer(srv, conversationSvc)
	return srv
}
