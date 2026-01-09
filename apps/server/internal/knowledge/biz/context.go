package biz

import (
	"context"

	"github.com/ZTH7/RAGDesk/apps/server/internal/tenant"
	"github.com/go-kratos/kratos/v2/errors"
)

func tenantIDFromContext(ctx context.Context) (string, error) {
	if value, ok := tenant.TenantID(ctx); ok {
		return value, nil
	}
	return "", errors.Forbidden("TENANT_MISSING", "tenant missing")
}

func withTenantID(ctx context.Context, tenantID string) context.Context {
	if tenantID == "" {
		return ctx
	}
	return tenant.WithTenantID(ctx, tenantID)
}
