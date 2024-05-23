package activities

import "context"

type RemoveCalendarEntryArgs struct {
	CalendarID string
	EventID    string
}

type RemoveCalendarEntryResult struct{}

func (a Activities) RemoveCalendarEntry(ctx context.Context, args RemoveCalendarEntryArgs) (RemoveCalendarEntryResult, error) {
	panic("implement me")
}
