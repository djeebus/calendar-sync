package workflows

import "go.temporal.io/sdk/worker"

func Register(w worker.Worker) {
	w.RegisterWorkflow(DiffCalendarWorkflow)
}
