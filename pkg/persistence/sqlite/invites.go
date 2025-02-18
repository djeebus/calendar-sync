package sqlite

import (
	"context"

	"github.com/pkg/errors"

	"calendar-sync/pkg/logs"
	"calendar-sync/pkg/persistence"
)

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

	logs.GetLogger(ctx).Info().
		Str("calendar-id", calendarID).
		Str("email-address", emailAddress).
		Msgf("created invite config")

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

	logs.GetLogger(ctx).Info().
		Str("invite-id", inviteID).
		Msgf("deleted invite config")

	return nil
}

func (d *Database) GetInviteConfig(ctx context.Context, id int64) (persistence.InviteConfig, error) {
	var config persistence.InviteConfig

	stmt, err := d.db.PrepareContext(ctx, `
SELECT id, calendarID, emailAddress
FROM invites
WHERE id = ?`)
	if err != nil {
		return persistence.InviteConfig{}, errors.Wrap(err, "failed to prepare statement")
	}
	defer stmt.Close()

	if err := stmt.QueryRowContext(ctx, id).Scan(&config.ID, &config.CalendarID, &config.EmailAddress); err != nil {
		return persistence.InviteConfig{}, errors.Wrap(err, "failed to parse row")
	}

	return config, nil
}

func (d *Database) queryInviteConfigs(ctx context.Context, query string, args ...any) ([]persistence.InviteConfig, error) {
	stmt, err := d.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare statement")
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query context")
	}
	defer rows.Close()

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to get rows")
	}

	var configs []persistence.InviteConfig
	for rows.Next() {
		var config persistence.InviteConfig
		if err = rows.Scan(&config.ID, &config.CalendarID, &config.EmailAddress); err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}

		configs = append(configs, config)
	}

	return configs, nil
}

func (d *Database) GetInviteConfigs(ctx context.Context) ([]persistence.InviteConfig, error) {
	return d.queryInviteConfigs(ctx, `
SELECT id, calendarID, emailAddress 
FROM invites
`)
}

func (d *Database) GetInviteConfigsBySourceCalendar(ctx context.Context, calendarID string) ([]persistence.InviteConfig, error) {
	return d.queryInviteConfigs(ctx, `
SELECT id, calendarID, emailAddress 
FROM invites
WHERE calendarID = ?
`, calendarID)
}
