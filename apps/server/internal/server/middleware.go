package server

import (
	"context"

	"github.com/ZTH7/RAGDesk/apps/server/internal/tenant"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

// AuthMiddleware is a placeholder for auth (API Key / JWT) enforcement.
func AuthMiddleware() middleware.Middleware {
	return func(next middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// TODO: inject auth checks and tenant context.
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

// TenantContextMiddleware extracts tenant_id from transport metadata.
func TenantContextMiddleware() middleware.Middleware {
	return func(next middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			if _, ok := tenant.TenantID(ctx); ok {
				return next(ctx, req)
			}
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return next(ctx, req)
			}
			tenantID := tr.RequestHeader().Get("X-Tenant-ID")
			if tenantID == "" {
				return next(ctx, req)
			}
			return next(tenant.WithTenantID(ctx, tenantID), req)
		}
	}
}
