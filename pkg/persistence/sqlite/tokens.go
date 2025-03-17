package sqlite

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

func (d *Database) GetTokens(ctx context.Context) (*oauth2.Token, error) {
	accessToken, err := d.GetSetting(ctx, accessTokenSetting)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get access token")
	}

	refreshToken, err := d.GetSetting(ctx, refreshTokenSetting)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get refresh token")
	}

	tokenType, err := d.GetSetting(ctx, tokenTypeSetting)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get token type setting")
	}

	expiryString, err := d.GetSetting(ctx, expirySetting)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get expiry setting")
	}

	expiry, err := time.Parse(expiryTimeFormat, expiryString)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse expiry string")
	}

	return &oauth2.Token{
		AccessToken:  accessToken,
		TokenType:    tokenType,
		RefreshToken: refreshToken,
		Expiry:       expiry,
	}, nil
}

func (d *Database) SetAccessToken(ctx context.Context, accessToken string) error {
	return d.SetSetting(ctx, accessTokenSetting, accessToken)
}

func (d *Database) SetExpiry(ctx context.Context, expiry time.Time) error {
	expiryString := expiry.Format(expiryTimeFormat)

	return d.SetSetting(ctx, expirySetting, expiryString)
}

var NoRefreshTokenErr = errors.New("no refresh token present")

func (d *Database) SetTokens(ctx context.Context, token *oauth2.Token) error {
	if token.RefreshToken == "" {
		return NoRefreshTokenErr
	}

	var err error
	if err = d.SetSetting(ctx, refreshTokenSetting, token.RefreshToken); err != nil {
		return errors.Wrap(err, "failed to store refresh token")
	}

	if err = d.SetAccessToken(ctx, token.AccessToken); err != nil {
		return errors.Wrap(err, "failed to store access token")
	}

	if err = d.SetExpiry(ctx, token.Expiry); err != nil {
		return errors.Wrap(err, "failed to store expiry")
	}

	if err = d.SetSetting(ctx, tokenTypeSetting, token.TokenType); err != nil {
		return errors.Wrap(err, "failed to store token type")
	}

	return nil
}

func (d *Database) RemoveTokens(ctx context.Context) error {
	var err error

	for _, setting := range []SettingType{
		tokenTypeSetting,
		refreshTokenSetting,
		accessTokenSetting,
		refreshTokenSetting,
		expirySetting,
	} {
		if err = d.removeSetting(ctx, setting); err != nil {
			return errors.Wrapf(err, "failed to remove %s", setting)
		}
	}

	return nil
}

func (d *Database) UpdateTokens(ctx context.Context, tokens *oauth2.Token) error {
	var err error

	// store new tokens
	if err = d.SetAccessToken(ctx, tokens.AccessToken); err != nil {
		return errors.Wrap(err, "failed to store new access token")
	}

	// store new expiry
	if err = d.SetExpiry(ctx, tokens.Expiry); err != nil {
		return errors.Wrap(err, "failed to store expiry date")
	}

	// store new token type
	if err = d.SetSetting(ctx, tokenTypeSetting, tokens.TokenType); err != nil {
		return errors.Wrap(err, "failed to store token type")
	}

	return nil
}

const expiryTimeFormat = time.RFC3339

const (
	accessTokenSetting  SettingType = "accessToken"
	refreshTokenSetting SettingType = "refreshToken"
	tokenTypeSetting    SettingType = "tokenType"
	expirySetting       SettingType = "expiry"
)
