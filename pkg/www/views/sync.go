package views

import (
	"net/url"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"calendar-sync/pkg/tasks/workflows"
)

func (v Views) SyncCopy(c echo.Context, vals url.Values) error {
	ctx := c.Request().Context()

	copyIDstr := vals.Get("copyID")
	if copyIDstr == "" {
		return errors.New("required field missing")
	}
	copyID, err := strconv.ParseInt(copyIDstr, 10, 64)
	if err != nil {
		return errors.Wrap(err, "failed to parse copyID")
	}

	config, err := v.ctr.Database.GetCopyConfig(ctx, copyID)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve copy row")
	}

	args := workflows.CopyCalendarWorkflowArgs{
		SourceCalendarID:      config.SourceID,
		DestinationCalendarID: config.DestinationID,
	}
	if err := v.workflows.CopyCalendarWorkflow(ctx, args); err != nil {
		return errors.Wrap(err, "failed to execute workflow")
	}

	return c.Redirect(302, "/")
}

func (v Views) SyncInvite(c echo.Context, vals url.Values) error {
	ctx := c.Request().Context()

	inviteIDstr := vals.Get("inviteID")
	if inviteIDstr == "" {
		return errors.New("required field missing")
	}
	inviteID, err := strconv.ParseInt(inviteIDstr, 10, 64)
	if err != nil {
		return errors.Wrap(err, "failed to parse inviteID")
	}

	config, err := v.ctr.Database.GetInviteConfig(ctx, inviteID)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve invite row")
	}

	args := workflows.InviteCalendarWorkflowArgs{
		CalendarID: config.CalendarID,
		EmailToAdd: config.EmailAddress,
	}
	if err := v.workflows.InviteCalendarWorkflow(ctx, args); err != nil {
		return errors.Wrap(err, "failed to execute workflow")
	}

	return c.Redirect(302, "/")
}
