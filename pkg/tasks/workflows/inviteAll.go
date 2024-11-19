package workflows

import (
	"context"

	"github.com/pkg/errors"
)

func (w *Workflows) InviteAllWorkflow(ctx context.Context) error {
	ctx, log := setupLogger(ctx, "InviteAllWorkflow")

	inviteConfigs, err := w.a.GetAllInvites(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get invites")
	}

	for _, inviteConfig := range inviteConfigs.InviteConfigs {
		args := InviteCalendarWorkflowArgs{
			inviteConfig.CalendarID,
			inviteConfig.EmailAddress,
		}
		if err := w.InviteCalendarWorkflow(ctx, args); err != nil {
			log.Error().Err(err).Msg("failed to trigger child workflow")
		}
	}

	return nil
}
