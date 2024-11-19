package activities

import (
	"context"

	"calendar-sync/pkg/container"
	"calendar-sync/pkg/logs"
)

type Activities struct {
	ctr container.Container
}

func New(ctr container.Container) *Activities {
	a := Activities{ctr}
	return &a
}

func setupLogger(ctx context.Context, activityName string) context.Context {
	log := logs.GetLogger(ctx).With().Str("activity-name", activityName).Logger()
	ctx = logs.SetLogger(ctx, log)
	return ctx
}
