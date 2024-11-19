package activities

import (
	"context"
	"github.com/pkg/errors"
	"google.golang.org/api/calendar/v3"
)

type GetCalendarItemByItemIDArgs struct {
	CalendarID, EventID string
}

type GetCalendarItemByItemIDResult struct {
	Event *calendar.Event
}

func (a Activities) GetCalendarItemByItemID(ctx context.Context, args GetCalendarItemByItemIDArgs) (GetCalendarItemByItemIDResult, error) {
	var result GetCalendarItemByItemIDResult

	client, err := a.ctr.GetCalendarClient(ctx)
	if err != nil {
		return result, errors.Wrap(err, "failed to get calendar client")
	}

	response, err := client.Events.Get(args.CalendarID, args.EventID).Context(ctx).Do()
	if err != nil {
		return result, errors.Wrap(err, "failed to get event")
	}

	result.Event = response

	return result, nil
}
