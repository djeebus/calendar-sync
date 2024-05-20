package views

import (
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"net/url"
)

func (v Views) CreateCopyConfig(c echo.Context, values url.Values) error {
	ctx := c.Request().Context()

	source := values.Get("source")
	if source == "" {
		return errors.New("missing required field 'source'")
	}
	destination := values.Get("destination")
	if destination == "" {
		return errors.New("missing required field 'destination'")
	}

	if err := v.ctr.Database.CreateCopyConfig(ctx, source, destination); err != nil {
		return errors.Wrap(err, "failed to create invite config")
	}

	return c.Redirect(302, "/")
}

func (v Views) CreateInviteConfig(c echo.Context, values url.Values) error {
	ctx := c.Request().Context()

	calendarID := values.Get("calendar")
	if calendarID == "" {
		return errors.New("missing required field 'calendar'")
	}
	email := values.Get("email")
	if email == "" {
		return errors.New("missing required field 'email'")
	}

	if err := v.ctr.Database.CreateInviteConfig(ctx, calendarID, email); err != nil {
		return errors.Wrap(err, "failed to create invite config")
	}

	return c.Redirect(302, "/")
}
