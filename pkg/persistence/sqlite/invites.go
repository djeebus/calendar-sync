package sqlite

import (
	"calendar-sync/pkg/persistence"
	"context"
	"github.com/pkg/errors"
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

func (d *Database) GetInviteConfigs(ctx context.Context) ([]persistence.InviteConfig, error) {
	stmt, err := d.db.PrepareContext(ctx, `
SELECT id, calendarID, emailAddress 
FROM invites
`)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare statement")
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx)
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
