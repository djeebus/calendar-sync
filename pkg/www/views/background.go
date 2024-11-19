package views

import (
	"context"

	"github.com/labstack/echo/v4"

	"calendar-sync/pkg/logs"
)

func (v Views) background(c echo.Context, fn func(ctx context.Context)) {
	ctx := c.Request().Context()
	logger := logs.GetLogger(ctx)

	ctx = context.Background()
	backgroundLogger := logger.With().Ctx(ctx).Logger()
	ctx = logs.SetLogger(ctx, backgroundLogger)

	go fn(ctx)
}
