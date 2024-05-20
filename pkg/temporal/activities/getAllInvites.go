package activities

import (
	"calendar-sync/pkg/persistence"
	"context"
)

type GetAllInviteConfigsResult struct {
	InviteConfigs []persistence.InviteConfig
}

func (a Activities) GetAllInvites(ctx context.Context) (GetAllInviteConfigsResult, error) {
	invites, err := a.ctr.Database.GetInviteConfigs(ctx)
	return GetAllInviteConfigsResult{InviteConfigs: invites}, err
}
