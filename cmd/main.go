package cmd

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.temporal.io/sdk/client"

	"calendar-sync/pkg"
	"calendar-sync/pkg/container"
	"calendar-sync/pkg/logs"
	"calendar-sync/pkg/temporal"
	"calendar-sync/pkg/temporal/workflows"
	"calendar-sync/pkg/www"
)

var CommitSHA = "unknown"
var CommitRef = "unknown"
var BuildDate = "unknown"

var rootCmd = &cobra.Command{
	Use:     "",
	Version: fmt.Sprintf("SHA:%s, build:%s, ref:%s", CommitSHA, BuildDate, CommitRef),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		errs := make(chan error)

		cfg, err := pkg.ReadConfig()
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		log.Info().Msgf("commit sha: %s", CommitSHA)
		log.Info().Msgf("commit ref: %s", CommitRef)
		log.Info().Msgf("build date: %s", BuildDate)

		ctr, err := container.New(ctx, cfg)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		defer ctr.Close()

		go startTemporalWorker(ctx, ctr, errs)

		go startWebserver(ctx, ctr, cfg.Listen, errs)

		go func() {
			// reschedule cron job every 24 hours.
			// if temporal restarts, these jobs disappear
			for {
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Hour * 24):
					triggerScheduledJobs(ctx, ctr)
				}
			}
		}()

		waitForInterrupt(ctx, errs)
	},
}

func triggerScheduledJobs(ctx context.Context, ctr container.Container) {
	log := logs.GetLogger(ctx)

	hourlySyncCheckOpts := client.StartWorkflowOptions{
		ID:           "hourly-sync-check",
		TaskQueue:    ctr.Config.TemporalTaskQueue,
		CronSchedule: "0 * * * *",
	}
	if _, err := ctr.TemporalClient.ExecuteWorkflow(ctx, hourlySyncCheckOpts, workflows.CopyAllWorkflow); err != nil {
		log.Err(err).Msg("failed to trigger copy all calendars cronjob")
	}

	hourlyInviteCheck := client.StartWorkflowOptions{
		ID:           "hourly-invite-check",
		TaskQueue:    ctr.Config.TemporalTaskQueue,
		CronSchedule: "15 * * * *",
	}
	if _, err := ctr.TemporalClient.ExecuteWorkflow(ctx, hourlyInviteCheck, workflows.InviteAllWorkflow); err != nil {
		log.Err(err).Msg("failed to trigger invite all calendars cronjob")
	}

	if ctr.Config.WebhookUrl != "" {
		opts3 := client.StartWorkflowOptions{
			ID:           "webhook-check",
			TaskQueue:    ctr.Config.TemporalTaskQueue,
			CronSchedule: "30 * * * *",
		}
		if _, err := ctr.TemporalClient.ExecuteWorkflow(ctx, opts3, workflows.WatchAll); err != nil {
			log.Err(err).Msg("failed to trigger watch all calendars cronjob")
		}
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
