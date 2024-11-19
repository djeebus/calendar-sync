package activities

import "context"

type UpsertCalendarEntryArgs struct {
	CalendarID string
	EventID    string
}

type UpsertCalendarEntryResult struct{}

func (a Activities) UpsertCalendarEntry(ctx context.Context, args UpsertCalendarEntryArgs) (UpsertCalendarEntryResult, error) {
	panic("implement me")
}
