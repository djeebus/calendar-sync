package sqlite

import (
	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"

	"calendar-sync/pkg"
)

const dbVersionSetting settingType = "db_version"

func NewDatabase(ctx context.Context, cfg pkg.Config) (*Database, error) {
	conn, err := sql.Open(cfg.DatabaseDriver, cfg.DatabaseSource)
	if err != nil {
		return nil, errors.Wrap(err, "failed to epen db")
	}

	db := &Database{db: conn}
	if err := migrate(ctx, db, conn); err != nil {
		return nil, errors.Wrap(err, "failed to migrate")
	}

	return db, nil
}

type Database struct {
	db *sql.DB
}

func (d *Database) Close() {
	d.db.Close()
}
