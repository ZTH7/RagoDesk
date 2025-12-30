package tenant

import (
	"context"
	"errors"
)

type ctxKey struct{}

// ErrTenantMissing is returned when tenant_id is not present in context.
var ErrTenantMissing = errors.New("tenant id missing")

// WithTenantID attaches tenant_id to context.
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, ctxKey{}, tenantID)
}

// TenantID extracts tenant_id from context.
func TenantID(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(ctxKey{}).(string)
	if !ok || value == "" {
		return "", false
	}
	return value, true
}

// RequireTenantID returns tenant_id or error when missing.
func RequireTenantID(ctx context.Context) (string, error) {
	if value, ok := TenantID(ctx); ok {
		return value, nil
	}
	return "", ErrTenantMissing
}
