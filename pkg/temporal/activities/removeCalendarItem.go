package activities

import (
	"context"
	"github.com/pkg/errors"
)

type RemoveCalendarItemArgs struct {
	CalendarID, EventID string
}

type RemoveCalendarItemResult struct{}

func (a Activities) RemoveCalendarItem(ctx context.Context, args RemoveCalendarItemArgs) (RemoveCalendarItemResult, error) {
	var result RemoveCalendarItemResult

	client, err := a.ctr.GetCalendarClient(ctx)
	if err != nil {
		return result, errors.Wrap(err, "failed to create client")
	}

	if err := client.Events.Delete(args.CalendarID, args.EventID).Do(); err != nil {
		return result, errors.Wrap(err, "failed to delete event")
	}

	return result, nil
}
