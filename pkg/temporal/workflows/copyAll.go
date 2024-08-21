package workflows

import (
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"calendar-sync/pkg/logs"
	"calendar-sync/pkg/persistence"
	"calendar-sync/pkg/temporal/activities"
)

func CopyAllWorkflow(ctx workflow.Context) error {
	log := logs.GetWorkflowLogger(ctx)

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

	ctx = workflow.WithActivityOptions(ctx, options)

	copyConfigs, err := getCopies(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get copies")
	}

	for _, copyConfig := range copyConfigs {
		args := CopyCalendarWorkflowArgs{
			copyConfig.SourceID,
			copyConfig.DestinationID,
		}
		if err := workflow.ExecuteChildWorkflow(ctx, CopyCalendarWorkflow, args).Get(ctx, nil); err != nil {
			log.Error().Err(err).Msg("failed to trigger child workflow")
		}
	}

	return nil
}

func getCopies(ctx workflow.Context) ([]persistence.CopyConfig, error) {
	var a activities.Activities

	var result activities.GetAllCopyConfigsResult
	if err := workflow.ExecuteActivity(ctx, a.GetAllCopies).Get(ctx, &result); err != nil {
		return nil, errors.Wrap(err, "failed to get all copies")
	}

	return result.CopyConfigs, nil
}
