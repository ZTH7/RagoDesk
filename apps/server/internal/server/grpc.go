package server

import (
	apimgmtv1 "github.com/ZTH7/RAGDesk/apps/server/api/apimgmt/v1"
	conversationv1 "github.com/ZTH7/RAGDesk/apps/server/api/conversation/v1"
	iamv1 "github.com/ZTH7/RAGDesk/apps/server/api/iam/v1"
	knowledgev1 "github.com/ZTH7/RAGDesk/apps/server/api/knowledge/v1"
	ragv1 "github.com/ZTH7/RAGDesk/apps/server/api/rag/v1"
	"github.com/ZTH7/RAGDesk/apps/server/internal/conf"
	apimgmtservice "github.com/ZTH7/RAGDesk/apps/server/internal/apimgmt/service"
	conversationservice "github.com/ZTH7/RAGDesk/apps/server/internal/conversation/service"
	iamservice "github.com/ZTH7/RAGDesk/apps/server/internal/iam/service"
	knowledgeservice "github.com/ZTH7/RAGDesk/apps/server/internal/knowledge/service"
	ragservice "github.com/ZTH7/RAGDesk/apps/server/internal/rag/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Server, logger log.Logger, iamSvc *iamservice.IAMService, knowledgeSvc *knowledgeservice.KnowledgeService, ragSvc *ragservice.RAGService, conversationSvc *conversationservice.ConversationService, apimgmtSvc *apimgmtservice.APIMgmtService) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			ErrorMiddleware(),
			TracingMiddleware(),
			LoggingMiddleware(),
			AuthMiddleware(c.Auth),
		),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)
	iamv1.RegisterPlatformIAMServer(srv, iamSvc)
	iamv1.RegisterConsoleIAMServer(srv, iamSvc)
	knowledgev1.RegisterConsoleKnowledgeServer(srv, knowledgeSvc)
	apimgmtv1.RegisterConsoleAPIMgmtServer(srv, apimgmtSvc)
	ragv1.RegisterRAGServer(srv, ragSvc)
	conversationv1.RegisterConversationServer(srv, conversationSvc)
	conversationv1.RegisterConsoleConversationServer(srv, conversationSvc)
	return srv
}
