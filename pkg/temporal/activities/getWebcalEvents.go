package activities

import (
	"calendar-sync/pkg"
	"calendar-sync/pkg/clients"
	"context"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"time"
)

type GetCalendarEventsActivityArgs struct {
	CalendarID string

	Token *oauth2.Token
}

type GetCalendarEventsActivityResult struct {
	Calendar pkg.Calendar
}

var SearchWindow = time.Duration(time.Hour * 24 * 14)

func rfc3339(t time.Time) string {
	return t.Format(time.RFC3339)
}

func GetCalendarEventsActivity(ctx context.Context, args GetCalendarEventsActivityArgs) (GetCalendarEventsActivityResult, error) {
	var result GetCalendarEventsActivityResult

	client, err := clients.GetClient(ctx, args.Token)
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
			result.Calendar.Items = append(result.Calendar.Items, pkg.CalendarItem{
				CalendarID:     c.Id,
				CalendarItemID: event.Id,
				Summary:        event.Summary,
				Guests:         event.Attendees,
			})
		}

		return nil
	}); err != nil {
		return result, errors.Wrap(err, "failed to retrieve events")
	}

	return result, nil
}
