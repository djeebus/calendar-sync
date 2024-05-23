package pkg

import "google.golang.org/api/calendar/v3"

type Calendar struct {
	CalendarID string
	Summary    string
	Items      []*calendar.Event
}

const SourceCalendarIDKey = "source-calendar-id"
const SourceCalendarItemIDKey = "source-calendar-item-id"
