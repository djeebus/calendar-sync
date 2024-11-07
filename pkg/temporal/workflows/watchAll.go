package workflows

import (
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"calendar-sync/pkg"
	"calendar-sync/pkg/persistence"
	"calendar-sync/pkg/temporal/activities"
)

func WatchAll(ctx workflow.Context) error {
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

	var futures []workflow.Future

	watchConfigs, err := getWatches(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get watches")
	}

	goodWatches, badWatches := splitWatches(watchConfigs)

	for _, watch := range badWatches {
		futures = append(futures, deleteCalendar(ctx, watch.ID))
	}

	watchConfigsByCalendarID := pkg.ToSet(goodWatches, func(i persistence.WatchConfig) string {
		return i.CalendarID
	})

	inviteConfigs, err := getInvites(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get invites")
	}

	for _, inviteConfig := range inviteConfigs {
		futures = watchCalendar(ctx, watchConfigsByCalendarID, inviteConfig.CalendarID, futures)
	}

	copyConfigs, err := getCopies(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get copies")
	}

	for _, copyConfig := range copyConfigs {
		futures = watchCalendar(ctx, watchConfigsByCalendarID, copyConfig.SourceID, futures)
		futures = watchCalendar(ctx, watchConfigsByCalendarID, copyConfig.DestinationID, futures)
	}

	for _, f := range futures {
		var result activities.WatchCalendarResult
		if err := f.Get(ctx, &result); err != nil {
			return err
		}
	}

	return nil
}

func splitWatches(configs []persistence.WatchConfig) ([]persistence.WatchConfig, []persistence.WatchConfig) {
	var good, bad []persistence.WatchConfig

	for _, config := range configs {
		if !config.Expiration.IsZero() && config.Expiration.After(time.Now()) {
			good = append(good, config)
			continue
		}

		bad = append(bad, config)
	}

	return good, bad
}

func watchCalendar(
	ctx workflow.Context, existingWatches map[string]struct{}, calendarID string, futures []workflow.Future,
) []workflow.Future {
	var a activities.Activities

	_, ok := existingWatches[calendarID]
	if !ok {
		args := activities.WatchCalendarArgs{
			CalendarID: calendarID,
		}
		future := workflow.ExecuteActivity(ctx, a.WatchCalendar, args)
		futures = append(futures, future)
		existingWatches[calendarID] = struct{}{}
	}

	return futures
}

func getWatches(ctx workflow.Context) ([]persistence.WatchConfig, error) {
	var a activities.Activities

	var result activities.GetAllWatchConfigsResult
	if err := workflow.ExecuteActivity(ctx, a.GetAllWatches).Get(ctx, &result); err != nil {
		return nil, errors.Wrap(err, "failed to get all watches")
	}

	return result.WatchConfigs, nil
}

func deleteCalendar(ctx workflow.Context, watchID int) workflow.Future {
	var a activities.Activities

	args := activities.DeleteWatchConfigArgs{
		WatchID: watchID,
	}
	future := workflow.ExecuteActivity(ctx, a.DeleteWatchConfig, args)
	return future
}
