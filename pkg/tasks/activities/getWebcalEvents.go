package activities

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/api/calendar/v3"

	"calendar-sync/pkg"
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
	ctx = setupLogger(ctx, "GetCalendarEventsActivity")

	var result GetCalendarEventsActivityResult

	client, err := a.ctr.GetCalendarClient(ctx)
	if err != nil {
		return result, errors.Wrap(err, "failed to create client")
	}

	c, err := client.Calendars.Get(args.CalendarID).Context(ctx).Do()
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
