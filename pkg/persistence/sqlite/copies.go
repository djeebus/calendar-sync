package sqlite

import (
	"context"

	"github.com/pkg/errors"

	"calendar-sync/pkg/persistence"
)

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

func (d *Database) GetCopyConfig(ctx context.Context, id int64) (persistence.CopyConfig, error) {
	var config persistence.CopyConfig

	stmt, err := d.db.PrepareContext(ctx, `
SELECT id, sourceID, destinationID
FROM copies
WHERE id = ?`)
	if err != nil {
		return persistence.CopyConfig{}, errors.Wrap(err, "failed to prepare statement")
	}
	defer stmt.Close()

	if err := stmt.QueryRowContext(ctx, id).Scan(&config.ID, &config.SourceID, &config.DestinationID); err != nil {
		return persistence.CopyConfig{}, errors.Wrap(err, "failed to parse row")
	}

	return config, nil
}

func (d *Database) GetCopyConfigs(ctx context.Context) ([]persistence.CopyConfig, error) {
	stmt, err := d.db.PrepareContext(ctx, `
SELECT id, sourceID, destinationID
FROM copies
`)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare statement")
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute statement")
	}
	defer rows.Close()

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error getting rows")
	}

	var configs []persistence.CopyConfig
	for rows.Next() {
		var config persistence.CopyConfig
		if err = rows.Scan(&config.ID, &config.SourceID, &config.DestinationID); err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}
		configs = append(configs, config)
	}

	return configs, nil
}
