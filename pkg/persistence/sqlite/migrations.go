package sqlite

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/pkg/errors"
)

var migrations = map[int]string{
	0: `
CREATE TABLE IF NOT EXISTS settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    value TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS settings_name ON settings(name);

CREATE TABLE IF NOT EXISTS copies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sourceID TEXT NOT NULL, 
    destinationID TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS copies_sourceID_destinationID ON copies (sourceID, destinationID);

CREATE TABLE IF NOT EXISTS invites (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    calendarID TEXT NOT NULL, 
    emailAddress TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS invites_calendarID_emailAddress ON invites (calendarID, emailAddress);
`,
	1: `
CREATE TABLE IF NOT EXISTS watches (
    id 			INTEGER PRIMARY KEY AUTOINCREMENT,
    calendarID 	TEXT 	NOT NULL,
    watchID    	TEXT 	NOT NULL,
    token    	TEXT 	NOT NULL,
   	expiration	DATE
);

CREATE UNIQUE INDEX IF NOT EXISTS watches_calendarID ON watches (calendarID);
CREATE UNIQUE INDEX IF NOT EXISTS watches_watchID ON watches (watchID);
`,
	2: `
ALTER TABLE watches DROP COLUMN expiration;
ALTER TABLE watches ADD COLUMN expiration DATETIME;
`,
}

func migrate(ctx context.Context, db *Database, conn *sql.DB) error {
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
			return errors.Wrapf(err, "failed to migrate to v%d", nextVersion)
		}
		if err = db.setSetting(ctx, dbVersionSetting, strconv.Itoa(nextVersion)); err != nil {
			return errors.Wrap(err, "failed to persist db version")
		}
		nextVersion++
	}
	return nil
}
