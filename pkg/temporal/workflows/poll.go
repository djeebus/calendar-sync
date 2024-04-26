package workflows

import (
	"calendar-sync/pkg/temporal/activities"
	"github.com/rs/zerolog/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"time"
)

type DiffCalendarWorkflowArgs struct {
	CalendarID string
	EmailToAdd string

	Token *oauth2.Token
}

func DiffCalendarWorkflow(ctx workflow.Context, args DiffCalendarWorkflowArgs) error {
	// setup
	retryPolicy := temporal.RetryPolicy{
		InitialInterval:    1 * time.Minute,
		BackoffCoefficient: 2.0,

		MaximumAttempts:        0, // infinite attempts
		MaximumInterval:        1 * time.Hour,
		NonRetryableErrorTypes: []string{},
	}

	options := workflow.ActivityOptions{
		StartToCloseTimeout: 15 * time.Minute,
		RetryPolicy:         &retryPolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, options)

	// get events from calendar
	eventArgs := activities.GetCalendarEventsActivityArgs{
		CalendarID: args.CalendarID,
		Token:      args.Token,
	}

	var eventResult activities.GetCalendarEventsActivityResult
	if err := workflow.ExecuteActivity(ctx, activities.GetCalendarEventsActivity, eventArgs).Get(ctx, &eventResult); err != nil {
		return nil
	}

	// find missing guests
	futures := make([]workflow.Future, 0)
	for _, item := range eventResult.Calendar.Items {
		if guestsContains(item.Guests, args.EmailToAdd) {
			continue
		}

		inviteArgs := activities.InviteGuestArgs{
			CalendarID:           item.CalendarID,
			CalendarItemID:       item.CalendarItemID,
			EmailAddressToInvite: args.EmailToAdd,
			Attendees:            item.Guests,
			Token:                args.Token,
		}
		inviteOpts := workflow.ActivityOptions{
			ActivityID: "",
		}
		fut := workflow.ExecuteActivity(workflow.WithActivityOptions(ctx, inviteOpts), activities.UpdateGuestList, inviteArgs)
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
