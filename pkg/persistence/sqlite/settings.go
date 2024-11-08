package sqlite

import (
	"context"
	"github.com/pkg/errors"
)

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
