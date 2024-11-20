package activities

import (
	"context"

	"github.com/pkg/errors"

	"calendar-sync/pkg/persistence"
)

type GetInviteConfigsForSourceCalendarArgs struct {
	CalendarID string
}

type GetInviteConfigsForSourceCalendarResult struct {
	Configs []persistence.InviteConfig
}

func (a Activities) GetInviteConfigsForSourceCalendar(ctx context.Context, args GetInviteConfigsForSourceCalendarArgs) (GetInviteConfigsForSourceCalendarResult, error) {
	var result GetInviteConfigsForSourceCalendarResult

	configs, err := a.ctr.Database.GetInviteConfigsBySourceCalendar(ctx, args.CalendarID)
	if err != nil {
		return result, errors.Wrap(err, "failed to get configs from the db")
	}

	result.Configs = configs
	return result, nil
}
