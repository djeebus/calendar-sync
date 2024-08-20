package tracing

import "github.com/rs/zerolog"

func ZerologHook() zerolog.Hook {
	return zerologHook{}
}

type zerologHook struct{}

func (l zerologHook) Run(e *zerolog.Event, level zerolog.Level, message string) {
	ctx := e.GetCtx()
	if ctx == nil {
		return
	}

	correlationID, ok := GetCorrelationID(ctx)
	if !ok {
		return
	}

	e.Str("correlation-id", correlationID)
}
