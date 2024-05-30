package activities

import (
	"context"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"google.golang.org/api/calendar/v3"
)

type WatchCalendarArgs struct {
	CalendarID string
}

type WatchCalendarResult struct {
	WatchID string
}

func (a Activities) WatchCalendar(ctx context.Context, args WatchCalendarArgs) (WatchCalendarResult, error) {
	var result WatchCalendarResult

	client, err := a.ctr.GetCalendarClient(ctx)
	if err != nil {
		return result, errors.Wrap(err, "failed to create client")
	}

	channel := calendar.Channel{
		Id:      uuid.NewString(),
		Type:    "web_hook",
		Address: a.ctr.Config.WebhookUrl,
		Token:   uuid.NewString(),
	}

	if _, err := client.Events.Watch(args.CalendarID, &channel).Do(); err != nil {
		return result, errors.Wrap(err, "failed to watch events")
	}

	if err := a.ctr.Database.CreateWatchConfig(ctx, args.CalendarID, channel.Id, channel.Token); err != nil {
		return result, errors.Wrap(err, "failed to write row")
	}

	result.WatchID = channel.Id

	return result, nil
}
