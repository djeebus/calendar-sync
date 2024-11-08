package sqlite

import (
	"context"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"strconv"

	"calendar-sync/pkg"
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
