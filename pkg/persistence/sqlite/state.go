package sqlite

import "context"

var stateSetting SettingType = "state"

func (d *Database) GetState(ctx context.Context) (string, error) {
	return d.GetSetting(ctx, stateSetting)
}

func (d *Database) SetState(ctx context.Context, state string) error {
	return d.SetSetting(ctx, stateSetting, state)
}
