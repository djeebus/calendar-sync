package activities

import (
	"context"

	"calendar-sync/pkg/persistence"
)

type GetAllCopyConfigsResult struct {
	CopyConfigs []persistence.CopyConfig
}

func (a Activities) GetAllCopies(ctx context.Context) (GetAllCopyConfigsResult, error) {
	copies, err := a.ctr.Database.GetCopyConfigs(ctx)
	return GetAllCopyConfigsResult{CopyConfigs: copies}, err
}
