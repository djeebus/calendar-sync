package views

import (
	"context"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

func (v Views) BeginAuth(c echo.Context) error {
	ctx := c.Request().Context()

	state := uuid.New().String()
	if err := v.ctr.Database.SetState(ctx, state); err != nil {
		return err
	}

	url := v.ctr.OAuth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return c.Redirect(302, url)
}

func (v Views) EndAuth(c echo.Context) error {
	r := c.Request()
	ctx := context.Background() // this request is too important to cancel

	if err := r.ParseForm(); err != nil {
		return errors.Wrap(err, "failed to parse form")
	}
	untrusted := r.Form.Get("state")
	trusted, err := v.ctr.Database.GetState(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get stored state")
	}

	if untrusted != trusted {
		return errors.New("state does not match")
	}

	code := r.Form.Get("code")
	tokens, err := v.ctr.OAuth2Config.Exchange(ctx, code)
	if err != nil {
		return errors.Wrap(err, "failed to exchange code for token")
	}

	if err := v.ctr.Database.SetTokens(ctx, tokens); err != nil {
		return errors.Wrap(err, "failed to store tokens")
	}

	return c.Redirect(302, "/")
}

func (v Views) Logout(c echo.Context) error {
	ctx := c.Request().Context()
	if err := v.ctr.Database.RemoveTokens(ctx); err != nil {
		return errors.Wrap(err, "failed to log out")
	}

	return c.Redirect(302, "/")
}
