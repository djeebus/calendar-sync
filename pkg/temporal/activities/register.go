package activities

import "go.temporal.io/sdk/worker"

func Register(w worker.Worker) {
	w.RegisterActivity(GetCalendarEventsActivity)
	w.RegisterActivity(UpdateGuestList)
}
