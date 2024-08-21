package logs

import (
	"context"

	"github.com/rs/zerolog"
)

func SetLogger(ctx context.Context, logger zerolog.Logger) context.Context {
	return logger.WithContext(ctx)
}

func GetLogger(ctx context.Context) *zerolog.Logger {
	return zerolog.Ctx(ctx)
}
