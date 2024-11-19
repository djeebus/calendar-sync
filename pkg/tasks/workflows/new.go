package workflows

import (
	"calendar-sync/pkg/container"
	"context"

	"github.com/rs/zerolog"

	"calendar-sync/pkg/logs"
	"calendar-sync/pkg/tasks/activities"
)

type Workflows struct {
	a   *activities.Activities
	ctr container.Container
}

func New(a *activities.Activities) *Workflows {
	return &Workflows{a: a}
}

func setupLogger(ctx context.Context, workflowName string) (context.Context, zerolog.Logger) {
	log := logs.GetLogger(ctx).With().Str("workflow-name", workflowName).Logger()
	ctx = logs.SetLogger(ctx, log)
	return ctx, log
}
