package activities

import (
	"calendar-sync/pkg/persistence"
	"context"
)

type GetAllWatchConfigsResult struct {
	WatchConfigs []persistence.WatchConfig
}

func (a Activities) GetAllWatches(ctx context.Context) (GetAllWatchConfigsResult, error) {
	watches, err := a.ctr.Database.GetWatchConfigs(ctx)
	return GetAllWatchConfigsResult{WatchConfigs: watches}, err
}
