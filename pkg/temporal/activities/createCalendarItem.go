package activities

import (
	"calendar-sync/pkg/clients"
	"context"
	"github.com/pkg/errors"
	"google.golang.org/api/calendar/v3"
)

type CreateCalendarItemArgs struct {
	CalendarID string
	Event      *calendar.Event
}

type CreateCalendarItemResult struct {
	CreatedItem *calendar.Event
}

func (a Activities) CreateCalendarItem(ctx context.Context, args CreateCalendarItemArgs) (CreateCalendarItemResult, error) {
	var result CreateCalendarItemResult

	tokens, err := a.ctr.Database.GetTokens(ctx)
	if err != nil {
		return result, errors.Wrap(err, "failed to get tokens")
	}

	client, err := clients.GetClient(ctx, tokens)
	if err != nil {
		return result, errors.Wrap(err, "failed to create client")
	}

	created, err := client.Events.Insert(args.CalendarID, args.Event).Do()
	if err != nil {
		return result, errors.Wrap(err, "failed to create event")
	}

	result.CreatedItem = created

	return result, nil
}
