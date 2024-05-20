package temporal

import (
	"calendar-sync/pkg/container"
	"calendar-sync/pkg/temporal/activities"
	"calendar-sync/pkg/temporal/workflows"
	"context"
	"go.temporal.io/sdk/worker"
)

func NewWorker(ctx context.Context, ctr container.Container) (worker.Worker, error) {
	opts := worker.Options{BackgroundActivityContext: ctx}
	w := worker.New(ctr.TemporalClient, ctr.TaskQueue, opts)

	workflows.Register(w)
	activities.Register(w, ctr)

	return w, nil
}
