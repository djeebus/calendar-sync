package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"go.temporal.io/sdk/client"

	"calendar-sync/pkg"
	"calendar-sync/pkg/container"
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

		go startWebserver(ctr, cfg.Listen, errs)

		go func() {
			triggerScheduledJobs(ctx, ctr, true)

			// reschedule cron job every 24 hours.
			// if temporal restarts, these jobs disappear
			for {
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Hour * 24):
					triggerScheduledJobs(ctx, ctr, false)
				}
			}
		}()

		log.Info().Msg("waiting for interrupts ...")
		waitForInterrupt(ctx, errs)
	},
}

var scheduledTasks = []struct {
	workflowID string
	schedule   string
	workflow   any
}{
	{
		workflowID: "hourly-sync-check",
		schedule:   "0 * * * *",
		workflow:   workflows.CopyAllWorkflow,
	},
	{
		workflowID: "hourly-invite-check",
		schedule:   "15 * * * *",
		workflow:   workflows.InviteAllWorkflow,
	},
	{
		workflowID: "webhook-check",
		schedule:   "30 * * * *",
		workflow:   workflows.WatchAll,
	},
}

func triggerScheduledJobs(ctx context.Context, ctr container.Container, now bool) {
	for _, scheduledTask := range scheduledTasks {
		opts := client.StartWorkflowOptions{
			ID:        scheduledTask.workflowID,
			TaskQueue: ctr.Config.TemporalTaskQueue,
		}
		if now {
			opts.CronSchedule = scheduledTask.schedule
		}

		log.Info().Msgf("trigger scheduled job: %s", opts.ID)
		if _, err := ctr.TemporalClient.ExecuteWorkflow(ctx, opts, scheduledTask.workflow); err != nil {
			log.Err(err).Msgf("failed to trigger %q job", opts.ID)
		}
	}
}

func waitForInterrupt(ctx context.Context, errs chan error) {
	sigTerm := make(chan os.Signal, 1)
	signal.Notify(sigTerm, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-sigTerm:
		log.Info().Str("signal", s.String()).Msg("terminate caught")
		break
	case err := <-errs:
		log.Info().Err(err).Msg("error encountered")
		break
	case <-ctx.Done():
		log.Info().Msg("global context canceled")
		break
	}
}

func startWebserver(ctr container.Container, listen string, errs chan error) {
	log.Info().Msg("starting web server")
	s := www.NewServer(ctr)
	if err := s.Start(listen); err != nil {
		errs <- err
	}
}

func startTemporalWorker(ctx context.Context, ctr container.Container, errs chan error) {
	log.Info().Msg("starting temporal worker")
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
