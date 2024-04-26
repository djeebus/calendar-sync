package cmd

import (
	"calendar-sync/pkg/temporal/activities"
	"calendar-sync/pkg/temporal/workflows"
	"context"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"golang.org/x/oauth2"
	"os"
	"os/signal"
	"syscall"
)

var workerCmd = cobra.Command{
	Use: "worker",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)

		c, err := client.Dial(client.Options{})
		if err != nil {
			cancel()
			return errors.Wrap(err, "failed to dial the temporal server")
		}
		defer c.Close()

		w := worker.New(c, "default", worker.Options{})

		workflows.Register(w)
		activities.Register(w)

		if err := runWorker(ctx, w); err != nil {
			log.Error().Err(err).Msg("failed to start worker")
		}

		go waitForInterrupt(cancel)

		select {
		case <-ctx.Done():
			break
		}

		return nil
	},
}

func waitForInterrupt(cancel func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		cancel()
	}()
}

func runWorker(ctx context.Context, w worker.Worker) error {
	return w.Start()
}

func triggerSchedule(ctx context.Context, c client.Client, webcal, gmail string, token *oauth2.Token, immediate bool) error {
	opts := client.StartWorkflowOptions{
		ID:        "sync-calendar",
		TaskQueue: "default",
	}

	if !immediate {
		opts.CronSchedule = "@every 15m"
	}

	args := workflows.DiffCalendarWorkflowArgs{
		CalendarID: webcal,
		EmailToAdd: gmail,
		Token:      token,
	}

	if _, err := c.ExecuteWorkflow(ctx, opts, workflows.DiffCalendarWorkflow, args); err != nil {
		return errors.Wrap(err, "failed to start cron job")
	}

	return nil
}

func init() {
	rootCmd.AddCommand(&workerCmd)
}
