package persistence

import "time"

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

type WatchConfig struct {
	ID         int
	CalendarID string
	WatchID    string
	Token      string
	Expiration time.Time
}
