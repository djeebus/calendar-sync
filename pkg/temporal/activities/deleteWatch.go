package activities

import (
	"context"
	"github.com/pkg/errors"
)

type DeleteWatchConfigArgs struct {
	WatchID int
}

type DeleteWatchConfigResult struct {
}

func (a Activities) DeleteWatchConfig(ctx context.Context, args DeleteWatchConfigArgs) (DeleteWatchConfigResult, error) {
	err := a.ctr.Database.DeleteWatchConfig(ctx, args.WatchID)
	if err != nil {
		return DeleteWatchConfigResult{}, errors.Wrap(err, "failed to delete watch config")
	}

	return DeleteWatchConfigResult{}, nil
}
