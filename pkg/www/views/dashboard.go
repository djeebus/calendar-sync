package views

import (
	"database/sql"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"google.golang.org/api/calendar/v3"

	"calendar-sync/pkg/logs"
	"calendar-sync/pkg/www/templates"
)

func (v Views) Dashboard(c echo.Context) error {
	ctx := c.Request().Context()

	client, err := v.ctr.GetCalendarClient(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Redirect(302, "/auth/begin")
		}

		return errors.Wrap(err, "failed to create client")
	}

	var calendarStubs []templates.CalendarStub
	calendarStubsById := make(map[string]templates.CalendarStub)
	if err = client.CalendarList.List().Pages(ctx, func(list *calendar.CalendarList) error {
		for _, c := range list.Items {
			stub := templates.CalendarStub{
				AccessRole: c.AccessRole,
				ID:         c.Id,
				Label:      c.Summary,
			}
			calendarStubsById[c.Id] = stub
			calendarStubs = append(calendarStubs, stub)
		}

		return nil
	}); err != nil {
		if strings.Contains(err.Error(), "oauth2: token expired and refresh token is not set") {
			if err = v.ctr.Database.RemoveTokens(ctx); err != nil {
				log := logs.GetLogger(ctx)
				log.Warn().Err(err).Msg("failed to remove tokens")
			}
			return c.Redirect(302, "/auth/begin")
		}
		return errors.Wrap(err, "failed to list calendars")
	}

	var inviteStubs []templates.InvitationStub
	invites, err := v.ctr.Database.GetInviteConfigs(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to collect invites")
	}
	for _, i := range invites {
		inviteStubs = append(inviteStubs, templates.InvitationStub{
			ID:           i.ID,
			Calendar:     calendarStubsById[i.CalendarID],
			EmailAddress: i.EmailAddress,
		})
	}

	var copyStubs []templates.CopyStub
	copies, err := v.ctr.Database.GetCopyConfigs(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to collect copies")
	}
	for _, cs := range copies {
		copyStubs = append(copyStubs, templates.CopyStub{
			ID:          cs.ID,
			Source:      calendarStubsById[cs.SourceID],
			Destination: calendarStubsById[cs.DestinationID],
		})
	}

	tokens, err := v.ctr.Database.GetTokens(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to collect tokens")
	}

	model := templates.Dashboard{
		AuthDuration:    time.Until(tokens.Expiry).String(),
		AuthExpiration:  tokens.Expiry.String(),
		Calendars:       calendarStubs,
		Copies:          copyStubs,
		Invitations:     inviteStubs,
		IsAuthenticated: true,
	}

	return c.Render(200, "index.html", model)
}
