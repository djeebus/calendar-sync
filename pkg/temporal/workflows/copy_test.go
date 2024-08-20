package workflows

import (
	"calendar-sync/pkg"
	"calendar-sync/pkg/temporal/activities"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
	"google.golang.org/api/calendar/v3"
	"os"
	"testing"
	"time"
)

func TestBuildPatch(t *testing.T) {
	testcases := map[string]struct {
		from, to calendar.Event
		expected *calendar.Event
	}{
		"minimal": {
			from:     calendar.Event{},
			to:       calendar.Event{},
			expected: nil,
		},
		"no diff": {
			from: calendar.Event{
				EventType:   "blah",
				Location:    "location",
				Status:      "status",
				Summary:     "an event",
				Description: "description",
				Recurrence:  []string{"one", "two"},
				Start: &calendar.EventDateTime{
					Date:     "start-date",
					DateTime: "start-datetime",
					TimeZone: "start-timezone",
				},
				End: &calendar.EventDateTime{
					Date:     "end-date",
					DateTime: "end-datetime",
					TimeZone: "end-timezone",
				},
			},
			to: calendar.Event{
				EventType:   "blah",
				Location:    "location",
				Status:      "status",
				Summary:     "an event",
				Description: "description",
				Recurrence:  []string{"one", "two"},
				Start: &calendar.EventDateTime{
					Date:     "start-date",
					DateTime: "start-datetime",
					TimeZone: "start-timezone",
				},
				End: &calendar.EventDateTime{
					Date:     "end-date",
					DateTime: "end-datetime",
					TimeZone: "end-timezone",
				},
			},
		},
		"everything is changed": {
			from: calendar.Event{
				EventType:   "blah",
				Location:    "location",
				Status:      "status",
				Summary:     "an event",
				Description: "description",
				Recurrence:  []string{"one", "two"},
				Start: &calendar.EventDateTime{
					Date:     "start-date",
					DateTime: "start-datetime",
					TimeZone: "start-timezone",
				},
				End: &calendar.EventDateTime{
					Date:     "end-date",
					DateTime: "end-datetime",
					TimeZone: "end-timezone",
				},
			},
			to: calendar.Event{
				EventType:   "blah2",
				Location:    "location2",
				Status:      "status2",
				Summary:     "an event2",
				Description: "description2",
				Recurrence:  []string{"one2", "two2"},
				Start: &calendar.EventDateTime{
					Date:     "start-date2",
					DateTime: "start-datetime2",
					TimeZone: "start-timezone2",
				},
				End: &calendar.EventDateTime{
					Date:     "end-date2",
					DateTime: "end-datetime2",
					TimeZone: "end-timezone2",
				},
			},
			expected: &calendar.Event{
				EventType:   "blah",
				Location:    "location",
				Status:      "status",
				Summary:     "an event",
				Description: "description",
				Recurrence:  []string{"one", "two"},
				Start: &calendar.EventDateTime{
					Date:     "start-date",
					DateTime: "start-datetime",
					TimeZone: "start-timezone",
				},
				End: &calendar.EventDateTime{
					Date:     "end-date",
					DateTime: "end-datetime",
					TimeZone: "end-timezone",
				},
			},
		},
	}

	for key, tc := range testcases {
		t.Run(key, func(t *testing.T) {
			log := zerolog.New(os.Stdout)
			actual := buildPatch(&log, tc.from, tc.to)
			if tc.expected == nil {
				assert.Nil(t, actual)
			} else {
				assert.Equal(t, *tc.expected, *actual)
			}
		})
	}
}

func TestCopyCalendarWorkflow(t *testing.T) {
	var ts testsuite.WorkflowTestSuite
	env := ts.NewTestWorkflowEnvironment()

	// setup source calendar
	sourceCalendarCopiedEvent := &calendar.Event{
		Id:    "source-calendar-id-copied-event",
		Start: &calendar.EventDateTime{DateTime: time.Now().Format(time.RFC3339)},
		End:   &calendar.EventDateTime{DateTime: time.Now().Format(time.RFC3339)},
	}
	extraSourceCalendarEvent := &calendar.Event{
		Id:    "source-calendar-id-new-event",
		Start: &calendar.EventDateTime{DateTime: time.Now().Format(time.RFC3339)},
		End:   &calendar.EventDateTime{DateTime: time.Now().Format(time.RFC3339)},
	}
	sourceCalendar := pkg.Calendar{
		CalendarID: "source-calendar-id",
		Summary:    "some interesting calendar",
		Items: []*calendar.Event{
			sourceCalendarCopiedEvent,
			extraSourceCalendarEvent,
		},
	}

	// setup destination calendar
	destinationCalendarCopiedEvent := &calendar.Event{
		Id:    "destination-calendar-id-copied-event",
		Start: &calendar.EventDateTime{DateTime: time.Now().Format(time.RFC3339)},
		End:   &calendar.EventDateTime{DateTime: time.Now().Format(time.RFC3339)},
		ExtendedProperties: &calendar.EventExtendedProperties{
			Private: map[string]string{
				pkg.SourceCalendarIDKey:     sourceCalendar.CalendarID,
				pkg.SourceCalendarItemIDKey: sourceCalendarCopiedEvent.Id,
			},
		},
	}
	destinationCalendarDuplicateCopiedEvent := &calendar.Event{
		Id:    "destination-calendar-id-duplicate-copied-event",
		Start: &calendar.EventDateTime{DateTime: time.Now().Format(time.RFC3339)},
		End:   &calendar.EventDateTime{DateTime: time.Now().Format(time.RFC3339)},
		ExtendedProperties: &calendar.EventExtendedProperties{
			Private: map[string]string{
				pkg.SourceCalendarIDKey:     sourceCalendar.CalendarID,
				pkg.SourceCalendarItemIDKey: sourceCalendarCopiedEvent.Id,
			},
		},
	}
	extraDestinationCalendarEvent := &calendar.Event{
		Id:    "destination-calendar-id-new-event",
		Start: &calendar.EventDateTime{DateTime: time.Now().Format(time.RFC3339)},
		End:   &calendar.EventDateTime{DateTime: time.Now().Format(time.RFC3339)},
	}
	destinationCalendar := pkg.Calendar{
		CalendarID: "destination-calendar-id",
		Summary:    "a calendar full of copies",
		Items: []*calendar.Event{
			destinationCalendarCopiedEvent,
			extraDestinationCalendarEvent,
			destinationCalendarDuplicateCopiedEvent,
		},
	}

	// set up workflows
	env.RegisterWorkflow(CopyCalendarWorkflow)

	// set up activities
	var a activities.Activities
	env.RegisterActivity(&a)
	env.
		OnActivity(
			a.GetCalendarEventsActivity,
			mock.Anything,
			activities.GetCalendarEventsActivityArgs{CalendarID: sourceCalendar.CalendarID},
		).
		Return(activities.GetCalendarEventsActivityResult{Calendar: sourceCalendar}, nil)

	env.
		OnActivity(
			a.GetCalendarEventsActivity,
			mock.Anything,
			activities.GetCalendarEventsActivityArgs{CalendarID: destinationCalendar.CalendarID},
		).
		Return(activities.GetCalendarEventsActivityResult{Calendar: destinationCalendar}, nil)
	env.
		OnActivity(
			a.CreateCalendarItem,
			mock.Anything,
			mock.MatchedBy(func(args activities.CreateCalendarItemArgs) bool {
				if args.CalendarID != destinationCalendar.CalendarID {
					return false
				}
				if args.Event == nil || args.Event.ExtendedProperties == nil || args.Event.ExtendedProperties.Private == nil {
					return false
				}

				private := args.Event.ExtendedProperties.Private

				if value, ok := private[pkg.SourceCalendarIDKey]; !ok {
					return false
				} else if value != sourceCalendar.CalendarID {
					return false
				}

				if value, ok := private[pkg.SourceCalendarItemIDKey]; !ok {
					return false
				} else if value != extraSourceCalendarEvent.Id {
					return false
				}

				return true
			}),
		).
		Return(activities.CreateCalendarItemResult{CreatedItem: &calendar.Event{}}, nil)

	// execute workflow
	args := CopyCalendarWorkflowArgs{
		SourceCalendarID:      sourceCalendar.CalendarID,
		DestinationCalendarID: destinationCalendar.CalendarID,
	}
	env.ExecuteWorkflow(CopyCalendarWorkflow, args)

	// verify result
	env.AssertExpectations(t)
}
