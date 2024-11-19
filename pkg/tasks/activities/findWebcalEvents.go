package activities

import (
	"calendar-sync/pkg"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"google.golang.org/api/calendar/v3"
)

type FindWebcalEventsArgs struct {
	DestinationCalendarID string
	SourceCalendarID      string
	SourceCalendarItemID  string
}

type FindWebcalEventsResults struct {
	Items []*calendar.Event
}

func (a Activities) FindDestinationWebcalEvent(ctx context.Context, args FindWebcalEventsArgs) (FindWebcalEventsResults, error) {
	var result FindWebcalEventsResults

	client, err := a.ctr.GetCalendarClient(ctx)
	if err != nil {
		return result, errors.Wrap(err, "failed to create calendar client")
	}

	response, err := client.Events.List(args.DestinationCalendarID).PrivateExtendedProperty(
		fmt.Sprintf("%s=%s", pkg.SourceCalendarIDKey, args.SourceCalendarID),
		fmt.Sprintf("%s=%s", pkg.SourceCalendarItemIDKey, args.SourceCalendarItemID),
	).Context(ctx).Do()
	if err != nil {
		return result, errors.Wrap(err, "failed to list events")
	}

	result.Items = response.Items
	return result, nil
}
