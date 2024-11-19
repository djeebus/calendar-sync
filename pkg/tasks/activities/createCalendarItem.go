package activities

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/api/calendar/v3"

	"calendar-sync/pkg/logs"
)

type CreateCalendarItemArgs struct {
	CalendarID string
	Event      *calendar.Event
}

type CreateCalendarItemResult struct {
	CreatedItem *calendar.Event
}

func (a Activities) CreateCalendarItem(ctx context.Context, args CreateCalendarItemArgs) (CreateCalendarItemResult, error) {
	ctx = setupLogger(ctx, "CreateCalendarItem")
	log := logs.GetLogger(ctx)

	var result CreateCalendarItemResult

	log.Info().Msg("get calendar client")
	client, err := a.ctr.GetCalendarClient(ctx)
	if err != nil {
		return result, errors.Wrap(err, "failed to create client")
	}

	log.Info().Str("calendar-id", args.CalendarID).Msg("insert event into calendar")
	created, err := client.Events.Insert(args.CalendarID, args.Event).Context(ctx).Do()
	if err != nil {
		return result, errors.Wrap(err, "failed to create event")
	}

	result.CreatedItem = created

	return result, nil
}
