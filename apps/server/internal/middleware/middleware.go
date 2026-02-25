package middleware

import (
	"context"
	"strings"
	"time"

	jwt "github.com/ZTH7/RagoDesk/apps/server/internal/kit/jwt"
	"github.com/ZTH7/RagoDesk/apps/server/internal/kit/tenant"
	"github.com/ZTH7/RagoDesk/apps/server/internal/conf"
	"github.com/go-kratos/kratos/v2/errors"
	kmmiddleware "github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

// AuthMiddleware enforces JWT auth for protected console/platform operations.
func AuthMiddleware(cfg *conf.Server_Auth) kmmiddleware.Middleware {
	return func(next kmmiddleware.Handler) kmmiddleware.Handler {
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
			claims, err := jwt.ParseHS256(token, cfg.JwtSecret, cfg.Issuer, cfg.Audience, time.Now())
			if err != nil {
				return nil, errors.Unauthorized("ADMIN_UNAUTHORIZED", err.Error())
			}
			ctx = jwt.WithClaims(ctx, claims)
			if claims.TenantID != "" {
				ctx = tenant.WithTenantID(ctx, claims.TenantID)
			}
			return next(ctx, req)
		}
	}
}

// LoggingMiddleware is a placeholder for structured request logging.
func LoggingMiddleware() kmmiddleware.Middleware {
	return func(next kmmiddleware.Handler) kmmiddleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// TODO: add structured logging with request/response metadata.
			return next(ctx, req)
		}
	}
}

// TracingMiddleware is a placeholder for distributed tracing.
func TracingMiddleware() kmmiddleware.Middleware {
	return func(next kmmiddleware.Handler) kmmiddleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// TODO: propagate trace/span context.
			return next(ctx, req)
		}
	}
}

// ErrorMiddleware is a placeholder for standardized error mapping.
func ErrorMiddleware() kmmiddleware.Middleware {
	return func(next kmmiddleware.Handler) kmmiddleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// TODO: map domain errors to API error codes.
			return next(ctx, req)
		}
	}
}

func isAdminOperation(operation string) bool {
	if strings.Contains(operation, "ConsoleAuth") || strings.Contains(operation, "PlatformAuth") {
		return false
	}
	return strings.Contains(operation, "PlatformIAM") ||
		strings.Contains(operation, "ConsoleIAM") ||
		strings.Contains(operation, "ConsoleKnowledge") ||
		strings.Contains(operation, "ConsoleBot") ||
		strings.Contains(operation, "ConsoleConversation") ||
		strings.Contains(operation, "ConsoleAPIMgmt") ||
		strings.Contains(operation, "ConsoleAnalytics")
}
