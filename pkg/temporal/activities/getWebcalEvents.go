package activities

import (
	"calendar-sync/pkg"
	"calendar-sync/pkg/clients"
	"context"
	"github.com/pkg/errors"
	"google.golang.org/api/calendar/v3"
	"time"
)

type GetCalendarEventsActivityArgs struct {
	CalendarID string
}

type GetCalendarEventsActivityResult struct {
	Calendar pkg.Calendar
}

var SearchWindow = time.Hour * 24 * 14

func rfc3339(t time.Time) string {
	return t.Format(time.RFC3339)
}

func (a Activities) GetCalendarEventsActivity(ctx context.Context, args GetCalendarEventsActivityArgs) (GetCalendarEventsActivityResult, error) {
	var result GetCalendarEventsActivityResult

	tokens, err := a.ctr.Database.GetTokens(ctx)
	if err != nil {
		return result, errors.Wrap(err, "failed to get tokens")
	}

	client, err := clients.GetClient(ctx, a.ctr.OAuth2Config, tokens)
	if err != nil {
		return result, errors.Wrap(err, "failed to create client")
	}

	c, err := client.Calendars.Get(args.CalendarID).Do()
	if err != nil {
		return result, errors.Wrap(err, "failed to retrieve calendar")
	}
	result.Calendar.CalendarID = c.Id
	result.Calendar.Summary = c.Summary

	now := time.Now()
	listCall := client.Events.List(args.CalendarID).
		TimeMin(rfc3339(now)).
		TimeMax(rfc3339(now.Add(SearchWindow))).
		MaxResults(100)

	if err = listCall.Pages(ctx, func(events *calendar.Events) error {
		for _, event := range events.Items {
			if event.Status == "cancelled" {
				continue
			}

			result.Calendar.Items = append(result.Calendar.Items, event)
		}

		return nil
	}); err != nil {
		return result, errors.Wrap(err, "failed to retrieve events")
	}

	return result, nil
}
