package logs

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

func CreateRequestLogger(logger zerolog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// get context
			r := c.Request()
			ctx := r.Context()

			// copy logger, attach context
			logger := logger.With().Ctx(ctx).Logger()

			// attach logger to context
			ctx = SetLogger(ctx, logger)

			// rebuild request
			r = r.WithContext(ctx)
			c.SetRequest(r)

			// handle function
			return next(c)
		}
	}
}

func LogRequest() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			r := c.Request()
			ctx := r.Context()
			logger := GetLogger(ctx)

			logger.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Msg("request_begin")
			start := time.Now()
			result := next(c)
			duration := time.Since(start)

			l := logger.Info().
				Int64("ms", duration.Milliseconds())

			if r.Response != nil {
				l = l.Int("status", r.Response.StatusCode)
			}

			l.Msg("request_end")

			return result
		}
	}
}
