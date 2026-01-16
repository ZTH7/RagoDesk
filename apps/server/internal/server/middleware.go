package server

import (
	"context"
	"strings"
	"time"

	"github.com/ZTH7/RAGDesk/apps/server/internal/auth"
	"github.com/ZTH7/RAGDesk/apps/server/internal/conf"
	"github.com/ZTH7/RAGDesk/apps/server/internal/tenant"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

// AuthMiddleware is a placeholder for auth (API Key / JWT) enforcement.
func AuthMiddleware(cfg *conf.Server_Auth) middleware.Middleware {
	return func(next middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return next(ctx, req)
			}
			if !isAdminOperation(tr.Operation()) {
				return next(ctx, req)
			}
			if cfg == nil || cfg.JwtSecret == "" {
				return nil, errors.Unauthorized("ADMIN_UNAUTHORIZED", "jwt config missing")
			}
			header := strings.TrimSpace(tr.RequestHeader().Get("Authorization"))
			token := header
			if strings.HasPrefix(strings.ToLower(header), "bearer ") {
				token = strings.TrimSpace(header[len("bearer "):])
			}
			if token == "" {
				return nil, errors.Unauthorized("ADMIN_UNAUTHORIZED", "missing authorization")
			}
			claims, err := auth.ParseHS256(token, cfg.JwtSecret, cfg.Issuer, cfg.Audience, time.Now())
			if err != nil {
				return nil, errors.Unauthorized("ADMIN_UNAUTHORIZED", err.Error())
			}
			ctx = auth.WithClaims(ctx, claims)
			if claims.TenantID != "" {
				ctx = tenant.WithTenantID(ctx, claims.TenantID)
			}
			return next(ctx, req)
		}
	}
}

// LoggingMiddleware is a placeholder for structured request logging.
func LoggingMiddleware() middleware.Middleware {
	return func(next middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// TODO: add structured logging with request/response metadata.
			return next(ctx, req)
		}
	}
}

// TracingMiddleware is a placeholder for distributed tracing.
func TracingMiddleware() middleware.Middleware {
	return func(next middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// TODO: propagate trace/span context.
			return next(ctx, req)
		}
	}
}

// ErrorMiddleware is a placeholder for standardized error mapping.
func ErrorMiddleware() middleware.Middleware {
	return func(next middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// TODO: map domain errors to API error codes.
			return next(ctx, req)
		}
	}
}

func isAdminOperation(operation string) bool {
	return strings.Contains(operation, "PlatformIAM") ||
		strings.Contains(operation, "ConsoleIAM") ||
		strings.Contains(operation, "ConsoleKnowledge") ||
		strings.Contains(operation, "ConsoleConversation")
}
