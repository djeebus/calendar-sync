package workflows

import "go.temporal.io/sdk/worker"

func Register(w worker.Worker) {
	w.RegisterWorkflow(InviteAllWorkflow)
	w.RegisterWorkflow(InviteCalendarWorkflow)
	w.RegisterWorkflow(CopyAllWorkflow)
	w.RegisterWorkflow(CopyCalendarWorkflow)
}
