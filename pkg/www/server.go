package www

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"github.com/ziflex/lecho/v3"

	"calendar-sync/pkg/container"
	"calendar-sync/pkg/logs"
	"calendar-sync/pkg/tasks/workflows"
	"calendar-sync/pkg/tracing"
	"calendar-sync/pkg/www/templates"
	"calendar-sync/pkg/www/views"
)

func NewServer(ctr container.Container, workflows *workflows.Workflows) *echo.Echo {
	e := echo.New()
	e.Debug = true
	e.HideBanner = true
	e.Logger = lecho.From(ctr.Logger)
	e.Renderer = templates.New()

	v := views.New(ctr, workflows)

	e.Use(tracing.GenerateCorrelationID())
	e.Use(logs.CreateRequestLogger(ctr.Logger))
	e.Use(logs.LogRequest())
	e.Use(middleware.Recover())
	e.Use(v.RequireClientToken("/auth/begin", "/auth/end", "/hooks/calendar", "/-/status"))
	e.Use(v.WipeTokenIfInvalid)

	e.GET("/auth/begin", v.BeginAuth)
	e.GET("/auth/end", v.EndAuth)
	e.GET("/", v.Dashboard)
	e.GET("/-/status", v.Status)
	e.POST("/hooks/calendar", v.Webhook)
	e.POST("/", func(c echo.Context) error {
		vals, err := c.FormParams()
		if err != nil {
			return errors.Wrap(err, "failed to get form params")
		}

		switch vals.Get("cmd") {
		case "copy":
			return v.CreateCopyConfig(c, vals)
		case "invite":
			return v.CreateInviteConfig(c, vals)
		case "sync copy":
			return v.SyncCopy(c, vals)
		case "sync invite":
			return v.SyncInvite(c, vals)
		case "delete invite":
			return v.DeleteInviteConfig(c, vals)
		case "delete copy":
			return v.DeleteCopyConfig(c, vals)
		case "renew token":
			return v.RenewToken(c)
		default:
			return echo.ErrMethodNotAllowed
		}
	})

	return e
}
