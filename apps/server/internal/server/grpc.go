package server

import (
	iamv1 "github.com/ZTH7/RAGDesk/apps/server/api/iam/v1"
	knowledgev1 "github.com/ZTH7/RAGDesk/apps/server/api/knowledge/v1"
	"github.com/ZTH7/RAGDesk/apps/server/internal/conf"
	iamservice "github.com/ZTH7/RAGDesk/apps/server/internal/iam/service"
	knowledgeservice "github.com/ZTH7/RAGDesk/apps/server/internal/knowledge/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Server, logger log.Logger, iamSvc *iamservice.IAMService, knowledgeSvc *knowledgeservice.KnowledgeService) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			ErrorMiddleware(),
			TracingMiddleware(),
			LoggingMiddleware(),
			AuthMiddleware(c.Auth),
			TenantContextMiddleware(),
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
	iamv1.RegisterTenantIAMServer(srv, iamSvc)
	knowledgev1.RegisterKnowledgeAdminServer(srv, knowledgeSvc)
	return srv
}
