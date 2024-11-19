package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"calendar-sync/pkg"
	"calendar-sync/pkg/container"
	"calendar-sync/pkg/tasks/activities"
	"calendar-sync/pkg/tasks/workflows"
	"calendar-sync/pkg/www"
)

var CommitSHA = "unknown"
var CommitRef = "unknown"
var BuildDate = "unknown"

type job struct {
	workflowID string
	workflow   func(ctx context.Context, w *workflows.Workflows) error
}

var jobs = []job{
	{
		workflow: func(ctx context.Context, w *workflows.Workflows) error {
			return w.CopyAllWorkflow(ctx)
		},
		workflowID: "hourly-sync-check",
	},
	{
		workflowID: "hourly-invite-check",
		workflow: func(ctx context.Context, w *workflows.Workflows) error {
			return w.InviteAllWorkflow(ctx)
		},
	},
	{
		workflowID: "hourly-webhook-check",
		workflow: func(ctx context.Context, w *workflows.Workflows) error {
			return w.WatchAll(ctx)
		},
	},
}

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

		a := activities.New(ctr)
		w := workflows.New(a)

		go startWebserver(ctr, w, cfg.Listen, errs)

		go func() {
			triggerScheduledJobs(ctx, w, true, jobs)

			// reschedule cron job every 24 hours.
			// if temporal restarts, these jobs disappear
			for {
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Hour):
					triggerScheduledJobs(ctx, w, false, jobs)
				}
			}
		}()

		log.Info().Msg("waiting for interrupts ...")
		waitForInterrupt(ctx, errs)
	},
}

func triggerScheduledJobs(ctx context.Context, w *workflows.Workflows, now bool, jobs []job) {
	for _, job := range jobs {
		jobID := job.workflowID
		if now {
			jobID += "-init"
		}

		log.Info().Msgf("trigger scheduled job: %s", jobID)
		if err := job.workflow(ctx, w); err != nil {
			log.Err(err).Msgf("failed to trigger %q job", jobID)
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

func startWebserver(ctr container.Container, workflows *workflows.Workflows, listen string, errs chan error) {
	log.Info().Msg("starting web server")
	s := www.NewServer(ctr, workflows)
	if err := s.Start(listen); err != nil {
		errs <- err
	}
}

func Main() error {
	return rootCmd.Execute()
}
