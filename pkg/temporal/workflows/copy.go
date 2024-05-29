package workflows

import (
	"calendar-sync/pkg"
	"calendar-sync/pkg/temporal/activities"
	"github.com/rs/zerolog/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/api/calendar/v3"
	"time"
)

type CopyCalendarWorkflowArgs struct {
	SourceCalendarID      string
	DestinationCalendarID string
}

func CopyCalendarWorkflow(ctx workflow.Context, args CopyCalendarWorkflowArgs) error {
	// setup
	var a activities.Activities

	retryPolicy := temporal.RetryPolicy{
		InitialInterval:    1 * time.Minute,
		BackoffCoefficient: 2.0,

		MaximumAttempts:        1,
		MaximumInterval:        1 * time.Hour,
		NonRetryableErrorTypes: []string{},
	}

	options := workflow.ActivityOptions{
		StartToCloseTimeout: 15 * time.Minute,
		RetryPolicy:         &retryPolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, options)

	// get source events
	sourceCalendarItems, err := getEvents(ctx, args.SourceCalendarID)
	if err != nil {
		return err
	}
	sourceItemsByID := pkg.ToMap(sourceCalendarItems, func(item *calendar.Event) string { return item.Id })

	// get destination events
	destinationCalendarItems, err := getEvents(ctx, args.DestinationCalendarID)
	if err != nil {
		return err
	}
	destinationCalendarItems = pkg.Filter(destinationCalendarItems, func(item *calendar.Event) bool {
		return getExtraByKey(item, pkg.SourceCalendarIDKey) == args.SourceCalendarID
	})
	destinationItemsBySourceItemID := pkg.ToMap(destinationCalendarItems, func(item *calendar.Event) string { return getExtraByKey(item, pkg.SourceCalendarItemIDKey) })

	// find missing destination events
	var futures []workflow.Future
	for key, sourceItem := range sourceItemsByID {
		if destItem, ok := destinationItemsBySourceItemID[key]; ok {
			if patch := buildPatch(*sourceItem, *destItem); patch != nil {
				updateArgs := activities.UpdateCalendarItemArgs{
					CalendarID:     args.DestinationCalendarID,
					CalendarItemID: destItem.Id,
					Patch:          patch,
				}
				f := workflow.ExecuteActivity(ctx, a.UpdateCalendarItem, updateArgs)
				futures = append(futures, f)
			}
			continue
		}

		createArgs := activities.CreateCalendarItemArgs{
			Event:      toInsert(args.SourceCalendarID, sourceItem),
			CalendarID: args.DestinationCalendarID,
		}
		f := workflow.ExecuteActivity(ctx, a.CreateCalendarItem, createArgs)
		futures = append(futures, f)
	}

	// find extra destination events
	for key, destItem := range destinationItemsBySourceItemID {
		if _, ok := sourceItemsByID[key]; ok {
			continue
		}

		removeArgs := activities.RemoveCalendarItemArgs{
			CalendarID: args.DestinationCalendarID,
			EventID:    destItem.Id,
		}
		f := workflow.ExecuteActivity(ctx, a.RemoveCalendarItem, removeArgs)
		futures = append(futures, f)
	}

	// collect futures
	for _, f := range futures {
		if err = f.Get(ctx, nil); err != nil {
			log.Err(err).Msg("some task did not execute correctly")
		}
	}

	return nil
}

func toInsert(sourceCalendarID string, e *calendar.Event) *calendar.Event {
	return &calendar.Event{
		Description: e.Description,
		End:         e.End,
		EventType:   e.EventType,
		ExtendedProperties: &calendar.EventExtendedProperties{
			Private: map[string]string{
				pkg.SourceCalendarIDKey:     sourceCalendarID,
				pkg.SourceCalendarItemIDKey: e.Id,
			},
		},
		Kind:     e.Kind,
		Location: e.Location,
		Start:    e.Start,
		Status:   e.Status,
		Summary:  e.Summary,
	}

}

func buildPatch(from, to calendar.Event) *calendar.Event {
	patch := new(calendar.Event)
	shouldPatch := false

	// perform some clean up
	if from.Summary == "" {
		from.Summary = "Busy"
	}

	if from.EventType == "" {
		from.EventType = "default"
	}

	// diff
	if from.EventType != to.EventType {
		patch.EventType = from.EventType
		shouldPatch = true
	}
	if from.Location != to.Location {
		patch.Location = from.Location
		shouldPatch = true
	}
	if from.Status != to.Status {
		patch.Status = from.Status
		shouldPatch = true
	}
	if from.Summary != to.Summary {
		patch.Summary = from.Summary
		shouldPatch = true
	}
	if from.Description != to.Description {
		patch.Description = from.Description
		shouldPatch = true
	}
	if len(from.Recurrence) != len(to.Recurrence) {
		patch.Recurrence = from.Recurrence
		shouldPatch = true
	} else {
		for idx := range from.Recurrence {
			if from.Recurrence[idx] != to.Recurrence[idx] {
				patch.Recurrence = from.Recurrence
				shouldPatch = true
				break
			}
		}
	}
	if from.Start != nil && (to.Start == nil || from.Start.DateTime != to.Start.DateTime) {
		patch.Start = &calendar.EventDateTime{DateTime: from.Start.DateTime}
		shouldPatch = true
	}
	if from.End != nil && (to.End == nil || from.End.DateTime != to.End.DateTime) {
		patch.End = &calendar.EventDateTime{DateTime: to.End.DateTime}
		shouldPatch = true
	}

	if !shouldPatch {
		return nil
	}

	return patch
}

func getExtraByKey(item *calendar.Event, key string) string {
	if item.ExtendedProperties == nil {
		return ""
	}

	if item.ExtendedProperties.Private == nil {
		return ""
	}

	return item.ExtendedProperties.Private[key]
}

func getEvents(ctx workflow.Context, sourceID string) ([]*calendar.Event, error) {
	var a activities.Activities
	var sourceEventsResult activities.GetCalendarEventsActivityResult
	sourceEventsArgs := activities.GetCalendarEventsActivityArgs{
		CalendarID: sourceID,
	}
	if err := workflow.ExecuteActivity(ctx, a.GetCalendarEventsActivity, sourceEventsArgs).Get(ctx, &sourceEventsResult); err != nil {
		return nil, err
	}

	return sourceEventsResult.Calendar.Items, nil
}
