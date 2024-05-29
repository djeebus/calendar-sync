package workflows

import (
	"calendar-sync/pkg/temporal/activities"
	"github.com/rs/zerolog/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/api/calendar/v3"
	"time"
)

type InviteCalendarWorkflowArgs struct {
	CalendarID string
	EmailToAdd string
}

func InviteCalendarWorkflow(ctx workflow.Context, args InviteCalendarWorkflowArgs) error {
	// setup
	var a activities.Activities

	retryPolicy := temporal.RetryPolicy{
		InitialInterval:    1 * time.Minute,
		BackoffCoefficient: 2.0,

		MaximumAttempts:        1,
		MaximumInterval:        1 * time.Hour,
		NonRetryableErrorTypes: []string{},
	}

	options := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy:         &retryPolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, options)

	// get events from calendar
	eventArgs := activities.GetCalendarEventsActivityArgs{
		CalendarID: args.CalendarID,
	}

	var eventResult activities.GetCalendarEventsActivityResult
	if err := workflow.ExecuteActivity(ctx, a.GetCalendarEventsActivity, eventArgs).Get(ctx, &eventResult); err != nil {
		return err
	}

	// find missing guests
	futures := make([]workflow.Future, 0)
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
		fut := workflow.ExecuteActivity(ctx, a.UpdateGuestList, inviteArgs)
		futures = append(futures, fut)
	}

	for _, f := range futures {
		var result activities.InviteGuestResult
		if err := f.Get(ctx, &result); err != nil {
			log.Error().Err(err).Msg("failed to invite guest")
		}
	}

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
