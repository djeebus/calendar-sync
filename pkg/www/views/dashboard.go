package views

import (
	"calendar-sync/pkg/clients"
	"database/sql"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"google.golang.org/api/calendar/v3"
)

type calendarStub struct {
	ID         string
	Label      string
	AccessRole string
}

type invitationStub struct {
	ID           int
	Calendar     calendarStub
	EmailAddress string
}

type copyStub struct {
	ID          int
	Source      calendarStub
	Destination calendarStub
}

type dashboard struct {
	IsAuthenticated bool
	Calendars       []calendarStub
	Invitations     []invitationStub
	Copies          []copyStub
}

func (v Views) Dashboard(c echo.Context) error {
	ctx := c.Request().Context()
	tokens, err := v.ctr.Database.GetTokens(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Render(200, "index.html", dashboard{IsAuthenticated: false})
		}
		return errors.Wrap(err, "failed to get tokens")
	}

	client, err := clients.GetClient(ctx, v.ctr.OAuth2Config, tokens)
	if err != nil {
		return errors.Wrap(err, "failed to create client")
	}

	var calendarStubs []calendarStub
	calendarStubsById := make(map[string]calendarStub)
	if err = client.CalendarList.List().Pages(ctx, func(list *calendar.CalendarList) error {
		for _, c := range list.Items {
			stub := calendarStub{
				AccessRole: c.AccessRole,
				ID:         c.Id,
				Label:      c.Summary,
			}
			calendarStubsById[c.Id] = stub
			calendarStubs = append(calendarStubs, stub)
		}

		return nil
	}); err != nil {
		return errors.Wrap(err, "failed to list calendars")
	}

	var inviteStubs []invitationStub
	invites, err := v.ctr.Database.GetInviteConfigs(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to collect invites")
	}
	for _, i := range invites {
		inviteStubs = append(inviteStubs, invitationStub{
			ID:           i.ID,
			Calendar:     calendarStubsById[i.CalendarID],
			EmailAddress: i.EmailAddress,
		})
	}

	var copyStubs []copyStub
	copies, err := v.ctr.Database.GetCopyConfigs(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to collect copies")
	}
	for _, cs := range copies {
		copyStubs = append(copyStubs, copyStub{
			ID:          cs.ID,
			Source:      calendarStubsById[cs.SourceID],
			Destination: calendarStubsById[cs.DestinationID],
		})
	}

	model := dashboard{
		Calendars:       calendarStubs,
		Copies:          copyStubs,
		Invitations:     inviteStubs,
		IsAuthenticated: true,
	}

	return c.Render(200, "index.html", model)
}
