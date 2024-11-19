package workflows

import (
	"context"

	"github.com/rs/zerolog"

	"calendar-sync/pkg/logs"
	"calendar-sync/pkg/tasks/activities"
)

type Workflows struct {
	a *activities.Activities
}

func New(a *activities.Activities) *Workflows {
	return &Workflows{a: a}
}

func setupLogger(ctx context.Context, workflowName string) (context.Context, zerolog.Logger) {
	log := logs.GetLogger(ctx).With().Str("workflow-name", workflowName).Logger()
	ctx = logs.SetLogger(ctx, log)
	return ctx, log
}
