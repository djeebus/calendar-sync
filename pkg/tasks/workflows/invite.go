package workflows

import (
	"context"
	"sync"

	"google.golang.org/api/calendar/v3"

	"calendar-sync/pkg/tasks/activities"
)

type InviteCalendarWorkflowArgs struct {
	CalendarID string
	EmailToAdd string
}

func (w *Workflows) InviteCalendarWorkflow(ctx context.Context, args InviteCalendarWorkflowArgs) error {
	ctx, log := setupLogger(ctx, "InviteCalendarWorkflow")

	// get events from calendar
	eventArgs := activities.GetCalendarEventsActivityArgs{
		CalendarID: args.CalendarID,
	}

	eventResult, err := w.a.GetCalendarEventsActivity(ctx, eventArgs)
	if err != nil {
		return err
	}

	// find missing guests
	var wg sync.WaitGroup
	for _, item := range eventResult.Calendar.Items {
		if guestsContains(item.Attendees, args.EmailToAdd) {
			continue
		}

		inviteArgs := activities.InviteGuestArgs{
			CalendarID:           args.CalendarID,
			CalendarItemID:       item.Id,
			EmailAddressToInvite: args.EmailToAdd,
			Attendees:            item.Attendees,
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := w.a.UpdateGuestList(ctx, inviteArgs); err != nil {
				log.Error().Err(err).
					Str("calendar-id", args.CalendarID).
					Str("calendar-item-id", item.Id).
					Str("email-address", args.EmailToAdd).
					Msg("failed to update guest list")
			}
		}()
	}

	wg.Wait()
	return nil
}

func guestsContains(guests []*calendar.EventAttendee, add string) bool {
	for _, guest := range guests {
		if guest.Email == add {
			return true
		}
	}

	return false
}
