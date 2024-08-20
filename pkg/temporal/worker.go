package temporal

import (
	"calendar-sync/pkg/container"
	"calendar-sync/pkg/logs"
	"calendar-sync/pkg/temporal/activities"
	"calendar-sync/pkg/temporal/workflows"
	"context"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"
)

func NewWorker(ctx context.Context, ctr container.Container) (worker.Worker, error) {
	opts := worker.Options{
		BackgroundActivityContext: ctx,
		Interceptors: []interceptor.WorkerInterceptor{
			logs.NewLoggingInterceptor(),
		},
	}
	w := worker.New(ctr.TemporalClient, ctr.Config.TemporalTaskQueue, opts)

	workflows.Register(w)
	activities.Register(w, ctr)

	return w, nil
}
