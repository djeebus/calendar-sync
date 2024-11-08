package sqlite

import "context"

var stateSetting settingType = "state"

func (d *Database) GetState(ctx context.Context) (string, error) {
	return d.getSetting(ctx, stateSetting)
}

func (d *Database) SetState(ctx context.Context, state string) error {
	return d.setSetting(ctx, stateSetting, state)
}
