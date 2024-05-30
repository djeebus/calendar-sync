package workflows

import (
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/calendar/v3"
	"testing"
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
			actual := buildPatch(tc.from, tc.to)
			if tc.expected == nil {
				assert.Nil(t, actual)
			} else {
				assert.Equal(t, *tc.expected, *actual)
			}
		})
	}
}
