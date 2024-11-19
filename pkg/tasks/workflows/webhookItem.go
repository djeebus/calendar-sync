package workflows

import (
	"context"

	"github.com/pkg/errors"

	"calendar-sync/pkg/persistence"
	"calendar-sync/pkg/tasks/activities"
)

type ProcessWebhookEventArgs struct {
	ChannelID     string
	MessageNumber string
	ResourceID    string
	ResourceState string
	ResourceUri   string
	ChannelToken  string
}

func (w *Workflows) ProcessWebhookEvent(ctx context.Context, args ProcessWebhookEventArgs) error {
	ctx, log := setupLogger(ctx, "ProcessWebhookEvent")

	// channel has been created, this doesn't really represent an event
	if args.ResourceState == "sync" {
		return nil
	}

	getWatchArgs := activities.GetWatchArgs{WatchID: args.ChannelID}
	getWatchResult, err := w.a.GetWatch(ctx, getWatchArgs)
	if err != nil {
		return errors.Wrap(err, "failed to get watch")
	}

	watch := getWatchResult.Watch
	if watch.Token != args.ChannelToken {
		log.Warn().
			Str("watch_id", args.ChannelID).
			Str("bad_token", args.ChannelToken).
			Str("real_token", watch.Token).
			Msg("token mismatch")
		return nil
	}

	switch args.ResourceState {
	case "exists":
		// resource was created or modified, do something!
		return w.processEventUpsert(ctx, args, watch)
	case "not_exists":
		return w.processEventDelete(ctx, args, watch)
	default:
		log.Warn().
			Str("state", args.ResourceState).
			Str("uri", args.ResourceUri).
			Msg("unknown resource state")
		return nil
	}
}

func (w *Workflows) processEventUpsert(ctx context.Context, args ProcessWebhookEventArgs, watch persistence.WatchConfig) error {
	activityArgs := activities.UpsertCalendarEntryArgs{CalendarID: watch.CalendarID, EventID: args.ResourceID}
	_, err := w.a.UpsertCalendarEntry(ctx, activityArgs)
	return err
}

func (w *Workflows) processEventDelete(ctx context.Context, args ProcessWebhookEventArgs, watch persistence.WatchConfig) error {
	activityArgs := activities.RemoveCalendarItemArgs{CalendarID: watch.CalendarID, EventID: args.ResourceID}
	_, err := w.a.RemoveCalendarItem(ctx, activityArgs)
	return err
}
