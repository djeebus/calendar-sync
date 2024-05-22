package views

import (
	"calendar-sync/pkg/clients"
	"context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"net/http"
	"slices"
	"time"
)

const authCookieName = ".auth"

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

	if !v.isValidUser(c.Request().Context(), tokens) {
		return errors.Wrap(err, "user is invalid")
	}

	if err := v.ctr.Database.SetTokens(ctx, tokens); err != nil {
		return errors.Wrap(err, "failed to store tokens")
	}

	v.setAuthCookie(c.Response())

	return c.Redirect(302, "/")
}

func (v Views) Logout(c echo.Context) error {
	ctx := c.Request().Context()
	if err := v.ctr.Database.RemoveTokens(ctx); err != nil {
		return errors.Wrap(err, "failed to log out")
	}

	return c.Redirect(302, "/")
}

func (v Views) RequireClientToken(noAuthPages ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			request := c.Request()
			cookie, err := request.Cookie(authCookieName)
			if cookie != nil && !v.isAuthCookieValid(cookie) {
				v.forgetAuthCookie(c.Response())
				cookie = nil
				err = http.ErrNoCookie
			}

			if errors.Is(err, http.ErrNoCookie) {
				// the only acceptable time to have no cookie is if you're trying to login
				if !slices.Contains(noAuthPages, request.URL.Path) {
					return c.Redirect(302, "/auth/begin")
				}
			}

			return next(c)
		}
	}
}

func (v Views) isValidUser(ctx context.Context, token *oauth2.Token) bool {
	client, err := clients.GetClient(ctx, v.ctr.OAuth2Config, token)
	if err != nil {
		log.Err(err).Msg("failed to get client")
		return false
	}

	var found bool
	if err = client.CalendarList.List().Pages(ctx, func(settings *calendar.CalendarList) error {
		for _, c := range settings.Items {
			if c.AccessRole != "owner" {
				continue
			}
			if c.Id != v.ctr.Config.OwnerEmailAddress {
				continue
			}

			found = true
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("failed to read settings")
	}
	return found
}

func (v Views) setAuthCookie(response *echo.Response) {
	expiration := time.Now().Add(v.ctr.Config.JwtDuration)

	alg := jwt.GetSigningMethod(v.ctr.Config.JwtAlgorithm)
	signer := jwt.NewWithClaims(alg, jwt.MapClaims{
		"iss": v.ctr.Config.JwtIssuer,
		"exp": expiration.Unix(),
	})
	token, err := signer.SignedString([]byte(v.ctr.Config.JwtSecretKey))
	if err != nil {
		log.Error().Err(err).Msg("failed to create jwt")
		return
	}

	cookie := http.Cookie{
		Name:    authCookieName,
		Value:   token,
		Path:    "/",
		Expires: expiration,
	}

	http.SetCookie(response.Writer, &cookie)
}

func (v Views) isAuthCookieValid(cookie *http.Cookie) bool {
	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		return []byte(v.ctr.Config.JwtSecretKey), nil
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to parse jwt")
		return false
	}

	validator := jwt.NewValidator()
	if err = validator.Validate(token.Claims); err != nil {
		log.Error().Err(err).Msg("claims are not valid")
		return false
	}

	return true
}

func (v Views) forgetAuthCookie(response *echo.Response) {
	cookie := http.Cookie{
		Name:    authCookieName,
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
	}
	http.SetCookie(response.Writer, &cookie)
}

func (v Views) WipeTokenIfInvalid(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := next(c)
		var oauth2err *oauth2.RetrieveError
		if errors.As(err, &oauth2err) && oauth2err.ErrorCode == "invalid_grant" {
			if err2 := v.ctr.Database.RemoveTokens(c.Request().Context()); err2 != nil {
				c.Logger().Warn("failed to remove invalid tokens", err2)
			}
			return c.Redirect(302, "/logout")
		}
		return err
	}
}
