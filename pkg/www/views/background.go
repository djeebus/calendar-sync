package views

import (
	"calendar-sync/pkg/logs"
	"context"
	"github.com/labstack/echo/v4"
)

func (v Views) background(c echo.Context, fn func(ctx context.Context)) {
	ctx := c.Request().Context()
	logger := logs.GetLogger(ctx)

	ctx = context.Background()
	backgroundLogger := logger.With().Ctx(ctx).Logger()
	ctx = logs.SetLogger(ctx, backgroundLogger)

	go fn(ctx)
}
