package tracing

import (
	"github.com/labstack/echo/v4"
)

func GenerateCorrelationID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// get context
			r := c.Request()
			ctx := r.Context()

			// add correlation id to context
			ctx = AddCorrelationID(ctx)

			// rebuild echo context
			r = r.WithContext(ctx)
			c.SetRequest(r)

			// handle request
			return next(c)
		}
	}
}
