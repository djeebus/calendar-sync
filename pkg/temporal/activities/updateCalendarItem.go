package activities

import (
	"context"
	"github.com/pkg/errors"
	"google.golang.org/api/calendar/v3"
)

type UpdateCalendarItemArgs struct {
	CalendarID     string
	CalendarItemID string
	Patch          *calendar.Event
}

type UpdateCalendarItemResult struct{}

func (a Activities) UpdateCalendarItem(ctx context.Context, args UpdateCalendarItemArgs) (UpdateCalendarItemResult, error) {
	var result UpdateCalendarItemResult

	client, err := a.ctr.GetCalendarClient(ctx)
	if err != nil {
		return result, errors.Wrap(err, "failed to create client")
	}

	if _, err := client.Events.Patch(args.CalendarID, args.CalendarItemID, args.Patch).Do(); err != nil {
		return result, errors.Wrap(err, "failed to patch event")
	}

	return result, nil
}
