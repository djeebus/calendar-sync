package activities

import (
	"calendar-sync/pkg/clients"
	"context"
	"github.com/pkg/errors"
	"google.golang.org/api/calendar/v3"
)

type InviteGuestArgs struct {
	CalendarID           string
	CalendarItemID       string
	Attendees            []*calendar.EventAttendee
	EmailAddressToInvite string
}

type InviteGuestResult struct {
}

func (a Activities) UpdateGuestList(ctx context.Context, args InviteGuestArgs) (InviteGuestResult, error) {
	var result InviteGuestResult

	tokens, err := a.ctr.Database.GetTokens(ctx)
	if err != nil {
		return result, errors.Wrap(err, "failed to get tokens")
	}

	client, err := clients.GetClient(ctx, tokens)
	if err != nil {
		return result, errors.Wrap(err, "failed to create client")
	}

	patch := client.Events.Patch(args.CalendarID, args.CalendarItemID, &calendar.Event{
		Attendees: append(args.Attendees, &calendar.EventAttendee{
			AdditionalGuests: 1,
			Email:            args.EmailAddressToInvite,
		}),
	})
	if _, err = patch.Do(); err != nil {
		return result, errors.Wrap(err, "failed to patch event")
	}

	return result, nil
}
