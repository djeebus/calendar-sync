package sqlite

import (
	"calendar-sync/pkg"
	"context"
	"database/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestDatabase(t *testing.T) {
	ctx := context.Background()
	cfg := pkg.Config{DatabaseDriver: "sqlite3", DatabaseSource: "test.db"}

	db, err := NewDatabase(ctx, cfg)
	require.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
		os.Remove(cfg.DatabaseSource)
	})

	// settings
	var testKey settingType = "some-setting-key"

	val, err := db.getSetting(ctx, testKey)
	require.ErrorIs(t, err, sql.ErrNoRows)
	assert.Equal(t, "", val)

	err = db.setSetting(ctx, testKey, "some-value")
	require.NoError(t, err)

	val, err = db.getSetting(ctx, testKey)
	require.NoError(t, err)
	assert.Equal(t, "some-value", val)

	// state
	_, err = db.GetState(ctx)
	require.ErrorIs(t, err, sql.ErrNoRows)

	err = db.SetState(ctx, "some-state")
	require.NoError(t, err)

	state, err := db.GetState(ctx)
	assert.Equal(t, "some-state", state)

	expiration := time.Now().UTC()

	err = db.CreateWatchConfig(ctx, "calendar-id", "watch-id", "token", expiration)
	require.NoError(t, err)

	w, err := db.GetWatchConfig(ctx, "watch-id")
	require.NoError(t, err)
	assert.Equal(t, "watch-id", w.WatchID)
	assert.Equal(t, "token", w.Token)
	assert.Equal(t, expiration, w.Expiration)
	assert.Equal(t, "calendar-id", w.CalendarID)

	err = db.DeleteWatchConfig(ctx, w.ID)
	require.NoError(t, err)

	_, err = db.GetWatchConfig(ctx, "watch-id")
	require.ErrorIs(t, err, sql.ErrNoRows)
}
