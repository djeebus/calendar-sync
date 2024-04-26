package pkg

import "google.golang.org/api/calendar/v3"

const ListenPort = 31425

type CalendarItem struct {
	CalendarID     string
	CalendarItemID string

	Summary string
	Guests  []*calendar.EventAttendee
}

type Calendar struct {
	CalendarID string
	Summary    string
	Items      []CalendarItem
}
