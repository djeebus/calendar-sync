package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

func NewDatabase() (*Database, error) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return nil, errors.Wrap(err, "failed to epen db")
	}

	if _, err = db.Exec(`
CREATE TABLE IF NOT EXISTS settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    value TEXT NOT NULL
)`); err != nil {
		return nil, errors.Wrap(err, "failed to create settings table")
	}

	if _, err = db.Exec(`
CREATE UNIQUE INDEX IF NOT EXISTS settings_name ON settings(name)`); err != nil {
		return nil, errors.Wrap(err, "failed to create unique settings index")
	}

	if _, err = db.Exec(`
CREATE TABLE IF NOT EXISTS copies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sourceID TEXT NOT NULL, 
    destinationID TEXT NOT NULL
)`); err != nil {
		return nil, errors.Wrap(err, "failed to create copies table")
	}

	if _, err = db.Exec(`
CREATE UNIQUE INDEX IF NOT EXISTS copies_sourceID_destinationID ON copies (sourceID, destinationID)
`); err != nil {
		return nil, errors.Wrap(err, "failed to create unique copies index")
	}

	if _, err = db.Exec(`
CREATE TABLE IF NOT EXISTS invites (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    calendarID TEXT NOT NULL, 
    emailAddress TEXT NOT NULL
)`); err != nil {
		return nil, errors.Wrap(err, "failed to create invites table")
	}

	if _, err = db.Exec(`
CREATE UNIQUE INDEX IF NOT EXISTS invites_calendarID_emailAddress ON invites (calendarID, emailAddress)
`); err != nil {
		return nil, errors.Wrap(err, "failed to create unique invites index")
	}

	return &Database{db: db}, nil
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

type settingType string

const (
	tokenSetting settingType = "tokens"
	stateSetting settingType = "state"
)

func (d *Database) GetTokens(ctx context.Context) (*oauth2.Token, error) {
	val, err := d.getSetting(ctx, tokenSetting)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get tokens")
	}

	var tokens oauth2.Token
	if err := json.Unmarshal([]byte(val), &tokens); err != nil {
		if err2 := d.removeSetting(ctx, tokenSetting); err2 != nil {
			log.Warn().Err(err2).Msg("failed to remove broken token")
		}
		return nil, errors.Wrap(err, "failed to unmarshal tokens")
	}

	return &tokens, nil
}

func (d *Database) RemoveTokens(ctx context.Context) error {
	return d.removeSetting(ctx, tokenSetting)
}

func (d *Database) SetTokens(ctx context.Context, tokens *oauth2.Token) error {
	data, err := json.Marshal(tokens)
	if err != nil {
		return errors.Wrap(err, "failed to marshal tokens")
	}

	return d.setSetting(ctx, tokenSetting, string(data))
}

func (d *Database) GetState(ctx context.Context) (string, error) {
	return d.getSetting(ctx, stateSetting)
}

func (d *Database) SetState(ctx context.Context, state string) error {
	return d.setSetting(ctx, stateSetting, state)
}
