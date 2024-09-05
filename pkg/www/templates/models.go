package templates

type CalendarStub struct {
	ID         string
	Label      string
	AccessRole string
}

type InvitationStub struct {
	ID           int
	Calendar     CalendarStub
	EmailAddress string
}

type CopyStub struct {
	ID          int
	Source      CalendarStub
	Destination CalendarStub
}

type Dashboard struct {
	IsAuthenticated bool
	AuthExpiration  string
	AuthDuration    string
	Calendars       []CalendarStub
	Invitations     []InvitationStub
	Copies          []CopyStub
}
