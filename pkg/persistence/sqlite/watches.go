package sqlite

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"calendar-sync/pkg/logs"
	"calendar-sync/pkg/persistence"
)

var ErrMustHaveExpirationTime = errors.New("must have an expiration time")

func (d *Database) CreateWatchConfig(ctx context.Context, calendarID, watchID, token string, expiration time.Time) error {
	if expiration.IsZero() {
		return ErrMustHaveExpirationTime
	}

	stmt, err := d.db.PrepareContext(ctx, `
INSERT INTO watches (calendarID, watchID, token, expiration)
VALUES (?, ?, ?, ?)
`)
	if err != nil {
		return errors.Wrap(err, "failed to prepare statement")
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, calendarID, watchID, token, expiration); err != nil {
		return errors.Wrap(err, "failed to execute statement")
	}

	logs.GetLogger(ctx).Info().
		Str("calendar-id", calendarID).
		Msgf("created new watch config")

	return nil
}

func (d *Database) GetWatchConfig(ctx context.Context, watchID string) (persistence.WatchConfig, error) {
	var config persistence.WatchConfig

	stmt, err := d.db.PrepareContext(ctx, `
SELECT id, calendarID, watchID, token, expiration
FROM watches
WHERE watchID = ?`)
	if err != nil {
		return config, errors.Wrap(err, "failed to prepare statement")
	}
	defer stmt.Close()

	if err := stmt.QueryRowContext(ctx, watchID).Scan(&config.ID, &config.CalendarID, &config.WatchID, &config.Token, &config.Expiration); err != nil {
		return config, errors.Wrap(err, "failed to parse row")
	}

	return config, nil
}

func (d *Database) GetWatchConfigs(ctx context.Context) ([]persistence.WatchConfig, error) {
	stmt, err := d.db.PrepareContext(ctx, `
SELECT id, calendarID, watchID, token, expiration
FROM watches
WHERE expiration IS NOT NULL
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
		return nil, errors.Wrap(err, "failed to get rows")
	}

	var watches []persistence.WatchConfig
	for rows.Next() {
		var watch persistence.WatchConfig
		if err = rows.Scan(&watch.ID, &watch.CalendarID, &watch.WatchID, &watch.Token, &watch.Expiration); err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}

		watches = append(watches, watch)
	}

	return watches, nil
}

func (d *Database) DeleteWatchConfig(ctx context.Context, watchID int) error {
	stmt, err := d.db.PrepareContext(ctx, `DELETE FROM watches WHERE ID = ?`)
	if err != nil {
		return errors.Wrap(err, "failed to prepare")
	}
	defer stmt.Close()

	if _, err = stmt.ExecContext(ctx, watchID); err != nil {
		return errors.Wrap(err, "failed to exec")
	}

	logs.GetLogger(ctx).Info().
		Int("watch-id", watchID).
		Msgf("deleted watch config")

	return nil
}
