package persistence

type InviteConfig struct {
	ID           int
	CalendarID   string
	EmailAddress string
}

type CopyConfig struct {
	ID            int
	SourceID      string
	DestinationID string
}
