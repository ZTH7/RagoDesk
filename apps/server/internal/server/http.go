package server

import (
	v1 "github.com/ZTH7/RAGDesk/apps/server/api/iam/v1"
	"github.com/ZTH7/RAGDesk/apps/server/internal/conf"
	iamservice "github.com/ZTH7/RAGDesk/apps/server/internal/iam/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, logger log.Logger, iamSvc *iamservice.IAMService) *http.Server {
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
	v1.RegisterAdminIAMHTTPServer(srv, iamSvc)
	return srv
}
