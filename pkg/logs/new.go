package logs

import (
	"io"
	"os"

	"github.com/rs/zerolog"

	"calendar-sync/pkg"
	"calendar-sync/pkg/tracing"
)

func New(cfg pkg.Config) zerolog.Logger {
	var w io.Writer
	if !cfg.LogJson {
		w = zerolog.NewConsoleWriter()
	} else {
		w = os.Stdout
	}

	logger := zerolog.New(w)
	logger = logger.Level(cfg.LogLevel)

	logger = logger.Hook(tracing.ZerologHook())

	zerolog.DefaultContextLogger = &logger

	return logger
}
