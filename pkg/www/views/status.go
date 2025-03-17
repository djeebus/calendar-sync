package views

import (
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (v Views) Status(e echo.Context) error {
	ctx := e.Request().Context()
	tokens, err := v.ctr.Database.GetTokens(ctx)
	if err != nil {
		return errors.Wrap(err, "")
	}

	return e.JSON(200, map[string]any{
		"expiration_seconds": tokens.ExpiresIn,
		"token_type":         tokens.TokenType,
		"expiry":             tokens.Expiry.String(),
		"is_valid":           tokens.Valid(),
	})
}
