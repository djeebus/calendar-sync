package tracing

import (
	"context"
	"github.com/google/uuid"
)

type ctxKey struct{}

var key = ctxKey{}

func AddCorrelationID(ctx context.Context) context.Context {
	correlationID := uuid.NewString()
	correlationID = correlationID[:6]

	return setCorrelationID(ctx, correlationID)
}

func GetCorrelationID(ctx context.Context) (string, bool) {
	return getCorrelationID(ctx)
}

func getCorrelationID(ctx interface{ Value(key any) any }) (string, bool) {
	result := ctx.Value(key)
	if result == nil {
		return "", false
	}

	id, ok := result.(string)
	if !ok {
		return "", false
	}

	return id, true
}

func setCorrelationID(ctx context.Context, correlationID string) context.Context {
	ctx = context.WithValue(ctx, key, correlationID)
	return ctx
}
