package workflows

import (
	"context"

	"github.com/pkg/errors"

	"calendar-sync/pkg/logs"
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
	case "exists", "not_exists":
		// resource was created or modified, do something!
		w.processInvites(ctx, watch.CalendarID)
		w.processCopyConfigs(ctx, watch.CalendarID)
	default:
		log.Warn().
			Str("state", args.ResourceState).
			Str("uri", args.ResourceUri).
			Msg("unknown resource state")
	}

	return nil
}

func (w *Workflows) processCopyConfigs(ctx context.Context, calendarID string) {
	log := logs.GetLogger(ctx)

	copyConfigResult, err := w.a.GetCopyConfigsForSourceCalendar(ctx, activities.GetCopyConfigsForSourceCalendarArgs{
		CalendarID: calendarID,
	})
	if err != nil {
		log.Error().Err(err).Str("calendar-id", calendarID).Msg("failed to get copy configs")
		return
	}

	for _, config := range copyConfigResult.CopyConfigs {
		if err = w.CopyCalendarWorkflow(ctx, CopyCalendarWorkflowArgs{
			SourceCalendarID:      config.SourceID,
			DestinationCalendarID: config.DestinationID,
		}); err != nil {
			log.Error().
				Err(err).
				Str("destination-calendar-id", config.DestinationID).
				Str("source-calendar-id", config.SourceID).
				Msg("failed to copy calendar events")
		}
	}
}

func (w *Workflows) processInvites(ctx context.Context, calendarID string) {
	log := logs.GetLogger(ctx)

	inviteConfigResult, err := w.a.GetInviteConfigsForSourceCalendar(ctx, activities.GetInviteConfigsForSourceCalendarArgs{
		CalendarID: calendarID,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to get invite configs")
		return
	}

	for _, config := range inviteConfigResult.Configs {
		args := InviteCalendarWorkflowArgs{
			CalendarID: calendarID,
			EmailToAdd: config.EmailAddress,
		}
		if err = w.InviteCalendarWorkflow(ctx, args); err != nil {
			log.Error().Err(err).
				Str("calendar-id", config.CalendarID).
				Str("email-address", config.EmailAddress).
				Msg("failed to update guest list")
		}
	}
}
