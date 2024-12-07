package workflows

import (
	"context"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"calendar-sync/pkg/logs"
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
	case "exists", "not_exists":
		// resource was created or modified, do something!
		return w.processEventUpsert(ctx, args, watch, args.ResourceState == "not_exists")
	default:
		log.Warn().
			Str("state", args.ResourceState).
			Str("uri", args.ResourceUri).
			Msg("unknown resource state")
		return nil
	}
}

func (w *Workflows) processEventUpsert(
	ctx context.Context,
	args ProcessWebhookEventArgs,
	watch persistence.WatchConfig,
	isDelete bool,
) error {
	log := logs.GetLogger(ctx)

	calendarItem, err := w.a.GetCalendarItemByItemID(ctx, activities.GetCalendarItemByItemIDArgs{
		CalendarID: watch.CalendarID,
		EventID:    args.ResourceID,
	})
	if err != nil {
		return errors.Wrap(err, "failed to get calendar item")
	}

	copyConfigResult, err := w.a.GetCopyConfigsForSourceCalendar(ctx, activities.GetCopyConfigsForSourceCalendarArgs{
		CalendarID: watch.CalendarID,
	})
	if err != nil {
		return errors.Wrap(err, "failed to get copy configs")
	}

	w.processInvites(ctx, watch.CalendarID, calendarItem)

	w.processCopyConfigs(ctx, args, watch, isDelete, copyConfigResult, log, calendarItem)

	return nil
}

func (w *Workflows) processCopyConfigs(ctx context.Context, args ProcessWebhookEventArgs, watch persistence.WatchConfig, isDelete bool, copyConfigResult activities.GetCopyConfigsForSourceCalendarResult, log *zerolog.Logger, calendarItem activities.GetCalendarItemByItemIDResult) {
	for _, config := range copyConfigResult.CopyConfigs {
		result, err := w.a.FindDestinationWebcalEvent(ctx, activities.FindWebcalEventsArgs{
			DestinationCalendarID: config.DestinationID,
			SourceCalendarID:      config.SourceID,
			SourceCalendarItemID:  args.ResourceID,
		})
		if err != nil {
			log.Error().
				Err(err).
				Str("destination-calendar-id", config.DestinationID).
				Str("source-calendar-id", config.SourceID).
				Str("source-calendar-item-id", args.ResourceID).
				Msg("failed to find copied calendar events")
			continue
		}

		// delete copied items
		if isDelete {
			for _, item := range result.Items {
				if _, err := w.a.RemoveCalendarItem(ctx, activities.RemoveCalendarItemArgs{
					CalendarID: config.DestinationID,
					EventID:    item.Id,
				}); err != nil {
					log.Error().
						Err(err).
						Str("calendar-id", config.DestinationID).
						Str("event-id", item.Id).
						Msg("failed to delete calendar item")
				}
			}
			continue
		}

		// create copied item
		if len(result.Items) == 0 {
			if _, err := w.a.CreateCalendarItem(ctx, activities.CreateCalendarItemArgs{
				CalendarID: config.DestinationID,
				Event:      toInsert(watch.CalendarID, calendarItem.Event),
			}); err != nil {
				log.Error().
					Err(err).
					Str("source-calendar-id", watch.CalendarID).
					Str("destination-calendar-id", config.DestinationID).
					Str("event-id", calendarItem.Event.Id).
					Msg("failed to create calendar item")
			}
			continue
		}

		// update copied items
		for _, item := range result.Items {
			if patch := buildPatch(*log, *calendarItem.Event, *item); patch != nil {
				if _, err := w.a.UpdateCalendarItem(ctx, activities.UpdateCalendarItemArgs{
					CalendarID:     config.DestinationID,
					CalendarItemID: item.Id,
					Patch:          patch,
				}); err != nil {
					log.Error().
						Err(err).
						Str("source-calendar-id", watch.CalendarID).
						Str("destination-calendar-id", config.DestinationID).
						Str("source-event-id", calendarItem.Event.Id).
						Str("destination-event-id", item.Id).
						Msg("failed to update calendar item")
				}
			}
		}
	}
}

func (w *Workflows) processInvites(ctx context.Context, calendarID string, calendarItem activities.GetCalendarItemByItemIDResult) {
	log := logs.GetLogger(ctx)

	inviteConfigResult, err := w.a.GetInviteConfigsForSourceCalendar(ctx, activities.GetInviteConfigsForSourceCalendarArgs{
		CalendarID: calendarID,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to get invite configs")
		return
	}

	for _, config := range inviteConfigResult.Configs {
		if guestsContains(calendarItem.Event.Attendees, config.EmailAddress) {
			continue
		}

		inviteArgs := activities.InviteGuestArgs{
			CalendarID:           config.CalendarID,
			CalendarItemID:       calendarItem.Event.Id,
			EmailAddressToInvite: config.EmailAddress,
			Attendees:            calendarItem.Event.Attendees,
		}

		if _, err := w.a.UpdateGuestList(ctx, inviteArgs); err != nil {
			log.Error().Err(err).
				Str("calendar-id", config.CalendarID).
				Str("calendar-item-id", calendarItem.Event.Id).
				Str("email-address", config.EmailAddress).
				Msg("failed to update guest list")
		}
	}
}
