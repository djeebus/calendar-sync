package workflows

import (
	"calendar-sync/pkg"
	"calendar-sync/pkg/logs"
	"calendar-sync/pkg/temporal/activities"
	"github.com/rs/zerolog"
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
	log := logs.GetWorkflowLogger(ctx)

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
			if patch := buildPatch(log, *sourceItem, *destItem); patch != nil {
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
	event := calendar.Event{
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

	cleanEvent(&event)

	return &event
}

func cleanEvent(e *calendar.Event) {
	if e.Summary == "" {
		e.Summary = "Busy"
	}

	if e.EventType == "" {
		e.EventType = "default"
	}
}

type patchable[P any] struct {
	log             zerolog.Logger
	from, to, patch *P
	shouldPatch     bool
}

func diffSlice[P any, T comparable](p *patchable[P], field string, fn func(e *P) *[]T) {
	from := *fn(p.from)
	to := *fn(p.to)

	var shouldPatch bool
	if len(from) != len(to) {
		p.log.Info().
			Any("source", from).
			Any("destination", to).
			Msgf("%s has different length", field)

		shouldPatch = true
	} else {
		for idx := range from {
			if from[idx] != to[idx] {
				p.log.Info().
					Any("source", from[idx]).
					Any("destination", to[idx]).
					Msgf("%s is different at index #%d", field, idx)

				shouldPatch = true
				break
			}
		}
	}

	if !shouldPatch {
		return
	}

	patch := fn(p.patch)
	*patch = from
	p.shouldPatch = true
}

func diff[P any, T comparable](p *patchable[P], field string, fn func(e *P) *T) {
	from := *fn(p.from)
	to := *fn(p.to)

	if from == to {
		return
	}

	p.log.Info().
		Any("source", from).
		Any("destination", to).
		Msgf("%s is different", field)

	patch := fn(p.patch)
	*patch = from
	p.shouldPatch = true
}

func buildPatch(log *zerolog.Logger, from, to calendar.Event) *calendar.Event {
	logger := log.With().
		Str("source_event_id", from.Id).
		Str("destination_event_id", to.Id).
		Logger()

	var patch calendar.Event

	// perform some clean up
	cleanEvent(&from)
	cleanEvent(&to)

	p := &patchable[calendar.Event]{
		log:   logger,
		from:  &from,
		to:    &to,
		patch: &patch,
	}

	// diff
	diff(p, "event_type", func(e *calendar.Event) *string { return &e.EventType })
	diff(p, "location", func(e *calendar.Event) *string { return &e.Location })
	diff(p, "status", func(e *calendar.Event) *string { return &e.Status })
	diff(p, "summary", func(e *calendar.Event) *string { return &e.Summary })
	diff(p, "description", func(e *calendar.Event) *string { return &e.Description })
	diffSlice(p, "recurrence", func(e *calendar.Event) *[]string { return &e.Recurrence })

	if update := patchDateTime(logger, "start", from.Start, to.Start); update != nil {
		patch.Start = update
		p.shouldPatch = true
	}
	if update := patchDateTime(logger, "end", from.End, to.End); update != nil {
		patch.End = update
		p.shouldPatch = true
	}

	if !p.shouldPatch {
		return nil
	}

	return &patch
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

func patchDateTime(log zerolog.Logger, field string, from *calendar.EventDateTime, to *calendar.EventDateTime) *calendar.EventDateTime {
	if from == nil {
		return nil
	}

	if to == nil {
		return from
	}

	patch := calendar.EventDateTime{}
	p := &patchable[calendar.EventDateTime]{
		log:   log.With().Str("datetime field", field).Logger(),
		from:  from,
		to:    to,
		patch: &patch,
	}

	diff(p, "date", func(dt *calendar.EventDateTime) *string { return &dt.Date })
	diff(p, "datetime", func(dt *calendar.EventDateTime) *string { return &dt.DateTime })
	diff(p, "timezone", func(dt *calendar.EventDateTime) *string { return &dt.TimeZone })

	if !p.shouldPatch {
		return nil
	}

	return &patch
}
