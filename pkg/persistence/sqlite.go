package persistence

import (
	"calendar-sync/pkg"
	"context"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"strconv"
	"time"
)

const dbVersionSetting settingType = "db_version"

func NewDatabase(ctx context.Context, cfg pkg.Config) (*Database, error) {
	conn, err := sql.Open(cfg.DatabaseDriver, cfg.DatabaseSource)
	if err != nil {
		return nil, errors.Wrap(err, "failed to epen db")
	}

	db := &Database{db: conn}
	var nextVersion = 0
	value, err := db.getSetting(ctx, dbVersionSetting)
	if err == nil && value != "" {
		nextVersion, _ = strconv.Atoi(value)
	}
	for {
		query, ok := migrations[nextVersion]
		if !ok {
			break
		}
		if _, err = conn.ExecContext(ctx, query); err != nil {
			return nil, errors.Wrapf(err, "failed to migrate to v%d", nextVersion)
		}
		if err = db.setSetting(ctx, dbVersionSetting, strconv.Itoa(nextVersion)); err != nil {
			return nil, errors.Wrap(err, "failed to persist db version")
		}
		nextVersion++
	}

	return db, nil
}

type Database struct {
	db *sql.DB
}

func (d *Database) Close() {
	d.db.Close()
}

func (d *Database) GetInviteConfig(ctx context.Context, id int64) (InviteConfig, error) {
	var config InviteConfig

	stmt, err := d.db.PrepareContext(ctx, `
SELECT id, calendarID, emailAddress
FROM invites
WHERE id = ?`)
	if err != nil {
		return InviteConfig{}, errors.Wrap(err, "failed to prepare statement")
	}
	defer stmt.Close()

	if err := stmt.QueryRowContext(ctx, id).Scan(&config.ID, &config.CalendarID, &config.EmailAddress); err != nil {
		return InviteConfig{}, errors.Wrap(err, "failed to parse row")
	}

	return config, nil
}

func (d *Database) GetInviteConfigs(ctx context.Context) ([]InviteConfig, error) {
	stmt, err := d.db.PrepareContext(ctx, `
SELECT id, calendarID, emailAddress 
FROM invites
`)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare statement")
	}
	defer stmt.Close()

	cur, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query context")
	}

	var configs []InviteConfig
	for cur.Next() {
		var config InviteConfig
		if err = cur.Scan(&config.ID, &config.CalendarID, &config.EmailAddress); err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}

		configs = append(configs, config)
	}

	return configs, nil
}

func (d *Database) GetCopyConfig(ctx context.Context, id int64) (CopyConfig, error) {
	var config CopyConfig

	stmt, err := d.db.PrepareContext(ctx, `
SELECT id, sourceID, destinationID
FROM copies
WHERE id = ?`)
	if err != nil {
		return CopyConfig{}, errors.Wrap(err, "failed to prepare statement")
	}
	defer stmt.Close()

	if err := stmt.QueryRowContext(ctx, id).Scan(&config.ID, &config.SourceID, &config.DestinationID); err != nil {
		return CopyConfig{}, errors.Wrap(err, "failed to parse row")
	}

	return config, nil
}

func (d *Database) GetCopyConfigs(ctx context.Context) ([]CopyConfig, error) {
	stmt, err := d.db.PrepareContext(ctx, `
SELECT id, sourceID, destinationID
FROM copies
`)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare statement")
	}
	defer stmt.Close()

	cur, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute statement")
	}

	var configs []CopyConfig
	for cur.Next() {
		var config CopyConfig
		if err = cur.Scan(&config.ID, &config.SourceID, &config.DestinationID); err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}
		configs = append(configs, config)
	}

	return configs, nil
}

func (d *Database) getSetting(ctx context.Context, name settingType) (string, error) {
	stmt, err := d.db.PrepareContext(ctx, `SELECT value FROM settings WHERE name = ?`)
	if err != nil {
		return "", errors.Wrap(err, "failed to prepare statement")
	}
	defer stmt.Close()

	var value string
	if err := stmt.QueryRowContext(ctx, name).Scan(&value); err != nil {
		return "", errors.Wrap(err, "failed to execute statement")
	}

	return value, nil
}

func (d *Database) setSetting(ctx context.Context, name settingType, value string) error {
	stmt, err := d.db.PrepareContext(ctx, `
INSERT INTO settings(name, value) VALUES(?, ?)
ON CONFLICT(name) DO 
UPDATE SET value=excluded.value
WHERE excluded.name = settings.name
`)
	if err != nil {
		return errors.Wrap(err, "failed to prepare statement")
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, string(name), value); err != nil {
		return err
	}

	return nil
}

func (d *Database) removeSetting(ctx context.Context, name settingType) error {
	stmt, err := d.db.PrepareContext(ctx, `
DELETE FROM settings
WHERE name = ?
`)
	if err != nil {
		return errors.Wrap(err, "failed to prepare statement")
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, string(name)); err != nil {
		return err
	}

	return nil
}

func (d *Database) CreateCopyConfig(ctx context.Context, sourceCalendarID, destinationCalendarID string) error {
	stmt, err := d.db.PrepareContext(ctx, `
INSERT INTO copies (sourceID, destinationID)
VALUES (?, ?)
`)
	if err != nil {
		return errors.Wrap(err, "failed to prepare statement")
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, sourceCalendarID, destinationCalendarID); err != nil {
		return errors.Wrap(err, "failed to execute statement")
	}

	return nil
}

func (d *Database) DeleteCopyConfig(ctx context.Context, copyID string) error {
	stmt, err := d.db.PrepareContext(ctx, `
DELETE FROM copies
WHERE id = ?`)
	if err != nil {
		return errors.Wrap(err, "failed to prepare statement")
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, copyID); err != nil {
		return errors.Wrap(err, "failed to execute statement")
	}

	return nil
}

func (d *Database) CreateWatchConfig(ctx context.Context, calendarID, watchID, token string) error {
	stmt, err := d.db.PrepareContext(ctx, `
INSERT INTO watches (calendarID, watchID, token)
VALUES (?, ?, ?)
`)
	if err != nil {
		return errors.Wrap(err, "failed to prepare statement")
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, calendarID, watchID, token); err != nil {
		return errors.Wrap(err, "failed to execute statement")
	}

	return nil
}

func (d *Database) GetWatchConfig(ctx context.Context, watchID string) (WatchConfig, error) {
	var config WatchConfig

	stmt, err := d.db.PrepareContext(ctx, `
SELECT id, calendarID, watchID, token
FROM watches
WHERE watchID = ?`)
	if err != nil {
		return config, errors.Wrap(err, "failed to prepare statement")
	}
	defer stmt.Close()

	if err := stmt.QueryRowContext(ctx, watchID).Scan(&config.ID, &config.CalendarID, &config.WatchID, &config.Token); err != nil {
		return config, errors.Wrap(err, "failed to parse row")
	}

	return config, nil
}

func (d *Database) GetWatchConfigs(ctx context.Context) ([]WatchConfig, error) {
	stmt, err := d.db.PrepareContext(ctx, `
SELECT id, calendarID, watchID, token
FROM watches
`)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare statement")
	}
	defer stmt.Close()

	cur, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute statement")
	}

	var watches []WatchConfig
	for cur.Next() {
		var watch WatchConfig
		if err = cur.Scan(&watch.ID, &watch.CalendarID, &watch.WatchID, &watch.Token); err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}
		watches = append(watches, watch)
	}

	return watches, nil
}

func (d *Database) CreateInviteConfig(ctx context.Context, calendarID, emailAddress string) error {
	stmt, err := d.db.PrepareContext(ctx, `
INSERT INTO invites (calendarID, emailAddress)
VALUES (?, ?)
`)
	if err != nil {
		return errors.Wrap(err, "failed to prepare statement")
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, calendarID, emailAddress); err != nil {
		return errors.Wrap(err, "failed to execute statement")
	}

	return nil
}

func (d *Database) DeleteInviteConfig(ctx context.Context, inviteID string) error {
	stmt, err := d.db.PrepareContext(ctx, `
DELETE FROM invites
WHERE id = ?`)
	if err != nil {
		return errors.Wrap(err, "failed to prepare statement")
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, inviteID); err != nil {
		return errors.Wrap(err, "failed to execute statement")
	}

	return nil
}

type settingType string

const expiryTimeFormat = time.RFC3339

const (
	stateSetting        settingType = "state"
	accessTokenSetting  settingType = "accessToken"
	refreshTokenSetting settingType = "refreshToken"
	tokenTypeSetting    settingType = "tokenType"
	expirySetting       settingType = "expiry"
)

func (d *Database) GetState(ctx context.Context) (string, error) {
	return d.getSetting(ctx, stateSetting)
}

func (d *Database) SetState(ctx context.Context, state string) error {
	return d.setSetting(ctx, stateSetting, state)
}

func (d *Database) GetTokens(ctx context.Context) (*oauth2.Token, error) {
	accessToken, err := d.getSetting(ctx, accessTokenSetting)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get access token")
	}

	refreshToken, err := d.getSetting(ctx, refreshTokenSetting)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get refresh token")
	}

	tokenType, err := d.getSetting(ctx, tokenTypeSetting)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get token type setting")
	}

	expiryString, err := d.getSetting(ctx, expirySetting)
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
	return d.setSetting(ctx, accessTokenSetting, accessToken)
}

func (d *Database) SetExpiry(ctx context.Context, expiry time.Time) error {
	expiryString := expiry.Format(expiryTimeFormat)

	return d.setSetting(ctx, expirySetting, expiryString)
}

var NoRefreshTokenErr = errors.New("no refresh token present")

func (d *Database) SetTokens(ctx context.Context, token *oauth2.Token) error {
	if token.RefreshToken == "" {
		return NoRefreshTokenErr
	}

	var err error
	if err = d.setSetting(ctx, refreshTokenSetting, token.RefreshToken); err != nil {
		return errors.Wrap(err, "failed to store refresh token")
	}

	if err = d.SetAccessToken(ctx, token.AccessToken); err != nil {
		return errors.Wrap(err, "failed to store access token")
	}

	if err = d.SetExpiry(ctx, token.Expiry); err != nil {
		return errors.Wrap(err, "failed to store expiry")
	}

	if err = d.setSetting(ctx, tokenTypeSetting, token.TokenType); err != nil {
		return errors.Wrap(err, "failed to store token type")
	}

	return nil
}

func (d *Database) RemoveTokens(ctx context.Context) error {
	var err error

	for _, setting := range []settingType{
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
	if err = d.setSetting(ctx, tokenTypeSetting, tokens.TokenType); err != nil {
		return errors.Wrap(err, "failed to store token type")
	}

	return nil
}
