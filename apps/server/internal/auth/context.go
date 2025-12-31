package auth

import "context"

type ctxKey struct{}

// WithClaims stores JWT claims in context.
func WithClaims(ctx context.Context, claims *Claims) context.Context {
	if claims == nil {
		return ctx
	}
	return context.WithValue(ctx, ctxKey{}, claims)
}

// ClaimsFromContext returns JWT claims from context.
func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	value, ok := ctx.Value(ctxKey{}).(*Claims)
	if !ok || value == nil {
		return nil, false
	}
	return value, true
}
