package activities

import (
	"context"

	"calendar-sync/pkg/persistence"
)

type GetWatchArgs struct {
	WatchID string
}

type GetWatchResult struct {
	Watch persistence.WatchConfig
}

func (a Activities) GetWatch(ctx context.Context, args GetWatchArgs) (GetWatchResult, error) {
	watch, err := a.ctr.Database.GetWatchConfig(ctx, args.WatchID)
	return GetWatchResult{Watch: watch}, err
}
