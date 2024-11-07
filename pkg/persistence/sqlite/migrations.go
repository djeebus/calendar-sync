package sqlite

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
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    calendarID TEXT NOT NULL,
    watchID    TEXT NOT NULL,
    token    TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS watches_calendarID ON watches (calendarID);
CREATE UNIQUE INDEX IF NOT EXISTS watches_watchID ON watches (watchID);
`,
	2: `
ALTER TABLE watches ADD COLUMN expiration DATE;
`,
}
