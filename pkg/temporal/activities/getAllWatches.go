package activities

import (
	"context"

	"calendar-sync/pkg/persistence"
)

type GetAllWatchConfigsResult struct {
	WatchConfigs []persistence.WatchConfig
}

func (a Activities) GetAllWatches(ctx context.Context) (GetAllWatchConfigsResult, error) {
	watches, err := a.ctr.Database.GetWatchConfigs(ctx)
	return GetAllWatchConfigsResult{WatchConfigs: watches}, err
}
