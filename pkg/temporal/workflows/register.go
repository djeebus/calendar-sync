package workflows

import "go.temporal.io/sdk/worker"

func Register(w worker.Worker) {
	w.RegisterWorkflow(CopyAllWorkflow)
	w.RegisterWorkflow(CopyCalendarWorkflow)
	w.RegisterWorkflow(InviteAllWorkflow)
	w.RegisterWorkflow(InviteCalendarWorkflow)
	w.RegisterWorkflow(WatchAll)
	w.RegisterWorkflow(ProcessWebhookEvent)
}
