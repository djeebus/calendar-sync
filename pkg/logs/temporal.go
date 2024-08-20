package logs

import (
	"context"
	"github.com/rs/zerolog"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/log"
)

type temporalLogger struct {
	log zerolog.Logger
}

func (t temporalLogger) submitLog(level zerolog.Level, msg string, keyvals []interface{}) {
	logger := t.log.With().Logger()
	logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
		for i := 0; i < len(keyvals); i += 2 {
			key, ok := keyvals[i].(string)
			if !ok {
				l2 := c.Interface("invalid-key", keyvals[i]).Logger()
				l2.Warn().Msg("key is not a string")
				continue
			}

			switch val := keyvals[i+1].(type) {
			case error:
				c = c.AnErr(key, val)
			case int:
				c = c.Int(key, val)
			case int32:
				c = c.Int32(key, val)
			case float32:
				c = c.Float32(key, val)
			case float64:
				c = c.Float64(key, val)
			case bool:
				c = c.Bool(key, val)
			case string:
				c = c.Str(key, val)
			default:
				c = c.Interface(key, val)
			}
		}

		return c
	})

	logger.WithLevel(level).Msg(msg)
}

func (t temporalLogger) Debug(msg string, keyvals ...interface{}) {
	t.submitLog(zerolog.DebugLevel, msg, keyvals)
}

func (t temporalLogger) Info(msg string, keyvals ...interface{}) {
	t.submitLog(zerolog.InfoLevel, msg, keyvals)
}

func (t temporalLogger) Warn(msg string, keyvals ...interface{}) {
	t.submitLog(zerolog.WarnLevel, msg, keyvals)
}

func (t temporalLogger) Error(msg string, keyvals ...interface{}) {
	t.submitLog(zerolog.ErrorLevel, msg, keyvals)
}

func NewTemporalLogger(log zerolog.Logger) log.Logger {
	return &temporalLogger{log: log}
}

func NewLoggingInterceptor() interceptor.WorkerInterceptor {
	return loggingInterceptor{}
}

type loggingInterceptor struct {
	interceptor.WorkerInterceptor
}

func (i loggingInterceptor) InterceptActivity(ctx context.Context, next interceptor.ActivityInboundInterceptor) interceptor.ActivityInboundInterceptor {
	return loggingActivityInterceptor{next}
}

type loggingActivityInterceptor struct {
	interceptor.ActivityInboundInterceptor
}

func (i loggingActivityInterceptor) ExecuteActivity(ctx context.Context, in *interceptor.ExecuteActivityInput) (interface{}, error) {
	return i.ActivityInboundInterceptor.ExecuteActivity(ctx, in)
}
