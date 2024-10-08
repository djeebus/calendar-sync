package logs

import (
	"context"
	"os"

	"github.com/rs/zerolog"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
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

	log zerolog.Logger
}

func (i loggingInterceptor) InterceptActivity(ctx context.Context, next interceptor.ActivityInboundInterceptor) interceptor.ActivityInboundInterceptor {
	return loggingActivityInterceptor{next, i.log}
}

func (i loggingInterceptor) InterceptWorkflow(ctx workflow.Context, next interceptor.WorkflowInboundInterceptor) interceptor.WorkflowInboundInterceptor {
	return loggingWorkflowInterceptor{WorkflowInboundInterceptor: next, log: i.log}
}

type loggingActivityInterceptor struct {
	interceptor.ActivityInboundInterceptor

	log zerolog.Logger
}

func (i loggingActivityInterceptor) ExecuteActivity(ctx context.Context, in *interceptor.ExecuteActivityInput) (interface{}, error) {
	info := activity.GetInfo(ctx)

	logger := i.log.With().
		Str("activity_name", info.ActivityType.Name).
		Str("activity_id", info.ActivityID).
		Bool("is_local", info.IsLocalActivity).
		Logger()

	ctx = SetLogger(ctx, logger)

	logger.Info().Msg("starting activity")

	result, err := i.ActivityInboundInterceptor.ExecuteActivity(ctx, in)

	if err != nil {
		logger.Warn().Err(err).Msg("activity finished with error")
	} else {
		logger.Info().Msg("activity finished")
	}

	return result, err
}

type loggingWorkflowInterceptor struct {
	interceptor.WorkflowInboundInterceptor

	log zerolog.Logger
}

type logKeyType struct{}

var logKey = logKeyType{}

func (i loggingWorkflowInterceptor) ExecuteWorkflow(ctx workflow.Context, in *interceptor.ExecuteWorkflowInput) (interface{}, error) {
	info := workflow.GetInfo(ctx)

	logger := i.log.With().
		Str("workflow_name", info.WorkflowType.Name).
		Logger()

	ctx = workflow.WithValue(ctx, logKey, logger)

	logger.Info().Msg("starting workflow")

	result, err := i.WorkflowInboundInterceptor.ExecuteWorkflow(ctx, in)

	if err != nil {
		logger.Warn().Err(err).Msg("workflow finished with error")
	} else {
		logger.Info().Msg("workflow finished")
	}

	return result, err
}

func GetWorkflowLogger(ctx workflow.Context) *zerolog.Logger {
	value := ctx.Value(logKey)
	if value != nil {
		if logger, ok := value.(*zerolog.Logger); ok {
			return logger
		}
	}

	if logger := zerolog.DefaultContextLogger; logger != nil {
		return zerolog.DefaultContextLogger
	}

	logger := zerolog.New(os.Stdout)
	return &logger
}
