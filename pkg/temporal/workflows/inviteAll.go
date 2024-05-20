package workflows

import (
	"calendar-sync/pkg/persistence"
	"calendar-sync/pkg/temporal/activities"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"time"
)

func InviteAllWorkflow(ctx workflow.Context) error {
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

	inviteConfigs, err := getInvites(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get copies")
	}

	for _, inviteConfig := range inviteConfigs {
		args := InviteCalendarWorkflowArgs{
			inviteConfig.CalendarID,
			inviteConfig.EmailAddress,
		}
		if err := workflow.ExecuteChildWorkflow(ctx, InviteCalendarWorkflow, args).Get(ctx, nil); err != nil {
			log.Error().Err(err).Msg("failed to trigger child workflow")
		}
	}

	return nil
}

func getInvites(ctx workflow.Context) ([]persistence.InviteConfig, error) {
	var a activities.Activities

	var result activities.GetAllInviteConfigsResult
	if err := workflow.ExecuteActivity(ctx, a.GetAllInvites).Get(ctx, &result); err != nil {
		return nil, errors.Wrap(err, "failed to get all invites")
	}

	return result.InviteConfigs, nil
}
