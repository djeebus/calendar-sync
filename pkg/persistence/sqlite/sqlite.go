package sqlite

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"calendar-sync/pkg"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
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
