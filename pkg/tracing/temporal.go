package tracing

import (
	"context"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/workflow"
)

func NewCorrelationIDPropagator() workflow.ContextPropagator {
	return CorrelationIDPropagator{
		converter.GetDefaultDataConverter(),
	}
}

type CorrelationIDPropagator struct {
	dataConverter converter.DataConverter
}

const headerKey = "correlation-id"

func (c CorrelationIDPropagator) Inject(ctx context.Context, writer workflow.HeaderWriter) error {
	value, ok := GetCorrelationID(ctx)
	if !ok {
		return nil
	}

	payload, err := c.dataConverter.ToPayload(value)
	if err != nil {
		return errors.Wrap(err, "failed to convert correlation id")
	}

	writer.Set(headerKey, payload)

	return nil
}

func (c CorrelationIDPropagator) Extract(ctx context.Context, reader workflow.HeaderReader) (context.Context, error) {
	if value, ok := reader.Get(headerKey); ok {
		var correlationID string
		if err := c.dataConverter.FromPayload(value, &correlationID); err != nil {
			return nil, errors.Wrap(err, "failed to convert from payload")
		}

		ctx = setCorrelationID(ctx, correlationID)
	}

	return ctx, nil
}

func (c CorrelationIDPropagator) InjectFromWorkflow(ctx workflow.Context, writer workflow.HeaderWriter) error {
	correlationID, ok := getCorrelationID(ctx)
	if !ok {
		return nil
	}

	payload, err := c.dataConverter.ToPayload(correlationID)
	if err != nil {
		return errors.Wrap(err, "failed to convert correlation id")
	}

	writer.Set(headerKey, payload)
	return nil
}

func (c CorrelationIDPropagator) ExtractToWorkflow(ctx workflow.Context, reader workflow.HeaderReader) (workflow.Context, error) {
	if value, ok := reader.Get(headerKey); ok {
		var correlationID string
		if err := c.dataConverter.FromPayload(value, &correlationID); err != nil {
			return nil, errors.Wrap(err, "failed to convert from payload")
		}

		ctx = workflow.WithValue(ctx, key, correlationID)
	}

	return ctx, nil
}
