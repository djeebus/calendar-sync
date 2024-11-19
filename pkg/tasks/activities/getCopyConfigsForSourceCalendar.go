package activities

import (
	"calendar-sync/pkg/persistence"
	"context"
	"github.com/pkg/errors"
)

type GetCopyConfigsForSourceCalendarArgs struct {
	CalendarID string
}

type GetCopyConfigsForSourceCalendarResult struct {
	CopyConfigs []persistence.CopyConfig
}

func (a Activities) GetCopyConfigsForSourceCalendar(ctx context.Context, args GetCopyConfigsForSourceCalendarArgs) (GetCopyConfigsForSourceCalendarResult, error) {
	var result GetCopyConfigsForSourceCalendarResult

	configs, err := a.ctr.Database.GetCopyConfigsBySourceCalendar(ctx, args.CalendarID)
	if err != nil {
		return result, errors.Wrap(err, "failed to get configs from the db")
	}

	result.CopyConfigs = configs
	return result, nil
}
