package workflows

import (
	"calendar-sync/pkg/logs"
	"calendar-sync/pkg/persistence"
	"calendar-sync/pkg/temporal/activities"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"time"
)

type ProcessWebhookEventArgs struct {
	ChannelID     string
	MessageNumber string
	ResourceID    string
	ResourceState string
	ResourceUri   string
	ChannelToken  string
}

func ProcessWebhookEvent(ctx workflow.Context, args ProcessWebhookEventArgs) error {
	log := logs.GetWorkflowLogger(ctx)

	// channel has been created, this doesn't really represent an event
	if args.ResourceState == "sync" {
		return nil
	}

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

	var a activities.Activities

	var getWatchResult activities.GetWatchResult
	getWatchArgs := activities.GetWatchArgs{WatchID: args.ChannelID}
	if err := workflow.ExecuteActivity(ctx, a.GetWatch, getWatchArgs).Get(ctx, &getWatchResult); err != nil {
		return errors.Wrap(err, "failed to get watch")
	}
	watch := getWatchResult.Watch
	if watch.Token != args.ChannelToken {
		log.Warn().
			Str("watch_id", args.ChannelID).
			Str("bad_token", args.ChannelToken).
			Str("real_token", watch.Token).
			Msg("token mismatch")
		return nil
	}

	switch args.ResourceState {
	case "exists":
		// resource was created or modified, do something!
		return processEventUpsert(ctx, args, watch)
	case "not_exists":
		return processEventDelete(ctx, args, watch)
	default:
		log.Warn().
			Str("state", args.ResourceState).
			Str("uri", args.ResourceUri).
			Msg("unknown resource state")
		return nil
	}
}

func processEventUpsert(ctx workflow.Context, args ProcessWebhookEventArgs, watch persistence.WatchConfig) error {
	var a activities.Activities

	activityArgs := activities.UpsertCalendarEntryArgs{CalendarID: watch.CalendarID, EventID: args.ResourceID}
	var result activities.UpsertCalendarEntryResult
	if err := workflow.ExecuteActivity(ctx, a.UpsertCalendarEntry, activityArgs).Get(ctx, &result); err != nil {
		return errors.Wrap(err, "failed to upsert calendar entry")
	}

	return nil
}

func processEventDelete(ctx workflow.Context, args ProcessWebhookEventArgs, watch persistence.WatchConfig) error {
	// resource was deleted
	var a activities.Activities

	activityArgs := activities.RemoveCalendarEntryArgs{CalendarID: watch.CalendarID, EventID: args.ResourceID}
	var result activities.RemoveCalendarEntryResult
	if err := workflow.ExecuteActivity(ctx, a.RemoveCalendarEntry, activityArgs).Get(ctx, &result); err != nil {
		return errors.Wrap(err, "failed to remove calendar entry")
	}

	return nil
}
