package activities

import (
	"context"

	"calendar-sync/pkg/persistence"
)

type GetAllInviteConfigsResult struct {
	InviteConfigs []persistence.InviteConfig
}

func (a Activities) GetAllInvites(ctx context.Context) (GetAllInviteConfigsResult, error) {
	ctx = setupLogger(ctx, "GetAllInvites")

	invites, err := a.ctr.Database.GetInviteConfigs(ctx)
	return GetAllInviteConfigsResult{InviteConfigs: invites}, err
}
