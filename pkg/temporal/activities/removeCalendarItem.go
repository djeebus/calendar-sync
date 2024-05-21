package activities

import (
	"calendar-sync/pkg/clients"
	"context"
	"github.com/pkg/errors"
)

type RemoveCalendarItemArgs struct {
	CalendarID, EventID string
}

type RemoveCalendarItemResult struct{}

func (a Activities) RemoveCalendarItem(ctx context.Context, args RemoveCalendarItemArgs) (RemoveCalendarItemResult, error) {
	var result RemoveCalendarItemResult

	tokens, err := a.ctr.Database.GetTokens(ctx)
	if err != nil {
		return result, errors.Wrap(err, "failed to get tokens")
	}

	client, err := clients.GetClient(ctx, a.ctr.OAuth2Config, tokens)
	if err != nil {
		return result, errors.Wrap(err, "failed to create client")
	}

	if err := client.Events.Delete(args.CalendarID, args.EventID).Do(); err != nil {
		return result, errors.Wrap(err, "failed to delete event")
	}

	return result, nil
}
