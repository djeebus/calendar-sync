package cmd

import (
	"calendar-sync/pkg/container"
	"calendar-sync/pkg/temporal"
	"calendar-sync/pkg/temporal/workflows"
	"calendar-sync/pkg/www"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"go.temporal.io/sdk/client"
	"os"
	"os/signal"
	"syscall"
)

var rootCmd = &cobra.Command{
	Use: "",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		errs := make(chan error)

		ctr, cleanup, err := container.New(ctx)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		defer cleanup()

		go startTemporalWorker(ctx, ctr, errs)

		go startWebserver(ctx, ctr, ":31425", errs)

		go triggerScheduledJobs(ctx, ctr)

		waitForInterrupt(ctx, errs)
	},
}

func triggerScheduledJobs(ctx context.Context, ctr container.Container) {
	opts1 := client.StartWorkflowOptions{
		ID:           "hourly-sync-check",
		TaskQueue:    ctr.TaskQueue,
		CronSchedule: "*/15 * * * *",
	}
	if _, err := ctr.TemporalClient.ExecuteWorkflow(ctx, opts1, workflows.CopyAllWorkflow); err != nil {
		log.Err(err).Msg("failed to trigger copy all calendars cronjob")
	}

	opts2 := client.StartWorkflowOptions{
		ID:           "hourly-invite-check",
		TaskQueue:    ctr.TaskQueue,
		CronSchedule: "*/15 * * * *",
	}
	if _, err := ctr.TemporalClient.ExecuteWorkflow(ctx, opts2, workflows.InviteAllWorkflow); err != nil {
		log.Err(err).Msg("failed to trigger invite all calendars cronjob")
	}
}

func waitForInterrupt(ctx context.Context, errs chan error) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	select {
	case <-c:
		break
	case <-errs:
		break
	case <-ctx.Done():
		break
	}
}

func startWebserver(ctx context.Context, ctr container.Container, listen string, errs chan error) {
	s := www.NewServer(ctr)
	if err := s.Start(listen); err != nil {
		errs <- err
	}
}

func startTemporalWorker(ctx context.Context, ctr container.Container, errs chan error) {
	w, err := temporal.NewWorker(ctx, ctr)
	if err != nil {
		errs <- errors.Wrap(err, "failed to create new worker")
	}

	if err := w.Start(); err != nil {
		errs <- errors.Wrap(err, "failed to start worker")
	}
}

func Main() error {
	return rootCmd.Execute()
}
