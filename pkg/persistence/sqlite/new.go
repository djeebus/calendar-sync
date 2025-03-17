package sqlite

import (
	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/simukti/sqldb-logger"

	"calendar-sync/pkg"
	"calendar-sync/pkg/logs"
)

const dbVersionSetting SettingType = "db_version"

type databaseLogger struct {
}

func (d databaseLogger) Log(ctx context.Context, level sqldblogger.Level, msg string, data map[string]interface{}) {
	log := logs.GetLogger(ctx)

	var lvl zerolog.Level

	switch level {
	case sqldblogger.LevelError:
		lvl = zerolog.ErrorLevel
	case sqldblogger.LevelInfo:
		lvl = zerolog.InfoLevel
	case sqldblogger.LevelDebug:
		lvl = zerolog.DebugLevel
	case sqldblogger.LevelTrace:
		lvl = zerolog.TraceLevel
	default:
		lvl = zerolog.DebugLevel
	}

	log.WithLevel(lvl).Fields(data).Msg(msg)

}

func NewDatabase(ctx context.Context, cfg pkg.Config) (*Database, error) {
	conn, err := sql.Open(cfg.DatabaseDriver, cfg.DatabaseSource)
	if err != nil {
		return nil, errors.Wrap(err, "failed to epen db")
	}

	conn = sqldblogger.OpenDriver(
		cfg.DatabaseSource,
		conn.Driver(),
		databaseLogger{},
		sqldblogger.WithPreparerLevel(sqldblogger.LevelDebug),
		sqldblogger.WithQueryerLevel(sqldblogger.LevelDebug),
		sqldblogger.WithExecerLevel(sqldblogger.LevelInfo),
	)

	db := &Database{db: conn}
	if err := migrate(ctx, db, conn); err != nil {
		return nil, errors.Wrap(err, "failed to migrate")
	}

	return db, nil
}

type Database struct {
	db *sql.DB
}

func (d *Database) Close() {
	d.db.Close()
}
