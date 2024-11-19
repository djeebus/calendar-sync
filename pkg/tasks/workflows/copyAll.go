package workflows

import (
	"context"
	"sync"

	"github.com/pkg/errors"
)

func (w *Workflows) CopyAllWorkflow(ctx context.Context) error {
	ctx, log := setupLogger(ctx, "CopyAllWorkflow")

	copyConfigs, err := w.a.GetAllCopies(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get copies")
	}

	var wg sync.WaitGroup
	for _, copyConfig := range copyConfigs.CopyConfigs {
		args := CopyCalendarWorkflowArgs{
			copyConfig.SourceID,
			copyConfig.DestinationID,
		}
		wg.Add(1)
		go func(args CopyCalendarWorkflowArgs) {
			err := w.CopyCalendarWorkflow(ctx, args)
			if err != nil {
				log.Error().Err(err).
					Str("source-id", copyConfig.SourceID).
					Str("destination-id", copyConfig.DestinationID).
					Msg("failed to copy calendar")
			}
		}(args)
	}

	wg.Done()
	return nil
}
